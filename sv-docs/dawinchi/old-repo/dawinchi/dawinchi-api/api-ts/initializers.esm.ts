import {Control} from "../internal-ts/control.esm.js";
import {ApplicationLogDao, SQLiteApplicationLogDao} from "../queue/applog.esm.js";
import ControlHandler from "../internal-ts/control.esm.js";
import {CommandQueue, CommandQueueImpl} from "../queue/cmdqueue.esm.js";
import {CronJobTaskService, SQLiteCronDao} from "../queue/cron.esm.js";
import sqlite3 from "sqlite3";

// can just import the applicationLog and motorControl variables
// and use them in the other files
export let applicationLog: ApplicationLogDao;
export let motorControl: Control;
export let commandQueue: CommandQueue;

export let cronJobDao: SQLiteCronDao;
export let cronJobScheduler: CronJobTaskService;

(async () => {
    const db = new sqlite3.Database('./dawinchi.db')

    applicationLog = new SQLiteApplicationLogDao(db);
    await applicationLog.createTable();

    motorControl = ControlHandler;
    commandQueue = new CommandQueueImpl(applicationLog, motorControl);

    // create the cronjob service so can be called from api
    cronJobDao = new SQLiteCronDao(db);
    cronJobScheduler = new CronJobTaskService(cronJobDao, commandQueue);
    await cronJobScheduler.scheduleExistingJobs()
})();

