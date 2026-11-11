package handler

import (
	"net/http"
	"time"

	"github.com/alanjm0532/agentmesh/configs"
	"github.com/alanjm0532/agentmesh/internal/middleware"
	"github.com/alanjm0532/agentmesh/internal/model"
	"github.com/alanjm0532/agentmesh/internal/service"
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

func NewRouter(config configs.Config, h *Handler, logger *zap.Logger) *gin.Engine {
	gin.SetMode(config.Server.Mode)
	router := gin.New()
	router.Use(
		middleware.RequestID(),
		ginzap.Ginzap(logger, time.RFC3339, true),
		ginzap.RecoveryWithZap(logger, true),
	)
	router.GET("/ping", h.ping)
	router.GET("/hello", h.hello)
	router.Any("/mcp", gin.WrapH(h.mcpHandler()))
	return router
}

func (h *Handler) ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func (h *Handler) hello(c *gin.Context) {
	input := model.HelloInput{Name: c.Query("name")}
	c.JSON(http.StatusOK, h.helloService.Hello(input))
}
