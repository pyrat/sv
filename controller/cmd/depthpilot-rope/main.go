package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"depthpilot/controller/internal/rope"
)

func main() {
	addr := flag.String("listen", "127.0.0.1:8080", "HTTP address; use :8080 for trusted LAN access")
	sim := flag.Bool("simulate", true, "simulate both drives")
	broadcast := flag.Bool("broadcast-start", false, "experimental common broadcast start when both drums move")
	port := flag.String("port", "/dev/serial/by-id/usb-1a86_USB_Serial-if00-port0", "shared 38400 baud 8N1 RS485 adapter")
	flag.Parse()
	var bus rope.Bus
	if *sim {
		bus = rope.NewSimBus()
	} else {
		b, e := rope.OpenBus(*port)
		if e != nil {
			log.Fatal(e)
		}
		bus = b
	}
	c, e := rope.NewWithStartMode(bus, *sim, *broadcast)
	if e != nil {
		bus.Close()
		log.Fatal(e)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	done := make(chan struct{})
	go func() { c.Run(ctx); close(done) }()
	h := &http.Server{Addr: *addr, Handler: c.Handler(), ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second}
	httpDone := make(chan struct{})
	go func() {
		defer close(httpDone)
		<-ctx.Done()
		s, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		h.Shutdown(s)
	}()
	log.Printf("Rope commissioning at http://%s simulation=%t; no startup movement", *addr, *sim)
	if e = h.ListenAndServe(); e != nil && e != http.ErrServerClosed {
		log.Print(e)
	}
	cancel()
	<-done
	<-httpDone
	if e = c.Close(); e != nil {
		log.Print("Shutdown: ", e)
	}
}
