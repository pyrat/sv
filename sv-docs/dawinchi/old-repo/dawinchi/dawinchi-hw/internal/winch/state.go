package winch

import (
	"github.com/CreateView/dawinchi-hw/internal/motor"
	"github.com/CreateView/dawinchi-hw/internal/panel"
)

type State struct {
	RightMotor   *motor.State `json:"right_motor"`
	LeftMotor    *motor.State `json:"left_motor"`
	ButtonsPanel *panel.State `json:"buttons_panel"`

	Status string `json:"status"`
	Error  string `json:"error"`
}

func (f *Firmware) State() *State {
	// TODO: implement status
	// by reading motor right and left
	// last irection, speed
	// counters
	// driver status
	// break status
	// ADS values

	winchState := &State{}

	winchState.RightMotor = f.motorRight.State()
	winchState.LeftMotor = f.motorLeft.State()
	winchState.ButtonsPanel = f.buttonsPanel.State()

	return winchState
}
