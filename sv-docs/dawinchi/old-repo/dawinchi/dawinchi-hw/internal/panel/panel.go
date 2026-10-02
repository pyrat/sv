package panel

import (
	"context"
	"sync"

	"github.com/stianeikeland/go-rpio/v4"
)

type Panel struct {
	action chan Action

	ledLocked rpio.Pin
	ledError  rpio.Pin

	buttonHome rpio.Pin
	buttonUp   rpio.Pin
	buttonDown rpio.Pin
	buttonIn   rpio.Pin
	buttonOut  rpio.Pin

	remoteLockCancel context.CancelFunc
	actionCancel     context.CancelFunc

	mu           sync.Mutex
	buttonsState *ButtonsState
	statusRight  Status
	statusLeft   Status
	// for all actions locked?
}

type Status int

const (
	StatusOff Status = iota
	StatusIdle
	StatusOn
	StatusLocked
)

func (s Status) String() string {
	switch s {
	case StatusOff:
		return "Off"
	case StatusIdle:
		return "Idle"
	case StatusOn:
		return "On"
	case StatusLocked:
		return "Locked"
	default:
		return "Unknown"
	}
}

func New(ctx context.Context) *Panel {
	ledLocked := rpio.Pin(LedLockedPin)
	ledLocked.Output()
	ledLocked.High()

	ledError := rpio.Pin(LedErrorPin)
	ledError.Output()
	ledError.High()

	buttonHome := rpio.Pin(ButtonHomePin)
	buttonUp := rpio.Pin(ButtonUpPin)
	buttonDown := rpio.Pin(ButtonDownPin)
	buttonIn := rpio.Pin(ButtonInPin)
	buttonOut := rpio.Pin(ButtonOutPin)

	// set pin to detect button press
	buttonHome.Input()
	buttonHome.Detect(rpio.AnyEdge)

	buttonUp.Input()
	buttonUp.Detect(rpio.AnyEdge)

	buttonDown.Input()
	buttonDown.Detect(rpio.AnyEdge)

	buttonIn.Input()
	buttonIn.Detect(rpio.AnyEdge)

	buttonOut.Input()
	buttonOut.Detect(rpio.AnyEdge)

	p := &Panel{
		action: make(chan Action),

		ledLocked: ledLocked,
		ledError:  ledError,

		buttonHome: buttonHome,
		buttonUp:   buttonUp,
		buttonDown: buttonDown,
		buttonIn:   buttonIn,
		buttonOut:  buttonOut,

		statusRight: StatusOff,
		statusLeft:  StatusOff,

		actionCancel:     func() {},
		remoteLockCancel: func() {},
	}

	// start action loop
	p.start(ctx)

	return p
}

func (p *Panel) Close() {
	p.remoteLockCancel()
	p.actionCancel()

	p.ledLocked.High()
	p.ledError.High()
}
