/** Import API base declaration */
import API from './api.esm.js';

/** Import error definition */
import Errors from '../internal-ts/errors.esm.js';

/** Import JSONWebToken wrapper class */
import JWT from '../internal-ts/token.esm.js';

/** Import logger */
import Logger from '../internal-ts/logger.esm.js';

/** Import Preset Positions class */
import PresetPositions from '../internal-ts/presetPositions.esm.js';

/** Destructure API export */
const { app, basePath } = API;

//Get all presetPositions from winch
app.get(`${basePath}/presetPositions`, JWT.authenticateToken, (req, res) => {
    Logger.log('debug', `[Preset Positions]: Received GET all preset positions from ${req.socket.localAddress}`);
    res.status(200).json(PresetPositions.getAllPresets());
});

//Get a preset from winch based on UUID or preset name
app.get(`${basePath}/presetPositions/:identifier`, JWT.authenticateToken, (req, res) => {
    Logger.log('debug', `[Preset Positions]: Received GET preset ${req.params.identifier} from ${req.socket.localAddress}`);
    PresetPositions.getPreset(req.params.identifier).then(preset => {
        res.status(200).json(preset); //Send retrieved preset along with 200 OK.
    }).catch(err => {
        res.status(Errors[err.fault].status).json({success: false, message: Errors[err.fault].message, faultCode: err.fault});
    });    
});

//Create a new preset or overwrite an existing one
app.post(`${basePath}/presetPositions`, JWT.authenticateToken, (req, res) => {
    if(!req.body.presetName) return res.status(Errors[105].status).json({success: false, message: Errors[105].message, faultCode: 105}); //Provide presetName!
    if(!req.body.xPosition && req.body.yPosition) return res.status(Errors[106].status).json({success: false, message: Errors[106].message, faultCode: 106}) //Provide xPosition!
    if(!req.body.yPosition && req.body.xPosition) return res.status(Errors[107].status).json({success: false, message: Errors[107].message, faultCode: 107}) //Provide yPosition!
    if(!req.body.yPosition && !req.body.xPosition) return res.status(501).json({success: false, message: 'Automatically collecting winch position not implemented!'})
    const { presetName, xPosition, yPosition, uuid } = req.body;
    Logger.log('debug', `[Preset Positions]: Received POST preset ${presetName} at X: ${xPosition}, Y: ${yPosition}`);
    PresetPositions.addPreset({ presetName, xPosition, yPosition, uuid })
    .then((preset) => res.status(200).json(preset))
    .catch(() => res.status(500).json({success: false, message: "Internal server error"}));
});

//Update an existing preset
app.put(`${basePath}/presetPositions/:identifier`, JWT.authenticateToken, (req, res) => {
    if(!req.body.presetName && !req.body.xPosition && !req.body.yPosition && !req.body.uuid) return res.status(Errors[108].status).json({success: false, message: Errors[108].message, faultCode: 108}); //Check if there's anything in the body of the request

    const { presetName, xPosition, yPosition, uuid } = req.body;
    Logger.log('debug', `[Preset Positions]: Received PUT, updating preset ${presetName} at X: ${xPosition}, Y: ${yPosition}`);
    PresetPositions.updatePreset(req.params.identifier, { presetName, xPosition, yPosition, uuid }).then((preset) => res.status(200).json(preset)).catch((err) => {
        res.status(Errors[err.fault].status).json({success: false, message: Errors[err.fault].message, faultCode: err.fault});
    });
});

//Delete an existing preset
app.delete(`${basePath}/presetPositions/:identifier`, JWT.authenticateToken, (req, res) => {
    Logger.log('debug', `[Preset Positions]: Received DELETE, DELETING preset ${req.params.identifier}.`);
    PresetPositions.deletePreset(req.params.identifier)
    .then(() => res.status(200).json({success: true, message: `Preset ${req.params.identifier} succesfully deleted.`}))
    .catch(err => res.status(Errors[err.fault].status).json({success: false, message: Errors[err.fault].message, faultCode: err.fault}));
});