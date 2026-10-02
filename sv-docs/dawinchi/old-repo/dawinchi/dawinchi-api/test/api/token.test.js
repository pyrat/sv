const request = require('supertest');
const dotenv = require('dotenv');
dotenv.config();

const basePath = process.env.BASE_PATH;
const baseURL = process.env.BASE_URL;

describe(`Test ${basePath}/authorize`, () => {
    it('[POST] should return 200', async () => {
        const body = {
            username: 'smartWinchAdmin'
        }
        const response = await request(baseURL).post(`/${basePath}/authorize`).send(body);
        expect(response.statusCode).toBe(200);
    });
});
