package rpc

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"sync"

	"github.com/CreateView/dawinchi-hw/internal/winch"
	"github.com/gin-gonic/gin"
	"github.com/hpcloud/tail"
	"github.com/pkg/errors"
	"github.com/sourcegraph/jsonrpc2"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
)

// WinchService is the service that will be used for JSON-RPC calls.
type WinchService struct {
	winch Winch
}

// @Summary Websocket(ws) endpoint for sending commands to the winch.
// @Description Creates a websocket connection with [JSON-RPC 2.0 protocol](https://www.jsonrpc.org/specification). _(batch requests are not supported)_
// @Description The client can send requests and receive notifications.
// @Description The server can receive requests, send responses and notifications.
// @Description
// @Description # Available methods:
// @Description ---
// @Description ## *move*:
// @Description  - moves the winch to the specified right, left motor counters (computed by client in relation to x,z cordiantes).
// @Description  - _response is streamed to the client, on [event](#/websocket/post_api_v1_event) endpoint, until the winch reaches the specified counters._
// @Description
// @Description #### _Request_:
// @Description `
// @Description {
// @Description   "id": 1,
// @Description   "jsonrpc": "2.0",
// @Description   "method": "move",
// @Description   "params": {
// @Description     "right_motor_counter": 22000,
// @Description     "left_motor_counter": 22000,
// @Description     "speed": 10
// @Description   }
// @Description }
// @Description `
// @Description
// @Description #### _Response_:
// @Description `
// @Description {
// @Description   "id": 1,
// @Description   "jsonrpc": "2.0",
// @Description   "result": {
// @Description     "status": "running",
// @Description    	"right_motor_counter": 1000,
// @Description    	"left_motor_counter": 1000,
// @Description    	"speed": 10
// @Description   }
// @Description }
// @Description `
// @Description ..........................................
// @Description `
// @Description {
// @Description   "id": 1,
// @Description   "jsonrpc": "2.0",
// @Description   "result": {
// @Description     "status": "done",
// @Description    	"right_motor_counter": 22000,
// @Description    	"left_motor_counter": 22000,
// @Description    	"speed": 10
// @Description   }
// @Description }
// @Description `
// @Description ---
// @Description ## *state*:
// @Description  - sends the current state of the winch to the client.
// @Description
// @Description #### _Request_:
// @Description `
// @Description {
// @Description   "id": 2,
// @Description   "jsonrpc": "2.0",
// @Description   "method": "state"
// @Description }
// @Description `
// @Description
// @Description #### _Response_:
// @Description `
// @Description {
// @Description	"jsonrpc": "2.0",
// @Description	"method": "state",
// @Description	"params": {
// @Description		"right_motor": {
// @Description			"direction": 0,
// @Description			"running": false,
// @Description			"speed": 0,
// @Description			"counter": 0,
// @Description			"driver_status": {
// @Description				"status": 1,
// @Description				"error": null
// @Description			},
// @Description			"break_status": {
// @Description				"status": 1,
// @Description				"error": null
// @Description			},
// @Description			"current_state": {
// @Description				"voltage": 0.070734375,
// @Description				"amperage": 2.08671875,
// @Description				"error": null
// @Description			}
// @Description		},
// @Description		"left_motor": {
// @Description			"direction": 0,
// @Description			"running": false,
// @Description			"speed": 0,
// @Description			"counter": 0,
// @Description			"driver_status": {
// @Description				"status": 1,
// @Description				"error": null
// @Description			},
// @Description			"break_status": {
// @Description				"status": 1,
// @Description				"error": null
// @Description			},
// @Description			"current_state": {
// @Description				"voltage": 0.0478125,
// @Description				"amperage": 0.9406249999999999,
// @Description				"error": null
// @Description			}
// @Description		},
// @Description		"buttons_panel": {
// @Description			"led_locked_state": 1,
// @Description			"led_error_state": 1,
// @Description			"button_home_state": "ButtonHomeRelease",
// @Description			"button_home_last_state": "ButtonHomeRelease",
// @Description			"button_up_state": "ButtonUpRelease",
// @Description			"button_up_last_state": "ButtonUpRelease",
// @Description			"button_down_state": "ButtonDownRelease",
// @Description			"button_down_last_state": "ButtonDownRelease",
// @Description			"button_in_state": "ButtonInRelease",
// @Description			"button_in_last_state": "ButtonInRelease",
// @Description			"button_out_state": "ButtonOutRelease",
// @Description			"button_out_last_state": "ButtonOutRelease",
// @Description			"action_status_right": "On",
// @Description			"action_status_left": "On"
// @Description		},
// @Description		"status": "",
// @Description		"error": ""
// @Description	  }
// @Description	}
// @Description `
// @Description ---
// @Description **JSON-RPC 2.0 protocol**:
// @Description
// @Description **Request** is a JSON object:
// @Description - id: An identifier established by the Client that MUST contain a Number.
// @Description - jsonrpc: A String specifying the version of the JSON-RPC protocol.
// @Description - method: A String containing the name of the method to be invoked.
// @Description - params: A Structured value that holds the parameter values to be used during the invocation of the method.
// @Description
// @Description **Response** is a JSON object:
// @Description - id: REQUIRED identifier matching **Request** id.
// @Description - jsonrpc: A String specifying the version of the JSON-RPC protocol.
// @Description - result: REQUIRED on success.
// @Description - error: REQUIRED on error.
// @Description
// @Description **Notification** is a JSON object:
// @Description - jsonrpc: A String specifying the version of the JSON-RPC protocol.
// @Description - method: A String containing the name of the method to be invoked.
// @Description - params: A Structured value that holds the parameter values to be used during the invocation of the method.
// @Description
// @Description **Error** is a JSON object:
// @Description - code: A Number that indicates the error type that occurred.
// @Description - message: A String providing a short description of the error.
// @Description - data: A Primitive or Structured value that contains additional information about the error.
// @Description
// @Tags websocket
// @Accept json
// @Produce json
// @Router /api/v1/cmd [post]
func (s *WinchService) Handle(ctx context.Context, c *jsonrpc2.Conn, r *jsonrpc2.Request) {
	switch r.Method {
	case "move":
		moveReq := &winchMoveRequest{}
		if err := json.Unmarshal(*r.Params, moveReq); err != nil {
			log.Println(err) // TODO: save to event log file

			respErr := &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInvalidRequest,
				Message: err.Error(),
			}
			err = c.ReplyWithError(ctx, r.ID, respErr)
			if err != nil {
				log.Println(err)
			}
			return
		}

		err := s.move(int(r.ID.Num), moveReq)
		if err != nil {
			log.Println(err) // TODO: save to event log file

			respErr := &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInvalidRequest,
				Message: err.Error(),
			}
			err = c.ReplyWithError(ctx, r.ID, respErr)
			if err != nil {
				log.Println(err)
			}
			return
		}

		status := "started"
		if moveReq.Sync {
			status = "done"
		}

		state := s.winch.State()
		moveResp := &winchMoveResponse{
			Status:            status,
			MotorRightCounter: state.RightMotor.Counter,
			MotorLeftCounter:  state.LeftMotor.Counter,
			Speed:             moveReq.Speed,
			Sync:              moveReq.Sync,
		}

		if err := c.Reply(ctx, r.ID, moveResp); err != nil {
			log.Println(err) // TODO: save to event log file
			return
		}

	default:
		err := &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: "Method not found"}
		if err := c.ReplyWithError(ctx, r.ID, err); err != nil {
			log.Println(err)
			return
		}
	}
}

type winchMoveRequest struct {
	MotorRightCounter uint32 `json:"motor_right_counter"`
	MotorLeftCounter  uint32 `json:"motor_left_counter"`
	Speed             uint32 `json:"speed"`
	Sync              bool   `json:"sync"`
}

type winchMoveResponse struct {
	Status            string `json:"status"`
	MotorRightCounter uint32 `json:"motor_right_counter"`
	MotorLeftCounter  uint32 `json:"motor_left_counter"`
	Speed             uint32 `json:"speed"`
	Sync              bool   `json:"sync"`
}

func (s *WinchService) move(requestID int, moveReq *winchMoveRequest) error {
	ctx := context.Background()

	if moveReq.Sync {
		err := s.winch.Move(ctx, requestID, moveReq.MotorRightCounter, moveReq.MotorLeftCounter, moveReq.Speed)
		if err != nil {
			return errors.Wrap(err, "winch.Move: sync")
		}
		return nil
	}

	go func() {
		err := s.winch.Move(ctx, requestID, moveReq.MotorRightCounter, moveReq.MotorLeftCounter, moveReq.Speed)
		if err != nil {
			return
		}
	}()

	return nil
}

func (s *Server) rpcCmdHandler(ctx *gin.Context) {
	w, r := ctx.Writer, ctx.Request

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Error().Err(err).Msg("rpcCmdHandler: Upgrade")
		return
	}

	jc := jsonrpc2.NewConn(
		r.Context(),
		websocketjsonrpc2.NewObjectStream(conn),
		&WinchService{
			winch: s.winch,
		},
		// jsonrpc2.LogMessages(&s.log),
	)
	defer jc.Close()

	<-jc.DisconnectNotify()
}

// WinchEventService is the service that will be used for JSON-RPC calls.
type WinchEventService struct {
	winch Winch

	mu           sync.Mutex
	cancelStream context.CancelFunc
}

// @Summary Websocket(ws) endpoint receiving events from the winch.
// @Description Creates a websocket connection with [JSON-RPC 2.0 protocol](https://www.jsonrpc.org/specification). _(batch requests are not supported)_
// @Description The client can send requests and receive notifications.
// @Description The server can receive requests, send responses and notifications.
// @Description
// @Description # Available methods:
// @Description ---
// @Description ## *event*:
// @Description  - starts streaming events from the winch.
// @Description  - start_id: starts stream from event ID to current event.
// @Description  - filter: can filter by method and request_id.
// @Description
// @Description #### _Request_:
// @Description `
// @Description {
// @Description 	"id": 1,
// @Description 	"jsonrpc": "2.0",
// @Description 	"method": "event",
// @Description 	"params": {
// @Description 		"start_id": 13,
// @Description 		"filter": {
// @Description 			"request_id": 1,
// @Description 			"method": "move"
// @Description 		}
// @Description 	}
// @Description }
// @Description `
// @Description
// @Description #### _Response_:
// @Description `
// @Description {
// @Description 	"id": 1,
// @Description 	"result": "started",
// @Description 	"jsonrpc": "2.0"
// @Description }
// @Description `
// @Description ---
// @Description
// @Description # Available events:
// @Description ---
// @Description ## *move*:
// @Description  - _streamed `move` method(command) until the winch reaches the specified counters._
// @Description
// @Description #### _Response_:
// @Description `
// @Description {
// @Description   "jsonrpc": "2.0",
// @Description   "method": "move",
// @Description   "params": {
// @Description       "event_id": 1,
// @Description       "status": "running",
// @Description       "right_motor_counter": 230,
// @Description       "left_motor_counter": 230,
// @Description       "speed": 10,
// @Description       "sync": true
// @Description   }
// @Description }
// @Description `
// @Description ..........................................
// @Description `
// @Description {
// @Description   "jsonrpc": "2.0",
// @Description   "method": "move",
// @Description   "params": {
// @Description       "event_id": 2,
// @Description       "status": "done",
// @Description       "right_motor_counter": 1000,
// @Description       "left_motor_counter": 1000,
// @Description       "speed": 0,
// @Description       "sync": true
// @Description   }
// @Description }
// @Description `
// @Description ---
// @Description ## *log*:
// @Description  - _streamed `log` winch._
// @Description
// @Description #### _Response_:
// @Description `
// @Description {
// @Description   "jsonrpc": "2.0",
// @Description   "method": "log",
// @Description   "params": {
// @Description       "event_id": 3,
// @Description       "level": "info",
// @Description       "message": "TODO: format for payload logs",
// @Description   	  "payload": {}
// @Description   }
// @Description }
// @Description `
// @Description ..........................................
// @Description `
// @Description {
// @Description   "jsonrpc": "2.0",
// @Description   "method": "log",
// @Description   "params": {
// @Description       "event_id": 4,
// @Description       "level": "info",
// @Description       "message": "TODO: format for payload logs",
// @Description   	  "payload": {}
// @Description   }
// @Description }
// @Description `
// @Description ---
// @Description ## *state*:
// @Description  - sends the current state of the winch to the client.
// @Description
// @Description #### _Response_:
// @Description `
// @Description {
// @Description	"jsonrpc": "2.0",
// @Description	"method": "state",
// @Description	"params": {
// @Description		"right_motor": {
// @Description			"direction": 0,
// @Description			"running": false,
// @Description			"speed": 0,
// @Description			"counter": 0,
// @Description			"driver_status": {
// @Description				"status": 1,
// @Description				"error": null
// @Description			},
// @Description			"break_status": {
// @Description				"status": 1,
// @Description				"error": null
// @Description			},
// @Description			"current_state": {
// @Description				"voltage": 0.070734375,
// @Description				"amperage": 2.08671875,
// @Description				"error": null
// @Description			}
// @Description		},
// @Description		"left_motor": {
// @Description			"direction": 0,
// @Description			"running": false,
// @Description			"speed": 0,
// @Description			"counter": 0,
// @Description			"driver_status": {
// @Description				"status": 1,
// @Description				"error": null
// @Description			},
// @Description			"break_status": {
// @Description				"status": 1,
// @Description				"error": null
// @Description			},
// @Description			"current_state": {
// @Description				"voltage": 0.0478125,
// @Description				"amperage": 0.9406249999999999,
// @Description				"error": null
// @Description			}
// @Description		},
// @Description		"buttons_panel": {
// @Description			"led_locked_state": 1,
// @Description			"led_error_state": 1,
// @Description			"button_home_state": "ButtonHomeRelease",
// @Description			"button_home_last_state": "ButtonHomeRelease",
// @Description			"button_up_state": "ButtonUpRelease",
// @Description			"button_up_last_state": "ButtonUpRelease",
// @Description			"button_down_state": "ButtonDownRelease",
// @Description			"button_down_last_state": "ButtonDownRelease",
// @Description			"button_in_state": "ButtonInRelease",
// @Description			"button_in_last_state": "ButtonInRelease",
// @Description			"button_out_state": "ButtonOutRelease",
// @Description			"button_out_last_state": "ButtonOutRelease",
// @Description			"action_status_right": "On",
// @Description			"action_status_left": "On"
// @Description		},
// @Description		"status": "",
// @Description		"error": ""
// @Description	  }
// @Description	}
// @Description `
// @Description ---
// @Description **JSON-RPC 2.0 protocol**:
// @Description
// @Description **Request** is a JSON object:
// @Description - id: An identifier established by the Client that MUST contain a Number.
// @Description - jsonrpc: A String specifying the version of the JSON-RPC protocol.
// @Description - method: A String containing the name of the method to be invoked.
// @Description - params: A Structured value that holds the parameter values to be used during the invocation of the method.
// @Description
// @Description **Response** is a JSON object:
// @Description - id: REQUIRED identifier matching **Request** id.
// @Description - jsonrpc: A String specifying the version of the JSON-RPC protocol.
// @Description - result: REQUIRED on success.
// @Description - error: REQUIRED on error.
// @Description
// @Description **Notification** is a JSON object:
// @Description - jsonrpc: A String specifying the version of the JSON-RPC protocol.
// @Description - method: A String containing the name of the method to be invoked.
// @Description - params: A Structured value that holds the parameter values to be used during the invocation of the method.
// @Description
// @Description **Error** is a JSON object:
// @Description - code: A Number that indicates the error type that occurred.
// @Description - message: A String providing a short description of the error.
// @Description - data: A Primitive or Structured value that contains additional information about the error.
// @Description
// @Tags websocket
// @Accept json
// @Produce json
// @Router /api/v1/event [post]
func (s *WinchEventService) Handle(ctx context.Context, c *jsonrpc2.Conn, r *jsonrpc2.Request) {
	switch r.Method {
	case "event":
		// stop previous stream
		s.cancelStream()

		err := s.streamEvent(ctx, c, r)
		if err != nil {
			log.Println(err) // TODO: save to event log file

			respErr := &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInvalidRequest,
				Message: err.Error(),
			}
			err = c.ReplyWithError(ctx, r.ID, respErr)
			if err != nil {
				log.Println(err)
			}
			return
		}

		if err := c.Reply(ctx, r.ID, "started"); err != nil {
			log.Println(err) // TODO: save to event log file
			return
		}

	default:
		err := &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: "Method not found"}
		if err := c.ReplyWithError(ctx, r.ID, err); err != nil {
			log.Println(err)
			return
		}
	}
}

type eventStreamRequest struct {
	StartID uint64            `json:"start_id"`
	Filter  eventStreamFilter `json:"filter"`
}

type eventStreamFilter struct {
	RequestID uint64 `json:"request_id"`
	Method    string `json:"method"`
}

type eventStreamResponse struct {
}

func (s *WinchEventService) streamEvent(ctx context.Context, c *jsonrpc2.Conn, r *jsonrpc2.Request) error {
	// TODO: on connection start stream winch log events to the client
	ctx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	s.cancelStream = cancel
	s.mu.Unlock()

	eventStreamReq := &eventStreamRequest{}
	if err := json.Unmarshal(*r.Params, eventStreamReq); err != nil {
		return errors.Wrap(err, "json.Unmarshal params")
	}

	if eventStreamReq.StartID != 0 {
		log.Println("start_id:", eventStreamReq.StartID)

		// TODO: send all events from the winch log file starting from start_id
		// TODO: check filter
		es, err := s.winch.EventLogsFromID(eventStreamReq.StartID)
		if err != nil {
			return errors.Wrap(err, "eventLogsFromID")
		}

		for _, e := range es {
			err = c.Notify(ctx, e.Method, e.Params)
			if err != nil {
				log.Println(err)
			}
		}
	}

	t, err := tail.TailFile(
		s.winch.EventLogFilePath(),
		tail.Config{Follow: true, Location: &tail.SeekInfo{Offset: 0, Whence: io.SeekEnd}},
	)
	if err != nil {
		log.Fatalf("tail file err: %v", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("streamEvent: ctx.Done")

				t.Cleanup()
				return

			case line := <-t.Lines:
				if line.Text != "" {
					event := winch.EventLog{}
					err = json.Unmarshal([]byte(line.Text), &event)
					if err != nil {
						log.Println(err)
						continue
					}

					err = c.Notify(ctx, event.Method, event.Params)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}()

	return nil
}

func (s *Server) rpcEventHandler(ctx *gin.Context) {
	w, r := ctx.Writer, ctx.Request

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Error().Err(err).Msg("rpcEventHandler: Upgrade")
		return
	}

	jc := jsonrpc2.NewConn(
		r.Context(),
		websocketjsonrpc2.NewObjectStream(conn),
		&WinchEventService{
			winch:        s.winch,
			cancelStream: func() {},
		},
		// jsonrpc2.LogMessages(&s.log),
	)
	defer jc.Close()

	<-jc.DisconnectNotify()
}
