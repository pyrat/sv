package panel

import (
	"context"
	"time"
)

// 54	GPIO04	Lock LED
// 34	GPIO05	Error LED

const (
	LedLockedPin = 4
	LedErrorPin  = 5
)

func (p *Panel) ShowError() {
	p.ledError.Low()
}
func (p *Panel) ClearError() {
	p.ledError.High()
}

func (p *Panel) ShowLocked() {
	p.ledLocked.Low()
}
func (p *Panel) ToggleLocked() {
	p.ledLocked.Toggle()
}

func (p *Panel) ClearLocked() {
	p.ledLocked.High()
}

const ledBlinkTime = 500 * time.Millisecond

func (p *Panel) ShowRemoteLocked() {
	ctx, cancel := context.WithCancel(context.Background())
	p.remoteLockCancel = cancel

	go func() {
		for {
			select {
			case <-ctx.Done():
				p.ClearLocked()
				return
			default:
				p.ToggleLocked()
				time.Sleep(ledBlinkTime)
			}
		}
	}()
}

func (p *Panel) ClearRemoteLocked() {
	p.remoteLockCancel()
}
