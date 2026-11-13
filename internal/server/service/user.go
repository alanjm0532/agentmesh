package service

import (
	"context"
	"errors"

	"github.com/alanjm0532/agentmesh/internal/server/ent"
	"github.com/alanjm0532/agentmesh/internal/server/repository"
	"github.com/alanjm0532/agentmesh/internal/server/util"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid username or password")

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

	accessToken, err := util.GenerateToken()
	if err != nil {
		return nil, err
	}
	return s.users.Create(ctx, username, string(hashedPassword), accessToken)
}

func (s *UserService) Login(ctx context.Context, username, password string) (*ent.User, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if ent.IsNotFound(err) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	// Verify the submitted password against the persisted bcrypt hash.
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (s *UserService) Profile(ctx context.Context, id uuid.UUID) (*ent.User, error) {
	return s.users.GetByID(ctx, id)
}
