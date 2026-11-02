package rope

import (
	"bytes"
	"encoding/csv"
	"math"
	"testing"
	"time"
)

func TestSequentialSamplesAlignWithoutFalseVerticalMotion(t *testing.T) {
	base := time.Unix(100, 0)
	trace := newTrace(Point{}, Point{X: .1}, [2]int32{}, [2]RawSample{})
	// Both drums move at equal opposite speed, but are sampled 20 ms apart.
	for _, ms := range []int64{0, 100, 200} {
		for i := 0; i < 2; i++ {
			offset := ms + int64(i)*20
			count := int32(offset * 10)
			if i == 1 {
				count = -count
			}
			at := base.Add(time.Duration(offset) * time.Millisecond)
			trace.add(i, RawSample{At: at, ReadStart: at.Add(-5 * time.Millisecond), ReadEnd: at.Add(5 * time.Millisecond), Feedback: count})
		}
		trace.align()
	}
	if len(trace.Points) != 2 {
		t.Fatalf("points: %+v", trace.Points)
	}
	for _, p := range trace.Points {
		if math.Abs(p.Z) > 1e-12 {
			t.Fatalf("false vertical deviation: %+v", p)
		}
	}
	if trace.MaxReadHalfWindowMS != 5 || trace.MaxBracketMS != 100 {
		t.Fatalf("timing metrics: %+v", trace)
	}
	copy := trace.clone()
	copy.Raw[0][0].Feedback = 999
	copy.Points[0].X = 999
	if trace.Raw[0][0].Feedback == 999 || trace.Points[0].X == 999 {
		t.Fatal("snapshot aliases trace")
	}
	var buf bytes.Buffer
	if err := trace.CSV(&buf); err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 9 {
		t.Fatalf("CSV rows %d", len(rows))
	}
}

func TestInterpolationRejectsExtrapolationAndGaps(t *testing.T) {
	base := time.Unix(100, 0)
	samples := []RawSample{{At: base}, {At: base.Add(time.Second), Feedback: 100}}
	for _, at := range []time.Time{base.Add(-time.Millisecond), base.Add(100 * time.Millisecond), base.Add(2 * time.Second)} {
		if _, _, ok := interpolate(samples, at); ok {
			t.Fatalf("invalid interpolation at %v", at)
		}
	}
}

func TestTraceMetricsAndInterruptedOutcome(t *testing.T) {
	if got := crossTrack(Point{X: .05, Z: .003}, Point{}, Point{X: .1}); math.Abs(got-.003) > 1e-12 {
		t.Fatal(got)
	}
	now := time.Now()
	trace := newTrace(Point{}, Point{X: .1}, [2]int32{}, [2]RawSample{})
	trigger := now.Add(-time.Second)
	trace.TriggerAt = &trigger
	trace.motionEnded(now)
	trace.Points = []TracePoint{{At: now, Point: Point{X: .099}}}
	trace.finish("completed", now)
	if trace.FinalErrorM == nil || math.Abs(*trace.FinalErrorM-.001) > 1e-12 || *trace.DurationSeconds != 1 {
		t.Fatalf("metrics %+v", trace)
	}
	interrupted := newTrace(Point{}, Point{X: .1}, [2]int32{}, [2]RawSample{})
	interrupted.Points = trace.Points
	interrupted.finish("stopped", now)
	if interrupted.FinalErrorM != nil {
		t.Fatal("interrupted move reported final error")
	}
}
