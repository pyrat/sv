package rope

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// SimBus executes the programmed trapezoids, including acceleration, using the
// same register protocol and orchestration as hardware.
type SimBus struct {
	mu     sync.Mutex
	drives [2]simDrive
}
type simDrive struct {
	regs                  map[uint16]uint16
	position              float64
	start, target         float64
	began                 time.Time
	seconds, accel, speed float64
	running, complete     bool
}

func NewSimBus() *SimBus {
	b := &SimBus{}
	for i := range b.drives {
		b.drives[i].regs = map[uint16]uint16{0x0001: 10000, 0x0233: 4000, 0x0191: 40, 0x0145: 8, 0x0157: 0x24, 0x0167: 250, 0x0169: 250}
	}
	b.drives[0].regs[0x0191] = 60
	b.drives[1].regs[0x0145] = 0x88
	return b
}
func (d *simDrive) update() {
	if !d.running {
		return
	}
	t := time.Since(d.began).Seconds()
	distance := math.Abs(d.target - d.start)
	var travel float64
	switch {
	case t >= d.seconds:
		d.position = d.target
		d.running = false
		d.complete = true
		return
	case t < d.accel:
		travel = .5 * d.speed * t * t / d.accel
	case t > d.seconds-d.accel:
		remaining := d.seconds - t
		travel = distance - .5*d.speed*remaining*remaining/d.accel
	default:
		travel = d.speed * (t - d.accel/2)
	}
	d.position = d.start + math.Copysign(travel, d.target-d.start)
}
func (b *SimBus) Read(sl byte, a, n uint16) ([]uint16, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if sl < 1 || sl > 2 {
		return nil, fmt.Errorf("invalid slave")
	}
	d := &b.drives[sl-1]
	d.update()
	out := make([]uint16, n)
	for i := range out {
		r := a + uint16(i)
		v := d.regs[r]
		switch r {
		case 0x1003:
			v = 0
			if d.regs[0xf] == 1 || d.regs[0x0145] == 0x88 {
				v |= 2
			}
			if d.running {
				v |= 4
			}
			if d.complete {
				v |= 48
			}
		case 0x6002:
			v = 0
			if d.running {
				v = 16
			}
		case 0x602a, 0x602c:
			v = uint16(uint32(int32(math.Round(d.position))) >> 16)
		case 0x602b, 0x602d:
			v = uint16(int32(math.Round(d.position)))
		}
		out[i] = v
	}
	return out, nil
}
func (b *SimBus) Write(sl byte, a, v uint16) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.write(sl, a, v)
}
func (b *SimBus) BroadcastStart() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	first := b.write(1, 0x6002, 0x10)
	second := b.write(2, 0x6002, 0x10)
	if first != nil {
		return first
	}
	if second != nil {
		return second
	}
	b.drives[1].began = b.drives[0].began
	return nil
}
func (b *SimBus) write(sl byte, a, v uint16) error {
	if sl < 1 || sl > 2 {
		return fmt.Errorf("invalid slave")
	}
	d := &b.drives[sl-1]
	d.update()
	if a == 0x6002 {
		if v == 0x40 {
			d.running = false
			d.complete = false
			return nil
		}
		if v != 16 {
			return fmt.Errorf("unsupported trigger")
		}
		if d.regs[0xf] != 1 && d.regs[0x0145] != 0x88 {
			return fmt.Errorf("disabled")
		}
		d.start = d.position
		d.target = float64(int32(uint32(d.regs[0x6201])<<16 | uint32(d.regs[0x6202])))
		d.speed = float64(d.regs[0x6203]) * PulsesPerTurn / 60
		if d.speed == 0 {
			return fmt.Errorf("zero speed")
		}
		d.accel = float64(d.regs[0x6204]) * float64(d.regs[0x6203]) / 1e6
		d.seconds = math.Abs(d.target-d.start)/d.speed + d.accel
		d.began = time.Now()
		d.running = true
		d.complete = false
		return nil
	}
	d.regs[a] = v
	return nil
}
func (b *SimBus) Program(sl byte, w []uint16) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if sl < 1 || sl > 2 || len(w) != 7 {
		return fmt.Errorf("bad path")
	}
	for i, v := range w {
		b.drives[sl-1].regs[0x6200+uint16(i)] = v
	}
	return nil
}
func (b *SimBus) Close() error { return nil }
