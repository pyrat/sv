package panel

import (
	"context"
	"fmt"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

type Action int

const (
	GoHome              Action = iota // press HOME to move to home position
	SetHome                           // press and hold HOME 5s to set home position
	MoveUp                            // press or hold Up to move up
	MoveDown                          // press or hold Down to move down
	MoveRightStop                     // release Up || Down to stop
	MoveRightLocked                   // panel right is locked
	MoveIn                            // press or hold In to move in
	MoveOut                           // press or hold Out to move out
	MoveLeftStop                      // release In || Out to stop
	MoveLeftLocked                    // panel left is locked
	TogglePrecisionMode               // press and hold Up && Down 5s reduce motors speed by 69%
	Reboot                            // press and hold In && Out 10s reboot i.e the Raspberry Pi
)

func (a Action) String() string {
	switch a {
	case GoHome:
		return "GoHome"
	case SetHome:
		return "SetHome"
	case MoveUp:
		return "MoveUp"
	case MoveDown:
		return "MoveDown"
	case MoveRightStop:
		return "MoveRightStop"
	case MoveRightLocked:
		return "MoveRightLocked"
	case MoveIn:
		return "MoveIn"
	case MoveOut:
		return "MoveOut"
	case MoveLeftStop:
		return "MoveLeftStop"
	case MoveLeftLocked:
		return "MoveLeftLocked"
	case TogglePrecisionMode:
		return "TogglePrecisionMode"
	case Reboot:
		return "Reboot"
	default:
		return "Unknown"
	}
}

type ButtonsState struct {
	homeState      ButtonEvent
	homeLastState  ButtonEvent
	homePushedTime time.Time

	upState      ButtonEvent
	upLastState  ButtonEvent
	upPushedTime time.Time

	downState      ButtonEvent
	downLastState  ButtonEvent
	downPushedTime time.Time

	inState      ButtonEvent
	inLastState  ButtonEvent
	inPushedTime time.Time

	outState      ButtonEvent
	outLastState  ButtonEvent
	outPushedTime time.Time
}

const buttonStateDebounce = 100 * time.Millisecond

func (p *Panel) start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	p.actionCancel = cancel

	p.mu.Lock()
	p.statusRight = StatusOn
	p.statusLeft = StatusOn
	p.mu.Unlock()

	pbs := &ButtonsState{
		homeState:      ButtonHomeRelease,
		homeLastState:  ButtonHomeRelease,
		homePushedTime: time.Now(),

		upState:      ButtonUpRelease,
		upLastState:  ButtonUpRelease,
		upPushedTime: time.Now(),

		downState:      ButtonDownRelease,
		downLastState:  ButtonDownRelease,
		downPushedTime: time.Now(),

		inState:      ButtonInRelease,
		inLastState:  ButtonInRelease,
		inPushedTime: time.Now(),

		outState:      ButtonOutRelease,
		outLastState:  ButtonOutRelease,
		outPushedTime: time.Now(),
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				p.mu.Lock()
				p.statusRight = StatusOff
				p.statusLeft = StatusOff
				p.mu.Unlock()
				return

			default:
				time.Sleep(buttonStateDebounce)

				if p.buttonHome.EdgeDetected() {
					pbs.downPushedTime = time.Now()

					pbs.homeState = ButtonHomeRelease
					if p.buttonHome.Read() == rpio.Low {
						pbs.homeState = ButtonHomePush
					}
				}

				if p.buttonUp.EdgeDetected() {
					pbs.upPushedTime = time.Now()

					pbs.upState = ButtonUpRelease
					if p.buttonUp.Read() == rpio.Low {
						pbs.upState = ButtonUpPush
					}
				}

				if p.buttonDown.EdgeDetected() {
					pbs.downPushedTime = time.Now()

					pbs.downState = ButtonDownRelease
					if p.buttonDown.Read() == rpio.Low {
						pbs.downState = ButtonDownPush
					}
				}

				// fmt.Println("buttonIn:", p.buttonIn.Read())
				if p.buttonIn.EdgeDetected() {
					fmt.Println("buttonIn.EdgeDetected():", p.buttonIn.Read())

					pbs.inPushedTime = time.Now()

					pbs.inState = ButtonInRelease
					if p.buttonIn.Read() == rpio.Low {
						pbs.inState = ButtonInPush
					}
				}

				// fmt.Println("buttonOut:", p.buttonOut.Read())
				if p.buttonOut.EdgeDetected() {
					fmt.Println("buttonOut.EdgeDetected():", p.buttonOut.Read())

					pbs.outPushedTime = time.Now()

					pbs.outState = ButtonOutRelease
					if p.buttonOut.Read() == rpio.Low {
						pbs.outState = ButtonOutPush
					}
				}

				p.mu.Lock()
				p.buttonsState = pbs
				p.mu.Unlock()

				// p.actionGeneral()
				p.actionRight()
				p.actionLeft()
			}
		}
	}()
}

func (p *Panel) Action() <-chan Action {
	return p.action
}

const (
	actionDebounce                = 500 * time.Millisecond
	actionPrecisionModeActivation = 5 * time.Second
	actionRebootActivation        = 10 * time.Second
	actionSetHomeActivation       = 5 * time.Second
)

// GoHome, SetHome, FactoryReset, ToggleServiceMode, ToggleRemoteLock
func (p *Panel) actionGeneral() {
	// TODO: keep panel locked until GoHome action is completed
	if (p.buttonsState.homeState == ButtonHomeRelease && p.buttonsState.homeLastState == ButtonUpPush) &&
		p.statusRight == StatusIdle && p.statusLeft == StatusIdle {
		p.mu.Lock()
		p.statusRight = StatusLocked
		p.statusLeft = StatusLocked
		p.mu.Lock()

		p.action <- GoHome
	}
	p.mu.Lock()
	p.buttonsState.homeLastState = p.buttonsState.homeState
	p.mu.Unlock()

	if (p.buttonsState.homeState == ButtonHomePush) &&
		time.Since(p.buttonsState.homePushedTime) > actionSetHomeActivation {
		p.action <- SetHome
	}
}

func (p *Panel) actionRight() {
	// check if button was push and released
	// if button was pushed and released, then stop action
	// unlock panel
	if p.buttonsState.upState == ButtonUpRelease && p.buttonsState.upLastState == ButtonUpPush {
		p.mu.Lock()
		p.statusRight = StatusIdle
		p.mu.Unlock()

		p.action <- MoveRightStop
	}
	p.mu.Lock()
	p.buttonsState.upLastState = p.buttonsState.upState
	p.mu.Unlock()

	if p.buttonsState.downState == ButtonDownRelease && p.buttonsState.downLastState == ButtonDownPush {
		p.mu.Lock()
		p.statusRight = StatusIdle
		p.mu.Unlock()

		p.action <- MoveRightStop
	}
	p.mu.Lock()
	p.buttonsState.downLastState = p.buttonsState.downState
	p.mu.Unlock()

	// stop action if panel is locked
	if p.statusRight == StatusLocked {
		fmt.Println("action right locked")
		return
	}

	//
	// actions
	//
	if (p.buttonsState.upState == ButtonUpPush && p.buttonsState.downState != ButtonDownPush) &&
		time.Since(p.buttonsState.upPushedTime) > actionDebounce {
		p.mu.Lock()
		p.statusRight = StatusLocked
		p.mu.Unlock()

		p.action <- MoveUp
	}
	if (p.buttonsState.downState == ButtonDownPush && p.buttonsState.upState != ButtonUpPush) &&
		time.Since(p.buttonsState.downPushedTime) > actionDebounce {
		p.mu.Lock()
		p.statusRight = StatusLocked
		p.mu.Unlock()

		p.action <- MoveDown
	}

	// Up, Down for 5s to activate precision mode
	if (p.buttonsState.upState == ButtonUpPush && p.buttonsState.downState == ButtonDownPush) &&
		time.Since(p.buttonsState.upPushedTime) > actionPrecisionModeActivation &&
		time.Since(p.buttonsState.downPushedTime) > actionPrecisionModeActivation {

		p.mu.Lock()
		p.statusRight = StatusLocked
		p.mu.Unlock()

		p.action <- TogglePrecisionMode
	}
}

func (p *Panel) actionLeft() {
	// check if button was push and released
	// if button was pushed and released, then stop action
	// unlock panel
	if p.buttonsState.inState == ButtonInRelease && p.buttonsState.inLastState == ButtonInPush {
		p.mu.Lock()
		p.statusLeft = StatusIdle
		p.mu.Unlock()

		p.action <- MoveLeftStop
	}
	p.mu.Lock()
	p.buttonsState.inLastState = p.buttonsState.inState
	p.mu.Unlock()

	if p.buttonsState.outState == ButtonOutRelease && p.buttonsState.outLastState == ButtonOutPush {
		p.mu.Lock()
		p.statusLeft = StatusIdle
		p.mu.Unlock()

		p.action <- MoveLeftStop
	}
	p.mu.Lock()
	p.buttonsState.outLastState = p.buttonsState.outState
	p.mu.Unlock()

	// stop action if panel is locked
	if p.statusLeft == StatusLocked {
		fmt.Println("action right locked")
		return
	}

	//
	// actions
	//
	if (p.buttonsState.inState == ButtonInPush && p.buttonsState.outState != ButtonOutPush) &&
		time.Since(p.buttonsState.inPushedTime) > actionDebounce {
		p.mu.Lock()
		p.statusLeft = StatusLocked
		p.mu.Unlock()

		p.action <- MoveIn
	}
	if (p.buttonsState.outState == ButtonOutPush && p.buttonsState.inState != ButtonInPush) &&
		time.Since(p.buttonsState.outPushedTime) > actionDebounce {
		p.mu.Lock()
		p.statusLeft = StatusLocked
		p.mu.Unlock()

		p.action <- MoveOut
	}

	// In, Out for 10s to reboot
	if (p.buttonsState.inState == ButtonInPush && p.buttonsState.outState == ButtonOutPush) &&
		time.Since(p.buttonsState.inPushedTime) > actionRebootActivation &&
		time.Since(p.buttonsState.outPushedTime) > actionRebootActivation {
		p.mu.Lock()
		p.statusLeft = StatusLocked
		p.mu.Unlock()

		p.action <- Reboot
	}
}
