package service_test

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alanjm0532/agentmesh/internal/server/ent/user"
	"github.com/alanjm0532/agentmesh/internal/server/service"
	"github.com/alanjm0532/agentmesh/internal/testutil"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestUserServiceLogin(t *testing.T) {
	t.Run("stores hashed password and user fields", func(t *testing.T) {
		deps := testutil.New(t)
		client, users := deps.Client, deps.UserService
		ctx := context.Background()
		username := "login-user-" + uuid.NewString()
		password := "login-password"
		startedAt := time.Now()

		created, err := users.Login(ctx, username, password)
		if err != nil {
			t.Fatalf("log in with new user: %v", err)
		}
		finishedAt := time.Now()

		// Read the persisted row to verify what was actually stored.
		stored, err := client.User.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("read created user: %v", err)
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
		if stored.AccessToken != created.AccessToken {
			t.Error("returned access token differs from the stored token")
		}
		// SQL databases may truncate fractional seconds when persisting timestamps.
		if stored.CreatedAt.Before(startedAt.Truncate(time.Second)) || stored.CreatedAt.After(finishedAt) {
			t.Error("created_at is outside the account creation time range")
		}
		if stored.UpdatedAt.Before(stored.CreatedAt) || stored.UpdatedAt.After(finishedAt) {
			t.Error("updated_at is outside the account creation time range")
		}

		existing, err := users.Login(ctx, username, password)
		if err != nil {
			t.Fatalf("log in with existing user: %v", err)
		}
		if existing.ID != stored.ID || existing.AccessToken != stored.AccessToken || existing.Password != stored.Password || !existing.CreatedAt.Equal(stored.CreatedAt) || !existing.UpdatedAt.Equal(stored.UpdatedAt) {
			t.Error("repeat login changed the existing account")
		}

		// Identical passwords must get independent salts and access tokens.
		other, err := users.Login(ctx, "another-"+username, password)
		if err != nil {
			t.Fatalf("log in with another new user: %v", err)
		}
		if other.Password == stored.Password {
			t.Error("users with the same password have identical hashes")
		}
		if other.AccessToken == stored.AccessToken {
			t.Error("users have identical access tokens")
		}
	})

	t.Run("rejects incorrect password for existing user", func(t *testing.T) {
		deps := testutil.New(t)
		users := deps.UserService
		ctx := context.Background()
		username := "existing-user-" + uuid.NewString()
		_, err := users.Login(ctx, username, "original-password")
		if err != nil {
			t.Fatalf("create original user: %v", err)
		}

		_, err = users.Login(ctx, username, "different-password")
		if !errors.Is(err, service.ErrInvalidCredentials) {
			t.Fatalf("login error = %v, want ErrInvalidCredentials", err)
		}
		if _, err := users.Login(ctx, username, "original-password"); err != nil {
			t.Fatalf("log in with original password after failed login: %v", err)
		}
	})

	t.Run("rejects passwords exceeding bcrypt limit", func(t *testing.T) {
		deps := testutil.New(t)
		client, users := deps.Client, deps.UserService
		ctx := context.Background()
		username := "long-password-user-" + uuid.NewString()

		_, err := users.Login(ctx, username, strings.Repeat("x", 73))
		if !errors.Is(err, bcrypt.ErrPasswordTooLong) {
			t.Fatalf("account creation error = %v, want bcrypt.ErrPasswordTooLong", err)
		}
		count, err := client.User.Query().Where(user.UsernameEQ(username)).Count(ctx)
		if err != nil {
			t.Fatalf("count users: %v", err)
		}
		if count != 0 {
			t.Errorf("user count = %d, want 0 after failed account creation", count)
		}
	})

	t.Run("concurrent first logins verify the winning password", func(t *testing.T) {
		deps := testutil.New(t)
		ctx := context.Background()
		username := "concurrent-login-user-" + uuid.NewString()
		type result struct {
			password string
			id       string
			err      error
		}
		start := make(chan struct{})
		results := make(chan result, 4)
		for _, password := range []string{"password-one", "password-two", "password-one", "password-two"} {
			go func() {
				<-start
				loggedIn, err := deps.UserService.Login(ctx, username, password)
				r := result{password: password, err: err}
				if loggedIn != nil {
					r.id = loggedIn.ID.String()
				}
				results <- r
			}()
		}
		close(start)
		var attempts []result
		for range 4 {
			attempts = append(attempts, <-results)
		}
		stored, err := deps.Client.User.Query().Where(user.UsernameEQ(username)).Only(ctx)
		if err != nil {
			t.Fatalf("read unique account after concurrent logins: %v", err)
		}
		for _, attempt := range attempts {
			if bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte(attempt.password)) == nil {
				if attempt.err != nil || attempt.id != stored.ID.String() {
					t.Errorf("matching password: id = %q, error = %v", attempt.id, attempt.err)
				}
			} else if !errors.Is(attempt.err, service.ErrInvalidCredentials) || attempt.id != "" {
				t.Errorf("incorrect password: id = %q, error = %v", attempt.id, attempt.err)
			}
		}
	})
}
