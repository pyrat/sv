package rope

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"time"
)

const traceLimit = 2000
const maxAlignmentGap = 500 * time.Millisecond

type RawSample struct {
	At        time.Time `json:"at"`
	ReadStart time.Time `json:"readStart"`
	ReadEnd   time.Time `json:"readEnd"`
	Profile   int32     `json:"profile"`
	Feedback  int32     `json:"feedback"`
}

type TracePoint struct {
	At time.Time `json:"at"`
	Point
	Feedback  [2]float64 `json:"feedback"`
	BracketMS float64    `json:"bracketMs"`
}

type MoveTrace struct {
	StartMode           string         `json:"startMode"`
	ID                  string         `json:"id"`
	CreatedAt           time.Time      `json:"createdAt"`
	TriggerAt           *time.Time     `json:"triggerAt"`
	MotionEndedAt       *time.Time     `json:"motionEndedAt"`
	EndedAt             *time.Time     `json:"endedAt"`
	Outcome             string         `json:"outcome"`
	From                Point          `json:"from"`
	Target              Point          `json:"target"`
	FeedbackOrigin      [2]int32       `json:"feedbackOrigin"`
	Raw                 [2][]RawSample `json:"raw"`
	Points              []TracePoint   `json:"points"`
	MaxCrossTrackM      float64        `json:"maxCrossTrackM"`
	FinalErrorM         *float64       `json:"finalErrorM"`
	DurationSeconds     *float64       `json:"durationSeconds"`
	MaxReadHalfWindowMS float64        `json:"maxReadHalfWindowMs"`
	MaxBracketMS        float64        `json:"maxBracketMs"`
	TriggerSkewMS       float64        `json:"triggerSkewMs"`
	SkippedAlignments   int            `json:"skippedAlignments"`
	Truncated           bool           `json:"truncated"`
}

func newTrace(from, target Point, origin [2]int32, initial [2]RawSample) *MoveTrace {
	now := time.Now()
	t := &MoveTrace{ID: now.UTC().Format("20060102T150405.000000000Z"), CreatedAt: now, Outcome: "preparing", From: from, Target: target, FeedbackOrigin: origin, Points: []TracePoint{}}
	for i, s := range initial {
		t.Raw[i] = []RawSample{}
		if !s.At.IsZero() {
			t.add(i, s)
		}
	}
	return t
}

func (t *MoveTrace) add(i int, s RawSample) {
	if t.EndedAt != nil {
		return
	}
	v := t.Raw[i]
	if len(v) > 0 && !s.At.After(v[len(v)-1].At) {
		return
	}
	t.Raw[i] = append(v, s)
	if len(t.Raw[i]) > traceLimit {
		t.Raw[i] = t.Raw[i][1:]
		t.Truncated = true
	}
	t.MaxReadHalfWindowMS = math.Max(t.MaxReadHalfWindowMS, float64(s.ReadEnd.Sub(s.ReadStart))/float64(2*time.Millisecond))
}

// Interpolate only between known samples, never extrapolate across the leading
// edge of the sequential reads or bridge a large telemetry gap.
func interpolate(samples []RawSample, at time.Time) (float64, time.Duration, bool) {
	i := sort.Search(len(samples), func(i int) bool { return !samples[i].At.Before(at) })
	if i == len(samples) {
		return 0, 0, false
	}
	if samples[i].At.Equal(at) {
		return float64(samples[i].Feedback), 0, true
	}
	if i == 0 {
		return 0, 0, false
	}
	a, b := samples[i-1], samples[i]
	gap := b.At.Sub(a.At)
	if gap <= 0 || gap > maxAlignmentGap {
		return 0, gap, false
	}
	f := float64(at.Sub(a.At)) / float64(gap)
	return float64(a.Feedback) + f*float64(int64(b.Feedback)-int64(a.Feedback)), gap, true
}

func crossTrack(p, a, b Point) float64 {
	dx, dz := b.X-a.X, b.Z-a.Z
	length := math.Hypot(dx, dz)
	if length == 0 {
		return math.Hypot(p.X-a.X, p.Z-a.Z)
	}
	return math.Abs(dx*(p.Z-a.Z)-dz*(p.X-a.X)) / length
}

func (t *MoveTrace) align() {
	if t.EndedAt != nil || len(t.Raw[0]) == 0 || len(t.Raw[1]) == 0 {
		return
	}
	at := t.Raw[0][len(t.Raw[0])-1].At
	if b := t.Raw[1][len(t.Raw[1])-1].At; b.Before(at) {
		at = b
	}
	if len(t.Points) > 0 && !at.After(t.Points[len(t.Points)-1].At) {
		return
	}
	var feedback [2]float64
	var maxGap time.Duration
	for i := range feedback {
		v, gap, ok := interpolate(t.Raw[i], at)
		if !ok {
			t.SkippedAlignments++
			return
		}
		feedback[i] = v
		if gap > maxGap {
			maxGap = gap
		}
	}
	d1 := (feedback[0] - float64(t.FeedbackOrigin[0])) * metresPerPulse
	d2 := (feedback[1] - float64(t.FeedbackOrigin[1])) * metresPerPulse
	p := TracePoint{At: at, Point: Point{X: d1, Z: -d1 - d2}, Feedback: feedback, BracketMS: float64(maxGap) / float64(time.Millisecond)}
	t.Points = append(t.Points, p)
	if len(t.Points) > traceLimit {
		t.Points = t.Points[1:]
		t.Truncated = true
	}
	t.MaxBracketMS = math.Max(t.MaxBracketMS, p.BracketMS)
	// Include the entire observed path (also settling), not just endpoint error.
	t.MaxCrossTrackM = math.Max(t.MaxCrossTrackM, crossTrack(p.Point, t.From, t.Target))
}

func (t *MoveTrace) motionEnded(now time.Time) {
	if t.MotionEndedAt != nil {
		return
	}
	t.MotionEndedAt = &now
	if t.TriggerAt != nil {
		v := now.Sub(*t.TriggerAt).Seconds()
		t.DurationSeconds = &v
	}
}
func (t *MoveTrace) finish(outcome string, now time.Time) {
	if t.EndedAt != nil {
		return
	}
	t.EndedAt = &now
	t.Outcome = outcome
	if len(t.Points) > 0 && outcome == "completed" {
		last := t.Points[len(t.Points)-1]
		if now.Sub(last.At) <= maxAlignmentGap {
			v := math.Hypot(last.X-t.Target.X, last.Z-t.Target.Z)
			t.FinalErrorM = &v
		}
	}
}
func (t *MoveTrace) clone() *MoveTrace {
	if t == nil {
		return nil
	}
	v := *t
	for i := range t.Raw {
		v.Raw[i] = append([]RawSample{}, t.Raw[i]...)
	}
	v.Points = append([]TracePoint{}, t.Points...)
	// Timestamp/metric pointers refer to immutable values, replaced on updates.
	return &v
}

func (t *MoveTrace) CSV(w io.Writer) error {
	c := csv.NewWriter(w)
	header := []string{"trace_id", "outcome", "kind", "time_utc", "seconds_from_trigger", "drum", "profile_count", "feedback_count", "aligned_drum1_count", "aligned_drum2_count", "x_m", "z_m", "read_start_utc", "read_end_utc", "interpolation_bracket_ms", "start_x_m", "start_z_m", "target_x_m", "target_z_m", "origin_drum1_count", "origin_drum2_count", "trigger_skew_ms"}
	header = append(header, "start_mode")
	if e := c.Write(header); e != nil {
		return e
	}
	f := func(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
	stamp := func(v time.Time) string { return v.UTC().Format(time.RFC3339Nano) }
	row := func(at time.Time, kind string) []string {
		r := make([]string, len(header))
		r[0] = t.ID
		r[1] = t.Outcome
		r[2] = kind
		r[3] = stamp(at)
		if t.TriggerAt != nil {
			r[4] = f(at.Sub(*t.TriggerAt).Seconds())
		}
		r[15] = f(t.From.X)
		r[16] = f(t.From.Z)
		r[17] = f(t.Target.X)
		r[18] = f(t.Target.Z)
		r[19] = fmt.Sprint(t.FeedbackOrigin[0])
		r[20] = fmt.Sprint(t.FeedbackOrigin[1])
		r[21] = f(t.TriggerSkewMS)
		r[22] = t.StartMode
		return r
	}
	type timedRow struct {
		at  time.Time
		row []string
	}
	var rows []timedRow
	for i, samples := range t.Raw {
		for _, s := range samples {
			r := row(s.At, "raw")
			r[5] = fmt.Sprint(i + 1)
			r[6] = fmt.Sprint(s.Profile)
			r[7] = fmt.Sprint(s.Feedback)
			r[12] = stamp(s.ReadStart)
			r[13] = stamp(s.ReadEnd)
			rows = append(rows, timedRow{s.At, r})
		}
	}
	for _, p := range t.Points {
		r := row(p.At, "aligned")
		r[8] = f(p.Feedback[0])
		r[9] = f(p.Feedback[1])
		r[10] = f(p.X)
		r[11] = f(p.Z)
		r[14] = f(p.BracketMS)
		rows = append(rows, timedRow{p.At, r})
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].at.Before(rows[j].at) })
	for _, r := range rows {
		if e := c.Write(r.row); e != nil {
			return e
		}
	}
	c.Flush()
	return c.Error()
}
