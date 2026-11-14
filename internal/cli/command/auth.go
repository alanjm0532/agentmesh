package command

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/alanjm0532/agentmesh/internal/cli/client"
	"github.com/alanjm0532/agentmesh/internal/cli/config"
	"github.com/alanjm0532/agentmesh/internal/cli/util"
	"github.com/alanjm0532/agentmesh/internal/protocol"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func init() {
	registerCommand(newAuthCommand)
}

func newLoginCommand(cfg *config.Config) *cobra.Command {
	var noBrowser bool
	var name string
	cmd := &cobra.Command{
		Use: "login", Short: "Log in through your browser", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := util.ResolveConfigPath(cmd)
			if err != nil {
				return err
			}
			api := client.New(cfg.Server, cfg.Timeout)
			if strings.TrimSpace(name) == "" {
				name, _ = os.Hostname()
				if name == "" {
					name = "AgentMesh CLI"
				}
			}
			if len([]rune(name)) > 100 {
				return fmt.Errorf("token name must not exceed 100 characters")
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
			defer cancel()
			grant, err := api.StartLogin(ctx, name)
			if err != nil {
				return err
			}
			deviceCode, err := uuid.Parse(grant.DeviceCode)
			if err != nil || deviceCode.Version() != 4 || len(grant.DeviceCode) != 36 || grant.ExpiresIn <= 0 || grant.Interval < 1 || grant.Interval > 60 || grant.UserCode == "" {
				return fmt.Errorf("server returned an invalid login response")
			}
			address := strings.TrimRight(cfg.Server, "/") + grant.VerificationURIComplete
			loginCtx, stop := context.WithTimeout(ctx, time.Duration(grant.ExpiresIn)*time.Second)
			defer stop()
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Open this URL to authorize your CLI:\n%s\nAuthorization code: %s\n", address, grant.UserCode)
			if !noBrowser {
				if err := openBrowser(address); err != nil {
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Unable to open a browser automatically. Open the URL above manually.")
				}
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Waiting for authorization...")
			response, err := waitForAuthorization(loginCtx, api, grant)
			if err != nil {
				return err
			}
			if err = config.SaveToken(path, response.AccessToken); err != nil {
				cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), cfg.Timeout)
				defer cleanupCancel()
				if revokeErr := api.WithToken(response.AccessToken).Revoke(cleanupCtx); revokeErr != nil {
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Unable to revoke the unsaved token. Revoke it from the web token management page.")
				}
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Logged in as %s. Token saved to %s\n", response.User.Username, path)
			return err
		},
	}
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Print the authorization URL without opening a browser")
	cmd.Flags().StringVar(&name, "name", "", "Token name (default: this machine's hostname)")
	return cmd
}

func waitForAuthorization(ctx context.Context, api *client.Client, grant protocol.DeviceCodeResponse) (protocol.AccessTokenResponse, error) {
	ticker := time.NewTicker(time.Duration(grant.Interval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return protocol.AccessTokenResponse{}, fmt.Errorf("login canceled or expired; run agentmesh auth login again: %w", ctx.Err())
		case <-ticker.C:
			response, err := api.Exchange(ctx, grant.DeviceCode)
			if err == nil {
				if response.AccessToken == "" || response.User.ID == "" || response.TokenType != "Bearer" {
					return protocol.AccessTokenResponse{}, fmt.Errorf("server returned an invalid token response")
				}
				return response, nil
			}
			var apiErr *client.APIError
			if !errors.As(err, &apiErr) {
				var transportErr net.Error
				if errors.As(err, &transportErr) || ctx.Err() != nil {
					continue
				}
				return protocol.AccessTokenResponse{}, err
			}
			if apiErr.Code == "authorization_pending" || apiErr.Status >= 500 || apiErr.Status == http.StatusTooManyRequests {
				continue
			}
			switch apiErr.Code {
			case "access_denied":
				return protocol.AccessTokenResponse{}, fmt.Errorf("authorization was denied")
			case "expired_token":
				return protocol.AccessTokenResponse{}, fmt.Errorf("login expired; run agentmesh auth login again")
			default:
				return protocol.AccessTokenResponse{}, err
			}
		}
	}
}

func newAuthCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{Use: "auth", Short: "Manage CLI authentication", Args: cobra.NoArgs}
	cmd.AddCommand(newLoginCommand(cfg), newLogoutCommand(cfg))
	cmd.AddCommand(&cobra.Command{Use: "status", Short: "Verify the current token with the server", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cfg.Token == "" {
				return fmt.Errorf("not logged in; run agentmesh auth login")
			}
			user, err := client.New(cfg.Server, cfg.Timeout).WithToken(cfg.Token).Me(cmd.Context())
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Logged in as %s\n", user.Username)
			return err
		},
	})
	return cmd
}

func newLogoutCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{Use: "logout", Short: "Revoke this CLI token and remove local credentials", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := util.ResolveConfigPath(cmd)
			if err != nil {
				return err
			}
			var revokeErr error
			if cfg.Token != "" {
				revokeErr = client.New(cfg.Server, cfg.Timeout).WithToken(cfg.Token).Revoke(cmd.Context())
				var apiErr *client.APIError
				if errors.As(revokeErr, &apiErr) && apiErr.Status == http.StatusUnauthorized {
					revokeErr = nil
				}
			}
			if err = config.SaveToken(path, ""); err != nil {
				return err
			}
			if revokeErr != nil {
				return fmt.Errorf("local token removed, but remote revocation failed; revoke it from the web token management page: %w", revokeErr)
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Logged out. Local token removed.")
			return err
		},
	}
}

func openBrowser(address string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", address)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", address)
	default:
		cmd = exec.Command("xdg-open", address)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
