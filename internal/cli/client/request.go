package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alanjm0532/agentmesh/internal/protocol"
)

func doRequest[T any](ctx context.Context, c *Client, method, path string) (T, error) {
	var zero T
	path = protocol.APIPrefix + path
	endpoint := c.http.BaseURL + path
	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		Execute(method, path)
	if err != nil {
		return zero, fmt.Errorf("%s %s request failed: %w", method, endpoint, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return zero, fmt.Errorf("%s %s request failed: HTTP %s", method, endpoint, resp.Status())
	}
	var response T
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		return zero, fmt.Errorf("decode %s %s response: %w", method, endpoint, err)
	}
	return response, nil
}
