package util

import (
	"context"
	"errors"
	"net/http"

	"github.com/alanjm0532/agentmesh/internal/protocol"
	"github.com/alanjm0532/agentmesh/internal/server/entity"
	"github.com/gin-gonic/gin"
)

func InvalidRequest(c *gin.Context) {
	c.JSON(http.StatusBadRequest, protocol.ErrorResponse{Error: entity.ErrInvalidRequest.Error(), Message: "Invalid request parameters"})
}

// HandleContextError maps a supplied cancellation or deadline error to an HTTP response.
// It never overwrites a committed response. Status 499 identifies a canceled request
// in access logs without writing a response body to a disconnected client.
func HandleContextError(c *gin.Context, err error) bool {
	deadline := errors.Is(err, context.DeadlineExceeded)
	if !deadline && !errors.Is(err, context.Canceled) {
		return false
	}
	c.Abort()
	if !c.Writer.Written() {
		if deadline {
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, protocol.ErrorResponse{
				Error: "request_timeout", Message: "Request timed out",
			})
		} else {
			c.Status(499)
		}
	}
	return true
}
