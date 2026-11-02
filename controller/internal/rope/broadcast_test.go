package rope

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/goburrow/modbus"
)

func TestBroadcastFrameAndNoReplyRequired(t *testing.T) {
	f, e := os.Create(filepath.Join(t.TempDir(), "wire"))
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	b := &SerialBus{h: modbus.NewRTUClientHandler("unused"), lock: f}
	if e = b.BroadcastStart(); e != nil {
		t.Fatal(e)
	}
	frame, e := os.ReadFile(f.Name())
	if e != nil {
		t.Fatal(e)
	}
	if len(frame) != 8 || !bytes.Equal(frame[:6], []byte{0, 6, 0x60, 2, 0, 0x10}) {
		t.Fatalf("bad frame %x", frame)
	}
	crc := uint16(0xffff)
	for _, v := range frame {
		crc ^= uint16(v)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xa001
			} else {
				crc >>= 1
			}
		}
	}
	if crc != 0 {
		t.Fatalf("bad CRC %x", frame)
	}
}

type shortBroadcastWriter struct{ calls int }

func (w *shortBroadcastWriter) Write(p []byte) (int, error) { w.calls++; return len(p) - 1, nil }
func TestBroadcastShortWriteNeverRetried(t *testing.T) {
	w := &shortBroadcastWriter{}
	var pauses []time.Duration
	e := sendBroadcast(w, make([]byte, 8), func(d time.Duration) { pauses = append(pauses, d) })
	if e != io.ErrShortWrite || w.calls != 1 || len(pauses) != 2 || pauses[1] < 20*time.Millisecond {
		t.Fatalf("error=%v writes=%d pauses=%v", e, w.calls, pauses)
	}
}
