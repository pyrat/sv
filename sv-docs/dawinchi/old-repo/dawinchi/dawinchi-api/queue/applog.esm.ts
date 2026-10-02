import sqlite3 from 'sqlite3';
import {promisify} from 'util';


export enum Level {
    INFO = 'INFO',
    ERROR = 'ERROR',
    FATAL = 'FATAL',
    WARNING = 'WARNING'
}

export enum Order {
    ASC = 'ASC',
    DESC = 'DESC'
}

export interface ApplicationLog {
    EventID: string;
    EventType: string;
    Payload: any;
    CreatedAt: string;
    Level: Level;
}

export interface ApplicationLogResult extends ApplicationLog {
    Id: number;
}

export interface ApplicationLogDao {
    createTable(): void;

    create(log: ApplicationLog): void;

    getById(eventId: string): Promise<ApplicationLogResult[] | null>;

    delete(eventId: string): void;

    filter(eventType?: string, level?: string, startDate?: string, endDate?: string, eventId?: string, order?: Order, limit?: number): Promise<ApplicationLogResult[]>;

    all(): Promise<ApplicationLogResult[]>;
}

export class SQLiteApplicationLogDao implements ApplicationLogDao {
    private db: sqlite3.Database;

    constructor(db: sqlite3.Database) {
        this.db = db;
    }

    static fromFilePath(dbFilePath: string) {
        return new SQLiteApplicationLogDao(new sqlite3.Database(dbFilePath));
    }

    async createTable() {
        const sql = `
            CREATE TABLE IF NOT EXISTS ApplicationLog
            (
                Id
                INTEGER
                PRIMARY
                KEY
                AUTOINCREMENT,
                EventID
                TEXT,
                EventType
                TEXT,
                Payload
                TEXT,
                CreatedAt
                TEXT,
                Level
                TEXT
                CHECK (
                Level
                IN
            (
                'INFO',
                'ERROR',
                'FATAL',
                'WARNING'
            ))
                )`;
        const runAsync = promisify((sql: string, params: any, callback: Function) => {
            this.db.run(sql, params, callback);
        });
        await runAsync(sql, []);
    }

    async create(log: ApplicationLog) {
        const runAsync = promisify((sql: string, params: any, callback: Function) => {
            this.db.run(sql, params, callback);
        });
        await runAsync(
            'INSERT INTO ApplicationLog (EventID, EventType, Payload, CreatedAt, Level) VALUES (?, ?, ?, ?, ?)',
            [log.EventID, log.EventType, JSON.stringify(log.Payload), log.CreatedAt, log.Level]
        );
    }

    async delete(eventId: string) {
        const runAsync = promisify((sql: string, params: any, callback: Function) => {
            this.db.run(sql, params, callback);
        });
        await runAsync(
            'DELETE FROM ApplicationLog WHERE EventID = ?',
            [eventId]
        );
    }

    async getById(eventId: string): Promise<ApplicationLogResult[] | null> {
        const params: string[] = [eventId];
        const query = 'SELECT * FROM ApplicationLog WHERE EventID = ? order by Id asc';

        const allAsync = promisify((sql: string, params: any[], callback: (err: Error | null, rows?: ApplicationLog[]) => void) => {
            this.db.all(sql, params, (err, rows: ApplicationLogResult[]) => {
                callback(err, rows);
            });
        });

        const rows = await allAsync(query, params) as ApplicationLog[];
        const logs: ApplicationLogResult[] = rows.map(row => {
            row.Payload = JSON.parse(row.Payload);
            return row as ApplicationLogResult;
        });
        return logs;
    }

    async filter(eventType?: string, level?: Level, startDate?: string, endDate?: string, eventId?: string, order = Order.DESC, limit: number = 20): Promise<ApplicationLogResult[]> {
        let query = 'SELECT * FROM ApplicationLog WHERE 1=1';
        const params: string[] = [];
        if (startDate && endDate) {
            query += ' AND CreatedAt BETWEEN ? AND ?';
            params.push(startDate, endDate);
        }
        if (eventType) {
            query += ' AND EventType = ?';
            params.push(eventType);
        }
        if (level) {
            query += ' AND Level = ?';
            params.push(level);
        }

        if (eventId) {
            query += ' AND EventID = ?';
            params.push(eventId);
        }


        query += ` ORDER BY CreatedAt ${order}`;

        if (limit) {
            query += ' LIMIT ?';
            params.push(limit.toString());
        }

        console.log(`QUERY : ${query}`);
        console.log(`PARAMS : ${params}`);

        const allAsync = promisify((sql: string, params: any[], callback: (err: Error | null, rows?: ApplicationLogResult[]) => void) => {
            this.db.all(sql, params, (err, rows: ApplicationLogResult[]) => {
                callback(err, rows);
            });
        });

        const rows = await allAsync(query, params) as ApplicationLog[];
        const logs: ApplicationLogResult[] = rows.map(row => {
            row.Payload = JSON.parse(row.Payload);
            return row as ApplicationLogResult;
        });
        return logs;
    }

    async all(): Promise<ApplicationLogResult[]> {
        const allAsync = promisify((sql: string, params: any[], callback: (err: Error | null, rows?: ApplicationLogResult[]) => void) => {
            this.db.all(sql, params, (err, rows: ApplicationLogResult[]) => {
                callback(err, rows);
            });
        });

        const rows = await allAsync('SELECT * FROM ApplicationLog', []) as ApplicationLogResult[];
        const logs: ApplicationLogResult[] = rows.map(row => {
            row.Payload = JSON.parse(row.Payload);
            return row as ApplicationLogResult;
        });
        return logs;
    }
}