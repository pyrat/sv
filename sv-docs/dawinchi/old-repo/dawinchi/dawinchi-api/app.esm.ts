/** Imports */
import express, {Request, Response} from 'express';
import Logger from './internal-ts/logger.esm.js';
import API from './api-ts/api.esm.js';
import swaggerUi, { JsonObject } from 'swagger-ui-express';
import {RegisterRoutes} from './build-tsoa/routes.js';
import fs from 'fs'
import {errorHandler} from "./api-ts/errhandler.esm.js";


/** Destructure API export */
const { app } = API;
// register express routes from defined application routes
RegisterRoutes(app);
app.use(errorHandler);

var swaggerJson = JSON.parse(fs.readFileSync('./build-tsoa/swagger.json', 'utf-8'))
app.use('/api-docs', swaggerUi.serve, swaggerUi.setup(swaggerJson));

function getEnvVar(key: string, defaultValue: string): string {
    const value = process.env[key];
    if (typeof value === 'undefined' || value === null) {
        return defaultValue;
    }
    return value;
}

/** Constants */

/** Express App */
const port = Number(getEnvVar('PORT', '3500'));
const server = app.listen(port, '0.0.0.0', () => { //Force IPv4
    console.log('swAPI is listening on localhost', port);
    Logger.log('silly', `wAPI is starting... Listening on port ${port}`);
});

