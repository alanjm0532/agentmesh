package entity

import "time"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ProfileResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}
