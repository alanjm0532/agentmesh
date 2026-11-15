package middleware

import (
	"context"
	"net/http"

	"github.com/alanjm0532/agentmesh/internal/server/ctxmeta"
	"github.com/gin-gonic/gin"
)

// AuthenticatedHandlerFunc explicitly requires verified identity from authentication middleware.
type AuthenticatedHandlerFunc func(ctx context.Context, c *gin.Context, identity ctxmeta.Identity)

// AdaptAuthenticated checks and injects identity before invoking a protected handler.
func AdaptAuthenticated(h AuthenticatedHandlerFunc) gin.HandlerFunc {
	return Adapt(func(ctx context.Context, c *gin.Context) {
		identity, ok := ctxmeta.IdentityFrom(ctx)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authenticated"})
			return
		}
		h(ctx, c, identity)
	})
}
