package panel

import (
	"context"
	"testing"
	"time"
)

func TestPanel_actions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	panel := &Panel{
		action: make(chan Action),
	}

	// no buttons pressed
	// no action
	pbs := &ButtonsState{
		homeState:      ButtonHomeRelease,
		homeLastState:  ButtonHomeRelease,
		homePushedTime: time.Now(),

		upState:      ButtonUpRelease,
		upLastState:  ButtonUpRelease,
		upPushedTime: time.Now(),

		downState:      ButtonDownRelease,
		downLastState:  ButtonDownRelease,
		downPushedTime: time.Now(),

		inState:      ButtonInRelease,
		inLastState:  ButtonInRelease,
		inPushedTime: time.Now(),

		outState:      ButtonOutRelease,
		outLastState:  ButtonOutRelease,
		outPushedTime: time.Now(),
	}
	panel.mu.Lock()
	panel.buttonsState = pbs
	panel.mu.Unlock()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(buttonStateDebounce)

				panel.mu.Lock()
				panel.buttonsState = pbs
				panel.mu.Unlock()

				panel.actionRight()
				panel.actionLeft()
			}
		}
	}()

	//
	// rigt motor
	//
	t.Run("up-push-release", func(t *testing.T) {
		// button Up pressed
		pbs.upState = ButtonUpPush
		pbs.upPushedTime = time.Now()
		action := <-panel.Action()
		if action != MoveUp {
			t.Errorf("expected action %v, got %v", MoveUp, action)
		}

		// button Up released
		pbs.upState = ButtonUpRelease
		pbs.upPushedTime = time.Now()
		action = <-panel.Action()
		if action != MoveRightStop {
			t.Errorf("expected action %v, got %v", MoveRightStop, action)
		}
	})

	t.Run("down-push-release", func(t *testing.T) {
		// button Down pressed
		pbs.downState = ButtonDownPush
		pbs.downPushedTime = time.Now()
		action := <-panel.Action()
		if action != MoveDown {
			t.Errorf("expected action %v, got %v", MoveDown, action)
		}

		// button Down released
		pbs.downState = ButtonDownRelease
		pbs.downPushedTime = time.Now()
		action = <-panel.Action()
		if action != MoveRightStop {
			t.Errorf("expected action %v, got %v", MoveRightStop, action)
		}
	})

	t.Run("up-down-push-5s-release", func(t *testing.T) {
		// button Up && Down pressed
		actionPrecissionStart := time.Now()
		pbs.upState = ButtonUpPush
		pbs.upPushedTime = time.Now()
		time.Sleep(actionDebounce / 2) // wait a little, user has a little time to match the buttons
		pbs.downState = ButtonDownPush
		pbs.downPushedTime = time.Now()
		action := <-panel.Action()
		if action != TogglePrecisionMode {
			t.Errorf("expected action %v, got %v", TogglePrecisionMode, action)
		}
		actionPrecissionTime := time.Since(actionPrecissionStart)
		if actionPrecissionTime < actionPrecisionModeActivation {
			t.Errorf("expected precision mode to be activated after %v, got %v", actionPrecisionModeActivation, actionPrecissionTime)
		}

		// button Up && Down released
		pbs.upState = ButtonUpRelease
		pbs.upPushedTime = time.Now()
		// TODO: real life a start followed by stop might be triggered here,
		// if the user is not fast enough to release the buttons
		pbs.downState = ButtonDownRelease
		pbs.downPushedTime = time.Now()
		action = <-panel.Action()
		if action != MoveRightStop {
			t.Errorf("expected action %v, got %v", MoveRightStop, action)
		}
		action = <-panel.Action()
		if action != MoveRightStop {
			t.Errorf("expected action %v, got %v", MoveRightStop, action)
		}
	})

	t.Run("up-push-down-push-release", func(t *testing.T) {
		// button Up pressed, Down pressed action disabled
		// Down release motor stop
		pbs.upState = ButtonUpPush
		pbs.upPushedTime = time.Now()
		action := <-panel.Action()
		if action != MoveUp {
			t.Errorf("expected action %v, got %v", MoveUp, action)
		}

		// no action, panel move right locked
		pbs.downState = ButtonDownPush
		pbs.downPushedTime = time.Now()
		time.Sleep(buttonStateDebounce * 2)
		pbs.downState = ButtonDownRelease
		pbs.downPushedTime = time.Now()

		pbs.upState = ButtonUpRelease
		pbs.upPushedTime = time.Now()
		action = <-panel.Action()
		if action != MoveRightStop {
			t.Errorf("expected action %v, got %v", MoveRightStop, action)
		}
		action = <-panel.Action()
		if action != MoveRightStop {
			t.Errorf("expected action %v, got %v", MoveRightStop, action)
		}
	})

	t.Run("down-push-up-push-release", func(t *testing.T) {
		// button Down pressed, Up pressed action disabled
		// Up release motor stop
		pbs.downState = ButtonDownPush
		pbs.downPushedTime = time.Now()
		action := <-panel.Action()
		if action != MoveDown {
			t.Errorf("expected action %v, got %v", MoveDown, action)
		}

		// no action, panel move right locked
		pbs.upState = ButtonUpPush
		pbs.upPushedTime = time.Now()
		time.Sleep(buttonStateDebounce * 2)
		pbs.upState = ButtonUpRelease
		pbs.upPushedTime = time.Now()

		pbs.downState = ButtonDownRelease
		pbs.downPushedTime = time.Now()
		action = <-panel.Action()
		if action != MoveRightStop {
			t.Errorf("expected action %v, got %v", MoveRightStop, action)
		}
		action = <-panel.Action()
		if action != MoveRightStop {
			t.Errorf("expected action %v, got %v", MoveRightStop, action)
		}
	})

	//
	// left motor
	//
	t.Run("in-push-release", func(t *testing.T) {
		// button In pressed
		pbs.inState = ButtonInPush
		pbs.inPushedTime = time.Now()
		action := <-panel.Action()
		if action != MoveIn {
			t.Errorf("expected action %v, got %v", MoveIn, action)
		}

		// button In released
		pbs.inState = ButtonInRelease
		pbs.inPushedTime = time.Now()
		action = <-panel.Action()
		if action != MoveLeftStop {
			t.Errorf("expected action %v, got %v", MoveLeftStop, action)
		}
	})

	t.Run("out-push-release", func(t *testing.T) {
		// button Out pressed
		pbs.outState = ButtonOutPush
		pbs.outPushedTime = time.Now()
		action := <-panel.Action()
		if action != MoveOut {
			t.Errorf("expected action %v, got %v", MoveOut, action)
		}

		// button Out released
		pbs.outState = ButtonOutRelease
		pbs.outPushedTime = time.Now()
		action = <-panel.Action()
		if action != MoveLeftStop {
			t.Errorf("expected action %v, got %v", MoveLeftStop, action)
		}
	})

	t.Run("in-out-push-10s-release", func(t *testing.T) {
		// button In && Out pressed
		actionRebootStart := time.Now()
		pbs.inState = ButtonInPush
		pbs.inPushedTime = time.Now()
		time.Sleep(actionDebounce / 2) // wait a little, user has a little time to match the buttons
		pbs.outState = ButtonOutPush
		pbs.outPushedTime = time.Now()
		action := <-panel.Action()
		if action != Reboot {
			t.Errorf("expected action %v, got %v", Reboot, action)
		}
		actionRebootTime := time.Since(actionRebootStart)
		if actionRebootTime < actionRebootActivation {
			t.Errorf("expected reboot to be activated after %v, got %v", actionRebootActivation, actionRebootTime)
		}

		// button In && Out released
		pbs.inState = ButtonInRelease
		pbs.inPushedTime = time.Now()
		// TODO: real life a start followed by stop might be triggered here,
		// if the user is not fast enough to release the buttons
		pbs.outState = ButtonOutRelease
		pbs.outPushedTime = time.Now()
		action = <-panel.Action()
		if action != MoveLeftStop {
			t.Errorf("expected action %v, got %v", MoveLeftStop, action)
		}
		action = <-panel.Action()
		if action != MoveLeftStop {
			t.Errorf("expected action %v, got %v", MoveLeftStop, action)
		}
	})

	t.Run("in-push-out-push-release", func(t *testing.T) {
		// button In pressed, Out pressed action disabled
		// Out release motor stop
		pbs.inState = ButtonInPush
		pbs.inPushedTime = time.Now()
		action := <-panel.Action()
		if action != MoveIn {
			t.Errorf("expected action %v, got %v", MoveIn, action)
		}

		// no action, panel left right locked
		pbs.outState = ButtonOutPush
		pbs.outPushedTime = time.Now()
		time.Sleep(buttonStateDebounce * 2)
		pbs.outState = ButtonOutRelease
		pbs.outPushedTime = time.Now()

		pbs.inState = ButtonInRelease
		pbs.inPushedTime = time.Now()
		action = <-panel.Action()
		if action != MoveLeftStop {
			t.Errorf("expected action %v, got %v", MoveLeftStop, action)
		}
		action = <-panel.Action()
		if action != MoveLeftStop {
			t.Errorf("expected action %v, got %v", MoveLeftStop, action)
		}
	})

	t.Run("out-push-in-push-release", func(t *testing.T) {
		// button Out pressed, In pressed action disabled
		// In release motor stop
		pbs.outState = ButtonOutPush
		pbs.outPushedTime = time.Now()
		action := <-panel.Action()
		if action != MoveOut {
			t.Errorf("expected action %v, got %v", MoveOut, action)
		}

		// no action, panel left right locked
		pbs.inState = ButtonInPush
		pbs.inPushedTime = time.Now()
		time.Sleep(buttonStateDebounce * 2)
		pbs.inState = ButtonInRelease
		pbs.inPushedTime = time.Now()

		pbs.outState = ButtonOutRelease
		pbs.outPushedTime = time.Now()
		action = <-panel.Action()
		if action != MoveLeftStop {
			t.Errorf("expected action %v, got %v", MoveLeftStop, action)
		}
		action = <-panel.Action()
		if action != MoveLeftStop {
			t.Errorf("expected action %v, got %v", MoveLeftStop, action)
		}
	})
}
