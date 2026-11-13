package handler

import (
	"net/http"
	"regexp"
	"time"

	"github.com/alanjm0532/agentmesh/internal/server/config"
	"github.com/alanjm0532/agentmesh/internal/server/middleware"
	"github.com/alanjm0532/agentmesh/internal/server/service"
	"github.com/alanjm0532/agentmesh/pkg/protocol"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	userService *service.UserService
}

func NewHandler(userService *service.UserService) *Handler {
	return &Handler{userService: userService}
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
	store := cookie.NewStore([]byte(config.Server.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/web",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   config.Server.SessionSecure,
		SameSite: http.SameSiteLaxMode,
	})
	h.registerWebRoutes(router, store)
	h.registerAPIRoutes(router)
	return router
}

func (h *Handler) registerWebRoutes(router *gin.Engine, store sessions.Store) {
	pages := router.Group("/")
	registerPages(pages)

	web := router.Group("/web")
	web.Use(sessions.Sessions("agentmesh_session", store))
	web.POST("/register", h.register)
	web.POST("/login", h.login)
	web.POST("/logout", h.logout)
	web.GET("/profile", middleware.SessionAuth(), h.profile)
}

func (h *Handler) registerAPIRoutes(router *gin.Engine) {
	api := router.Group(protocol.APIPrefix)
	api.GET(protocol.PingPath, h.apiPing)
}

func (h *Handler) apiPing(c *gin.Context) {
	c.JSON(http.StatusOK, protocol.PingResponse{Message: "pong"})
}
