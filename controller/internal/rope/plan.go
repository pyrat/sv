// Package rope implements unloaded commissioning of the two winding drums.
package rope

import (
	"fmt"
	"math"
)

const (
	Diameter      = 0.080 // Virtual bare-barrel diameter, metres; no rope calibration.
	GearRatio     = 20.0
	PulsesPerTurn = 10000.0 // Verified command and reported-position scale on both drives.
	MaxStep       = 0.25
	MaxSpeed      = 0.05
	MaxRPM        = 240
)

var metresPerPulse = math.Pi * Diameter / GearRatio / PulsesPerTurn

type Point struct {
	X float64 `json:"x"`
	Z float64 `json:"z"`
}

type Profile struct {
	Target  int32   `json:"target"`
	Delta   int64   `json:"delta"`
	RPM     uint16  `json:"rpm"`
	Ramp    uint16  `json:"ramp"`
	Seconds float64 `json:"seconds"`
}

type Plan struct {
	Target           Point      `json:"target"`
	Profiles         [2]Profile `json:"profiles"`
	Seconds          float64    `json:"seconds"`
	DurationMismatch float64    `json:"durationMismatch"`
}

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }

// Coordinates uses independent feedback origins; capturing a virtual origin does
// not write zero/home registers or pretend the mechanism has physically homed.
func Coordinates(counts, origin [2]int32) Point {
	d1 := float64(int64(counts[0])-int64(origin[0])) * metresPerPulse
	d2 := float64(int64(counts[1])-int64(origin[1])) * metresPerPulse
	return Point{X: d1, Z: -d1 - d2}
}

// MakePlan converts a Cartesian target to absolute command counters. Each drive
// runs a trapezoid with a common nominal ramp time and proportional cruise speed.
// Integer RPM/ramp quantization is explicit: reject profiles whose completion
// times diverge by >25 ms rather than silently claiming exact synchronization.
func MakePlan(target, from Point, speed float64, origin, current [2]int32) (Plan, error) {
	return makePlan(target, from, speed, origin, current, nil)
}

// Unchanged rope targets may hold a settled counter within 25 counts (~0.031 mm)
// of the absolute ideal. This is not a per-move deadband: changed targets still
// use absolute origin-based counters, and offsets cannot accumulate each move.
func makePlan(target, from Point, speed float64, origin, current [2]int32, previous *Point) (Plan, error) {
	p := Plan{Target: target}
	if !finite(target.X) || !finite(target.Z) || !finite(speed) || speed < 0.005 || speed > MaxSpeed {
		return p, fmt.Errorf("finite targets and speed 0.005..%.3f m/s required", MaxSpeed)
	}
	if math.Abs(target.X) > 1 || math.Abs(target.Z) > 1 {
		return p, fmt.Errorf("unloaded virtual coordinates must stay within ±1 m")
	}
	distance := math.Hypot(target.X-from.X, target.Z-from.Z)
	if distance > MaxStep+1e-9 || distance < 0.0001 {
		return p, fmt.Errorf("move length must be 0.1..250 mm in virtual X/Z space")
	}
	targetTurns := [2]float64{target.X / metresPerPulse, -(target.X + target.Z) / metresPerPulse}
	cruiseTime := math.Max(distance/speed, 0.25)
	var turns [2]float64
	maxTurns := 0.0
	for i := range p.Profiles {
		v := math.Round(targetTurns[i]) + float64(origin[i])
		if v < math.MinInt32 || v > math.MaxInt32 {
			return p, fmt.Errorf("target exceeds signed counter range")
		}
		m := &p.Profiles[i]
		m.Target = int32(v)
		m.Delta = int64(m.Target) - int64(current[i])
		if previous != nil {
			prior := [2]float64{previous.X / metresPerPulse, -(previous.X + previous.Z) / metresPerPulse}
			if math.Round(prior[i]) == math.Round(targetTurns[i]) {
				if math.Abs(float64(m.Delta)) > 25 {
					return p, fmt.Errorf("stationary drum %d is more than 25 counts from its unchanged target; recapture origin after inspection", i+1)
				}
				m.Target = current[i]
				m.Delta = 0
			}
		}
		turns[i] = math.Abs(float64(m.Delta)) / PulsesPerTurn
		maxTurns = math.Max(maxTurns, turns[i])
	}
	// Select the fastest compatible integer RPM pair no faster than requested.
	// A 1:2 distance ratio selects e.g. 67/134 rather than mismatched 68/135 RPM.
	limit := int(math.Min(MaxRPM, math.Floor(maxTurns*60/cruiseTime)))
	for dominantRPM := limit; dominantRPM >= 4; dominantRPM-- {
		candidate := p
		valid := true
		for i := range candidate.Profiles {
			m := &candidate.Profiles[i]
			if m.Delta == 0 {
				continue
			}
			rpm := math.Round(turns[i] / maxTurns * float64(dominantRPM))
			if rpm < 4 || rpm > MaxRPM {
				valid = false
				break
			}
			m.RPM = uint16(rpm)
			m.Ramp = uint16(math.Round(250000 / rpm))
			accel := float64(m.Ramp) * rpm / 1e6
			constantSpeedTime := turns[i] * 60 / rpm
			if constantSpeedTime < accel || constantSpeedTime+1e-9 < cruiseTime {
				valid = false
				break
			}
			m.Seconds = constantSpeedTime + accel
			candidate.Seconds = math.Max(candidate.Seconds, m.Seconds)
		}
		if !valid {
			continue
		}
		if candidate.Profiles[0].Delta != 0 && candidate.Profiles[1].Delta != 0 {
			candidate.DurationMismatch = math.Abs(candidate.Profiles[0].Seconds - candidate.Profiles[1].Seconds)
		}
		if candidate.DurationMismatch <= 0.025 && candidate.Seconds <= 30 {
			return candidate, nil
		}
	}
	return p, fmt.Errorf("no matched 4..%d RPM profiles within 25 ms mismatch and 30 s duration; change distance/speed", MaxRPM)
}
