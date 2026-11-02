package rope

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type recordingBus struct {
	*SimBus
	writes          []string
	failProgram     byte
	failTrigger     byte
	failRead        bool
	pendingDisabled uint16
	disableLagReads int
	holdBrakeOutput bool
	enableShift     int32
	failBroadcast   bool
	ignoreBroadcast bool
}

func (b *recordingBus) BroadcastStart() error {
	b.writes = append(b.writes, "broadcast:start")
	if b.ignoreBroadcast {
		return nil
	}
	e := b.SimBus.BroadcastStart()
	if b.failBroadcast {
		return fmt.Errorf("uncertain broadcast write")
	}
	return e
}

func TestBroadcastPreparedOnceAndSingleDrumFallback(t *testing.T) {
	for _, target := range []Point{{.1, 0}, {0, .1}, {.1, -.1}} {
		c, b := setup(t)
		c.state.StartMode = "broadcast"
		c.Reference()
		if e := c.Move(target, .02); e != nil {
			t.Fatal(e)
		}
		c.readyAt = time.Now().Add(-time.Second)
		c.tick(time.Now())
		if c.state.Phase != "moving" {
			t.Fatal(c.state.Error)
		}
		both := target == (Point{.1, 0})
		if contains(b.writes, "broadcast:start") != both {
			t.Fatalf("wrong trigger route: %v", b.writes)
		}
		if both {
			want := "1:000f=0001,program:1,program:2,broadcast:start"
			if strings.Join(b.writes, ",") != want {
				t.Fatal(b.writes)
			}
		}
	}
}

func TestBroadcastFailureStopsBothWithoutRetry(t *testing.T) {
	for _, ignored := range []bool{false, true} {
		c, b := setup(t)
		c.state.StartMode = "broadcast"
		b.failBroadcast = !ignored
		b.ignoreBroadcast = ignored
		begin(t, c)
		if ignored {
			c.deadline = time.Now().Add(-time.Second)
			c.tick(time.Now())
		}
		assertStops(t, c, b)
		n := 0
		for _, w := range b.writes {
			if w == "broadcast:start" {
				n++
			}
		}
		if n != 1 {
			t.Fatal(b.writes)
		}
	}
}

func TestBroadcastNotSentAfterProgrammingFailure(t *testing.T) {
	c, b := setup(t)
	c.state.StartMode = "broadcast"
	b.failProgram = 2
	begin(t, c)
	if contains(b.writes, "broadcast:start") {
		t.Fatal("broadcast after partial preparation")
	}
	assertStops(t, c, b)
}

func (b *recordingBus) Read(sl byte, address, count uint16) ([]uint16, error) {
	if b.failRead {
		b.failRead = false
		return nil, fmt.Errorf("telemetry timeout")
	}
	if sl == 1 && address == 0x6002 && b.pendingDisabled != 0 {
		return []uint16{b.pendingDisabled}, nil
	}
	v, e := b.SimBus.Read(sl, address, count)
	if e == nil && sl == 1 && address == 0x1003 && b.disableLagReads > 0 {
		v[0] |= 2
		b.disableLagReads--
	}
	if e == nil && sl == 1 && address == 0x017b && b.holdBrakeOutput {
		v[0] |= 1
	}
	return v, e
}

func (b *recordingBus) Write(sl byte, r, v uint16) error {
	b.writes = append(b.writes, fmt.Sprintf("%d:%04x=%04x", sl, r, v))
	if sl == 1 && r == 0xf && v == 1 {
		b.drives[0].position += float64(b.enableShift)
	}
	if sl == 1 && b.pendingDisabled != 0 {
		if r == 0x6002 && v == 0x40 {
			b.pendingDisabled = 0x40
		}
		if r == 0xf && v == 1 {
			if b.pendingDisabled != 0x40 {
				return fmt.Errorf("old path could resume")
			}
			b.pendingDisabled = 0
		}
	}
	// Apply the trigger before reporting failure to model a lost acknowledgement.
	e := b.SimBus.Write(sl, r, v)
	if r == 0x6002 && v == 16 && sl == b.failTrigger {
		return fmt.Errorf("lost trigger ACK")
	}
	return e
}
func (b *recordingBus) Program(sl byte, w []uint16) error {
	b.writes = append(b.writes, fmt.Sprintf("program:%d", sl))
	if sl == b.failProgram {
		return fmt.Errorf("program failed")
	}
	return b.SimBus.Program(sl, w)
}
func setup(t *testing.T) (*Controller, *recordingBus) {
	t.Helper()
	b := &recordingBus{SimBus: NewSimBus()}
	c, e := New(b, true)
	if e != nil {
		t.Fatal(e)
	}
	return c, b
}
func contains(v []string, s string) bool {
	for _, x := range v {
		if x == s {
			return true
		}
	}
	return false
}
func begin(t *testing.T, c *Controller) {
	t.Helper()
	if e := c.Reference(); e != nil {
		t.Fatal(e)
	}
	if e := c.Move(Point{.1, 0}, .02); e != nil {
		t.Fatal(e)
	}
	c.readyAt = time.Now().Add(-time.Second)
	c.tick(time.Now())
}
func TestNoStartupMotionAndReferenceRequired(t *testing.T) {
	c, b := setup(t)
	if len(b.writes) != 0 {
		t.Fatal(b.writes)
	}
	if e := c.Move(Point{.1, 0}, .02); e == nil {
		t.Fatal("unreferenced move accepted")
	}
	if e := c.Reference(); e != nil {
		t.Fatal(e)
	}
	if len(b.writes) != 0 {
		t.Fatal("reference wrote registers")
	}
}

func TestDisabledPendingPathIsStoppedBeforeEnable(t *testing.T) {
	c, b := setup(t)
	b.pendingDisabled = 0x200
	c.tick(time.Now())
	if len(b.writes) != 0 || c.state.Phase != "unreferenced" {
		t.Fatalf("startup wrote registers: %v", b.writes)
	}
	begin(t, c)
	if len(b.writes) < 2 || b.writes[0] != "1:6002=0040" || b.writes[1] != "1:000f=0001" {
		t.Fatalf("did not cancel pending path before enable: %v", b.writes)
	}
	if c.state.Phase != "moving" {
		t.Fatalf("phase %s: %s", c.state.Phase, c.state.Error)
	}
}
func TestPrepareBothBeforeTriggersAndBrakeLifecycle(t *testing.T) {
	c, b := setup(t)
	begin(t, c)
	want := []string{"1:000f=0001", "program:1", "program:2", "1:6002=0010", "2:6002=0010"}
	if strings.Join(b.writes, ",") != strings.Join(want, ",") {
		t.Fatalf("wrong sequence %v", b.writes)
	}
	if e := c.Move(Point{.2, 0}, .02); e == nil {
		t.Fatal("overlapping move accepted")
	}
	for i := range b.drives {
		b.drives[i].began = time.Now().Add(-time.Minute)
	}
	c.tick(time.Now())
	if c.state.Phase != "settling" {
		t.Fatalf("phase %s", c.state.Phase)
	}
	c.settleAt = time.Now().Add(-time.Second)
	c.tick(time.Now())
	if c.state.Phase != "disabling" {
		t.Fatalf("missing disable wait: %s", c.state.Phase)
	}
	c.tick(c.disableNotBefore.Add(time.Millisecond))
	if c.state.Phase != "idle" || c.state.Drums[0].Enabled {
		t.Fatalf("not disabled and idle: %+v", c.state)
	}
	if !contains(b.writes, "1:000f=0000") {
		t.Fatal("enable not restored")
	}
	for _, w := range b.writes {
		if strings.Contains(w, "1801") {
			t.Fatal("unexpected EEPROM/control write")
		}
	}
}

func readyToDisable(t *testing.T) (*Controller, *recordingBus) {
	t.Helper()
	c, b := setup(t)
	begin(t, c)
	for i := range b.drives {
		b.drives[i].began = time.Now().Add(-time.Minute)
	}
	c.tick(time.Now())
	c.settleAt = time.Now().Add(-time.Second)
	c.tick(time.Now())
	if c.state.Phase != "disabling" {
		t.Fatalf("phase %s", c.state.Phase)
	}
	return c, b
}

func TestDelayedDisableAcknowledgementDoesNotFault(t *testing.T) {
	c, b := readyToDisable(t)
	b.disableLagReads = 2
	for i := 0; i < 2; i++ {
		c.tick(c.disableNotBefore.Add(time.Duration(i) * 100 * time.Millisecond))
		if c.state.Phase != "disabling" || c.state.Error != "" {
			t.Fatalf("premature fault: %+v", c.state)
		}
		if e := c.Move(Point{.2, 0}, .02); e == nil {
			t.Fatal("move accepted during disable")
		}
	}
	c.tick(c.disableNotBefore.Add(300 * time.Millisecond))
	if c.state.Phase != "idle" || c.enabledByUs || !c.state.Referenced {
		t.Fatalf("did not complete delayed disable: %+v", c.state)
	}
	n := 0
	for _, w := range b.writes {
		if w == "1:000f=0000" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("disable rewritten %d times", n)
	}
}

func TestDisableTimeoutAndBrakeOutputConfirmation(t *testing.T) {
	for _, brake := range []bool{false, true} {
		c, b := readyToDisable(t)
		if brake {
			b.holdBrakeOutput = true
		} else {
			b.disableLagReads = 100
		}
		c.tick(c.disableDeadline.Add(time.Millisecond))
		if c.state.Phase != "fault" || c.state.Referenced || !c.enabledByUs {
			t.Fatalf("unconfirmed disable accepted: %+v", c.state)
		}
	}
}
func TestPartialProgramFailureNeverTriggers(t *testing.T) {
	c, b := setup(t)
	b.failProgram = 2
	begin(t, c)
	if contains(b.writes, "1:6002=0010") || contains(b.writes, "2:6002=0010") {
		t.Fatal("triggered partially programmed plan")
	}
	assertStops(t, c, b)
}
func assertStops(t *testing.T, c *Controller, b *recordingBus) {
	t.Helper()
	if !contains(b.writes, "1:6002=0040") || !contains(b.writes, "2:6002=0040") || c.state.Referenced {
		t.Fatalf("missing paired stop/invalidation %v", b.writes)
	}
}
func TestLostTriggerAcknowledgementStopsBothWithoutRetry(t *testing.T) {
	c, b := setup(t)
	b.failTrigger = 2
	begin(t, c)
	assertStops(t, c, b)
	n := 0
	for _, w := range b.writes {
		if w == "2:6002=0010" {
			n++
		}
	}
	if n != 1 {
		t.Fatal("trigger retried")
	}
}

func TestTelemetryFailureStopsBoth(t *testing.T) {
	c, b := setup(t)
	begin(t, c)
	b.failRead = true
	c.tick(time.Now())
	assertStops(t, c, b)
}

func TestDeadlineStopsBoth(t *testing.T) {
	c, b := setup(t)
	begin(t, c)
	c.deadline = time.Now().Add(-time.Second)
	c.tick(time.Now())
	assertStops(t, c, b)
}

func TestEncoderFaultStopsBoth(t *testing.T) {
	c, b := setup(t)
	begin(t, c)
	b.drives[1].regs[0x2203] = 8
	c.tick(time.Now())
	assertStops(t, c, b)
}
func TestStopDuringBrakeDelayNeverStartsMotion(t *testing.T) {
	c, b := setup(t)
	c.Reference()
	if e := c.Move(Point{.1, 0}, .02); e != nil {
		t.Fatal(e)
	}
	c.Stop()
	c.readyAt = time.Now().Add(-time.Second)
	c.tick(time.Now())
	if contains(b.writes, "1:6002=0010") {
		t.Fatal("started after stop")
	}
	assertStops(t, c, b)
}
func TestOneStationaryDrumIsNotTriggered(t *testing.T) {
	c, b := setup(t)
	c.Reference()
	if e := c.Move(Point{.1, -.1}, .02); e != nil {
		t.Fatal(e)
	}
	c.readyAt = time.Now().Add(-time.Second)
	c.tick(time.Now())
	if contains(b.writes, "2:6002=0010") || contains(b.writes, "program:2") {
		t.Fatal("stationary drum was programmed/triggered")
	}
}

func TestRightLowerRaiseLeftKeepsStationaryDrumBraked(t *testing.T) {
	c, b := setup(t)
	b.enableShift = 22
	if e := c.Reference(); e != nil {
		t.Fatal(e)
	}
	for _, target := range []Point{{.1, 0}, {.1, .1}, {.1, 0}, {0, 0}} {
		b.writes = nil
		if e := c.Move(target, .02); e != nil {
			t.Fatal(e)
		}
		stationary := c.state.Plan.Profiles[0].Delta == 0
		c.readyAt = time.Now().Add(-time.Second)
		c.tick(time.Now())
		if c.state.Phase != "moving" {
			t.Fatalf("start %+v: %s", target, c.state.Error)
		}
		for i := range b.drives {
			b.drives[i].began = time.Now().Add(-time.Minute)
		}
		c.tick(time.Now())
		c.settleAt = time.Now().Add(-time.Second)
		c.tick(time.Now())
		if c.state.Phase == "disabling" {
			// Drive may settle one count away while engaging the brake.
			b.drives[0].position++
			c.tick(c.disableNotBefore.Add(time.Millisecond))
		}
		if c.state.Phase != "idle" || !c.state.Referenced {
			t.Fatalf("finish %+v: %s / %s", target, c.state.Phase, c.state.Error)
		}
		if stationary {
			for _, w := range b.writes {
				if strings.HasPrefix(w, "1:") || w == "program:1" {
					t.Fatalf("stationary drum written: %s", w)
				}
			}
			if c.state.Drums[0].Enabled {
				t.Fatal("stationary drum enabled")
			}
		}
	}
}

func TestStationaryDrumMovementStillStopsBoth(t *testing.T) {
	c, b := setup(t)
	c.Reference()
	if e := c.Move(Point{0, .1}, .02); e != nil {
		t.Fatal(e)
	}
	c.readyAt = time.Now().Add(-time.Second)
	c.tick(time.Now())
	b.drives[0].position += 22
	c.tick(time.Now())
	assertStops(t, c, b)
	if c.state.Error != "stationary drum 1 moved" {
		t.Fatal(c.state.Error)
	}
}
func TestCounterResetInvalidatesReference(t *testing.T) {
	c, b := setup(t)
	b.drives[0].position = 10000
	c.Reference()
	b.drives[0].position = 0
	c.tick(time.Now())
	assertStops(t, c, b)
}
func TestHTTPRejectsCrossOriginMissingFieldsAndDuplicateMove(t *testing.T) {
	c, _ := setup(t)
	h := c.Handler()
	for _, tc := range []struct {
		path, body, origin string
		code               int
	}{{"move", `{"x":0.1,"z":0,"speed":0.02}`, "http://evil.test", 403}, {"move", `{"x":0.1}`, "", 400}, {"move", `{"x":0.1,"z":0,"speed":0.02}`, "", 409}} {
		r := httptest.NewRequest("POST", "http://localhost/api/"+tc.path, strings.NewReader(tc.body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", tc.origin)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.code {
			t.Fatalf("status %d: %s", w.Code, w.Body.String())
		}
	}
}
