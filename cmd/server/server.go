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

// run 启动 Wire 组装的服务，并在返回前清理依赖。
func run(configPath string) int {
	server, cleanup, err := wireServer(configPath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "初始化服务失败:", err)
		return 1
	}
	defer cleanup()

	if err = server.Run(); err != nil {
		server.logger.Error("服务退出", zap.Error(err))
		return 1
	}
	return 0
}
