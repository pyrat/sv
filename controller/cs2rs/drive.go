// Package cs2rs controls the Leadshine CS2RS-D507 using the V1.5 register map.
package cs2rs

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

// Registers is implemented by a Modbus client. Addresses are zero-based wire addresses.
type Registers interface {
	ReadHoldingRegisters(address, quantity uint16) ([]byte, error)
	WriteSingleRegister(address, value uint16) ([]byte, error)
	WriteMultipleRegisters(address, quantity uint16, value []byte) ([]byte, error)
}

type Drive struct {
	bus Registers
	mu  sync.Mutex
}

func New(bus Registers) *Drive { return &Drive{bus: bus} }

type Status struct {
	Fault           bool   `json:"fault"`
	Enabled         bool   `json:"enabled"`
	Running         bool   `json:"running"`
	CommandComplete bool   `json:"command_complete"`
	PathComplete    bool   `json:"path_complete"`
	Homed           bool   `json:"homed"`
	Alarm           uint16 `json:"alarm"`
	PRAlarm         uint16 `json:"pr_alarm"`
	Trigger         uint16 `json:"trigger"`
	ProfilePosition int32  `json:"profile_position"`
	ActualPosition  int32  `json:"actual_position"`
}

func (d *Drive) read(address, count uint16) ([]byte, error) {
	b, err := d.bus.ReadHoldingRegisters(address, count)
	if err != nil {
		return nil, fmt.Errorf("read 0x%04x: %w", address, err)
	}
	if len(b) != int(count)*2 {
		return nil, fmt.Errorf("read 0x%04x: expected %d bytes, got %d", address, count*2, len(b))
	}
	return b, nil
}

func (d *Drive) status() (Status, error) {
	return d.statusWithJog(nil)
}

// StatusDuringJog interleaves a jog refresh before each telemetry transaction,
// rather than letting a multi-register status poll starve the <50ms jog stream.
func (d *Drive) StatusDuringJog(positive bool) (Status, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.statusWithJog(&positive)
}

func (d *Drive) statusWithJog(positive *bool) (Status, error) {
	return d.statusWithHeartbeat(func() error {
		if positive != nil {
			v := uint16(0x4002)
			if *positive {
				v = 0x4001
			}
			if _, err := d.bus.WriteSingleRegister(0x1801, v); err != nil {
				return err
			}
		}
		return nil
	})
}

// StatusWithHeartbeat services another drive's jog between status transactions.
// The callback must not call methods on this same Drive (whose lock is held).
func (d *Drive) StatusWithHeartbeat(heartbeat func() error) (Status, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.statusWithHeartbeat(heartbeat)
}

func (d *Drive) statusWithHeartbeat(heartbeat func() error) (Status, error) {
	read := func(address, count uint16) ([]byte, error) {
		if heartbeat != nil {
			if err := heartbeat(); err != nil {
				return nil, err
			}
		}
		return d.read(address, count)
	}
	var s Status
	b, err := read(0x1003, 1)
	if err != nil {
		return s, err
	}
	w := binary.BigEndian.Uint16(b)
	s.Fault, s.Enabled, s.Running = w&1 != 0, w&2 != 0, w&4 != 0
	s.CommandComplete, s.PathComplete, s.Homed = w&16 != 0, w&32 != 0, w&64 != 0
	for _, r := range []struct {
		address uint16
		value   *uint16
	}{{0x2203, &s.Alarm}, {0x601D, &s.PRAlarm}, {0x6002, &s.Trigger}} {
		b, err = read(r.address, 1)
		if err != nil {
			return s, err
		}
		*r.value = binary.BigEndian.Uint16(b)
	}
	b, err = read(0x602A, 4)
	if err != nil {
		return s, err
	}
	s.ProfilePosition = int32(binary.BigEndian.Uint32(b[:4]))
	s.ActualPosition = int32(binary.BigEndian.Uint32(b[4:]))
	return s, nil
}

func (d *Drive) Status() (Status, error) { d.mu.Lock(); defer d.mu.Unlock(); return d.status() }

// Stop requests the drive's quick-stop ramp (Pr8.23). It does not disable torque.
func (d *Drive) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.bus.WriteSingleRegister(0x6002, 0x40)
	return err
}

type Move struct {
	Position     int32 // command pulses, signed; not encoder counts or metres
	Relative     bool
	RPM          uint16
	Acceleration uint16 // ms/1000 rpm
	Deceleration uint16 // ms/1000 rpm
}

// PrepareJog programs RS485 jog velocity and ramp without starting motion.
// V1.5 section 5.3 uses Pr6.00/Pr6.03 (not the external-I/O jog registers).
func (d *Drive) PrepareJog(rpm, ramp uint16) error {
	if rpm == 0 || rpm > 5000 || ramp == 0 {
		return fmt.Errorf("invalid jog profile")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	s, err := d.status()
	if err != nil {
		return err
	}
	if !s.Enabled || s.Running || s.Fault || s.Alarm != 0 || s.PRAlarm != 0 || s.Trigger > 15 {
		return fmt.Errorf("drive is not ready for jog")
	}
	if _, err = d.bus.WriteSingleRegister(0x01E1, rpm); err != nil {
		return err
	}
	_, err = d.bus.WriteSingleRegister(0x01E7, ramp)
	return err
}

// RefreshJog must be called at intervals below 50ms for continuous RS485 jog.
// Expiry is also followed by explicit Stop; the manual does not specify the
// stopping distance/time of a missing refresh.
func (d *Drive) RefreshJog(positive bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	v := uint16(0x4002)
	if positive {
		v = 0x4001
	}
	_, err := d.bus.WriteSingleRegister(0x1801, v)
	return err
}

// Start programs volatile PR path 0, then triggers it exactly once. A transport
// error on the trigger leaves the outcome unknown; callers must not retry a move.
func (d *Drive) Start(m Move) error {
	if m.RPM == 0 || m.RPM > 5000 || m.Acceleration == 0 || m.Deceleration == 0 {
		return fmt.Errorf("require RPM 1..5000 and nonzero acceleration/deceleration")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	s, err := d.status()
	if err != nil {
		return err
	}
	if s.Fault || s.Alarm != 0 || s.PRAlarm != 0 {
		return fmt.Errorf("drive fault: alarm=0x%04x PR=0x%04x", s.Alarm, s.PRAlarm)
	}
	if !s.Enabled || s.Running || s.Trigger > 15 {
		return fmt.Errorf("drive is not enabled and idle")
	}
	mode := uint16(1)
	if m.Relative {
		mode |= 1 << 6
	}
	p := uint32(m.Position)
	words := []uint16{mode, uint16(p >> 16), uint16(p), m.RPM, m.Acceleration, m.Deceleration, 0}
	b := make([]byte, len(words)*2)
	for i, w := range words {
		binary.BigEndian.PutUint16(b[i*2:], w)
	}
	// Do not write 0x6207: it maps directly to the trigger register.
	if _, err = d.bus.WriteMultipleRegisters(0x6200, 7, b); err != nil {
		return fmt.Errorf("program path: %w", err)
	}
	if _, err = d.bus.WriteSingleRegister(0x6002, 0x10); err != nil {
		return fmt.Errorf("trigger outcome unknown: %w", err)
	}
	return nil
}

// Wait waits for path 0 completion, including acceptance of the trigger. A
// canceled context, fault or lost communications causes one best-effort stop.
func (d *Drive) Wait(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			if stopErr := d.Stop(); stopErr != nil {
				err = fmt.Errorf("%w; stop failed: %v", err, stopErr)
			}
		}
	}()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
		}
		s, e := d.Status()
		if e != nil {
			return e
		}
		if s.Fault || s.Alarm != 0 || s.PRAlarm != 0 || !s.Enabled {
			return fmt.Errorf("motion interrupted: %+v", s)
		}
		if s.Trigger == 0 && !s.Running && s.PathComplete {
			return nil
		}
	}
}
