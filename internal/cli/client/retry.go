package client

import (
	"errors"
	"net"
	"net/http"
)

// Retryable reports whether a request failed with a network or transient HTTP error.
func Retryable(err error) bool {
	if api, ok := errors.AsType[*APIError](err); ok {
		return api.Status == http.StatusRequestTimeout || api.Status == http.StatusTooManyRequests || api.Status >= 500
	}
	_, ok := errors.AsType[net.Error](err)
	return ok
}
