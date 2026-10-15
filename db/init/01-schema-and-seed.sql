-- Runs once, when the Postgres data volume is first created.
-- To re-run it: docker compose down -v && docker compose up -d
--
-- The schema mirrors internal/database/migrate.go. Keep them in sync.
-- Seeds only sample cafes. Create accounts with cmd/createadmin or by registering.

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    password_hash VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    email_verified_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS cafes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    closed_at TIMESTAMP,
    background_image_path TEXT
);

CREATE TABLE IF NOT EXISTS reviews (
    id SERIAL PRIMARY KEY,
    cafe_id INTEGER NOT NULL REFERENCES cafes(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    body TEXT NOT NULL,
    visited_on DATE NOT NULL DEFAULT (CURRENT_DATE),
    created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);

CREATE TABLE IF NOT EXISTS review_images (
    id SERIAL PRIMARY KEY,
    review_id INTEGER NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);

ALTER TABLE cafes ADD COLUMN IF NOT EXISTS background_image_id INTEGER REFERENCES review_images(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    csrf_token VARCHAR(64) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS email_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    kind VARCHAR(20) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX IF NOT EXISTS idx_email_tokens_user_id ON email_tokens(user_id);
-- One review per user per cafe per day, so a user can record each visit.
CREATE UNIQUE INDEX IF NOT EXISTS reviews_user_cafe_visit_unique ON reviews(user_id, cafe_id, visited_on);
CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,
    scope VARCHAR(20) NOT NULL DEFAULT 'cafe' CHECK (scope IN ('cafe', 'review', 'image')),
    status VARCHAR(20) NOT NULL DEFAULT 'approved' CHECK (status IN ('approved', 'pending')),
    created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    UNIQUE (scope, name)
);

CREATE TABLE IF NOT EXISTS cafe_tags (
    cafe_id INTEGER NOT NULL REFERENCES cafes(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (cafe_id, tag_id)
);
CREATE INDEX IF NOT EXISTS idx_cafe_tags_tag_id ON cafe_tags(tag_id);

CREATE TABLE IF NOT EXISTS review_image_tags (
    image_id INTEGER NOT NULL REFERENCES review_images(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (image_id, tag_id)
);
CREATE INDEX IF NOT EXISTS idx_review_image_tags_tag_id ON review_image_tags(tag_id);

CREATE TABLE IF NOT EXISTS review_tags (
    review_id INTEGER NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (review_id, tag_id)
);
CREATE INDEX IF NOT EXISTS idx_review_tags_tag_id ON review_tags(tag_id);

CREATE TABLE IF NOT EXISTS review_votes (
    review_id INTEGER NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    PRIMARY KEY (review_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_review_votes_user_id ON review_votes(user_id);

CREATE TABLE IF NOT EXISTS review_revisions (
    id SERIAL PRIMARY KEY,
    review_id INTEGER NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL,
    body TEXT NOT NULL,
    visited_on DATE NOT NULL,
    replaced_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX IF NOT EXISTS idx_review_revisions_review_id ON review_revisions(review_id);

CREATE TABLE IF NOT EXISTS lists (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX IF NOT EXISTS idx_lists_user_id ON lists(user_id);

CREATE TABLE IF NOT EXISTS list_cafes (
    list_id INTEGER NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    cafe_id INTEGER NOT NULL REFERENCES cafes(id) ON DELETE CASCADE,
    added_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    PRIMARY KEY (list_id, cafe_id)
);
CREATE INDEX IF NOT EXISTS idx_list_cafes_cafe_id ON list_cafes(cafe_id);

CREATE TABLE IF NOT EXISTS list_reviews (
    list_id INTEGER NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    review_id INTEGER NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    added_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    PRIMARY KEY (list_id, review_id)
);
CREATE INDEX IF NOT EXISTS idx_list_reviews_review_id ON list_reviews(review_id);

CREATE TABLE IF NOT EXISTS list_users (
    list_id INTEGER NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    added_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    PRIMARY KEY (list_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_list_users_user_id ON list_users(user_id);

CREATE INDEX IF NOT EXISTS idx_reviews_cafe_id ON reviews(cafe_id);
CREATE INDEX IF NOT EXISTS idx_review_images_review_id ON review_images(review_id);

-- Starter cafe tags. Shared with review and image tags through the scope column.
INSERT INTO tags (name, category, scope) VALUES
    ('Downtown', 'location', 'cafe'),
    ('Near campus', 'location', 'cafe'),
    ('Waterfront', 'location', 'cafe'),
    ('Quiet', 'vibe', 'cafe'),
    ('Lively', 'vibe', 'cafe'),
    ('Study-friendly', 'vibe', 'cafe'),
    ('Cozy', 'vibe', 'cafe'),
    ('Espresso', 'drinks', 'cafe'),
    ('Pour-over', 'drinks', 'cafe'),
    ('Cold brew', 'drinks', 'cafe'),
    ('Matcha', 'drinks', 'cafe'),
    ('Tea', 'drinks', 'cafe'),
    ('Drink', 'photo', 'image'),
    ('Interior', 'photo', 'image'),
    ('Exterior', 'photo', 'image'),
    ('Seasonal menu', 'seasonal', 'review'),
    ('Limited edition', 'seasonal', 'review'),
    ('Event', 'event', 'review');

-- Seed data: sample cafes only. No accounts or reviews are created.

INSERT INTO cafes (name, address, description) VALUES
    ('Bean There', '12 Market St', 'Single-origin pour-overs and a quiet back room.'),
    ('Morning Crumb', '48 Elm Ave', 'Fresh pastries baked before sunrise.'),
    ('Corner Roast', '7 Harbor Rd', 'Espresso bar with a good selection of cold brew.');
