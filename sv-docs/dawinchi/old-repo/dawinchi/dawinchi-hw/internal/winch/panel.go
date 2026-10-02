package winch

import (
	"context"

	"github.com/CreateView/dawinchi-hw/internal/motor"
	"github.com/CreateView/dawinchi-hw/internal/panel"
	"github.com/fatih/structs"
)

var defaultMotorsSpeed uint32 = 15

type logPanel struct {
	Status            string `json:"status" structs:"status"`
	MotorRightCounter uint32 `json:"motor_right_counter" structs:"motor_right_counter"`
	MotorLeftCounter  uint32 `json:"motor_left_counter" structs:"motor_left_counter"`
	Speed             uint32 `json:"speed" structs:"speed"`
}

func (f *Firmware) Panel(ctx context.Context) {
	logEntry := logPanel{
		Status:            "start",
		MotorRightCounter: f.motorRight.ReadCounter(),
		MotorLeftCounter:  f.motorLeft.ReadCounter(),
		Speed:             0,
	}
	lm := structs.Map(logEntry)

	if err := f.AddEventLog("panel", lm); err != nil {
		f.log.Error().Err(err).Msg("write event log")
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				f.log.Err(ctx.Err()).Msg("panel: context done")

				logEntry := logPanel{
					Status:            "stop",
					MotorRightCounter: f.motorRight.ReadCounter(),
					MotorLeftCounter:  f.motorLeft.ReadCounter(),
					Speed:             0,
				}
				lm := structs.Map(logEntry)

				if err := f.AddEventLog("panel", lm); err != nil {
					f.log.Error().Err(err).Msg("write event log")
				}

				return

			case action := <-f.buttonsPanel.Action():
				actionStatus := "start"
				actionSpeed := defaultMotorsSpeed

				if action == panel.MoveUp {
					if err := f.motorRight.Start(motor.CW, defaultMotorsSpeed); err != nil {
						actionStatus = "error"

						// TODO: propagate error to the caller
						// TODO: stop both motors?
						// TODO: enable break on both motors?

						f.log.Error().Err(err).Msg("motor right start")
					}

				}
				if action == panel.MoveDown {
					if err := f.motorRight.Start(motor.CCW, defaultMotorsSpeed); err != nil {
						actionStatus = "error"
						f.log.Error().Err(err).Msg("motor right start")
					}
				}
				if action == panel.MoveRightStop {
					actionSpeed = 0
					f.motorRight.Stop()
				}

				if action == panel.MoveIn {
					if err := f.motorLeft.Start(motor.CCW, defaultMotorsSpeed); err != nil {
						actionStatus = "error"
						f.log.Error().Err(err).Msg("motor left start")
					}
				}
				if action == panel.MoveOut {
					if err := f.motorLeft.Start(motor.CW, defaultMotorsSpeed); err != nil {
						actionStatus = "error"
						f.log.Error().Err(err).Msg("motor left start")
					}
				}
				if action == panel.MoveLeftStop {
					actionSpeed = 0
					f.motorLeft.Stop()
				}

				if action == panel.TogglePrecisionMode {
					// precisionMode
					actionSpeed = 0
					f.buttonsPanel.ShowRemoteLocked()
				}
				if action == panel.Reboot {
					// reboot
					actionSpeed = 0
					f.buttonsPanel.ClearRemoteLocked()
				}

				logEntry := logPanel{
					Status:            actionStatus,
					MotorRightCounter: f.motorRight.ReadCounter(),
					MotorLeftCounter:  f.motorLeft.ReadCounter(),
					Speed:             actionSpeed,
				}
				lm := structs.Map(logEntry)

				if err := f.AddEventLog(action.String(), lm); err != nil {
					f.log.Error().Err(err).Msg("write event log")
				}
			}
		}
	}()
}
