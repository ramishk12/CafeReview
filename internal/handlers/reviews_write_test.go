package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cafe-review/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fetchReviewMock(m sqlmock.Sqlmock, id int, now time.Time) {
	m.ExpectQuery(`FROM reviews r JOIN users u ON u.id = r.user_id WHERE r.id = \$1`).WithArgs(id).
		WillReturnRows(sqlmock.NewRows(reviewCols).AddRow(id, 1, 7, "Demo User", 4, "Lovely.", now, now, now))
	m.ExpectQuery(`FROM review_images`).WillReturnRows(sqlmock.NewRows(imageCols))
}

func TestValidateReviewRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     reviewRequest
		wantMsg string
		body    string
	}{
		{"valid", reviewRequest{Rating: 5, Body: "  Great  "}, "", "Great"},
		{"rating too low", reviewRequest{Rating: 0, Body: "x"}, "Rating must be between 1 and 5", ""},
		{"rating too high", reviewRequest{Rating: 6, Body: "x"}, "Rating must be between 1 and 5", ""},
		{"blank body", reviewRequest{Rating: 3, Body: "   "}, "Review text is required", ""},
		{"body too long", reviewRequest{Rating: 3, Body: strings.Repeat("a", 2001)}, "Review text must be at most 2000 characters", ""},
		{"body at limit", reviewRequest{Rating: 3, Body: strings.Repeat("a", 2000)}, "", strings.Repeat("a", 2000)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.req
			assert.Equal(t, tt.wantMsg, validateReviewRequest(&req))
			if tt.wantMsg == "" {
				assert.Equal(t, tt.body, req.Body, "body should be trimmed")
			}
		})
	}
}

func TestCreateReview(t *testing.T) {
	now := time.Now()
	params := gin.Params{{Key: "id", Value: "1"}}

	t.Run("Creates a review for the current user and returns it", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT closed_at FROM cafes WHERE id = \$1`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"closed_at"}).AddRow(nil))
			m.ExpectQuery(`INSERT INTO reviews`).WithArgs(1, 7, 4, "Lovely.", sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
			fetchReviewMock(m, 10, now)

			c, w := newTestContext(t, http.MethodPost, "/api/cafes/1/reviews",
				map[string]any{"rating": 4, "body": "Lovely."}, params, 7, models.RoleUser)
			CreateReview(c)

			assert.Equal(t, http.StatusCreated, w.Code)
			var review models.Review
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &review))
			assert.Equal(t, 7, review.UserID)
			assert.NotNil(t, review.Images, "images should be an empty list, not null")
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 409 when the user already reviewed this cafe", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT closed_at FROM cafes WHERE id = \$1`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"closed_at"}).AddRow(nil))
			m.ExpectQuery(`INSERT INTO reviews`).WillReturnError(&pq.Error{Code: "23505"})

			c, w := newTestContext(t, http.MethodPost, "/api/cafes/1/reviews",
				map[string]any{"rating": 4, "body": "Again."}, params, 7, models.RoleUser)
			CreateReview(c)

			assert.Equal(t, http.StatusConflict, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 404 when the cafe does not exist", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT closed_at FROM cafes WHERE id = \$1`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"closed_at"}))

			c, w := newTestContext(t, http.MethodPost, "/api/cafes/1/reviews",
				map[string]any{"rating": 4, "body": "Hi."}, params, 7, models.RoleUser)
			CreateReview(c)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})

	t.Run("Returns 400 for an invalid rating without an insert", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT closed_at FROM cafes WHERE id = \$1`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"closed_at"}).AddRow(nil))
			c, w := newTestContext(t, http.MethodPost, "/api/cafes/1/reviews",
				map[string]any{"rating": 9, "body": "Hi."}, params, 7, models.RoleUser)
			CreateReview(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 500 when the insert fails for another reason", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT closed_at FROM cafes WHERE id = \$1`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"closed_at"}).AddRow(nil))
			m.ExpectQuery(`INSERT INTO reviews`).WillReturnError(errors.New("disk full"))

			c, w := newTestContext(t, http.MethodPost, "/api/cafes/1/reviews",
				map[string]any{"rating": 4, "body": "Hi."}, params, 7, models.RoleUser)
			CreateReview(c)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})
}

func TestUpdateReview(t *testing.T) {
	now := time.Now()
	params := gin.Params{{Key: "id", Value: "10"}}

	t.Run("Lets the owner update rating and text", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews WHERE id = \$1`).WithArgs(10).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
			m.ExpectQuery(`SELECT c\.closed_at FROM reviews r JOIN cafes c`).WithArgs(10).
				WillReturnRows(sqlmock.NewRows([]string{"closed_at"}).AddRow(nil))
			m.ExpectExec(`UPDATE reviews SET rating = \$1, body = \$2, visited_on = \$3`).WithArgs(2, "Changed my mind.", sqlmock.AnyArg(), 10).
				WillReturnResult(sqlmock.NewResult(0, 1))
			fetchReviewMock(m, 10, now)

			c, w := newTestContext(t, http.MethodPut, "/api/reviews/10",
				map[string]any{"rating": 2, "body": "Changed my mind."}, params, 7, models.RoleUser)
			UpdateReview(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 403 when another user tries to edit", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews WHERE id = \$1`).WithArgs(10).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(9))

			c, w := newTestContext(t, http.MethodPut, "/api/reviews/10",
				map[string]any{"rating": 2, "body": "Mine now."}, params, 7, models.RoleUser)
			UpdateReview(c)

			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 404 when the review does not exist", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews WHERE id = \$1`).WithArgs(10).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}))

			c, w := newTestContext(t, http.MethodPut, "/api/reviews/10",
				map[string]any{"rating": 2, "body": "Hi."}, params, 7, models.RoleUser)
			UpdateReview(c)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})

	t.Run("Returns 400 for blank text without updating", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews WHERE id = \$1`).WithArgs(10).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
			m.ExpectQuery(`SELECT c\.closed_at FROM reviews r JOIN cafes c`).WithArgs(10).
				WillReturnRows(sqlmock.NewRows([]string{"closed_at"}).AddRow(nil))

			c, w := newTestContext(t, http.MethodPut, "/api/reviews/10",
				map[string]any{"rating": 2, "body": "  "}, params, 7, models.RoleUser)
			UpdateReview(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})
}

func TestDeleteReview(t *testing.T) {
	params := gin.Params{{Key: "id", Value: "10"}}

	t.Run("Deletes the owner's review and its photo files", func(t *testing.T) {
		name := "review-delete-test.jpg"
		path := writeUpload(t, name)

		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews WHERE id = \$1`).WithArgs(10).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
			m.ExpectQuery(`SELECT filename FROM review_images WHERE review_id = \$1`).WithArgs(10).
				WillReturnRows(sqlmock.NewRows([]string{"filename"}).AddRow(name))
			m.ExpectExec(`DELETE FROM reviews WHERE id = \$1`).WithArgs(10).WillReturnResult(sqlmock.NewResult(0, 1))

			c, w := newTestContext(t, http.MethodDelete, "/api/reviews/10", nil, params, 7, models.RoleUser)
			DeleteReview(c)
			c.Writer.WriteHeaderNow()

			assert.Equal(t, http.StatusNoContent, w.Code)
			_, err := os.Stat(path)
			assert.True(t, os.IsNotExist(err), "photo file should be removed")
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 403 when deleting someone else's review", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews WHERE id = \$1`).WithArgs(10).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(9))

			c, w := newTestContext(t, http.MethodDelete, "/api/reviews/10", nil, params, 7, models.RoleUser)
			DeleteReview(c)

			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 500 and keeps files when the delete fails", func(t *testing.T) {
		name := "review-delete-fail.jpg"
		path := writeUpload(t, name)

		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews WHERE id = \$1`).WithArgs(10).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
			m.ExpectQuery(`SELECT filename FROM review_images`).WithArgs(10).
				WillReturnRows(sqlmock.NewRows([]string{"filename"}).AddRow(name))
			m.ExpectExec(`DELETE FROM reviews`).WillReturnError(errors.New("locked"))

			c, w := newTestContext(t, http.MethodDelete, "/api/reviews/10", nil, params, 7, models.RoleUser)
			DeleteReview(c)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
			_, err := os.Stat(filepath.Clean(path))
			assert.NoError(t, err, "files must not be removed when the review was not deleted")
		})
	})
}

func TestNormalizeVisitedOn(t *testing.T) {
	today := time.Now().UTC().Format("2006-01-02")

	t.Run("empty defaults to today", func(t *testing.T) {
		v, msg := normalizeVisitedOn("")
		assert.Equal(t, today, v)
		assert.Empty(t, msg)
	})

	t.Run("accepts a past date", func(t *testing.T) {
		v, msg := normalizeVisitedOn("2020-01-02")
		assert.Equal(t, "2020-01-02", v)
		assert.Empty(t, msg)
	})

	t.Run("rejects a future date", func(t *testing.T) {
		future := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
		_, msg := normalizeVisitedOn(future)
		assert.Contains(t, msg, "future")
	})

	t.Run("rejects a malformed date", func(t *testing.T) {
		_, msg := normalizeVisitedOn("02/01/2020")
		assert.Contains(t, msg, "YYYY-MM-DD")
	})
}

func TestCreateReviewAcceptsVisitedOn(t *testing.T) {
	now := time.Now()
	params := gin.Params{{Key: "id", Value: "1"}}

	withMockDB(t, func(m sqlmock.Sqlmock) {
		m.ExpectQuery(`SELECT closed_at FROM cafes WHERE id = \$1`).WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"closed_at"}).AddRow(nil))
		m.ExpectQuery(`INSERT INTO reviews`).WithArgs(1, 7, 4, "Lovely.", "2020-05-05").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
		fetchReviewMock(m, 10, now)

		c, w := newTestContext(t, http.MethodPost, "/api/cafes/1/reviews",
			map[string]any{"rating": 4, "body": "Lovely.", "visited_on": "2020-05-05"}, params, 7, models.RoleUser)
		CreateReview(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.NoError(t, m.ExpectationsWereMet())
	})
}

func TestCreateReviewWhenClosed(t *testing.T) {
	params := gin.Params{{Key: "id", Value: "1"}}

	withMockDB(t, func(m sqlmock.Sqlmock) {
		m.ExpectQuery(`SELECT closed_at FROM cafes WHERE id = \$1`).WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"closed_at"}).AddRow(time.Now()))

		c, w := newTestContext(t, http.MethodPost, "/api/cafes/1/reviews",
			map[string]any{"rating": 4, "body": "Hi."}, params, 7, models.RoleUser)
		CreateReview(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "closed")
		assert.NoError(t, m.ExpectationsWereMet(), "no insert should happen for a closed cafe")
	})
}

func TestUpdateReviewWhenClosed(t *testing.T) {
	params := gin.Params{{Key: "id", Value: "10"}}

	withMockDB(t, func(m sqlmock.Sqlmock) {
		m.ExpectQuery(`SELECT user_id FROM reviews WHERE id = \$1`).WithArgs(10).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
		m.ExpectQuery(`SELECT c\.closed_at FROM reviews r JOIN cafes c`).WithArgs(10).
			WillReturnRows(sqlmock.NewRows([]string{"closed_at"}).AddRow(time.Now()))

		c, w := newTestContext(t, http.MethodPut, "/api/reviews/10",
			map[string]any{"rating": 2, "body": "Edit."}, params, 7, models.RoleUser)
		UpdateReview(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "closed")
		assert.NoError(t, m.ExpectationsWereMet(), "no update should happen for a closed cafe")
	})
}
