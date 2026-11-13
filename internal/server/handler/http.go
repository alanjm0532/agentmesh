package handler

import (
	"net/http"
	"regexp"
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
	userService  *service.UserService
}

func NewHandler(helloService *service.HelloService, userService *service.UserService) *Handler {
	return &Handler{helloService: helloService, userService: userService}
}

func NewRouter(config config.Config, h *Handler, logger *zap.Logger) *gin.Engine {
	gin.SetMode(config.Server.Mode)
	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(
		middleware.RequestID(),
		ginzap.GinzapWithConfig(logger, &ginzap.Config{
			TimeFormat: time.RFC3339,
			UTC:        true,
			SkipPaths:  []string{"/favicon.ico"},
			SkipPathRegexps: []*regexp.Regexp{
				regexp.MustCompile(`^/static(/|$)`),
			},
		}),
		ginzap.RecoveryWithZap(logger, true),
	)
	h.registerWebRoutes(router)
	h.registerAPIRoutes(router)
	return router
}

func (h *Handler) registerWebRoutes(router *gin.Engine) {
	pages := router.Group("/")
	registerPages(pages)

	web := router.Group("/web")
	web.GET("/ping", h.webPing)
	web.GET("/hello", h.hello)
	web.POST(protocol.RegisterPath, h.register)
}

func (h *Handler) registerAPIRoutes(router *gin.Engine) {
	api := router.Group(protocol.APIPrefix)
	api.GET(protocol.PingPath, h.apiPing)
}

func (h *Handler) apiPing(c *gin.Context) {
	c.JSON(http.StatusOK, protocol.PingResponse{Message: "pong"})
}

func (h *Handler) webPing(c *gin.Context) {
	c.JSON(http.StatusOK, protocol.PingResponse{Message: "pong"})
}

func (h *Handler) hello(c *gin.Context) {
	input := protocol.HelloRequest{Name: c.Query("name")}
	c.JSON(http.StatusOK, h.helloService.Hello(input))
}
