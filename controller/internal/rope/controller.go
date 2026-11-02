package rope

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"depthpilot/controller/cs2rs"
)

type State struct {
	Mode          string          `json:"mode"`
	Phase         string          `json:"phase"`
	Referenced    bool            `json:"referenced"`
	Position      Point           `json:"position"`
	Target        Point           `json:"target"`
	Drums         [2]cs2rs.Status `json:"drums"`
	Plan          *Plan           `json:"plan,omitempty"`
	Error         string          `json:"error"`
	Updated       time.Time       `json:"updated"`
	TriggerSkewMS float64         `json:"triggerSkewMs"`
	Events        []string        `json:"events"`
	LastMove      *MoveTrace      `json:"lastMove,omitempty"`
	StartMode     string          `json:"startMode"`
}
type Controller struct {
	mu                                sync.Mutex
	bus                               Bus
	state                             State
	commandOrigin, feedbackOrigin     [2]int32
	idleCounters                      [2]int32
	readyAt, deadline, settleAt       time.Time
	disableNotBefore, disableDeadline time.Time
	enabledByUs                       bool
	started                           [2]bool
	closed                            bool
	lastSamples                       [2]RawSample
}

func New(bus Bus, simulation bool) (*Controller, error) {
	return NewWithStartMode(bus, simulation, false)
}

func NewWithStartMode(bus Bus, simulation, broadcast bool) (*Controller, error) {
	mode := "hardware · NO ROPE"
	if simulation {
		mode = "simulation"
	}
	c := &Controller{bus: bus, state: State{Mode: mode, Phase: "unreferenced", Events: []string{}}}
	c.state.StartMode = "sequential"
	if broadcast {
		c.state.StartMode = "broadcast"
	}
	if err := c.inspect(); err != nil {
		return nil, err
	}
	c.event("Ready for unloaded commissioning; capture a virtual origin before moving")
	return c, nil
}
func (c *Controller) event(s string) {
	c.state.Events = append(c.state.Events, time.Now().Format("15:04:05")+" · "+s)
	if len(c.state.Events) > 30 {
		c.state.Events = c.state.Events[len(c.state.Events)-30:]
	}
}
func (c *Controller) inspect() error {
	defer func() {
		if c.state.LastMove != nil {
			c.state.LastMove.align()
		}
	}()
	for i := 0; i < 2; i++ {
		var sample RawSample
		s, e := sampledStatus(c.bus, byte(i+1), &sample)
		if e != nil {
			return fmt.Errorf("drum %d: %w", i+1, e)
		}
		c.state.Drums[i] = s
		c.lastSamples[i] = sample
		if c.state.LastMove != nil {
			c.state.LastMove.add(i, sample)
		}
		if s.Fault || s.Alarm != 0 || s.PRAlarm != 0 {
			return fmt.Errorf("drum %d fault: alarm=%04x PR=%04x", i+1, s.Alarm, s.PRAlarm)
		}
		if c.state.Referenced && c.state.Phase == "idle" && s.ProfilePosition != c.idleCounters[i] {
			return fmt.Errorf("drum %d counter changed outside commanded motion; recapture origin", i+1)
		}
	}
	c.state.Updated = time.Now()
	if c.state.Referenced {
		c.state.Position = Coordinates([2]int32{c.state.Drums[0].ActualPosition, c.state.Drums[1].ActualPosition}, c.feedbackOrigin)
	}
	return nil
}
func (c *Controller) verifyConfig() error {
	for i := byte(1); i <= 2; i++ {
		checks := map[uint16]uint16{0x0001: 10000, 0x0233: 4000}
		if i == 1 {
			checks[0x0145] = 8
			checks[0x0157] = 0x24
			checks[0x0167] = 250
			checks[0x0169] = 250
			checks[0x0191] = 60
		} else {
			checks[0x0191] = 40
		}
		for reg, want := range checks {
			v, e := c.bus.Read(i, reg, 1)
			if e != nil {
				return e
			}
			if v[0] != want {
				return fmt.Errorf("drum %d register %04x=%d, expected %d", i, reg, v[0], want)
			}
		}
	}
	return nil
}
func (c *Controller) Snapshot() State {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := c.state
	s.Events = append([]string(nil), s.Events...)
	s.LastMove = s.LastMove.clone()
	if s.Plan != nil {
		p := *s.Plan
		s.Plan = &p
	}
	return s
}
func (c *Controller) Reference() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return fmt.Errorf("controller closed")
	}
	if c.state.Phase != "idle" && c.state.Phase != "unreferenced" && c.state.Phase != "fault" {
		return fmt.Errorf("stop before capturing an origin")
	}
	if c.enabledByUs {
		return fmt.Errorf("wait for brake/disable sequence")
	}
	if e := c.inspect(); e != nil {
		return e
	}
	if e := c.verifyConfig(); e != nil {
		return e
	}
	for i, s := range c.state.Drums {
		if s.Running || (s.Enabled && s.Trigger > 15) {
			return fmt.Errorf("drum %d not idle", i+1)
		}
		c.commandOrigin[i] = s.ProfilePosition
		c.feedbackOrigin[i] = s.ActualPosition
		c.idleCounters[i] = s.ProfilePosition
	}
	c.state.Referenced = true
	c.state.Position = Point{}
	c.state.Target = Point{}
	c.state.Plan = nil
	c.state.Error = ""
	c.state.Phase = "idle"
	c.event("Temporary virtual origin captured; no physical homing or EEPROM writes")
	return nil
}
func (c *Controller) Move(target Point, speed float64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed || !c.state.Referenced || c.state.Phase != "idle" {
		return fmt.Errorf("capture origin and wait for idle before moving")
	}
	if e := c.inspect(); e != nil {
		c.fail(e)
		return e
	}
	if e := c.verifyConfig(); e != nil {
		c.fail(e)
		return e
	}
	current := [2]int32{}
	for i, s := range c.state.Drums {
		if s.Running || (s.Enabled && s.Trigger > 15) {
			return fmt.Errorf("drive already moving")
		}
		current[i] = s.ProfilePosition
	}
	// Use the command frame for exact cancellation (dx+dz=0). Feedback remains
	// independently displayed and checked; it is not silently used as a new datum.
	from := Coordinates(current, c.commandOrigin)
	p, e := makePlan(target, from, speed, c.commandOrigin, current, &c.state.Target)
	if e != nil {
		return e
	}
	if !c.state.Drums[1].Enabled {
		return fmt.Errorf("drum 2 must already be enabled with its brake powered")
	}
	forced, e := c.bus.Read(1, 0x000f, 1)
	if e != nil {
		return e
	}
	if forced[0] != 0 || c.state.Drums[0].Enabled {
		return fmt.Errorf("drum 1 must start disabled with software enable zero")
	}
	c.state.Plan = &p
	c.state.LastMove = newTrace(from, target, c.feedbackOrigin, c.lastSamples)
	c.state.Target = target
	c.started = [2]bool{}
	c.settleAt = time.Time{}
	c.state.TriggerSkewMS = 0
	if p.Profiles[0].Delta != 0 {
		c.enabledByUs = true // Also restore after an ambiguous enable acknowledgement.
		// Disabled CS2RS firmware can retain a pending path/stop value. Replace it
		// with quick-stop BEFORE enabling so no old path can resume on brake release.
		if c.state.Drums[0].Trigger > 15 {
			if e = c.bus.Write(1, 0x6002, 0x40); e != nil {
				c.fail(e)
				return e
			}
		}
		if e = c.bus.Write(1, 0x000f, 1); e != nil {
			c.fail(e)
			return e
		}
	}
	c.state.Phase = "preparing"
	c.readyAt = time.Now().Add(2 * time.Second)
	c.deadline = c.readyAt.Add(time.Duration((p.Seconds + 5) * float64(time.Second)))
	if c.enabledByUs {
		c.event(fmt.Sprintf("Move queued: x=%.3f z=%.3f m; waiting for automatic brake release", target.X, target.Z))
	} else {
		c.event(fmt.Sprintf("Move queued: x=%.3f z=%.3f m; drum 1 remains disabled with brake engaged", target.X, target.Z))
	}
	return nil
}
func (c *Controller) stopWrites() error {
	var es []error
	for sl := byte(1); sl <= 2; sl++ {
		if e := c.bus.Write(sl, 0x6002, 0x40); e != nil {
			es = append(es, fmt.Errorf("drum %d stop: %w", sl, e))
		}
	}
	return errors.Join(es...)
}
func (c *Controller) fail(e error) {
	if t := c.state.LastMove; t != nil && t.EndedAt == nil {
		t.Outcome = "fault"
	}
	c.state.Error = e.Error()
	c.state.Referenced = false
	c.state.Phase = "stopping"
	c.settleAt = time.Time{}
	if se := c.stopWrites(); se != nil {
		c.state.Error += "; " + se.Error()
	}
	c.event(c.state.Error)
}
func (c *Controller) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return fmt.Errorf("controller closed")
	}
	e := c.stopWrites()
	if t := c.state.LastMove; t != nil && t.EndedAt == nil && t.Outcome != "fault" {
		t.Outcome = "stopped"
	}
	c.state.Phase = "stopping"
	c.settleAt = time.Time{}
	// A partial/interrupted trajectory requires a deliberate new virtual origin.
	c.state.Referenced = false
	if e != nil {
		c.state.Error = e.Error()
	}
	c.event("Stop requested on both drives; capture a new origin after rest")
	return e
}
func (c *Controller) beginDisable(now time.Time) error {
	if e := c.bus.Write(1, 0x000f, 0); e != nil {
		return e
	}
	// A write acknowledgement confirms receipt, not completion of the drive's
	// brake/disable sequence. Keep servicing telemetry and the stop endpoint.
	c.state.Phase = "disabling"
	c.disableNotBefore = now.Add(300 * time.Millisecond)
	c.disableDeadline = now.Add(2 * time.Second)
	c.event("Drum 1 disable requested; waiting for enable bit and brake output to clear")
	return nil
}

func (c *Controller) finishRest() {
	if t := c.state.LastMove; t != nil && t.EndedAt == nil {
		outcome := t.Outcome
		if c.state.Error != "" {
			outcome = "fault"
		} else if outcome != "stopped" && outcome != "fault" {
			outcome = "completed"
		}
		t.finish(outcome, time.Now())
	}
	for i, s := range c.state.Drums {
		c.idleCounters[i] = s.ProfilePosition
	}
	if c.state.Error != "" {
		c.state.Phase = "fault"
	} else if c.state.Referenced {
		c.state.Phase = "idle"
	} else {
		c.state.Phase = "unreferenced"
	}
	c.event("At rest; drum 1 software enable off / automatic brake requested")
}
func (c *Controller) tick(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	if e := c.inspect(); e != nil {
		if c.state.Phase != "fault" && c.state.Phase != "stopping" {
			c.fail(e)
		} else {
			c.state.Error = e.Error()
		}
		return
	}
	phase := c.state.Phase
	if phase == "preparing" || phase == "moving" {
		if now.After(c.deadline) {
			c.fail(fmt.Errorf("motion deadline exceeded"))
			return
		}
		for i, s := range c.state.Drums {
			if c.state.Plan.Profiles[i].Delta != 0 && !s.Enabled {
				c.fail(fmt.Errorf("drum %d disabled during motion", i+1))
				return
			}
			if i == 0 && c.state.Plan.Profiles[0].Delta == 0 && s.Enabled {
				c.fail(fmt.Errorf("stationary drum 1 unexpectedly enabled"))
				return
			}
		}
	}
	if phase == "preparing" && now.After(c.readyAt) {
		for _, s := range c.state.Drums {
			if s.Running || (s.Enabled && s.Trigger > 15) {
				c.fail(fmt.Errorf("unexpected motion before trigger"))
				return
			}
		}
		for i, m := range c.state.Plan.Profiles {
			if m.Delta == 0 {
				continue
			}
			p := uint32(m.Target)
			words := []uint16{1, uint16(p >> 16), uint16(p), m.RPM, m.Ramp, m.Ramp, 0}
			if e := c.bus.Program(byte(i+1), words); e != nil {
				c.fail(e)
				return
			}
			v, e := c.bus.Read(byte(i+1), 0x6200, 7)
			if e != nil {
				c.fail(e)
				return
			}
			for j := range words {
				if v[j] != words[j] {
					c.fail(fmt.Errorf("path readback mismatch"))
					return
				}
			}
		}
		if c.state.StartMode == "broadcast" && c.state.Plan.Profiles[0].Delta != 0 && c.state.Plan.Profiles[1].Delta != 0 {
			n := time.Now()
			c.state.LastMove.TriggerAt = &n
			c.state.LastMove.Outcome = "moving"
			c.state.LastMove.StartMode = "broadcast"
			c.started = [2]bool{true, true}
			if e := c.bus.BroadcastStart(); e != nil {
				c.fail(fmt.Errorf("broadcast trigger outcome unknown: %w", e))
				return
			}
		} else {
			c.state.LastMove.StartMode = "individual"
			var first time.Time
			for i, m := range c.state.Plan.Profiles {
				if m.Delta == 0 {
					continue
				}
				n := time.Now()
				if first.IsZero() {
					first = n
					c.state.LastMove.TriggerAt = &n
					c.state.LastMove.Outcome = "moving"
				} else {
					c.state.TriggerSkewMS = float64(n.Sub(first)) / float64(time.Millisecond)
				}
				c.state.LastMove.TriggerSkewMS = c.state.TriggerSkewMS
				c.started[i] = true
				if e := c.bus.Write(byte(i+1), 0x6002, 0x10); e != nil {
					c.fail(fmt.Errorf("trigger outcome unknown: %w", e))
					return
				}
			}
		}
		if c.state.TriggerSkewMS > 50 {
			c.fail(fmt.Errorf("trigger skew exceeded 50 ms"))
			return
		}
		c.state.Phase = "moving"
		if c.state.LastMove.StartMode == "broadcast" {
			c.event("One broadcast start sent (no acknowledgement); checking each drive individually")
		} else {
			c.event(fmt.Sprintf("Paths started; trigger request skew %.1f ms", c.state.TriggerSkewMS))
		}
		return
	}
	if phase == "moving" {
		done := true
		for i, s := range c.state.Drums {
			m := c.state.Plan.Profiles[i]
			if math.Abs(float64(int64(s.ProfilePosition)-int64(s.ActualPosition))) > 1000 {
				c.fail(fmt.Errorf("drum %d tracking error >1000 counts", i+1))
				return
			}
			if !c.started[i] {
				if s.Running || s.ProfilePosition != m.Target {
					c.fail(fmt.Errorf("stationary drum %d moved", i+1))
					return
				}
				continue
			}
			if s.Running || s.Trigger != 0 || !s.CommandComplete {
				done = false
				continue
			}
			if s.ProfilePosition != m.Target || math.Abs(float64(int64(s.ActualPosition)-int64(m.Target))) > 100 {
				c.fail(fmt.Errorf("drum %d stopped short", i+1))
				return
			}
		}
		if done {
			c.state.LastMove.motionEnded(time.Now())
			c.state.Phase = "settling"
			c.settleAt = now.Add(time.Second)
			c.event("Both paths completed; settling before brake engagement")
		}
	}
	if phase == "stopping" || phase == "settling" {
		for _, s := range c.state.Drums {
			if s.Running {
				return
			}
		}
		if t := c.state.LastMove; t != nil && t.EndedAt == nil {
			t.motionEnded(time.Now())
		}
		if c.settleAt.IsZero() {
			c.settleAt = now.Add(time.Second)
		}
		if now.Before(c.settleAt) {
			return
		}
		if c.enabledByUs {
			if e := c.beginDisable(now); e != nil {
				c.state.Error = e.Error()
				c.state.Referenced = false
				c.state.Phase = "fault"
				if t := c.state.LastMove; t != nil {
					t.finish("fault", time.Now())
				}
			}
		} else {
			c.finishRest()
		}
		return
	}
	if phase == "disabling" {
		if c.state.Drums[0].Running {
			c.fail(fmt.Errorf("drum 1 moved during disable sequence"))
			return
		}
		if now.Before(c.disableNotBefore) {
			return
		}
		if !c.state.Drums[0].Enabled {
			forced, e := c.bus.Read(1, 0x000f, 1)
			if e != nil {
				c.fail(e)
				return
			}
			outputs, e := c.bus.Read(1, 0x017b, 1)
			if e != nil {
				c.fail(e)
				return
			}
			if forced[0] == 0 && outputs[0]&1 == 0 {
				c.enabledByUs = false
				c.finishRest()
				return
			}
		}
		if !now.Before(c.disableDeadline) {
			c.state.Error = "drum 1 disable/brake output did not clear within 2 seconds"
			c.state.Referenced = false
			c.state.Phase = "fault"
			c.event(c.state.Error)
			if t := c.state.LastMove; t != nil {
				t.finish("fault", time.Now())
			}
		}
	}
	if phase == "idle" || phase == "unreferenced" {
		for _, s := range c.state.Drums {
			if s.Running || (s.Enabled && s.Trigger > 15) {
				c.fail(fmt.Errorf("unexpected external motion"))
				return
			}
		}
	}
}
func (c *Controller) Run(ctx context.Context) {
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			c.tick(now)
		}
	}
}
func (c *Controller) Close() error {
	_ = c.Stop()
	// Poll before disabling: zeroing forced enable is not a quick-stop command.
	for i := 0; i < 50; i++ {
		time.Sleep(100 * time.Millisecond)
		c.tick(time.Now())
		s := c.Snapshot()
		if s.Phase != "stopping" && s.Phase != "settling" && s.Phase != "disabling" {
			break
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	var e error
	if c.enabledByUs {
		e = fmt.Errorf("could not confirm drum 1 disabled during shutdown")
	}
	return errors.Join(e, c.bus.Close())
}
