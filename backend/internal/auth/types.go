package auth

import "time"

// HouseholdRole identifies a member's permissions within a household.
type HouseholdRole string

const (
	HouseholdRoleViewer HouseholdRole = "viewer"
	HouseholdRoleEditor HouseholdRole = "editor"
	HouseholdRoleOwner  HouseholdRole = "owner"
)

// UserView is the authenticated user snapshot.
type UserView struct {
	ID            string
	DisplayName   string
	Email         string
	Role          string
	LastLoginAt   *time.Time
	EmailVerified bool
}

// HouseholdView is the active household snapshot.
type HouseholdView struct {
	ID   string
	Name string
	Slug string
	Role HouseholdRole
}

// SessionView is returned by auth use-cases.
type SessionView struct {
	Authenticated   bool
	User            *UserView
	ActiveHousehold *HouseholdView
}

// SessionCookies contains cookie payloads written after auth flows.
type SessionCookies struct {
	AccessToken    string
	RefreshToken   string
	CSRFToken      string
	AccessExpires  time.Time
	RefreshExpires time.Time
}

// Result is the common output for auth mutations.
type Result struct {
	Session SessionView
	Cookies SessionCookies
}

// RegisterInput creates a local user and bootstrap household.
type RegisterInput struct {
	DisplayName string
	Email       string
	Password    string
}

// LoginInput authenticates a local user.
type LoginInput struct {
	Email    string
	Password string
}

// SwitchHouseholdInput changes the active household context.
type SwitchHouseholdInput struct {
	HouseholdID string
}

// AccessSession is the auth context carried in the access cookie.
type AccessSession struct {
	UserID            string
	UserRole          string
	ActiveHouseholdID string
	HouseholdRole     HouseholdRole
}
