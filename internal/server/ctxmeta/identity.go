// Package ctxmeta carries request metadata independently of the HTTP framework.
package ctxmeta

import (
	"context"

	"github.com/google/uuid"
)

type identityKey struct{}

// Identity contains verified identity metadata. Session authentication only sets UserID.
type Identity struct {
	UserID   uuid.UUID
	TokenID  uuid.UUID
	Username string
}

func WithIdentity(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, identityKey{}, identity)
}

func IdentityFrom(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(identityKey{}).(Identity)
	return identity, ok && identity.UserID != uuid.Nil
}
