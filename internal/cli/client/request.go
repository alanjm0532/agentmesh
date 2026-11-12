package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func doRequest[T any](ctx context.Context, c *Client, method, path string) (T, error) {
	var zero T
	endpoint := c.http.BaseURL + path
	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		Execute(method, path)
	if err != nil {
		return zero, fmt.Errorf("%s %s 请求失败: %w", method, endpoint, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return zero, fmt.Errorf("%s %s 请求失败: HTTP %s", method, endpoint, resp.Status())
	}
	var response T
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		return zero, fmt.Errorf("解析 %s %s 响应失败: %w", method, endpoint, err)
	}
	return response, nil
}
