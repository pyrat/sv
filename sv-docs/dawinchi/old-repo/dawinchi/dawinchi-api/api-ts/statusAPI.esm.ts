import API from './api.esm.js';
const {app, basePath} = API;
app.get(`${basePath}/`, async (req, res) =>{
    res.sendStatus(501);
})