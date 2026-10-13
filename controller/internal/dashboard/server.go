package dashboard

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const leaseDuration = 400 * time.Millisecond

type Axis struct {
	ID              string    `json:"id"`
	Position        float64   `json:"position"`
	Target          float64   `json:"target"`
	Min             float64   `json:"min"`
	Max             float64   `json:"max"`
	MaxSpeed        float64   `json:"maxSpeed"`
	State           string    `json:"state"`
	Error           string    `json:"error"`
	Available       bool      `json:"available"`
	Updated         time.Time `json:"updated"`
	ProfileCounter  int32     `json:"profileCounter"`
	FeedbackCounter int32     `json:"feedbackCounter"`
	config          AxisConfig
	motor           motor
	lease           string
	expires         time.Time
	direction       int
	speed           float64
	deadline        time.Time
}
type Event struct {
	ID      int64     `json:"id"`
	Time    time.Time `json:"time"`
	Text    string    `json:"text"`
	Session string    `json:"session"`
}
type Server struct {
	mu          sync.Mutex
	axes        map[string]*Axis
	simulation  bool
	events      []Event
	lastPoll    time.Time
	closed      bool
	motion      *Motion
	home        *Home
	homePath    string
	activity    *activityStore
	nextEventID int64
	session     string
}

func New(simulation bool, configs map[string]AxisConfig) (*Server, error) {
	s := &Server{axes: map[string]*Axis{}, simulation: simulation, events: []Event{}}
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	s.session = hex.EncodeToString(b)
	ports := map[string]bool{}
	for _, id := range []string{"x", "y"} {
		a := &Axis{ID: id, State: "unconfigured", Error: "Axis calibration required"}
		s.axes[id] = a
		c, ok := configs[id]
		if !ok {
			continue
		}
		if err := c.validate(); err != nil {
			s.Close()
			return nil, fmt.Errorf("axis %s: %w", id, err)
		}
		a.config = c
		a.Min = c.Min
		a.Max = c.Max
		a.MaxSpeed = c.MaxSpeed
		if !simulation && !c.ReferenceValid {
			a.Error = "Reference must be established before enabling movement"
			continue
		}
		var err error
		if simulation {
			p := 2.45
			if id == "y" {
				p = 4.8
			}
			p = math.Max(c.Min, math.Min(c.Max, p))
			a.motor = &simulator{position: p, target: p, last: time.Now(), config: c}
		} else {
			if ports[c.Port] {
				s.Close()
				return nil, fmt.Errorf("each axis needs its own serial channel")
			}
			ports[c.Port] = true
			a.motor, err = newHardware(c)
		}
		if err != nil {
			s.Close()
			return nil, fmt.Errorf("axis %s: %w", id, err)
		}
		a.Available = true
		a.State = "idle"
		a.Error = ""
	}
	// Each hardware axis has an independent serial channel. Polling an idle
	// axis must still service the other axis's <50ms jog refresh stream.
	for _, a := range s.axes {
		if h, ok := a.motor.(*hardware); ok {
			h.heartbeat = func() error {
				for _, other := range s.axes {
					if other.motor != h && other.lease != "" {
						if !time.Now().Before(other.expires) {
							return fmt.Errorf("jog lease expired during telemetry")
						}
						if err := other.motor.refresh(); err != nil {
							return err
						}
					}
				}
				return nil
			}
		}
	}
	s.poll(time.Now())
	s.event("Control station online")
	return s, nil
}
func (s *Server) event(text string) {
	s.nextEventID++
	s.appendEvent(Event{ID: s.nextEventID, Time: time.Now().UTC(), Text: text, Session: s.session})
}
func (s *Server) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	for _, a := range s.axes {
		if a.motor != nil {
			a.motor.close()
		}
	}
	if s.activity != nil {
		if err := s.activity.close(); err != nil {
			log.Printf("flush activity on shutdown: %v", err)
		}
	}
}
func (s *Server) stopAll() error {
	if s.motion != nil && s.motion.Status == "running" {
		s.motion.Status = "cancelled"
	}
	var errs []error
	for _, id := range []string{"x", "y"} {
		a := s.axes[id]
		a.lease = ""
		a.deadline = time.Time{}
		a.direction = 0
		if a.motor != nil {
			if err := a.motor.stop(); err != nil {
				a.State = "fault"
				a.Error = err.Error()
				errs = append(errs, fmt.Errorf("%s stop: %w", id, err))
			} else {
				if a.State != "fault" {
					a.State = "stopping"
				}
				a.Target = a.Position
			}
		}
	}
	return errors.Join(errs...)
}
func (s *Server) poll(now time.Time) {
	for _, id := range []string{"x", "y"} {
		a := s.axes[id]
		if a.motor == nil {
			continue
		}
		r, err := a.motor.read(now)
		if err != nil || r.Fault != "" {
			message := r.Fault
			if err != nil {
				message = err.Error()
			}
			if a.State != "fault" {
				s.failMotion(a, fmt.Errorf("%s", message))
			}
			continue
		}
		a.Position = r.Position
		a.ProfileCounter, a.FeedbackCounter = r.ProfileCounter, r.FeedbackCounter
		a.Updated = now
		if a.State == "fault" {
			continue
		} // faults latch until service restart / diagnosis
		if a.State == "queued" && r.Running {
			s.failMotion(a, fmt.Errorf("queued axis %s moved before its turn", id))
			continue
		}
		if a.Position < a.Min-0.001 || a.Position > a.Max+0.001 {
			s.failMotion(a, fmt.Errorf("position outside configured travel"))
			continue
		}
		if a.State == "idle" && r.Running {
			a.State = "moving"
			a.deadline = now.Add(30 * time.Second)
			s.event(id + ": motion detected on drive")
		}
		if a.direction != 0 {
			// Include a conservative host polling margin plus nominal drive ramp.
			rpm := a.speed * a.config.PulsesPerMetre * 60 / a.config.PulsesPerRevolution
			margin := a.speed * (0.5 + float64(a.config.Ramp)*rpm/1e6)
			if (a.direction > 0 && a.Position >= a.Max-margin) || (a.direction < 0 && a.Position <= a.Min+margin) {
				s.stopAll()
				s.event("Jog stopped at travel boundary")
			}
		} else if !r.Running && (a.State == "moving" || a.State == "stopping") {
			if a.State == "moving" && s.motion != nil && s.motion.Status == "running" && s.motion.ActiveAxis == id {
				if !r.Completed {
					continue
				}
				if math.Abs(a.Position-a.Target) > tolerance(a) {
					s.failMotion(a, fmt.Errorf("axis %s stopped short of its target", id))
					continue
				}
			}
			a.State = "idle"
			a.deadline = time.Time{}
			s.event(id + " axis at rest")
		}
		if a.State == "idle" {
			a.Target = a.Position
		}
	}
	s.advanceMotion(now)
}
func (s *Server) tick(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range []string{"x", "y"} {
		a := s.axes[id]
		if a.lease != "" && !now.Before(a.expires) {
			s.stopAll()
			s.event("Jog released · command lease expired")
			break
		}
		if !a.deadline.IsZero() && now.After(a.deadline) {
			if s.motion != nil && s.motion.Status == "running" {
				s.motion.Error = "Motion deadline reached"
			}
			s.stopAll()
			s.event("Motion deadline reached")
			break
		}
	}
	for _, id := range []string{"x", "y"} {
		a := s.axes[id]
		if a.lease != "" {
			if err := a.motor.refresh(); err != nil {
				a.State = "fault"
				a.Error = err.Error()
				s.stopAll()
				s.event("Jog communication fault")
				break
			}
		}
	}
	if now.Sub(s.lastPoll) >= 100*time.Millisecond {
		s.poll(now)
		s.lastPoll = now
	}
}
func (s *Server) Run(ctx context.Context) {
	t := time.NewTicker(20 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			s.mu.Lock()
			s.closed = true
			s.event("Control station stopping")
			s.stopAll()
			s.mu.Unlock()
			return
		case now := <-t.C:
			s.tick(now)
		}
	}
}

type command struct {
	Targets   map[string]float64 `json:"targets"`
	Speed     float64            `json:"speed"`
	Axis      string             `json:"axis"`
	Direction int                `json:"direction"`
	Lease     string             `json:"lease"`
	Delta     float64            `json:"delta"`
	Source    string             `json:"source"`
	Counters  map[string]int32   `json:"counters"`
}

func ready(a *Axis, v float64) error {
	if a == nil || !a.Available {
		return fmt.Errorf("axis not configured")
	}
	if a.State != "idle" || time.Since(a.Updated) > time.Second {
		return fmt.Errorf("axis unavailable or busy")
	}
	if !finite(v) || v <= 0 || v > a.MaxSpeed {
		return fmt.Errorf("speed must be > 0 and <= %.3f m/s", a.MaxSpeed)
	}
	return nil
}
func (s *Server) execute(path string, c command) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, fmt.Errorf("controller is shutting down")
	}
	switch path {
	case "/api/home/set":
		if err := s.setHome(c); err != nil {
			return nil, err
		}
		return s.homeStatus(), nil
	case "/api/home/clear":
		if err := s.idle(); err != nil {
			return nil, err
		}
		if err := s.persistHome(nil); err != nil {
			return nil, err
		}
		s.home = nil
		s.event("Home position cleared")
		return s.homeStatus(), nil
	case "/api/home/return":
		h := s.homeStatus()
		if !h.Usable {
			return nil, fmt.Errorf("%s", h.Reason)
		}
		targets := map[string]float64{}
		for id, p := range s.home.Axes {
			targets[id] = p.Position
		}
		if err := s.startMotion("home", targets, c.Speed); err != nil {
			return nil, err
		}
		return map[string]bool{"ok": true}, nil
	case "/api/stop":
		err := s.stopAll()
		s.event("Stop requested · all axes")
		return map[string]bool{"ok": err == nil}, err
	case "/api/jog/renew":
		for _, a := range s.axes {
			if c.Lease != "" && a.lease == c.Lease && time.Now().Before(a.expires) {
				a.expires = time.Now().Add(leaseDuration)
				return map[string]bool{"ok": true}, nil
			}
		}
		return nil, fmt.Errorf("jog lease expired or revoked")
	case "/api/jog":
		if err := s.idle(); err != nil {
			return nil, err
		}
		a := s.axes[c.Axis]
		if err := ready(a, c.Speed); err != nil {
			return nil, err
		}
		if c.Direction != 1 && c.Direction != -1 {
			return nil, fmt.Errorf("direction must be -1 or 1")
		}
		for _, other := range s.axes {
			if other.State == "moving" || other.State == "jogging" || other.State == "stopping" {
				return nil, fmt.Errorf("another motion is active")
			}
		}
		if (c.Direction < 0 && a.Position <= a.Min) || (c.Direction > 0 && a.Position >= a.Max) {
			return nil, fmt.Errorf("travel boundary reached")
		}
		if err := a.motor.jog(c.Direction, c.Speed); err != nil {
			return nil, errors.Join(err, s.stopAll())
		}
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		a.lease = hex.EncodeToString(b)
		a.expires = time.Now().Add(leaseDuration)
		a.direction = c.Direction
		a.speed = c.Speed
		a.State = "jogging"
		a.Target = a.Position
		s.event(fmt.Sprintf("%s jog · %+.0f", c.Axis, float64(c.Direction)))
		return map[string]string{"lease": a.lease}, nil
	case "/api/move", "/api/step":
		if path == "/api/step" {
			a := s.axes[c.Axis]
			if a == nil || !finite(c.Delta) || c.Delta == 0 {
				return nil, fmt.Errorf("invalid axis or step")
			}
			c.Targets = map[string]float64{c.Axis: a.Position + c.Delta}
		}
		if err := s.startMotion("outbound", c.Targets, c.Speed); err != nil {
			return nil, err
		}
		return map[string]bool{"ok": true}, nil
	}
	return nil, fmt.Errorf("unknown command")
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func (s *Server) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /api/state", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		writeJSON(w, 200, map[string]any{"simulation": s.simulation, "axes": []*Axis{s.axes["x"], s.axes["y"]}, "events": s.events, "time": time.Now(), "motion": s.motion, "home": s.homeStatus(), "activity": s.activityStatus()})
	})
	m.HandleFunc("GET /api/activity", s.activityHandler)
	for _, path := range []string{"/api/move", "/api/step", "/api/jog", "/api/jog/renew", "/api/stop", "/api/home/set", "/api/home/clear", "/api/home/return"} {
		path := path
		m.HandleFunc("POST "+path, func(w http.ResponseWriter, r *http.Request) {
			if origin := r.Header.Get("Origin"); origin != "" {
				u, err := url.Parse(origin)
				if err != nil || u.Host != r.Host {
					writeJSON(w, 403, map[string]string{"error": "cross-origin control is not allowed"})
					return
				}
			}
			if r.Header.Get("Content-Type") != "application/json" {
				writeJSON(w, 415, map[string]string{"error": "application/json required"})
				return
			}
			var c command
			d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
			d.DisallowUnknownFields()
			if err := d.Decode(&c); err != nil {
				writeJSON(w, 400, map[string]string{"error": "invalid command JSON"})
				return
			}
			if err := d.Decode(&struct{}{}); err != io.EOF {
				writeJSON(w, 400, map[string]string{"error": "one JSON object required"})
				return
			}
			result, err := s.execute(path, c)
			if err != nil {
				writeJSON(w, 409, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, 200, result)
		})
	}
	return m
}
