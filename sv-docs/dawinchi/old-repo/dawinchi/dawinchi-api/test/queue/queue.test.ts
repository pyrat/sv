import { Database } from 'sqlite3';
import { SQLiteApplicationLogDao, ApplicationLog, Level, Order } from '../../queue/applog.esm';
import { CommandQueueImpl, CommandTypes, QeueuEventTypes } from '../../queue/cmdqueue.esm';
import {Control, ControlHandler} from '../../internal-ts/control.esm.js';

import * as fs from 'fs';

describe('Queue', () => {
    let db: Database;
    let dao: SQLiteApplicationLogDao;
    let queue: CommandQueueImpl;
    let motorControl: Control;

    const dbPath = './test-queue.db';

    beforeEach(async () => {
        db = new Database(dbPath);
        dao = new SQLiteApplicationLogDao(db);
        await dao.createTable();

        motorControl = new ControlHandler();

        queue = new CommandQueueImpl(dao, motorControl);
    });

    afterEach(() => {
        db.close();
        fs.unlinkSync(dbPath); // Delete the test database file
    });

    test("should lock and unlock", async () => {  
        expect(queue.isLocked()).toBe(false);
        queue.lock();
        expect(queue.isLocked()).toBe(true);
        queue.unlock();
    });

    test("should push a command to a locked queue", async () => {
        queue.lock();
        const command = {
            CommandType: CommandTypes.SIMPLE,
            Parameters: { test: 'payload' },
        };

        expect(() => queue.push(command)).toThrowError('Queue is locked');
    });


    test("should push a command", async () => {
        const command = {
            CommandType: CommandTypes.SIMPLE,
            Parameters: { test: 'payload' },
        };

        const taskId = queue.push(command);

        
        // sleep for 1 second
        await new Promise(resolve => setTimeout(resolve, 2000));

        expect(queue.isLocked()).toBe(false);
        const logs = await dao.filter(QeueuEventTypes.COMMAND_PROCESSED)
        console.log(logs);
        expect(logs.length).toBe(1);
        expect(logs[0].Payload.TaskId).toEqual(taskId);
        expect(logs[0].Payload.Command.Parameters).toEqual(command.Parameters);
        expect(logs[0].Payload.Command.CommandType).toEqual(command.CommandType);

    });


    test("should push a command that fails", async () => {
        const command = {
            CommandType: CommandTypes.FAILURE,
            Parameters: { test: 'payload' },
        };

        const taskID = queue.push(command);
        
        // sleep for 1 second
        await new Promise(resolve => setTimeout(resolve, 2000));

        expect(queue.isLocked()).toBe(false);
        const logs = await dao.filter(QeueuEventTypes.COMMAND_FAILED)
        console.log(logs);
        expect(logs.length).toBe(1);
        expect(logs[0].Payload.Command.TaskId).toEqual(taskID);
        expect(logs[0].EventID).toEqual(taskID);
    });

    test("should push a command that is a bit longer to process", async () => {
        const command = {
            CommandType: CommandTypes.SLEEP,
            Parameters: { test: 'payload' },
        };

        const taskId = queue.push(command);
        
        // sleep for 1 second
        await new Promise(resolve => setTimeout(resolve, 2000));

        expect(queue.isLocked()).toBe(false);
        const logs = await dao.filter(QeueuEventTypes.COMMAND_PROCESSED)
        console.log(logs);
        expect(logs.length).toBe(1);
        expect(logs[0].Payload.Command.TaskId).toEqual(taskId);
        expect(logs[0].EventID).toEqual(taskId);
    });

    test("should push a command that fails with exception", async () => {
        const command = {
            CommandType: CommandTypes.EXCEPTION,
            Parameters: { test: 'payload' },
        };

        const taskId = queue.push(command);
        
        // sleep for 1 second
        await new Promise(resolve => setTimeout(resolve, 2000));

        expect(queue.isLocked()).toBe(false);
        const logs = await dao.filter(QeueuEventTypes.COMMAND_FAILED)
        console.log(logs);
        expect(logs.length).toBe(1);
        expect(logs[0].Payload.Command.TaskId).toEqual(taskId);
        expect(logs[0].Payload.Error).toContain('Command failed with exception');
        expect(logs[0].EventID).toEqual(taskId);
    });

    test("should check the status of a command", async () => {
        const command = {
            CommandType: CommandTypes.SIMPLE,
            Parameters: { test: 'payload' },
        };

        const taskId = queue.push(command);

        
        // sleep for 1 second
        await new Promise(resolve => setTimeout(resolve, 2000));

        expect(queue.isLocked()).toBe(false);
        const logs = await dao.filter(QeueuEventTypes.COMMAND_PROCESSED)
        console.log(logs);
        expect(logs.length).toBe(1);
        expect(logs[0].Payload.TaskId).toEqual(taskId);
        expect(logs[0].Payload.Command.Parameters).toEqual(command.Parameters);
        expect(logs[0].Payload.Command.CommandType).toEqual(command.CommandType); 

        // check the status of the command
        const commandResult = await queue.getStatus(taskId);
        expect(commandResult).toEqual({...logs[0].Payload, "Status": "COMPLETED"});
        
        const noFilter = await dao.all();
        console.log("All logs");
        console.log(noFilter);

        // this is the last log entry let's check the rest of them
        const allLogs = await dao.getById(taskId);
        // console.log(allLogs);
        expect(allLogs).not.toBeNull();
        if (allLogs != null){
            expect(allLogs.length).toBe(2);
        }

    });

});