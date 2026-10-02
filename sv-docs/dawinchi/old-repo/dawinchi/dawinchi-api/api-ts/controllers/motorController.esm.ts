import {Body, Controller, Middlewares, Post, Response, Route, Tags} from 'tsoa';

import Errors, {ErrorResponse, ErrorTypes, getErrorResponse} from '../../internal-ts/errors.esm.js';
import {commandQueue, cronJobScheduler} from '../../api-ts/initializers.esm.js';
import {Directions} from '../../internal-ts/control.esm.js';
import JWT from '../../internal-ts/token.esm.js';
import {MoveToCoordinates, QueueMoveToRequest} from "../../queue/cmdqueue.esm.js"; // Assuming that this is a JWT middleware function

interface MoveResponse {
    success: boolean;
    taskId?: string;
    directions?: Directions;
}

interface RampResponse {
    success: boolean;
    message?: string;
}

interface MoveToRequest {
    coordinates: MoveToCoordinates[];
    schedule?: string;
}

interface MoveToResponse {
    success: boolean;
    taskId?: string;
}

@Route('/api/v1/motor') // TSOA uses this as the basePath
@Tags('Motor')
export class MoveController extends Controller {
    @Response<MoveResponse>(200, 'Success')
    @Response<ErrorResponse>(400, 'Error')
    @Response<ErrorResponse>(500, 'Error')
    @Post('move') // The method decorator matches the HTTP verb, e.g., @Post, @Get, @Put, etc.
    async move(@Body() directions: Directions): Promise<MoveResponse> {
        console.log(`[MOVE]: Received move command : `, directions);
        if (!directions || Object.keys(directions).length === 0) {
            throw getErrorResponse(ErrorTypes.INVALID_DIRECTION_REQUEST);
        }

        const taskId = commandQueue.push({CommandType: 'MOVE', Parameters: directions})
        return {
            success: true,
            taskId,
            directions
        };
    }

    @Response<MoveToResponse>(200, 'Success')
    @Response<ErrorResponse>(400, 'Error')
    @Response<ErrorResponse>(500, 'Error')
    @Post('moveto') // The method decorator matches the HTTP verb, e.g., @Post, @Get, @Put, etc.
    async moveTo(@Body() directions: MoveToRequest): Promise<MoveToResponse> {
        console.log(`[MOVETO]: Received move command : `, directions);
        if (!directions || directions.coordinates.length === 0) {
            throw getErrorResponse(ErrorTypes.INVALID_DIRECTION_REQUEST);
        }

        for (let coord of directions.coordinates) {
            if (!coord.x || !coord.y) {
                throw getErrorResponse(ErrorTypes.INVALID_DIRECTION_REQUEST);
            }

            if (coord.x < 0 || coord.x > 100 || coord.y < 0 || coord.y > 100) {
                throw getErrorResponse(ErrorTypes.INVALID_DIRECTION_REQUEST);
            }
        }

        let taskId: string | null = null;

        if (directions.schedule) {
            console.log(`[MOVETO]: Scheduling job : `, directions);
            const cronJob = await cronJobScheduler.scheduleNewJob({
                TaskName: 'MOVETO',
                Schedule: directions.schedule,
                TaskParameters: {Coordinates: directions.coordinates} as QueueMoveToRequest,
            });
            taskId = cronJob.Id.toString();
        } else {
            console.log(`[MOVETO]: Pushing job : `, directions)
            taskId = commandQueue.push({CommandType: 'MOVETO', Parameters: {Coordinates: directions.coordinates} as QueueMoveToRequest});
        }

        return {
            success: true,
            taskId,
        };
    }


    @Response<RampResponse>(200, 'Success')
    @Response<ErrorResponse>(400, 'Error') // This is just a guess. Replace with your actual Error model and status code.
    @Post('stop') // The method decorator matches the HTTP verb, e.g., @Post, @Get, @Put, etc.
    async stop()
        :
        Promise<RampResponse> {
        try {
            // await Control.stop();
            return {success: true};
        } catch (err: any) {
            throw {
                message: Errors[err.fault].message,
            };
        }
    }
}




