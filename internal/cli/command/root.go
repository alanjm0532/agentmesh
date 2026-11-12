package command

import (
	"fmt"
	"time"

	"github.com/alanjm0532/agentmesh/internal/cli/config"
	"github.com/spf13/cobra"
)

func Execute() error {
	cfg := config.New()

	root := &cobra.Command{
		Use:           "agentmesh",
		Short:         "AgentMesh 命令行客户端",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}
			return config.Load(cfg, configPath, cmd.Flags().Changed("config"))
		},
	}
	flags := root.PersistentFlags()
	flags.String("config", "", "CLI TOML 配置文件路径（默认查找 deployments/cli.toml）")
	flags.String("server", "http://127.0.0.1:8080", "HTTP 服务地址")
	flags.Duration("timeout", 5*time.Second, "请求超时时间")
	if err := cfg.BindPFlags(flags); err != nil {
		return fmt.Errorf("绑定 CLI 配置失败: %w", err)
	}
	root.AddCommand(newPingCommand(cfg))
	return root.Execute()
}
