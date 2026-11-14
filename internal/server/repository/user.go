package repository

import (
	"context"

	"github.com/alanjm0532/agentmesh/internal/server/ent"
	"github.com/alanjm0532/agentmesh/internal/server/ent/user"
	"github.com/google/uuid"
)

type UserRepository struct {
	client *ent.Client
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*ent.User, error) {
	return r.client.User.Query().Where(user.UsernameEQ(username)).Only(ctx)
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.User, error) {
	return r.client.User.Get(ctx, id)
}

func NewUserRepository(client *ent.Client) *UserRepository {
	return &UserRepository{client: client}
}

func (r *UserRepository) Create(ctx context.Context, username, hashedPassword string) (*ent.User, error) {
	return r.client.User.Create().
		SetUsername(username).
		SetPassword(hashedPassword).
		Save(ctx)
}
