package client

import (
	"context"
	"net/http"

	"github.com/alanjm0532/agentmesh/internal/protocol"
)

func (c *Client) StartLogin(ctx context.Context, name string) (protocol.DeviceCodeResponse, error) {
	return request[protocol.DeviceCodeResponse](ctx, c, http.MethodPost, protocol.DeviceCodePath, protocol.DeviceCodeRequest{Name: name})
}

func (c *Client) Exchange(ctx context.Context, device string) (protocol.AccessTokenResponse, error) {
	return request[protocol.AccessTokenResponse](ctx, c, http.MethodPost, protocol.DeviceTokenPath, protocol.DeviceTokenRequest{DeviceCode: device})
}

func (c *Client) Me(ctx context.Context) (protocol.IdentityResponse, error) {
	return doRequest[protocol.IdentityResponse](ctx, c, http.MethodGet, protocol.MePath)
}

func (c *Client) Revoke(ctx context.Context) error {
	_, err := request[map[string]string](ctx, c, http.MethodPost, protocol.RevokeTokenPath, struct{}{})
	return err
}
