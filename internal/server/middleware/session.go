package middleware

import (
	"net/http"

	"github.com/alanjm0532/agentmesh/internal/server/ctxmeta"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SessionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only a user ID from the signed session establishes the Web identity.
		value, ok := sessions.Default(c).Get("user").(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not logged in"})
			return
		}
		id, err := uuid.Parse(value)
		if err != nil || id == uuid.Nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid session"})
			return
		}
		ctx := ctxmeta.WithIdentity(c.Request.Context(), ctxmeta.Identity{UserID: id})
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
