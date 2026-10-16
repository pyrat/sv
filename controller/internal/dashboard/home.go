package dashboard

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"
)

type HomeAxis struct {
	Position        float64 `json:"position"`
	CommandCounter  int32   `json:"commandCounter"`
	ProfileCounter  *int32  `json:"profileCounter,omitempty"`
	FeedbackCounter *int32  `json:"feedbackCounter,omitempty"`
}
type Home struct {
	Version     int                 `json:"version"`
	Simulation  bool                `json:"simulation"`
	Calibration string              `json:"calibration"`
	SavedAt     time.Time           `json:"savedAt"`
	Source      string              `json:"source"`
	Axes        map[string]HomeAxis `json:"axes"`
}
type HomeStatus struct {
	Saved  *Home  `json:"saved"`
	Usable bool   `json:"usable"`
	AtHome bool   `json:"atHome"`
	Reason string `json:"reason"`
}

func (s *Server) calibrationID() string {
	configs := map[string]AxisConfig{}
	for id, a := range s.axes {
		if a.config.PulsesPerMetre > 0 {
			configs[id] = a.config
		}
	}
	b, _ := json.Marshal(configs)
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])
}

// LoadHome configures the state file; "-" explicitly selects ephemeral state.
// Counter snapshots are never written back to the drive on startup.
func (s *Server) LoadHome(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if path == "-" {
		s.homePath = ""
		return nil
	}
	if path == "" {
		return fmt.Errorf("home state path required")
	}
	s.homePath = path
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var h *Home
	if err = json.Unmarshal(b, &h); err != nil {
		return fmt.Errorf("read home state: %w", err)
	}
	if h != nil && (h.Version != 1 || len(h.Axes) == 0) {
		return fmt.Errorf("invalid home state version or axes")
	}
	s.home = h
	return nil
}

func (s *Server) persistHome(h *Home) error {
	if s.homePath == "" {
		return nil
	}
	dir := filepath.Dir(s.homePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".depthpilot-home-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = json.NewEncoder(f).Encode(h); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), s.homePath)
}

func (s *Server) homeStatus() HomeStatus {
	h := HomeStatus{Saved: s.home, Reason: "No home position saved"}
	if s.home == nil {
		return h
	}
	if s.home.Simulation != s.simulation || s.home.Calibration != s.calibrationID() {
		h.Reason = "Axis calibration or reference changed; set home again"
		return h
	}
	for id, a := range s.axes {
		p, ok := s.home.Axes[id]
		if !ok {
			if a.Available {
				h.Reason = "Home does not include axis " + id
				return h
			}
			continue
		}
		if !a.Available || !a.config.ReferenceValid || (!s.simulation && a.config.ReferenceID == "") {
			h.Reason = "Axis " + id + " reference is not established"
			return h
		}
		if !finite(p.Position) || p.Position < a.Min || p.Position > a.Max {
			h.Reason = "Saved home is outside travel limits"
			return h
		}
		counter := math.Round(p.Position * a.config.PulsesPerMetre * float64(a.config.Direction))
		if counter != float64(p.CommandCounter) {
			h.Reason = "Saved home counter does not match its position"
			return h
		}
		if a.State == "fault" || time.Since(a.Updated) > time.Second {
			h.Reason = "Axis " + id + " feedback is unavailable"
			return h
		}
	}
	for id := range s.home.Axes {
		if s.axes[id] == nil {
			h.Reason = "Unknown axis in saved home"
			return h
		}
	}
	h.Usable = true
	h.Reason = ""
	h.AtHome = true
	for id, p := range s.home.Axes {
		a := s.axes[id]
		if a.State != "idle" || math.Abs(a.Position-p.Position) > tolerance(a) {
			h.AtHome = false
		}
	}
	return h
}

func (s *Server) setHome(c command) error {
	if err := s.idle(); err != nil {
		return err
	}
	// Read fresh counters while stationary rather than saving browser readouts.
	s.poll(time.Now())
	if err := s.idle(); err != nil {
		return err
	}
	source := c.Source
	if source == "" {
		source = "current"
	}
	if source != "current" && source != "positions" && source != "counters" {
		return fmt.Errorf("home source must be current, positions or counters")
	}
	if source == "current" && (len(c.Targets) > 0 || len(c.Counters) > 0) || source == "positions" && len(c.Counters) > 0 || source == "counters" && len(c.Targets) > 0 {
		return fmt.Errorf("home sources cannot be mixed")
	}
	h := &Home{Version: 1, Simulation: s.simulation, Calibration: s.calibrationID(), SavedAt: time.Now().UTC(), Source: source, Axes: map[string]HomeAxis{}}
	for _, id := range []string{"x", "y"} {
		a := s.axes[id]
		if !a.Available {
			continue
		}
		if !s.simulation && a.config.ReferenceID == "" {
			return fmt.Errorf("axis %s requires a reference_id before saving home", id)
		}
		if err := ready(a, a.MaxSpeed); err != nil {
			return err
		}
		p := a.Position
		if source == "positions" {
			var ok bool
			p, ok = c.Targets[id]
			if !ok {
				return fmt.Errorf("home position required for %s", id)
			}
		}
		if source == "counters" {
			count, ok := c.Counters[id]
			if !ok {
				return fmt.Errorf("home command counter required for %s", id)
			}
			p = float64(count) / a.config.PulsesPerMetre * float64(a.config.Direction)
		}
		if !finite(p) || p < a.Min || p > a.Max {
			return fmt.Errorf("%s home is outside travel limits", id)
		}
		entry := HomeAxis{Position: p, CommandCounter: int32(math.Round(p * a.config.PulsesPerMetre * float64(a.config.Direction)))}
		if source == "current" {
			profile, feedback := a.ProfileCounter, a.FeedbackCounter
			entry.ProfileCounter = &profile
			entry.FeedbackCounter = &feedback
		}
		h.Axes[id] = entry
	}
	if len(h.Axes) == 0 {
		return fmt.Errorf("no configured axes")
	}
	if source == "positions" && len(c.Targets) != len(h.Axes) || source == "counters" && len(c.Counters) != len(h.Axes) {
		return fmt.Errorf("home contains unavailable axes")
	}
	if err := s.persistHome(h); err != nil {
		return fmt.Errorf("save home: %w", err)
	}
	s.home = h
	s.event("Home position saved · counters unchanged")
	return nil
}
