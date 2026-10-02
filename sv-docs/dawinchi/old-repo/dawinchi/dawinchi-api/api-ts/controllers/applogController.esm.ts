import {Controller, Get, Query, Response, Route, Tags} from 'tsoa';

import {ErrorResponse} from '../../internal-ts/errors.esm.js';
import {applicationLog} from '../../api-ts/initializers.esm.js';
import {ApplicationLogResult, Order} from '../../queue/applog.esm.js';

interface AppLogResponse {
    logs: ApplicationLogResult[]
}


@Route('/api/v1')
@Tags('Logs')
export class AppLogController extends Controller {

    @Response<AppLogResponse>(200, 'Success')
    @Response<ErrorResponse>(400, 'Error') // This is just a guess. Replace with your actual Error model and status code.
    @Response<ErrorResponse>(500, 'Internal Server Error')
    @Get('applog')
    public async applog(
        @Query() eventType?: string,
        @Query() level?: string,
        @Query() eventId?: string,
        @Query() from?: string,
        @Query() to?: string,
        @Query() limit?: number,
    ): Promise<AppLogResponse> {
        try {
            const logs = await applicationLog.filter(
                eventType,
                level,
                from,
                to,
                eventId,
                Order.DESC,
                limit
            )
            return {logs};
        } catch (err: any) {
            this.setStatus(500);
            throw{
                message: err.message,
            };
        }
    }
}