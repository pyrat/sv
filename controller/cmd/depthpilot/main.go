package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"depthpilot/controller/cs2rs"
	"github.com/goburrow/modbus"
)

func run() error {
	port := flag.String("port", "", "USB RS485 serial device (prefer /dev/serial/by-id/...)")
	id := flag.Int("id", 1, "Modbus address 1..127")
	baud := flag.Int("baud", 38400, "baud rate matching drive DIP switches")
	position := flag.Int64("position", 0, "signed command pulses")
	relative := flag.Bool("relative", false, "relative rather than absolute move")
	rpm := flag.Uint("rpm", 60, "motor rpm, 1..5000")
	ramp := flag.Uint("ramp", 1000, "acceleration/deceleration in ms/1000 rpm")
	timeout := flag.Duration("timeout", 30*time.Second, "motion deadline")
	flag.Parse()
	if flag.NArg() != 1 || *port == "" || *id < 1 || *id > 127 {
		return fmt.Errorf("usage: depthpilot -port DEVICE [flags] status|move|stop")
	}
	command := flag.Arg(0)
	if command != "status" && command != "move" && command != "stop" {
		return fmt.Errorf("unknown command %q", command)
	}
	if *position < -2147483648 || *position > 2147483647 || *rpm < 1 || *rpm > 5000 || *ramp < 1 || *ramp > 65535 || *timeout <= 0 {
		return fmt.Errorf("invalid motion parameters")
	}
	h := modbus.NewRTUClientHandler(*port)
	h.BaudRate, h.DataBits, h.Parity, h.StopBits = *baud, 8, "N", 1
	h.SlaveId, h.Timeout = byte(*id), time.Second
	if err := h.Connect(); err != nil {
		return err
	}
	defer h.Close()
	d := cs2rs.New(modbus.NewClient(h))
	switch command {
	case "status":
		s, err := d.Status()
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(s)
	case "stop":
		return d.Stop()
	default:
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		ctx, deadline := context.WithTimeout(ctx, *timeout)
		defer deadline()
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := d.Start(cs2rs.Move{Position: int32(*position), Relative: *relative, RPM: uint16(*rpm), Acceleration: uint16(*ramp), Deceleration: uint16(*ramp)}); err != nil {
			if stopErr := d.Stop(); stopErr != nil {
				return fmt.Errorf("%w; stop failed: %v", err, stopErr)
			}
			return err
		}
		return d.Wait(ctx)
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
