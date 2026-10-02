import express from 'express';
import cors from 'cors';
import bodyParser from 'body-parser';
import {errorHandler} from "./errhandler.esm.js";



class API {
    public debug: boolean;
    public port: number;
    public basePath: string;
    public app: express.Application;

    constructor() {
        console.log(`TOKEN_USER: ${process.env.TOKEN_USER}`);
        /** Initialize App */
        this.debug = true;
        this.port = 3500;
        this.basePath = '/api/v1';

        this.app = express();
        this.app.use(cors());
        this.app.use(bodyParser.json());

    }
}

export default new API();