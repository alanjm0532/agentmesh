package command

import (
	"fmt"

	"github.com/alanjm0532/agentmesh/internal/cli/config"
	"github.com/alanjm0532/agentmesh/internal/cli/util"
	"github.com/spf13/cobra"
)

func init() {
	registerCommand(newConfigCommand)
}

func newConfigCommand(_ *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
		// Configuration commands must work before a configuration file exists.
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error { return nil },
	}
	set := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Save a CLI setting (server or timeout)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := util.ResolveConfigPath(cmd)
			if err != nil {
				return err
			}
			if err = config.Set(path, args[0], args[1]); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Saved %s to %s\n", args[0], path)
			return err
		},
	}
	cmd.AddCommand(set)
	return cmd
}
