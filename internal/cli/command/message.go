package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"time"

	"github.com/alanjm0532/agentmesh/internal/cli/client"
	"github.com/alanjm0532/agentmesh/internal/cli/config"
	"github.com/alanjm0532/agentmesh/internal/cli/util"
	"github.com/alanjm0532/agentmesh/internal/protocol"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func init() { registerCommand(newMessageCommand) }

func newMessageCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{Use: "message", Short: "Send and receive private messages"}
	cmd.AddCommand(newSendMessageCommand(cfg), newReceiveMessageCommand(cfg), newMessageHistoryCommand(cfg))
	return cmd
}

func newSendMessageCommand(cfg *config.Config) *cobra.Command {
	var file, format string
	var stdin bool
	cmd := &cobra.Command{
		Use: "send <user> [text]", Short: "Send UTF-8 text to a user (maximum 128 KiB)", Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := messageOptions(cfg, format, "text", "json"); err != nil {
				return err
			}
			sources := 0
			if len(args) == 2 {
				sources++
			}
			if cmd.Flags().Changed("file") {
				sources++
			}
			if stdin {
				sources++
			}
			if sources != 1 {
				return fmt.Errorf("provide exactly one of text, --file, or --stdin")
			}
			var body string
			if len(args) == 2 {
				body = args[1]
			} else {
				reader := cmd.InOrStdin()
				if cmd.Flags().Changed("file") {
					f, err := os.Open(file)
					if err != nil {
						return err
					}
					defer func() { _ = f.Close() }()
					reader = f
				}
				data, err := io.ReadAll(io.LimitReader(reader, protocol.MaxMessageBytes+1))
				if err != nil {
					return err
				}
				body = string(data)
			}
			if !protocol.ValidMessageBody(body) {
				return fmt.Errorf("body must be valid UTF-8 without NUL and contain 1 to %d bytes", protocol.MaxMessageBytes)
			}
			id, err := uuid.NewRandom()
			if err != nil {
				return err
			}
			input := protocol.SendMessageRequest{Recipient: args[0], ClientMessageID: id.String(), EncryptionAlgorithm: protocol.EncryptionAlgorithmNone, ContentType: "text/plain", Body: body}
			api := client.New(cfg.Server, cfg.Timeout).WithToken(cfg.Token)
			var result protocol.Message
			err = util.Retry(cmd.Context(), 3, client.Retryable, func() error {
				var sendErr error
				result, sendErr = api.SendMessage(cmd.Context(), input)
				return sendErr
			})
			if err != nil {
				return err
			}
			if format == "json" {
				err = json.NewEncoder(cmd.OutOrStdout()).Encode(result)
			} else {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Sent %s to %s\n", result.ID, args[0])
			}
			if err != nil {
				return err
			}
			notifyFeishu(cmd, cfg, args[0], result)
			return nil
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "Read the message body from a file")
	cmd.Flags().BoolVar(&stdin, "stdin", false, "Read the message body from standard input")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text or json")
	cmd.Flags().Bool("feishu", false, "Override the configured Feishu notification setting (--feishu or --feishu=false)")
	return cmd
}

func newReceiveMessageCommand(cfg *config.Config) *cobra.Command {
	var follow bool
	var format, output string
	var limit, durationMinutes int
	cmd := &cobra.Command{
		Use: "recv [user]", Short: "Receive unread messages and mark successfully output messages as read", Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (resultErr error) {
			if err := messageOptions(cfg, format, "text", "jsonl"); err != nil {
				return err
			}
			if err := validMessageLimit(limit); err != nil {
				return err
			}
			if cmd.Flags().Changed("duration") && !follow {
				return fmt.Errorf("--duration requires --follow")
			}
			if durationMinutes < 1 || int64(durationMinutes) > int64((1<<63-1)/time.Minute) {
				return fmt.Errorf("duration must be a positive number of minutes within the supported time range")
			}
			ctx := cmd.Context()
			if follow {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, time.Duration(durationMinutes)*time.Minute)
				defer cancel()
			}
			writer := cmd.OutOrStdout()
			var outputFile *os.File
			if cmd.Flags().Changed("output") {
				if output == "" {
					return fmt.Errorf("output path must not be empty")
				}
				var err error
				outputFile, err = os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
				if err != nil {
					return fmt.Errorf("open message output file: %w", err)
				}
				defer func() {
					if err := outputFile.Close(); err != nil && resultErr == nil {
						resultErr = fmt.Errorf("close message output file: %w", err)
					}
				}()
				writer = outputFile
			}
			from := ""
			if len(args) == 1 {
				from = args[0]
			}
			api := client.New(cfg.Server, cfg.Timeout).WithToken(cfg.Token)
			backoff := 200 * time.Millisecond
			for {
				if follow && ctx.Err() != nil {
					return nil
				}
				page, err := api.UnreadMessages(ctx, from, limit)
				if err != nil {
					if ctx.Err() != nil {
						return nil
					}
					if !follow || !client.Retryable(err) {
						return err
					}
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Receive failed; retrying: %v\n", err)
					if pauseMessage(ctx, backoff) != nil {
						return nil
					}
					backoff = min(backoff*2, 5*time.Second)
					continue
				}
				backoff = 200 * time.Millisecond
				for _, msg := range page.Messages {
					if err = outputMessage(writer, msg, format); err != nil {
						return err
					}
				}
				if outputFile != nil && len(page.Messages) > 0 {
					if err = outputFile.Sync(); err != nil {
						return fmt.Errorf("sync message output file: %w", err)
					}
				}
				if len(page.ReadPositions) > 0 {
					if err = util.Retry(ctx, 3, client.Retryable, func() error { return api.ReadMessages(ctx, page.ReadPositions) }); err != nil {
						if follow && ctx.Err() != nil {
							return nil
						}
						return err
					}
				}
				if !follow {
					if page.HasMore {
						_, err = fmt.Fprintln(cmd.ErrOrStderr(), "More unread messages are available; run recv again or use --follow")
						return err
					}
					return nil
				}
				if !page.HasMore && pauseMessage(ctx, 5*time.Second) != nil {
					return nil
				}
			}
		},
	}
	cmd.Flags().BoolVar(&follow, "follow", false, "Poll for new messages every 5 seconds and reconnect automatically")
	cmd.Flags().IntVar(&durationMinutes, "duration", 5, "Maximum follow duration in minutes (requires --follow)")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum messages per batch (1-200)")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text or jsonl")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Append received messages to a file instead of stdout")
	return cmd
}

func newMessageHistoryCommand(cfg *config.Config) *cobra.Command {
	var before, format string
	var limit int
	cmd := &cobra.Command{
		Use: "history <user>", Short: "View private message history without changing read positions", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := messageOptions(cfg, format, "text", "json"); err != nil {
				return err
			}
			if err := validMessageLimit(limit); err != nil {
				return err
			}
			api := client.New(cfg.Server, cfg.Timeout).WithToken(cfg.Token)
			page, err := api.MessageHistory(cmd.Context(), args[0], before, limit)
			if err != nil {
				return err
			}
			if format == "json" {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(page)
			}
			for _, v := range slices.Backward(page.Messages) {
				if err = outputMessage(cmd.OutOrStdout(), v, "text"); err != nil {
					return err
				}
			}
			if page.HasMore {
				_, err = fmt.Fprintf(cmd.ErrOrStderr(), "More history: use --before %s\n", page.NextCursor)
			}
			return err
		},
	}
	cmd.Flags().StringVar(&before, "before", "", "Cursor returned by the previous history page")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum messages per page (1-200)")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text or json")
	return cmd
}

func outputMessage(w io.Writer, msg protocol.Message, format string) error {
	// Normalize names from newer servers so unrecognized algorithms are unknown too.
	if msg.EncryptionAlgorithm.String() == protocol.EncryptionAlgorithmUnknown.String() {
		return fmt.Errorf("unsupported encryption algorithm %q", string(msg.EncryptionAlgorithm))
	}
	if format == "jsonl" {
		return json.NewEncoder(w).Encode(msg)
	}
	_, err := fmt.Fprintf(w, "[%s] %s: %s\n", msg.CreatedAt.Format(time.RFC3339), msg.SenderUsername, msg.Body)
	return err
}

func messageOptions(cfg *config.Config, format string, allowed ...string) error {
	if cfg.Token == "" {
		return fmt.Errorf("not logged in; run agentmesh auth login")
	}
	if slices.Contains(allowed, format) {
		return nil
	}
	return fmt.Errorf("unsupported output format %q (allowed: %v)", format, allowed)
}

func validMessageLimit(limit int) error {
	if limit < 1 || limit > protocol.MaxMessagePageSize {
		return fmt.Errorf("limit must be between 1 and %d", protocol.MaxMessagePageSize)
	}
	return nil
}

func pauseMessage(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
