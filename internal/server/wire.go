//go:build wireinject

package server

import (
	"github.com/alanjm0532/agentmesh/internal/server/config"
	"github.com/alanjm0532/agentmesh/internal/server/handler"
	"github.com/alanjm0532/agentmesh/internal/server/service"
	"github.com/google/wire"
)

func wireServer(configPath string) (*Server, func(), error) {
	// Assemble configuration, logging, services, and the server with a cleanup function.
	panic(wire.Build(
		config.Load,
		newLogger,
		service.ProviderSet,
		handler.ProviderSet,
		newServer,
	))
}
