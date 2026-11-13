//go:build wireinject

package testutil

import (
	"github.com/alanjm0532/agentmesh/internal/server/config"
	"github.com/alanjm0532/agentmesh/internal/server/repository"
	"github.com/alanjm0532/agentmesh/internal/server/service"
	"github.com/google/wire"
	"go.uber.org/zap"
)

func initialize(configPath string) (*Dependencies, func(), error) {
	// Reuse business providers with the test database connection and silent logger.
	panic(wire.Build(
		config.Load,
		zap.NewNop,
		repository.ProviderSet,
		service.ProviderSet,
		wire.Struct(new(Dependencies), "*"),
	))
}
