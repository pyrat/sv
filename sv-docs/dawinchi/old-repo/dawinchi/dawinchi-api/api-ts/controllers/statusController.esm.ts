import express from 'express';
import { Body, Get, Route, Response, Request, Middlewares } from 'tsoa';

import Errors from '../../internal-ts/errors.esm.js';
import Control from '../../internal-ts/control.esm.js';
import Logger from '../../internal-ts/logger.esm.js';
import JWT from '../../internal-ts/token.esm.js'; // Assuming that this is a JWT middleware function

interface StatusResponse {
    success: boolean;
    status?: string;
}

@Route('/api/v1/status') // TSOA uses this as the basePath
export class StatusController {
    @Response<StatusResponse>(200, 'Success')
    @Response<StatusResponse>(400, 'Error') // This is just a guess. Replace with your actual Error model and status code.
    @Get() // The method decorator matches the HTTP verb, e.g., @Post, @Get, @Put, etc.
    @Middlewares([JWT.authenticateToken])
    async status(): Promise<StatusResponse> {
        try {
            return { success: true};
        } catch (err: any) {
            throw {
                status: Errors[err.fault].status,
                success: false,
                message: Errors[err.fault].message,
                faultCode: err.fault,
            };
        }
    }
}