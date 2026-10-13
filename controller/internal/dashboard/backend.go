package dashboard

import (
	"fmt"
	"math"
	"time"

	"depthpilot/controller/cs2rs"
	"github.com/goburrow/modbus"
)

type reading struct {
	Position        float64
	Running         bool
	Fault           string
	Completed       bool
	ProfileCounter  int32
	FeedbackCounter int32
}
type motor interface {
	read(time.Time) (reading, error)
	move(float64, float64) error
	jog(int, float64) error
	refresh() error
	stop() error
	close()
}

type simulator struct {
	position, target, speed float64
	direction               int
	last                    time.Time
	config                  AxisConfig
}

func (s *simulator) read(now time.Time) (reading, error) {
	dt := now.Sub(s.last).Seconds()
	s.last = now
	if s.direction != 0 {
		s.position += float64(s.direction) * s.speed * dt
	} else {
		d := s.target - s.position
		s.position += math.Copysign(math.Min(math.Abs(d), s.speed*dt), d)
	}
	running := s.direction != 0 || math.Abs(s.target-s.position) > 1e-7
	return reading{Position: s.position, Running: running, Completed: !running, ProfileCounter: int32(math.Round(s.position * s.config.PulsesPerMetre * float64(s.config.Direction))), FeedbackCounter: int32(math.Round(s.position * s.config.FeedbackCountsPerMetre * float64(s.config.Direction)))}, nil
}
func (s *simulator) move(p, v float64) error    { s.target = p; s.speed = v; s.direction = 0; return nil }
func (s *simulator) jog(d int, v float64) error { s.direction = d; s.speed = v; return nil }
func (s *simulator) refresh() error             { return nil }
func (s *simulator) stop() error                { s.direction = 0; s.target = s.position; return nil }
func (s *simulator) close()                     {}

type hardware struct {
	drive     *cs2rs.Drive
	handler   *modbus.RTUClientHandler
	config    AxisConfig
	positive  bool
	jogging   bool
	heartbeat func() error
}

func newHardware(c AxisConfig) (motor, error) {
	if c.Port == "" || c.Slave < 1 || c.Slave > 127 || c.Baud <= 0 {
		return nil, fmt.Errorf("port, slave 1..127 and baud required")
	}
	h := modbus.NewRTUClientHandler(c.Port)
	h.BaudRate = c.Baud
	h.DataBits = 8
	h.Parity = "N"
	h.StopBits = 1
	h.SlaveId = c.Slave
	h.Timeout = 150 * time.Millisecond
	if err := h.Connect(); err != nil {
		return nil, err
	}
	return &hardware{drive: cs2rs.New(modbus.NewClient(h)), handler: h, config: c}, nil
}
func (h *hardware) read(_ time.Time) (reading, error) {
	var s cs2rs.Status
	var err error
	if h.jogging {
		s, err = h.drive.StatusDuringJog(h.positive)
	} else if h.heartbeat != nil {
		s, err = h.drive.StatusWithHeartbeat(h.heartbeat)
	} else {
		s, err = h.drive.Status()
	}
	if err != nil {
		return reading{}, err
	}
	r := reading{Position: float64(s.ActualPosition) / h.config.FeedbackCountsPerMetre * float64(h.config.Direction), Running: s.Running || s.Trigger > 15}
	r.Completed = !s.Running && s.Trigger == 0 && s.PathComplete
	r.ProfileCounter, r.FeedbackCounter = s.ProfilePosition, s.ActualPosition
	if s.Fault || s.Alarm != 0 || s.PRAlarm != 0 || !s.Enabled {
		r.Fault = fmt.Sprintf("drive unavailable: enabled=%t alarm=%04x PR=%04x", s.Enabled, s.Alarm, s.PRAlarm)
	}
	return r, nil
}
func (h *hardware) rpm(v float64) (uint16, error) {
	r := math.Round(v * h.config.PulsesPerMetre * 60 / h.config.PulsesPerRevolution)
	if r < 1 || r > 5000 {
		return 0, fmt.Errorf("speed outside motor RPM range")
	}
	return uint16(r), nil
}
func (h *hardware) move(p, v float64) error {
	r, err := h.rpm(v)
	if err != nil {
		return err
	}
	return h.drive.Start(cs2rs.Move{Position: int32(math.Round(p * h.config.PulsesPerMetre * float64(h.config.Direction))), RPM: r, Acceleration: h.config.Ramp, Deceleration: h.config.Ramp})
}
func (h *hardware) jog(d int, v float64) error {
	r, err := h.rpm(v)
	if err != nil {
		return err
	}
	h.positive = d*h.config.Direction > 0
	if err := h.drive.PrepareJog(r, h.config.Ramp); err != nil {
		return err
	}
	h.jogging = true
	return nil
}
func (h *hardware) refresh() error { return h.drive.RefreshJog(h.positive) }
func (h *hardware) stop() error    { h.jogging = false; return h.drive.Stop() }
func (h *hardware) close()         { h.handler.Close() }
