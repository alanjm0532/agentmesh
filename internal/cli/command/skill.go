package command

import (
	"fmt"

	"github.com/alanjm0532/agentmesh/internal/cli/config"
	"github.com/alanjm0532/agentmesh/internal/cli/skill"
	"github.com/alanjm0532/agentmesh/internal/cli/version"
	"github.com/spf13/cobra"
)

func init() {
	registerCommand(newSkillCommand)
}

func newSkillCommand(_ *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "skill",
		Short:             "Manage the AgentMesh agent skill",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error { return nil },
	}
	requestedVersion := version.Version
	install := &cobra.Command{
		Use:   "install",
		Short: "Download and install the AgentMesh skill",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			result, err := skill.Install(cmd.Context(), requestedVersion)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Installed skill %s to %s\n", result.Version, result.Path)
			return err
		},
	}
	install.Flags().StringVar(&requestedVersion, "version", requestedVersion, "Skill release tag or latest")
	cmd.AddCommand(install)
	return cmd
}
