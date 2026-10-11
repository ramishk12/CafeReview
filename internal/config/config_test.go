package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"PORT", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "JWT_SECRET", "UPLOAD_DIR"} {
		t.Setenv(k, "")
	}
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)

	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, 5432, cfg.DBPort)
	assert.Equal(t, "cafe_review", cfg.DBName)
	assert.Equal(t, "uploads", cfg.UploadDir)
	assert.Empty(t, cfg.JWTSecret)
}

func TestLoadReadsEnvironment(t *testing.T) {
	clearEnv(t)
	t.Setenv("PORT", "9000")
	t.Setenv("DB_PORT", "6543")
	t.Setenv("JWT_SECRET", "s3cret")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, "9000", cfg.Port)
	assert.Equal(t, 6543, cfg.DBPort)
	assert.Equal(t, "s3cret", cfg.JWTSecret)
}

func TestLoadRejectsInvalidDBPort(t *testing.T) {
	clearEnv(t)
	t.Setenv("DB_PORT", "not-a-port")

	_, err := Load()
	assert.Error(t, err)
}
