package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config contains API runtime settings.
type Config struct {
	AppEnv                        string
	DatabaseURL                   string
	AuthJWTSecret                 string
	AuthCSRFSecret                string
	CookieSecure                  bool
	TrustedProxies                []string
	MaxUploadBytes                int64
	AllowedUploadMIMEs            []string
	AllowedCORSOrigins            []string
	AuthLoginRateLimitPerMinute   int
	AuthRefreshRateLimitPerMinute int
}

// Load returns runtime config from the environment.
func Load() (Config, error) {
	uploadLimit := int64(10 << 20)
	if raw := strings.TrimSpace(os.Getenv("MAX_UPLOAD_BYTES")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return Config{}, fmt.Errorf("parse MAX_UPLOAD_BYTES: %w", err)
		}
		uploadLimit = value
	}

	cfg := Config{
		AppEnv:                        firstNonEmpty(strings.TrimSpace(os.Getenv("APP_ENV")), "development"),
		DatabaseURL:                   strings.TrimSpace(os.Getenv("DATABASE_URL")),
		AuthJWTSecret:                 strings.TrimSpace(os.Getenv("AUTH_JWT_SECRET")),
		AuthCSRFSecret:                strings.TrimSpace(os.Getenv("AUTH_CSRF_SECRET")),
		CookieSecure:                  parseBool(os.Getenv("COOKIE_SECURE")),
		TrustedProxies:                csvDefault(os.Getenv("TRUSTED_PROXIES"), []string{}),
		MaxUploadBytes:                uploadLimit,
		AllowedUploadMIMEs:            csvDefault(os.Getenv("ALLOWED_UPLOAD_MIME_TYPES"), []string{"image/jpeg", "image/png", "image/webp"}),
		AllowedCORSOrigins:            csvDefault(os.Getenv("ALLOWED_CORS_ORIGINS"), []string{}),
		AuthLoginRateLimitPerMinute:   intDefault(os.Getenv("AUTH_LOGIN_RATE_LIMIT_PER_MINUTE"), 5),
		AuthRefreshRateLimitPerMinute: intDefault(os.Getenv("AUTH_REFRESH_RATE_LIMIT_PER_MINUTE"), 30),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.AppEnv == "production" {
		if cfg.AuthJWTSecret == "" {
			return Config{}, fmt.Errorf("AUTH_JWT_SECRET is required in production")
		}
		if cfg.AuthCSRFSecret == "" {
			return Config{}, fmt.Errorf("AUTH_CSRF_SECRET is required in production")
		}
		if !cfg.CookieSecure {
			return Config{}, fmt.Errorf("COOKIE_SECURE must be true in production")
		}
	} else {
		cfg.AuthJWTSecret = firstNonEmpty(cfg.AuthJWTSecret, "dev-jwt-secret-not-for-development")
		cfg.AuthCSRFSecret = firstNonEmpty(cfg.AuthCSRFSecret, "dev-csrf-secret-not-for-development")
	}

	return cfg, nil
}

func parseBool(raw string) bool {
	value, err := strconv.ParseBool(strings.TrimSpace(raw))
	return err == nil && value
}

func intDefault(raw string, fallback int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}

	return value
}

func csv(raw string) []string {
	return csvDefault(raw, nil)
}

func csvDefault(raw string, fallback []string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}

	return values
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}

	return ""
}
