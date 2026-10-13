package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// placeholderPrefix marks the placeholders in .env.example. A value that starts
// with it has not been replaced.
const placeholderPrefix = "replace-me"

// Config holds runtime settings read from environment variables.
type Config struct {
	Port       string
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	UploadDir  string
	// TrustedProxies are the addresses whose X-Forwarded-For header is trusted for
	// the client IP. Empty means trust none, so the client IP is the direct peer.
	TrustedProxies []string
	// CookieSecure sets the Secure flag on session cookies. Defaults to true.
	// Set COOKIE_SECURE=false for local http development.
	CookieSecure bool

	// Email settings. When SMTPHost is empty, emails are written to the log instead.
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	MailFrom     string
	// AppBaseURL is the public origin of the frontend, used to build links in emails.
	AppBaseURL string
}

// Load reads configuration from the environment with defaults for local development.
func Load() (Config, error) {
	cfg := Config{
		Port:       getEnv("PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     getEnv("DB_NAME", "cafe_review"),
		UploadDir:  getEnv("UPLOAD_DIR", "uploads"),

		TrustedProxies: splitList(os.Getenv("TRUSTED_PROXIES")),

		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPUser:     os.Getenv("SMTP_USER"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		MailFrom:     getEnv("MAIL_FROM", "CafeReview <no-reply@cafereview.local>"),
		AppBaseURL:   getEnv("APP_BASE_URL", "http://localhost:3000"),
	}

	smtpPort, err := strconv.Atoi(getEnv("SMTP_PORT", "1025"))
	if err != nil {
		return cfg, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}
	cfg.SMTPPort = smtpPort

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return cfg, fmt.Errorf("invalid DB_PORT: %w", err)
	}
	cfg.DBPort = dbPort

	cookieSecure, err := parseBoolDefault(os.Getenv("COOKIE_SECURE"), true)
	if err != nil {
		return cfg, fmt.Errorf("invalid COOKIE_SECURE: %w", err)
	}
	cfg.CookieSecure = cookieSecure

	if err := validateDBPassword(cfg.DBPassword); err != nil {
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

// parseBoolDefault parses a boolean, returning fallback when the value is empty.
func parseBoolDefault(value string, fallback bool) (bool, error) {
	if strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	return strconv.ParseBool(value)
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
