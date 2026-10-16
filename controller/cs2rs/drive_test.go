package cs2rs

import (
	"context"
	"encoding/binary"
	"errors"
	"reflect"
	"testing"
	"time"
)

type fakeBus struct {
	state   uint16
	trigger uint16
	writes  []uint16
	payload []byte
	fail    bool
	short   bool
}

func (f *fakeBus) ReadHoldingRegisters(a, n uint16) ([]byte, error) {
	if f.short {
		return nil, nil
	}
	b := make([]byte, 2*n)
	switch a {
	case 0x1003:
		binary.BigEndian.PutUint16(b, f.state)
	case 0x6002:
		binary.BigEndian.PutUint16(b, f.trigger)
	case 0x602A:
		binary.BigEndian.PutUint32(b, 0xffffffff)
		binary.BigEndian.PutUint32(b[4:], 0x80000000)
	}
	return b, nil
}
func (f *fakeBus) WriteSingleRegister(a, v uint16) ([]byte, error) {
	f.writes = append(f.writes, a, v)
	return nil, nil
}
func (f *fakeBus) WriteMultipleRegisters(a, n uint16, b []byte) ([]byte, error) {
	f.writes = append(f.writes, a, n)
	f.payload = append([]byte(nil), b...)
	if f.fail {
		return nil, errors.New("timeout")
	}
	return nil, nil
}

func TestRelativeMoveWireFormat(t *testing.T) {
	f := &fakeBus{state: 2}
	d := New(f)
	if err := d.Start(Move{-2, true, 60, 1000, 2000}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(f.writes, []uint16{0x6200, 7, 0x6002, 0x10}) {
		t.Fatal(f.writes)
	}
	want := []byte{0, 0x41, 0xff, 0xff, 0xff, 0xfe, 0, 60, 3, 0xe8, 7, 0xd0, 0, 0}
	if !reflect.DeepEqual(f.payload, want) {
		t.Fatalf("got %x want %x", f.payload, want)
	}
}
func TestRejectUnavailableDrive(t *testing.T) {
	for _, f := range []*fakeBus{{state: 0}, {state: 3}, {state: 6}, {state: 2, trigger: 0x10}, {state: 2, short: true}} {
		if err := New(f).Start(Move{RPM: 60, Acceleration: 1000, Deceleration: 1000}); err == nil {
			t.Fatal("expected rejection")
		}
		if len(f.writes) != 0 {
			t.Fatal("wrote to unavailable drive")
		}
	}
}
func TestFailedProgrammingDoesNotTrigger(t *testing.T) {
	f := &fakeBus{state: 2, fail: true}
	if err := New(f).Start(Move{RPM: 60, Acceleration: 1000, Deceleration: 1000}); err == nil {
		t.Fatal("expected error")
	}
	if len(f.writes) != 2 {
		t.Fatal(f.writes)
	}
}
func TestSignedTelemetry(t *testing.T) {
	s, err := New(&fakeBus{state: 2}).Status()
	if err != nil {
		t.Fatal(err)
	}
	if s.ProfilePosition != -1 || s.ActualPosition != -2147483648 {
		t.Fatal(s)
	}
}
func TestCanceledWaitStops(t *testing.T) {
	f := &fakeBus{state: 6}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := New(f).Wait(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(f.writes, []uint16{0x6002, 0x40}) {
		t.Fatal(f.writes)
	}
}
func TestPendingTriggerNotMistakenForCompletion(t *testing.T) {
	f := &fakeBus{state: 2 | 32, trigger: 0x10}
	ctx, cancel := context.WithTimeout(context.Background(), 125*time.Millisecond)
	defer cancel()
	if err := New(f).Wait(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("pending motion completed: %v", err)
	}
}

func TestRS485JogRegistersAndHeartbeat(t *testing.T) {
	f := &fakeBus{state: 2}
	d := New(f)
	if err := d.PrepareJog(60, 1000); err != nil {
		t.Fatal(err)
	}
	if err := d.RefreshJog(false); err != nil {
		t.Fatal(err)
	}
	want := []uint16{0x01E1, 60, 0x01E7, 1000, 0x1801, 0x4002}
	if !reflect.DeepEqual(f.writes, want) {
		t.Fatalf("got %x", f.writes)
	}
	f.writes = nil
	if _, err := d.StatusDuringJog(true); err != nil {
		t.Fatal(err)
	}
	if len(f.writes) != 10 {
		t.Fatal("missing interleaved refreshes", f.writes)
	}
	for i := 0; i < len(f.writes); i += 2 {
		if f.writes[i] != 0x1801 || f.writes[i+1] != 0x4001 {
			t.Fatal(f.writes)
		}
	}
}
