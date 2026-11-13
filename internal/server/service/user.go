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

func (s *UserService) createUser(ctx context.Context, username, password string) (*ent.User, error) {
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
	if len(password) > 72 {
		return nil, bcrypt.ErrPasswordTooLong
	}
	user, err := s.users.GetByUsername(ctx, username)
	if ent.IsNotFound(err) {
		user, err = s.createUser(ctx, username, password)
		if err == nil {
			return user, nil
		}
		if ent.IsConstraintError(err) {
			// A concurrent login may have created this username first.
			// Verify its password before allowing access to that account.
			existing, lookupErr := s.users.GetByUsername(ctx, username)
			if lookupErr == nil {
				user, err = existing, nil
			}
		}
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
