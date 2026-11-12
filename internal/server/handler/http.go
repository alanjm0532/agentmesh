package handler

import (
	"net/http"
	"time"

	"github.com/alanjm0532/agentmesh/internal/server/config"
	"github.com/alanjm0532/agentmesh/internal/server/middleware"
	"github.com/alanjm0532/agentmesh/internal/server/service"
	"github.com/alanjm0532/agentmesh/pkg/protocol"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	helloService *service.HelloService
}

func NewHandler(helloService *service.HelloService) *Handler {
	return &Handler{helloService: helloService}
}

func NewRouter(config config.Config, h *Handler, logger *zap.Logger) *gin.Engine {
	gin.SetMode(config.Server.Mode)
	router := gin.New()
	router.Use(
		middleware.RequestID(),
		ginzap.Ginzap(logger, time.RFC3339, true),
		ginzap.RecoveryWithZap(logger, true),
	)
	router.GET(protocol.PingPath, h.ping)
	router.GET("/hello", h.hello)
	return router
}

func (h *Handler) ping(c *gin.Context) {
	c.JSON(http.StatusOK, protocol.PingResponse{Message: "pong"})
}

func (h *Handler) hello(c *gin.Context) {
	input := protocol.HelloRequest{Name: c.Query("name")}
	c.JSON(http.StatusOK, h.helloService.Hello(input))
}
