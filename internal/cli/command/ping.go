package command

import (
	"fmt"

	"github.com/alanjm0532/agentmesh/internal/cli/client"
	"github.com/alanjm0532/agentmesh/internal/cli/config"
	"github.com/spf13/cobra"
)

func newPingCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "Check HTTP server availability",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			api := client.New(cfg.Server, cfg.Timeout)
			response, err := api.Ping(cmd.Context())
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), response.Message)
			return err
		},
	}
}
