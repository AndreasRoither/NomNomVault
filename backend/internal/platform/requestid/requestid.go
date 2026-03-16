package requestid

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// HeaderName is the canonical response header for request IDs.
	HeaderName = "X-Request-ID"
	contextKey = "request_id"
)

// Middleware injects a request ID into the context and response headers.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(HeaderName)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(contextKey, requestID)
		c.Writer.Header().Set(HeaderName, requestID)

		c.Next()
	}
}

// FromContext returns the request ID associated with the current request.
func FromContext(c *gin.Context) string {
	if value, ok := c.Get(contextKey); ok {
		if requestID, ok := value.(string); ok {
			return requestID
		}
	}

	return c.Writer.Header().Get(HeaderName)
}

// AttachToRequest stores the request ID in the standard request header.
func AttachToRequest(req *http.Request, requestID string) {
	if requestID != "" {
		req.Header.Set(HeaderName, requestID)
	}
}
