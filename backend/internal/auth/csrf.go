package auth

import "fmt"

// CSRFManager signs and validates double-submit cookie tokens.
type CSRFManager struct {
	secret string
}

// NewCSRFManager creates a CSRF helper.
func NewCSRFManager(secret string) *CSRFManager {
	return &CSRFManager{secret: secret}
}

// Generate creates a signed CSRF token suitable for cookie/header echo.
func (m *CSRFManager) Generate() (string, error) {
	raw, err := NewOpaqueToken()
	if err != nil {
		return "", fmt.Errorf("generate csrf token: %w", err)
	}

	return SignValue(m.secret, raw), nil
}

// Validate returns true when the signed CSRF token is valid.
func (m *CSRFManager) Validate(token string) bool {
	return ValidateSignedValue(m.secret, token)
}
