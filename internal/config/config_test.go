package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validSecret is 32 bytes, the minimum accepted length.
const validSecret = "0123456789abcdef0123456789abcdef"

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"PORT", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "JWT_SECRET", "UPLOAD_DIR", "TRUSTED_PROXIES"} {
		t.Setenv(k, "")
	}
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_SECRET", validSecret)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, 5432, cfg.DBPort)
	assert.Equal(t, "cafe_review", cfg.DBName)
	assert.Equal(t, "uploads", cfg.UploadDir)
	assert.Equal(t, validSecret, cfg.JWTSecret)
}

func TestLoadReadsEnvironment(t *testing.T) {
	clearEnv(t)
	t.Setenv("PORT", "9000")
	t.Setenv("DB_PORT", "6543")
	t.Setenv("JWT_SECRET", validSecret)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "9000", cfg.Port)
	assert.Equal(t, 6543, cfg.DBPort)
}

func TestLoadRejectsInvalidDBPort(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_SECRET", validSecret)
	t.Setenv("DB_PORT", "not-a-port")

	_, err := Load()
	assert.Error(t, err)
}

func TestLoadRejectsWeakJWTSecret(t *testing.T) {
	tests := []struct {
		name     string
		secret   string
		contains string
	}{
		{"missing", "", "JWT_SECRET must be set"},
		{"placeholder from .env.example", "replace-me-with-openssl-rand-base64-32", "placeholder"},
		{"too short", "short", "at least 32 bytes (got 5)"},
		{"31 bytes", strings.Repeat("a", 31), "at least 32 bytes (got 31)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			t.Setenv("JWT_SECRET", tt.secret)

			_, err := Load()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.contains)
			assert.Contains(t, err.Error(), "openssl rand -base64 32", "error should say how to fix it")
		})
	}
}

func TestValidateJWTSecretAcceptsMinimumLength(t *testing.T) {
	assert.NoError(t, validateJWTSecret(strings.Repeat("a", 32)))
	assert.NoError(t, validateJWTSecret(validSecret))
	assert.NoError(t, validateJWTSecret(strings.Repeat("a", 64)))
}

func TestValidateJWTSecretCountsBytesNotCharacters(t *testing.T) {
	// 16 multi-byte runes are 32 bytes: accepted. 15 are 30 bytes: rejected.
	assert.NoError(t, validateJWTSecret(strings.Repeat("é", 16)))
	assert.Error(t, validateJWTSecret(strings.Repeat("é", 15)))
}

func TestSplitList(t *testing.T) {
	assert.Nil(t, splitList(""))
	assert.Nil(t, splitList("   "))
	assert.Equal(t, []string{"10.0.0.1"}, splitList("10.0.0.1"))
	assert.Equal(t, []string{"10.0.0.1", "10.0.0.2"}, splitList(" 10.0.0.1 , 10.0.0.2 "))
}

func TestLoadReadsTrustedProxies(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_SECRET", validSecret)
	t.Setenv("TRUSTED_PROXIES", "10.0.0.1, 10.0.0.2")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"10.0.0.1", "10.0.0.2"}, cfg.TrustedProxies)
}

func TestLoadTrustedProxiesEmptyByDefault(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_SECRET", validSecret)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Nil(t, cfg.TrustedProxies)
}
