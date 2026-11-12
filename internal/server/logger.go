package server

import (
	"fmt"

	"github.com/alanjm0532/agentmesh/internal/server/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newLogger(config config.Config) (*zap.Logger, func(), error) {
	// 按运行模式创建日志实例，并提供给 Wire 调用的清理函数。
	logConfig := zap.NewProductionConfig()
	logConfig.OutputPaths = []string{"stdout"}
	if config.Server.Mode != "release" {
		logConfig.Encoding = "console"
		logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	logger, err := logConfig.Build()
	if err != nil {
		return nil, nil, fmt.Errorf("初始化日志失败: %w", err)
	}
	return logger, func() { _ = logger.Sync() }, nil
}
