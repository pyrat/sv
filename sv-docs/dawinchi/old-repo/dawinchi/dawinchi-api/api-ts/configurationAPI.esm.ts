/** Import API base declaration */
import API from './api.esm.js';

/** Import error definition */
import Errors from '../internal-ts/errors.esm.js';

/** Import JSONWebToken wrapper class */
import JWT from '../internal-ts/token.esm.js';

/** Import winch control class */
import Control from '../internal-ts/control.esm.js';

/** Import logger */
import Logger from '../internal-ts/logger.esm.js';

/** Destructure API export */
const { app, basePath } = API;

app.get(`${basePath}/configuration/encoders`, JWT.authenticateToken, (req, res) =>{
    // if(req.body.directions === undefined || Object.keys(req.body.directions).length === 0) 
    //     return res.status(Errors[109].status).json({success: false, message: Errors[109].message, faultCode: 109});
    // const directions = req.body.directions;
    // Logger.log('debug', `[MOVE]: Received move command from ${req.socket.localAddress}`, directions);
    // Control.move(directions).then(() =>{
    //     res.status(200).json({success: true, directions});
    // }).catch(err =>{
    //     console.log(err);
    //     Logger.log('error', `[MOVE]: Move command from ${req.socket.localAddress} failed`, { statusCode: Errors[err.fault].status, msg: Errors[err.fault].message, faultCode: err.fault });
    //     res.status(Errors[err.fault].status).json({success: false, message: Errors[err.fault].message, faultCode: err.fault});
    // });
});