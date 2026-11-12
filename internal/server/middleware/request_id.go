package middleware

import (
	"crypto/rand"

	"github.com/gin-gonic/gin"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate a request ID and return it in the response header.
		c.Header("X-Request-ID", rand.Text())
		c.Next()
	}
}
