package rope

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"time"

	"depthpilot/controller/cs2rs"
	"github.com/goburrow/modbus"
)

type Bus interface {
	Read(slave byte, address, count uint16) ([]uint16, error)
	Write(slave byte, address, value uint16) error
	Program(slave byte, words []uint16) error
	BroadcastStart() error
	Close() error
}

// One handler owns the physical serial port. Slave selection and each complete
// transaction are locked together, including requests made by telemetry.
type SerialBus struct {
	mu   sync.Mutex
	h    *modbus.RTUClientHandler
	c    modbus.Client
	lock *os.File
}

func OpenBus(port string) (*SerialBus, error) {
	f, err := os.OpenFile(port, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("serial port locked: %w", err)
	}
	h := modbus.NewRTUClientHandler(port)
	h.BaudRate = 38400
	h.DataBits = 8
	h.Parity = "N"
	h.StopBits = 1
	h.Timeout = 250 * time.Millisecond
	h.IdleTimeout = 0 // Keep serial configuration live for the write-only broadcast descriptor.
	if err = h.Connect(); err != nil {
		f.Close()
		return nil, err
	}
	return &SerialBus{h: h, c: modbus.NewClient(h), lock: f}, nil
}
func (b *SerialBus) Read(sl byte, addr, n uint16) ([]uint16, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.h.SlaveId = sl
	r, e := b.c.ReadHoldingRegisters(addr, n)
	if e != nil {
		return nil, e
	}
	if len(r) != int(n)*2 {
		return nil, fmt.Errorf("short register read")
	}
	v := make([]uint16, n)
	for i := range v {
		v[i] = binary.BigEndian.Uint16(r[2*i:])
	}
	return v, nil
}
func (b *SerialBus) Write(sl byte, addr, v uint16) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.h.SlaveId = sl
	_, e := b.c.WriteSingleRegister(addr, v)
	return e
}

// BroadcastStart sends one FC06 frame, never reads an acknowledgement and never
// retries. The already-open lock descriptor refers to the same configured TTY;
// all access through either descriptor is serialized by mu. The RTU handler is
// kept connected, so there is no second serial owner or configuration change.
func (b *SerialBus) BroadcastStart() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.h.SlaveId = 0
	frame, e := b.h.Encode(&modbus.ProtocolDataUnit{FunctionCode: 6, Data: []byte{0x60, 0x02, 0, 0x10}})
	if e != nil {
		return e
	}
	return sendBroadcast(b.lock, frame, time.Sleep)
}

func sendBroadcast(w io.Writer, frame []byte, pause func(time.Duration)) error {
	pause(5 * time.Millisecond) // RTU inter-frame silence before transmission.
	n, e := w.Write(frame)
	// 8 bytes take ~2.1 ms at 38400 8N1. Allow USB buffering, transmission,
	// frame silence and slave processing before the next addressed transaction.
	pause(20 * time.Millisecond)
	if e != nil {
		return e
	}
	if n != len(frame) {
		return io.ErrShortWrite
	}
	return nil
}
func (b *SerialBus) Program(sl byte, words []uint16) error {
	if len(words) != 7 {
		return fmt.Errorf("path must contain exactly seven words, excluding trigger")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.h.SlaveId = sl
	v := make([]byte, 14)
	for i, w := range words {
		binary.BigEndian.PutUint16(v[2*i:], w)
	}
	_, e := b.c.WriteMultipleRegisters(0x6200, 7, v)
	return e
}
func (b *SerialBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	e := b.h.Close()
	b.lock.Close()
	return e
}

func status(bus Bus, sl byte) (cs2rs.Status, error) {
	return sampledStatus(bus, sl, nil)
}

func sampledStatus(bus Bus, sl byte, sample *RawSample) (cs2rs.Status, error) {
	var s cs2rs.Status
	v, e := bus.Read(sl, 0x1003, 1)
	if e != nil {
		return s, e
	}
	w := v[0]
	s.Enabled = w&2 != 0
	s.Running = w&4 != 0
	s.Fault = w&1 != 0
	s.CommandComplete = w&16 != 0
	s.PathComplete = w&32 != 0
	for _, r := range []struct {
		addr uint16
		dst  *uint16
	}{{0x2203, &s.Alarm}, {0x601d, &s.PRAlarm}, {0x6002, &s.Trigger}} {
		v, e = bus.Read(sl, r.addr, 1)
		if e != nil {
			return s, e
		}
		*r.dst = v[0]
	}
	start := time.Now()
	v, e = bus.Read(sl, 0x602a, 4)
	end := time.Now()
	if e != nil {
		return s, e
	}
	s.ProfilePosition = int32(uint32(v[0])<<16 | uint32(v[1]))
	s.ActualPosition = int32(uint32(v[2])<<16 | uint32(v[3]))
	if sample != nil {
		*sample = RawSample{At: start.Add(end.Sub(start) / 2), ReadStart: start, ReadEnd: end, Profile: s.ProfilePosition, Feedback: s.ActualPosition}
	}
	return s, nil
}
