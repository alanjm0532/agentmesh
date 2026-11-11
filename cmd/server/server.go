package main

import (
	"fmt"
	"os"

	"github.com/alanjm0532/agentmesh/configs"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	config configs.Config
	router *gin.Engine
	logger *zap.Logger
}

func newServer(config configs.Config, router *gin.Engine, logger *zap.Logger) *Server {
	return &Server{config: config, router: router, logger: logger}
}

func (s *Server) Run() error {
	s.logger.Info("启动服务", zap.String("addr", s.config.Server.Addr), zap.String("mode", s.config.Server.Mode))
	return s.router.Run(s.config.Server.Addr)
}

// run 在退出进程前刷新日志，并将同一个 logger 交给 Wire 注入。
func run(configPath string) int {
	logger, err := zap.NewProduction()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "初始化日志失败:", err)
		return 1
	}
	defer func() { _ = logger.Sync() }()

	server, err := wireServer(configPath, logger)
	if err != nil {
		logger.Error("初始化服务失败", zap.Error(err))
		return 1
	}
	if err = server.Run(); err != nil {
		logger.Error("服务退出", zap.Error(err))
		return 1
	}
	return 0
}
