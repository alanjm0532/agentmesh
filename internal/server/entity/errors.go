package entity

import "errors"

var (
	ErrInvalidRequest       = errors.New("invalid_request")
	ErrInvalidCredentials   = errors.New("invalid username or password")
	ErrPasswordTooLong      = errors.New("password must not exceed 72 bytes")
	ErrAuthorizationPending = errors.New("authorization_pending")
	ErrAccessDenied         = errors.New("access_denied")
	ErrExpiredToken         = errors.New("expired_token")
	ErrInvalidGrant         = errors.New("invalid_grant")
	ErrInvalidToken         = errors.New("invalid_token")
)
