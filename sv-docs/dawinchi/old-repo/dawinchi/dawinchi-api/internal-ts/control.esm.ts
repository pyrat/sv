// import MotorDriver from './motorDriver.esm.js';
import {Motor, MockMotor} from './interfaces.esm.js';

export interface Directions {
    up?: boolean,
    down?: boolean,
    wind?: boolean,
    unwind?: boolean,
    speed?: number
}

export interface Control {
    move(directions: Directions): Promise<void>;

    moveTo(xPosition: number, yPosition: number): Promise<void>;
}

function delay(ms: number) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

export class ControlHandler implements Control {
    debug: boolean;
    defaultSpeed: number;
    maxAbsoluteSpeed: number;
    rightMotorPWMPin: number;
    leftMotorPWMPin: number;
    rightMotor: Motor;
    leftMotor: Motor;

    constructor() {
        this.debug = true;
        this.defaultSpeed = 100;
        this.maxAbsoluteSpeed = 100;

        this.rightMotorPWMPin = 31;
        this.leftMotorPWMPin = 28;

        // this.rightMotor = new MotorDriver(this.rightMotorPWMPin);
        // this.leftMotor  = new MotorDriver(this.leftMotorPWMPin);


        this.rightMotor = new MockMotor();
        this.leftMotor = new MockMotor();

        // setInterval(() => {
        //     if(this.rightMotor.currentSpeed > 0) console.log(this.rightMotor.checkCurrent());
        //     if(this.leftMotor.currentSpeed > 0) console.log(this.leftMotor.checkCurrent());
        // }, 1000);
    }

    /**
     * Receives directions object and commands the winch depending on what the object contains
     * If a direction is set to true, it will use the default speed, if it's set to a number, the direction will move with the specified speed.
     * @example
     * {
     *      up: true,
     *      in: 50
     * }
     * @param {object} directions
     */
    move(directions: Directions): Promise<void> {
        console.log(`Moving winch with directions: ${JSON.stringify(directions)} => type : ${typeof (directions)}`);
        return new Promise((resolve, reject) => {
            console.log(directions);
            if ((directions.up !== undefined && directions.up) && (directions.down !== undefined && directions.down)) {
                return reject({fault: 100}); //Cannot move both up and down!
            }
            if ((directions.wind !== undefined && directions.wind) && (directions.unwind !== undefined && directions.unwind)) {
                return reject({fault: 101});
            } //Cannot move both in and out!

            if (directions.speed === undefined) directions.speed = this.defaultSpeed;

            if (directions.speed > this.maxAbsoluteSpeed) return reject({fault: 102});

            //!Do motor movement logic & logging here
            if (directions.up !== undefined || directions.down !== undefined) {
                this.rightMotor.move({cw: directions.up, ccw: directions.down}).then(console.log).catch(console.log);
                console.log('Right motor')
            }
            if (directions.wind !== undefined || directions.unwind !== undefined) {
                this.leftMotor.move({cw: directions.wind, ccw: directions.unwind}).then(console.log).catch(console.log);
                console.log('Left motor')
            }
            return resolve();
        });
    }

    async moveTo(xPosition: number, yPosition: number): Promise<void> {
        const startTime = new Date().getTime();
        console.log(`Moving winch to position: ${xPosition}, ${yPosition}`)
        if (xPosition === undefined || xPosition === null) throw {fault: 106};
        if (yPosition === undefined || yPosition === null) throw {fault: 107};
        // do some random sleep to simulate movement
        await delay(10000);
        const endTime = new Date().getTime();
        console.log(`Moved winch to position: ${xPosition}, ${yPosition} it took ${endTime - startTime}`)
    }

    stop(): Promise<void> {
        return new Promise((resolve, reject) => {
            this.rightMotor.ramp(10, 10).then(console.log).catch(console.log);
            this.leftMotor.ramp(10, 10).then(console.log).catch(console.log);
            return resolve();
        });
    }
}

export default (new ControlHandler());
