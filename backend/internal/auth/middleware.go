package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/httpx"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/securitylog"
)

const sessionContextKey = "auth_session"

// Middleware authenticates access cookies for protected routes.
func Middleware(signer *TokenSigner, cookies *CookieManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := cookies.ReadAccessToken(c)
		if !ok {
			securitylog.Log(c, "auth.authorization.denied", map[string]string{"reason": "missing_access_cookie"})
			httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Authentication is required.", nil)
			c.Abort()
			return
		}

		session, err := signer.ParseAccessToken(token)
		if err != nil {
			securitylog.Log(c, "auth.authorization.denied", map[string]string{"reason": "invalid_access_token"})
			httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Authentication is required.", nil)
			c.Abort()
			return
		}

		c.Set(sessionContextKey, session)
		c.Next()
	}
}

// CSRFMiddleware enforces signed double-submit validation.
func CSRFMiddleware(csrf *CSRFManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookieValue, err := c.Cookie(CSRFCookieName)
		if err != nil {
			securitylog.Log(c, "auth.csrf.failure", map[string]string{"reason": "missing_csrf_cookie"})
			httpx.WriteError(c, http.StatusForbidden, "csrf_invalid", "A valid CSRF token is required.", nil)
			c.Abort()
			return
		}

		headerValue := c.GetHeader("X-CSRF-Token")
		if headerValue == "" || headerValue != cookieValue || !csrf.Validate(headerValue) {
			securitylog.Log(c, "auth.csrf.failure", map[string]string{"reason": "csrf_mismatch"})
			httpx.WriteError(c, http.StatusForbidden, "csrf_invalid", "A valid CSRF token is required.", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// SessionFromGin returns the authenticated access session.
func SessionFromGin(c *gin.Context) (AccessSession, bool) {
	value, ok := c.Get(sessionContextKey)
	if !ok {
		return AccessSession{}, false
	}

	session, ok := value.(AccessSession)
	return session, ok
}
