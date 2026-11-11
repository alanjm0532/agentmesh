package middleware

import (
	"crypto/rand"

	"github.com/gin-gonic/gin"
)

// RequestID 为每次请求生成标识，并通过响应头返回给调用方。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Request-ID", rand.Text())
		c.Next()
	}
}
