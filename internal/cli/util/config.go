package util

import (
	"github.com/alanjm0532/agentmesh/internal/cli/config"
	"github.com/spf13/cobra"
)

// ResolveConfigPath returns the explicit config flag or the OS default path.
func ResolveConfigPath(cmd *cobra.Command) (string, error) {
	if cmd.Flags().Changed("config") {
		return cmd.Flags().GetString("config")
	}
	return config.DefaultPath()
}
