package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alanjm0532/agentmesh/internal/protocol"
)

type APIError struct {
	Status int
	Code   string
}

func (e *APIError) Error() string {
	if e.Status == http.StatusUnauthorized {
		return "authentication failed; run agentmesh login again"
	}
	if e.Status == http.StatusForbidden {
		return "permission denied"
	}
	return fmt.Sprintf("request failed: HTTP %d (%s)", e.Status, e.Code)
}

func doRequest[T any](ctx context.Context, c *Client, method, path string) (T, error) {
	return request[T](ctx, c, method, path, nil)
}

func request[T any](ctx context.Context, c *Client, method, path string, body any) (T, error) {
	var zero T
	path = protocol.APIPrefix + path
	req := c.http.R().SetContext(ctx).SetHeader("Accept", "application/json")
	if body != nil {
		req.SetBody(body)
	}
	resp, err := req.Execute(method, path)
	if err != nil {
		return zero, fmt.Errorf("%s request failed: %w", path, err)
	}
	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		var response protocol.ErrorResponse
		_ = json.Unmarshal(resp.Body(), &response)
		return zero, &APIError{Status: resp.StatusCode(), Code: response.Error}
	}
	var response T
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		return zero, fmt.Errorf("decode %s response: %w", path, err)
	}
	return response, nil
}
