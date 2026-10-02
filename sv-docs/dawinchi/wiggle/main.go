package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

var (
	slackWebhook = ""
	jwtToken     = "eHh4"
	wsBasePath   = "ws://127.0.0.1:9001/api/v1/"
	wsQuery      = wsBasePath + "query"
	wsCmd        = wsBasePath + "cmd"
	dialer       = websocket.DefaultDialer

	wiggleUpCounter   = 650000
	wiggleDownCounter = 952242
	wiggleMoveDelay   = 5 * time.Second
	wiggleSpeed       = 30
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	slackWebhook = os.Getenv("SLACK_WEBHOOK")
	if slackWebhook == "" {
		log.Fatal("SLACK_WEBHOOK env var is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sc := slackNew(slackWebhook)

	if err := run(ctx); err != nil {
		errSC := sc.SendAlert(IconError, err.Error())
		if errSC != nil {
			log.Println("slack alert err:", errSC)
		}

		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	header := http.Header{}
	header.Add("Authorization", "Bearer "+jwtToken)

	connQuery, _, err := dialer.DialContext(ctx, wsQuery, header)
	if err != nil {
		return fmt.Errorf("wsQuery: %w", err)
	}
	defer connQuery.Close()

	connCmd, _, err := dialer.DialContext(ctx, wsCmd, header)
	if err != nil {
		return fmt.Errorf("wsCmd: %w", err)
	}
	defer connCmd.Close()

	connStopCmd, _, err := dialer.DialContext(ctx, wsCmd, header)
	if err != nil {
		return fmt.Errorf("wsStopCmd: %w", err)
	}
	defer connCmd.Close()

	// stop motors on exit
	defer sendStopCommand(connStopCmd)

	//
	// wiggle up and down
	//

	// get start counters
	counterResp, err := getCounters(connQuery)
	if err != nil {
		return fmt.Errorf("getCounters: %w", err)
	}

	startInOutCounter := counterResp.Result.InOutMotorCounter

	allDone := make(chan struct{})
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		select {
		case <-ctx.Done():
			err := sendStopCommand(connStopCmd)
			if err != nil {
				return fmt.Errorf("sendStopCommand: %w", err)
			}
			return ctx.Err()
		case <-allDone:
			return nil
		}
	})

	g.Go(func() error {
		defer close(allDone)

		for {
			select {
			case <-ctx.Done():
				err := sendStopCommand(connStopCmd)
				if err != nil {
					return fmt.Errorf("sendStopCommand: %w", err)
				}
				return ctx.Err()

			default:
				err = wiggleUpDown(connCmd, startInOutCounter)
				if err != nil {
					return fmt.Errorf("wiggleUpDown: %w", err)
				}
			}
		}
	})

	err = g.Wait()
	if err != nil {
		return err
	}

	return nil
}

func wiggleUpDown(conn *websocket.Conn, startInOutCounter int) error {
	// no wiggle hours
	currentHour := time.Now().Hour()
	if currentHour >= 0 && currentHour <= 6 {
		fmt.Println("no wiggle hours: 00:00 - 06:00")
		time.Sleep(wiggleMoveDelay)
		return nil
	}

	// move up
	err := sendMoveCommand(conn, startInOutCounter, wiggleUpCounter)
	if err != nil {
		return fmt.Errorf("move up: %w", err)
	}

	time.Sleep(wiggleMoveDelay)

	// move down
	err = sendMoveCommand(conn, startInOutCounter, wiggleDownCounter)
	if err != nil {
		return fmt.Errorf("move down: %w", err)
	}

	time.Sleep(wiggleMoveDelay)

	return nil
}

type CounterResp struct {
	ID     string `json:"id"`
	Result struct {
		Status             string `json:"status"`
		UpDownMotorCounter int    `json:"up_down_motor_counter"`
		UpDownMotorSpeed   int    `json:"up_down_motor_speed"`
		InOutMotorCounter  int    `json:"in_out_motor_counter"`
		InOutMotorSpeed    int    `json:"in_out_motor_speed"`
	} `json:"result"`
	Jsonrpc string `json:"jsonrpc"`
}

func getCounters(conn *websocket.Conn) (*CounterResp, error) {
	jsonMessage :=
		`{
  "id": "wiggle",
  "jsonrpc": "2.0",
  "method": "counter"
}`
	err := sendCommand(conn, jsonMessage)
	if err != nil {
		return nil, err
	}
	resp, err := readResponse(conn)
	if err != nil {
		return nil, err
	}

	var counterResp CounterResp
	err = json.Unmarshal(resp, &counterResp)
	if err != nil {
		return nil, err
	}

	return &counterResp, nil
}

func sendMoveCommand(conn *websocket.Conn, inOutCounter, upDownCounter int) error {
	sio := strconv.Itoa(inOutCounter)
	sud := strconv.Itoa(upDownCounter)
	speed := strconv.Itoa(wiggleSpeed)

	jsonMessage := `
{
    "id": "wiggle",
    "jsonrpc": "2.0",
    "method": "move",
    "params": {
        "up_down_motor_counter": ` + sud + `,
        "up_down_motor_speed": ` + speed + `,
        "in_out_motor_counter": ` + sio + `,
        "in_out_motor_speed": 10,
        "sync": true
    }
}`
	err := sendCommand(conn, jsonMessage)
	if err != nil {
		return err
	}

	_, err = readResponse(conn)
	if err != nil {
		return err
	}

	return nil
}

func sendStopCommand(conn *websocket.Conn) error {
	jsonMessage := `
{
    "id": "wiggle",
    "jsonrpc": "2.0",
    "method": "stop_move"
}`
	err := sendCommand(conn, jsonMessage)
	if err != nil {
		return err
	}

	_, err = readResponse(conn)
	if err != nil {
		return err
	}

	return nil
}

func sendCommand(conn *websocket.Conn, jsonMessage string) error {
	log.Printf("sending command: %s\n", jsonMessage)

	err := conn.WriteMessage(websocket.TextMessage, []byte(jsonMessage))
	if err != nil {
		return err
	}
	return nil
}

type ErrorResponse struct {
	ID    string `json:"id"`
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	Jsonrpc string `json:"jsonrpc"`
}

func readResponse(conn *websocket.Conn) ([]byte, error) {
	_, message, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	log.Printf("received response: %s\n", message)

	// check if response is an error
	var errorResp ErrorResponse
	err = json.Unmarshal(message, &errorResp)
	if err != nil {
		return nil, err
	}

	if errorResp.Error.Code != 0 {
		return nil, fmt.Errorf("response: %s", errorResp.Error.Message)
	}

	return message, nil
}
