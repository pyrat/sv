package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CreateView/dawinchi-hw/internal/ads1115"
	"github.com/CreateView/dawinchi-hw/internal/ls7366r"
	"github.com/CreateView/dawinchi-hw/internal/motor"
	"github.com/CreateView/dawinchi-hw/internal/panel"
	"github.com/CreateView/dawinchi-hw/internal/rpc"
	"github.com/CreateView/dawinchi-hw/internal/winch"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stianeikeland/go-rpio/v4"
)

const (
	env            = "development"
	APIPort        = "9001"
	APIBearerToken = "xxx"
)

var (
	adsBus     = 1
	adsAddress = 0x48

	defaultMotorsSpeed uint32 = 5

	rightMotorCounterChip       uint8 = 1
	rightMotorPin                     = 12
	rightMotorDirectionPin            = 2
	rightMotorBreakPin                = 18
	rightMotorDriverStatusPin         = 20
	rightMotorADSMUX                  = 0
	rightMotorDrumSize                = 0
	rightMotorDrumFullSpinCount       = 22000
	rightMotorRopeSize                = 0

	leftMotorCounterChip       uint8 = 0
	leftMotorPin                     = 13
	leftMotorDirectionPin            = 3
	leftMotorBreakPin                = 19
	leftMotorDriverStatusPin         = 21
	leftMotorADSMUX                  = 1
	leftMotorDrumSize                = 0
	leftMotorDrumFullSpinCount       = 22000
	leftMotorRopeSize                = 0
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatal().Err(err).Msg("main run")
	}
}

func run(ctx context.Context) error {
	// log to be used globally
	// pass as dependency to other packages
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// if env == "development" {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	// }

	// initializes the Raspberry Pi GPIO and SPI,
	// it must be called before any other packages,
	// close must be called when the program exits to release resources.
	if err := rpio.Open(); err != nil {
		return errors.Wrap(err, "rpio open")
	}
	defer rpio.Close()

	if err := rpio.SpiBegin(rpio.Spi0); err != nil {
		return errors.Wrap(err, "rpio spi begin")
	}
	defer rpio.SpiEnd(rpio.Spi0)

	// ads1115 shared by both motors
	ads, err := ads1115.New(adsBus, adsAddress)
	if err != nil {
		return errors.Wrap(err, "init ADS1115")
	}
	defer ads.Close()

	// ls7366r shared by both motors
	counter := ls7366r.New()

	motorRight := motor.New(
		rightMotorPin,
		rightMotorDirectionPin,
		rightMotorBreakPin,
		rightMotorDriverStatusPin,
		counter,
		rightMotorCounterChip,
		ads,
		rightMotorADSMUX,
	)

	motorLeft := motor.New(
		leftMotorPin,
		leftMotorDirectionPin,
		leftMotorBreakPin,
		leftMotorDriverStatusPin,
		counter,
		leftMotorCounterChip,
		ads,
		leftMotorADSMUX,
	)

	butonsPanel := panel.New(ctx)
	defer butonsPanel.Close()

	defer func() {
		if r := recover(); r != nil {
			log.Error().Msgf("recovered from panic: %v", r)
			butonsPanel.ShowError()
		}
	}()

	// winch init
	winch, err := winch.New(ctx, motorRight, motorLeft, butonsPanel, log)
	if err != nil {
		return errors.Wrap(err, "init winch")
	}
	defer winch.Close()

	// counter 0
	winch.Reset()

	// go func() {
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			return
	// 		default:
	// 			time.Sleep(3000 * time.Millisecond)
	// 			state := winch.State()

	// 			sj, err := json.Marshal(state)
	// 			if err != nil {
	// 				fmt.Println("json marshal error:", err)
	// 			}
	// 			fmt.Println("winch state:", string(sj))
	// 		}
	// 	}
	// }()

	// time.Sleep(3 * time.Second)

	// winch.Move(ctx, 20, 20, 10)

	// fmt.Println("winch move: done")
	// time.Sleep(3 * time.Second)

	// winch.Stop()

	// websocket jsonrpc2
	rpcAPI, err := rpc.New(APIBearerToken, winch, log)

	bind := fmt.Sprintf(":%s", APIPort)
	log.Debug().Msgf("starting API server %s", bind)

	srv := &http.Server{
		Addr:              bind,
		Handler:           rpcAPI,
		ReadHeaderTimeout: time.Second * 20,
		WriteTimeout:      time.Second * 180,
		ReadTimeout:       time.Second * 120,
		IdleTimeout:       time.Second * 120,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server listen")
		}
	}()

	<-ctx.Done()

	return nil
}
