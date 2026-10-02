package rpc

import (
	"net/http"

	_ "github.com/CreateView/dawinchi-hw/docs/swagger"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Dawinchi HW API
// @version 1.0
// @basePath /
// @schemes http ws
func (s *Server) routes() {
	s.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	apiV1 := s.Group("/api/v1")
	apiV1.GET("/", s.rootRoute)
	apiV1.GET("/health", s.healthRoute)
	apiV1.GET("/ping", s.pingRoute)

	authV1 := apiV1.Group("/", validateToken(s.bearerToken))
	authV1.GET("cmd", s.rpcCmdHandler)
	authV1.GET("event", s.rpcEventHandler)
}

func (s *Server) rootRoute(c *gin.Context) {
	c.JSON(http.StatusOK, "Welcome to Dawinchi HW API V1")
}

// @Summary Health endpoint.
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {string} ok
// @Router /api/v1/health [get]
func (s *Server) healthRoute(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

// @Summary Ping endpoint.
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {string} pong
// @Router /api/v1/ping [get]
func (s *Server) pingRoute(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}
