package server

import (
	"fmt"

	"github.com/alanjm0532/agentmesh/internal/server/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newLogger(config config.Config) (*zap.Logger, func(), error) {
	// Configure logging for the server mode and provide a cleanup function to Wire.
	logConfig := zap.NewProductionConfig()
	logConfig.OutputPaths = []string{"stdout"}
	if config.Server.Mode != "release" {
		logConfig.Encoding = "console"
		logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	logger, err := logConfig.Build()
	if err != nil {
		return nil, nil, fmt.Errorf("initialize logger: %w", err)
	}
	return logger, func() { _ = logger.Sync() }, nil
}
