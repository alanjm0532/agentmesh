package infra

import (
	"net/http"
	"regexp"
	"time"

	"github.com/alanjm0532/agentmesh/internal/protocol"
	"github.com/alanjm0532/agentmesh/internal/server/config"
	"github.com/alanjm0532/agentmesh/internal/server/handler"
	"github.com/alanjm0532/agentmesh/internal/server/middleware"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewRouter(config config.Config, h *handler.Handler, logger *zap.Logger) *gin.Engine {
	configureValidation()
	gin.SetMode(config.Server.Mode)
	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(
		ginzap.GinzapWithConfig(logger, &ginzap.Config{
			TimeFormat: time.RFC3339,
			UTC:        true,
			SkipPaths:  []string{"/favicon.ico"},
			SkipPathRegexps: []*regexp.Regexp{
				regexp.MustCompile(`^/static(/|$)`),
				regexp.MustCompile(`\.html$`),
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
	registerWebRoutes(router, h, store)
	registerAPIRoutes(router)
	return router
}

func registerWebRoutes(router *gin.Engine, h *handler.Handler, store sessions.Store) {
	pages := router.Group("/")
	registerPages(pages)

	web := router.Group("/web")
	web.Use(sessions.Sessions("agentmesh_session", store))
	web.POST("/login", h.Login)
	web.POST("/logout", h.Logout)
	web.GET("/profile", middleware.SessionAuth(), h.Profile)
	web.GET("/token", middleware.SessionAuth(), h.Token)
	web.POST("/token/reset", middleware.SessionAuth(), h.ResetToken)
}

func registerAPIRoutes(router *gin.Engine) {
	api := router.Group(protocol.APIPrefix)
	api.GET(protocol.PingPath, apiPing)
}

func apiPing(c *gin.Context) {
	c.JSON(http.StatusOK, protocol.PingResponse{Message: "pong"})
}
