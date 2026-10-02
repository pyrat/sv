/** Import conf */
import confmodule from 'conf';
const conf = confmodule.default;

/** Import Logger */
import Logger from './logger.esm.js';

/** Import UUIDv4 */
import { v4 as uuidv4 } from 'uuid';

interface Preset {
    presetName: string;
    xPosition: string;
    yPosition: string;
    uuid?: string;
}

class PresetPositions {
    private debug: boolean;
    private presetPositions: any;

    constructor(){
        this.debug = true;
        this.presetPositions = new conf({
            configName: 'presetPositions',
            cwd: 'config',
            defaults:{
                positions:{}
            },
            watch: true
        });
        console.log(this.presetPositions.path);
    }

    getAllPresets(): Preset[]{
        return this._getPresets();
    }

    getPreset(presetIdentifier: string): Promise<Preset>{
        return new Promise((resolve, reject) => {
            const presets = this._getPresets();
            presets.forEach((value: Preset) => {
                if(presetIdentifier === value.presetName || presetIdentifier === value.uuid) return resolve(value);
            });
            return reject({fault: 104})
        });
    }

    addPreset(sanitizedPreset: Preset): Promise<Preset>{
        return new Promise((resolve, reject) => {
            if(!sanitizedPreset.uuid) sanitizedPreset.uuid = uuidv4();
            let presets = this._getPresets();
            let conflict = false;
            presets.forEach((preset: Preset, index: number) => {
                if(preset.presetName === sanitizedPreset.presetName){
                    presets[index] = sanitizedPreset;
                    this.presetPositions.set('positions', presets);
                    conflict = true;
                    return resolve(sanitizedPreset);
                }
            });
    
            if(!conflict) this.presetPositions.set('positions', [...presets, sanitizedPreset]);
            Logger.log('debug', `[Preset Positions]: Added new preset ${sanitizedPreset.presetName} at X: ${sanitizedPreset.xPosition} / Y: ${sanitizedPreset.yPosition}`);
            return resolve(sanitizedPreset);
        })
    }

    deletePreset(presetIdentifier: string): Promise<void>{
        return new Promise((resolve, reject) => {
            const presets = this._getPresets();
            presets.forEach((preset: Preset, index: number) => {
                if(preset.presetName === presetIdentifier || preset.uuid === presetIdentifier){
                    presets.splice(index, 1);
                    this.presetPositions.set('positions', presets);
                    return resolve();
                }
            });
            return reject({fault: 104});
        })
    }

    updatePreset(presetIdentifier: string, presetObject: Partial<Preset>): Promise<Preset | undefined>{
        return new Promise((resolve, reject) => {
            const presets = this._getPresets();
            presets.forEach((preset: Preset, index: number) => {
                if(preset.presetName === presetIdentifier || preset.uuid === presetIdentifier){
                    presets[index] = { ...presets[index], ...presetObject };
                    this.presetPositions.set('positions', presets);
                    return resolve(presets[index]);
                }
            });
            return reject({fault: 104});
        })
    }

    private _getPresets(): Preset[]{
        return this.presetPositions.get('positions');
    }
}

export default(new PresetPositions());

// const test = new PresetPositions();
// test.getPreset('Test').then(console.log);
// test.updatePreset('Test Preset 5', {
//     yPosition: "150",
//     xPosition: "123",
//     presetName: "Test Preset 55"
// }).catch(() => console.log('cannot find Test Preset 5'))

// test.deletePreset('Test Preset 1').then(() => console.log('deleted')).catch(err => console.log('error'));
