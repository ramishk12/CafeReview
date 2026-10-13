package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validDBPassword = "test-db-password"

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"PORT", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "UPLOAD_DIR", "TRUSTED_PROXIES", "COOKIE_SECURE", "SMTP_HOST", "SMTP_PORT", "SMTP_USER", "SMTP_PASSWORD", "MAIL_FROM", "APP_BASE_URL"} {
		t.Setenv(k, "")
	}
}

// setValid sets the required variable so Load succeeds.
func setValid(t *testing.T) {
	t.Helper()
	t.Setenv("DB_PASSWORD", validDBPassword)
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)
	setValid(t)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, 5432, cfg.DBPort)
	assert.Equal(t, "cafe_review", cfg.DBName)
	assert.Equal(t, "uploads", cfg.UploadDir)
	assert.True(t, cfg.CookieSecure, "cookies default to Secure")
}

func TestLoadReadsEnvironment(t *testing.T) {
	clearEnv(t)
	setValid(t)
	t.Setenv("PORT", "9000")
	t.Setenv("DB_PORT", "6543")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "9000", cfg.Port)
	assert.Equal(t, 6543, cfg.DBPort)
}

func TestLoadRejectsInvalidDBPort(t *testing.T) {
	clearEnv(t)
	setValid(t)
	t.Setenv("DB_PORT", "not-a-port")

	_, err := Load()
	assert.Error(t, err)
}

func TestLoadRejectsMissingDBPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		contains string
	}{
		{"missing", "", "DB_PASSWORD must be set"},
		{"placeholder from .env.example", "replace-me-with-a-strong-password", "placeholder"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			t.Setenv("DB_PASSWORD", tt.password)

			_, err := Load()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.contains)
		})
	}
}

func TestValidateDBPassword(t *testing.T) {
	assert.NoError(t, validateDBPassword("a-strong-password"))
	assert.Error(t, validateDBPassword(""))
	assert.Error(t, validateDBPassword("replace-me-with-a-strong-password"))
}

func TestCookieSecure(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    bool
		wantErr bool
	}{
		{"default is secure", "", true, false},
		{"explicit true", "true", true, false},
		{"explicit false for local http", "false", false, false},
		{"invalid", "notabool", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			setValid(t)
			t.Setenv("COOKIE_SECURE", tt.value)

			cfg, err := Load()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg.CookieSecure)
		})
	}
}

func TestSplitList(t *testing.T) {
	assert.Nil(t, splitList(""))
	assert.Nil(t, splitList("   "))
	assert.Equal(t, []string{"10.0.0.1"}, splitList("10.0.0.1"))
	assert.Equal(t, []string{"10.0.0.1", "10.0.0.2"}, splitList(" 10.0.0.1 , 10.0.0.2 "))
}

func TestLoadReadsTrustedProxies(t *testing.T) {
	clearEnv(t)
	setValid(t)
	t.Setenv("TRUSTED_PROXIES", "10.0.0.1, 10.0.0.2")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"10.0.0.1", "10.0.0.2"}, cfg.TrustedProxies)
}

func TestLoadTrustedProxiesEmptyByDefault(t *testing.T) {
	clearEnv(t)
	setValid(t)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Nil(t, cfg.TrustedProxies)
}

func TestEmailDefaults(t *testing.T) {
	clearEnv(t)
	setValid(t)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Empty(t, cfg.SMTPHost)
	assert.Equal(t, 1025, cfg.SMTPPort)
	assert.Equal(t, "http://localhost:3000", cfg.AppBaseURL)
	assert.Equal(t, "CafeReview <no-reply@cafereview.local>", cfg.MailFrom)
}

func TestLoadRejectsInvalidSMTPPort(t *testing.T) {
	clearEnv(t)
	setValid(t)
	t.Setenv("SMTP_PORT", "not-a-port")

	_, err := Load()
	assert.Error(t, err)
}
