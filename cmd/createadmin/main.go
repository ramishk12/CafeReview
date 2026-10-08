// createadmin creates an admin user, or promotes an existing user to admin.
//
// Usage:
//
//	ADMIN_EMAIL=you@example.com ADMIN_PASSWORD=secret123 ADMIN_NAME=Admin go run ./cmd/createadmin
package main

import (
	"log"
	"os"
	"strings"

	"cafe-review/internal/auth"
	"cafe-review/internal/config"
	"cafe-review/internal/database"
	"cafe-review/internal/models"
)

func main() {
	email := strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_EMAIL")))
	password := os.Getenv("ADMIN_PASSWORD")
	name := strings.TrimSpace(os.Getenv("ADMIN_NAME"))
	if name == "" {
		name = "Admin"
	}
	if email == "" || len(password) < 8 {
		log.Fatal("ADMIN_EMAIL and ADMIN_PASSWORD (min 8 characters) must be set")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Connect(database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
	}); err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatal(err)
	}

	// Existing users are promoted but keep their current password.
	var id int
	err = database.DB.QueryRow(
		`INSERT INTO users (email, display_name, password_hash, role)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (email) DO UPDATE SET role = EXCLUDED.role
		 RETURNING id`,
		email, name, hash, models.RoleAdmin,
	).Scan(&id)
	if err != nil {
		log.Fatalf("could not create admin: %v", err)
	}
	log.Printf("admin user ready: %s (id %d)", email, id)
}
