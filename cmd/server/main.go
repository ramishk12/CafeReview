package main

import (
	"log"
	"net/http"
	"os"
	"time"

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

	auth.CookieSecure = cfg.CookieSecure

	uploads, err := storage.NewLocal(cfg.UploadDir, "/uploads")
	if err != nil {
		log.Fatalf("upload directory: %v", err)
	}
	handlers.Uploads = uploads

	r := server.NewRouter(cfg.UploadDir)
	// Trust no proxy by default, so a spoofed X-Forwarded-For cannot bypass the
	// per-IP rate limits. Set TRUSTED_PROXIES when running behind a proxy.
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		log.Fatalf("trusted proxies: %v", err)
	}

	// Explicit timeouts replace http.DefaultServer. ReadHeaderTimeout stops slow
	// header attacks (slowloris). ReadTimeout allows a 5 MB upload on a slow link.
	// WriteTimeout must cover reading the body and writing the response.
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	log.Printf("listening on :%s", cfg.Port)
	log.Fatal(srv.ListenAndServe())
}

// loadDotEnv reads .env if present. Variables already set in the environment take precedence.
func loadDotEnv() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("reading .env: %v", err)
	}
}
