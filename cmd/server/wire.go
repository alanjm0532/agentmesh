//go:build wireinject

package main

import (
	"github.com/alanjm0532/agentmesh/configs"
	"github.com/alanjm0532/agentmesh/internal/handler"
	"github.com/alanjm0532/agentmesh/internal/service"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// wireServer 组装配置、业务服务、HTTP/MCP 接口与启动实例。
func wireServer(configPath string, logger *zap.Logger) (*Server, error) {
	panic(wire.Build(
		configs.Load,
		service.ProviderSet,
		handler.ProviderSet,
		newServer,
	))
}
