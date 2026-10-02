package winch

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/CreateView/dawinchi-hw/internal/hwerrors"
	"github.com/CreateView/dawinchi-hw/internal/hwlog"
	"github.com/CreateView/dawinchi-hw/internal/motor"
	"github.com/pkg/errors"
)

type Motor interface {
	TurnOn()
	TurnOff()

	EnableBrake() error

	Start(ctx context.Context) error

	Direction() motor.Direction
	Speed() uint32
	SetSpeed(ctx context.Context, speed uint32) error
	State() *motor.State

	ReadCounter() int32
	LoadCntr(value int32) error
	ResetCounter()
	ReadVoltage() (float64, error)

	MinSpeed() uint32
	MaxSpeed() uint32

	MinCounter() int32
	MaxCounter() int32

	SetMinCounter(minCounter int32)
	SetMaxCounter(maxCounter int32)
	SetMaxSpeed(maxSpeed uint32)

	RampDownWindow() float64

	SetDirection(direction motor.Direction) error

	AccelerateWithRamp(ctx context.Context, speed uint32) error
	AccelerateLocalAction(ctx context.Context, speed uint32) error
	Stop() error
	StopWithRamp(ctx context.Context) error

	MoveTo(ctx context.Context, counter int32, speed uint32) error
	StopMoveTo()

	SelfTest(ctx context.Context, counterMove int32, speed uint32) error
	SetKeepUpDownLevelWithInOutMove(cancel context.CancelFunc)
}

func (f *Firmware) StopMotors() error {
	eventLog := hwlog.NewEvent("", hwlog.HwStopMotors)
	defer f.hwlog.AddEvent(eventLog)

	// record last action source
	f.SetLastActionSource(hwlog.EventSourceHw)
	eventLog.Source = f.lastActionSource

	if err := f.upDownMotor.Stop(); err != nil {
		eventLog.SetError(hwerrors.UpDownMotorStop, err.Error())
		return errors.Wrap(err, "up/down motor stop")
	}

	if err := f.upDownMotor.EnableBrake(); err != nil {
		eventLog.SetError(hwerrors.UpDownMotorBrakeEnable, err.Error())
		return errors.Wrap(err, "up/down motor enable brake")
	}

	if err := f.inOutMotor.Stop(); err != nil {
		eventLog.SetError(hwerrors.InOutMotorStop, err.Error())
		return errors.Wrap(err, "in/out motor stop")
	}

	if err := f.inOutMotor.EnableBrake(); err != nil {
		eventLog.SetError(hwerrors.InOutMotorBrakeEnable, err.Error())
		return errors.Wrap(err, "in/out motor enable brake")
	}

	return nil
}

func (f *Firmware) KeepUpDownLevelWithInOutMove(ctx context.Context, upDownStartCounter int32, inOutStartCounter int32, upDownInOutCounterRatio float64) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	f.mu.Lock()
	f.upDownMotor.SetKeepUpDownLevelWithInOutMove(cancel)
	f.mu.Unlock()

	speedAdjustment := uint32(2)

	f.log.Debug().Msgf("in/out start counter: %d, up/down start counter: %d", inOutStartCounter, upDownStartCounter)
	f.log.Debug().Msgf("up/down - in/out ratio: %.2f", upDownInOutCounterRatio)

	inOutStartDropCounter := calculateInOutDrop(float64(inOutStartCounter))
	f.log.Debug().Msgf("in/out start drop counter: %d", int32(inOutStartDropCounter))

	// slow tick up/down bouncy, not in/out
	// fast tick in/out bouncy, not up/down
	counterCheckTick := 25 * time.Millisecond
	ticker := time.NewTicker(counterCheckTick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			f.log.Info().Msgf("keep up/down level with in/out move canceled: %s", ctx.Err())

			return
		case <-ticker.C:
			if f.inOutMotor.Speed() == 0 {
				f.log.Debug().Msg("in/out motor speed is zero")
				if err := f.upDownMotor.Stop(); err != nil {
					f.log.Error().Err(err).Msg("up/down motor stop")
				}
				return
			}

			// keep direction
			if err := f.upDownMotor.SetDirection(f.inOutMotor.Direction()); err != nil {
				f.log.Error().Err(err).Msg("set up/down direction")
				return
			}

			// if err := f.checkCountersSafetyRatio(); err != nil {
			// 	f.log.Error().Err(err).Msg("check counters safety ratio")
			// 	return
			// }

			inOutCounter := f.inOutMotor.ReadCounter()
			upDownCounter := f.upDownMotor.ReadCounter()
			// f.log.Debug().Msgf("in/out counter: %d, up/down counter: %d", inOutCounter, upDownCounter)

			inOutAddedCounter := math.Abs(float64(inOutCounter - inOutStartCounter))
			upDownAddedCounter := math.Abs(float64(upDownCounter - upDownStartCounter))
			// f.log.Debug().Msgf("in/out added counter: %.2f, up/down added counter: %.2f", inOutAddedCounter, upDownAddedCounter)

			addInOutDropCounter := calculateInOutDrop(float64(inOutCounter))
			// f.log.Debug().Msgf("in/out drop counter: %.2f", addInOutDropCounter)

			addInOutDropCounter = addInOutDropCounter - inOutStartDropCounter
			// f.log.Debug().Msgf("in/out drop: %.2f", addInOutDropCounter)

			inOutAddedCounter = inOutAddedCounter + math.Abs(addInOutDropCounter)
			// f.log.Debug().Msgf("in/out added counter with drop: %.2f", inOutAddedCounter)

			if inOutCounter <= 0 {
				f.log.Debug().Msg("in/out counter less than zero")
				return
			}
			if upDownCounter <= 0 {
				f.log.Debug().Msg("up/down counter less than zero")
				return
			}

			ratio := upDownAddedCounter / inOutAddedCounter

			// f.log.Debug().Msgf("current ratio: %.2f", ratio)

			newSpeed := f.upDownMotor.Speed()
			// f.log.Debug().Msgf("up/down current speed: %d", newSpeed)

			if ratio > upDownInOutCounterRatio {
				if newSpeed < f.upDownMotor.MinSpeed()+speedAdjustment {
					f.log.Debug().Msgf("up/down speed: %d, less than min speed: %d + adjustment value: %d", newSpeed, f.upDownMotor.MinSpeed(), speedAdjustment)
					continue
				}

				newSpeed = f.upDownMotor.Speed() - speedAdjustment
			}

			if ratio < upDownInOutCounterRatio {
				if newSpeed > f.upDownMotor.MaxSpeed()-speedAdjustment {
					f.log.Debug().Msgf("up/down speed: %d, higer than max speed:%d - adjustment value: %d", newSpeed, f.upDownMotor.MaxSpeed(), speedAdjustment)
					continue
				}

				newSpeed = f.upDownMotor.Speed() + speedAdjustment
			}

			// f.log.Debug().Msgf("up/down new speed: %d", newSpeed)

			if err := f.upDownMotor.SetSpeed(ctx, newSpeed); err != nil {
				f.log.Error().Err(err).Msg("set up/down speed")
			}
		}
	}
}

var (
	inOutUpDownSpeedRatioCW            = 0.68
	inOutUpDownSpeedRatioCWCalibration = 0.003

	inOutUpDownSpeedRatioCCW            = 0.78
	inOutUpDownSpeedRatioCCWCalibration = 0.002

	upDownInOutCounterRatio = 0.63
)

func KeepUpDownLevelSpeed(inOutSpeed uint32, inOutDirection motor.Direction) uint32 {
	r := inOutUpDownSpeedRatioCW - (float64(inOutSpeed) * inOutUpDownSpeedRatioCWCalibration)

	if inOutDirection == motor.CCW {
		r = inOutUpDownSpeedRatioCCW - (float64(inOutSpeed) * inOutUpDownSpeedRatioCCWCalibration)
	}

	return uint32(float64(inOutSpeed) * r)
}

var (
	inOutDropAtCenter = 66000.0   // 120 cm, 550 counts / cm
	inOutLength       = 1759992.0 // 2400 cm, 733.33 counts / cm
)

func calculateInOutDrop(inOutCounter float64) float64 {
	return (4 * inOutDropAtCenter / (inOutLength * inOutLength)) * inOutCounter * (inOutLength - inOutCounter)
}

func (f *Firmware) checkCountersSafetyRatio() error {
	inOutCounter := f.inOutMotor.ReadCounter()
	upDownCounter := f.upDownMotor.ReadCounter()

	if inOutCounter <= 0 {
		f.log.Debug().Msg("in/out counter less than zero")
		return fmt.Errorf("in/out counter less than zero")
	}
	if upDownCounter <= 0 {
		f.log.Debug().Msg("up/down counter less than zero")
		return fmt.Errorf("up/down counter less than zero")
	}

	// safety ratio check
	safetyRatio := 0.56
	countersRatio := float64(upDownCounter) / float64(inOutCounter)
	f.log.Debug().Msgf("counters ratio: %.2f, safety: %.2f", countersRatio, safetyRatio)

	if countersRatio < safetyRatio {
		f.log.Debug().Msgf("counters ration: %.2f, under safety: %.2f", countersRatio, safetyRatio)

		f.apiMoveCancel()
		f.apiCcpMoveConnCheckCancel()

		if err := f.inOutMotor.Stop(); err != nil {
			f.log.Error().Err(err).Msg("in/out motor stop")
		}
		if err := f.upDownMotor.Stop(); err != nil {
			f.log.Error().Err(err).Msg("up/down motor stop")
		}

		eventLog := hwlog.NewEvent("", hwlog.HwCounterMinMaxGuard)
		eventLog.SetError(hwerrors.Counter, "counters ratio under safety")
		f.hwlog.AddEvent(eventLog)

		return fmt.Errorf("counters ratio under safety")
	}

	return nil
}
