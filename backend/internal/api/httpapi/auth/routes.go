package httpapi

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AndreasRoither/NomNomVault/backend/internal/api/apicontract"
)

// RegisterRoutes binds the auth endpoints for the first generated API contract.
func RegisterRoutes(api *gin.RouterGroup) {
	auth := api.Group("/auth")
	auth.POST("/register", register)
	auth.POST("/login", login)
	auth.POST("/logout", logout)
	auth.POST("/refresh", refresh)
	auth.GET("/session", session)
}

// register godoc
// @Summary Register a local user
// @Description Create a user and bootstrap the default household context.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body AuthRegisterRequest true "Register payload"
// @Success 200 {object} AuthSessionResponse
// @Failure 400 {object} apicontract.ErrorResponse
// @Router /auth/register [post]
func register(context *gin.Context) {
	var request AuthRegisterRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		apicontract.WriteError(
			context,
			http.StatusBadRequest,
			"validation_error",
			"Request payload is invalid.",
			[]apicontract.ValidationError{{Field: "payload", Message: err.Error()}},
		)
		return
	}

	context.JSON(http.StatusOK, sessionFromInput(request.DisplayName, request.Email))
}

// login godoc
// @Summary Log in with email and password
// @Description Validate the local login contract and return the active session snapshot.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body AuthLoginRequest true "Login payload"
// @Success 200 {object} AuthSessionResponse
// @Failure 400 {object} apicontract.ErrorResponse
// @Router /auth/login [post]
func login(context *gin.Context) {
	var request AuthLoginRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		apicontract.WriteError(
			context,
			http.StatusBadRequest,
			"validation_error",
			"Request payload is invalid.",
			[]apicontract.ValidationError{{Field: "payload", Message: err.Error()}},
		)
		return
	}

	context.JSON(http.StatusOK, sessionFromInput("NomNom User", request.Email))
}

// logout godoc
// @Summary Log out the current session
// @Description Clear the current session in the cookie-first auth flow.
// @Tags auth
// @Success 204
// @Router /auth/logout [post]
func logout(context *gin.Context) {
	context.Status(http.StatusNoContent)
}

// refresh godoc
// @Summary Refresh the current session
// @Description Rotate the refresh record and return the current active session snapshot.
// @Tags auth
// @Produce json
// @Success 200 {object} AuthSessionResponse
// @Router /auth/refresh [post]
func refresh(context *gin.Context) {
	context.JSON(http.StatusOK, sampleSession())
}

// session godoc
// @Summary Fetch the active browser session
// @Description Return the authenticated user and active household context.
// @Tags auth
// @Produce json
// @Success 200 {object} AuthSessionResponse
// @Router /auth/session [get]
func session(context *gin.Context) {
	context.JSON(http.StatusOK, sampleSession())
}

func sessionFromInput(displayName string, email string) AuthSessionResponse {
	now := time.Now().UTC()

	return AuthSessionResponse{
		Authenticated: true,
		User: AuthenticatedUser{
			ID:            "user_sample",
			DisplayName:   displayName,
			Email:         email,
			Role:          "user",
			LastLoginAt:   &now,
			EmailVerified: true,
		},
		ActiveHousehold: AuthenticatedHousehold{
			ID:   "household_sample",
			Name: "NomNom Household",
			Slug: "nomnom-household",
			Role: "owner",
		},
	}
}

func sampleSession() AuthSessionResponse {
	return sessionFromInput("NomNom User", "cook@nomnomvault.local")
}
