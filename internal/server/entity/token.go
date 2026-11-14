package entity

import "time"

type TokenRequest struct {
	Name          string `json:"name" binding:"required,notblank,max=100"`
	ExpiresInDays int    `json:"expires_in_days" binding:"required,min=1,max=365"`
}

type RenameTokenRequest struct {
	Name string `json:"name" binding:"required,notblank,max=100"`
}

type AuthorizeRequest struct {
	UserCode string `json:"user_code" binding:"required,uuid4_rfc4122"`
	Approve  *bool  `json:"approve" binding:"required"`
}

type TokenView struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type CreatedTokenResponse struct {
	Token       TokenView `json:"token"`
	AccessToken string    `json:"access_token"`
}
