/*
ADS1115 is a 16-bit analog-digital converter.
More details about ADS1115, please ref to: https://www.ti.com/lit/ds/symlink/ads1115.pdf

Config Raspberry Pi:
 1. $ sudo apt-get install -y python-smbus
 2. $ sudo apt-get install -y i2c-tools
 3. $ sudo raspi-config
 4. -> [5 Interface Options] -> [P5 I2C] -> [yes] -> [ok]
 5. $ sudo reboot now
 6. check: $ sudo i2cdetect -y 1
    it works if you saw following message:
    ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
    0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f
    00:          -- -- -- -- -- -- -- -- -- -- -- -- --
    10: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
    20: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
    30: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
    40: -- -- -- -- -- -- -- -- 48 -- -- -- -- -- -- --
    50: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
    60: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
    70: -- -- -- -- -- -- -- --
    ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Connect to Raspberry Pi:
- VCC: any 3.3v pin
- GND: any GND pin
- SDA: GPIO-22 (SDA)
- SCL: GPIO-23 (SCL)
*/
package ads1115

import (
	"github.com/pkg/errors"
	"gobot.io/x/gobot/v2/drivers/i2c"
	"gobot.io/x/gobot/v2/platforms/raspi"
)

var (
	defaultGain     = 4
	defaultDataRate = 860

	voltageOffset        = 0.029
	voltageToAmpereRatio = 0.020 // ADS1115 will read 20mV for every ampere current the motor draws.
)

type ADS struct {
	i2c *i2c.ADS1x15Driver
}

func New(bus, address int) (*ADS, error) {
	board := raspi.NewAdaptor()
	err := board.Connect()
	if err != nil {
		return nil, errors.Wrap(err, "connect raspi")
	}

	ads1115 := i2c.NewADS1115Driver(
		board,
		i2c.WithBus(bus),
		i2c.WithAddress(address),
		i2c.WithADS1x15Gain(defaultGain),
		i2c.WithADS1x15DataRate(defaultDataRate),
	)

	if err := ads1115.Start(); err != nil {
		return nil, errors.Wrap(err, "start ADS1115")
	}

	return &ADS{ads1115}, nil
}

func (a *ADS) Read(channel int) (float64, error) {
	v, err := a.i2c.ReadWithDefaults(channel)
	if err != nil {
		return 0, errors.Wrap(err, "read ADS1115")
	}

	return v, nil
}

func (a *ADS) Close() error {
	return a.i2c.Connection().Finalize()
}

func VoltageToAmps(v float64) float64 {
	amps := (v - voltageOffset) / voltageToAmpereRatio
	return amps
}
