package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alanjm0532/agentmesh/pkg/protocol"
)

type Client struct {
	server string
	http   *http.Client
}

func New(server string, timeout time.Duration) *Client {
	return &Client{
		server: strings.TrimRight(server, "/"),
		http:   &http.Client{Timeout: timeout},
	}
}

func (c *Client) Ping(ctx context.Context) (protocol.PingResponse, error) {
	endpoint := c.server + "/ping"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return protocol.PingResponse{}, fmt.Errorf("创建 ping 请求失败: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return protocol.PingResponse{}, fmt.Errorf("ping %s 失败: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return protocol.PingResponse{}, fmt.Errorf("ping %s 失败: HTTP %s", endpoint, resp.Status)
	}
	var response protocol.PingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return protocol.PingResponse{}, fmt.Errorf("解析 ping 响应失败: %w", err)
	}
	if response.Message != "pong" {
		return protocol.PingResponse{}, fmt.Errorf("ping 响应异常: message = %q", response.Message)
	}
	return response, nil
}
