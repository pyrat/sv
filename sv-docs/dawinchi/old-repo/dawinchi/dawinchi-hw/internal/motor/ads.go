package motor

import "github.com/CreateView/dawinchi-hw/internal/ads1115"

type ADS interface {
	Read(channel int) (float64, error)
}

func (m *Motor) ReadVoltage() (float64, error) {
	return m.ads.Read(m.adsMUX)
}

func VoltageToAmps(v float64) float64 {
	return ads1115.VoltageToAmps(v)
}
