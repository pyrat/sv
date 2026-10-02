package motor

type State struct {
	Direction Direction `json:"direction"`
	Running   bool      `json:"running"`
	Speed     uint32    `json:"speed"`
	Counter   uint32    `json:"counter"`

	DriverStatus struct {
		Status int   `json:"status"`
		Error  error `json:"error"`
	} `json:"driver_status"`

	BreakStatus struct {
		Status int   `json:"status"`
		Error  error `json:"error"`
	} `json:"break_status"`

	CurrentState struct {
		Voltage  float64 `json:"voltage"`
		Amperage float64 `json:"amperage"`
		Error    error   `json:"error"`
	} `json:"current_state"`
}

func (m *Motor) State() *State {
	running := false
	if m.speed != 0 {
		running = true
	}

	state := &State{
		Direction: m.lastDirection,
		Running:   running,
		Speed:     m.speed,
	}

	state.Counter = m.ReadCounter()

	driverStatus, err := m.ReadDriverStatus()
	state.DriverStatus.Status = driverStatus
	state.DriverStatus.Error = err

	breakStatus, err := m.ReadBreakStatus()
	state.BreakStatus.Status = breakStatus
	state.BreakStatus.Error = err

	voltage, err := m.ReadVoltage()
	state.CurrentState.Voltage = voltage
	state.CurrentState.Amperage = VoltageToAmps(voltage)
	state.CurrentState.Error = err

	return state
}
