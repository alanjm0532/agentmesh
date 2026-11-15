package middleware

import (
	"context"
	"time"

	"github.com/alanjm0532/agentmesh/internal/server/util"
	"github.com/gin-gonic/gin"
)

// HandlerFunc separates request lifetime and metadata from HTTP operations.
type HandlerFunc func(ctx context.Context, c *gin.Context)

// Adapt reads the latest context after preceding middleware has enriched it.
func Adapt(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if util.HandleContextError(c, ctx.Err()) {
			return
		}
		h(ctx, c)
	}
}

// Timeout shares one deadline across authentication and downstream operations.
// Cancellation is cooperative; the handler runs synchronously on Gin's goroutine.
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		if util.HandleContextError(c, ctx.Err()) {
			return
		}
		c.Next()
		if !c.Writer.Written() {
			util.HandleContextError(c, ctx.Err())
		}
	}
}
