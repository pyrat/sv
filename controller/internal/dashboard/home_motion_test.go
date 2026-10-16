package dashboard

import (
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func do(t *testing.T, s *Server, path string, c command) {
	t.Helper()
	if _, err := s.execute(path, c); err != nil {
		t.Fatal(err)
	}
}

func TestHomeCapturePersistsCountersWithoutZeroing(t *testing.T) {
	s := testServer(t)
	path := filepath.Join(t.TempDir(), "home.json")
	if err := s.LoadHome(path); err != nil {
		t.Fatal(err)
	}
	before := s.axes["x"].Position
	do(t, s, "/api/home/set", command{})
	h := s.home.Axes["x"]
	if s.axes["x"].Position != before || h.Position != before || h.CommandCounter != 129850 || h.FeedbackCounter == nil || *h.FeedbackCounter != 324625 {
		t.Fatalf("wrong home/counters: %+v", h)
	}
	if !s.homeStatus().AtHome {
		t.Fatal("current position should be home")
	}
	restarted := testServer(t)
	if err := restarted.LoadHome(path); err != nil {
		t.Fatal(err)
	}
	if !restarted.homeStatus().Usable || restarted.home.Axes["x"].CommandCounter != 129850 {
		t.Fatal("home did not survive restart")
	}
	// Loading a saved home never overwrites live counters or current position.
	if restarted.axes["x"].Position != before {
		t.Fatal("load changed live position")
	}
}

func TestHomeReturnWaitsForYThenStartsX(t *testing.T) {
	s := testServer(t)
	do(t, s, "/api/home/set", command{Source: "positions", Targets: map[string]float64{"x": 2, "y": 4}})
	do(t, s, "/api/home/return", command{Speed: .1})
	if s.motion.ActiveAxis != "y" || s.axes["x"].State != "queued" || s.axes["x"].motor.(*simulator).target != 2.45 {
		t.Fatal("X started before Y")
	}
	s.poll(time.Now().Add(4 * time.Second))
	if s.motion.ActiveAxis != "y" || s.axes["x"].Position != 2.45 {
		t.Fatal("X started while Y still moving")
	}
	s.poll(time.Now().Add(9 * time.Second))
	if s.motion.ActiveAxis != "x" || s.axes["y"].Position != 4 {
		t.Fatal("did not advance after Y arrived")
	}
	s.poll(time.Now().Add(15 * time.Second))
	if s.motion.Status != "completed" || !s.homeStatus().AtHome {
		t.Fatalf("did not arrive home: %+v", s.homeStatus())
	}
}

func TestStopCancelsRemainingStageAndRejectsOverlappingCommands(t *testing.T) {
	s := testServer(t)
	do(t, s, "/api/move", command{Targets: map[string]float64{"x": 3, "y": 5}, Speed: .1})
	for _, item := range []struct {
		path string
		c    command
	}{{"/api/home/set", command{}}, {"/api/jog", command{Axis: "y", Direction: 1, Speed: .1}}, {"/api/move", command{Targets: map[string]float64{"y": 6}, Speed: .1}}} {
		if _, err := s.execute(item.path, item.c); err == nil {
			t.Fatalf("accepted overlapping %s", item.path)
		}
	}
	do(t, s, "/api/stop", command{})
	s.poll(time.Now().Add(20 * time.Second))
	if s.axes["y"].Position != 4.8 || s.motion.Status != "cancelled" {
		t.Fatal("queued Y moved after stop")
	}
}

func TestFaultAndDeadlineCancelQueuedAxis(t *testing.T) {
	for _, fault := range []bool{true, false} {
		t.Run(map[bool]string{true: "fault", false: "deadline"}[fault], func(t *testing.T) {
			s := testServer(t)
			do(t, s, "/api/move", command{Targets: map[string]float64{"x": 3, "y": 5}, Speed: .1})
			if fault {
				s.axes["x"].motor = &failingMotor{motor: s.axes["x"].motor, failRead: true}
				s.poll(time.Now())
			} else {
				s.axes["x"].deadline = time.Now().Add(-time.Second)
				s.tick(time.Now())
			}
			if s.motion.Status != "cancelled" || s.axes["y"].motor.(*simulator).target != 4.8 {
				t.Fatal("remaining stage was not cancelled")
			}
		})
	}
}

type completionMotor struct {
	motor
	position  float64
	completed bool
}

func (m *completionMotor) read(_ time.Time) (reading, error) {
	return reading{Position: m.position, Completed: m.completed}, nil
}

func TestStandstillAloneDoesNotReleaseNextAxis(t *testing.T) {
	s := testServer(t)
	m := &completionMotor{motor: s.axes["x"].motor, position: 3}
	s.axes["x"].motor = m
	do(t, s, "/api/move", command{Targets: map[string]float64{"x": 3, "y": 5}, Speed: .1})
	s.poll(time.Now())
	if s.motion.ActiveAxis != "x" || s.axes["y"].State != "queued" {
		t.Fatal("advanced before path completion")
	}
	m.completed = true
	m.position = 2.7
	s.poll(time.Now())
	if s.motion.Status != "cancelled" || s.axes["x"].State != "fault" {
		t.Fatal("short arrival did not fault")
	}
}

func TestOneMillimetreStepIsNotSkipped(t *testing.T) {
	s := testServer(t)
	do(t, s, "/api/step", command{Axis: "x", Delta: .001, Speed: .1})
	if s.motion.ActiveAxis != "x" {
		t.Fatal("1mm move skipped")
	}
	s.poll(time.Now().Add(time.Second))
	if math.Abs(s.axes["x"].Position-2.451) > 1e-8 {
		t.Fatal(s.axes["x"].Position)
	}
}

func TestHomeCounterEditingAndCalibrationInvalidation(t *testing.T) {
	s := testServer(t)
	do(t, s, "/api/home/set", command{Source: "counters", Counters: map[string]int32{"x": 106000, "y": 212000}})
	if s.home.Axes["x"].Position != 2 || s.home.Axes["y"].Position != 4 || s.axes["x"].Position != 2.45 {
		t.Fatal("editing home changed counters or position")
	}
	s.axes["x"].config.ReferenceID = "new-zero"
	if s.homeStatus().Usable {
		t.Fatal("old home reused after reference change")
	}
	if _, err := s.execute("/api/home/return", command{Speed: .1}); err == nil {
		t.Fatal("invalid home accepted")
	}
}

func TestInvalidHomeDoesNotOverwriteSavedHome(t *testing.T) {
	s := testServer(t)
	do(t, s, "/api/home/set", command{})
	original := s.home
	for _, c := range []command{{Source: "counters", Counters: map[string]int32{"x": -1, "y": 1}}, {Source: "positions", Targets: map[string]float64{"x": 2}}, {Source: "positions", Targets: map[string]float64{"x": 2, "y": s.axes["y"].Max + 1}}, {Source: "current", Counters: map[string]int32{"x": 1}}, {Source: "invalid"}} {
		if _, err := s.execute("/api/home/set", c); err == nil {
			t.Fatal("invalid home accepted")
		}
		if s.home != original {
			t.Fatal("invalid command replaced home")
		}
	}
	s.homePath = t.TempDir() // renaming a file onto a directory must fail
	if _, err := s.execute("/api/home/set", command{}); err == nil {
		t.Fatal("missing persistence error")
	}
	if s.home != original {
		t.Fatal("published unsaved home")
	}
}

func TestClearPersistsAndCorruptFileIsRejected(t *testing.T) {
	s := testServer(t)
	path := filepath.Join(t.TempDir(), "home.json")
	if err := s.LoadHome(path); err != nil {
		t.Fatal(err)
	}
	do(t, s, "/api/home/set", command{})
	do(t, s, "/api/home/clear", command{})
	restarted := testServer(t)
	if err := restarted.LoadHome(path); err != nil {
		t.Fatal(err)
	}
	if restarted.home != nil {
		t.Fatal("cleared home restored")
	}
	if err := os.WriteFile(path, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := restarted.LoadHome(path); err == nil {
		t.Fatal("corrupt state silently accepted")
	}
}

func TestSingleAxisHome(t *testing.T) {
	configs := DemoConfig()
	delete(configs, "y")
	s, err := New(true, configs)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	do(t, s, "/api/home/set", command{Source: "positions", Targets: map[string]float64{"x": 2}})
	do(t, s, "/api/home/return", command{Speed: .1})
	if len(s.motion.Order) != 1 || s.motion.Order[0] != "x" {
		t.Fatal(s.motion)
	}
}
