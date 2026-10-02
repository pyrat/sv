export interface ErrorDetail {
    message: string;
    status: number;
}

export class ErrorResponse extends Error {
    code?: number;

    constructor(message: string, code?: number) {
        super(message);
        this.code = code;
    }
}

export enum ErrorTypes {
    INVALID_DIRECTION = 100,
    INVALID_WIND = 101,
    INVALID_SPEED = 102,
    INVALID_UUID = 103,
    INVALID_PRESET = 104,
    INVALID_PRESET_NAME = 105,
    INVALID_X_POSITION = 106,
    INVALID_Y_POSITION = 107,
    INVALID_PRESET_REQUEST = 108,
    INVALID_DIRECTION_REQUEST = 109,
    INVALID_MOVE_TO_REQUEST = 110,
    INVALID_REQUEST = 400,
    MISSING_BEARER_TOKEN = 401,
    INVALID_BEARER_TOKEN = 403,
    INTERNAL_SERVER_ERROR = 500,
    RESOURCE_NOT_FOUND = 404,
}

const errorDetails: { [key: number]: ErrorDetail } = {
    100: {
        message: "Cannot move both up and down simultaneously!",
        status: 400,
    },
    101: {
        message: "Cannot wind and unwind simultaneously!",
        status: 400,
    },
    102: {
        message: "Requested speed is out of bounds, valid range is 0-100",
        status: 400,
    },
    103: {
        message: "Please specify either UUID or preset name to retrieve a preset!",
        status: 400,
    },
    104: {
        message: "The preset requested doesn't exist",
        status: 400,
    },
    105: {
        message: "Please provide a preset name!",
        status: 400,
    },
    106: {
        message: "Please provide xPosition!",
        status: 400,
    },
    107: {
        message: "Please provide yPosition!",
        status: 400,
    },
    108: {
        message: "Please provide either a preset name, xPosition or yPosition!",
        status: 400,
    },
    109: {
        message: "Please provide a directions object with directions to the payload!",
        status: 400,
    },
    400: {
        message: "Bad request!",
        status: 400,
    },
    401: {
        message: "Please provide a bearer token with the request!",
        status: 401,
    },
    403: {
        message: "The bearer token is not valid!",
        status: 403,
    },
    404: {
        message: "The requested resource was not found!",
        status: 404,
    },
    500: {
        message: "Something went wrong on the server!",
        status: 500,
    }
};

export class ErrorResponseWithcode extends ErrorResponse {
    statusCode: number;

    constructor(code: number, message: string, statusCode: number) {
        super(message, code);
        this.statusCode = statusCode;
    }
}

export function getErrorResponse(errorCode: ErrorTypes, message?: string): ErrorResponseWithcode {
    let errDetail = errorDetails[errorCode];
    if (!errDetail) {
        errDetail = errorDetails[500];
    }

    return new ErrorResponseWithcode(errorCode, message || errDetail.message, errDetail.status);
}

export default errorDetails;
