//go:build wireinject

package main

import (
	"github.com/alanjm0532/agentmesh/configs"
	"github.com/alanjm0532/agentmesh/internal/handler"
	"github.com/alanjm0532/agentmesh/internal/service"
	"github.com/google/wire"
)

// wireServer 组装配置、日志、业务服务与启动实例，并返回清理函数。
func wireServer(configPath string) (*Server, func(), error) {
	panic(wire.Build(
		configs.Load,
		newLogger,
		service.ProviderSet,
		handler.ProviderSet,
		newServer,
	))
}
