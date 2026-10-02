package winch

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"

	"github.com/CreateView/dawinchi-hw/internal/hwerrors"
	"github.com/CreateView/dawinchi-hw/internal/hwlog"
	"github.com/CreateView/dawinchi-hw/internal/motor"
	"github.com/CreateView/dawinchi-hw/internal/panel"
)

type Firmware struct {
	upDownMotor  Motor
	inOutMotor   Motor
	buttonsPanel ButtonsPanel
	hwlog        HWLog
	log          zerolog.Logger
	sysCmd       SysCmd
	stateDB      *sqlx.DB

	mu                    sync.Mutex
	homePosition          *HomePosition
	precisionMode         bool
	remoteLock            bool
	serviceMode           bool
	keepUpDownLevel       bool
	factoryResetInitiated bool
	localCommandLock      bool
	lastActionSource      string

	apiMoveCancel              context.CancelFunc
	apiCcpMoveConnCheckStarted bool
	apiCcpMoveConnCheckCtx     context.Context
	apiCcpMoveConnCheckCancel  context.CancelFunc
}

// Init returns new winch firmware instance with two motors.
// Starts listening to buttons panel actions.
func Init(ctx context.Context, upDownMotor, inOutMotor Motor, buttonsPanel ButtonsPanel, hl HWLog, log zerolog.Logger) (*Firmware, error) {
	f := &Firmware{
		upDownMotor:               upDownMotor,
		inOutMotor:                inOutMotor,
		buttonsPanel:              buttonsPanel,
		hwlog:                     hl,
		log:                       log,
		sysCmd:                    &sysCmd{},
		apiMoveCancel:             func() {},
		apiCcpMoveConnCheckCancel: func() {},
		lastActionSource:          "hw",
	}

	eventLog := hwlog.NewEvent("", hwlog.HwInit)
	eventLog.Source = hwlog.EventSourceHw
	defer f.hwlog.AddEvent(eventLog)

	// load state
	if err := f.initState(ctx, defaultStateFilePath); err != nil {
		eventLog.SetError(hwerrors.LoadState, err.Error())
		f.hwlog.AddEvent(eventLog)
		eventLog.Error = nil

		f.log.Err(err).Msg("init state: activating remote lock")
		f.remoteLock = true
		f.buttonsPanel.ShowRemoteLocked()
	}

	if err := f.LoadState(); err != nil {
		eventLog.SetError(hwerrors.LoadState, err.Error())
		f.hwlog.AddEvent(eventLog)
		eventLog.Error = nil

		f.log.Err(err).Msg("load state: activating remote lock")
		f.remoteLock = true
		f.buttonsPanel.ShowRemoteLocked()
	}

	if err := f.upDownMotor.SelfTest(ctx, 0, 0); err != nil {
		eventLog.SetError(hwerrors.UpDownMotorSelfTest, err.Error())
		f.hwlog.AddEvent(eventLog)
		eventLog.Error = nil

		f.log.Err(err).Msg("motor up/down self test failed")
		f.buttonsPanel.SignalError()
	}
	if err := f.upDownMotor.Start(ctx); err != nil {
		eventLog.SetError(hwerrors.UpDownMotorStart, err.Error())
		return nil, errors.Wrap(err, "motor up/down start")
	}

	if err := f.inOutMotor.SelfTest(ctx, 0, 0); err != nil {
		eventLog.SetError(hwerrors.InOutMotorSelfTest, err.Error())
		f.hwlog.AddEvent(eventLog)
		eventLog.Error = nil

		f.log.Err(err).Msg("motor in/out self test failed")
		f.buttonsPanel.SignalError()
	}
	if err := f.inOutMotor.Start(ctx); err != nil {
		eventLog.SetError(hwerrors.InOutMotorStart, err.Error())
		return nil, errors.Wrap(err, "motor in/out start")
	}

	// start listening to buttons panel actions
	f.Panel(ctx)

	return f, nil
}

func (f *Firmware) Close() {
	f.upDownMotor.Stop()
	f.inOutMotor.Stop()

	f.buttonsPanel.Close()
	f.hwlog.Close()
}

func (f *Firmware) SetLastActionSource(source string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.lastActionSource = source
}

func (f *Firmware) RebootSysFromApi(ctx context.Context, requestID string) error {
	eventLog := hwlog.NewEvent(requestID, hwlog.ApiRebootSys)

	// record last action source
	f.SetLastActionSource(hwlog.EventSourceApi)
	eventLog.Source = f.lastActionSource

	// ceck remote lock
	if f.remoteLock {
		eventLog.SetError(hwerrors.RemoteLock, "remote lock on")
		return errors.New("remote lock on")
	}

	// check local command running
	if f.localCommandLock {
		eventLog.SetError(hwerrors.LocalCommandLock, "local command running, lock on")
		return errors.New("local command lock on")
	}

	if err := f.StopMotors(); err != nil {
		eventLog.SetError(hwerrors.StopMotors, err.Error())
		f.log.Err(err).Msg("stop motors")
	}

	if err := f.SaveState(); err != nil {
		f.log.Err(err).Msg("save state")
	}

	f.log.Info().Msg("reboot system")
	f.hwlog.AddEvent(eventLog)

	// wait for state to be saved, log event
	time.Sleep(2 * time.Second)

	if err := f.sysCmd.Reboot(ctx); err != nil {
		f.log.Err(err).Msg("reboot system")

		eventLog.SetError(hwerrors.Reboot, err.Error())
		f.hwlog.AddEvent(eventLog)

		return errors.Wrap(err, "reboot system")
	}

	return nil
}

// InitTestFirmware returns winch firmware instance for testing!!!.
func InitTestFirmware(t *testing.T, upDownMotor, inOutMotor *motor.Mock, buttonsPanel *panel.Mock) *Firmware {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()

	hwlogFile, err := os.CreateTemp(t.TempDir(), "test-winch.log")
	if err != nil {
		t.Fatal(err)
	}

	hl, err := hwlog.New(context.Background(), hwlogFile.Name(), log)
	if err != nil {
		t.Fatal(err)
	}

	syscmdMock := &SysCmdMock{}
	syscmdMock.On("Reboot", mock.Anything).Return(nil)

	f := &Firmware{
		upDownMotor:               upDownMotor,
		inOutMotor:                inOutMotor,
		buttonsPanel:              buttonsPanel,
		log:                       log,
		hwlog:                     hl,
		sysCmd:                    syscmdMock,
		apiMoveCancel:             func() {},
		apiCcpMoveConnCheckCancel: func() {},
		lastActionSource:          "hw",
		keepUpDownLevel:           true,
	}

	// init state
	stateFile, err := os.CreateTemp(t.TempDir(), "test-winch.conf")
	if err != nil {
		t.Fatal(err)
	}

	if err := f.initState(context.Background(), stateFile.Name()); err != nil {
		t.Fatal(err)
	}

	return f
}
