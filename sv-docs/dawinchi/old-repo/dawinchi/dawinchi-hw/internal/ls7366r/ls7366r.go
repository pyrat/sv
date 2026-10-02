// Package ls7366r implements a driver for the LS7366R quadrature encoder over SPI interface.
//
// https://lsicsi.com/datasheets/LS7366R.pdf
package ls7366r

import (
	"sync"

	"github.com/stianeikeland/go-rpio/v4"
)

// TODO: implement
// counter value from disk - for reboot

const (
	// MDR0 configuration data
	//
	// Count modes
	NQuad   = 0x00 // Non-quadrature mode
	QuadRx1 = 0x01 // X1 quadrature mode
	QuadRx2 = 0x02 // X2 quadrature mode
	QuadRx4 = 0x03 // X4 quadrature mode

	// Running modes
	FreeRun     = 0x00
	SingleCycle = 0x04
	RangeLimit  = 0x08
	ModuloN     = 0x0C

	// Index modes
	DisableIndx   = 0x00 // Index disabled
	IndxLoadCNTR  = 0x10 // Index load CNTR
	IndxResetCNTR = 0x20 // Index reset CNTR
	IndxLoadOTR   = 0x30 // Index load OTR
	AsynchIndx    = 0x00 // Asynchronous index
	SynchIndx     = 0x80 // Synchronous index

	// Clock filter modes
	Filter1 = 0x00 // Filter clock frequency division factor 1
	Filter2 = 0x80 // Filter clock frequency division factor 2

	// MDR1 configuration data
	//
	// 1 to 4 bytes data-width
	ModeByte4 = 0x00 // 4-byte counter mode
	ModeByte3 = 0x01 // 3-byte counter mode
	ModeByte2 = 0x02 // 2-byte counter mode
	ModeByte1 = 0x03 // 1-byte counter mode

	// Enable/disable counter
	EnableCntr  = 0x00 // Counting enabled
	DisableCntr = 0x04 // Counting disabled

	//Flags
	NoFlags = 0x00 // All flags disabled
	IdxFlag = 0x10 // IDX flag
	CmpFlag = 0x20 // CMP flag
	BwFlag  = 0x40 // BW flag
	CyFlag  = 0x80 // CY flag

	// LS7366R op-code list
	ClrMdr0   = 0x08
	ClrMdr1   = 0x10
	ClrCntr   = 0x20
	ClrStr    = 0x30
	ReadMdr0  = 0x48
	ReadMdr1  = 0x50
	ReadCntr  = 0x60
	ReadOtr   = 0x68
	ReadStr   = 0x70
	WriteMdr0 = 0x88
	WriteMdr1 = 0x90
	WriteDtr  = 0x98
	LoadCntr  = 0xE0
	LoadOtr   = 0xE4
)

const defaultFreq = 1000000 // 1MHz
const defaultDataWidth = 5  // cmd + 4 bytes (synced with motor HS30B Optical Encoder)

type Counter struct {
	mu sync.Mutex
}

// New returns a new LS7366R driver with the given chip and frequency.
// rpio.SpiBegin(SPI0) must be called before calling this function.
// Motor right and left share the same SPI pins (MOSI, MISO, SCLK).
// Motor right uses CE0, motor left uses CE1.
func New() *Counter {
	rpio.SpiSpeed(defaultFreq)

	c := &Counter{}

	// default mode
	// as per HS30B Optical Kit Encode - A, B Channel Quadrature TTL Outputs with Index
	c.SetCountMode(0, QuadRx4|FreeRun|IndxLoadOTR, ModeByte4|EnableCntr|BwFlag|CyFlag)
	c.SetCountMode(1, QuadRx4|FreeRun|IndxLoadOTR, ModeByte4|EnableCntr|BwFlag|CyFlag)

	return c
}

// SetCountMode sets up the operating mode by writing to MDR0 and MDR1 registers.
// Data width is ReadCntr data width.
func (c *Counter) SetCountMode(chip uint8, mdr0 byte, mdr1 byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	rpio.SpiChipSelect(chip)

	// use IR register example
	// select MDR0 register
	// B7 B6 B5 B4 B3 B2 B1 B0
	// 1  0  0  0  1  0  0  0

	// set MDR0 that sets up the operating mode
	// (upon power up MDR0 is cleared to zero)
	rpio.SpiTransmit(WriteMdr0, mdr0)

	// set MDR1 that appended to MDR0 for additional modes
	// (upon power-up MDR1 is cleared to zero)
	rpio.SpiTransmit(WriteMdr1, mdr1)
}

func (c *Counter) ReadCntr(chip uint8) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	rpio.SpiChipSelect(chip)

	b := make([]byte, defaultDataWidth)
	b[0] = ReadCntr // command
	rpio.SpiExchange(b)

	var counter int64
	for _, bc := range b[1:] {
		counter = (counter << 8) | int64(bc)
	}

	return counter
}

// ReadOtr reads the OTR register, that serves as a convenient dump site for instantaneous CNTR data,
// which can then be read without interfering with the counting process.
func (c *Counter) ReadOtr(chip uint8) uint32 {
	c.mu.Lock()
	defer c.mu.Unlock()

	rpio.SpiChipSelect(chip)

	b := make([]byte, defaultDataWidth)
	b[0] = ReadOtr // command
	rpio.SpiExchange(b)

	var counter uint32
	for _, bc := range b[1:] {
		counter = (counter << 8) | uint32(bc)
	}

	return counter
}

// ReadStr reads the STR, an 8-bit status register which stores count related status information.
//
// CY: Carry (CNTR overflow) latch
// BW: Borrow (CNTR underflow) latch
// CMP: Compare (CNTR = DTR) latch
// IDX: Index latch
// CEN: Count enable status: 0: counting disabled, 1: counting enabled
// PLS: Power loss indicator latch; set upon power up
// U/D: Count direction indicator: 0: count down, 1: count up
// S: Sign bit. 1: negative, 0: positive
//
// Example:
// CY BW CMP IDX CEN PLS U/D S
// 1  0  1   0   1   0   0   0
func (c *Counter) ReadStr(chip uint8) byte {
	c.mu.Lock()
	defer c.mu.Unlock()

	rpio.SpiChipSelect(chip)

	b := make([]byte, 2)
	b[0] = ReadStr // command
	rpio.SpiExchange(b)

	return b[1]
}

func (c *Counter) ResetCntr(chip uint8) {
	c.mu.Lock()
	defer c.mu.Unlock()

	rpio.SpiChipSelect(chip)
	rpio.SpiTransmit(ClrCntr)
}

func (c *Counter) ResetStr(chip uint8) {
	c.mu.Lock()
	defer c.mu.Unlock()

	rpio.SpiChipSelect(chip)
	rpio.SpiTransmit(ClrStr)
}
