package motor

import (
	"github.com/pkg/errors"
	"github.com/stianeikeland/go-rpio/v4"
)

func (m *Motor) SetDriverStatusPin(faultStatePin int) {
	m.DriverStatusPin = faultStatePin
}

func (m *Motor) ReadDriverStatus() (int, error) {
	if m.DriverStatusPin == 0 {
		return 0, errors.New("fault state pin not set")
	}

	p := rpio.Pin(m.DriverStatusPin)
	return int(p.Read()), nil
}
