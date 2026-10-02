package winch

import (
	"context"
	"testing"
	"time"

	"github.com/CreateView/dawinchi-hw/internal/motor"
	"github.com/CreateView/dawinchi-hw/internal/panel"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type PanelSuite struct {
	suite.Suite

	upDownMotor  *motor.Mock
	inOutMotor   *motor.Mock
	buttonsPanel *panel.Mock

	firmware *Firmware
}

func TestPanelSuite(t *testing.T) {
	suite.Run(t, &PanelSuite{})
}

func (ps *PanelSuite) SetupSuite() {
}

func (ps *PanelSuite) SetupSubTest() {
	ps.upDownMotor = &motor.Mock{}
	ps.inOutMotor = &motor.Mock{}
	ps.buttonsPanel = &panel.Mock{}
	ps.firmware = InitTestFirmware(ps.T(), ps.upDownMotor, ps.inOutMotor, ps.buttonsPanel)
}

func (ps *PanelSuite) TearDownSubTest() {
	ps.firmware.hwlog.Close()
}

func (ps *PanelSuite) TearDownTest() {
	ps.upDownMotor.AssertExpectations(ps.T())
	ps.inOutMotor.AssertExpectations(ps.T())
	ps.buttonsPanel.AssertExpectations(ps.T())
}

func (ps *PanelSuite) TestFirmware_Panel() {
	ps.Run("action-UpDownMotorIncr", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		maxSpeed := uint32(100)

		ps.upDownMotor.On("Direction").Return(motor.CW)
		ps.upDownMotor.On("SetDirection", motor.CW).Return(nil)
		ps.upDownMotor.On("AccelerateLocalAction", ctx, maxSpeed).Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action:   panel.UpDownMotorIncr,
			MaxSpeed: maxSpeed,
			Ctx:      ctx,
		}

		time.Sleep(1 * time.Second)
		ps.Equal(false, ps.firmware.remoteLock)
		ps.Equal(true, ps.firmware.localCommandLock)
	})

	ps.Run("action-UpDownMotorDecr", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		maxSpeed := uint32(100)

		ps.upDownMotor.On("Direction").Return(motor.CW)
		ps.upDownMotor.On("Stop").Return(nil)

		ps.upDownMotor.On("SetDirection", motor.CCW).Return(nil)
		ps.upDownMotor.On("AccelerateLocalAction", ctx, maxSpeed).Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action:   panel.UpDownMotorDecr,
			MaxSpeed: maxSpeed,
			Ctx:      ctx,
		}

		time.Sleep(1 * time.Second)
		ps.Equal(false, ps.firmware.remoteLock)
		ps.Equal(true, ps.firmware.localCommandLock)
	})

	ps.Run("action-UpDownMotorStop", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		ps.upDownMotor.On("Stop").Return(nil)
		ps.upDownMotor.On("SetSpeed", uint32(0)).Return(nil)
		ps.upDownMotor.On("EnableBrake").Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action:   panel.UpDownMotorStop,
			MaxSpeed: 100,
			Ctx:      ctx,
		}

		time.Sleep(1 * time.Second)
		ps.Equal(false, ps.firmware.remoteLock)
		ps.Equal(false, ps.firmware.localCommandLock)
	})

	ps.Run("action-InOutMotorIncr", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		maxSpeed := uint32(100)

		ps.inOutMotor.On("Direction").Return(motor.CW)
		ps.inOutMotor.On("SetDirection", motor.CW).Return(nil)
		ps.inOutMotor.On("AccelerateLocalAction", mock.Anything, maxSpeed).Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		ps.inOutMotor.On("Direction").Return(motor.CW)
		ps.upDownMotor.On("SetDirection", motor.CW).Return(nil)

		upDownMaxSpeed := KeepUpDownLevelSpeed(maxSpeed, motor.CW)
		ps.upDownMotor.On("AccelerateLocalAction", mock.Anything, upDownMaxSpeed).Return(nil)

		inOutCounter := int32(100)
		upDownCounter := int32(100)
		ps.inOutMotor.On("ReadCounter").Return(inOutCounter)
		ps.upDownMotor.On("ReadCounter").Return(upDownCounter)

		ps.upDownMotor.On("SetKeepUpDownLevelWithInOutMove", mock.Anything).Return()
		ps.inOutMotor.On("Speed").Return(maxSpeed)
		ps.upDownMotor.On("Speed").Return(upDownMaxSpeed)

		ps.upDownMotor.On("SetSpeed", mock.Anything, mock.Anything).Return(nil)

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action:   panel.InOutMotorIncr,
			MaxSpeed: maxSpeed,
			Ctx:      ctx,
		}

		time.Sleep(1 * time.Second)
		ps.Equal(false, ps.firmware.remoteLock)
		ps.Equal(true, ps.firmware.localCommandLock)
	})

	ps.Run("action-InOutMotorDecr", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		maxSpeed := uint32(100)

		ps.inOutMotor.On("Direction").Return(motor.CW)
		ps.inOutMotor.On("Stop").Return(nil)

		ps.inOutMotor.On("SetDirection", motor.CCW).Return(nil)
		ps.inOutMotor.On("AccelerateLocalAction", mock.Anything, maxSpeed).Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		ps.upDownMotor.On("SetDirection", motor.CCW).Return(nil)

		upDownMaxSpeed := KeepUpDownLevelSpeed(maxSpeed, motor.CCW)
		ps.upDownMotor.On("AccelerateLocalAction", mock.Anything, upDownMaxSpeed).Return(nil)

		inOutCounter := int32(100)
		upDownCounter := int32(100)
		ps.inOutMotor.On("ReadCounter").Return(inOutCounter)
		ps.upDownMotor.On("ReadCounter").Return(upDownCounter)

		ps.upDownMotor.On("SetKeepUpDownLevelWithInOutMove", mock.Anything).Return()

		ps.inOutMotor.On("Speed").Return(maxSpeed)
		ps.upDownMotor.On("Speed").Return(upDownMaxSpeed)

		ps.upDownMotor.On("SetSpeed", mock.Anything, mock.Anything).Return(nil)

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action:   panel.InOutMotorDecr,
			MaxSpeed: maxSpeed,
			Ctx:      ctx,
		}

		// reset direction
		time.Sleep(10 * time.Millisecond)
		for i, call := range ps.inOutMotor.ExpectedCalls {
			if call.Method == "Direction" {
				ps.inOutMotor.ExpectedCalls[i].ReturnArguments[0] = motor.CCW
			}
		}

		time.Sleep(1 * time.Second)
		ps.Equal(false, ps.firmware.remoteLock)
		ps.Equal(true, ps.firmware.localCommandLock)
	})

	ps.Run("action-InOutMotorDecr-keepUpDownLevel-false", func() {
		ps.firmware.keepUpDownLevel = false

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		maxSpeed := uint32(100)

		ps.inOutMotor.On("Direction").Return(motor.CW)
		ps.inOutMotor.On("Stop").Return(nil)

		ps.inOutMotor.On("SetDirection", motor.CCW).Return(nil)
		ps.inOutMotor.On("AccelerateLocalAction", mock.Anything, maxSpeed).Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		ps.upDownMotor.On("SetDirection", motor.CCW).Return(nil)

		inOutCounter := int32(100)
		upDownCounter := int32(100)
		ps.inOutMotor.On("ReadCounter").Return(inOutCounter)
		ps.upDownMotor.On("ReadCounter").Return(upDownCounter)

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action:   panel.InOutMotorDecr,
			MaxSpeed: maxSpeed,
			Ctx:      ctx,
		}

		// reset direction
		time.Sleep(10 * time.Millisecond)
		for i, call := range ps.inOutMotor.ExpectedCalls {
			if call.Method == "Direction" {
				ps.inOutMotor.ExpectedCalls[i].ReturnArguments[0] = motor.CCW
			}
		}

		time.Sleep(1 * time.Second)
		ps.Equal(false, ps.firmware.remoteLock)
		ps.Equal(true, ps.firmware.localCommandLock)

		ps.firmware.keepUpDownLevel = true
	})

	ps.Run("action-InOutMotorStop", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		ps.inOutMotor.On("Stop").Return(nil)
		ps.inOutMotor.On("SetSpeed", uint32(0)).Return(nil)
		ps.inOutMotor.On("EnableBrake").Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		ps.upDownMotor.On("Stop").Return(nil)

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action:   panel.InOutMotorStop,
			MaxSpeed: 100,
			Ctx:      ctx,
		}

		time.Sleep(1 * time.Second)
		ps.Equal(false, ps.firmware.remoteLock)
		ps.Equal(false, ps.firmware.localCommandLock)
	})

	ps.Run("action-SetHome", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		upDownMotorCounter := int32(1000)
		inOutMotorCounter := int32(1000)

		ps.upDownMotor.On("ReadCounter").Return(upDownMotorCounter)
		ps.inOutMotor.On("ReadCounter").Return(inOutMotorCounter)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		// TOOD: remove it, only for debug
		ps.buttonsPanel.On("SignalError").Return(nil)

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action:   panel.SetHome,
			MaxSpeed: 100,
			Ctx:      ctx,
		}

		time.Sleep(1 * time.Second)

		expectedHomePosition := &HomePosition{
			UpDownMotorCounter: upDownMotorCounter,
			InOutMotorCounter:  inOutMotorCounter,
		}
		ps.Equal(expectedHomePosition, ps.firmware.homePosition)
		ps.Equal(false, ps.firmware.remoteLock)
		ps.Equal(false, ps.firmware.localCommandLock)
	})

	ps.Run("action-GoHome", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		ps.firmware.homePosition = &HomePosition{
			UpDownMotorCounter: 1000,
			InOutMotorCounter:  1000,
		}

		speed := uint32(65)

		upDownMotorCounter := ps.firmware.homePosition.UpDownMotorCounter
		inOutMotorCounter := ps.firmware.homePosition.InOutMotorCounter

		ps.upDownMotor.On("MoveTo", mock.Anything, upDownMotorCounter, speed).Return(nil)
		ps.inOutMotor.On("MoveTo", mock.Anything, inOutMotorCounter, speed).Return(nil)

		ps.inOutMotor.On("Direction").Return(motor.CW)
		ps.upDownMotor.On("SetDirection", motor.CW).Return(nil)
		upDownMaxSpeed := KeepUpDownLevelSpeed(speed, motor.CW)
		ps.upDownMotor.On("AccelerateWithRamp", mock.Anything, upDownMaxSpeed).Return(nil)

		ps.upDownMotor.On("SetKeepUpDownLevelWithInOutMove", mock.Anything).Return()
		ps.inOutMotor.On("Speed").Return(speed)

		ps.inOutMotor.On("ReadCounter").Return(int32(100))
		ps.upDownMotor.On("ReadCounter").Return(int32(100))
		ps.upDownMotor.On("Stop").Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action:   panel.GoHome,
			MaxSpeed: speed,
			Ctx:      ctx,
		}

		time.Sleep(1 * time.Second)
		ps.Equal(false, ps.firmware.remoteLock)
		ps.Equal(false, ps.firmware.localCommandLock)
	})

	// CCP
	ps.Run("action-CcpUpDownMotorIncr", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		maxSpeed := uint32(100)

		ps.upDownMotor.On("Direction").Return(motor.CW)

		ps.upDownMotor.On("SetDirection", motor.CW).Return(nil)
		ps.upDownMotor.On("AccelerateWithRamp", mock.Anything, maxSpeed).Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action:   panel.CcpUpDownMotorIncr,
			MaxSpeed: maxSpeed,
			Ctx:      ctx,
		}
		time.Sleep(1 * time.Second)
		ps.Equal(false, ps.firmware.remoteLock)
		ps.Equal(false, ps.firmware.localCommandLock)
	})

	ps.Run("action-CcpUpDownMotorDecr", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		maxSpeed := uint32(100)

		ps.upDownMotor.On("Direction").Return(motor.CW)
		ps.upDownMotor.On("Stop").Return(nil)

		ps.upDownMotor.On("SetDirection", motor.CCW).Return(nil)
		ps.upDownMotor.On("AccelerateWithRamp", ctx, maxSpeed).Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action:   panel.CcpUpDownMotorDecr,
			MaxSpeed: maxSpeed,
			Ctx:      ctx,
		}

		time.Sleep(1 * time.Second)
		ps.Equal(false, ps.firmware.remoteLock)
		ps.Equal(false, ps.firmware.localCommandLock)
	})

	ps.Run("action-CcpUpDownMotorStop", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		ps.upDownMotor.On("Stop").Return(nil)
		ps.upDownMotor.On("SetSpeed", uint32(0)).Return(nil)
		ps.upDownMotor.On("EnableBrake").Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action:   panel.CcpUpDownMotorStop,
			MaxSpeed: 100,
			Ctx:      ctx,
		}

		time.Sleep(1 * time.Second)
		ps.Equal(false, ps.firmware.remoteLock)
		ps.Equal(false, ps.firmware.localCommandLock)
	})

	ps.Run("action-CcpInOutMotorIncr", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		maxSpeed := uint32(100)

		ps.inOutMotor.On("Direction").Return(motor.CW)

		ps.inOutMotor.On("SetDirection", motor.CW).Return(nil)
		ps.inOutMotor.On("AccelerateWithRamp", mock.Anything, maxSpeed).Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		ps.inOutMotor.On("Direction").Return(motor.CW)
		ps.upDownMotor.On("SetDirection", motor.CW).Return(nil)

		upDownMaxSpeed := KeepUpDownLevelSpeed(maxSpeed, motor.CW)
		ps.upDownMotor.On("AccelerateWithRamp", mock.Anything, upDownMaxSpeed).Return(nil)

		inOutCounter := int32(100)
		upDownCounter := int32(100)
		ps.inOutMotor.On("ReadCounter").Return(inOutCounter)
		ps.upDownMotor.On("ReadCounter").Return(upDownCounter)

		ps.upDownMotor.On("SetKeepUpDownLevelWithInOutMove", mock.Anything).Return()
		ps.inOutMotor.On("Speed").Return(maxSpeed)
		ps.upDownMotor.On("Speed").Return(upDownMaxSpeed)

		ps.upDownMotor.On("SetSpeed", mock.Anything, mock.Anything).Return(nil)

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action:   panel.CcpInOutMotorIncr,
			MaxSpeed: 100,
			Ctx:      ctx,
		}
		time.Sleep(1 * time.Second)
		ps.Equal(false, ps.firmware.remoteLock)
		ps.Equal(false, ps.firmware.localCommandLock)
	})

	ps.Run("action-CcpInOutMotorIncr-keepUpDownLevel-false", func() {
		ps.firmware.keepUpDownLevel = false

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		maxSpeed := uint32(100)

		ps.inOutMotor.On("Direction").Return(motor.CW)

		ps.inOutMotor.On("SetDirection", motor.CW).Return(nil)
		ps.inOutMotor.On("AccelerateWithRamp", mock.Anything, maxSpeed).Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		ps.inOutMotor.On("Direction").Return(motor.CW)
		ps.upDownMotor.On("SetDirection", motor.CW).Return(nil)

		inOutCounter := int32(100)
		upDownCounter := int32(100)
		ps.inOutMotor.On("ReadCounter").Return(inOutCounter)
		ps.upDownMotor.On("ReadCounter").Return(upDownCounter)

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action:   panel.CcpInOutMotorIncr,
			MaxSpeed: 100,
			Ctx:      ctx,
		}
		time.Sleep(1 * time.Second)
		ps.Equal(false, ps.firmware.remoteLock)
		ps.Equal(false, ps.firmware.localCommandLock)

		ps.firmware.keepUpDownLevel = true
	})

	ps.Run("action-CcpInOutMotorDecr", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		maxSpeed := uint32(100)

		ps.inOutMotor.On("Direction").Return(motor.CW)
		ps.inOutMotor.On("Stop").Return(nil)

		ps.inOutMotor.On("SetDirection", motor.CCW).Return(nil)
		ps.inOutMotor.On("AccelerateWithRamp", mock.Anything, maxSpeed).Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		ps.inOutMotor.On("Direction").Return(motor.CCW)
		ps.upDownMotor.On("SetDirection", motor.CCW).Return(nil)

		upDownMaxSpeed := KeepUpDownLevelSpeed(maxSpeed, motor.CCW)
		ps.upDownMotor.On("AccelerateWithRamp", mock.Anything, upDownMaxSpeed).Return(nil)

		inOutCounter := int32(100)
		upDownCounter := int32(100)
		ps.inOutMotor.On("ReadCounter").Return(inOutCounter)
		ps.upDownMotor.On("ReadCounter").Return(upDownCounter)

		ps.upDownMotor.On("SetKeepUpDownLevelWithInOutMove", mock.Anything).Return()
		ps.inOutMotor.On("Speed").Return(maxSpeed)
		ps.upDownMotor.On("Speed").Return(upDownMaxSpeed)

		ps.upDownMotor.On("SetSpeed", mock.Anything, mock.Anything).Return(nil)
		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action:   panel.CcpInOutMotorDecr,
			MaxSpeed: maxSpeed,
			Ctx:      ctx,
		}

		// reset direction
		time.Sleep(10 * time.Millisecond)
		for i, call := range ps.inOutMotor.ExpectedCalls {
			if call.Method == "Direction" {
				ps.inOutMotor.ExpectedCalls[i].ReturnArguments[0] = motor.CCW
			}
		}

		time.Sleep(1 * time.Second)
		ps.Equal(false, ps.firmware.remoteLock)
		ps.Equal(false, ps.firmware.localCommandLock)
	})

	ps.Run("action-CcpInOutMotorDecr-keepUpDownLevel-false", func() {
		ps.firmware.keepUpDownLevel = false

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		maxSpeed := uint32(100)

		ps.inOutMotor.On("Direction").Return(motor.CW)
		ps.inOutMotor.On("Stop").Return(nil)

		ps.inOutMotor.On("SetDirection", motor.CCW).Return(nil)
		ps.inOutMotor.On("AccelerateWithRamp", mock.Anything, maxSpeed).Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		ps.inOutMotor.On("Direction").Return(motor.CCW)
		ps.upDownMotor.On("SetDirection", motor.CCW).Return(nil)

		inOutCounter := int32(100)
		upDownCounter := int32(100)
		ps.inOutMotor.On("ReadCounter").Return(inOutCounter)
		ps.upDownMotor.On("ReadCounter").Return(upDownCounter)

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action:   panel.CcpInOutMotorDecr,
			MaxSpeed: maxSpeed,
			Ctx:      ctx,
		}

		// reset direction
		time.Sleep(10 * time.Millisecond)
		for i, call := range ps.inOutMotor.ExpectedCalls {
			if call.Method == "Direction" {
				ps.inOutMotor.ExpectedCalls[i].ReturnArguments[0] = motor.CCW
			}
		}

		time.Sleep(1 * time.Second)
		ps.Equal(false, ps.firmware.remoteLock)
		ps.Equal(false, ps.firmware.localCommandLock)

		ps.firmware.keepUpDownLevel = true
	})

	ps.Run("action-CcpInOutMotorStop", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		ps.inOutMotor.On("Stop").Return(nil)
		ps.inOutMotor.On("SetSpeed", uint32(0)).Return(nil)
		ps.inOutMotor.On("EnableBrake").Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		ps.upDownMotor.On("Stop").Return(nil)

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action:   panel.CcpInOutMotorStop,
			MaxSpeed: 100,
			Ctx:      ctx,
		}

		time.Sleep(1 * time.Second)
		ps.Equal(false, ps.firmware.remoteLock)
		ps.Equal(false, ps.firmware.localCommandLock)
	})

	ps.Run("action-TogglePrecisionMode", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ps.firmware.precisionMode = false

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		// TOOD: remove it, only for debug
		ps.buttonsPanel.On("SignalError").Return(nil)

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action: panel.TogglePrecisionMode,
			Ctx:    ctx,
		}

		time.Sleep(1 * time.Second)
		ps.Equal(true, ps.firmware.precisionMode)
		ps.Equal(false, ps.firmware.localCommandLock)
	})

	ps.Run("action-ToggleRemoteLock", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		ps.buttonsPanel.On("ClearRemoteLocked").Return()
		ps.buttonsPanel.On("ShowRemoteLocked").Return().Once()

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action: panel.ToggleRemoteLock,
			Ctx:    ctx,
		}

		time.Sleep(1 * time.Second)
		ps.Equal(true, ps.firmware.remoteLock)
		ps.Equal(false, ps.firmware.localCommandLock)
	})

	ps.Run("action-ToggleServiceMode", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		ps.buttonsPanel.On("ClearError").Return()
		ps.buttonsPanel.On("ShowError").Return().Once()

		ps.upDownMotor.On("TurnOn").Return()
		ps.inOutMotor.On("TurnOn").Return()

		ps.upDownMotor.On("EnableBrake").Return(nil)
		ps.inOutMotor.On("EnableBrake").Return(nil)

		ps.upDownMotor.On("TurnOff").Return()
		ps.inOutMotor.On("TurnOff").Return()

		ps.upDownMotor.On("Stop").Return(nil)
		ps.inOutMotor.On("Stop").Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action: panel.ToggleServiceMode,
			Ctx:    ctx,
		}

		time.Sleep(2 * time.Second)
		ps.Equal(true, ps.firmware.serviceMode)
		ps.Equal(true, ps.firmware.remoteLock)
		ps.Equal(false, ps.firmware.localCommandLock)
	})

	ps.Run("action-Reboot", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		ps.upDownMotor.On("Stop").Return(nil)
		ps.inOutMotor.On("Stop").Return(nil)

		ps.upDownMotor.On("EnableBrake").Return(nil)
		ps.inOutMotor.On("EnableBrake").Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action: panel.Reboot,
			Ctx:    ctx,
		}

		// wait for reboot
		time.Sleep(3 * time.Second)
	})

	ps.Run("action-EnableFactoryReset", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		action := make(chan panel.ButtonAction)
		ps.buttonsPanel.On("Action").Return(
			action,
		)

		ps.upDownMotor.On("Stop").Return(nil)
		ps.inOutMotor.On("Stop").Return(nil)

		ps.upDownMotor.On("EnableBrake").Return(nil)
		ps.inOutMotor.On("EnableBrake").Return(nil)

		ps.upDownMotor.On("State").Return(&motor.State{})
		ps.inOutMotor.On("State").Return(&motor.State{})
		ps.buttonsPanel.On("State").Return(&panel.State{})

		// start listening to buttons panel actions
		ps.firmware.Panel(ctx)

		// send button action
		action <- panel.ButtonAction{
			Action: panel.InitFactoryReset,
			Ctx:    ctx,
		}

		// wait for factory reset
		time.Sleep(1 * time.Second)
		ps.Equal(true, ps.firmware.factoryResetInitiated)
		ps.Equal(true, ps.firmware.remoteLock)
	})
}
