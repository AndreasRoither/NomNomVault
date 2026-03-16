package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"golang.org/x/crypto/argon2"

	"github.com/AndreasRoither/NomNomVault/backend/internal/ent"
	entgen "github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/household"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/householdmember"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/refreshsession"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/user"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/clock"
	"github.com/AndreasRoither/NomNomVault/backend/internal/platform/httpx"
)

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

// Service implements the cookie-first auth flows.
type Service struct {
	db     *ent.Client
	clock  clock.Clock
	signer *TokenSigner
	csrf   *CSRFManager
}

// NewService creates an auth service.
func NewService(db *ent.Client, clock clock.Clock, signer *TokenSigner, csrf *CSRFManager) *Service {
	return &Service{db: db, clock: clock, signer: signer, csrf: csrf}
}

// NewCSRFToken creates a new CSRF token and its cookie expiry.
func (s *Service) NewCSRFToken() (string, time.Time, error) {
	token, err := s.csrf.Generate()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("create csrf token: %w", err)
	}

	return token, s.clock.Now().Add(RefreshTokenTTL), nil
}

// Register creates a user, default household, and active session.
func (s *Service) Register(ctx context.Context, in RegisterInput) (Result, error) {
	displayName := strings.TrimSpace(in.DisplayName)
	if displayName == "" {
		return Result{}, httpx.StatusError{Status: http.StatusBadRequest, Code: "validation_error", Message: "Display name is required."}
	}

	email := normalizeEmail(in.Email)
	passwordHash, err := hashPassword(in.Password)
	if err != nil {
		return Result{}, fmt.Errorf("hash password: %w", err)
	}

	now := s.clock.Now()
	refreshToken, err := NewOpaqueToken()
	if err != nil {
		return Result{}, fmt.Errorf("create refresh token: %w", err)
	}
	csrfToken, err := s.csrf.Generate()
	if err != nil {
		return Result{}, fmt.Errorf("create csrf token: %w", err)
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("start tx: %w", err)
	}

	userEntity, householdEntity, memberEntity, err := s.createBootstrapIdentity(ctx, tx, displayName, email, passwordHash, now)
	if err != nil {
		_ = tx.Rollback()
		return Result{}, err
	}

	refreshExpires := now.Add(RefreshTokenTTL)
	refreshEntity, err := tx.RefreshSession.Create().
		SetUserID(userEntity.ID).
		SetActiveHouseholdID(householdEntity.ID).
		SetTokenHash(HashOpaqueToken(refreshToken)).
		SetExpiresAt(refreshExpires).
		SetLastUsedAt(now).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return Result{}, mapConstraintError("create refresh session", err)
	}

	if err := tx.Commit(); err != nil {
		return Result{}, fmt.Errorf("commit register tx: %w", err)
	}

	return s.buildResult(userEntity, householdEntity, HouseholdRole(memberEntity.Role), refreshEntity, refreshToken, csrfToken)
}

// Login authenticates a local user and creates a new refresh session.
func (s *Service) Login(ctx context.Context, in LoginInput) (Result, error) {
	email := normalizeEmail(in.Email)
	userEntity, err := s.db.User.Query().Where(user.EmailEQ(email)).Only(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return Result{}, httpx.StatusError{Status: http.StatusUnauthorized, Code: "invalid_credentials", Message: "Email or password is invalid."}
		}

		return Result{}, fmt.Errorf("query user: %w", err)
	}

	if !verifyPassword(userEntity.PasswordHash, in.Password) {
		return Result{}, httpx.StatusError{Status: http.StatusUnauthorized, Code: "invalid_credentials", Message: "Email or password is invalid."}
	}

	memberEntity, householdEntity, err := s.firstMembership(ctx, userEntity.ID)
	if err != nil {
		return Result{}, err
	}

	now := s.clock.Now()
	refreshToken, err := NewOpaqueToken()
	if err != nil {
		return Result{}, fmt.Errorf("create refresh token: %w", err)
	}
	csrfToken, err := s.csrf.Generate()
	if err != nil {
		return Result{}, fmt.Errorf("create csrf token: %w", err)
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("start tx: %w", err)
	}

	userEntity, err = tx.User.UpdateOneID(userEntity.ID).SetLastLoginAt(now).Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return Result{}, fmt.Errorf("update last login: %w", err)
	}

	refreshExpires := now.Add(RefreshTokenTTL)
	refreshEntity, err := tx.RefreshSession.Create().
		SetUserID(userEntity.ID).
		SetActiveHouseholdID(householdEntity.ID).
		SetTokenHash(HashOpaqueToken(refreshToken)).
		SetExpiresAt(refreshExpires).
		SetLastUsedAt(now).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return Result{}, mapConstraintError("create refresh session", err)
	}

	if err := tx.Commit(); err != nil {
		return Result{}, fmt.Errorf("commit login tx: %w", err)
	}

	return s.buildResult(userEntity, householdEntity, HouseholdRole(memberEntity.Role), refreshEntity, refreshToken, csrfToken)
}

// Session builds the current session snapshot from a browser access token.
func (s *Service) Session(ctx context.Context, accessToken string) (SessionView, error) {
	if strings.TrimSpace(accessToken) == "" {
		return SessionView{Authenticated: false}, nil
	}

	session, err := s.signer.ParseAccessToken(accessToken)
	if err != nil {
		return SessionView{Authenticated: false}, nil
	}

	return s.sessionFromAccess(ctx, session)
}

// Refresh rotates the current refresh session and returns a fresh browser session.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (Result, error) {
	refreshEntity, userEntity, householdEntity, memberRole, err := s.loadRefreshSession(ctx, refreshToken)
	if err != nil {
		return Result{}, err
	}

	now := s.clock.Now()
	newRefreshToken, err := NewOpaqueToken()
	if err != nil {
		return Result{}, fmt.Errorf("create refresh token: %w", err)
	}
	csrfToken, err := s.csrf.Generate()
	if err != nil {
		return Result{}, fmt.Errorf("create csrf token: %w", err)
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("start tx: %w", err)
	}

	if _, err := tx.RefreshSession.UpdateOneID(refreshEntity.ID).
		SetRevoked(true).
		SetLastUsedAt(now).
		Save(ctx); err != nil {
		_ = tx.Rollback()
		return Result{}, fmt.Errorf("revoke refresh session: %w", err)
	}

	refreshExpires := now.Add(RefreshTokenTTL)
	newRefreshEntity, err := tx.RefreshSession.Create().
		SetUserID(userEntity.ID).
		SetActiveHouseholdID(householdEntity.ID).
		SetTokenHash(HashOpaqueToken(newRefreshToken)).
		SetExpiresAt(refreshExpires).
		SetLastUsedAt(now).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return Result{}, fmt.Errorf("create rotated refresh session: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return Result{}, fmt.Errorf("commit refresh tx: %w", err)
	}

	return s.buildResult(userEntity, householdEntity, memberRole, newRefreshEntity, newRefreshToken, csrfToken)
}

// Logout revokes the current refresh session if present.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return nil
	}

	_, err := s.db.RefreshSession.Update().
		Where(refreshsession.TokenHashEQ(HashOpaqueToken(refreshToken)), refreshsession.RevokedEQ(false)).
		SetRevoked(true).
		SetLastUsedAt(s.clock.Now()).
		Save(ctx)
	if err != nil && !entgen.IsNotFound(err) {
		return fmt.Errorf("revoke refresh session: %w", err)
	}

	return nil
}

// SwitchHousehold rotates the refresh record and changes the active household context.
func (s *Service) SwitchHousehold(ctx context.Context, refreshToken string, in SwitchHouseholdInput) (Result, error) {
	refreshEntity, userEntity, _, _, err := s.loadRefreshSession(ctx, refreshToken)
	if err != nil {
		return Result{}, err
	}

	memberEntity, householdEntity, err := s.membershipForHousehold(ctx, userEntity.ID, in.HouseholdID)
	if err != nil {
		return Result{}, err
	}

	now := s.clock.Now()
	newRefreshToken, err := NewOpaqueToken()
	if err != nil {
		return Result{}, fmt.Errorf("create refresh token: %w", err)
	}
	csrfToken, err := s.csrf.Generate()
	if err != nil {
		return Result{}, fmt.Errorf("create csrf token: %w", err)
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("start tx: %w", err)
	}

	if _, err := tx.RefreshSession.UpdateOneID(refreshEntity.ID).
		SetRevoked(true).
		SetLastUsedAt(now).
		Save(ctx); err != nil {
		_ = tx.Rollback()
		return Result{}, fmt.Errorf("revoke refresh session: %w", err)
	}

	refreshExpires := now.Add(RefreshTokenTTL)
	newRefreshEntity, err := tx.RefreshSession.Create().
		SetUserID(userEntity.ID).
		SetActiveHouseholdID(householdEntity.ID).
		SetTokenHash(HashOpaqueToken(newRefreshToken)).
		SetExpiresAt(refreshExpires).
		SetLastUsedAt(now).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return Result{}, fmt.Errorf("create switched refresh session: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return Result{}, fmt.Errorf("commit switch tx: %w", err)
	}

	return s.buildResult(userEntity, householdEntity, HouseholdRole(memberEntity.Role), newRefreshEntity, newRefreshToken, csrfToken)
}

func (s *Service) createBootstrapIdentity(ctx context.Context, tx *ent.Tx, displayName string, email string, passwordHash string, now time.Time) (*entgen.User, *entgen.Household, *entgen.HouseholdMember, error) {
	userEntity, err := tx.User.Create().
		SetDisplayName(strings.TrimSpace(displayName)).
		SetEmail(email).
		SetPasswordHash(passwordHash).
		SetLastLoginAt(now).
		Save(ctx)
	if err != nil {
		return nil, nil, nil, mapConstraintError("create user", err)
	}

	householdName := defaultHouseholdName(displayName)
	slug, err := s.allocateHouseholdSlug(ctx, tx.Client(), householdName)
	if err != nil {
		return nil, nil, nil, err
	}

	householdEntity, err := tx.Household.Create().
		SetName(householdName).
		SetSlug(slug).
		Save(ctx)
	if err != nil {
		return nil, nil, nil, mapConstraintError("create household", err)
	}

	memberEntity, err := tx.HouseholdMember.Create().
		SetHouseholdID(householdEntity.ID).
		SetUserID(userEntity.ID).
		SetRole(householdmember.RoleOwner).
		Save(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create household membership: %w", err)
	}

	return userEntity, householdEntity, memberEntity, nil
}

func (s *Service) sessionFromAccess(ctx context.Context, access AccessSession) (SessionView, error) {
	userEntity, err := s.db.User.Get(ctx, access.UserID)
	if err != nil {
		if entgen.IsNotFound(err) {
			return SessionView{Authenticated: false}, nil
		}

		return SessionView{}, fmt.Errorf("load session user: %w", err)
	}

	memberEntity, householdEntity, err := s.membershipForHousehold(ctx, access.UserID, access.ActiveHouseholdID)
	if err != nil {
		if statusErr, ok := err.(httpx.StatusError); ok && statusErr.Status == http.StatusForbidden {
			return SessionView{Authenticated: false}, nil
		}

		return SessionView{}, err
	}

	return buildSessionView(userEntity, householdEntity, HouseholdRole(memberEntity.Role)), nil
}

func (s *Service) loadRefreshSession(ctx context.Context, refreshToken string) (*entgen.RefreshSession, *entgen.User, *entgen.Household, HouseholdRole, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, nil, nil, "", httpx.StatusError{Status: http.StatusUnauthorized, Code: "unauthenticated", Message: "Authentication is required."}
	}

	refreshEntity, err := s.db.RefreshSession.Query().
		Where(refreshsession.TokenHashEQ(HashOpaqueToken(refreshToken)), refreshsession.RevokedEQ(false)).
		WithUser().
		WithActiveHousehold().
		Only(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return nil, nil, nil, "", httpx.StatusError{Status: http.StatusUnauthorized, Code: "unauthenticated", Message: "Authentication is required."}
		}

		return nil, nil, nil, "", fmt.Errorf("query refresh session: %w", err)
	}

	if refreshEntity.ExpiresAt.Before(s.clock.Now()) {
		return nil, nil, nil, "", httpx.StatusError{Status: http.StatusUnauthorized, Code: "unauthenticated", Message: "Authentication is required."}
	}

	memberEntity, householdEntity, err := s.membershipForHousehold(ctx, refreshEntity.UserID, refreshEntity.ActiveHouseholdID)
	if err != nil {
		return nil, nil, nil, "", err
	}

	return refreshEntity, refreshEntity.Edges.User, householdEntity, HouseholdRole(memberEntity.Role), nil
}

func (s *Service) firstMembership(ctx context.Context, userID string) (*entgen.HouseholdMember, *entgen.Household, error) {
	memberEntity, err := s.db.HouseholdMember.Query().
		Where(householdmember.UserIDEQ(userID)).
		Order(entgen.Asc(householdmember.FieldCreatedAt)).
		WithHousehold().
		First(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return nil, nil, httpx.StatusError{Status: http.StatusForbidden, Code: "household_required", Message: "The user does not belong to a household."}
		}

		return nil, nil, fmt.Errorf("query user membership: %w", err)
	}

	return memberEntity, memberEntity.Edges.Household, nil
}

func (s *Service) membershipForHousehold(ctx context.Context, userID string, householdID string) (*entgen.HouseholdMember, *entgen.Household, error) {
	memberEntity, err := s.db.HouseholdMember.Query().
		Where(householdmember.UserIDEQ(userID), householdmember.HouseholdIDEQ(householdID)).
		WithHousehold().
		Only(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return nil, nil, httpx.StatusError{Status: http.StatusForbidden, Code: "forbidden", Message: "The current account cannot access that household."}
		}

		return nil, nil, fmt.Errorf("query household membership: %w", err)
	}

	return memberEntity, memberEntity.Edges.Household, nil
}

func (s *Service) allocateHouseholdSlug(ctx context.Context, client *ent.Client, householdName string) (string, error) {
	base := slugify(householdName)
	if base == "" {
		base = "household"
	}

	candidate := base
	for suffix := 1; suffix < 1000; suffix++ {
		exists, err := client.Household.Query().Where(household.SlugEQ(candidate)).Exist(ctx)
		if err != nil {
			return "", fmt.Errorf("check household slug: %w", err)
		}
		if !exists {
			return candidate, nil
		}

		candidate = fmt.Sprintf("%s-%d", base, suffix)
	}

	return "", fmt.Errorf("allocate household slug for %q", householdName)
}

func (s *Service) buildResult(userEntity *entgen.User, householdEntity *entgen.Household, memberRole HouseholdRole, refreshEntity *entgen.RefreshSession, refreshToken string, csrfToken string) (Result, error) {
	now := s.clock.Now()
	accessExpires := now.Add(AccessTokenTTL)
	accessToken, err := s.signer.SignAccessToken(AccessSession{
		UserID:            userEntity.ID,
		UserRole:          string(userEntity.Role),
		ActiveHouseholdID: householdEntity.ID,
		HouseholdRole:     memberRole,
	}, now, accessExpires)
	if err != nil {
		return Result{}, fmt.Errorf("sign access token: %w", err)
	}

	return Result{
		Session: buildSessionView(userEntity, householdEntity, memberRole),
		Cookies: SessionCookies{
			AccessToken:    accessToken,
			RefreshToken:   refreshToken,
			CSRFToken:      csrfToken,
			AccessExpires:  accessExpires,
			RefreshExpires: refreshEntity.ExpiresAt,
		},
	}, nil
}

func buildSessionView(userEntity *entgen.User, householdEntity *entgen.Household, memberRole HouseholdRole) SessionView {
	return SessionView{
		Authenticated: true,
		User: &UserView{
			ID:            userEntity.ID,
			DisplayName:   userEntity.DisplayName,
			Email:         userEntity.Email,
			Role:          string(userEntity.Role),
			LastLoginAt:   userEntity.LastLoginAt,
			EmailVerified: userEntity.EmailVerified,
		},
		ActiveHousehold: &HouseholdView{
			ID:   householdEntity.ID,
			Name: householdEntity.Name,
			Slug: householdEntity.Slug,
			Role: memberRole,
		},
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func defaultHouseholdName(displayName string) string {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return "NomNomVault Household"
	}

	return fmt.Sprintf("%s's Household", displayName)
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = slugSanitizer.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	return value
}

func mapConstraintError(action string, err error) error {
	if sqlgraph.IsConstraintError(err) {
		if strings.Contains(strings.ToLower(err.Error()), "users_email_key") || strings.Contains(strings.ToLower(err.Error()), "users.email") {
			return httpx.StatusError{Status: http.StatusBadRequest, Code: "registration_failed", Message: "Registration could not be completed."}
		}
	}

	return fmt.Errorf("%s: %w", action, err)
}

func hashPassword(password string) (string, error) {
	var salt [16]byte
	if _, err := rand.Read(salt[:]); err != nil {
		return "", fmt.Errorf("read salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt[:], 3, 64*1024, 2, 32)
	return fmt.Sprintf(
		"$argon2id$v=19$m=65536,t=3,p=2$%s$%s",
		base64.RawStdEncoding.EncodeToString(salt[:]),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func verifyPassword(encoded string, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	computed := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, uint32(len(expectedHash)))
	return subtle.ConstantTimeCompare(expectedHash, computed) == 1
}
