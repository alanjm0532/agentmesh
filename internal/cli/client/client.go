package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/alanjm0532/agentmesh/pkg/protocol"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	http *resty.Client
}

func New(server string, timeout time.Duration) *Client {
	return &Client{
		http: resty.New().SetBaseURL(server).SetTimeout(timeout),
	}
}

func (c *Client) Ping(ctx context.Context) (protocol.PingResponse, error) {
	response, err := doRequest[protocol.PingResponse](ctx, c, http.MethodGet, protocol.PingPath)
	if err != nil {
		return protocol.PingResponse{}, err
	}
	if response.Message != "pong" {
		return protocol.PingResponse{}, fmt.Errorf("unexpected ping response: message = %q", response.Message)
	}
	return response, nil
}
