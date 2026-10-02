import sqlite3 from 'sqlite3';
import {promisify} from 'util';
import {CommandQueue, CommandRequest} from "./cmdqueue.esm.js";
import cron from 'node-cron';
import cronvalidator from 'cron-validator';
import {getErrorResponse} from "../internal-ts/errors.esm.js";


export interface CronJobTaskCmd {
    Schedule: string;
    TaskName: string;
    TaskParameters: any;
}

export interface CronJobTask extends CronJobTaskCmd {
    Id: number;
    LastRun: string;
    CreatedAt: string;
}

export class SQLiteCronDao {

    private db: sqlite3.Database;

    constructor(db: sqlite3.Database) {
        this.db = db;
    }

    static fromFilePath(dbFilePath: string) {
        return new SQLiteCronDao(new sqlite3.Database(dbFilePath));
    }

    async createTable() {
        const sql = `
            CREATE TABLE IF NOT EXISTS cron_jobs
            (
                Id
                INTEGER
                PRIMARY
                KEY,
                Schedule
                TEXT
                NOT
                NULL,
                TaskName
                TEXT
                NOT
                NULL,
                TaskParameters
                TEXT
                NOT
                NULL,
                LastRun
                TEXT,
                CreatedAt
                TEXT
            )`;
        const runAsync = promisify((sql: string, params: any, callback: Function) => {
            this.db.run(sql, params, callback);
        });
        await runAsync(sql, []);
    }

    async create(cronTask: CronJobTaskCmd): Promise<number> {
        const runAsync = (sql: string, params: any[]): Promise<number> => {
            return new Promise<number>((resolve, reject) => {
                this.db.run(sql, params, function (err) {
                    if (err) {
                        reject(err);
                    } else {
                        resolve(this.lastID);
                    }
                });
            });
        };

        try {
            const lastId = await runAsync(
                'INSERT INTO cron_jobs (Schedule, TaskName, TaskParameters, CreatedAt) VALUES (?, ?, ?, ?)',
                [cronTask.Schedule, cronTask.TaskName, JSON.stringify(cronTask.TaskParameters), new Date().toISOString()]
            );
            return lastId;
        } catch (err) {
            console.error("Error inserting into database:", err);
            throw err;  // or handle it differently based on your requirements
        }
    }

    async all(): Promise<CronJobTask[]> {
        const allAsync = promisify((sql: string, params: any[], callback: (err: Error | null, rows?: CronJobTask[]) => void) => {
            this.db.all(sql, params, (err, rows: CronJobTask[]) => {
                callback(err, rows);
            });
        });

        const rows = await allAsync('SELECT * FROM cron_jobs', []) as CronJobTask[];
        return rows.map((row: any) => {
            return {
                Id: row.Id,
                Schedule: row.Schedule,
                TaskName: row.TaskName,
                TaskParameters: JSON.parse(row.TaskParameters),
                LastRun: row.LastRun,
                CreatedAt: row.CreatedAt,
            };
        });
    }


    async delete(id: string) {
        const runAsync = promisify((sql: string, params: any, callback: Function) => {
            this.db.run(sql, params, callback);
        });
        await runAsync('DELETE FROM cron_jobs WHERE id = ?', [id]);
    }

    async getById(id: number): Promise<CronJobTask | null> {
        const params: number[] = [id];
        const query = 'SELECT * FROM cron_jobs WHERE Id = ? order by Id asc';

        const allAsync = promisify((sql: string, params: any[], callback: (err: Error | null, rows?: CronJobTask[]) => void) => {
            this.db.all(sql, params, (err, rows: CronJobTask[]) => {
                callback(err, rows);
            });
        });

        const rows = await allAsync(query, params) as CronJobTask[];
        const crons: CronJobTask[] = rows.map(row => {
            const fetchedRow = row as CronJobTask;
            fetchedRow.TaskParameters = JSON.parse(fetchedRow.TaskParameters);
            return fetchedRow;
        });
        if (crons.length > 0) {
            return crons[0];
        }
        return null;
    }

    async updateLastRun(id: number, lastRun: string) {
        const runAsync = promisify((sql: string, params: any, callback: Function) => {
            this.db.run(sql, params, callback);
        });
        await runAsync('UPDATE cron_jobs SET LastRun = ? WHERE Id = ?', [lastRun, id]);
    }

}

export class CronJobTaskService {

    private dao: SQLiteCronDao;
    private cmdQueue: CommandQueue;

    constructor(dao: SQLiteCronDao, cmdQueue: CommandQueue) {
        this.dao = dao;
        this.cmdQueue = cmdQueue;
    }

    async scheduleExistingJobs() {
        await this.dao.createTable();

        const cronJobs = await this.dao.all();
        cronJobs.forEach(cronJob => {
            console.log(`Loading cron job: ${cronJob.TaskName} with schedule: ${cronJob.Schedule}`);
            const task = cron.schedule(cronJob.Schedule, async () => {
                console.log(`Submitting task: ${cronJob.TaskName} with schedule: ${cronJob.Schedule}`);
                const taskId = await this.runCronJob(cronJob);
                console.log(`Task submitted to queue with id: ${taskId}`);
            });
        });
    }

    async scheduleNewJob(cronJob: CronJobTaskCmd): Promise<CronJobTask> {
        if (!cronvalidator.isValidCron(cronJob.Schedule)){
            throw getErrorResponse(400, "Invalid cron expression")
        }
        const cronId = await this.dao.create(cronJob);
        const cronTask = await this.dao.getById(cronId);
        if (cronTask == null) {
            throw new Error("Error creating cron job");
        }

        const task = cron.schedule(cronTask.Schedule, async () => {
            console.log(`Submitting task: ${cronTask.TaskName} with schedule: ${cronTask.Schedule}`);
            const taskId = await this.runCronJob(cronTask);
            console.log(`Task submitted to queue with id: ${taskId}`);
        });

        return cronTask;
    }

    private async runCronJob(cronJob: CronJobTask): Promise<string> {
        const queueTask: CommandRequest = {
            CommandType: cronJob.TaskName,
            Parameters: {...cronJob.TaskParameters, CronJobId: cronJob.Id}
        }
        const taskId = this.cmdQueue.push(queueTask);
        await this.dao.updateLastRun(cronJob.Id, new Date().toISOString());
        return taskId;
    }
}