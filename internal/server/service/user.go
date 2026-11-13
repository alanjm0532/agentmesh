package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/alanjm0532/agentmesh/internal/server/ent"
	"github.com/alanjm0532/agentmesh/internal/server/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	users *repository.UserRepository
}

func NewUserService(users *repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) Register(ctx context.Context, username, password string) (*ent.User, error) {
	// Persist only the bcrypt hash, using the default bcrypt cost.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Generate an independent 256-bit access token for each user.
	var token [32]byte
	if _, err := rand.Read(token[:]); err != nil {
		return nil, err
	}
	return s.users.Create(ctx, username, string(hashedPassword), hex.EncodeToString(token[:]))
}
