import {Controller, Middlewares, Path, Response, Get, Route, Tags} from 'tsoa';

import {ErrorResponse, ErrorTypes, getErrorResponse} from '../../internal-ts/errors.esm.js';
import {commandQueue, cronJobDao} from '../../api-ts/initializers.esm.js';
import JWT from '../../internal-ts/token.esm.js';
import {CommandResult} from "../../queue/cmdqueue.esm.js";
import {errorHandler} from "../errhandler.esm.js";
import {CronJobTask, CronJobTaskService} from "../../queue/cron.esm.js"; // Assuming that this is a JWT middleware function

interface CronJobsResponse {
    crons: CronJobTask[];
}

@Route('/api/v1/queue')
@Tags('Queue')
export class QueueController extends Controller {

    @Response<CommandResult>(200, 'Success')
    @Response<ErrorResponse>(400, 'Error')
    @Response<ErrorResponse>(500, 'Error')
    @Get('{taskId}/status')
    async status(@Path() taskId: string): Promise<CommandResult> {
        console.log(`[QUEUE]: Received status command : `, taskId)
        let result: CommandResult | null = null;
        try {
            result = await commandQueue.getStatus(taskId);
            if (result != null) {
                return result
            }
        } catch (err: any) {
            throw getErrorResponse(ErrorTypes.INTERNAL_SERVER_ERROR, err.message);
        }

        console.log("[QUEUE]: Task not found returning 404")
        throw getErrorResponse(ErrorTypes.RESOURCE_NOT_FOUND);
    }

    @Response<CronJobsResponse>(200, 'Success')
    @Response<ErrorResponse>(400, 'Error')
    @Response<ErrorResponse>(500, 'Error')
    @Get('crons')
    async getCrons(): Promise<CronJobsResponse> {
        console.log(`[QUEUE]: Received getCrons command : `)
        let result: CronJobTask[] | null = null;
        try {
            result = await cronJobDao.all();
            if (result != null) {
                return {crons: result}
            }
        } catch (err: any) {
            throw getErrorResponse(ErrorTypes.INTERNAL_SERVER_ERROR, err.message);
        }

        console.log("[QUEUE CRON]: Task not found returning 404")
        throw getErrorResponse(ErrorTypes.RESOURCE_NOT_FOUND);
    }

    @Response<CronJobsResponse>(200, 'Success')
    @Response<ErrorResponse>(400, 'Error')
    @Response<ErrorResponse>(500, 'Error')
    @Get('crons/{cronId}')
    async getCron(@Path() cronId: number): Promise<CronJobTask> {
        console.log(`[QUEUE]: Received getCron command : ${cronId}`)
        let result: CronJobTask | null = null;
        try {
            result = await cronJobDao.getById(cronId);
            if (result != null) {
                return result
            }
        } catch (err: any) {
            throw getErrorResponse(ErrorTypes.INTERNAL_SERVER_ERROR, err.message);
        }

        console.log("[QUEUE CRON]: Task not found returning 404")
        throw getErrorResponse(ErrorTypes.RESOURCE_NOT_FOUND);
    }
}