package middleware

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// WebCSRF requires a custom header that cross-origin forms cannot supply.
// Cross-origin scripts cannot supply it without a CORS preflight, which is not enabled.
func WebCSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
			c.Next()
			return
		}
		valid := c.GetHeader("X-AgentMesh-CSRF") == "1" && c.GetHeader("Sec-Fetch-Site") != "cross-site"
		if origin := c.GetHeader("Origin"); origin != "" {
			parsed, err := url.Parse(origin)
			valid = valid && err == nil && (parsed.Scheme == "https" || parsed.Scheme == "http") && parsed.Host == c.Request.Host && parsed.User == nil
		}
		if !valid {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Invalid request origin"})
			return
		}
		c.Next()
	}
}
