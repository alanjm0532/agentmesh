//go:build wireinject

package server

import (
	"github.com/alanjm0532/agentmesh/internal/server/config"
	"github.com/alanjm0532/agentmesh/internal/server/handler"
	"github.com/alanjm0532/agentmesh/internal/server/infra"
	"github.com/alanjm0532/agentmesh/internal/server/repository"
	"github.com/alanjm0532/agentmesh/internal/server/service"
	"github.com/google/wire"
)

func wireServer(configPath string) (*Server, func(), error) {
	// Assemble configuration, logging, persistence, services, and the server.
	panic(wire.Build(
		config.Load,
		infra.ProviderSet,
		repository.ProviderSet,
		service.ProviderSet,
		handler.ProviderSet,
		newServer,
	))
}
