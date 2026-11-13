package testutil

import (
	"path/filepath"
	"testing"

	"github.com/alanjm0532/agentmesh/internal/server/ent"
	"github.com/alanjm0532/agentmesh/internal/server/repository"
	"github.com/alanjm0532/agentmesh/internal/server/service"
)

const configPath = "deployments/server.toml"

type Dependencies struct {
	Client         *ent.Client
	UserRepository *repository.UserRepository
	UserService    *service.UserService
}

func New(t testing.TB) *Dependencies {
	t.Helper()

	// Wire assembles services and repositories using the test database client.
	deps, cleanup, err := initialize(filepath.Join(projectRoot(t), configPath))
	if err != nil {
		t.Fatalf("initialize test dependencies: %v", err)
	}
	t.Cleanup(cleanup)
	return deps
}
