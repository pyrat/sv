package winch

import (
	"context"
	"fmt"
	"time"

	"github.com/CreateView/dawinchi-hw/internal/hwerrors"
	"github.com/CreateView/dawinchi-hw/internal/hwlog"
	"github.com/CreateView/dawinchi-hw/internal/motor"
	"github.com/CreateView/dawinchi-hw/internal/panel"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type ButtonsPanel interface {
	Action() chan panel.ButtonAction
	SendButtonAction(ctx context.Context, action panel.Action, maxSpeed uint32)
	State() *panel.State
	Close()

	ShowError()
	ClearError()
	ShowLocked()
	ClearLocked()
	ShowRemoteLocked()
	ClearRemoteLocked()

	SignalError()
}

func (f *Firmware) Panel(ctx context.Context) {
	eventLog := hwlog.NewEvent("", hwlog.HwPanel)
	eventLog.Source = f.lastActionSource
	f.hwlog.AddEvent(eventLog)

	go func() {
		for {
			select {
			case <-ctx.Done():
				f.log.Err(ctx.Err()).Msg("panel: listening to buttons stopped")

				f.apiMoveCancel()
				f.apiCcpMoveConnCheckCancel()

				f.hwlog.AddEvent(eventLog)
				return

			case buttonAction := <-f.buttonsPanel.Action():
				eventLog := hwlog.NewEvent("", hwlog.HwPanel)

				f.log.Info().Msgf("panel: button action: %s: %d", buttonAction.Action, buttonAction.MaxSpeed)

				// stop any API move cmd
				f.apiMoveCancel()

				// apply precision mode
				maxSpeed := f.SpeedWithPrecisionMode(buttonAction.MaxSpeed)

				switch buttonAction.Action {
				case panel.UpDownMotorIncr, panel.CcpUpDownMotorIncr:
					if buttonAction.Action == panel.UpDownMotorIncr {
						f.EnableLocalCommandLock()
						eventLog.Method = hwlog.HwPanelUpDownMotorIncr
					}

					// handle direction change
					if f.upDownMotor.Direction() != motor.CW {
						if err := f.upDownMotor.Stop(); err != nil {
							f.log.Err(err).Msg("up/down motor stop")

							eventLog.SetError(hwerrors.UpDownMotorStop, err.Error())
							f.hwlog.AddEvent(eventLog)
						}
					}
					f.upDownMotor.SetDirection(motor.CW)

					go func() {
						var err error
						if buttonAction.Action == panel.UpDownMotorIncr {
							err = f.upDownMotor.AccelerateLocalAction(buttonAction.Ctx, maxSpeed)
						}
						if buttonAction.Action == panel.CcpUpDownMotorIncr {
							err = f.upDownMotor.AccelerateWithRamp(buttonAction.Ctx, maxSpeed)
						}

						if err != nil {
							f.log.Err(err).Msg("up/down motor move to counter")

							eventLog.SetError(hwerrors.UpDownMotorMoveTo, err.Error())
							f.hwlog.AddEvent(eventLog)

							var counterLimitErr *hwerrors.Error
							if errors.As(err, &counterLimitErr) && counterLimitErr.Code == hwerrors.Counter {
								f.buttonsPanel.SignalError()
							}
						}
					}()

				case panel.UpDownMotorDecr, panel.CcpUpDownMotorDecr:
					// if err := f.checkCountersSafetyRatio(); err != nil {
					// 	continue
					// }

					if buttonAction.Action == panel.UpDownMotorDecr {
						f.EnableLocalCommandLock()
						eventLog.Method = hwlog.HwPanelUpDownMotorDecr
					}

					// handle direction change
					if f.upDownMotor.Direction() != motor.CCW {
						if err := f.upDownMotor.Stop(); err != nil {
							f.log.Err(err).Msg("up/down motor stop")

							eventLog.SetError(hwerrors.UpDownMotorStop, err.Error())
							f.hwlog.AddEvent(eventLog)
							continue
						}
					}
					f.upDownMotor.SetDirection(motor.CCW)

					go func() {
						var err error
						if buttonAction.Action == panel.UpDownMotorDecr {
							err = f.upDownMotor.AccelerateLocalAction(buttonAction.Ctx, maxSpeed)
						}
						if buttonAction.Action == panel.CcpUpDownMotorDecr {
							err = f.upDownMotor.AccelerateWithRamp(buttonAction.Ctx, maxSpeed)
						}

						if err != nil {
							f.log.Err(err).Msg("up/down motor move to counter")

							eventLog.SetError(hwerrors.UpDownMotorMoveTo, err.Error())
							f.hwlog.AddEvent(eventLog)

							var counterLimitErr *hwerrors.Error
							if errors.As(err, &counterLimitErr) && counterLimitErr.Code == hwerrors.Counter {
								f.buttonsPanel.SignalError()
							}
						}

						// check counters ratio safety
						// go func() {
						// 	for f.upDownMotor.Speed() > 0 {
						// 		err = f.checkCountersSafetyRatio()
						// 		if err != nil {
						// 			f.buttonsPanel.SignalError()
						// 		}
						// 		time.Sleep(50 * time.Millisecond)
						// 	}
						// }()
					}()

				case panel.UpDownMotorStop:
					f.DisableLocalCommandLock()
					eventLog.Method = hwlog.HwPanelUpDownMotorStop

					if err := f.upDownMotor.Stop(); err != nil {
						f.log.Err(err).Msg("up/down motor stop")
						eventLog.SetError(hwerrors.UpDownMotorStop, err.Error())
					}

				case panel.CcpUpDownMotorStop:
					if err := f.upDownMotor.Stop(); err != nil {
						f.log.Err(err).Msg("up/down motor stop")
						eventLog.SetError(hwerrors.UpDownMotorStop, err.Error())
					}

				case panel.InOutMotorIncr, panel.CcpInOutMotorIncr:
					if buttonAction.Action == panel.InOutMotorIncr {
						f.EnableLocalCommandLock()
						eventLog.Method = hwlog.HwPanelInOutMotorIncr
					}

					// handle direction change
					if f.inOutMotor.Direction() != motor.CW {
						if err := f.inOutMotor.Stop(); err != nil {
							f.log.Err(err).Msg("in/out motor stop")

							eventLog.SetError(hwerrors.InOutMotorStop, err.Error())
							f.hwlog.AddEvent(eventLog)
							continue
						}
					}
					f.inOutMotor.SetDirection(motor.CW)

					// keep Up/Down level with In/Out move
					f.upDownMotor.SetDirection(motor.CW)

					go func() {
						upDownMaxSpeed := KeepUpDownLevelSpeed(maxSpeed, motor.CW)

						inOutCounter := f.inOutMotor.ReadCounter()
						upDownCounter := f.upDownMotor.ReadCounter()

						var err error
						if buttonAction.Action == panel.InOutMotorIncr {
							g, ctx := errgroup.WithContext(buttonAction.Ctx)
							g.Go(func() error {
								return f.inOutMotor.AccelerateLocalAction(ctx, maxSpeed)
							})

							g.Go(func() error {
								if !f.keepUpDownLevel {
									return nil
								}

								return f.upDownMotor.AccelerateLocalAction(ctx, upDownMaxSpeed)
							})

							err = g.Wait()
						}
						if buttonAction.Action == panel.CcpInOutMotorIncr {
							g, ctx := errgroup.WithContext(buttonAction.Ctx)
							g.Go(func() error {
								return f.inOutMotor.AccelerateWithRamp(ctx, maxSpeed)
							})

							g.Go(func() error {
								if !f.keepUpDownLevel {
									return nil
								}

								return f.upDownMotor.AccelerateWithRamp(ctx, upDownMaxSpeed)
							})

							err = g.Wait()
						}

						if err != nil {
							f.log.Err(err).Msg("in/out incr accelerate")

							if err := f.inOutMotor.Stop(); err != nil {
								f.log.Err(err).Msg("in/out motor stop")
							}
							if err := f.upDownMotor.Stop(); err != nil {
								f.log.Err(err).Msg("up/down motor stop")
							}

							eventLog.SetError(hwerrors.InOutMotorMoveTo, err.Error())
							f.hwlog.AddEvent(eventLog)

							var counterLimitErr *hwerrors.Error
							if errors.As(err, &counterLimitErr) && counterLimitErr.Code == hwerrors.Counter {
								f.buttonsPanel.SignalError()
							}
						} else {
							// start keep up/down level with in/out move
							go func() {
								if !f.keepUpDownLevel {
									return
								}

								upDownInOutCounterRatio := 0.63
								f.KeepUpDownLevelWithInOutMove(buttonAction.Ctx, upDownCounter, inOutCounter, upDownInOutCounterRatio)
							}()
						}
					}()

				case panel.InOutMotorDecr, panel.CcpInOutMotorDecr:
					if buttonAction.Action == panel.InOutMotorDecr {
						f.EnableLocalCommandLock()
						eventLog.Method = hwlog.HwPanelInOutMotorDecr
					}

					// handle direction change
					if f.inOutMotor.Direction() != motor.CCW {
						fmt.Println("inOutMotor.Direction() != motor.CCW")

						if err := f.inOutMotor.Stop(); err != nil {
							f.log.Err(err).Msg("in/out motor stop")

							eventLog.SetError(hwerrors.InOutMotorStop, err.Error())
							f.hwlog.AddEvent(eventLog)
							continue
						}
					}
					f.inOutMotor.SetDirection(motor.CCW)

					// keep Up/Down level with In/Out move
					f.upDownMotor.SetDirection(motor.CCW)

					go func() {
						upDownMaxSpeed := KeepUpDownLevelSpeed(maxSpeed, motor.CCW)

						inOutStartCounter := f.inOutMotor.ReadCounter()
						upDownStartCounter := f.upDownMotor.ReadCounter()

						var err error
						if buttonAction.Action == panel.InOutMotorDecr {
							g, ctx := errgroup.WithContext(buttonAction.Ctx)
							g.Go(func() error {
								return f.inOutMotor.AccelerateLocalAction(ctx, maxSpeed)
							})

							g.Go(func() error {
								if !f.keepUpDownLevel {
									return nil
								}

								return f.upDownMotor.AccelerateLocalAction(ctx, upDownMaxSpeed)
							})

							err = g.Wait()

						}
						if buttonAction.Action == panel.CcpInOutMotorDecr {
							g, ctx := errgroup.WithContext(buttonAction.Ctx)
							g.Go(func() error {
								return f.inOutMotor.AccelerateWithRamp(ctx, maxSpeed)
							})

							g.Go(func() error {
								if !f.keepUpDownLevel {
									return nil
								}

								return f.upDownMotor.AccelerateWithRamp(ctx, upDownMaxSpeed)
							})

							err = g.Wait()

						}

						if err != nil {
							f.log.Err(err).Msg("up/down decr accelerate")

							if err := f.inOutMotor.Stop(); err != nil {
								f.log.Err(err).Msg("in/out motor stop")
							}
							if err := f.upDownMotor.Stop(); err != nil {
								f.log.Err(err).Msg("up/down motor stop")
							}

							eventLog.SetError(hwerrors.InOutMotorMoveTo, err.Error())
							f.hwlog.AddEvent(eventLog)

							var counterLimitErr *hwerrors.Error
							if errors.As(err, &counterLimitErr) && counterLimitErr.Code == hwerrors.Counter {
								f.buttonsPanel.SignalError()
							}
						} else {
							// start keep up/down level with in/out move
							go func() {
								if !f.keepUpDownLevel {
									return
								}

								upDownInOutCounterRatio := 0.62
								f.KeepUpDownLevelWithInOutMove(buttonAction.Ctx, upDownStartCounter, inOutStartCounter, upDownInOutCounterRatio)
							}()
						}
					}()

				case panel.InOutMotorStop:
					f.DisableLocalCommandLock()
					eventLog.Method = hwlog.HwPanelInOutMotorStop

					if err := f.inOutMotor.Stop(); err != nil {
						f.log.Err(err).Msg("in/out motor stop")
						eventLog.SetError(hwerrors.InOutMotorStop, err.Error())
					}

					if err := f.upDownMotor.Stop(); err != nil {
						f.log.Err(err).Msg("up/down motor stop")
						eventLog.SetError(hwerrors.UpDownMotorStop, err.Error())
					}

				case panel.CcpInOutMotorStop:
					if err := f.inOutMotor.Stop(); err != nil {
						f.log.Err(err).Msg("in/out motor stop")
						eventLog.SetError(hwerrors.InOutMotorStop, err.Error())
					}

					if err := f.upDownMotor.Stop(); err != nil {
						f.log.Err(err).Msg("up/down motor stop")
						eventLog.SetError(hwerrors.UpDownMotorStop, err.Error())
					}

				case panel.TogglePrecisionMode:
					f.TogglePrecisionMode()

				case panel.ToggleRemoteLock:
					f.ToggleRemoteLock()

				case panel.ToggleServiceMode:
					f.ToggleServiceMode(buttonAction.Ctx)

				case panel.GoHome:
					f.GoHome(buttonAction.Ctx, maxSpeed)

				case panel.SetHome:
					f.SetHomePosition()

				case panel.InitFactoryReset:
					err := f.InitFactoryReset()
					if err != nil {
						f.log.Err(err).Msg("Panel: enable factory reset")
					}

				case panel.Reboot:
					go f.RebootSys(buttonAction.Ctx)
				}

				// save state for counter resync
				if err := f.SaveState(); err != nil {
					f.log.Error().Err(err).Msg("Panel: save state")
				}

				switch buttonAction.Action {
				case panel.CcpUpDownMotorIncr, panel.CcpUpDownMotorDecr, panel.CcpUpDownMotorStop, panel.CcpInOutMotorIncr, panel.CcpInOutMotorDecr, panel.CcpInOutMotorStop:
					// event for CCP already logged
					continue
				}

				// record last action source
				f.SetLastActionSource(hwlog.EventSourceHw)
				eventLog.Source = f.lastActionSource
				f.hwlog.AddEvent(eventLog)
			}
		}
	}()
}

func (f *Firmware) ToggleRemoteLock() {
	eventLog := hwlog.NewEvent("", hwlog.HwToggleRemoteLockOff)
	defer f.hwlog.AddEvent(eventLog)

	// record last action source
	f.SetLastActionSource(hwlog.EventSourceHw)
	eventLog.Source = f.lastActionSource

	f.mu.Lock()
	f.remoteLock = !f.remoteLock
	f.mu.Unlock()

	f.buttonsPanel.ClearRemoteLocked()
	if f.remoteLock {
		f.buttonsPanel.ShowRemoteLocked()
		eventLog.Method = hwlog.HwToggleRemoteLockOn
	}
}

func (f *Firmware) EnableRemoteLock() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.remoteLock = true
}

func (f *Firmware) disableRemoteLock() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.remoteLock = false
}

func (f *Firmware) TogglePrecisionMode() {
	eventLog := hwlog.NewEvent("", hwlog.HwTogglePrecisionModeOff)
	defer f.hwlog.AddEvent(eventLog)

	// record last action source
	f.SetLastActionSource(hwlog.EventSourceHw)
	eventLog.Source = f.lastActionSource

	f.mu.Lock()
	f.precisionMode = !f.precisionMode
	f.mu.Unlock()

	if f.precisionMode {
		eventLog.Method = hwlog.HwTogglePrecisionModeOn
	}

	// TOOD: remove it, only for debug
	f.buttonsPanel.SignalError()
}

func (f *Firmware) ToggleServiceMode(ctx context.Context) {
	eventLog := hwlog.NewEvent("", hwlog.HwToggleServiceModeOff)
	defer f.hwlog.AddEvent(eventLog)

	// record last action source
	f.SetLastActionSource(hwlog.EventSourceHw)
	eventLog.Source = f.lastActionSource

	f.mu.Lock()
	f.serviceMode = !f.serviceMode
	f.remoteLock = false
	f.mu.Unlock()

	f.buttonsPanel.ClearError()

	// restart motors
	f.upDownMotor.TurnOn()
	f.inOutMotor.TurnOn()

	if f.serviceMode {
		eventLog.Method = hwlog.HwToggleServiceModeOn

		f.mu.Lock()
		f.remoteLock = true
		f.mu.Unlock()

		f.buttonsPanel.ShowError()

		// stop motors
		if err := f.StopMotors(); err != nil {
			eventLog.SetError(hwerrors.StopMotors, err.Error())
			f.log.Err(err).Msg("stop motors")
			return
		}

		// turn motors off
		f.upDownMotor.TurnOff()
		f.inOutMotor.TurnOff()
	}
}

// SpeedWithPrecisionMode returns speed with precision mode applied.
// If precision mode is enabled, it reduces the speed by 70%.
func (f *Firmware) SpeedWithPrecisionMode(speed uint32) uint32 {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.precisionMode {
		return uint32(float32(speed) * 0.3)
	}

	return speed
}

func (f *Firmware) SetHomePosition() {
	eventLog := hwlog.NewEvent("", hwlog.HwSetHomePosition)
	defer f.hwlog.AddEvent(eventLog)

	// record last action source
	f.SetLastActionSource(hwlog.EventSourceHw)
	eventLog.Source = f.lastActionSource

	upDownMotorCounter := f.upDownMotor.ReadCounter()
	inOutMotorCounter := f.inOutMotor.ReadCounter()

	f.mu.Lock()
	f.homePosition = &HomePosition{
		UpDownMotorCounter: upDownMotorCounter,
		InOutMotorCounter:  inOutMotorCounter,
	}
	f.mu.Unlock()

	if err := f.SaveState(); err != nil {
		f.log.Err(err).Msg("save state")
	}

	// TOOD: remove it, only for debug
	f.buttonsPanel.SignalError()
}

func (f *Firmware) GoHome(ctx context.Context, speed uint32) {
	eventLog := hwlog.NewEvent("", hwlog.HwGoHome)
	defer f.hwlog.AddEvent(eventLog)

	// record last action source
	f.SetLastActionSource(hwlog.EventSourceHw)
	eventLog.Source = f.lastActionSource

	if f.homePosition == nil {
		eventLog.SetError(hwerrors.HomePositionNotSet, "home position not set")
		return
	}

	upDownmotorCounter := f.homePosition.UpDownMotorCounter
	inOutmotorCounter := f.homePosition.InOutMotorCounter

	// in/out move and keep up/down level
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := f.inOutMotor.MoveTo(gCtx, inOutmotorCounter, speed); err != nil {
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

		upDownMaxSpeed := KeepUpDownLevelSpeed(speed, inOutDirection)
		if err := f.upDownMotor.AccelerateWithRamp(gCtx, upDownMaxSpeed); err != nil {
			eventLog.SetError(hwerrors.UpDownMotorStop, err.Error())
			return errors.Wrap(err, "up/down motor accelerate")
		}

		go func() {
			if !f.keepUpDownLevel {
				return
			}

			f.KeepUpDownLevelWithInOutMove(gCtx, upDownStartCounter, inOutStartCounter, upDownInOutCounterRatio)
		}()

		return nil
	})

	eventLog.Method = hwlog.ApiMoveStarted
	f.hwlog.AddEvent(eventLog)

	if err := g.Wait(); err != nil {
		eventLog.SetError(hwerrors.InOutMotorMoveTo, err.Error())
		f.log.Err(err).Msg("in/out motor move to counter")
		return
	}

	// make sure up/down motor is stopped
	if err := f.upDownMotor.Stop(); err != nil {
		eventLog.SetError(hwerrors.UpDownMotorStop, err.Error())
		return
	}

	// now go to up/down counter
	if err := f.upDownMotor.MoveTo(ctx, upDownmotorCounter, speed); err != nil {
		eventLog.SetError(hwerrors.UpDownMotorStop, err.Error())
		f.log.Err(err).Msg("up/down motor move to counter")
		return
	}
}

func (f *Firmware) RebootSys(ctx context.Context) {
	eventLog := hwlog.NewEvent("", hwlog.HwRebootSys)

	// record last action source
	f.SetLastActionSource(hwlog.EventSourceHw)
	eventLog.Source = f.lastActionSource

	// stop any API move cmd
	f.apiMoveCancel()

	if err := f.StopMotors(); err != nil {
		eventLog.SetError(hwerrors.StopMotors, err.Error())
		f.hwlog.AddEvent(eventLog)
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
	}
}

func (f *Firmware) InitFactoryReset() error {
	eventLog := hwlog.NewEvent("", hwlog.HwFactoryReset)
	defer f.hwlog.AddEvent(eventLog)

	// record last action source
	f.SetLastActionSource(hwlog.EventSourceHw)
	eventLog.Source = f.lastActionSource

	// stop any API move cmd
	f.mu.Lock()
	f.remoteLock = true
	f.mu.Unlock()
	f.apiMoveCancel()

	// stop motors
	if err := f.StopMotors(); err != nil {
		eventLog.SetError(hwerrors.StopMotors, err.Error())

		f.mu.Lock()
		f.remoteLock = false
		f.mu.Unlock()
		return errors.Wrap(err, "stop motors")
	}

	// enable factory reset
	f.mu.Lock()
	f.factoryResetInitiated = true
	f.mu.Unlock()

	return nil
}

func (f *Firmware) EnableLocalCommandLock() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.apiMoveCancel()
	f.apiCcpMoveConnCheckCancel()

	f.localCommandLock = true
}

func (f *Firmware) DisableLocalCommandLock() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.apiMoveCancel()
	f.apiCcpMoveConnCheckCancel()

	f.localCommandLock = false
}
