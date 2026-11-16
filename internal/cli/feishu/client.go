package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/alanjm0532/agentmesh/internal/cli/config"
	"github.com/alanjm0532/agentmesh/internal/cli/repository"
	"github.com/go-resty/resty/v2"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type Client struct {
	config config.FeishuConfig
	store  *repository.FeishuRepository
	http   *resty.Client
	im     *larkim.V1
}

func New(cfg config.FeishuConfig, store *repository.FeishuRepository, timeout time.Duration) *Client {
	httpClient := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	// 只初始化消息 SDK；显式传入用户令牌，关闭应用令牌的自动获取。
	sdkConfig := &larkcore.Config{
		BaseUrl: "https://open.feishu.cn", AppId: cfg.AppID, AppSecret: cfg.AppSecret,
		AppType: larkcore.AppTypeSelfBuilt, HttpClient: httpClient,
		Logger: sdkLogger{}, Serializable: &larkcore.DefaultSerialization{},
	}
	return &Client{config: cfg, store: store, http: resty.NewWithClient(httpClient), im: larkim.New(sdkConfig)}
}

func (c *Client) Send(ctx context.Context, username, messageID string) error {
	email, err := c.store.Email(ctx, username)
	if err != nil {
		return err
	}
	accessToken, err := c.store.AccessToken(ctx, c.refresh)
	if err != nil {
		return err
	}
	// 飞书只发送配置中的提醒文案，消息 ID 仅用于请求去重。
	content, err := json.Marshal(struct {
		Text string `json:"text"`
	}{Text: c.config.NotificationText})
	if err != nil {
		return err
	}
	request := larkim.NewCreateMessageReqBuilder().ReceiveIdType("email").
		Body(larkim.NewCreateMessageReqBodyBuilder().ReceiveId(email).MsgType("text").
			Content(string(content)).Uuid(messageID).Build()).Build()
	response, err := c.im.Message.Create(ctx, request, larkcore.WithUserAccessToken(accessToken))
	if err != nil {
		return fmt.Errorf("send Feishu message: %w", err)
	}
	if response.StatusCode != http.StatusOK || !response.Success() {
		return fmt.Errorf("feishu send failed (HTTP %d, code %d): %s; request ID: %s", response.StatusCode, response.Code, response.Msg, response.RequestId())
	}
	return nil
}

type sdkLogger struct{}

func (sdkLogger) Debug(ctx context.Context, args ...interface{}) {
	slog.DebugContext(ctx, fmt.Sprint(args...))
}

func (sdkLogger) Info(ctx context.Context, args ...interface{}) {
	slog.InfoContext(ctx, fmt.Sprint(args...))
}

func (sdkLogger) Warn(ctx context.Context, args ...interface{}) {
	slog.WarnContext(ctx, fmt.Sprint(args...))
}

func (sdkLogger) Error(ctx context.Context, args ...interface{}) {
	slog.ErrorContext(ctx, fmt.Sprint(args...))
}
