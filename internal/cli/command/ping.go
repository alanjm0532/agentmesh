package command

import (
	"fmt"
	"os"

	"github.com/alanjm0532/agentmesh/internal/cli/client"
	"github.com/alanjm0532/agentmesh/pkg/protocol"
	"github.com/charmbracelet/x/term"
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
			var response protocol.PingResponse
			var err error
			// 交互界面写入 stderr；管道和重定向场景直接请求并输出结果。
			if term.IsTerminal(os.Stdin.Fd()) && term.IsTerminal(os.Stdout.Fd()) && term.IsTerminal(os.Stderr.Fd()) {
				response, err = runPingUI(cmd.Context(), api, server, cmd.InOrStdin(), cmd.ErrOrStderr())
			} else {
				response, err = api.Ping(cmd.Context())
			}
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), response.Message)
			return err
		},
	}
}
