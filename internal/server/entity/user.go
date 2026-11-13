package entity

import "time"

type LoginRequest struct {
	Username string `json:"username" binding:"required,notblank,max=255"`
	Password string `json:"password" binding:"required,maxbytes=72"`
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
