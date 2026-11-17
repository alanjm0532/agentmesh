package middleware

import (
	"net/http"

	"github.com/alanjm0532/agentmesh/internal/server/ctxmeta"
	"github.com/gin-gonic/gin"
)

type AuthenticatedHandlerFunc func(c *gin.Context, identity ctxmeta.Identity)

func AdaptAuthenticated(h AuthenticatedHandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Identity remains on the request context even though database calls use c.
		identity, ok := ctxmeta.IdentityFrom(c.Request.Context())
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authenticated"})
			return
		}
		h(c, identity)
	}
}
