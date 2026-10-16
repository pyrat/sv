package dashboard

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// A motion is one ordered transaction. No second axis starts until the first
// reports path completion, standstill and arrival at its target.
type Motion struct {
	Kind       string             `json:"kind"`
	Order      []string           `json:"order"`
	Stage      int                `json:"stage"`
	ActiveAxis string             `json:"activeAxis"`
	Status     string             `json:"status"`
	Targets    map[string]float64 `json:"targets"`
	Speed      float64            `json:"speed"`
	Error      string             `json:"error,omitempty"`
}

func tolerance(a *Axis) float64 {
	return math.Max(.001, math.Max(1/a.config.PulsesPerMetre, 1/a.config.FeedbackCountsPerMetre))
}

func (s *Server) idle() error {
	if s.motion != nil && s.motion.Status == "running" {
		return fmt.Errorf("a movement sequence is active")
	}
	for _, a := range s.axes {
		if a.Available && a.State != "idle" {
			return fmt.Errorf("axis %s is %s", a.ID, a.State)
		}
	}
	return nil
}

func (s *Server) startMotion(kind string, targets map[string]float64, speed float64) error {
	if err := s.idle(); err != nil {
		return err
	}
	if len(targets) == 0 {
		return fmt.Errorf("targets required")
	}
	for id, p := range targets {
		a := s.axes[id]
		if err := ready(a, speed); err != nil {
			return err
		}
		if !finite(p) || p < a.Min || p > a.Max {
			return fmt.Errorf("%s target outside %.3f–%.3f m", id, a.Min, a.Max)
		}
	}
	order := []string{"x", "y"}
	if kind == "home" {
		order = []string{"y", "x"}
	}
	m := &Motion{Kind: kind, Order: []string{}, Status: "running", Targets: map[string]float64{}, Speed: speed}
	for _, id := range order {
		if p, ok := targets[id]; ok {
			m.Order = append(m.Order, id)
			m.Targets[id] = p
			s.axes[id].Target = p
			s.axes[id].State = "queued"
		}
	}
	s.motion = m
	s.event(fmt.Sprintf("%s sequence accepted · %v", kind, m.Order))
	return s.advanceMotion(time.Now())
}

func (s *Server) advanceMotion(now time.Time) error {
	m := s.motion
	if m == nil || m.Status != "running" {
		return nil
	}
	if m.ActiveAxis != "" {
		if s.axes[m.ActiveAxis].State != "idle" {
			return nil
		}
		m.Stage++
		m.ActiveAxis = ""
	}
	for m.Stage < len(m.Order) {
		a := s.axes[m.Order[m.Stage]]
		if a.State != "queued" || !a.Available || now.Sub(a.Updated) > time.Second {
			return s.failMotion(a, fmt.Errorf("queued axis %s is no longer ready", a.ID))
		}
		p := m.Targets[a.ID]
		if math.Abs(a.Position-p) < .5/a.config.PulsesPerMetre {
			a.State = "idle"
			a.Target = a.Position
			m.Stage++
			s.event(a.ID + " already at target")
			continue
		}
		if err := a.motor.move(p, m.Speed); err != nil {
			return s.failMotion(a, err)
		}
		a.State = "moving"
		a.Target = p
		a.deadline = now.Add(time.Duration((math.Abs(p-a.Position)/m.Speed + 15) * float64(time.Second)))
		m.ActiveAxis = a.ID
		s.event(fmt.Sprintf("%s · moving %s (%d/%d)", m.Kind, a.ID, m.Stage+1, len(m.Order)))
		return nil
	}
	m.Status = "completed"
	s.event(m.Kind + " sequence complete")
	return nil
}

func (s *Server) failMotion(a *Axis, err error) error {
	a.State = "fault"
	a.Error = err.Error()
	if s.motion != nil && s.motion.Status == "running" {
		s.motion.Error = err.Error()
	}
	s.event(a.ID + ": " + err.Error())
	return errors.Join(err, s.stopAll())
}
