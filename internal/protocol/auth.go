package protocol

import "time"

const (
	DeviceCodePath  = "/auth/device/code"
	DeviceTokenPath = "/auth/device/token"
	MePath          = "/me"
	RevokeTokenPath = "/auth/token/revoke"
)

type DeviceCodeRequest struct {
	Name string `json:"name" binding:"required,notblank,max=100"`
}

type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type DeviceTokenRequest struct {
	DeviceCode string `json:"device_code" binding:"required,uuid4_rfc4122"`
}

type IdentityResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type AccessTokenResponse struct {
	AccessToken string           `json:"access_token"`
	TokenType   string           `json:"token_type"`
	ExpiresIn   int              `json:"expires_in"`
	ExpiresAt   time.Time        `json:"expires_at"`
	User        IdentityResponse `json:"user"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
