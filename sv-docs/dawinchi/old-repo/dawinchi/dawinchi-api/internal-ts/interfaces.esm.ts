export interface Motor {
    move(directions: { cw?: boolean | number; ccw?: boolean | number ; }): Promise<void>;
    ramp(speed: number, tick: number): Promise<void>;
}

class MockMotor implements Motor {
    private generateDelay(): number {
      return Math.floor(Math.random() * 5 + 1) * 1000; // Generate a delay between 1-5 seconds
    }
  
    public async move(directions: { cw?: boolean | number; ccw?: boolean | number; }): Promise<void> {
      return new Promise<void>((resolve, reject) => {
        console.log(`Moving motor in direction: ${JSON.stringify(directions)}`);
        setTimeout(() => {
          console.log(`Moved motor in direction: ${JSON.stringify(directions)}`);
          resolve();
        }, this.generateDelay());
      });
    }
  
    public async ramp(speed: number, tick: number): Promise<void> {
      return new Promise<void>((resolve, reject) => {
        console.log(`Ramping motor to speed: ${speed} over ${tick} ticks`);
        setTimeout(() => {
          console.log(`Ramped motor to speed: ${speed} over ${tick} ticks`);
          resolve();
        }, this.generateDelay());
      });
    }
  }

  export { MockMotor };