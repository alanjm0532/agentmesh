package infra

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewDatabase,
	NewLogger,
	NewRouter,
)
