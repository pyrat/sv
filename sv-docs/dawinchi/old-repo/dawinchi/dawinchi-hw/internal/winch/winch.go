package winch

import (
	"context"
	"sync"

	"github.com/CreateView/dawinchi-hw/internal/motor"
	"github.com/CreateView/dawinchi-hw/internal/panel"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Motor interface {
	ReadBreakStatus() (int, error)
	EnableBreak() error
	DisableBreak() error

	Start(direction motor.Direction, speed uint32) error
	SetSpeed(speed uint32) error
	State() *motor.State
	Stop()

	ReadCounter() uint32
	ResetCounter()
	ReadVoltage() (float64, error)
}

type ButtonsPanel interface {
	Action() <-chan panel.Action
	State() *panel.State
	Close()

	ShowError()
	ClearError()
	ShowLocked()
	ClearLocked()
	ShowRemoteLocked()
	ClearRemoteLocked()
}

type Firmware struct {
	motorRight    Motor
	motorLeft     Motor
	precisionMode bool
	buttonsPanel  ButtonsPanel
	log           zerolog.Logger

	mu       sync.Mutex
	eventLog firmwareEventLog
}

type firmwareEventLog struct {
	id       uint64
	filePath string
}

// New returns new winch firmware instance with two motors.
// Starts listening to buttons panel actions.
func New(ctx context.Context, motorRight, motorLeft Motor, buttonsPanel ButtonsPanel, log zerolog.Logger) (*Firmware, error) {
	f := &Firmware{
		motorRight:   motorRight,
		motorLeft:    motorLeft,
		buttonsPanel: buttonsPanel,
		eventLog: firmwareEventLog{
			id:       0,
			filePath: defaultEventLogFilePath,
		},
		log: log,
	}

	if err := f.motorRight.DisableBreak(); err != nil {
		return nil, errors.Wrap(err, "motor right disable break")
	}
	if err := f.motorLeft.DisableBreak(); err != nil {
		return nil, errors.Wrap(err, "motor left disable break")
	}

	if err := f.eventLogLoadLastID(); err != nil {
		return nil, errors.Wrap(err, "load event log last id")
	}

	if err := f.AddEventLog("init", nil); err != nil {
		return nil, errors.Wrap(err, "write event log")
	}

	// start listening to buttons panel actions
	f.Panel(ctx)

	return f, nil
}

func (f *Firmware) Reset() error {
	// TODO: implement reset
	// rest motor right and left - stop, enable break, reset counters
	// roll back to initial position?

	f.Stop()

	f.motorRight.ResetCounter()
	f.motorLeft.ResetCounter()

	if err := f.AddEventLog("reset", nil); err != nil {
		return errors.Wrap(err, "write event log")
	}

	return errors.New("not implemented")
}

func (f *Firmware) Stop() {
	f.motorRight.Stop()
	f.motorLeft.EnableBreak()

	f.motorLeft.Stop()
	f.motorLeft.EnableBreak()

	if err := f.AddEventLog("reset", nil); err != nil {
		f.log.Err(err).Msg("write event log")
	}
}

func (f *Firmware) Close() {
	// TODO: enable break on both motors?

	f.motorRight.Stop()
	f.motorLeft.Stop()
}
