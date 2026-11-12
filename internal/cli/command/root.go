package command

import (
	"github.com/alanjm0532/agentmesh/internal/cli/config"
	"github.com/spf13/cobra"
)

func Execute() error {
	var cfg config.Config

	root := &cobra.Command{
		Use:           "agentmesh",
		Short:         "AgentMesh command-line client",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}
			if !cmd.Flags().Changed("config") {
				configPath, err = config.DefaultPath()
				if err != nil {
					return err
				}
			}
			cfg, err = config.Load(configPath)
			return err
		},
	}
	flags := root.PersistentFlags()
	flags.String("config", "", "CLI TOML configuration file path (default: agentmesh/config.toml in the OS user configuration directory)")
	root.AddCommand(newPingCommand(&cfg))
	return root.Execute()
}
