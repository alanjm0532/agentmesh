package command

import (
	"github.com/alanjm0532/agentmesh/internal/cli/config"
	"github.com/alanjm0532/agentmesh/internal/cli/util"
	"github.com/spf13/cobra"
)

func Execute() error {
	return newRootCommand(new(config.Config)).Execute()
}

func newRootCommand(cfg *config.Config) *cobra.Command {
	root := &cobra.Command{
		Use:           "agentmesh",
		Short:         "AgentMesh command-line client",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			configPath, err := util.ResolveConfigPath(cmd)
			if err != nil {
				return err
			}
			*cfg, err = config.Load(configPath)
			return err
		},
	}
	flags := root.PersistentFlags()
	flags.String("config", "", "CLI TOML configuration file path (default: agentmesh/config.toml in the OS user configuration directory)")
	for _, factory := range commandFactories {
		root.AddCommand(factory(cfg))
	}
	return root
}
