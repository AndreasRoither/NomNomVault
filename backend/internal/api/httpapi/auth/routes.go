package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AndreasRoither/NomNomVault/backend/internal/api/apicontract"
	authsvc "github.com/AndreasRoither/NomNomVault/backend/internal/auth"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/httpx"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/ratelimit"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/securitylog"
)

type handler struct {
	service        *authsvc.Service
	cookies        *authsvc.CookieManager
	csrf           *authsvc.CSRFManager
	loginLimiter   *ratelimit.Limiter
	refreshLimiter *ratelimit.Limiter
}

// RegisterRoutes binds the real auth endpoints for cookie-first browser auth.
func RegisterRoutes(
	api *gin.RouterGroup,
	service *authsvc.Service,
	cookies *authsvc.CookieManager,
	csrf *authsvc.CSRFManager,
	loginLimiter *ratelimit.Limiter,
	refreshLimiter *ratelimit.Limiter,
) {
	h := &handler{
		service:        service,
		cookies:        cookies,
		csrf:           csrf,
		loginLimiter:   loginLimiter,
		refreshLimiter: refreshLimiter,
	}

	auth := api.Group("/auth")
	auth.GET("/session", h.session)

	writes := auth.Group("")
	writes.Use(authsvc.CSRFMiddleware(csrf))
	writes.POST("/register", h.register)
	writes.POST("/login", h.login)
	writes.POST("/logout", h.logout)
	writes.POST("/refresh", h.refresh)
}

// register godoc
// @Summary Register a local user
// @Description Create a user and bootstrap the default household context.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body AuthRegisterRequest true "Register payload"
// @Success 201 {object} AuthSessionResponse
// @Failure 400 {object} apicontract.ErrorResponse
// @Failure 403 {object} apicontract.ErrorResponse
// @Router /auth/register [post]
func (h *handler) register(c *gin.Context) {
	var request AuthRegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		securitylog.Log(c, "auth.register.failure", map[string]string{"reason": "validation_error"})
		httpx.WriteValidationError(c, validationErrors("payload", err.Error()))
		return
	}

	result, err := h.service.Register(c.Request.Context(), authsvc.RegisterInput{
		DisplayName: request.DisplayName,
		Email:       request.Email,
		Password:    request.Password,
	})
	if err != nil {
		securitylog.Log(c, "auth.register.failure", map[string]string{"reason": serviceReason(err)})
		httpx.WriteServiceError(c, err)
		return
	}

	h.cookies.WriteSessionCookies(c, result.Cookies)
	securitylog.Log(c, "auth.register.success", map[string]string{
		"user_id":      result.Session.User.ID,
		"household_id": result.Session.ActiveHousehold.ID,
	})
	c.JSON(http.StatusCreated, mapSessionResponse(result.Session))
}

// login godoc
// @Summary Log in with email and password
// @Description Authenticate a local user and return the active session snapshot.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body AuthLoginRequest true "Login payload"
// @Success 200 {object} AuthSessionResponse
// @Failure 400 {object} apicontract.ErrorResponse
// @Failure 403 {object} apicontract.ErrorResponse
// @Failure 429 {object} apicontract.ErrorResponse
// @Router /auth/login [post]
func (h *handler) login(c *gin.Context) {
	var request AuthLoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		securitylog.Log(c, "auth.login.failure", map[string]string{"reason": "validation_error"})
		httpx.WriteValidationError(c, validationErrors("payload", err.Error()))
		return
	}

	if retryAfter, limited := limitRequest(h.loginLimiter, loginRateLimitKey(request.Email, c.ClientIP())); limited {
		c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())+1))
		securitylog.Log(c, "auth.login.failure", map[string]string{"reason": "rate_limited"})
		httpx.WriteError(c, http.StatusTooManyRequests, "rate_limited", "Too many login attempts. Please try again later.", nil)
		return
	}

	result, err := h.service.Login(c.Request.Context(), authsvc.LoginInput{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		securitylog.Log(c, "auth.login.failure", map[string]string{"reason": serviceReason(err)})
		httpx.WriteServiceError(c, err)
		return
	}

	h.cookies.WriteSessionCookies(c, result.Cookies)
	securitylog.Log(c, "auth.login.success", map[string]string{
		"user_id":      result.Session.User.ID,
		"household_id": result.Session.ActiveHousehold.ID,
	})
	c.JSON(http.StatusOK, mapSessionResponse(result.Session))
}

// logout godoc
// @Summary Log out the current session
// @Description Clear the current session in the cookie-first auth flow.
// @Tags auth
// @Success 204
// @Failure 401 {object} apicontract.ErrorResponse
// @Failure 403 {object} apicontract.ErrorResponse
// @Router /auth/logout [post]
func (h *handler) logout(c *gin.Context) {
	refreshToken, ok := h.cookies.ReadRefreshToken(c)
	if !ok {
		securitylog.Log(c, "auth.logout.failure", map[string]string{"reason": "missing_refresh_cookie"})
		httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Authentication is required.", nil)
		return
	}

	if err := h.service.Logout(c.Request.Context(), refreshToken); err != nil {
		securitylog.Log(c, "auth.logout.failure", map[string]string{"reason": serviceReason(err)})
		httpx.WriteServiceError(c, err)
		return
	}

	h.cookies.ClearSessionCookies(c)
	securitylog.Log(c, "auth.logout.success", map[string]string{"reason": "revoked"})
	c.Status(http.StatusNoContent)
}

// refresh godoc
// @Summary Refresh the current session
// @Description Rotate the refresh record and return the current active session snapshot.
// @Tags auth
// @Produce json
// @Success 200 {object} AuthSessionResponse
// @Failure 401 {object} apicontract.ErrorResponse
// @Failure 403 {object} apicontract.ErrorResponse
// @Failure 429 {object} apicontract.ErrorResponse
// @Router /auth/refresh [post]
func (h *handler) refresh(c *gin.Context) {
	refreshToken, ok := h.cookies.ReadRefreshToken(c)
	if !ok {
		securitylog.Log(c, "auth.refresh.failure", map[string]string{"reason": "missing_refresh_cookie"})
		httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Authentication is required.", nil)
		return
	}

	if retryAfter, limited := limitRequest(h.refreshLimiter, refreshRateLimitKey(refreshToken, c.ClientIP())); limited {
		c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())+1))
		securitylog.Log(c, "auth.refresh.failure", map[string]string{"reason": "rate_limited"})
		httpx.WriteError(c, http.StatusTooManyRequests, "rate_limited", "Too many refresh attempts. Please try again later.", nil)
		return
	}

	result, err := h.service.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		securitylog.Log(c, "auth.refresh.failure", map[string]string{"reason": serviceReason(err)})
		httpx.WriteServiceError(c, err)
		return
	}

	h.cookies.WriteSessionCookies(c, result.Cookies)
	securitylog.Log(c, "auth.refresh.success", map[string]string{
		"user_id":      result.Session.User.ID,
		"household_id": result.Session.ActiveHousehold.ID,
	})
	c.JSON(http.StatusOK, mapSessionResponse(result.Session))
}

// session godoc
// @Summary Fetch the active browser session
// @Description Return the authenticated user and active household context.
// @Tags auth
// @Produce json
// @Success 200 {object} AuthSessionResponse
// @Router /auth/session [get]
func (h *handler) session(c *gin.Context) {
	if err := h.ensureCSRFCookie(c); err != nil {
		httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "The request could not be completed.", nil)
		return
	}

	accessToken, _ := h.cookies.ReadAccessToken(c)
	session, err := h.service.Session(c.Request.Context(), accessToken)
	if err != nil {
		httpx.WriteServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapSessionResponse(session))
}

func (h *handler) ensureCSRFCookie(c *gin.Context) error {
	if value, err := c.Cookie(authsvc.CSRFCookieName); err == nil && h.csrf.Validate(value) {
		return nil
	}

	token, expiresAt, err := h.service.NewCSRFToken()
	if err != nil {
		return err
	}
	h.cookies.WriteCSRFCookie(c, token, expiresAt)
	return nil
}

func mapSessionResponse(session authsvc.SessionView) AuthSessionResponse {
	response := AuthSessionResponse{Authenticated: session.Authenticated}
	if session.User != nil {
		response.User = AuthenticatedUser{
			ID:            session.User.ID,
			DisplayName:   session.User.DisplayName,
			Email:         session.User.Email,
			Role:          session.User.Role,
			LastLoginAt:   session.User.LastLoginAt,
			EmailVerified: session.User.EmailVerified,
		}
	}
	if session.ActiveHousehold != nil {
		response.ActiveHousehold = AuthenticatedHousehold{
			ID:   session.ActiveHousehold.ID,
			Name: session.ActiveHousehold.Name,
			Slug: session.ActiveHousehold.Slug,
			Role: string(session.ActiveHousehold.Role),
		}
	}
	return response
}

func limitRequest(limiter *ratelimit.Limiter, key string) (time.Duration, bool) {
	allowed, retryAfter := limiter.Allow(key, time.Now().UTC())
	return retryAfter, !allowed
}

func loginRateLimitKey(email string, clientIP string) string {
	return strings.ToLower(strings.TrimSpace(email)) + "|" + strings.TrimSpace(clientIP)
}

func refreshRateLimitKey(refreshToken string, clientIP string) string {
	tokenHash := authsvc.HashOpaqueToken(refreshToken)
	if len(tokenHash) > 16 {
		tokenHash = tokenHash[:16]
	}
	return strings.TrimSpace(clientIP) + "|" + tokenHash
}

func serviceReason(err error) string {
	var statusErr httpx.StatusError
	if errors.As(err, &statusErr) {
		return statusErr.Code
	}
	return "internal_error"
}

func validationErrors(field string, message string) []apicontract.ValidationError {
	return []apicontract.ValidationError{{Field: field, Message: message}}
}
