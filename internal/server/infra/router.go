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
	registerWebRoutes(router, h, store, config.Server.RequestTimeout)
	registerAPIRoutes(router, h, config.Server.RequestTimeout)
	return router
}

func registerWebRoutes(router *gin.Engine, h *handler.Handler, store sessions.Store, timeout time.Duration) {
	pages := router.Group("/")
	registerPages(pages)

	web := router.Group("/web")
	web.Use(middleware.Timeout(timeout), sessions.Sessions("agentmesh_session", store), middleware.WebCSRF(), func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
		c.Next()
	})
	web.POST("/login", middleware.Adapt(h.Login))
	web.POST("/logout", middleware.Adapt(h.Logout))
	protected := web.Group("", middleware.SessionAuth())
	protected.GET("/profile", middleware.AdaptAuthenticated(h.Profile))
	protected.GET("/tokens", middleware.AdaptAuthenticated(h.ListTokens))
	protected.POST("/tokens/:id/rename", middleware.AdaptAuthenticated(h.RenameToken))
	protected.POST("/tokens/:id/revoke", middleware.AdaptAuthenticated(h.RevokeToken))
	protected.GET("/cli/request", middleware.AdaptAuthenticated(h.CLIRequest))
	protected.POST("/cli/authorize", middleware.AdaptAuthenticated(h.AuthorizeCLI))
}

func registerAPIRoutes(router *gin.Engine, h *handler.Handler, timeout time.Duration) {
	api := router.Group(protocol.APIPrefix, middleware.Timeout(timeout))
	api.GET(protocol.PingPath, middleware.Adapt(h.APIPing))
	auth := api.Group("/auth/device", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
		c.Next()
	})
	auth.POST("/code", middleware.Adapt(h.DeviceCode))
	auth.POST("/token", middleware.Adapt(h.DeviceToken))
	protected := api.Group("", h.RequireToken(), func(c *gin.Context) { c.Header("Cache-Control", "no-store"); c.Next() })
	protected.GET(protocol.MePath, middleware.AdaptAuthenticated(h.Me))
	protected.POST(protocol.RevokeTokenPath, middleware.AdaptAuthenticated(h.RevokeCurrentToken))
	protected.POST(protocol.MessagesPath, middleware.AdaptAuthenticated(h.SendMessage))
	protected.GET(protocol.UnreadMessagesPath, middleware.AdaptAuthenticated(h.UnreadMessages))
	protected.GET(protocol.MessageHistoryPath, middleware.AdaptAuthenticated(h.MessageHistory))
	protected.POST(protocol.ReadMessagesPath, middleware.AdaptAuthenticated(h.ReadMessages))
}
