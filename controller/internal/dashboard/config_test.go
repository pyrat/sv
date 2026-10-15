package dashboard

import (
	"strings"
	"testing"
)

func TestDriverSelection(t *testing.T) {
	for _, model := range []string{"", "cs2rs-d507", "cl57pr"} {
		c := DemoConfig()["x"]
		c.Driver = model
		if err := c.validate(); err != nil {
			t.Fatalf("model %q: %v", model, err)
		}
	}
	c := DemoConfig()["x"]
	c.Driver = "cl57rs"
	if err := c.validate(); err == nil {
		t.Fatal("unknown driver accepted")
	}
}

func TestUnimplementedDriverRejectedBeforeSerialConnection(t *testing.T) {
	for _, model := range []string{"cl57pr", "unknown"} {
		c := DemoConfig()["x"]
		c.Driver, c.Port, c.Slave, c.Baud = model, "/nonexistent/serial", 1, 38400
		m, err := newHardware(c)
		if m != nil || err == nil {
			t.Fatalf("model %q: motor=%v error=%v", model, m, err)
		}
		want := "unknown driver"
		if model == "cl57pr" {
			want = "confirmed CL57PR register map"
		}
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("model %q: expected model error before serial open, got %v", model, err)
		}
	}
}
