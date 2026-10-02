package winch

import (
	"context"
	"math/rand"
	"time"

	"github.com/CreateView/dawinchi-hw/internal/motor"
	"github.com/fatih/structs"
	"github.com/pkg/errors"
)

const (
	counterSpeedMargin = 8
	maxCounter         = 4294967295
	minCounter         = 0
	maxSpeed           = 100
	minSpeed           = 5
)

type logMove struct {
	RequestID         int    `json:"request_id" structs:"request_id"`
	Status            string `json:"status" structs:"status"`
	MotorRightCounter uint32 `json:"motor_right_counter" structs:"motor_right_counter"`
	MotorLeftCounter  uint32 `json:"motor_left_counter" structs:"motor_left_counter"`
	Speed             uint32 `json:"speed" structs:"speed"`
}

func (f *Firmware) Move(ctx context.Context, requestID int, moveToRightCounter, moveToLeftCounter uint32, speed uint32) error {
	if f.motorRight.State().Running || f.motorLeft.State().Running {
		return errors.New("motors are running")
	}

	// max, min counter
	if moveToRightCounter >= maxCounter {
		moveToRightCounter = maxCounter
	}
	if moveToRightCounter <= minCounter {
		moveToRightCounter = minCounter
	}

	if moveToLeftCounter >= maxCounter {
		moveToLeftCounter = maxCounter
	}
	if moveToLeftCounter <= minCounter {
		moveToLeftCounter = minCounter
	}

	// max, min speed
	if speed > maxSpeed {
		speed = maxSpeed
	}
	if speed < minSpeed {
		speed = minSpeed
	}

	// direction based on current counter
	motorRightCounter := f.motorRight.ReadCounter()
	motorLeftCounter := f.motorLeft.ReadCounter()

	motorRightDirection := motor.CW
	if motorRightCounter >= moveToRightCounter {
		motorRightDirection = motor.CCW
	}

	motorLeftDirection := motor.CW
	if motorLeftCounter >= moveToLeftCounter {
		motorLeftDirection = motor.CCW
	}

	// add counter margin based on speed and direction
	if motorRightDirection == motor.CW {
		moveToRightCounter = moveToRightCounter - (counterSpeedMargin * speed)
	}
	if motorRightDirection == motor.CCW {
		moveToRightCounter = moveToRightCounter + (counterSpeedMargin * speed)
	}
	if motorLeftDirection == motor.CW {
		moveToLeftCounter = moveToLeftCounter - (counterSpeedMargin * speed)
	}
	if motorLeftDirection == motor.CCW {
		moveToLeftCounter = moveToLeftCounter + (counterSpeedMargin * speed)
	}

	if err := f.motorRight.DisableBreak(); err != nil {
		return errors.Wrap(err, "motor right disable break")
	}

	if err := f.motorRight.Start(motorRightDirection, speed); err != nil {
		return errors.Wrap(err, "motor right start")
	}

	if err := f.motorLeft.DisableBreak(); err != nil {
		return errors.Wrap(err, "motor left disable break")
	}

	if err := f.motorLeft.Start(motorLeftDirection, speed); err != nil {
		f.motorRight.Stop()
		return errors.Wrap(err, "motor left start")
	}

	// rate for event logging
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	logRate := 0.0001

	motorRightDone := false
	motorLeftDone := false
	for {
		select {
		case <-ctx.Done():
			f.motorRight.Stop()
			f.motorLeft.Stop()
			return ctx.Err()

		default:
			motorRightCounter := f.motorRight.ReadCounter()
			switch motorRightDirection {
			case motor.CW:
				if motorRightCounter >= moveToRightCounter && !motorRightDone {
					f.motorRight.Stop()
					motorRightDone = true
				}
			case motor.CCW:
				if motorRightCounter <= moveToRightCounter && !motorRightDone {
					f.motorRight.Stop()
					motorRightDone = true
				}
			}

			motorLeftCounter = f.motorLeft.ReadCounter()
			switch motorLeftDirection {
			case motor.CW:
				if motorLeftCounter >= moveToLeftCounter && !motorLeftDone {
					f.motorLeft.Stop()
					motorLeftDone = true
				}
			case motor.CCW:
				if motorLeftCounter <= moveToLeftCounter && !motorLeftDone {
					f.motorLeft.Stop()
					motorLeftDone = true
				}
			}

			// log sample
			logEntry := logMove{
				RequestID:         requestID,
				Status:            "running",
				MotorRightCounter: motorRightCounter,
				MotorLeftCounter:  motorLeftCounter,
				Speed:             speed,
			}

			if r.Float64() < logRate {
				lm := structs.Map(logEntry)

				if err := f.AddEventLog("move", lm); err != nil {
					f.log.Error().Err(err).Msg("write event log")
				}
			}

			if motorRightDone && motorLeftDone {
				logEntry.Status = "done"
				logEntry.Speed = 0
				lm := structs.Map(logEntry)

				if err := f.AddEventLog("move", lm); err != nil {
					f.log.Error().Err(err).Msg("write event log")
				}
				return nil
			}

			time.Sleep(10 * time.Nanosecond)
		}
	}
}
