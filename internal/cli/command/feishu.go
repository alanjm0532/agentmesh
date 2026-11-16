package command

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/alanjm0532/agentmesh/internal/cli/config"
	"github.com/alanjm0532/agentmesh/internal/cli/feishu"
	"github.com/alanjm0532/agentmesh/internal/cli/repository"
	"github.com/alanjm0532/agentmesh/internal/cli/storage"
	"github.com/alanjm0532/agentmesh/internal/cli/util"
	"github.com/alanjm0532/agentmesh/internal/protocol"
	"github.com/spf13/cobra"
)

func init() {
	registerCommand(newFeishuCommand)
}

func newFeishuCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{Use: "feishu", Short: "Manage local Feishu authorization and email mappings", Args: cobra.NoArgs}
	cmd.AddCommand(newFeishuLoginCommand(cfg))
	cmd.AddCommand(&cobra.Command{
		Use: "bind <user> <email>", Short: "Save an AgentMesh username to Feishu email mapping in local SQLite", Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, cleanup, err := openFeishuRepository(cmd, cfg)
			if err != nil {
				return err
			}
			defer cleanup()
			if err = store.Bind(cmd.Context(), args[0], args[1]); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Bound %s to %s\n", args[0], args[1])
			return err
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use: "unbind <user>", Short: "Remove a local Feishu email mapping", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, cleanup, err := openFeishuRepository(cmd, cfg)
			if err != nil {
				return err
			}
			defer cleanup()
			if err = store.Unbind(cmd.Context(), args[0]); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed Feishu mapping for %s\n", args[0])
			return err
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use: "bindings", Short: "List local Feishu email mappings for the configured server and app", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, cleanup, err := openFeishuRepository(cmd, cfg)
			if err != nil {
				return err
			}
			defer cleanup()
			bindings, err := store.Bindings(cmd.Context())
			if err != nil {
				return err
			}
			for _, binding := range bindings {
				if _, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", binding.Username, binding.Email); err != nil {
					return err
				}
			}
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use: "logout", Short: "Remove local Feishu user tokens (does not revoke the grant in Feishu)", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, cleanup, err := openFeishuRepository(cmd, cfg)
			if err != nil {
				return err
			}
			defer cleanup()
			if err = store.Logout(cmd.Context()); err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Local Feishu tokens removed. Email mappings retained.")
			return err
		},
	})
	return cmd
}

func newFeishuLoginCommand(cfg *config.Config) *cobra.Command {
	var noBrowser bool
	cmd := &cobra.Command{
		Use: "login", Short: "Authorize your Feishu user through the browser with PKCE", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, cleanup, err := openFeishuRepository(cmd, cfg)
			if err != nil {
				return err
			}
			defer cleanup()
			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
			defer cancel()
			api := feishu.New(cfg.Feishu, store, cfg.Timeout)
			err = api.Login(ctx, func(address string) {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Open this URL to authorize Feishu:\n%s\nWaiting for authorization...\n", address)
				if !noBrowser {
					if err := openBrowser(address); err != nil {
						_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Unable to open a browser automatically. Open the URL above manually.")
					}
				}
			})
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Feishu authorized. User tokens saved locally.")
			return err
		},
	}
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Print the authorization URL without opening a browser")
	return cmd
}

func openFeishuRepository(cmd *cobra.Command, cfg *config.Config) (*repository.FeishuRepository, func(), error) {
	path, err := util.ResolveConfigPath(cmd)
	if err != nil {
		return nil, nil, err
	}
	profile, err := filepath.Abs(path)
	if err != nil {
		return nil, nil, err
	}
	database, err := storage.Open(cmd.Context(), profile, cfg.Database)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { _ = database.Close() }
	return repository.NewFeishuRepository(database, profile, cfg.Server, cfg.Feishu.AppID), cleanup, nil
}

func sendFeishu(cmd *cobra.Command, cfg *config.Config, recipient string, message protocol.Message) error {
	store, cleanup, err := openFeishuRepository(cmd, cfg)
	if err != nil {
		return err
	}
	defer cleanup()
	return feishu.New(cfg.Feishu, store, cfg.Timeout).Send(cmd.Context(), recipient, message.ID)
}

func notifyFeishu(cmd *cobra.Command, cfg *config.Config, recipient string, message protocol.Message) {
	enabled := cfg.Feishu.Enabled
	if cmd.Flags().Changed("feishu") {
		enabled, _ = cmd.Flags().GetBool("feishu")
	}
	if !enabled {
		return
	}
	if err := sendFeishu(cmd, cfg, recipient, message); err != nil {
		// 主消息已经发送成功，辅助通知失败只输出警告，避免调用方重试主消息。
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "AgentMesh message %s was sent, but Feishu notification failed: %v\n", message.ID, err)
		return
	}
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Feishu notification sent.")
}
