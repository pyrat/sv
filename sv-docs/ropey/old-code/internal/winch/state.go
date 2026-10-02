package winch

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	_ "modernc.org/sqlite"

	"github.com/CreateView/dawinchi-hw/internal/hwerrors"
	"github.com/CreateView/dawinchi-hw/internal/motor"
	"github.com/CreateView/dawinchi-hw/internal/panel"
)

const (
	defaultStateFilePath = "/etc/winch.conf"
)

const create_winch_state = `
	CREATE TABLE IF NOT EXISTS winch_state (
	id INTEGER NOT NULL PRIMARY KEY,
	state_json TEXT NOT NULL,
	time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`

func (f *Firmware) initState(ctx context.Context, dns string) error {
	if dns == "" {
		return errors.Errorf("dns is empty")
	}

	db, err := sqlx.ConnectContext(ctx, "sqlite", dns)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)

	err = db.Ping()
	if err != nil {
		return errors.Wrap(err, "ping db")
	}

	// provision db
	_, err = db.Exec(create_winch_state)
	if err != nil {
		return errors.Wrap(err, "create winch state table")
	}

	f.stateDB = db

	return nil
}

type State struct {
	UpDownMotor           *motor.State    `json:"up_down_motor"`
	InOutMotor            *motor.State    `json:"in_out_motor"`
	ButtonsPanel          *panel.State    `json:"buttons_panel"`
	PrecisionMode         bool            `json:"precision_mode"`
	ServiceMode           bool            `json:"service_mode"`
	RemoteLock            bool            `json:"remote_lock"`
	KeepUpDownLevel       bool            `json:"keep_up_down_level"`
	LastActionSource      string          `json:"last_action_source"`
	HomePosition          *HomePosition   `json:"home_position"`
	FactoryResetInitiated bool            `json:"factory_reset_initiated"`
	Error                 *hwerrors.Error `json:"error"`
}

type HomePosition struct {
	UpDownMotorCounter int32 `json:"up_down_motor_counter"`
	InOutMotorCounter  int32 `json:"in_out_motor_counter"`
}

func (f *Firmware) State() *State {
	winchState := &State{}

	winchState.UpDownMotor = f.upDownMotor.State()
	winchState.InOutMotor = f.inOutMotor.State()
	winchState.ButtonsPanel = f.buttonsPanel.State()

	winchState.PrecisionMode = f.precisionMode
	winchState.ServiceMode = f.serviceMode
	winchState.RemoteLock = f.remoteLock
	winchState.KeepUpDownLevel = f.keepUpDownLevel
	winchState.LastActionSource = f.lastActionSource
	winchState.FactoryResetInitiated = f.factoryResetInitiated
	winchState.HomePosition = f.homePosition

	return winchState
}

func (f *Firmware) SaveState() error {
	stateJson, err := json.Marshal(f.State())
	if err != nil {
		return errors.Wrap(err, "marshal state")
	}

	query := "INSERT INTO winch_state (state_json) VALUES (?)"

	_, err = f.stateDB.Exec(query, string(stateJson))
	if err != nil {
		return errors.Wrap(err, "save state")
	}

	return nil
}

func (f *Firmware) LoadState() error {
	query := `SELECT state_json FROM winch_state ORDER BY id DESC LIMIT 1`

	var stateJson string
	err := f.stateDB.Get(&stateJson, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return errors.Wrap(err, "load state")
	}

	state := &State{}
	err = json.Unmarshal([]byte(stateJson), state)
	if err != nil {
		return errors.Wrap(err, "decode state")
	}

	f.precisionMode = state.PrecisionMode
	f.serviceMode = state.ServiceMode
	f.remoteLock = state.RemoteLock
	f.keepUpDownLevel = state.KeepUpDownLevel
	f.homePosition = state.HomePosition
	f.factoryResetInitiated = state.FactoryResetInitiated

	if f.remoteLock {
		f.buttonsPanel.ShowRemoteLocked()
	}

	if f.serviceMode {
		f.buttonsPanel.ClearRemoteLocked()
		f.remoteLock = true
		f.buttonsPanel.ShowError()
	}

	if state.UpDownMotor != nil {
		if err := f.upDownMotor.LoadCntr(state.UpDownMotor.Counter); err != nil {
			return errors.Wrap(err, "load up/down motor counter")
		}

		if state.UpDownMotor.Settings != nil {
			f.upDownMotor.SetMinCounter(state.UpDownMotor.Settings.MinCounter)
			f.upDownMotor.SetMaxCounter(state.UpDownMotor.Settings.MaxCounter)
			f.upDownMotor.SetMaxSpeed(state.UpDownMotor.Settings.MaxSpeed)
		}
	}

	if state.InOutMotor != nil {
		if err := f.inOutMotor.LoadCntr(state.InOutMotor.Counter); err != nil {
			return errors.Wrap(err, "load in/out motor counter")
		}

		if state.InOutMotor.Settings != nil {
			f.inOutMotor.SetMinCounter(state.InOutMotor.Settings.MinCounter)
			f.inOutMotor.SetMaxCounter(state.InOutMotor.Settings.MaxCounter)
			f.inOutMotor.SetMaxSpeed(state.InOutMotor.Settings.MaxSpeed)
		}
	}

	return nil
}
