//go:build wireinject

package testutil

import (
	"github.com/alanjm0532/agentmesh/internal/server/config"
	"github.com/alanjm0532/agentmesh/internal/server/repository"
	"github.com/alanjm0532/agentmesh/internal/server/service"
	"github.com/google/wire"
)

func initialize(configPath string) (*Dependencies, func(), error) {
	// Reuse business providers with the test database connection.
	panic(wire.Build(
		config.Load,
		repository.ProviderSet,
		service.ProviderSet,
		wire.Struct(new(Dependencies), "*"),
	))
}
