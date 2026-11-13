package service_test

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alanjm0532/agentmesh/internal/server/ent"
	"github.com/alanjm0532/agentmesh/internal/server/ent/user"
	"github.com/alanjm0532/agentmesh/internal/testutil"
	"golang.org/x/crypto/bcrypt"
)

func TestUserServiceRegister(t *testing.T) {
	t.Run("stores hashed password and user fields", func(t *testing.T) {
		deps := testutil.New(t)
		client, users := deps.Client, deps.UserService
		ctx := context.Background()
		username := "registration-user"
		password := "registration-password"
		startedAt := time.Now()

		registered, err := users.Register(ctx, username, password)
		if err != nil {
			t.Fatalf("register user: %v", err)
		}
		finishedAt := time.Now()

		// Read the persisted row to verify what was actually stored.
		stored, err := client.User.Get(ctx, registered.ID)
		if err != nil {
			t.Fatalf("read registered user: %v", err)
		}
		if stored.Username != username {
			t.Errorf("username = %q, want %q", stored.Username, username)
		}
		if stored.ID.Version() != 7 {
			t.Errorf("UUID version = %d, want 7", stored.ID.Version())
		}
		if err = bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte(password)); err != nil {
			t.Errorf("stored password does not match: %v", err)
		}
		cost, err := bcrypt.Cost([]byte(stored.Password))
		if err != nil {
			t.Fatalf("read bcrypt cost: %v", err)
		}
		if cost != bcrypt.DefaultCost {
			t.Errorf("bcrypt cost = %d, want %d", cost, bcrypt.DefaultCost)
		}
		tokenValue, hasPrefix := strings.CutPrefix(stored.AccessToken, "agentmesh_")
		if !hasPrefix {
			t.Fatal("access token must start with agentmesh_")
		}
		token, err := hex.DecodeString(tokenValue)
		if err != nil {
			t.Fatalf("decode access token: %v", err)
		}
		if len(token) != 32 {
			t.Errorf("access token length = %d bytes, want 32", len(token))
		}
		if stored.AccessToken != registered.AccessToken {
			t.Error("returned access token differs from the stored token")
		}
		// SQL databases may truncate fractional seconds when persisting timestamps.
		if stored.CreatedAt.Before(startedAt.Truncate(time.Second)) || stored.CreatedAt.After(finishedAt) {
			t.Error("created_at is outside the registration time range")
		}
		if stored.UpdatedAt.Before(stored.CreatedAt) || stored.UpdatedAt.After(finishedAt) {
			t.Error("updated_at is outside the registration time range")
		}

		// Identical passwords must get independent salts and access tokens.
		other, err := users.Register(ctx, "another-user", password)
		if err != nil {
			t.Fatalf("register another user: %v", err)
		}
		if other.Password == stored.Password {
			t.Error("users with the same password have identical hashes")
		}
		if other.AccessToken == stored.AccessToken {
			t.Error("users have identical access tokens")
		}
	})

	t.Run("rejects duplicate username", func(t *testing.T) {
		deps := testutil.New(t)
		users := deps.UserService
		ctx := context.Background()
		username := "duplicate-user"
		_, err := users.Register(ctx, username, "original-password")
		if err != nil {
			t.Fatalf("register original user: %v", err)
		}

		_, err = users.Register(ctx, username, "different-password")
		if !ent.IsConstraintError(err) {
			t.Fatalf("duplicate registration error = %v, want constraint error", err)
		}
	})

	t.Run("rejects passwords exceeding bcrypt limit", func(t *testing.T) {
		deps := testutil.New(t)
		client, users := deps.Client, deps.UserService
		ctx := context.Background()
		username := "long-password-user"

		_, err := users.Register(ctx, username, strings.Repeat("x", 73))
		if !errors.Is(err, bcrypt.ErrPasswordTooLong) {
			t.Fatalf("registration error = %v, want bcrypt.ErrPasswordTooLong", err)
		}
		count, err := client.User.Query().Where(user.UsernameEQ(username)).Count(ctx)
		if err != nil {
			t.Fatalf("count users: %v", err)
		}
		if count != 0 {
			t.Errorf("user count = %d, want 0 after failed registration", count)
		}
	})
}
