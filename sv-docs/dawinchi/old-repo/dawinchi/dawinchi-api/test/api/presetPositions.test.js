const dotenv = require('dotenv');
const request = require('supertest');
dotenv.config();

const basePath = process.env.BASE_PATH;
const baseURL = process.env.BASE_URL;

const newPreset = {
    presetName: "[Automated API Tester]: Winch Position Preset",
    xPosition: "30",
    yPosition: "40",
    uuid: "uuidv4"
};

describe(`Test ${baseURL}/${basePath}/presetPositions`, () =>{ 
    it('[GET] should return 200 and all presets, if any', async () =>{
        const response = await request(baseURL)
        .get(`/${basePath}/presetPositions`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT);
        expect(response.statusCode).toBe(200);
    });

    it('[POST] should return 200 and add a new preset', async () => {
        const response = await request(baseURL)
        .post(`/${basePath}/presetPositions`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        .send(newPreset);
        
        expect(response.statusCode).toBe(200);
        expect(response.body.presetName).toBe(newPreset.presetName);
        expect(response.body.xPosition).toBe(newPreset.xPosition);
        expect(response.body.yPosition).toBe(newPreset.yPosition);
        expect(response.body.uuid).toBe(newPreset.uuid);
    });

    it('[GET] should return 200 and return the previously added preset based on presetName', async () => {
        const response = await request(baseURL)
        .get(`/${basePath}/presetPositions/${newPreset.presetName}`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT);

        expect(response.statusCode).toBe(200);
        expect(response.body.presetName).toBe(newPreset.presetName);
        expect(response.body.xPosition).toBe(newPreset.xPosition);
        expect(response.body.yPosition).toBe(newPreset.yPosition);
        expect(response.body.uuid).toBe(newPreset.uuid);
    });

    it('[GET] should return 200 and return the previously added preset based on uuid', async () => {
        const response = await request(baseURL)
        .get(`/${basePath}/presetPositions/${newPreset.uuid}`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT);

        expect(response.statusCode).toBe(200);
        expect(response.body.presetName).toBe(newPreset.presetName);
        expect(response.body.xPosition).toBe(newPreset.xPosition);
        expect(response.body.yPosition).toBe(newPreset.yPosition);
        expect(response.body.uuid).toBe(newPreset.uuid);
    });

    it('[PUT] should return 200 and edit the previously added preset', async () => {
        const oldUUID = newPreset.uuid;
        newPreset.presetName = "[Automated API Tester]: Edited Winch Position Preset";
        newPreset.xPosition = "10";
        newPreset.yPosition = "20";
        newPreset.uuid = "uuidv4Edited";
        const response = await request(baseURL)
        .put(`/${basePath}/presetPositions/${oldUUID}`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        .send(newPreset);

        expect(response.statusCode).toBe(200);
        expect(response.body.presetName).toBe(newPreset.presetName);
        expect(response.body.xPosition).toBe(newPreset.xPosition);
        expect(response.body.yPosition).toBe(newPreset.yPosition);
        expect(response.body.uuid).toBe(newPreset.uuid);
    });

    it('[DELETE] should return 200 and delete the previously added preset', async () => {
        const response = await request(baseURL)
        .delete(`/${basePath}/presetPositions/${newPreset.uuid}`,)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT);

        expect(response.statusCode).toBe(200);
        expect(response.body.success).toBe(true);
    });

    it('[GET] should return 400 because the preset doesn\'t exist anymore', async () => {
        const response = await request(baseURL)
        .delete(`/${basePath}/presetPositions/${newPreset.uuid}`,)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT);

        expect(response.statusCode).toBe(400);
        expect(response.body.success).toBe(false);
        expect(response.body.faultCode).toBe(104);
    });

    it('[POST] should return 400 because the preset doesn\'t contain a preset name', async () => {
        const response = await request(baseURL)
        .post(`/${basePath}/presetPositions`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        .send({xPosition: newPreset.xPosition, yPosition: newPreset.yPosition, uuid: newPreset.uuid});
        
        expect(response.statusCode).toBe(400);
        expect(response.body.success).toBe(false);
        expect(response.body.faultCode).toBe(105);
    });

    it('[POST] should return 400 because the preset doesn\'t contain X position', async () => {
        const response = await request(baseURL)
        .post(`/${basePath}/presetPositions`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        .send({presetName: newPreset.presetName, yPosition: newPreset.yPosition, uuid: newPreset.uuid});
        
        expect(response.statusCode).toBe(400);
        expect(response.body.success).toBe(false);
        expect(response.body.faultCode).toBe(106);
    });

    it('[POST] should return 400 because the preset doesn\'t contain Y position', async () => {
        const response = await request(baseURL)
        .post(`/${basePath}/presetPositions`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        .send({presetName: newPreset.presetName, xPosition: newPreset.xPosition, uuid: newPreset.uuid});
        
        expect(response.statusCode).toBe(400);
        expect(response.body.success).toBe(false);
        expect(response.body.faultCode).toBe(107);
    });

    it('[POST] should return 200 and add a new preset without xPosition and yPosition', async () => {
        delete newPreset.xPosition;
        delete newPreset.yPosition;
        const response = await request(baseURL)
        .post(`/${basePath}/presetPositions`)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT)
        .send(newPreset);
        
        expect(response.statusCode).toBe(200);
        expect(response.body.presetName).toBe(newPreset.presetName);
        expect(response.body.uuid).toBe(newPreset.uuid);
    });

    it('[DELETE] should return 200 and delete the previously added preset', async () => {
        const response = await request(baseURL)
        .delete(`/${basePath}/presetPositions/${newPreset.uuid}`,)
        .set('Authorization', 'bearer ' + process.env.TOKEN_CLIENT);

        expect(response.statusCode).toBe(200);
        expect(response.body.success).toBe(true);
    });
    
});