package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/alanjm0532/agentmesh/internal/protocol"
	"github.com/alanjm0532/agentmesh/internal/server/entity"
	"github.com/alanjm0532/agentmesh/internal/server/service"
	"github.com/gin-gonic/gin"
)

func TokenAuth(tokens *service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		fields := strings.Fields(c.GetHeader("Authorization"))
		if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
			rejectToken(c)
			return
		}
		token, user, err := tokens.Authenticate(c.Request.Context(), fields[1])
		if errors.Is(err, entity.ErrInvalidToken) {
			rejectToken(c)
			return
		}
		if err != nil {
			_ = c.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, protocol.ErrorResponse{Error: "server_error", Message: "Unable to verify token"})
			return
		}
		c.Set("userID", user.ID)
		c.Set("tokenID", token.ID)
		c.Set("username", user.Username)
		c.Next()
	}
}

func rejectToken(c *gin.Context) {
	c.Header("WWW-Authenticate", `Bearer error="invalid_token"`)
	c.AbortWithStatusJSON(http.StatusUnauthorized, protocol.ErrorResponse{Error: "invalid_token", Message: "Token is missing, invalid, expired, or revoked"})
}
