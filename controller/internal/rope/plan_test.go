package rope

import (
	"math"
	"testing"
)

func TestRopeDirectionsAndInverse(t *testing.T) {
	for _, tc := range []struct {
		name string
		p    Point
		sign [2]int
	}{{"right", Point{.1, 0}, [2]int{1, -1}}, {"left", Point{-.1, 0}, [2]int{-1, 1}}, {"lower", Point{0, .1}, [2]int{0, -1}}, {"raise", Point{0, -.1}, [2]int{0, 1}}, {"right and raise", Point{.1, -.1}, [2]int{1, 0}}} {
		t.Run(tc.name, func(t *testing.T) {
			origin := [2]int32{123456, -654321}
			p, e := MakePlan(tc.p, Point{}, .02, origin, origin)
			if e != nil {
				t.Fatal(e)
			}
			var targets [2]int32
			for i, m := range p.Profiles {
				targets[i] = m.Target
				sign := 0
				if m.Delta > 0 {
					sign = 1
				} else if m.Delta < 0 {
					sign = -1
				}
				if sign != tc.sign[i] {
					t.Fatalf("drum %d sign %d", i, sign)
				}
			}
			got := Coordinates(targets, origin)
			if math.Abs(got.X-tc.p.X) > metresPerPulse || math.Abs(got.Z-tc.p.Z) > 2*metresPerPulse {
				t.Fatalf("inverse: %+v", got)
			}
		})
	}
}
func TestPlanLimits(t *testing.T) {
	for _, tc := range []struct {
		p      Point
		speed  float64
		origin [2]int32
	}{{Point{math.NaN(), 0}, .02, [2]int32{}}, {Point{.3, 0}, .02, [2]int32{}}, {Point{.1, 0}, 0, [2]int32{}}, {Point{.1, 0}, .1, [2]int32{}}, {Point{.1, 0}, .02, [2]int32{math.MaxInt32, 0}}} {
		_, e := MakePlan(tc.p, Point{}, tc.speed, tc.origin, tc.origin)
		if e == nil {
			t.Fatalf("accepted invalid plan %+v", tc)
		}
	}
}
func TestMatchedTrapezoids(t *testing.T) {
	p, e := MakePlan(Point{.1, 0}, Point{}, .02, [2]int32{}, [2]int32{})
	if e != nil {
		t.Fatal(e)
	}
	if p.Profiles[0].RPM != p.Profiles[1].RPM || p.Profiles[0].Ramp != p.Profiles[1].Ramp || p.DurationMismatch != 0 {
		t.Fatalf("unmatched horizontal profiles: %+v", p)
	}
	if p.Seconds < 5 || p.Seconds > 5.5 {
		t.Fatalf("unexpected duration %f", p.Seconds)
	}
}

func TestUnequalDrumDistancesChooseCompatibleRPMs(t *testing.T) {
	p, e := MakePlan(Point{.1, .1}, Point{}, .02, [2]int32{}, [2]int32{})
	if e != nil {
		t.Fatal(e)
	}
	if p.Profiles[1].RPM != 2*p.Profiles[0].RPM || p.DurationMismatch > .025 {
		t.Fatalf("bad coordinated profiles %+v", p)
	}
	if p.Seconds < math.Hypot(.1, .1)/.02 {
		t.Fatal("exceeded requested speed")
	}
}

func TestUnchangedTargetHoldsSettledCounterWithoutAccumulatingDrift(t *testing.T) {
	origin := [2]int32{}
	previous := Point{.1, 0}
	right, e := MakePlan(previous, Point{}, .02, origin, origin)
	if e != nil {
		t.Fatal(e)
	}
	for _, offset := range []int32{-25, -1, 1, 25} {
		current := [2]int32{right.Profiles[0].Target + offset, right.Profiles[1].Target}
		for _, target := range []Point{{.1, .1}, {.1, 0}, {.1, .1}} {
			p, e := makePlan(target, Coordinates(current, origin), .02, origin, current, &previous)
			if e != nil {
				t.Fatal(e)
			}
			if p.Profiles[0].Delta != 0 || p.Profiles[0].Target != current[0] {
				t.Fatal("stationary counter corrected")
			}
			current[1] = p.Profiles[1].Target
			previous = target
		}
		current[0] = right.Profiles[0].Target + 26
		if _, e := makePlan(Point{.1, .2}, Coordinates(current, origin), .02, origin, current, &previous); e == nil {
			t.Fatal("excessive absolute drift accepted")
		}
		previous = Point{.1, 0}
	}
}
