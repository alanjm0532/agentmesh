//go:build wireinject

package server

import (
	"github.com/alanjm0532/agentmesh/internal/server/config"
	"github.com/alanjm0532/agentmesh/internal/server/handler"
	"github.com/alanjm0532/agentmesh/internal/server/service"
	"github.com/google/wire"
)

func wireServer(configPath string) (*Server, func(), error) {
	// 组装配置、日志、业务服务与启动实例，并返回清理函数。
	panic(wire.Build(
		config.Load,
		newLogger,
		service.ProviderSet,
		handler.ProviderSet,
		newServer,
	))
}
