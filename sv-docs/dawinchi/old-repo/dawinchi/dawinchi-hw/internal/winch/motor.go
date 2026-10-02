package winch

func (f *Firmware) SetMotorRightSpeed(speed uint32) error {
	return f.motorRight.SetSpeed(speed)
}

func (f *Firmware) SetMotorLeftSpeed(speed uint32) error {
	return f.motorLeft.SetSpeed(speed)
}
