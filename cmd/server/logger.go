package main

import (
	"fmt"

	"github.com/alanjm0532/agentmesh/configs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// newLogger 按运行模式创建 logger，并提供给 Wire 调用的清理函数。
func newLogger(config configs.Config) (*zap.Logger, func(), error) {
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
