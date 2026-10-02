import rpio from 'rpio';
import Logger from './logger.esm.js';
import { openPromisified } from 'i2c-bus';
import ADS1115 from 'ads1115';
import LS7366 from './LS7366.esm.js';
import { Motor } from './interfaces.esm.js'


class MotorDriver  implements Motor{
    private debug: boolean;
    private motorPosition: string;
    private defaultSpeed: number;
    private currentSpeed: number;
    private currentDirection: number;
    private currentPWM: number;
    private pwmPin: number;
    private directionPin: number;
    private pwmRange: number;
    private ADSAddress: number;
    private ADSChannel: number;
    private ADS1115: typeof ADS1115;
    private ADSMaxVoltage: number;
    private ADSResolution: number;
    private voltageOffset: number;
    private voltageToAmpereRatio: number;
    private bus: number;
    private LS7366: LS7366;
    private interval: NodeJS.Timeout;

    constructor(pwmPin: number, motorPosition: string = ''){
        this.debug = false;
        this.motorPosition = ""
        this.interval = setTimeout(() => {}, 0);
       
       //** ---------------------------------------------------- General Configuration ----------------------------------------------------*/
       //Configure motorPosition based on PWM pins. Can be forced by supplying motorPosition in constructor
       if(motorPosition === '' || motorPosition === undefined) pwmPin === 31 ? this.motorPosition = 'right' : this.motorPosition = 'left'; 
       if(motorPosition !== '' && motorPosition !== undefined) this.motorPosition = motorPosition;

        this.defaultSpeed = 70;                                     // 70% speed
        this.currentSpeed = 0;                                      // Current winch speed
        this.currentDirection = 0;                                  // Current direction selection
        //** ---------------------------------------------------- General Configuration ----------------------------------------------------*/

        //** ------------------------------------------------------ RPIO Configuration -----------------------------------------------------*/
        this.currentPWM = 0;                                        //Keep track of current PWM output.
        this.pwmPin = pwmPin;                                       //The pin at where the PWM output for the motor driver is connected.
        this.directionPin = this.motorPosition === 'right' ? 58:56; //The pin at where the direction is set for the motor driver. LOW = CW, HIGH = CCW
        this.pwmRange = 1024;                                       //PWM Range - 0 - 1024
        rpio.open(this.pwmPin, rpio.PWM);                           //Set pwmPin to PWM output for motor control.
        rpio.pwmSetClockDivider(2);                                 //Set the PWM frequency correlating to the RPi PWM frequency of 19.2MHz
        rpio.pwmSetRange(this.pwmPin, this.pwmRange);               //19.2e6 / 2 / 1024 --> RPi PWM Clock Frequency / pwmClockDivider / pwmRange
        rpio.open(this.directionPin, rpio.OUTPUT);                  //Direction pin choose direction of the winch motor
        rpio.write(this.directionPin, 0);                           //Set LOW by default
        //** ------------------------------------------------------ RPIO Configuration -----------------------------------------------------*/

        //** -------------------------------------------------- ADS1115 ADC configuration --------------------------------------------------*/
        /*   Gain Setting       Full Scale Range      Resolution (1 LSB)
            |------------|----|------------------|-------------------------|
            |     2/3    |    |     ±6411mV      |         187.6µV         |
            |      1     |    |     ±4096mV      |         125µV           |
            |      2     |    |     ±2048mV      |         62.5µV          |
            |      4     |    |     ±1024mV      |         31.25µV         |
            |      8     |    |     ±512mV       |         15.625µV        |
            |     16     |    |     ±256mV       |         7.8125µV        |
            |--------------------------------------------------------------|
        */
        this.ADSAddress = 0x48;                                     //I2C address of ADS1115 chip
        this.ADSChannel = this.motorPosition === 'right' ? 0 : 1;   //ADC channel on ADS1115 chip for right motor is 0, left is 1
        this.ADS1115 = ADS1115;                                     //ADS1115 ADC library
        this.ADSMaxVoltage = 0.256                                  //Max voltage ADS1115 can record when set to gain level 16 (+/- 256mV (0.256V))
        this.ADSResolution = 32768                                  //ADS1115 has a 16-bit resolution.
        this.voltageOffset = 0.029;                                     //Initialize voltage offset. Will be recorded when the class initalizes to get this sessions voltage offset.
        this.voltageToAmpereRatio = 0.020;                          //ADS1115 will read 20mV for every ampere current the motor draws.
        this.bus = 6;
        //Set ADS1115 gain to mode 16 i.e +/- 256mV
        openPromisified(this.bus).then(async (bus) => {
            const ads1115 = await this.ADS1115(bus);
            ads1115.gain = 16;                                      // Set gain to +/- 256mV (0.256V)
            bus.close().catch(console.log);
        });

        //** -------------------------------------------------- ADS1115 ADC configuration --------------------------------------------------*/

        //** ------------------------------------------------- LS7366 Counter configuration ------------------------------------------------*/
        this.LS7366 = new LS7366(this.motorPosition === 'right' ? 1 : 0); //If right motor, use Chip Select 0, Chip Select 1 for left motor.

        //** ------------------------------------------------- LS7366 Counter configuration ------------------------------------------------*
    }

    /**
     * Move the winch payload in the direction of directions object.
     * @param {object} directions 
     */
    move(directions: { cw?: boolean | number; ccw?: boolean | number ; }): Promise<void>{
        return new Promise((resolve, reject) => {
            const { cw, ccw } = directions;
            this.currentSpeed = this.defaultSpeed;
            if(cw && typeof(cw) !== 'boolean') this.currentSpeed = cw;                //If cw/ccw isn't a boolean, take the number from cw/ccw and set it to speed in %.
            if(ccw && typeof(ccw) !== 'boolean') this.currentSpeed = ccw;              //If cw/ccw isn't a boolean, take the number from cw/ccw and set it to speed in %.
            if(ccw) this.currentSpeed = -this.currentSpeed;
            let pwm = Math.floor(1024 * (Math.abs(this.currentSpeed) / 100));   // Max PWM * (ABSOLUTE speed in percent / 100) = PWM output. Math.floor to remove decimals.

            if(this.currentSpeed > 0){                                          //Means it's a positive number, set direction
                if(this.currentDirection !== 1){
                    // this.ramp()
                }
                this.selectDirection(true);
                this.debug && console.log('set direction false');
            }else{                                                              //Negative number, set direction.
                this.selectDirection(false);
                this.debug && console.log('set direction true');
            }

            this.debug && console.log({motor: this.motorPosition, cw, ccw, speed: this.currentSpeed, pwm, direction: rpio.read(this.directionPin)});
            if(cw == false && ccw == false) this.setPWM(0);
            else this.setPWM(pwm);
            resolve();
        });
    }  

    /**
     * Ramp speed to 0. Specify speed as a percent per tick. Defaults to 5%/tick. Specify tick as a value in ms. Defaults to 250ms
     * @param {number} speed 
     * @param {number} tick 
     */
    ramp(speed: number = 5, tick: number = 100): Promise<void>{
        console.log(`Ramping from PWM: ${this.currentPWM} to PWM: 0. Speed: ${speed}%/${tick}ms`);
        return new Promise((resolve, reject) => {
            let initialPWM = this.currentPWM;
            let currentPWM = initialPWM;

            const interval = setInterval(() => {
                currentPWM = Math.floor(currentPWM - ((initialPWM / 100) * speed));
                if(currentPWM < 0){
                    this.setPWM(0);
                    console.log({initialPWM, currentPWM})
                    clearInterval(interval);
                } else { 
                    console.log({initialPWM, currentPWM})
                    this.setPWM(currentPWM); 
                }
            }, tick);
        });
    }

    /**
     * Select direction of the motor. False means GPIO LOW, direction CW. True means set the GPIO HIGH which sets the direction CCW.
     * @param {boolean} state 
     * @returns Error if not succesfull
     */
    private selectDirection(state: boolean): string | void {
        const dir = state === true ? 1 : 0;
        this.currentDirection = dir;
        rpio.write(this.directionPin, dir);
    }

    /**
     * Set motor PWM
     * @param {number} PWM 
     */
    private setPWM(PWM: number): void{
        if(PWM > this.pwmRange) throw new Error(`PWM value ${PWM} is higher than max PWM range 0-${this.pwmRange}`);
        rpio.pwmSetData(this.pwmPin, PWM);
    }

    /**
     * Check current draw from motor.
     * @returns 
     */
    private checkCurrent(): void{
        // // return new Promise((resolve, reject) =>{
        // //     this.i2c.openPromisified(6).then(async (bus) =>{
        // //         const ads1115 = await this.ADS1115(bus);
        // //         let value = await ads1115.measure(`${this.ADSChannel}+GND`);
        // //         const current = this.convertCurrent(value);
        // //         console.log(`${this.motorPosition} motor: ${current}A`);
        // //         return resolve(current);
        // //         // Logger.log('info', `${value} ${mV.toFixed(4)} ${a.toFixed(4)} ${this.currentPWM}`);
        // //     });
        // // });

        openPromisified(6).then(async (bus) =>{
            const ads1115 = await this.ADS1115(bus);
            ads1115.gain = 16;

            const cb = async () => {
                let value = await ads1115.measure('0+GND');
                let maxVoltage = 0.256;
                const vPerBit = maxVoltage / 32768;
                const mV = (value * vPerBit) - this.voltageOffset;
                const a = (mV - 0) / 0.020;
                
                console.log(value, mV.toFixed(4), a.toFixed(4), this.currentPWM);
                Logger.log('info', `${value} ${mV.toFixed(4)} ${a.toFixed(4)} ${this.currentPWM}`)
            }

            this.interval = setInterval(cb, 250);
        });
    }

    /**
     * 
     * @param {number} value - ADC value to convert.
     * @returns Motor current draw in Ampere.
     */
    private convertCurrent(value: number): number{
        return (((this.ADSMaxVoltage / this.ADSResolution) * value) - this.voltageOffset) / this.voltageToAmpereRatio; 
    }
}




export default MotorDriver;

// const m = new MotorDriver(31);
// m.checkCurrent().then(console.log).catch(console.log);
// setInterval(() => {
//     console.log(m.LS7366.readCounter());
// }, 500);
// m.ramp();
// m.move({
//     up: 5
// })
