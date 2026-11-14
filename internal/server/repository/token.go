package repository

import (
	"context"
	"time"

	"github.com/alanjm0532/agentmesh/internal/server/ent"
	"github.com/alanjm0532/agentmesh/internal/server/ent/usertoken"
	"github.com/alanjm0532/agentmesh/internal/server/entity"
	"github.com/google/uuid"
)

type TokenRepository struct{ client *ent.Client }

func NewTokenRepository(client *ent.Client) *TokenRepository { return &TokenRepository{client: client} }

func (r *TokenRepository) CreateLogin(ctx context.Context, name, deviceHash, userCode string, expires time.Time) (*ent.UserToken, error) {
	return r.client.UserToken.Create().SetName(name).SetDeviceCodeHash(deviceHash).SetUserCode(userCode).SetExpiresAt(expires).Save(ctx)
}

func (r *TokenRepository) ByDevice(ctx context.Context, hash string) (*ent.UserToken, error) {
	return r.client.UserToken.Query().Where(usertoken.DeviceCodeHashEQ(hash)).Only(ctx)
}

func (r *TokenRepository) ByCode(ctx context.Context, code string) (*ent.UserToken, error) {
	return r.client.UserToken.Query().Where(usertoken.UserCodeEQ(code)).Only(ctx)
}

func (r *TokenRepository) Decide(ctx context.Context, code string, userID uuid.UUID, approve bool, now time.Time) (bool, error) {
	update := r.client.UserToken.Update().Where(usertoken.UserCodeEQ(code), usertoken.StatusEQ(usertoken.StatusPending), usertoken.ExpiresAtGT(now)).SetUserID(userID)
	if approve {
		update.SetStatus(usertoken.StatusApproved)
	} else {
		update.SetStatus(usertoken.StatusDenied)
	}
	count, err := update.Save(ctx)
	return count == 1, err
}

// Exchange uses a conditional write inside a transaction so a grant can be redeemed only once.
func (r *TokenRepository) Exchange(ctx context.Context, grant *ent.UserToken, hash string, now, expires time.Time) (*ent.User, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if grant.UserID == nil {
		return nil, entity.ErrInvalidGrant
	}
	user, err := tx.User.Get(ctx, *grant.UserID)
	if ent.IsNotFound(err) {
		return nil, entity.ErrInvalidGrant
	}
	if err != nil {
		return nil, err
	}
	count, err := tx.UserToken.Update().Where(usertoken.IDEQ(grant.ID), usertoken.StatusEQ(usertoken.StatusApproved), usertoken.ExpiresAtGT(now), usertoken.UserIDEQ(user.ID)).
		SetStatus(usertoken.StatusActive).SetTokenHash(hash).SetExpiresAt(expires).ClearDeviceCodeHash().ClearUserCode().Save(ctx)
	if err != nil {
		return nil, err
	}
	if count != 1 {
		return nil, entity.ErrInvalidGrant
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return user, nil
}

func (r *TokenRepository) Create(ctx context.Context, userID uuid.UUID, name, hash string, expires time.Time) (*ent.UserToken, error) {
	return r.client.UserToken.Create().SetUserID(userID).SetName(name).SetStatus(usertoken.StatusActive).SetTokenHash(hash).SetExpiresAt(expires).Save(ctx)
}

func (r *TokenRepository) List(ctx context.Context, userID uuid.UUID) ([]*ent.UserToken, error) {
	return r.client.UserToken.Query().Where(usertoken.UserIDEQ(userID), usertoken.StatusIn(usertoken.StatusActive, usertoken.StatusRevoked)).Order(ent.Desc(usertoken.FieldCreatedAt), ent.Desc(usertoken.FieldID)).All(ctx)
}

func (r *TokenRepository) Rename(ctx context.Context, userID, id uuid.UUID, name string) (bool, error) {
	count, err := r.client.UserToken.Update().Where(usertoken.IDEQ(id), usertoken.UserIDEQ(userID), usertoken.StatusIn(usertoken.StatusActive, usertoken.StatusRevoked)).SetName(name).Save(ctx)
	if err == nil && count == 0 {
		return r.client.UserToken.Query().Where(usertoken.IDEQ(id), usertoken.UserIDEQ(userID), usertoken.StatusIn(usertoken.StatusActive, usertoken.StatusRevoked)).Exist(ctx)
	}
	return count == 1, err
}

func (r *TokenRepository) Revoke(ctx context.Context, userID, id uuid.UUID) (bool, error) {
	count, err := r.client.UserToken.Update().Where(usertoken.IDEQ(id), usertoken.UserIDEQ(userID), usertoken.StatusIn(usertoken.StatusActive, usertoken.StatusRevoked)).SetStatus(usertoken.StatusRevoked).Save(ctx)
	if err == nil && count == 0 {
		return r.client.UserToken.Query().Where(usertoken.IDEQ(id), usertoken.UserIDEQ(userID), usertoken.StatusIn(usertoken.StatusActive, usertoken.StatusRevoked)).Exist(ctx)
	}
	return count == 1, err
}

func (r *TokenRepository) Authenticate(ctx context.Context, hash string, now time.Time) (*ent.UserToken, *ent.User, error) {
	token, err := r.client.UserToken.Query().Where(usertoken.TokenHashEQ(hash), usertoken.StatusEQ(usertoken.StatusActive), usertoken.ExpiresAtGT(now)).WithUser().Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil, entity.ErrInvalidToken
	}
	if err != nil {
		return nil, nil, err
	}
	if token.Edges.User == nil {
		return nil, nil, entity.ErrInvalidToken
	}
	return token, token.Edges.User, nil
}
