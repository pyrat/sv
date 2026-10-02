package winch

import (
	"context"

	"github.com/pkg/errors"

	"github.com/CreateView/dawinchi-hw/internal/hwerrors"
	"github.com/CreateView/dawinchi-hw/internal/hwlog"
	"github.com/CreateView/dawinchi-hw/internal/panel"
)

type CcpMoveFromApiParams struct {
	Up                     bool
	Down                   bool
	UpDownMotorStop        bool
	In                     bool
	Out                    bool
	InOutMotorStop         bool
	Source                 string
	KeepUpDownLevelDisable bool
	MaxSpeed               uint32
}

func (f *Firmware) CcpMoveFromApi(ctx context.Context, requestID string, params CcpMoveFromApiParams) error {
	eventLog := hwlog.NewEvent(requestID, hwlog.ApiCcpMove)
	defer f.hwlog.AddEvent(eventLog)

	// record last action source
	switch params.Source {
	case "api":
		f.SetLastActionSource(hwlog.EventSourceApi)
	case "ccp":
		f.SetLastActionSource(hwlog.EventSourceCcp)
	default:
		f.SetLastActionSource(hwlog.EventSourceUnknown)
	}
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

	// cancel previous move
	f.apiMoveCancel()

	if f.apiCcpMoveConnCheckStarted {
		f.log.Info().Msg("apiCcpMoveConnCheckStarted")
		f.apiCcpMoveConnCheckCancel()
	}

	var panelAction panel.Action
	var panelActionErr error

	switch {
	case params.Up:
		panelAction = panel.CcpUpDownMotorDecr
		eventLog.Method = hwlog.ApiCcpMoveUpDownMotorDecr

		// if err := f.checkCountersSafetyRatio(); err != nil {
		// 	return errors.Wrap(err, "ccp move")
		// }

	case params.Down:
		panelAction = panel.CcpUpDownMotorIncr
		eventLog.Method = hwlog.ApiCcpMoveUpDownMotorIncr
	case params.UpDownMotorStop:
		panelAction = panel.CcpUpDownMotorStop
		eventLog.Method = hwlog.ApiCcpMoveUpDownMotorStop

	case params.In:
		panelAction = panel.CcpInOutMotorDecr
		eventLog.Method = hwlog.ApiCcpMoveInOutMotorDecr
	case params.Out:
		panelAction = panel.CcpInOutMotorIncr
		eventLog.Method = hwlog.ApiCcpMoveInOutMotorIncr
	case params.InOutMotorStop:
		panelAction = panel.CcpInOutMotorStop
		eventLog.Method = hwlog.ApiCcpMoveInOutMotorStop

	default:
		panelAction = panel.Unknown
		panelActionErr = errors.New("unknown action")
	}

	if panelActionErr != nil {
		eventLog.SetError(hwerrors.CcpMoveAction, panelActionErr.Error())
		return errors.Wrap(panelActionErr, "ccp move")
	}

	// send action to panel
	f.buttonsPanel.SendButtonAction(ctx, panelAction, params.MaxSpeed)

	// check for disconnect
	go func() {
		f.mu.Lock()
		f.apiCcpMoveConnCheckStarted = true
		cppMoveCtx, cppMoveCancel := context.WithCancel(context.Background())
		f.apiCcpMoveConnCheckCtx = cppMoveCtx
		f.apiCcpMoveConnCheckCancel = cppMoveCancel
		f.mu.Unlock()

		for {
			select {
			case <-ctx.Done():
				f.log.Info().Msg("ccp move from api: context done")

				f.mu.Lock()
				f.apiCcpMoveConnCheckStarted = false
				f.mu.Unlock()

				if err := f.upDownMotor.Stop(); err != nil {
					f.log.Err(err).Msg("up/down motor stop")
					eventLog.SetError(hwerrors.UpDownMotorStop, err.Error())
					f.hwlog.AddEvent(eventLog)
				}
				if err := f.inOutMotor.Stop(); err != nil {
					f.log.Err(err).Msg("in/out motor stop")
					eventLog.SetError(hwerrors.InOutMotorStop, err.Error())
					f.hwlog.AddEvent(eventLog)
				}

				return
			case <-f.apiCcpMoveConnCheckCtx.Done():
				return
			}
		}
	}()

	return nil
}
