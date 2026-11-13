package repository

import (
	"context"

	"github.com/alanjm0532/agentmesh/internal/server/ent"
)

type UserRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) *UserRepository {
	return &UserRepository{client: client}
}

func (r *UserRepository) Create(ctx context.Context, username, hashedPassword, accessToken string) (*ent.User, error) {
	return r.client.User.Create().
		SetUsername(username).
		SetPassword(hashedPassword).
		SetAccessToken(accessToken).
		Save(ctx)
}
