package database

import (
	"fmt"
)

// Migrate creates the schema if it does not already exist.
func Migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			display_name VARCHAR(100) NOT NULL,
			password_hash VARCHAR(100) NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
			created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
			email_verified_at TIMESTAMP
		)`,
		// Add the column to databases created before it existed, and mark their
		// existing users verified so they are not locked out. Runs only once,
		// when the column is absent.
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'users' AND column_name = 'email_verified_at'
			) THEN
				ALTER TABLE users ADD COLUMN email_verified_at TIMESTAMP;
				UPDATE users SET email_verified_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC');
			END IF;
		END $$;`,
		`CREATE TABLE IF NOT EXISTS cafes (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			address TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		)`,
		`CREATE TABLE IF NOT EXISTS reviews (
			id SERIAL PRIMARY KEY,
			cafe_id INTEGER NOT NULL REFERENCES cafes(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
			body TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
			updated_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
			CONSTRAINT reviews_user_cafe_unique UNIQUE (user_id, cafe_id)
		)`,
		`CREATE TABLE IF NOT EXISTS review_images (
			id SERIAL PRIMARY KEY,
			review_id INTEGER NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
			filename VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id SERIAL PRIMARY KEY,
			token_hash VARCHAR(64) NOT NULL UNIQUE,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			csrf_token VARCHAR(64) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
			expires_at TIMESTAMP NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions(token_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at)`,
		`CREATE TABLE IF NOT EXISTS email_tokens (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash VARCHAR(64) NOT NULL UNIQUE,
			kind VARCHAR(20) NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			used_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		)`,
		`CREATE INDEX IF NOT EXISTS idx_email_tokens_user_id ON email_tokens(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_reviews_cafe_id ON reviews(cafe_id)`,
		`CREATE INDEX IF NOT EXISTS idx_review_images_review_id ON review_images(review_id)`,
	}

	for i, stmt := range migrations {
		if _, err := DB.Exec(stmt); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}
	return nil
}
