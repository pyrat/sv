package dashboard

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	s, err := New(true, DemoConfig())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	return s
}

func TestMoveAndSignedStep(t *testing.T) {
	s := testServer(t)
	if _, err := s.execute("/api/move", command{Targets: map[string]float64{"x": 2.5, "y": 4.9}, Speed: .1}); err != nil {
		t.Fatal(err)
	}
	s.poll(time.Now().Add(2 * time.Second))
	if s.axes["y"].Position != 4.8 || s.motion.ActiveAxis != "y" {
		t.Fatal("Y must start only after X completes")
	}
	s.poll(time.Now().Add(4 * time.Second))
	if s.axes["x"].Position != 2.5 || s.axes["y"].Position != 4.9 || s.axes["x"].State != "idle" {
		t.Fatal("did not reach targets")
	}
	if _, err := s.execute("/api/step", command{Axis: "x", Delta: -.01, Speed: .1}); err != nil {
		t.Fatal(err)
	}
	s.poll(time.Now().Add(5 * time.Second))
	if math.Abs(s.axes["x"].Position-2.49) > 1e-6 {
		t.Fatal(s.axes["x"].Position)
	}
}

func TestValidateEntireMoveBeforeStarting(t *testing.T) {
	s := testServer(t)
	for _, c := range []command{{Targets: map[string]float64{"x": 3, "y": s.axes["y"].Max + 1}, Speed: .1}, {Targets: map[string]float64{"z": 1}, Speed: .1}, {Targets: map[string]float64{"x": 3}, Speed: 0}, {Targets: map[string]float64{"x": math.NaN()}, Speed: .1}, {Targets: map[string]float64{"x": 3}, Speed: .3}} {
		if _, err := s.execute("/api/move", c); err == nil {
			t.Fatal("accepted invalid command")
		}
		if s.axes["x"].State != "idle" {
			t.Fatal("partial move started")
		}
	}
}

func TestJogExpiresAndCannotBeResurrected(t *testing.T) {
	s := testServer(t)
	v, err := s.execute("/api/jog", command{Axis: "x", Direction: 1, Speed: .1})
	if err != nil {
		t.Fatal(err)
	}
	lease := v.(map[string]string)["lease"]
	s.tick(time.Now().Add(time.Second))
	if s.axes["x"].lease != "" || s.axes["x"].direction != 0 {
		t.Fatal("jog continued after expiry")
	}
	if _, err = s.execute("/api/jog/renew", command{Lease: lease}); err == nil {
		t.Fatal("revived expired jog")
	}
}

func TestStopRevokesLeaseAndNewJogHasNewToken(t *testing.T) {
	s := testServer(t)
	v, _ := s.execute("/api/jog", command{Axis: "y", Direction: -1, Speed: .1})
	old := v.(map[string]string)["lease"]
	if _, err := s.execute("/api/stop", command{}); err != nil {
		t.Fatal(err)
	}
	s.poll(time.Now())
	v, err := s.execute("/api/jog", command{Axis: "y", Direction: 1, Speed: .1})
	if err != nil {
		t.Fatal(err)
	}
	if v.(map[string]string)["lease"] == old {
		t.Fatal("reused token")
	}
	if _, err = s.execute("/api/jog/renew", command{Lease: old}); err == nil {
		t.Fatal("old token accepted")
	}
}

func TestJogStopsBeforeBoundary(t *testing.T) {
	s := testServer(t)
	a := s.axes["x"]
	sim := a.motor.(*simulator)
	sim.position = s.axes["x"].Max - .01
	sim.target = sim.position
	s.poll(time.Now())
	if _, err := s.execute("/api/jog", command{Axis: "x", Direction: 1, Speed: .1}); err != nil {
		t.Fatal(err)
	}
	s.poll(time.Now())
	if a.lease != "" {
		t.Fatal("jog not stopped at boundary")
	}
}

func TestUnconfiguredHardwareCannotMove(t *testing.T) {
	s, err := New(false, map[string]AxisConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.execute("/api/move", command{Targets: map[string]float64{"x": 1}, Speed: .1}); err == nil {
		t.Fatal("unconfigured axis accepted move")
	}
}

func TestHTTPContracts(t *testing.T) {
	s := testServer(t)
	h := s.Handler()
	for _, tc := range []struct {
		body, origin, content string
		want                  int
	}{{`{"targets":{"x":2.6},"speed":0.1}`, "http://evil.example", "application/json", 403}, {`{"unknown":1}`, "", "application/json", 400}, {`{} {}`, "", "application/json", 400}, {`{}`, "", "text/plain", 415}} {
		r := httptest.NewRequest("POST", "http://localhost/api/move", strings.NewReader(tc.body))
		r.Header.Set("Origin", tc.origin)
		r.Header.Set("Content-Type", tc.content)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Fatalf("got %d: %s", w.Code, w.Body.String())
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/state", nil))
	var state struct {
		Simulation bool   `json:"simulation"`
		Axes       []Axis `json:"axes"`
	}
	if err := json.NewDecoder(bytes.NewReader(w.Body.Bytes())).Decode(&state); err != nil {
		t.Fatal(err)
	}
	if !state.Simulation || len(state.Axes) != 2 {
		t.Fatal(state)
	}
}

type failingMotor struct {
	motor
	failMove bool
	failRead bool
	failStop bool
	stopped  bool
}

func (f *failingMotor) move(p, v float64) error {
	if f.failMove {
		return errors.New("trigger outcome unknown")
	}
	return f.motor.move(p, v)
}
func (f *failingMotor) read(now time.Time) (reading, error) {
	if f.failRead {
		return reading{}, errors.New("serial link lost")
	}
	return f.motor.read(now)
}
func (f *failingMotor) stop() error {
	f.stopped = true
	if f.failStop {
		return errors.New("stop transmission failed")
	}
	return f.motor.stop()
}

func TestPartialMoveFailureStopsBothAxes(t *testing.T) {
	s := testServer(t)
	x := &failingMotor{motor: s.axes["x"].motor}
	y := &failingMotor{motor: s.axes["y"].motor, failMove: true}
	s.axes["x"].motor = x
	s.axes["y"].motor = y
	if _, err := s.execute("/api/move", command{Targets: map[string]float64{"x": 3, "y": 5}, Speed: .1}); err != nil {
		t.Fatal(err)
	}
	s.poll(time.Now().Add(7 * time.Second))
	if s.motion.Status != "cancelled" || s.motion.Error == "" {
		t.Fatal("missing second-stage error")
	}
	if !x.stopped || !y.stopped {
		t.Fatal("did not stop both motors")
	}
}
func TestTelemetryFaultRevokesJogAndReportsFailedStop(t *testing.T) {
	s := testServer(t)
	if _, err := s.execute("/api/jog", command{Axis: "x", Direction: 1, Speed: .1}); err != nil {
		t.Fatal(err)
	}
	f := &failingMotor{motor: s.axes["x"].motor, failRead: true, failStop: true}
	s.axes["x"].motor = f
	s.poll(time.Now())
	if !f.stopped || s.axes["x"].lease != "" || s.axes["x"].State != "fault" || !strings.Contains(s.axes["x"].Error, "stop transmission failed") {
		t.Fatal(s.axes["x"])
	}
}
