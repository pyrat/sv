/** Import API base declaration */
import API from './api.esm.js';
/** Import JSONWebToken wrapper class */
import JWT from '../internal-ts/token.esm.js';

/** Import logger */
import Logger from '../internal-ts/logger.esm.js';

/** Destructure API export */
const { app, basePath } = API;
console.log(basePath);
app.post(`${basePath}/authorize`, (req, res) => {
    console.log(req.body);
    Logger.log('http', `User ${req.body.username} is trying to authenticate...`);
    if(process.env.TOKEN_USER !== req.body.username) return res.sendStatus(400);
    const token = JWT.generateAccessToken({username: req.body.username});
    res.status(200).json({token});
    Logger.log('http', `User ${req.body.username} authenticated!`);
});

app.get(`${basePath}/test`, JWT.authenticateToken, (req, res) => {
    res.sendStatus(200);
});