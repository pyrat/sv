package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"time"

	"github.com/CreateView/ropey/internal/motordriver"
	"github.com/CreateView/ropey/internal/rotaryencoder"
	"github.com/stianeikeland/go-rpio/v4"
)

func main() {
	ctx := context.Background()

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// initializes the Raspberry Pi GPIO and SPI,
	// it must be called before any other packages,
	// close must be called when the program exits to release resources.
	if err := rpio.Open(); err != nil {
		log.Fatalf("init rpio: %v", err)
	}
	defer rpio.Close()

	encoder1 := rotaryencoder.New(23, 24)
	encoder1.Start(ctx)
	defer encoder1.Stop()

	md1 := motordriver.New(12, 17, 27)
	md1.SetDirection("forward")
	md1.SetSpeed(10)

	rampDownWindow := RampDownWindow(20, 1000, 100)
	fmt.Println("Ramp down window: ", rampDownWindow)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				log.Printf("Encoder count: %d", encoder1.ReadCounter())
				if math.Abs(float64(encoder1.ReadCounter())) > (float64(2096*2) - rampDownWindow) {
					md1.SetSpeed(0)
					if md1.Direction() == "backward" {
						cancel()
						break
					}

					md1.SetDirection("backward")
					encoder1.ResetCounter()
					time.Sleep(1 * time.Second)
					md1.SetSpeed(10)
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	<-ctx.Done()

	encoder2 := rotaryencoder.New(20, 21)
	encoder2.Start(ctx)
	defer encoder2.Stop()

	md2 := motordriver.New(13, 5, 6)
	md2.SetDirection("forward")
	md2.SetSpeed(10)
	time.Sleep(3 * time.Second)
	md2.SetSpeed(0)

	md1.SetSpeed(0)
	md2.SetSpeed(0)
	fmt.Println("encoder1 count: ", encoder1.ReadCounter())
	fmt.Println("encoder2 count: ", encoder2.ReadCounter())
}

var rampTime = 2000 * time.Millisecond

func RampDownWindow(speed, pwmFrequency, pwmCycleLen uint32) float64 {
	return (float64(pwmFrequency) / 3.0) * (float64(speed) / float64(pwmCycleLen)) * rampTime.Seconds()
}
