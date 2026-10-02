import { Get, Controller, Body, Post, Route, Response, Request, Middlewares, ValidateError } from 'tsoa';
import {ErrorResponse} from '../../internal-ts/errors.esm.js';
import JWT from '../../internal-ts/token.esm.js';
import Logger from '../../internal-ts/logger.esm.js';

interface AuthRequest {
    username: string;
}

interface AuthResponse {
    token: string;
}

interface AuthStatus {
    authenticated: boolean;
}

@Route('/api/v1/auth')
export class AuthController extends Controller{

    @Response<AuthResponse>(200, 'Success')
    @Response<ErrorResponse>(401, 'Error') // This is just a guess. Replace with your actual Error model and status code.
    @Response<ErrorResponse>(500, 'Error') // This is just a guess. Replace with your actual Error model and status code.
    @Post('authorize') // The method decorator matches the HTTP verb, e.g., @Post, @Get, @Put, etc.
    async authorize(@Body() authReq: AuthRequest): Promise<AuthResponse> {
        console.log(`authorize is called with ${authReq.username}`);
        console.log(authReq);
        if (process.env.TOKEN_USER !== authReq.username){
            this.setStatus(401);
            throw {
                message: 'Invalid username',
            }
        }
        const token = JWT.generateAccessToken({username: authReq.username});
        Logger.log('http', `User ${authReq.username} authenticated!`);
        return {
            token
        }
    }

    @Response<AuthStatus>(200, 'Success')
    @Response<ErrorResponse>(401, 'Error') // This is just a guess. Replace with your actual Error model and status code.
    @Get('status') // The method decorator matches the HTTP verb, e.g., @Post, @Get, @Put, etc.
    @Middlewares([JWT.authenticateToken])
    async status(): Promise<AuthStatus>{
        return {
            authenticated: true
        }
    }
}