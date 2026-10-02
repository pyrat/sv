import {Database} from "sqlite3";
import fs from "fs";
import {SQLiteCronDao} from "../../queue/cron.esm";

describe('SQLiteCronDao', () => {
    let db: Database;
    let dao: SQLiteCronDao;
    const dbPath = './test-cron.db';

    beforeEach(async () => {
        db = new Database(dbPath);
        dao = new SQLiteCronDao(db);
        await dao.createTable();
    });

    afterEach(() => {
        db.close();
        fs.unlinkSync(dbPath); // Delete the test database file
    });

    test('should insert and retrieve a cron', async () => {
        const cron = {
            Schedule: '* * * * *',
            TaskName: 'Test',
            TaskParameters: {test: 'payload'},
        };

        const cronId = await dao.create(cron);
        console.log(`Created cron with id ${cronId}`);

        const retrievedCron = await dao.getById(cronId);
        expect(retrievedCron).not.toBeNull();
        console.log(`Retrieved cron : ${JSON.stringify(retrievedCron)}`);

        if (retrievedCron !== null) {
            expect(retrievedCron.Id).toEqual(cronId);
            expect(retrievedCron.Schedule).toEqual(cron.Schedule);
            expect(retrievedCron.TaskName).toEqual(cron.TaskName);
            expect(retrievedCron.TaskParameters).toEqual(cron.TaskParameters);
            expect(retrievedCron.LastRun).toBeNull();
            expect(retrievedCron.CreatedAt).not.toBeNull();
        }
    });

    test('should delete a cron', async () => {
    });

    test('should filter crons', async () => {
       const cron1 = {
            Schedule: '* * * * *',
            TaskName: 'Test',
            TaskParameters: {test: 'payload'},
        };

         const cron2 = {
            Schedule: '* 5 * * *',
            TaskName: 'Test2',
            TaskParameters: {test: 'payload2'},
        };

         const id1 = await dao.create(cron1);
         const id2 = await dao.create(cron2);

         const retrievedCrons = await dao.all();
         expect(retrievedCrons).not.toBeNull();
         expect(retrievedCrons).toHaveLength(2);
         expect(retrievedCrons[0]).toEqual({...cron1, Id: id1, LastRun: null, CreatedAt: expect.any(String)});
         expect(retrievedCrons[1]).toEqual({...cron2, Id: id2, LastRun: null, CreatedAt: expect.any(String)});
    });
});