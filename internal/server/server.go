package server

import (
	"fmt"
	"os"

	"github.com/alanjm0532/agentmesh/internal/server/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	config config.Config
	router *gin.Engine
	logger *zap.Logger
}

func newServer(config config.Config, router *gin.Engine, logger *zap.Logger) *Server {
	return &Server{config: config, router: router, logger: logger}
}

func (s *Server) Run() error {
	s.logger.Info("starting HTTP server", zap.String("addr", s.config.Server.Addr), zap.String("mode", s.config.Server.Mode))
	return s.router.Run(s.config.Server.Addr)
}

func Run(configPath string) int {
	// Initialize the server through Wire and clean up dependencies before returning.
	server, cleanup, err := wireServer(configPath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "initialize server:", err)
		return 1
	}
	defer cleanup()

	if err = server.Run(); err != nil {
		server.logger.Error("HTTP server stopped", zap.Error(err))
		return 1
	}
	return 0
}
