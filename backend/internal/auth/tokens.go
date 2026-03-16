package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AccessTokenTTL is the short-lived browser access lifetime.
const AccessTokenTTL = 15 * time.Minute

// RefreshTokenTTL is the server-tracked refresh lifetime.
const RefreshTokenTTL = 30 * 24 * time.Hour

type accessClaims struct {
	UserID            string `json:"uid"`
	UserRole          string `json:"urole"`
	ActiveHouseholdID string `json:"hid"`
	HouseholdRole     string `json:"hrole"`
	jwt.RegisteredClaims
}

// TokenSigner signs and validates access tokens.
type TokenSigner struct {
	secret []byte
}

// NewTokenSigner creates a new access token signer.
func NewTokenSigner(secret string) *TokenSigner {
	return &TokenSigner{secret: []byte(secret)}
}

// SignAccessToken returns a signed access token.
func (s *TokenSigner) SignAccessToken(session AccessSession, issuedAt time.Time, expiresAt time.Time) (string, error) {
	claims := accessClaims{
		UserID:            session.UserID,
		UserRole:          session.UserRole,
		ActiveHouseholdID: session.ActiveHouseholdID,
		HouseholdRole:     string(session.HouseholdRole),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   session.UserID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ParseAccessToken validates and decodes the browser access token.
func (s *TokenSigner) ParseAccessToken(tokenValue string) (AccessSession, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	token, err := parser.ParseWithClaims(tokenValue, &accessClaims{}, func(token *jwt.Token) (any, error) {
		return s.secret, nil
	})
	if err != nil {
		return AccessSession{}, err
	}

	claims, ok := token.Claims.(*accessClaims)
	if !ok || !token.Valid {
		return AccessSession{}, fmt.Errorf("invalid access token")
	}
	if claims.UserID == "" || claims.ActiveHouseholdID == "" || claims.HouseholdRole == "" {
		return AccessSession{}, errors.New("invalid access token claims")
	}
	if claims.Subject != "" && claims.Subject != claims.UserID {
		return AccessSession{}, errors.New("invalid access token subject")
	}

	return AccessSession{
		UserID:            claims.UserID,
		UserRole:          claims.UserRole,
		ActiveHouseholdID: claims.ActiveHouseholdID,
		HouseholdRole:     HouseholdRole(claims.HouseholdRole),
	}, nil
}

// NewOpaqueToken returns a random URL-safe token.
func NewOpaqueToken() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("read random token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

// HashOpaqueToken hashes an opaque token for database storage.
func HashOpaqueToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// SignValue returns a signed string suitable for cookies and headers.
func SignValue(secret string, value string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(value))
	signature := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s.%s", value, signature)
}

// ValidateSignedValue verifies a signed cookie/header value.
func ValidateSignedValue(secret string, signed string) bool {
	value, signature, ok := splitSignedValue(signed)
	if !ok {
		return false
	}

	expected := SignValue(secret, value)
	return hmac.Equal([]byte(expected), []byte(signed)) && signature != ""
}

func splitSignedValue(signed string) (string, string, bool) {
	for index := len(signed) - 1; index >= 0; index-- {
		if signed[index] == '.' {
			return signed[:index], signed[index+1:], true
		}
	}

	return "", "", false
}
