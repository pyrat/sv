package motor

type Counter interface {
	ReadCntr(chip uint8) int64
	ReadOtr(chip uint8) uint32
	ReadStr(chip uint8) byte
	ResetCntr(chip uint8)
	ResetStr(chip uint8)
}

// ReadCounter reads the counter from the counter chip associated with the motor.
// Call it in a loop to keep track of the number of rotations.
func (m *Motor) ReadCounter() uint32 {
	return m.counter.ReadOtr(m.counterChip)
	// return m.counter.ReadCntr(m.counterChip)
}

func (m *Motor) ResetCounter() {
	m.counter.ResetCntr(m.counterChip)
	m.counter.ResetStr(m.counterChip)
}
