package service

import (
	"context"
	"encoding/hex"
	"strings"
	"time"

	"github.com/alanjm0532/agentmesh/internal/protocol"
	"github.com/alanjm0532/agentmesh/internal/server/ent"
	"github.com/alanjm0532/agentmesh/internal/server/ent/usertoken"
	"github.com/alanjm0532/agentmesh/internal/server/entity"
	"github.com/alanjm0532/agentmesh/internal/server/repository"
	"github.com/alanjm0532/agentmesh/internal/server/util"
	"github.com/google/uuid"
)

const LoginLifetime = 5 * time.Minute
const TokenLifetime = 30 * 24 * time.Hour

type TokenService struct{ tokens *repository.TokenRepository }

func NewTokenService(tokens *repository.TokenRepository) *TokenService {
	return &TokenService{tokens: tokens}
}

func (s *TokenService) StartLogin(ctx context.Context, name string) (protocol.DeviceCodeResponse, error) {
	device, err := uuid.NewRandom()
	if err != nil {
		return protocol.DeviceCodeResponse{}, err
	}
	code, err := uuid.NewRandom()
	if err != nil {
		return protocol.DeviceCodeResponse{}, err
	}
	userCode := code.String()
	deviceCode := device.String()
	_, err = s.tokens.CreateLogin(ctx, strings.TrimSpace(name), util.HashToken(deviceCode), userCode, util.NowUTC().Add(LoginLifetime))
	if err != nil {
		return protocol.DeviceCodeResponse{}, err
	}
	return protocol.DeviceCodeResponse{DeviceCode: deviceCode, UserCode: userCode, VerificationURI: "/cli-login.html", VerificationURIComplete: "/cli-login.html?user_code=" + userCode, ExpiresIn: int(LoginLifetime.Seconds()), Interval: 3}, nil
}

func (s *TokenService) LoginRequest(ctx context.Context, code string) (*ent.UserToken, error) {
	grant, err := s.tokens.ByCode(ctx, strings.ToLower(strings.TrimSpace(code)))
	if ent.IsNotFound(err) {
		return nil, entity.ErrInvalidGrant
	}
	if err != nil {
		return nil, err
	}
	if !grant.ExpiresAt.After(util.NowUTC()) {
		return nil, entity.ErrExpiredToken
	}
	if grant.Status != usertoken.StatusPending {
		return nil, entity.ErrInvalidGrant
	}
	return grant, nil
}

func (s *TokenService) Authorize(ctx context.Context, code string, userID uuid.UUID, approve bool) error {
	if _, err := s.LoginRequest(ctx, code); err != nil {
		return err
	}
	ok, err := s.tokens.Decide(ctx, strings.ToLower(strings.TrimSpace(code)), userID, approve, util.NowUTC())
	if err != nil {
		return err
	}
	if !ok {
		return entity.ErrInvalidGrant
	}
	return nil
}

func (s *TokenService) Exchange(ctx context.Context, device string) (protocol.AccessTokenResponse, error) {
	grant, err := s.tokens.ByDevice(ctx, util.HashToken(strings.ToLower(device)))
	if ent.IsNotFound(err) {
		return protocol.AccessTokenResponse{}, entity.ErrInvalidGrant
	}
	if err != nil {
		return protocol.AccessTokenResponse{}, err
	}
	now := util.NowUTC()
	if !grant.ExpiresAt.After(now) {
		return protocol.AccessTokenResponse{}, entity.ErrExpiredToken
	}
	switch grant.Status {
	case usertoken.StatusPending:
		return protocol.AccessTokenResponse{}, entity.ErrAuthorizationPending
	case usertoken.StatusDenied:
		return protocol.AccessTokenResponse{}, entity.ErrAccessDenied
	case usertoken.StatusApproved:
	default:
		return protocol.AccessTokenResponse{}, entity.ErrInvalidGrant
	}
	raw, err := util.GenerateToken()
	if err != nil {
		return protocol.AccessTokenResponse{}, err
	}
	expires := now.Add(TokenLifetime)
	user, err := s.tokens.Exchange(ctx, grant, util.HashToken(raw), now, expires)
	if err != nil {
		return protocol.AccessTokenResponse{}, err
	}
	return protocol.AccessTokenResponse{AccessToken: raw, TokenType: "Bearer", ExpiresIn: int(TokenLifetime.Seconds()), ExpiresAt: expires, User: protocol.IdentityResponse{ID: user.ID.String(), Username: user.Username}}, nil
}

func tokenView(token *ent.UserToken) entity.TokenView {
	status := string(token.Status)
	if token.Status == usertoken.StatusActive && !token.ExpiresAt.After(util.NowUTC()) {
		status = "expired"
	}
	return entity.TokenView{ID: token.ID.String(), Name: token.Name, Status: status, CreatedAt: token.CreatedAt.UTC(), ExpiresAt: token.ExpiresAt.UTC()}
}

func (s *TokenService) List(ctx context.Context, userID uuid.UUID) ([]entity.TokenView, error) {
	tokens, err := s.tokens.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]entity.TokenView, 0, len(tokens))
	for _, token := range tokens {
		result = append(result, tokenView(token))
	}
	return result, nil
}

func (s *TokenService) Rename(ctx context.Context, userID, id uuid.UUID, name string) (bool, error) {
	return s.tokens.Rename(ctx, userID, id, strings.TrimSpace(name))
}

func (s *TokenService) Revoke(ctx context.Context, userID, id uuid.UUID) (bool, error) {
	return s.tokens.Revoke(ctx, userID, id)
}

func (s *TokenService) Authenticate(ctx context.Context, raw string) (*ent.UserToken, *ent.User, error) {
	if len(raw) != len("agentmesh_")+64 || !strings.HasPrefix(raw, "agentmesh_") {
		return nil, nil, entity.ErrInvalidToken
	}
	if _, err := hex.DecodeString(strings.TrimPrefix(raw, "agentmesh_")); err != nil {
		return nil, nil, entity.ErrInvalidToken
	}
	return s.tokens.Authenticate(ctx, util.HashToken(raw), util.NowUTC())
}
