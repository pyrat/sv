import rpio from 'rpio';

class ButtonPanel {
    debug: boolean;
    homeButtonPin: number;
    upButtonPin: number;
    downButtonPin: number;
    inButtonPin: number;
    outButtonPin: number;
    lockLEDPin: number;
    errorLEDPin: number;

    home: boolean;
    up: boolean;
    down: boolean;
    in: boolean;
    out: boolean;

    constructor(){
        this.debug              = true;
        this.homeButtonPin      = 30; // GPIO6
        this.upButtonPin        = 55; // GPIO14
        this.downButtonPin      = 51; // GPIO15
        this.inButtonPin        = 29; // GPIO16
        this.outButtonPin       = 50; // GPIO17
        this.lockLEDPin         = 54; // GPIO4
        this.errorLEDPin        = 34; // GPIO5

        this.home               = false; // Home button pressed value
        this.up                 = false; // Up button pressed value
        this.down               = false; // Down button pressed value
        this.in                 = false; // In button pressed value
        this.out                = false; // Out button pressed value
        
        rpio.open(this.homeButtonPin, rpio.INPUT, rpio.PULL_UP);
        rpio.open(this.upButtonPin, rpio.INPUT, rpio.PULL_UP);
        rpio.open(this.downButtonPin, rpio.INPUT, rpio.PULL_UP);
        rpio.open(this.inButtonPin, rpio.INPUT, rpio.PULL_UP);
        rpio.open(this.outButtonPin, rpio.INPUT, rpio.PULL_UP);
        rpio.open(this.lockLEDPin, rpio.OUTPUT);
        rpio.open(this.errorLEDPin, rpio.OUTPUT);

        rpio.poll(this.homeButtonPin, (pin: number) => {
            console.log(`[RPIO Poll]: Home(pin ${this.homeButtonPin}): ${rpio.read(pin)}`);
            rpio.read(pin) === 1 ? this.home = false : this.home = true;
        }, rpio.POLL_BOTH);

        rpio.poll(this.upButtonPin, (pin: number) => {
            console.log(`[RPIO Poll]: Up(pin ${this.upButtonPin}): ${rpio.read(pin)}`);
            rpio.read(pin) === 1 ? this.up = false : this.up = true;
        }, rpio.POLL_BOTH);

        rpio.poll(this.downButtonPin, (pin: number) => {
            console.log(`[RPIO Poll]: Down(pin ${this.downButtonPin}): ${rpio.read(pin)}`);
            rpio.read(pin) === 1 ? this.down = false : this.down = true;
        }, rpio.POLL_BOTH);

        rpio.poll(this.inButtonPin, (pin: number) => {
            console.log(`[RPIO Poll]: In(pin ${this.inButtonPin}): ${rpio.read(pin)}`);
            rpio.read(pin) === 1 ? this.in = false : this.in = true;
        }, rpio.POLL_BOTH);

        rpio.poll(this.outButtonPin, (pin: number) => {
            console.log(`[RPIO Poll]: Out(pin ${this.outButtonPin}): ${rpio.read(pin)}`);
            rpio.read(pin) === 1 ? this.out = false : this.out = true;
        }, rpio.POLL_BOTH);

        // setInterval(() => {
        //     console.log({home: this.home, up: this.up, down: this.down, in: this.in, out: this.out});
        //     rpio.write(this.errorLEDPin, rpio.HIGH);
        //     rpio.write(this.lockLEDPin, rpio.HIGH);
        //     setTimeout(() => {
        //     rpio.write(this.errorLEDPin, rpio.LOW);
        //     rpio.write(this.lockLEDPin, rpio.LOW);
                
        //     }, 400);
        // }, 500);
    } 
}

export default (new ButtonPanel);
