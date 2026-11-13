package entity

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrPasswordTooLong    = errors.New("password must not exceed 72 bytes")
)
