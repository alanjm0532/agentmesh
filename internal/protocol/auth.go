package protocol

import "time"

const (
	DeviceCodePath  = "/auth/device/code"
	DeviceTokenPath = "/auth/device/token"
	MePath          = "/me"
	RevokeTokenPath = "/auth/token/revoke"
)

type DeviceCodeRequest struct {
	// Name is the human-readable label for the CLI authorization request (1-100 characters).
	Name string `json:"name" binding:"required,notblank,max=100"`
}

type DeviceCodeResponse struct {
	// DeviceCode is the secret polling credential; it must not be shown to other users.
	DeviceCode string `json:"device_code"`
	// UserCode identifies the login request approved through the browser.
	UserCode string `json:"user_code"`
	// VerificationURI is the server-relative browser authorization path.
	VerificationURI string `json:"verification_uri"`
	// VerificationURIComplete includes the user code in the authorization path.
	VerificationURIComplete string `json:"verification_uri_complete"`
	// ExpiresIn is the remaining authorization-request lifetime in seconds.
	ExpiresIn int `json:"expires_in"`
	// Interval is the minimum delay between device-token polling requests, in seconds.
	Interval int `json:"interval"`
}

type DeviceTokenRequest struct {
	// DeviceCode is the UUIDv4 polling credential returned by the device-code endpoint.
	DeviceCode string `json:"device_code" binding:"required,uuid4_rfc4122"`
}

type IdentityResponse struct {
	// ID is the stable UUID of the authenticated user.
	ID string `json:"id"`
	// Username is the user's account name and private-message address.
	Username string `json:"username"`
}

type AccessTokenResponse struct {
	// AccessToken is the secret credential sent in the Authorization header.
	AccessToken string `json:"access_token"`
	// TokenType is the authorization scheme, currently Bearer.
	TokenType string `json:"token_type"`
	// ExpiresIn is the token lifetime in seconds at the time of issuance.
	ExpiresIn int `json:"expires_in"`
	// ExpiresAt is the token's absolute expiration time in UTC.
	ExpiresAt time.Time `json:"expires_at"`
	// User is the identity to which the access token belongs.
	User IdentityResponse `json:"user"`
}

type ErrorResponse struct {
	// Error is the machine-readable error code.
	Error string `json:"error"`
	// Message describes the failure for the caller.
	Message string `json:"message"`
}
