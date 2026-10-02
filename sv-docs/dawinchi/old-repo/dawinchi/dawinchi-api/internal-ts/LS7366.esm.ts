import rpio from 'rpio';

import confmodule from 'conf';
const conf = confmodule.default;

class LS7366 {
    configuration: any;
    chipSelect: number;
    direction: string;
    CLEAR_COUNTER: number;
    CLEAR_STATUS: number;
    READ_COUNTER: number;
    READ_STATUS: number;
    WRITE_MODE0: number;
    WRITE_MODE1: number;
    WRITE_DTR: number;
    LOAD_CNTR: number;
    QUADRATURE_COUNT_MODE: number;
    FOURBYTE_COUNTER: number;
    THREEBYTE_COUNTER: number;
    TWOBYTE_COUNTER: number;
    ONEBYTE_COUNTER: number;
    BYTE_MODE: number[];
    max_val: number;
    counterSize: number;

    constructor(CSX: number, BTMD: number = 4) {
        this.configuration = new conf({
            configName: 'LS7366',
            cwd: 'config',
            defaults:{
                rightCount: 0,
                leftCount: 0
            },
            watch: true
        });
        this.chipSelect = CSX;
        this.chipSelect === 0 ? this.direction = "right" : this.direction = "left"; // If chip select is pin 0, then it's the right motor.

        // Constants
        this.CLEAR_COUNTER = 0x20;
        this.CLEAR_STATUS = 0x30;
        this.READ_COUNTER = 0x60;
        this.READ_STATUS = 0x70;
        this.WRITE_MODE0 = 0x88;
        this.WRITE_MODE1 = 0x90;
        this.WRITE_DTR = 0x98;
        this.LOAD_CNTR = 0xE0;
        this.QUADRATURE_COUNT_MODE = 0x00;
        this.FOURBYTE_COUNTER = 0x00;
        this.THREEBYTE_COUNTER = 0x01;
        this.TWOBYTE_COUNTER = 0x02;
        this.ONEBYTE_COUNTER = 0x03;
        this.BYTE_MODE = [this.ONEBYTE_COUNTER, this.TWOBYTE_COUNTER, this.THREEBYTE_COUNTER, this.FOURBYTE_COUNTER];
        this.max_val = 4294967295;

        // Global Variables
        this.counterSize = BTMD // Default 4

        // SPI configuration
        rpio.spiBegin();
        rpio.spiChipSelect(this.chipSelect);
        rpio.spiSetClockDivider(248); /* Set SPI speed to ~1MHz */

        // SPI initial commands
        rpio.spiWrite(Buffer.from([this.WRITE_MODE0, 0x03]), 2);
        rpio.spiWrite(Buffer.from([this.WRITE_MODE1, 0x00]), 2);

        // Set counter value if it is defined in configuration
        if(this.get(`${this.direction}Count`) !== undefined) this.setLS7366Count(this.get(`${this.direction}Count`));

        // Periodically update the counter value in configuration
        setInterval(() => {
            this._set(`${this.direction}Count`, this.readCounter().toString());
        }, 500);
    }

    _set(key: string, value: string): void {
        this.configuration.set(key, value);
    }

    get(key: string): any {
        return this.configuration.get(key);
    }

    setLS7366Count(number: number): string | void {
        if(Math.abs(number) > this.max_val / 2) return 'Number is too high!';
        rpio.spiWrite(Buffer.from([this.WRITE_DTR, ...this.numberToByteArray(number)]), 5);
        rpio.spiWrite(Buffer.from([this.LOAD_CNTR]), 1);
    }

    clearCounter(): void { 
        rpio.spiWrite(Buffer.from([this.CLEAR_COUNTER]), 1);
    }

    clearStatus(): void {
        rpio.spiWrite(Buffer.from([this.CLEAR_STATUS]), 1);
    }

    readCounter(): number {
        rpio.spiChipSelect(this.chipSelect);
        const txBuffer = Buffer.from([this.READ_COUNTER, 0, 0, 0, 0]);
        const rxBuffer = Buffer.alloc(txBuffer.length);

        rpio.spiTransfer(txBuffer, rxBuffer, txBuffer.length);
        console.log(rxBuffer.join(' '));
        let enc = 0;
        enc = (enc << 8) + parseInt(rxBuffer[1].toString());
        enc = (enc << 8) + parseInt(rxBuffer[2].toString());
        enc = (enc << 8) + parseInt(rxBuffer[3].toString());
        enc = (enc << 8) + parseInt(rxBuffer[4].toString());
        return enc;
    }

    numberToByteArray(number: number): Uint8Array {
        const arr = new ArrayBuffer(4);
        const view = new DataView(arr);
        view.setUint32(0, number, false);
        return new Uint8Array(arr);
    }
}

export default LS7366;
