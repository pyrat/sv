import dotenv from 'dotenv';
import jwt, { Secret } from 'jsonwebtoken';
import { Request, Response, NextFunction } from 'express';
dotenv.config();

/** Import Errors */
/** Import error definition */
import Errors, {ErrorTypes} from '../internal-ts/errors.esm.js';

interface User {
    username: string;
}

class JWT{
    private secret: Secret;

    constructor(){
        this.secret = process.env.TOKEN_SECRET as Secret;
        this.authenticateToken = this.authenticateToken.bind(this);
        console.log(`JWT : TOKEN_SECRET: ${this.secret}`)
    }

    generateAccessToken(username: User): string {
        return jwt.sign(username, this.secret);
    }

    authenticateToken(req: Request, res: Response, next: NextFunction){
        const authHeader = req.headers['authorization'];
        const token = authHeader && authHeader.split(' ')[1]; // Isolate "BEARER <JSON_WEB_TOKEN>" by " " to keep the web token only.

        if(token === null || typeof token === 'undefined') {
            console.log(`Missing bearer token in authenticateToken method`);
            const errResp = Errors[ErrorTypes.MISSING_BEARER_TOKEN];
            return res.status(errResp.status).json({message: errResp.message});
        }
        
        jwt.verify(token, this.secret, (err, user) => {
            const errResp = Errors[ErrorTypes.INVALID_BEARER_TOKEN];
            if(err) {
                console.log(`jwt.verify error : ${err}`);
                return res.status(errResp.status).json({message: errResp.message});
            }

            console.log(`jwt.verify user : ${JSON.stringify(user)} successfully authenticated`);
            (req as any).user = user;
            next();
        });
    }
}

export default (new JWT());
