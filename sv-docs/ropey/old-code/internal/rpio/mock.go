package rpio

import (
	"github.com/stianeikeland/go-rpio/v4"
	"github.com/stretchr/testify/mock"
)

// Mock is a mock for the RpioPin interface.
type Mock struct {
	mock.Mock
}

// Input is a mock for the Input method
func (m *Mock) Input() {
	m.Called()
}

// Output is a mock for the Output method
func (m *Mock) Output() {
	m.Called()
}

// Pwm is a mock for the Pwm method
func (m *Mock) Pwm() {
	m.Called()
}

// High is a mock for the High method
func (m *Mock) High() {
	m.Called()
}

// Low is a mock for the Low method
func (m *Mock) Low() {
	m.Called()
}

// Toggle is a mock for the Toggle method
func (m *Mock) Toggle() {
	m.Called()
}

// Freq is a mock for the Freq method
func (m *Mock) Freq(freq int) {
	m.Called(freq)
}

// DutyCycle is a mock for the DutyCycle method
func (m *Mock) DutyCycle(dutyLen, cycleLen uint32) {
	m.Called(dutyLen, cycleLen)
}

// DutyCycleWithPwmMode is a mock for the DutyCycleWithPwmMode method
func (m *Mock) DutyCycleWithPwmMode(dutyLen, cycleLen uint32, mode bool) {
	m.Called(dutyLen, cycleLen, mode)
}

// Mode is a mock for the Mode method
func (m *Mock) Mode(mode rpio.Mode) {
	m.Called(mode)
}

// Write is a mock for the Write method
func (m *Mock) Write(state rpio.State) {
	m.Called(state)
}

// Read is a mock for the Read method
func (m *Mock) Read() rpio.State {
	args := m.Called()
	return args.Get(0).(rpio.State)
}

// PullUp is a mock for the PullUp method
func (m *Mock) PullUp() {
	m.Called()
}

// PullDown is a mock for the PullDown method
func (m *Mock) PullDown() {
	m.Called()
}

// PullOff is a mock for the PullOff method
func (m *Mock) PullOff() {
	m.Called()
}

// Detect is a mock for the Detect method
func (m *Mock) Detect(edge rpio.Edge) {
	m.Called(edge)
}

// EdgeDetected is a mock for the EdgeDetected method
func (m *Mock) EdgeDetected() bool {
	args := m.Called()
	return args.Get(0).(bool)
}
