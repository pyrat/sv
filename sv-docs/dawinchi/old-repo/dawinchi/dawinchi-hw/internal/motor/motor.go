package motor

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/stianeikeland/go-rpio/v4"
)

type Motor struct {
	Pin             int
	DirectionPin    int
	BreakPin        int
	DriverStatusPin int
	counter         Counter
	counterChip     uint8
	ads             ADS
	adsMUX          int

	drumSize          float64
	drumFullSpinCount float64
	ropeSize          float64

	mu            sync.Mutex
	lastDirection Direction
	speed         uint32
}

func New(pin, directionPin, breakPin, driverStatusPin int, counter Counter, counterChip uint8, ads ADS, adsMUX int) *Motor {
	m := &Motor{
		Pin:             pin,
		DirectionPin:    directionPin,
		BreakPin:        breakPin,
		DriverStatusPin: driverStatusPin,
		counter:         counter,
		counterChip:     counterChip,
		ads:             ads,
		adsMUX:          adsMUX,
		lastDirection:   CW,
	}
	return m
}

// Start starts the motor in the given direction at the given speed.
// Speed is from 0 to 100 and is a percentage of the maximum speed.
// Direction is either CW or CCW.
// Must be followed by a call to Stop() to stop the motor.
func (m *Motor) Start(direction Direction, speed uint32) error {
	// check driver fault status
	driverStatus, err := m.ReadDriverStatus()
	if err != nil {
		return errors.Wrap(err, "read driver status")
	}
	if driverStatus == int(rpio.Low) {
		return errors.Errorf("driver state fault: %d", driverStatus)
	}

	// check break
	breakStatus, err := m.ReadBreakStatus()
	if err != nil {
		return errors.Wrap(err, "read break status")
	}
	if breakStatus == int(rpio.Low) {
		return errors.Errorf("break on: %d", breakStatus)
	}

	// set direction
	pinDirection := rpio.Pin(m.DirectionPin)
	pinDirection.Output()
	switch direction {
	case CW:
		pinDirection.High()
	case CCW:
		pinDirection.Low()
	default:
		return errors.Errorf("invalid direction: %d", direction)
	}

	// start motor
	err = m.SetSpeed(speed)
	if err != nil {
		return errors.Wrap(err, "set speed")
	}

	m.mu.Lock()
	m.lastDirection = direction
	m.mu.Unlock()

	return nil
}

// SetSpeed sets the speed of the motor.
// Speed is from 0 to 100 and is a percentage of the maximum speed.
// Start the motor by setting PWM, frequency, and duty cycle.
// Must be followed by a call to Stop() to stop the motor.
func (m *Motor) SetSpeed(speed uint32) error {
	if speed > 100 || speed < 0 {
		return errors.Errorf("invalid speed: %d", speed)
	}

	pinMotor := rpio.Pin(m.Pin)
	pinMotor.Mode(rpio.Pwm)
	pinMotor.Freq(37500 * 100)
	pinMotor.DutyCycle(speed, 100)
	// motor at 37.5 kHz pulse
	// source frequency divided by cycle length

	m.mu.Lock()
	m.speed = speed
	m.mu.Unlock()

	return nil
}

func (m *Motor) Reset() {
	m.Stop()

	m.ResetCounter()
	m.counter.ResetStr(m.counterChip)

	// enable break?
}

// Stop stops the motor.
func (m *Motor) Stop() {
	p := rpio.Pin(m.Pin)
	p.Output()
	p.Low()

	m.mu.Lock()
	m.speed = 0
	m.mu.Unlock()
}
