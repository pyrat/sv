const dotenv = require('dotenv');
const request = require('supertest');
dotenv.config();

const basePath = process.env.BASE_PATH;
const baseURL = process.env.BASE_URL;

describe(`Test ${baseURL}/${basePath}/move`, () =>{
    it('[POST] should return 200 and move UP', async () => {
        const response = await request(baseURL)
        .post(`/${basePath}/move`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        .send({
            directions: {
                up: true
            }
        });
        expect(response.statusCode).toBe(200);
    });

    it('[POST] should return 200 and move DOWN', async () => {
        const response = await request(baseURL)
        .post(`/${basePath}/move`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        .send({
            directions: {
                down: true
            }
        });
        expect(response.statusCode).toBe(200);
    });

    it('[POST] should return 200 and move IN', async () => {
        const response = await request(baseURL)
        .post(`/${basePath}/move`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        .send({
            directions: {
                in: true
            }
        });
        expect(response.statusCode).toBe(200);
    });

    it('[POST] should return 200 and move OUT', async () => {
        const response = await request(baseURL)
        .post(`/${basePath}/move`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        .send({
            directions: {
                out: true
            }
        });
        expect(response.statusCode).toBe(200);
    });

    it('[POST] should return 200 and move UP@75% speed', async () => {
        const response = await request(baseURL)
        .post(`/${basePath}/move`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        .send({
            directions: {
                up: 75
            }
        });
        expect(response.statusCode).toBe(200);
        expect(response.body.directions.up).toBe(75);
    });

    it('[POST] should return 200 and move DOWN@75% speed', async () => {
        const response = await request(baseURL)
        .post(`/${basePath}/move`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        .send({
            directions: {
                down: 75
            }
        });
        expect(response.statusCode).toBe(200);
        expect(response.body.directions.down).toBe(75);
    });

    it('[POST] should return 200 and move IN@75% speed', async () => {
        const response = await request(baseURL)
        .post(`/${basePath}/move`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        .send({
            directions: {
                in: 75
            }
        });
        expect(response.statusCode).toBe(200);
        expect(response.body.directions.in).toBe(75);
    });

    it('[POST] should return 200 and move OUT@75% speed', async () => {
        const response = await request(baseURL)
        .post(`/${basePath}/move`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        .send({
            directions: {
                out: 75
            }
        });
        expect(response.statusCode).toBe(200);
        expect(response.body.directions.out).toBe(75);
    });

    it('[POST] should return 400 because of out of bounds speed', async () => {
        const response = await request(baseURL)
        .post(`/${basePath}/move`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        .send({
            directions: {
                out: 150
            }
        });
        expect(response.statusCode).toBe(400);
        expect(response.body.success).toBe(false);
        expect(response.body.faultCode).toBe(102);
    });

    it('[POST] should return 400 because up&down are called simultaneously', async () => {
        const response = await request(baseURL)
        .post(`/${basePath}/move`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        .send({
            directions: {
                up: 50,
                down: true
            }
        });
        expect(response.statusCode).toBe(400);
        expect(response.body.success).toBe(false);
        expect(response.body.faultCode).toBe(100);
    });

    it('[POST] should return 400 because in&out are called simultaneously', async () => {
        const response = await request(baseURL)
        .post(`/${basePath}/move`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        .send({
            directions: {
                in: 50,
                out: true
            }
        });
        expect(response.statusCode).toBe(400);
        expect(response.body.success).toBe(false);
        expect(response.body.faultCode).toBe(101);
    });
});