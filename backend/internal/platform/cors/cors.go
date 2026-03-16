package cors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var allowedMethods = "GET, POST, OPTIONS"
var allowedHeaders = "Content-Type, X-CSRF-Token"

// Middleware allows credentialed requests from an explicit origin allowlist.
func Middleware(origins []string) gin.HandlerFunc {
	if len(origins) == 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	originSet := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		if origin != "" {
			originSet[origin] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		if _, ok := originSet[origin]; !ok {
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.Next()
			return
		}

		headers := c.Writer.Header()
		headers.Set("Access-Control-Allow-Origin", origin)
		headers.Set("Access-Control-Allow-Credentials", "true")
		headers.Set("Access-Control-Allow-Methods", allowedMethods)
		headers.Set("Access-Control-Allow-Headers", allowedHeaders)
		headers.Add("Vary", "Origin")
		headers.Add("Vary", "Access-Control-Request-Method")
		headers.Add("Vary", "Access-Control-Request-Headers")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
