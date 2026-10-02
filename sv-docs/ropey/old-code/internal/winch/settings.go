package winch

import (
	"context"

	"github.com/pkg/errors"

	"github.com/CreateView/dawinchi-hw/internal/hwerrors"
	"github.com/CreateView/dawinchi-hw/internal/hwlog"
)

type SetHomePositionFromApiParams struct {
	UpDownMotorCounter int32
	InOutMotorCounter  int32
}

func (f *Firmware) SetHomePositionFromApi(requestID string, params SetHomePositionFromApiParams) error {
	eventLog := hwlog.NewEvent(requestID, hwlog.ApiSetHomePosition)
	defer f.hwlog.AddEvent(eventLog)

	// record last action source
	f.SetLastActionSource(hwlog.EventSourceApi)
	eventLog.Source = f.lastActionSource

	// check remote lock
	if f.remoteLock {
		eventLog.SetError(hwerrors.RemoteLock, "remote lock on")
		return errors.New("remote lock on")
	}

	if params.UpDownMotorCounter > f.upDownMotor.MaxCounter() || params.UpDownMotorCounter < f.upDownMotor.MinCounter() {
		return errors.New("invalid up/down motor counter")
	}
	if params.InOutMotorCounter > f.inOutMotor.MaxCounter() || params.InOutMotorCounter < f.inOutMotor.MinCounter() {
		return errors.New("invalid in/out motor counter")
	}

	f.mu.Lock()
	f.homePosition = &HomePosition{
		UpDownMotorCounter: params.UpDownMotorCounter,
		InOutMotorCounter:  params.InOutMotorCounter,
	}
	f.mu.Unlock()

	if err := f.SaveState(); err != nil {
		return errors.Wrap(err, "save state")
	}

	return nil
}

const (
	factorySettingsMotorMinCounter = -2147483648
	factorySettingsMotorMaxCounter = 2147483647
	factorySettingsMinSpeed        = 10
	factorySettingsMaxSpeed        = 100
)

type SetSettingsFromApiParams struct {
	UpDownMotorMinCounter int32
	UpDownMotorMaxCounter int32
	UpDownMotorMaxSpeed   uint32

	InOutMotorMinCounter int32
	InOutMotorMaxCounter int32
	InOutMotorMaxSpeed   uint32

	KeepUpDownLevel bool
}

func (f *Firmware) SetSettingsFromApi(requestID string, params SetSettingsFromApiParams) error {
	eventLog := hwlog.NewEvent(requestID, hwlog.ApiSetSettings)
	defer f.hwlog.AddEvent(eventLog)

	// record last action source
	f.SetLastActionSource(hwlog.EventSourceApi)
	eventLog.Source = f.lastActionSource

	// ceck remote lock
	if f.remoteLock {
		eventLog.SetError(hwerrors.RemoteLock, "remote lock on")

		return errors.New("remote lock on")
	}

	if err := f.settingsFromApiValidation(params); err != nil {
		return errors.Wrap(err, "set settings from api validation")
	}

	f.upDownMotor.SetMinCounter(params.UpDownMotorMinCounter)
	f.upDownMotor.SetMaxCounter(params.UpDownMotorMaxCounter)
	f.upDownMotor.SetMaxSpeed(params.UpDownMotorMaxSpeed)

	f.inOutMotor.SetMinCounter(params.InOutMotorMinCounter)
	f.inOutMotor.SetMaxCounter(params.InOutMotorMaxCounter)
	f.inOutMotor.SetMaxSpeed(params.InOutMotorMaxSpeed)

	f.mu.Lock()
	f.keepUpDownLevel = params.KeepUpDownLevel
	f.mu.Unlock()

	if err := f.SaveState(); err != nil {
		return errors.Wrap(err, "save state")
	}

	return nil
}

func (f *Firmware) settingsFromApiValidation(params SetSettingsFromApiParams) error {
	if err := f.settingsValidation(params); err != nil {
		return errors.Wrap(err, "settings validation")
	}

	// max, min can't be inside current counter
	if params.UpDownMotorMinCounter > f.upDownMotor.ReadCounter() {
		return errors.New("invalid up/down motor min counter: min > current counter")
	}
	if params.UpDownMotorMaxCounter < f.upDownMotor.ReadCounter() {
		return errors.New("invalid up/down motor max counter: max < current counter")
	}

	// max, min can't be inside current counter
	if params.InOutMotorMinCounter > f.inOutMotor.ReadCounter() {
		return errors.New("invalid in/out motor min counter: min > current counter")
	}
	if params.InOutMotorMaxCounter < f.inOutMotor.ReadCounter() {
		return errors.New("invalid in/out motor max counter: max < current counter")
	}

	return nil
}

func (f *Firmware) settingsValidation(params SetSettingsFromApiParams) error {
	// up/down motor
	if params.UpDownMotorMinCounter > params.UpDownMotorMaxCounter {
		return errors.New("invalid up/down motor min counter: min > max")
	}
	if params.UpDownMotorMaxCounter < params.UpDownMotorMinCounter {
		return errors.New("invalid up/down motor max counter: max < min")
	}
	if params.UpDownMotorMinCounter < factorySettingsMotorMinCounter || params.UpDownMotorMinCounter > factorySettingsMotorMaxCounter {
		return errors.New("invalid up/down motor min counter")
	}
	if params.UpDownMotorMaxCounter < factorySettingsMotorMinCounter || params.UpDownMotorMaxCounter > factorySettingsMotorMaxCounter {
		return errors.New("invalid up/down motor max counter")
	}
	if params.UpDownMotorMaxSpeed > factorySettingsMaxSpeed || params.UpDownMotorMaxSpeed < factorySettingsMinSpeed {
		return errors.New("invalid up/down motor max speed")
	}

	// in/out motor
	if params.InOutMotorMinCounter > params.InOutMotorMaxCounter {
		return errors.New("invalid in/out motor min counter: min > max")
	}
	if params.InOutMotorMaxCounter < params.InOutMotorMinCounter {
		return errors.New("invalid in/out motor max counter: max < min")
	}
	if params.InOutMotorMinCounter < factorySettingsMotorMinCounter || params.InOutMotorMinCounter > factorySettingsMotorMaxCounter {
		return errors.Errorf("invalid in/out motor min counter")
	}
	if params.InOutMotorMaxCounter < factorySettingsMotorMinCounter || params.InOutMotorMaxCounter > factorySettingsMotorMaxCounter {
		return errors.Errorf("invalid in/out motor max counter")
	}
	if params.InOutMotorMaxSpeed > factorySettingsMaxSpeed || params.InOutMotorMaxSpeed < factorySettingsMinSpeed {
		return errors.New("invalid in/out motor max speed")
	}

	return nil
}

func (f *Firmware) InitFactoryResetFromApi(ctx context.Context, requestID string) error {
	eventLog := hwlog.NewEvent(requestID, hwlog.ApiFactoryResetInitiated)
	defer f.hwlog.AddEvent(eventLog)

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

	// stop any API move cmd
	f.apiMoveCancel()

	// stop motors
	if err := f.StopMoveFromApi(ctx, requestID); err != nil {
		return errors.Wrap(err, "stop move from api")
	}

	// turn off motors
	f.upDownMotor.TurnOff()
	f.inOutMotor.TurnOff()

	f.mu.Lock()
	f.remoteLock = true
	f.factoryResetInitiated = true
	f.mu.Unlock()

	return nil
}

type FactorySettingsFromApiParams struct {
	UpDownMotorMinCounter int32
	UpDownMotorMaxCounter int32
	UpDownMotorMaxSpeed   uint32

	InOutMotorMinCounter int32
	InOutMotorMaxCounter int32
	InOutMotorMaxSpeed   uint32

	HomePositionUpDownMotorCounter int32
	HomePositionInOutMotorCounter  int32

	KeepUpDownLevel bool
}

func (f *Firmware) SetFactorySettingsFromApi(ctx context.Context, requestID string, params FactorySettingsFromApiParams) error {
	eventLog := hwlog.NewEvent(requestID, hwlog.ApiSetFactoryResetSettings)
	defer f.hwlog.AddEvent(eventLog)

	// record last action source
	f.SetLastActionSource(hwlog.EventSourceApi)
	eventLog.Source = f.lastActionSource

	// EnableFactoryReset must be called first
	if !f.factoryResetInitiated {
		eventLog.SetError(hwerrors.FactoryResetNotInitiated, "factory reset not initiated")
		return errors.New("factory reset not initiated")
	}

	// stop any API move cmd
	f.apiMoveCancel()

	// set settings
	settingsParams := SetSettingsFromApiParams{
		UpDownMotorMinCounter: params.UpDownMotorMinCounter,
		UpDownMotorMaxCounter: params.UpDownMotorMaxCounter,
		UpDownMotorMaxSpeed:   params.UpDownMotorMaxSpeed,

		InOutMotorMinCounter: params.InOutMotorMinCounter,
		InOutMotorMaxCounter: params.InOutMotorMaxCounter,
		InOutMotorMaxSpeed:   params.InOutMotorMaxSpeed,
	}
	if err := f.settingsValidation(settingsParams); err != nil {
		return errors.Wrap(err, "settings validation")
	}

	if params.HomePositionUpDownMotorCounter > params.UpDownMotorMaxCounter || params.HomePositionUpDownMotorCounter < params.UpDownMotorMinCounter {
		return errors.New("invalid home up/down motor counter")
	}
	if params.HomePositionInOutMotorCounter > params.InOutMotorMaxCounter || params.HomePositionInOutMotorCounter < params.InOutMotorMinCounter {
		return errors.New("invalid home in/out motor counter")
	}

	// reset counters
	f.upDownMotor.ResetCounter()
	f.inOutMotor.ResetCounter()

	f.upDownMotor.SetMinCounter(params.UpDownMotorMinCounter)
	f.upDownMotor.SetMaxCounter(params.UpDownMotorMaxCounter)
	f.upDownMotor.SetMaxSpeed(params.UpDownMotorMaxSpeed)

	f.inOutMotor.SetMinCounter(params.InOutMotorMinCounter)
	f.inOutMotor.SetMaxCounter(params.InOutMotorMaxCounter)
	f.inOutMotor.SetMaxSpeed(params.InOutMotorMaxSpeed)

	// reset remote lock, factory reset initiation
	f.mu.Lock()
	f.remoteLock = false
	f.keepUpDownLevel = params.KeepUpDownLevel
	f.factoryResetInitiated = false
	f.homePosition = &HomePosition{
		UpDownMotorCounter: params.HomePositionUpDownMotorCounter,
		InOutMotorCounter:  params.HomePositionInOutMotorCounter,
	}
	f.mu.Unlock()

	// turn on motors
	f.upDownMotor.TurnOn()
	f.inOutMotor.TurnOn()

	// reset counters to clear self test
	f.upDownMotor.ResetCounter()
	f.inOutMotor.ResetCounter()

	if err := f.SaveState(); err != nil {
		return errors.Wrap(err, "save state")
	}

	return nil
}

func (f *Firmware) SelfTestFromApi(ctx context.Context, requestID string, counterMove int32, speed uint32) error {
	eventLog := hwlog.NewEvent(requestID, hwlog.ApiSelfTest)
	defer f.hwlog.AddEvent(eventLog)

	// record last action source
	f.SetLastActionSource(hwlog.EventSourceApi)
	eventLog.Source = f.lastActionSource

	// ceck remote lock
	if f.remoteLock {
		eventLog.SetError(hwerrors.RemoteLock, "remote lock on")
		return errors.New("remote lock on")
	}

	// cancel previous move
	f.apiMoveCancel()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	f.apiMoveCancel = cancel

	if err := f.upDownMotor.Stop(); err != nil {
		eventLog.SetError(hwerrors.UpDownMotorStop, err.Error())
		return errors.Wrap(err, "up/dow mnotor stop")
	}

	if err := f.inOutMotor.Stop(); err != nil {
		eventLog.SetError(hwerrors.InOutMotorStop, err.Error())
		return errors.Wrap(err, "in/out motor stop")
	}

	if err := f.upDownMotor.SelfTest(ctx, counterMove, speed); err != nil {
		eventLog.SetError(hwerrors.UpDownMotorSelfTest, err.Error())
		return errors.Wrap(err, "up/down motor self test")
	}

	if err := f.inOutMotor.SelfTest(ctx, counterMove, speed); err != nil {
		eventLog.SetError(hwerrors.InOutMotorSelfTest, err.Error())
		return errors.Wrap(err, "in/out motor self test")
	}

	// all good - turn on motor
	f.upDownMotor.TurnOn()
	f.inOutMotor.TurnOn()

	return nil
}
