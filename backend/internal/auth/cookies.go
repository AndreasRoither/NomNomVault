package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	AccessCookieName  = "nnv_access"
	RefreshCookieName = "nnv_refresh"
	CSRFCookieName    = "nnv_csrf"
)

// CookieManager writes and clears browser cookies.
type CookieManager struct {
	secure bool
}

// NewCookieManager creates a cookie writer.
func NewCookieManager(secure bool) *CookieManager {
	return &CookieManager{secure: secure}
}

// WriteSessionCookies writes the current auth cookie set.
func (m *CookieManager) WriteSessionCookies(c *gin.Context, cookies SessionCookies) {
	writeCookie(c.Writer, AccessCookieName, cookies.AccessToken, "/", cookies.AccessExpires, true, m.secure)
	writeCookie(c.Writer, RefreshCookieName, cookies.RefreshToken, "/api/v1/auth", cookies.RefreshExpires, true, m.secure)
	writeCookie(c.Writer, CSRFCookieName, cookies.CSRFToken, "/", cookies.RefreshExpires, false, m.secure)
}

// EnsureCSRFCookie writes a CSRF cookie only when missing.
func (m *CookieManager) EnsureCSRFCookie(c *gin.Context, csrfToken string, expiresAt time.Time) {
	if _, err := c.Cookie(CSRFCookieName); err == nil {
		return
	}

	m.WriteCSRFCookie(c, csrfToken, expiresAt)
}

// WriteCSRFCookie writes the CSRF cookie regardless of existing state.
func (m *CookieManager) WriteCSRFCookie(c *gin.Context, csrfToken string, expiresAt time.Time) {
	writeCookie(c.Writer, CSRFCookieName, csrfToken, "/", expiresAt, false, m.secure)
}

// ClearSessionCookies clears all auth cookies.
func (m *CookieManager) ClearSessionCookies(c *gin.Context) {
	clearCookie(c.Writer, AccessCookieName, "/", true, m.secure)
	clearCookie(c.Writer, RefreshCookieName, "/api/v1/auth", true, m.secure)
	clearCookie(c.Writer, CSRFCookieName, "/", false, m.secure)
}

// ReadAccessToken returns the access token from cookies.
func (m *CookieManager) ReadAccessToken(c *gin.Context) (string, bool) {
	value, err := c.Cookie(AccessCookieName)
	return value, err == nil && value != ""
}

// ReadRefreshToken returns the refresh token from cookies.
func (m *CookieManager) ReadRefreshToken(c *gin.Context) (string, bool) {
	value, err := c.Cookie(RefreshCookieName)
	return value, err == nil && value != ""
}

func writeCookie(writer http.ResponseWriter, name string, value string, path string, expiresAt time.Time, httpOnly bool, secure bool) {
	http.SetCookie(writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Expires:  expiresAt,
		MaxAge:   maxAge(expiresAt),
		HttpOnly: httpOnly,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	})
}

func clearCookie(writer http.ResponseWriter, name string, path string, httpOnly bool, secure bool) {
	http.SetCookie(writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	})
}

func maxAge(expiresAt time.Time) int {
	seconds := int(time.Until(expiresAt).Seconds())
	if seconds < 0 {
		return 0
	}

	return seconds
}
