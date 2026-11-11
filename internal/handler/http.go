package handler

import (
	"net/http"

	"github.com/alanjm0532/agentmesh/internal/middleware"
	"github.com/alanjm0532/agentmesh/internal/model"
	"github.com/alanjm0532/agentmesh/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	helloService *service.HelloService
}

func NewRouter(helloService *service.HelloService) *gin.Engine {
	h := &Handler{helloService: helloService}
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), middleware.RequestID())
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
