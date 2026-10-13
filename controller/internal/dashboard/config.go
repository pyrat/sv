package dashboard

import (
	"fmt"
	"math"
)

type AxisConfig struct {
	Port                   string  `json:"port"`
	Slave                  byte    `json:"slave"`
	Baud                   int     `json:"baud"`
	PulsesPerMetre         float64 `json:"pulses_per_metre"`
	FeedbackCountsPerMetre float64 `json:"feedback_counts_per_metre"`
	PulsesPerRevolution    float64 `json:"pulses_per_revolution"`
	Min                    float64 `json:"min_m"`
	Max                    float64 `json:"max_m"`
	MaxSpeed               float64 `json:"max_speed_mps"`
	Direction              int     `json:"direction"`
	Ramp                   uint16  `json:"ramp_ms_per_1000rpm"`
	ReferenceValid         bool    `json:"reference_valid"`
	ReferenceID            string  `json:"reference_id"`
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func (c AxisConfig) validate() error {
	for _, v := range []float64{c.PulsesPerMetre, c.FeedbackCountsPerMetre, c.PulsesPerRevolution, c.Min, c.Max, c.MaxSpeed} {
		if !finite(v) {
			return fmt.Errorf("calibration must be finite")
		}
	}
	if c.PulsesPerMetre <= 0 || c.FeedbackCountsPerMetre <= 0 || c.PulsesPerRevolution <= 0 || c.Min >= c.Max || c.MaxSpeed <= 0 || (c.Direction != 1 && c.Direction != -1) || c.Ramp == 0 {
		return fmt.Errorf("invalid calibration, direction, speed or travel limits")
	}
	if math.Max(math.Abs(c.Min), math.Abs(c.Max))*c.PulsesPerMetre > math.MaxInt32 {
		return fmt.Errorf("travel exceeds signed command range")
	}
	if c.MaxSpeed*c.PulsesPerMetre*60/c.PulsesPerRevolution > 5000 {
		return fmt.Errorf("speed exceeds driver RPM range")
	}
	return nil
}

func DemoConfig() map[string]AxisConfig {
	x := AxisConfig{PulsesPerMetre: 53000, FeedbackCountsPerMetre: 132500, PulsesPerRevolution: 1600, Min: 0, Max: 50, MaxSpeed: 0.2, Direction: 1, Ramp: 1000, ReferenceValid: true}
	y := x
	y.Max = 30
	return map[string]AxisConfig{"x": x, "y": y}
}
