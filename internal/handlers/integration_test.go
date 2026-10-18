package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"cafe-review/internal/database"
	"cafe-review/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// These tests run against a real PostgreSQL database, so every query is parsed
// and executed by the server instead of being matched by sqlmock. They are
// skipped unless TEST_DATABASE_URL is set, so `go test ./...` stays fast and
// database-free:
//
//	TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5432/cafe_review_test?sslmode=disable' go test ./...
//
// The database name must end in "_test": every table is truncated between tests.

// integrationTables lists every table in the schema so each test starts clean.
var integrationTables = []string{
	"users", "sessions", "email_tokens",
	"cafes", "cafe_tags",
	"reviews", "review_images", "review_image_tags", "review_tags", "review_votes", "review_revisions",
	"tags",
	"lists", "list_cafes", "list_users", "list_reviews",
}

// integrationDB connects to the test database, applies the migrations, and
// empties every table. database.DB is replaced for the duration of the test.
func integrationDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run the integration tests (see the README)")
	}

	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}
	name := strings.TrimPrefix(u.Path, "/")
	if !strings.HasSuffix(name, "_test") {
		t.Fatalf("TEST_DATABASE_URL must name a database ending in _test, got %q", name)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	if _, err := db.Exec(`SET timezone = 'UTC'`); err != nil {
		t.Fatalf("set timezone: %v", err)
	}

	previous := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = previous })

	if err := database.Migrate(); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	if _, err := db.Exec(`TRUNCATE ` + strings.Join(integrationTables, ", ") + ` RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
	return db
}

// --- fixtures ---

func testUser(t *testing.T, db *sql.DB, name string) int {
	t.Helper()
	email := strings.ToLower(strings.ReplaceAll(name, " ", ".")) + "@example.com"
	var id int
	err := db.QueryRow(
		`INSERT INTO users (email, display_name, password_hash, role, email_verified_at)
		 VALUES ($1, $2, 'integration-test', 'user', (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'))
		 RETURNING id`,
		email, name,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert user %q: %v", name, err)
	}
	return id
}

func testCafe(t *testing.T, db *sql.DB, name, address string) int {
	t.Helper()
	var id int
	if err := db.QueryRow(
		`INSERT INTO cafes (name, address) VALUES ($1, $2) RETURNING id`, name, address,
	).Scan(&id); err != nil {
		t.Fatalf("insert cafe %q: %v", name, err)
	}
	return id
}

func testReview(t *testing.T, db *sql.DB, cafeID, userID, rating int, body, visitedOn string) int {
	t.Helper()
	var id int
	err := db.QueryRow(
		`INSERT INTO reviews (cafe_id, user_id, rating, body, visited_on) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		cafeID, userID, rating, body, visitedOn,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert review: %v", err)
	}
	return id
}

func testTag(t *testing.T, db *sql.DB, name, scope string) int {
	t.Helper()
	var id int
	err := db.QueryRow(
		`INSERT INTO tags (name, category, scope) VALUES ($1, 'test', $2) RETURNING id`, name, scope,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert tag %q: %v", name, err)
	}
	return id
}

func testImage(t *testing.T, db *sql.DB, reviewID int, filename string) int {
	t.Helper()
	var id int
	err := db.QueryRow(
		`INSERT INTO review_images (review_id, filename) VALUES ($1, $2) RETURNING id`, reviewID, filename,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert review image: %v", err)
	}
	return id
}

// pastDate returns a date in the past, so a visit is never rejected as future.
func pastDate(days int) string {
	return time.Now().UTC().AddDate(0, 0, -days).Format("2006-01-02")
}

// idParam builds the gin params for a route with an :id.
func idParam(id int) gin.Params {
	return gin.Params{{Key: "id", Value: strconv.Itoa(id)}}
}

// decodeJSON decodes a response body, failing with the raw body on error.
func decodeJSON(t *testing.T, w *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), target); err != nil {
		t.Fatalf("decode response: %v (body: %s)", err, w.Body.String())
	}
}

// requireStatus fails with the response body when the status is unexpected.
func requireStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("status %d, want %d (body: %s)", w.Code, want, w.Body.String())
	}
}

// --- tests ---

// TestIntegrationListCafesWithReviews is the regression test for the GROUP BY
// bug the mocked tests missed (#86): the list query joins reviews and groups by
// the cafe, so a column that is neither grouped nor aggregated is rejected by
// PostgreSQL while sqlmock still passes.
func TestIntegrationListCafesWithReviews(t *testing.T) {
	db := integrationDB(t)

	alice := testUser(t, db, "Alice")
	bob := testUser(t, db, "Bob")
	bean := testCafe(t, db, "Bean There", "12 Market St")
	testCafe(t, db, "Morning Crumb", "48 Elm Ave")
	testReview(t, db, bean, alice, 5, "Excellent.", pastDate(3))
	testReview(t, db, bean, bob, 3, "Fine.", pastDate(2))

	c, w := newTestContext(t, http.MethodGet, "/api/cafes?sort=rating", nil, nil, 0, "")
	ListCafes(c)
	requireStatus(t, w, http.StatusOK)
	assert.Equal(t, "2", w.Header().Get("X-Total-Count"))

	var cafes []models.Cafe
	decodeJSON(t, w, &cafes)
	if assert.Len(t, cafes, 2) {
		assert.Equal(t, "Bean There", cafes[0].Name)
		assert.Equal(t, 4.0, cafes[0].AvgRating)
		assert.Equal(t, 2, cafes[0].ReviewCount)
		assert.Equal(t, 2, cafes[0].ReviewerCount)
	}

	// The address search and the tag filter exercise the LIKE escape and the
	// grouped HAVING clause against the real planner.
	quiet := testTag(t, db, "Quiet", "cafe")
	if _, err := db.Exec(`INSERT INTO cafe_tags (cafe_id, tag_id) VALUES ($1, $2)`, bean, quiet); err != nil {
		t.Fatalf("tag cafe: %v", err)
	}

	c2, w2 := newTestContext(t, http.MethodGet, "/api/cafes?q=market", nil, nil, 0, "")
	ListCafes(c2)
	requireStatus(t, w2, http.StatusOK)
	var matches []models.Cafe
	decodeJSON(t, w2, &matches)
	if assert.Len(t, matches, 1) {
		assert.Equal(t, "Bean There", matches[0].Name)
	}

	c3, w3 := newTestContext(t, http.MethodGet, "/api/cafes?tags="+strconv.Itoa(quiet), nil, nil, 0, "")
	ListCafes(c3)
	requireStatus(t, w3, http.StatusOK)
	var tagged []models.Cafe
	decodeJSON(t, w3, &tagged)
	if assert.Len(t, tagged, 1) {
		assert.Equal(t, "Bean There", tagged[0].Name)
		if assert.Len(t, tagged[0].Tags, 1) {
			assert.Equal(t, "Quiet", tagged[0].Tags[0].Name)
		}
	}
}

func TestIntegrationGetCafe(t *testing.T) {
	db := integrationDB(t)

	alice := testUser(t, db, "Alice")
	cafe := testCafe(t, db, "Bean There", "12 Market St")
	testReview(t, db, cafe, alice, 4, "Good.", pastDate(1))
	quiet := testTag(t, db, "Quiet", "cafe")
	if _, err := db.Exec(`INSERT INTO cafe_tags (cafe_id, tag_id) VALUES ($1, $2)`, cafe, quiet); err != nil {
		t.Fatalf("tag cafe: %v", err)
	}

	c, w := newTestContext(t, http.MethodGet, "/api/cafes/1", nil, idParam(cafe), 0, "")
	GetCafe(c)
	requireStatus(t, w, http.StatusOK)

	var got models.Cafe
	decodeJSON(t, w, &got)
	assert.Equal(t, "Bean There", got.Name)
	assert.Equal(t, 4.0, got.AvgRating)
	assert.Equal(t, 1, got.ReviewCount)
	if assert.Len(t, got.Tags, 1) {
		assert.Equal(t, "Quiet", got.Tags[0].Name)
	}

	c2, w2 := newTestContext(t, http.MethodGet, "/api/cafes/9999", nil, idParam(9999), 0, "")
	GetCafe(c2)
	assert.Equal(t, http.StatusNotFound, w2.Code)
}

func TestIntegrationListReviewsSortsAndFilters(t *testing.T) {
	db := integrationDB(t)

	alice := testUser(t, db, "Alice")
	bob := testUser(t, db, "Bob")
	carla := testUser(t, db, "Carla")
	dan := testUser(t, db, "Dan")
	cafe := testCafe(t, db, "Bean There", "12 Market St")

	plain := testReview(t, db, cafe, bob, 2, "Too cold.", pastDate(1))
	rich := testReview(t, db, cafe, alice, 5, "Excellent.", pastDate(2))
	testImage(t, db, rich, "photo.jpg")
	tag := testTag(t, db, "Seasonal menu", "review")
	if _, err := db.Exec(`INSERT INTO review_tags (review_id, tag_id) VALUES ($1, $2)`, rich, tag); err != nil {
		t.Fatalf("tag review: %v", err)
	}
	// Two helpful votes on the plain review, so the helpful sort must lift it.
	if _, err := db.Exec(
		`INSERT INTO review_votes (review_id, user_id) VALUES ($1, $2), ($1, $3)`,
		plain, carla, dan,
	); err != nil {
		t.Fatalf("vote: %v", err)
	}

	list := func(t *testing.T, query string) ([]models.Review, string) {
		t.Helper()
		c, w := newTestContext(t, http.MethodGet, "/api/cafes/1/reviews"+query, nil, idParam(cafe), 0, "")
		ListReviews(c)
		requireStatus(t, w, http.StatusOK)
		var reviews []models.Review
		decodeJSON(t, w, &reviews)
		return reviews, w.Header().Get("X-Total-Count")
	}

	// Newest first by default, with the total in the header.
	reviews, total := list(t, "")
	assert.Equal(t, "2", total)
	if assert.Len(t, reviews, 2) {
		assert.Equal(t, rich, reviews[0].ID)
	}

	// Most helpful orders by vote count, not by date.
	reviews, _ = list(t, "?sort=helpful")
	if assert.Len(t, reviews, 2) {
		assert.Equal(t, plain, reviews[0].ID)
		assert.Equal(t, 2, reviews[0].HelpfulCount)
	}

	// Minimum rating.
	reviews, total = list(t, "?rating=4")
	assert.Equal(t, "1", total)
	if assert.Len(t, reviews, 1) {
		assert.Equal(t, rich, reviews[0].ID)
	}

	// Reviews with photos, including the loaded image.
	reviews, total = list(t, "?photos=1")
	assert.Equal(t, "1", total)
	if assert.Len(t, reviews, 1) {
		assert.Equal(t, rich, reviews[0].ID)
		if assert.Len(t, reviews[0].Images, 1) {
			assert.Equal(t, "photo.jpg", strings.TrimPrefix(reviews[0].Images[0].URL, "/uploads/"))
		}
	}

	// Review tag.
	reviews, total = list(t, "?tags="+strconv.Itoa(tag))
	assert.Equal(t, "1", total)
	if assert.Len(t, reviews, 1) {
		assert.Equal(t, rich, reviews[0].ID)
	}

	// Reviewer.
	reviews, total = list(t, "?user_id="+strconv.Itoa(bob))
	assert.Equal(t, "1", total)
	if assert.Len(t, reviews, 1) {
		assert.Equal(t, plain, reviews[0].ID)
	}

	// Paging keeps the total.
	reviews, total = list(t, "?limit=1&offset=1")
	assert.Equal(t, "2", total)
	assert.Len(t, reviews, 1)

	// Invalid filters are rejected before touching the database.
	for _, query := range []string{"?rating=9", "?user_id=0"} {
		c, w := newTestContext(t, http.MethodGet, "/api/cafes/1/reviews"+query, nil, idParam(cafe), 0, "")
		ListReviews(c)
		assert.Equal(t, http.StatusBadRequest, w.Code, query)
	}
}

func TestIntegrationReviewLifecycle(t *testing.T) {
	db := integrationDB(t)

	alice := testUser(t, db, "Alice")
	cafe := testCafe(t, db, "Bean There", "12 Market St")
	date := pastDate(1)

	create := func(t *testing.T, body map[string]any) *httptest.ResponseRecorder {
		t.Helper()
		c, w := newTestContext(t, http.MethodPost, "/api/cafes/1/reviews", body, idParam(cafe), alice, models.RoleUser)
		CreateReview(c)
		return w
	}

	w := create(t, map[string]any{"rating": 5, "body": "Excellent.", "visited_on": date})
	requireStatus(t, w, http.StatusCreated)

	var created models.Review
	decodeJSON(t, w, &created)
	assert.Equal(t, 5, created.Rating)

	// One review per user per cafe per day.
	w = create(t, map[string]any{"rating": 4, "body": "Again.", "visited_on": date})
	assert.Equal(t, http.StatusConflict, w.Code)

	// Editing keeps the previous version.
	c, w := newTestContext(t, http.MethodPut, "/api/reviews/1",
		map[string]any{"rating": 3, "body": "Colder today.", "visited_on": date}, idParam(created.ID), alice, models.RoleUser)
	UpdateReview(c)
	requireStatus(t, w, http.StatusOK)

	c2, w2 := newTestContext(t, http.MethodGet, "/api/reviews/1/revisions", nil, idParam(created.ID), alice, models.RoleUser)
	GetReviewRevisions(c2)
	requireStatus(t, w2, http.StatusOK)

	var revisions []models.ReviewRevision
	decodeJSON(t, w2, &revisions)
	if assert.Len(t, revisions, 1) {
		assert.Equal(t, 5, revisions[0].Rating)
		assert.Equal(t, "Excellent.", revisions[0].Body)
	}

	// The list reflects the edit and its revision count.
	c3, w3 := newTestContext(t, http.MethodGet, "/api/cafes/1/reviews", nil, idParam(cafe), 0, "")
	ListReviews(c3)
	requireStatus(t, w3, http.StatusOK)

	var reviews []models.Review
	decodeJSON(t, w3, &reviews)
	if assert.Len(t, reviews, 1) {
		assert.Equal(t, 3, reviews[0].Rating)
		assert.Equal(t, 1, reviews[0].RevisionCount)
	}
}

func TestIntegrationLists(t *testing.T) {
	db := integrationDB(t)

	user := testUser(t, db, "Alice")
	reviewer := testUser(t, db, "Bob")
	cafe := testCafe(t, db, "Bean There", "12 Market St")
	review := testReview(t, db, cafe, reviewer, 5, "Excellent.", pastDate(1))

	c, w := newTestContext(t, http.MethodPost, "/api/me/lists", map[string]any{"name": "Visited"}, nil, user, models.RoleUser)
	CreateList(c)
	requireStatus(t, w, http.StatusCreated)

	var list models.List
	decodeJSON(t, w, &list)

	add := func(t *testing.T, body map[string]any, fn func(*gin.Context)) {
		t.Helper()
		ctx, rec := newTestContext(t, http.MethodPost, "/api/lists/1/items", body, idParam(list.ID), user, models.RoleUser)
		fn(ctx)
		ctx.Writer.WriteHeaderNow() // flush the 204 like the engine would
		requireStatus(t, rec, http.StatusNoContent)
	}
	add(t, map[string]any{"cafe_id": cafe}, AddCafeToList)
	add(t, map[string]any{"user_id": reviewer}, AddUserToList)
	add(t, map[string]any{"review_id": review}, AddReviewToList)

	c2, w2 := newTestContext(t, http.MethodGet, "/api/lists/1", nil, idParam(list.ID), user, models.RoleUser)
	GetList(c2)
	requireStatus(t, w2, http.StatusOK)

	var detail models.ListDetail
	decodeJSON(t, w2, &detail)
	assert.Equal(t, "Visited", detail.Name)
	assert.Equal(t, 1, detail.CafeCount)
	if assert.Len(t, detail.Cafes, 1) {
		assert.Equal(t, cafe, detail.Cafes[0].ID)
	}
	if assert.Len(t, detail.Users, 1) {
		assert.Equal(t, reviewer, detail.Users[0].ID)
	}
	if assert.Len(t, detail.Reviews, 1) {
		assert.Equal(t, review, detail.Reviews[0].ID)
	}
}

func TestIntegrationUserProfile(t *testing.T) {
	db := integrationDB(t)

	user := testUser(t, db, "Alice")
	cafe := testCafe(t, db, "Bean There", "12 Market St")
	testReview(t, db, cafe, user, 5, "Excellent.", pastDate(1))

	c, w := newTestContext(t, http.MethodGet, "/api/users/1", nil, idParam(user), 0, "")
	GetUserProfile(c)
	requireStatus(t, w, http.StatusOK)

	var profile struct {
		ID          int    `json:"id"`
		DisplayName string `json:"display_name"`
		ReviewCount int    `json:"review_count"`
	}
	decodeJSON(t, w, &profile)
	assert.Equal(t, user, profile.ID)
	assert.Equal(t, "Alice", profile.DisplayName)
	assert.Equal(t, 1, profile.ReviewCount)

	c2, w2 := newTestContext(t, http.MethodGet, "/api/users/1/reviews", nil, idParam(user), 0, "")
	ListUserReviews(c2)
	requireStatus(t, w2, http.StatusOK)
	assert.Equal(t, "1", w2.Header().Get("X-Total-Count"))

	var reviews []models.Review
	decodeJSON(t, w2, &reviews)
	if assert.Len(t, reviews, 1) {
		assert.Equal(t, "Bean There", reviews[0].CafeName)
	}
}

func TestIntegrationVotes(t *testing.T) {
	db := integrationDB(t)

	author := testUser(t, db, "Alice")
	voter := testUser(t, db, "Bob")
	cafe := testCafe(t, db, "Bean There", "12 Market St")
	review := testReview(t, db, cafe, author, 5, "Excellent.", pastDate(1))

	type voteResult struct {
		HelpfulCount int  `json:"helpful_count"`
		Voted        bool `json:"voted"`
	}

	c, w := newTestContext(t, http.MethodPost, "/api/reviews/1/votes", nil, idParam(review), voter, models.RoleUser)
	VoteReview(c)
	requireStatus(t, w, http.StatusOK)

	var result voteResult
	decodeJSON(t, w, &result)
	assert.Equal(t, 1, result.HelpfulCount)
	assert.True(t, result.Voted)

	// Voting again is a no-op.
	c2, w2 := newTestContext(t, http.MethodPost, "/api/reviews/1/votes", nil, idParam(review), voter, models.RoleUser)
	VoteReview(c2)
	requireStatus(t, w2, http.StatusOK)

	var again voteResult
	decodeJSON(t, w2, &again)
	assert.Equal(t, 1, again.HelpfulCount)

	// Authors cannot vote on their own review.
	c3, w3 := newTestContext(t, http.MethodPost, "/api/reviews/1/votes", nil, idParam(review), author, models.RoleUser)
	VoteReview(c3)
	assert.Equal(t, http.StatusForbidden, w3.Code)

	// Undo.
	c4, w4 := newTestContext(t, http.MethodDelete, "/api/reviews/1/votes", nil, idParam(review), voter, models.RoleUser)
	UnvoteReview(c4)
	requireStatus(t, w4, http.StatusOK)

	var removed voteResult
	decodeJSON(t, w4, &removed)
	assert.Equal(t, 0, removed.HelpfulCount)
	assert.False(t, removed.Voted)
}

// TestIntegrationMigrationsAndSeeds proves Migrate and the seeds can run
// repeatedly against a real database, so fresh and existing databases both work.
func TestIntegrationMigrationsAndSeeds(t *testing.T) {
	db := integrationDB(t)

	assert.NoError(t, database.Migrate(), "migrations should be idempotent")
	assert.NoError(t, database.SeedTags())
	assert.NoError(t, database.SeedCafes())
	assert.NoError(t, database.SeedTags(), "seeding twice should be a no-op")
	assert.NoError(t, database.SeedCafes(), "seeding twice should be a no-op")

	var tags, cafes int
	if err := db.QueryRow(`SELECT COUNT(*) FROM tags`).Scan(&tags); err != nil {
		t.Fatalf("count tags: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM cafes`).Scan(&cafes); err != nil {
		t.Fatalf("count cafes: %v", err)
	}
	assert.Equal(t, 18, tags, "12 cafe + 3 image + 3 review tags")
	assert.Equal(t, 3, cafes, "three seeded sample cafes")
}
