package command

import (
	"fmt"

	"github.com/alanjm0532/agentmesh/internal/cli/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newPingCommand(cfg *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "检查 HTTP 服务是否可用",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			timeout := cfg.GetDuration("timeout")
			if timeout <= 0 {
				return fmt.Errorf("timeout 必须大于 0")
			}
			server := cfg.GetString("server")
			api := client.New(server, timeout)
			response, err := api.Ping(cmd.Context())
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), response.Message)
			return err
		},
	}
}
