package command

import (
	"github.com/alanjm0532/agentmesh/internal/cli/config"
	"github.com/spf13/cobra"
)

type commandFactory func(*config.Config) *cobra.Command

var commandFactories []commandFactory

// registerCommand is called from command init functions. Factories create fresh
// commands for each root; configuration is loaded later in PersistentPreRunE.
func registerCommand(factory commandFactory) {
	commandFactories = append(commandFactories, factory)
}
