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
			created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
			closed_at TIMESTAMP,
			background_image_path TEXT
		)`,
		// Columns added after the table existed, for databases created earlier.
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'cafes' AND column_name = 'closed_at'
			) THEN
				ALTER TABLE cafes ADD COLUMN closed_at TIMESTAMP;
			END IF;
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'cafes' AND column_name = 'background_image_path'
			) THEN
				ALTER TABLE cafes ADD COLUMN background_image_path TEXT;
			END IF;
		END $$;`,
		`CREATE TABLE IF NOT EXISTS reviews (
			id SERIAL PRIMARY KEY,
			cafe_id INTEGER NOT NULL REFERENCES cafes(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
			body TEXT NOT NULL,
			visited_on DATE NOT NULL DEFAULT (CURRENT_DATE),
			created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
			updated_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		)`,
		// Reviews used to be one per user per cafe. They are now one per user per
		// cafe per day, so a user can record each visit. Existing databases keep
		// their rows: visited_on is backfilled from created_at.
		`ALTER TABLE reviews DROP CONSTRAINT IF EXISTS reviews_user_cafe_unique`,
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'reviews' AND column_name = 'visited_on'
			) THEN
				ALTER TABLE reviews ADD COLUMN visited_on DATE;
				UPDATE reviews SET visited_on = (created_at AT TIME ZONE 'UTC')::date;
				ALTER TABLE reviews ALTER COLUMN visited_on SET NOT NULL;
				ALTER TABLE reviews ALTER COLUMN visited_on SET DEFAULT (CURRENT_DATE);
			END IF;
		END $$;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS reviews_user_cafe_visit_unique ON reviews(user_id, cafe_id, visited_on)`,
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
		`CREATE TABLE IF NOT EXISTS tags (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			category VARCHAR(50) NOT NULL,
			scope VARCHAR(20) NOT NULL DEFAULT 'cafe' CHECK (scope IN ('cafe', 'review', 'image')),
			status VARCHAR(20) NOT NULL DEFAULT 'approved' CHECK (status IN ('approved', 'pending')),
			created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
			UNIQUE (scope, name)
		)`,
		`CREATE TABLE IF NOT EXISTS cafe_tags (
			cafe_id INTEGER NOT NULL REFERENCES cafes(id) ON DELETE CASCADE,
			tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
			PRIMARY KEY (cafe_id, tag_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cafe_tags_tag_id ON cafe_tags(tag_id)`,
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
