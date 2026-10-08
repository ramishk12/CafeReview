package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"cafe-review/internal/auth"
	"cafe-review/internal/config"
	"cafe-review/internal/database"
	"cafe-review/internal/handlers"
	"cafe-review/internal/server"
	"cafe-review/internal/storage"
)

func main() {
	loadDotEnv()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET must be set")
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
	if err := database.SeedCafes(); err != nil {
		log.Fatalf("seeding failed: %v", err)
	}

	auth.Secret = []byte(cfg.JWTSecret)

	uploads, err := storage.NewLocal(cfg.UploadDir, "/uploads")
	if err != nil {
		log.Fatalf("upload directory: %v", err)
	}
	handlers.Uploads = uploads

	r := server.NewRouter(cfg.UploadDir)
	log.Printf("listening on :%s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}

// loadDotEnv reads .env if present. Variables already set in the environment take precedence.
func loadDotEnv() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("reading .env: %v", err)
	}
}
