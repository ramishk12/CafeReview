package config

import (
	"fmt"
	"os"
	"strconv"
)

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
}

// Load reads configuration from the environment with defaults for local development.
func Load() (Config, error) {
	cfg := Config{
		Port:       getEnv("PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "cafe_review"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		UploadDir:  getEnv("UPLOAD_DIR", "uploads"),
	}

	port, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return cfg, fmt.Errorf("invalid DB_PORT: %w", err)
	}
	cfg.DBPort = port

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
