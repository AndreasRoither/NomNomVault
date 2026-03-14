package httpapi

import "time"

// AuthRegisterRequest is the payload for the register endpoint.
type AuthRegisterRequest struct {
	DisplayName string `json:"displayName" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
}

// AuthLoginRequest is the payload for the login endpoint.
type AuthLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthenticatedUser is the user snapshot returned in session responses.
type AuthenticatedUser struct {
	ID            string     `json:"id"`
	DisplayName   string     `json:"displayName"`
	Email         string     `json:"email"`
	Role          string     `json:"role"`
	LastLoginAt   *time.Time `json:"lastLoginAt,omitempty"`
	EmailVerified bool       `json:"emailVerified"`
}

// AuthenticatedHousehold is the active household snapshot in session responses.
type AuthenticatedHousehold struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Role string `json:"role"`
}

// AuthSessionResponse is returned by the auth/session endpoints.
type AuthSessionResponse struct {
	Authenticated   bool                   `json:"authenticated"`
	User            AuthenticatedUser      `json:"user"`
	ActiveHousehold AuthenticatedHousehold `json:"activeHousehold"`
}
