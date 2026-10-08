-- Runs once, when the Postgres data volume is first created.
-- To re-run it: docker compose down -v && docker compose up -d
--
-- The schema mirrors internal/database/migrate.go. Keep them in sync.
-- Seed passwords are for local development only:
--   admin@example.com / changeme123   (admin)
--   demo@example.com  / password123   (regular user)

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    password_hash VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);

CREATE TABLE IF NOT EXISTS cafes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);

CREATE TABLE IF NOT EXISTS reviews (
    id SERIAL PRIMARY KEY,
    cafe_id INTEGER NOT NULL REFERENCES cafes(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    body TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    CONSTRAINT reviews_user_cafe_unique UNIQUE (user_id, cafe_id)
);

CREATE TABLE IF NOT EXISTS review_images (
    id SERIAL PRIMARY KEY,
    review_id INTEGER NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);

CREATE INDEX IF NOT EXISTS idx_reviews_cafe_id ON reviews(cafe_id);
CREATE INDEX IF NOT EXISTS idx_review_images_review_id ON review_images(review_id);

-- Seed data

INSERT INTO cafes (name, address, description) VALUES
    ('Bean There', '12 Market St', 'Single-origin pour-overs and a quiet back room.'),
    ('Morning Crumb', '48 Elm Ave', 'Fresh pastries baked before sunrise.'),
    ('Corner Roast', '7 Harbor Rd', 'Espresso bar with a good selection of cold brew.');

-- pgcrypto's bcrypt ($2a$) hashes are accepted by Go's bcrypt.CompareHashAndPassword.
INSERT INTO users (email, display_name, password_hash, role) VALUES
    ('admin@example.com', 'Admin', crypt('changeme123', gen_salt('bf')), 'admin'),
    ('demo@example.com', 'Demo User', crypt('password123', gen_salt('bf')), 'user');

INSERT INTO reviews (cafe_id, user_id, rating, body)
SELECT c.id, u.id, 5, 'Excellent pour-over and the back room is great for working.'
FROM cafes c, users u
WHERE c.name = 'Bean There' AND u.email = 'demo@example.com';

INSERT INTO reviews (cafe_id, user_id, rating, body)
SELECT c.id, u.id, 4, 'Best croissants in the area. Gets busy before 9am.'
FROM cafes c, users u
WHERE c.name = 'Morning Crumb' AND u.email = 'demo@example.com';
