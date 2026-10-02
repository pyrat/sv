package rpc

import (
	"context"
	"net/http"

	"github.com/CreateView/dawinchi-hw/internal/winch"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type Winch interface {
	Move(ctx context.Context, requestID int, motorRightCounter, motorLeftCounter uint32, speed uint32) error
	EventLogFilePath() string
	EventLogsFromID(id uint64) ([]winch.EventLog, error)
	State() *winch.State
}

type Server struct {
	*gin.Engine
	bearerToken string

	upgrader websocket.Upgrader

	winch Winch

	log zerolog.Logger
}

func New(bearerToken string, winch Winch, log zerolog.Logger) (*Server, error) {
	r := gin.Default()

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// fmt.Println("CheckOrigin RemoteAddr:", r.RemoteAddr)
			// fmt.Println("CheckOrigin Host:", r.Host)

			// if r.RemoteAddr != r.Host {
			// 	return false
			// }

			// Allow all origins for this example; you may want to restrict this in production.
			return true
		},
	}

	s := &Server{
		Engine:      r,
		bearerToken: bearerToken,
		upgrader:    upgrader,
		winch:       winch,
		log:         log,
	}

	// load routes, middlewares
	s.routes()

	return s, nil
}
