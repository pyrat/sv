const dotenv = require('dotenv');
const request = require('supertest');
dotenv.config();

const basePath = process.env.BASE_PATH;
const baseURL = process.env.BASE_URL;

describe(`Test ${baseURL}/${basePath}/configuration/encoders`, () =>{
    it('[GET] should return 200 and current encoder positions', async () => {
        const response = await request(baseURL)
        .get(`/${basePath}/configuration/encoder`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        expect(response.statusCode).toBe(200);
        expect(response.body.encoderPosition !== undefined);
    });

    it('[GET] should return 200 and current encoder positions in meters', async () => {
        const response = await request(baseURL)
        .get(`/${basePath}/configuration/encoder`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        expect(response.statusCode).toBe(200);
        expect(response.body.encoderPositions !== undefined);
        expect(response.body.encoderPositionsMeter !== undefined);
    });
});