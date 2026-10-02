package rpio

import "github.com/stianeikeland/go-rpio/v4"

type State = rpio.State

const (
	Low  = rpio.Low
	High = rpio.High
)

// Edge events
const (
	NoEdge   = rpio.NoEdge
	RiseEdge = rpio.RiseEdge
	FallEdge = rpio.FallEdge
	AnyEdge  = rpio.AnyEdge
)

const (
	Pwm = rpio.Pwm
)

type Pin interface {
	Input()
	Output()
	Pwm()
	High()
	Low()
	Toggle()
	Freq(freq int)
	DutyCycle(dutyLen, cycleLen uint32)
	DutyCycleWithPwmMode(dutyLen, cycleLen uint32, mode bool)
	Mode(mode rpio.Mode)
	Write(state rpio.State)
	Read() rpio.State
	PullUp()
	PullDown()
	PullOff()
	Detect(edge rpio.Edge)
	EdgeDetected() bool
}

func NewPin(pin uint8) Pin {
	return rpio.Pin(pin)
}
