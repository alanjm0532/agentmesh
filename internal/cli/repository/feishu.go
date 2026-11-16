package repository

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/alanjm0532/agentmesh/internal/cli/ent"
	"github.com/alanjm0532/agentmesh/internal/cli/ent/feishuaccountmapping"
	"github.com/alanjm0532/agentmesh/internal/cli/ent/feishuoauthtoken"
)

type FeishuRepository struct {
	client  *ent.Client
	profile string
	server  string
	appID   string
}

type FeishuToken struct {
	AccessToken      string
	RefreshToken     string
	ExpiresAt        int64
	RefreshExpiresAt int64
	Scope            string
}

func NewFeishuRepository(client *ent.Client, profile, server, appID string) *FeishuRepository {
	return &FeishuRepository{client: client, profile: profile, server: strings.TrimRight(server, "/"), appID: appID}
}

func (r *FeishuRepository) Bind(ctx context.Context, username, email string) error {
	address, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid Feishu email: %w", err)
	}
	if address.Address != email {
		return fmt.Errorf("use a plain Feishu email address without a display name")
	}
	return r.client.FeishuAccountMapping.Create().SetServer(r.server).SetAppID(r.appID).
		SetUsername(username).SetEmail(email).SetUpdatedAt(time.Now().UTC()).
		OnConflictColumns(feishuaccountmapping.FieldServer, feishuaccountmapping.FieldAppID, feishuaccountmapping.FieldUsername).
		UpdateNewValues().Exec(ctx)
}

func (r *FeishuRepository) Unbind(ctx context.Context, username string) error {
	_, err := r.client.FeishuAccountMapping.Delete().Where(
		feishuaccountmapping.ServerEQ(r.server), feishuaccountmapping.AppIDEQ(r.appID), feishuaccountmapping.UsernameEQ(username),
	).Exec(ctx)
	return err
}

func (r *FeishuRepository) Bindings(ctx context.Context) ([]*ent.FeishuAccountMapping, error) {
	return r.client.FeishuAccountMapping.Query().Where(
		feishuaccountmapping.ServerEQ(r.server), feishuaccountmapping.AppIDEQ(r.appID),
	).Order(ent.Asc(feishuaccountmapping.FieldUsername)).All(ctx)
}

func (r *FeishuRepository) Email(ctx context.Context, username string) (string, error) {
	mapping, err := r.client.FeishuAccountMapping.Query().Where(
		feishuaccountmapping.ServerEQ(r.server), feishuaccountmapping.AppIDEQ(r.appID), feishuaccountmapping.UsernameEQ(username),
	).Only(ctx)
	if ent.IsNotFound(err) {
		return "", fmt.Errorf("no Feishu email mapping for %q; run agentmesh feishu bind <user> <email>", username)
	}
	if err != nil {
		return "", err
	}
	return mapping.Email, nil
}

func (r *FeishuRepository) SaveToken(ctx context.Context, value FeishuToken) error {
	return r.saveToken(ctx, r.client, value)
}

func (r *FeishuRepository) saveToken(ctx context.Context, client *ent.Client, value FeishuToken) error {
	return client.FeishuOAuthToken.Create().SetProfile(r.profile).SetAppID(r.appID).
		SetAccessToken(value.AccessToken).SetRefreshToken(value.RefreshToken).
		SetExpiresAt(value.ExpiresAt).SetRefreshExpiresAt(value.RefreshExpiresAt).SetScope(value.Scope).
		OnConflictColumns(feishuoauthtoken.FieldProfile, feishuoauthtoken.FieldAppID).UpdateNewValues().Exec(ctx)
}

func (r *FeishuRepository) Logout(ctx context.Context) error {
	_, err := r.client.FeishuOAuthToken.Delete().Where(feishuoauthtoken.ProfileEQ(r.profile), feishuoauthtoken.AppIDEQ(r.appID)).Exec(ctx)
	return err
}

func (r *FeishuRepository) AccessToken(ctx context.Context, refresh func(context.Context, string) (FeishuToken, error)) (string, error) {
	// 事务独立于发送的取消信号，确保飞书轮换令牌后仍能提交新凭证。
	transactionCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	defer cancel()
	tx, err := r.client.Tx(transactionCtx)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	// 通用 SQLite 驱动采用 immediate 事务，多进程刷新时先锁定再读取令牌。
	value, err := tx.FeishuOAuthToken.Query().Where(feishuoauthtoken.ProfileEQ(r.profile), feishuoauthtoken.AppIDEQ(r.appID)).Only(ctx)
	if ent.IsNotFound(err) {
		return "", fmt.Errorf("Feishu is not authorized; run agentmesh feishu login")
	}
	if err != nil {
		return "", err
	}
	accessToken := value.AccessToken
	if time.Now().Add(time.Minute).Unix() >= value.ExpiresAt {
		if time.Now().Unix() >= value.RefreshExpiresAt {
			return "", fmt.Errorf("Feishu authorization expired; run agentmesh feishu login")
		}
		refreshed, err := refresh(ctx, value.RefreshToken)
		if err != nil {
			return "", err
		}
		persistCtx, stop := context.WithTimeout(transactionCtx, 5*time.Second)
		defer stop()
		if err = r.saveToken(persistCtx, tx.Client(), refreshed); err != nil {
			return "", fmt.Errorf("save refreshed Feishu credentials; reauthorize if this persists: %w", err)
		}
		accessToken = refreshed.AccessToken
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return accessToken, nil
}
