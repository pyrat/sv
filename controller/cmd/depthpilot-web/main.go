package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"depthpilot/controller/internal/dashboard"
)

func run() error {
	addr := flag.String("listen", "127.0.0.1:8080", "HTTP listen address; use :8080 for LAN")
	sim := flag.Bool("simulate", true, "use two simulated axes (no serial devices opened)")
	config := flag.String("config", "", "hardware axis calibration JSON")
	web := flag.String("web", "../web/dist", "built web UI directory")
	state := flag.String("state", "", "home state JSON path; '-' for ephemeral state")
	activity := flag.String("activity", "", "activity SQLite path; default alongside home state; '-' for memory only")
	activityLimit := flag.Int("activity-limit", 10000, "maximum retained activity events")
	flag.Parse()
	configs := dashboard.DemoConfig()
	if !*sim {
		if *config == "" {
			log.Fatal("hardware mode requires -config")
		}
		b, err := os.ReadFile(*config)
		if err != nil {
			log.Fatal(err)
		}
		configs = nil
		if err = json.Unmarshal(b, &configs); err != nil {
			log.Fatal(err)
		}
	}
	s, err := dashboard.New(*sim, configs)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	if *state == "" {
		if *sim {
			*state = ".depthpilot-sim-home.json"
		} else {
			*state = ".depthpilot-home.json"
		}
	}
	if err := s.LoadHome(*state); err != nil {
		return err
	}
	if *activity == "" {
		if *state == "-" {
			*activity = "-"
		} else {
			*activity = strings.TrimSuffix(*state, filepath.Ext(*state)) + ".activity.sqlite"
		}
	}
	if err := s.OpenActivity(*activity, *activityLimit); err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	m := http.NewServeMux()
	m.Handle("/api/", s.Handler())
	m.Handle("/", http.FileServer(http.Dir(*web)))
	h := &http.Server{Addr: *addr, Handler: m, ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second}
	done := make(chan struct{})
	httpDone := make(chan struct{})
	go func() { s.Run(ctx); close(done) }()
	go func() {
		defer close(httpDone)
		<-ctx.Done()
		shutdown, stop := context.WithTimeout(context.Background(), 3*time.Second)
		defer stop()
		h.Shutdown(shutdown)
	}()
	log.Printf("DepthPilot http://%s · simulation=%t", *addr, *sim)
	if err = h.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Print(err)
		cancel()
	}
	cancel()
	<-done
	<-httpDone
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
