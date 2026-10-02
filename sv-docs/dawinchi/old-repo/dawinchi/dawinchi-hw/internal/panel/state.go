package panel

type State struct {
	LedLockedState uint8 `json:"led_locked_state"`
	LedErrorState  uint8 `json:"led_error_state"`

	ButtonHomeState     string `json:"button_home_state"`
	ButtonHomeLastState string `json:"button_home_last_state"`

	ButtonUpState     string `json:"button_up_state"`
	ButtonUpLastState string `json:"button_up_last_state"`

	ButtonDownState     string `json:"button_down_state"`
	ButtonDownLastState string `json:"button_down_last_state"`

	ButtonInState     string `json:"button_in_state"`
	ButtonInLastState string `json:"button_in_last_state"`

	ButtonOutState     string `json:"button_out_state"`
	ButtonOutLastState string `json:"button_out_last_state"`

	ActionStatusRight string `json:"action_status_right"`
	ActionStatusLeft  string `json:"action_status_left"`
}

func (p *Panel) State() *State {
	p.mu.Lock()
	defer p.mu.Unlock()

	return &State{
		LedLockedState: uint8(p.ledLocked.Read()),
		LedErrorState:  uint8(p.ledError.Read()),

		ButtonHomeState:     p.buttonsState.homeState.String(),
		ButtonHomeLastState: p.buttonsState.homeLastState.String(),

		ButtonUpState:     p.buttonsState.upState.String(),
		ButtonUpLastState: p.buttonsState.upLastState.String(),

		ButtonDownState:     p.buttonsState.downState.String(),
		ButtonDownLastState: p.buttonsState.downLastState.String(),

		ButtonInState:     p.buttonsState.inState.String(),
		ButtonInLastState: p.buttonsState.inLastState.String(),

		ButtonOutState:     p.buttonsState.outState.String(),
		ButtonOutLastState: p.buttonsState.outLastState.String(),

		ActionStatusRight: p.statusRight.String(),
		ActionStatusLeft:  p.statusLeft.String(),
	}
}
