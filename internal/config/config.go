package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// minJWTSecretBytes is the minimum length for HS256 signing keys: the key should be at least
// as long as the hash output (256 bits), per RFC 7518 section 3.2.
const minJWTSecretBytes = 32

// placeholderPrefix marks the placeholders in .env.example. A value that starts
// with it has not been replaced.
const placeholderPrefix = "replace-me"

const jwtSecretHelp = "generate one with: openssl rand -base64 32"

// Config holds runtime settings read from environment variables.
type Config struct {
	Port       string
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	UploadDir  string
	// TrustedProxies are the addresses whose X-Forwarded-For header is trusted for
	// the client IP. Empty means trust none, so the client IP is the direct peer.
	TrustedProxies []string
}

// Load reads configuration from the environment with defaults for local development.
func Load() (Config, error) {
	cfg := Config{
		Port:       getEnv("PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     getEnv("DB_NAME", "cafe_review"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		UploadDir:  getEnv("UPLOAD_DIR", "uploads"),

		TrustedProxies: splitList(os.Getenv("TRUSTED_PROXIES")),
	}

	port, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return cfg, fmt.Errorf("invalid DB_PORT: %w", err)
	}
	cfg.DBPort = port

	if err := validateDBPassword(cfg.DBPassword); err != nil {
		return cfg, err
	}
	if err := validateJWTSecret(cfg.JWTSecret); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// validateDBPassword requires a database password and rejects the .env.example
// placeholder, so no working default credential ships with the code.
func validateDBPassword(password string) error {
	switch {
	case password == "":
		return errors.New("DB_PASSWORD must be set")
	case strings.HasPrefix(password, placeholderPrefix):
		return errors.New("DB_PASSWORD is still the placeholder from .env.example; set a strong password")
	}
	return nil
}

// validateJWTSecret rejects missing, placeholder, and short signing secrets. Length is counted in bytes.
func validateJWTSecret(secret string) error {
	switch {
	case secret == "":
		return errors.New("JWT_SECRET must be set; " + jwtSecretHelp)
	case strings.HasPrefix(secret, placeholderPrefix):
		return errors.New("JWT_SECRET is still the placeholder from .env.example; " + jwtSecretHelp)
	case len(secret) < minJWTSecretBytes:
		return fmt.Errorf("JWT_SECRET must be at least %d bytes (got %d); %s", minJWTSecretBytes, len(secret), jwtSecretHelp)
	}
	return nil
}

// splitList splits a comma-separated list, trimming blanks. An empty string returns nil.
func splitList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
