import {ValidateError} from 'tsoa';
import {Request, Response, NextFunction} from 'express';
import {ErrorResponse, ErrorResponseWithcode} from "../internal-ts/errors.esm.js";

export function errorHandler(
    err: unknown,
    req: Request,
    res: Response,
    next: NextFunction
): Response | void {
    console.log(`[ERROR_HANDLER] is called : ${err}`)
    if (err instanceof ValidateError) {
        console.warn(`Caught Validation Error for ${req.path}:`, err.fields);
        return res.status(422).json({
            message: "Validation Failed",
            details: err?.fields,
        });
    }

    if (err instanceof ErrorResponseWithcode) {
        return res.status(err.statusCode).json({
            message: err.message,
        } as ErrorResponse);
    }

    if (err instanceof Error) {
        return res.status(500).json({
            message: "Internal Server Error",
        } as ErrorResponse);
    }

    next();
}