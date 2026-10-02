package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/CreateView/xzmove/internal/hwlog"
	"github.com/CreateView/xzmove/internal/rpc"
	"github.com/pkg/errors"
	"github.com/stianeikeland/go-rpio/v4"
)

// https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/

// Version is set during build.
var Version = "dev"

func main() {
	ctx := context.Background()

	if err := run(ctx, os.Stdout, os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, w io.Writer, getenv func(string) string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// env vars
	env := getenv("HW_ENV")
	if env == "" {
		env = "debug"
	}

	apiPort := getenv("HW_API_PORT")
	if apiPort == "" {
		apiPort = "9001"
	}

	apiToken := getenv("HW_API_BEARER_TOKEN")
	if apiToken == "" {
		return fmt.Errorf("HW_API_BEARER_TOKEN is required")
	}

	hwLogFilePath := getenv("HW_LOG_FILE_PATH")
	if hwLogFilePath == "" {
		hwLogFilePath = "/var/log/winch.log"
	}

	// log to be used globally
	// pass as dependency to other packages
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if env == "debug" {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	// init hardware logs
	hl, err := hwlog.New(ctx, hwLogFilePath, logger)
	if err != nil {
		return errors.Wrap(err, "init hwlog")
	}

	// initializes the Raspberry Pi GPIO and SPI,
	// it must be called before any other packages,
	// close must be called when the program exits to release resources.
	if err := rpio.Open(); err != nil {
		return errors.Wrap(err, "rpio open")
	}
	defer rpio.Close()

	// init firmware
	err = firmware.Init(ctx, logger)
	if err != nil {
		return errors.Wrap(err, "init firmware")
	}

	// websocket jsonrpc2
	rpcAPI, err := rpc.New(apiToken, winch, hl, logger)
	if err != nil {
		return errors.Wrap(err, "init rpc")
	}

	bind := fmt.Sprintf(":%s", apiPort)
	logger.Info("starting API server:", bind)

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
			logger.Error("server listen:", err)
		}
	}()

	<-ctx.Done()

	return nil
}

// ----------------------------------

// func main() {
// 	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
//
// 	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
// 	defer stop()
//
// 	if err := run(ctx); err != nil {
// 		log.Fatal().Err(err).Msg("main run")
// 	}
// }
//
// func run(ctx context.Context) error {
// 	cfg := &Config{}
// 	err := config.Load(
// 		cfg,
// 		validation.Field(
// 			&cfg.Env,
// 			validation.Required,
// 			validation.In(config.ModeDebug, config.ModeTest, config.ModeRelease).Error(
// 				fmt.Sprintf(
// 					"'%s' is not a valid value. Valid values: %s, %s, %s", cfg.Env, config.ModeDebug, config.ModeTest, config.ModeRelease),
// 			),
// 		),
// 		validation.Field(&cfg.APIPort, validation.Required),
// 		validation.Field(&cfg.APIBearerToken, validation.Required),
// 	)
// 	if err != nil {
// 		return errors.Wrap(err, "config load")
// 	}
//
// 	// log to be used globally
// 	// pass as dependency to other packages
// 	log := config.NewLogger(cfg.Env)
//
// 	// initializes the Raspberry Pi GPIO and SPI,
// 	// it must be called before any other packages,
// 	// close must be called when the program exits to release resources.
// 	if err := rpio.Open(); err != nil {
// 		return errors.Wrap(err, "rpio open")
// 	}
// 	defer rpio.Close()
//
// 	if err := rpio.SpiBegin(rpio.Spi0); err != nil {
// 		return errors.Wrap(err, "rpio spi begin")
// 	}
// 	defer rpio.SpiEnd(rpio.Spi0)
//
// 	// init buttons panel
// 	butonsPanel := panel.New(ctx, log)
// 	defer butonsPanel.Close()
//
// 	defer func() {
// 		if r := recover(); r != nil {
// 			butonsPanel.ShowError()
// 			log.Error().Msgf("recover panic: %+v", r)
// 		}
// 	}()
//
// 	// signal boot up
// 	log.Debug().Msg("signal boot up")
// 	butonsPanel.SignalBootup()
//
// 	// init hardware logs
// 	hl, err := hwlog.New(ctx, defaultHWLogFilePath, log)
// 	if err != nil {
// 		butonsPanel.ShowError()
//
// 		log.Warn().Err(err).Msg("init hwlog: starting in memory")
//
// 		hl, err = hwlog.New(ctx, ":memory:", log)
// 		if err != nil {
// 			return errors.Wrap(err, "init hwlog: in memory")
// 		}
// 	}
//
// 	// ads1115 shared by both motors
// 	ads, err := ads1115.New(adsBus, adsAddress)
// 	if err != nil {
// 		return errors.Wrap(err, "init ADS1115")
// 	}
// 	defer ads.Close()
//
// 	// validate ADS is working
// 	_, err = ads.Read(upDownMotorADSMUX)
// 	if err != nil {
// 		return errors.Wrap(err, "ADS up/down failed")
// 	}
//
// 	_, err = ads.Read(inOutMotorADSMUX)
// 	if err != nil {
// 		return errors.Wrap(err, "ADS in/out failed")
// 	}
//
// 	// ls7366r shared by both motors
// 	counter := ls7366r.New()
//
// 	// reset counters
// 	// counter.ResetCntr(upDownMotorCounterChip)
// 	// counter.ResetStr(upDownMotorCounterChip)
//
// 	// counter.ResetCntr(inOutMotorCounterChip)
// 	// counter.ResetStr(inOutMotorCounterChip)
//
// 	upDownMotorParams := &motor.NewParams{
// 		Name:            "Up/Down motor",
// 		MotorPin:        upDownMotorPin,
// 		DirectionPin:    upDownMotorDirectionPin,
// 		BrakePin:        upDownMotorBrakePin,
// 		DriverStatusPin: upDownMotorDriverStatusPin,
// 		Counter:         counter,
// 		CounterChip:     upDownMotorCounterChip,
// 		ADS:             ads,
// 		ADSMUX:          upDownMotorADSMUX,
// 		HWLog:           hl,
// 		Log:             log,
// 	}
// 	upDownMotor, err := motor.New(upDownMotorParams)
// 	if err != nil {
// 		return errors.Wrap(err, "init up/down motor")
// 	}
// 	defer upDownMotor.Stop()
//
// 	inOutMotorParams := &motor.NewParams{
// 		Name:            "In/Out motor",
// 		MotorPin:        inOutMotorPin,
// 		DirectionPin:    inOutMotorDirectionPin,
// 		BrakePin:        inOutMotorBrakePin,
// 		DriverStatusPin: inOutMotorDriverStatusPin,
// 		Counter:         counter,
// 		CounterChip:     inOutMotorCounterChip,
// 		ADS:             ads,
// 		ADSMUX:          inOutMotorADSMUX,
// 		HWLog:           hl,
// 		Log:             log,
// 	}
// 	inOutMotor, err := motor.New(inOutMotorParams)
// 	if err != nil {
// 		return errors.Wrap(err, "init in/out motor")
// 	}
// 	defer inOutMotor.Stop()
//
// 	// winch init
// 	winch, err := winch.Init(ctx, upDownMotor, inOutMotor, butonsPanel, hl, log)
// 	if err != nil {
// 		return errors.Wrap(err, "init winch")
// 	}
// 	defer winch.Close()
//
// 	defer func() {
// 		err := winch.SaveState()
// 		if err != nil {
// 			log.Error().Err(err).Msg("save state")
// 		}
// 	}()
//
// 	// websocket jsonrpc2
// 	rpcAPI, err := rpc.New(cfg.APIBearerToken, winch, hl, log)
// 	if err != nil {
// 		return errors.Wrap(err, "init rpc")
// 	}
//
// 	bind := fmt.Sprintf(":%s", cfg.APIPort)
// 	log.Info().Msgf("starting API server %s", bind)
//
// 	srv := &http.Server{
// 		Addr:              bind,
// 		Handler:           rpcAPI,
// 		ReadHeaderTimeout: time.Second * 20,
// 		WriteTimeout:      time.Second * 180,
// 		ReadTimeout:       time.Second * 120,
// 		IdleTimeout:       time.Second * 120,
// 	}
//
// 	go func() {
// 		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 			log.Fatal().Err(err).Msg("server listen")
// 		}
// 	}()
//
// 	<-ctx.Done()
//
// 	return nil
// }
