package motor

import (
	"github.com/pkg/errors"
	"github.com/stianeikeland/go-rpio/v4"
)

func (m *Motor) ReadBreakStatus() (int, error) {
	if m.BreakPin == 0 {
		return 0, errors.New("break pin not set")
	}

	p := rpio.Pin(m.BreakPin)
	return int(p.Read()), nil
}

func (m *Motor) DisableBreak() error {
	if m.BreakPin == 0 {
		return errors.New("break pin not set")
	}

	p := rpio.Pin(m.BreakPin)
	p.Output()
	p.High()

	return nil
}

func (m *Motor) EnableBreak() error {
	if m.BreakPin == 0 {
		return errors.New("break pin not set")
	}

	p := rpio.Pin(m.BreakPin)
	p.Output()
	p.Low()

	return nil
}
