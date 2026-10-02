package winch

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/CreateView/dawinchi-hw/internal/hwerrors"
	"github.com/CreateView/dawinchi-hw/internal/hwlog"
)

type MoveFromApiParams struct {
	MoveToUpDownCounter int32
	MoveToUpDownSpeed   uint32
	MoveToInOutCounter  int32
	MoveToInOutSpeed    uint32
}

func (f *Firmware) MoveFromApi(ctx context.Context, requestID string, params MoveFromApiParams) error {
	eventLog := hwlog.NewEvent(requestID, hwlog.ApiMove)
	defer f.hwlog.AddEvent(eventLog)

	// record last action source
	f.SetLastActionSource(hwlog.EventSourceApi)
	eventLog.Source = f.lastActionSource

	// ceck remote lock
	if f.remoteLock {
		eventLog.SetError(hwerrors.RemoteLock, "remote lock on")
		return errors.New("remote lock on")
	}

	if f.upDownMotor.State().Running || f.inOutMotor.State().Running {
		eventLog.SetError(hwerrors.MotorsAreRunning, "motors are running")

		return errors.New("motors are running")
	}

	// apply precision mode
	moveUpDownSpeed := f.SpeedWithPrecisionMode(params.MoveToUpDownSpeed)
	moveInOutSpeed := f.SpeedWithPrecisionMode(params.MoveToInOutSpeed)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	f.apiMoveCancel = cancel

	// in/out move and keep up/down level
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := f.inOutMotor.MoveTo(gCtx, params.MoveToInOutCounter, moveInOutSpeed); err != nil {
			eventLog.SetError(hwerrors.InOutMotorStop, err.Error())
			return errors.Wrap(err, "in/out motor move to")
		}

		f.log.Debug().Msg("in/out motor move finished")

		return nil
	})

	g.Go(func() error {
		if !f.keepUpDownLevel {
			return nil
		}

		// set up/down direction as in/out
		retryCount := 0
		maxRetry := 11

		for retryCount < maxRetry {
			if retryCount >= maxRetry-1 {
				f.log.Debug().Msg("in/out motor speed 0 after max retry, done")
				return nil
			}

			if f.inOutMotor.Speed() == 0 {
				f.log.Debug().Msg("in/out motor speed is 0, waiting for move, to find direction")

				retryCount++
				time.Sleep(50 * time.Millisecond)
				continue
			}

			retryCount = maxRetry + 1
		}

		inOutDirection := f.inOutMotor.Direction()
		f.upDownMotor.SetDirection(inOutDirection)

		inOutStartCounter := f.inOutMotor.ReadCounter()
		upDownStartCounter := f.upDownMotor.ReadCounter()

		upDownMaxSpeed := KeepUpDownLevelSpeed(moveInOutSpeed, inOutDirection)
		if err := f.upDownMotor.AccelerateWithRamp(gCtx, upDownMaxSpeed); err != nil {
			eventLog.SetError(hwerrors.UpDownMotorStop, err.Error())
			return errors.Wrap(err, "up/down motor accelerate")
		}

		go func() {
			f.KeepUpDownLevelWithInOutMove(gCtx, upDownStartCounter, inOutStartCounter, upDownInOutCounterRatio)
		}()

		return nil
	})

	eventLog.Method = hwlog.ApiMoveStarted
	f.hwlog.AddEvent(eventLog)

	if err := g.Wait(); err != nil {
		eventLog.Method = hwlog.ApiMove
		eventLog.SetError(hwerrors.MotorsApiMove, err.Error())
		return err
	}

	// make sure up/down motor is stopped
	if err := f.upDownMotor.Stop(); err != nil {
		eventLog.SetError(hwerrors.UpDownMotorStop, err.Error())
		return errors.Wrap(err, "up/down motor stop")
	}

	// now go to up/down counter
	if err := f.upDownMotor.MoveTo(ctx, params.MoveToUpDownCounter, moveUpDownSpeed); err != nil {
		eventLog.SetError(hwerrors.UpDownMotorStop, err.Error())
		return errors.Wrap(err, "up/down motor move to")
	}

	// save state for counter resync
	if err := f.SaveState(); err != nil {
		f.log.Error().Err(err).Msg("MoveFromApi: save state")
	}

	eventLog.Method = hwlog.ApiMoveFinished
	return nil
}

func (f *Firmware) StopMoveFromApi(ctx context.Context, requestID string) error {
	eventLog := hwlog.NewEvent(requestID, hwlog.ApiStopMove)
	defer f.hwlog.AddEvent(eventLog)

	// record last action source
	f.SetLastActionSource(hwlog.EventSourceApi)
	eventLog.Source = f.lastActionSource

	// stop any API move cmd
	f.apiMoveCancel()

	if err := f.upDownMotor.Stop(); err != nil {
		eventLog.SetError(hwerrors.UpDownMotorStop, err.Error())
		f.hwlog.AddEvent(eventLog)
	}

	if err := f.inOutMotor.Stop(); err != nil {
		eventLog.SetError(hwerrors.InOutMotorStop, err.Error())
		f.hwlog.AddEvent(eventLog)
	}

	// save state for counter resync
	if err := f.SaveState(); err != nil {
		f.log.Error().Err(err).Msg("StopMoveFromApi: save state")
	}

	return nil
}
