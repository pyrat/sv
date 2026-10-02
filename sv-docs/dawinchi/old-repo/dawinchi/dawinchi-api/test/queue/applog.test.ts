import {Database} from 'sqlite3';
import {SQLiteApplicationLogDao, ApplicationLog, Level, Order, ApplicationLogResult} from '../../queue/applog.esm';
import * as fs from 'fs';

describe('SQLiteApplicationLogDao', () => {
    let db: Database;
    let dao: SQLiteApplicationLogDao;
    const dbPath = './test.db';

    beforeEach(async () => {
        db = new Database(dbPath);
        dao = new SQLiteApplicationLogDao(db);
        await dao.createTable();
    });

    afterEach(() => {
        db.close();
        fs.unlinkSync(dbPath); // Delete the test database file
    });

    test('should insert and retrieve a log', async () => {
        const log: ApplicationLogResult = {
            Id: 1,
            EventID: '1',
            EventType: 'Test',
            Payload: {test: 'payload'},
            CreatedAt: '2023-07-05',
            Level: Level.INFO,
        };

        await dao.create(log);

        const retrievedLog = await dao.getById('1');
        expect(retrievedLog).toEqual([log]);
    });

    test('should delete a log', async () => {
        await dao.delete('1');

        const retrievedLog = await dao.getById('1');
        expect(retrievedLog).toEqual([]);
    });

    test('should filter logs', async () => {
        const log1: ApplicationLogResult = {
            Id: 1,
            EventID: '2',
            EventType: 'Filter Test',
            Payload: {test: 'payload'},
            CreatedAt: '2023-07-04',
            Level: Level.INFO,
        };

        const log2: ApplicationLogResult = {
            Id: 2,
            EventID: '3',
            EventType: 'Filter Test',
            Payload: {test: 'payload'},
            CreatedAt: '2023-07-07',
            Level: Level.ERROR,
        };

        await dao.create(log1);
        await dao.create(log2);

        const all = await dao.all();
        // console.log(all);
        expect(all).toEqual([log1, log2]);

        const filteredLogs = await dao.filter('Filter Test', Level.ERROR, '2023-07-06', '2023-07-08');
        expect(filteredLogs).toEqual([log2]);


        const filteredLogs2 = await dao.filter('Filter Test', Level.INFO);
        expect(filteredLogs2).toEqual([log1]);

        // test ordering
        const filteredLogs3 = await dao.filter(
            'Filter Test',
            undefined,
            undefined,
            undefined,
             undefined,
             Order.ASC);
        expect(filteredLogs3).toEqual([log1, log2]);

        // test ordering
        const filteredLogs4 = await dao.filter('Filter Test');
        console.log(filteredLogs4);
        expect(filteredLogs4).toEqual([log2, log1]);

    });
});
