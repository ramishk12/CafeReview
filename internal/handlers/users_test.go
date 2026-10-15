package handlers

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"cafe-review/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserProfile(t *testing.T) {
	now := time.Now()
	params := gin.Params{{Key: "id", Value: "7"}}

	t.Run("Returns the public profile", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM users u`).WithArgs(7).
				WillReturnRows(sqlmock.NewRows([]string{"id", "display_name", "created_at", "count"}).AddRow(7, "Demo User", now, 3))

			c, w := newTestContext(t, http.MethodGet, "/api/users/7", nil, params, 0, "")
			GetUserProfile(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.NotContains(t, w.Body.String(), "email")
			var profile struct {
				DisplayName string `json:"display_name"`
				ReviewCount int    `json:"review_count"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &profile))
			assert.Equal(t, "Demo User", profile.DisplayName)
			assert.Equal(t, 3, profile.ReviewCount)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 404 when the user does not exist", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM users u`).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"id", "display_name", "created_at", "count"}))

			c, w := newTestContext(t, http.MethodGet, "/api/users/7", nil, params, 0, "")
			GetUserProfile(c)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})
}

func TestListUserReviews(t *testing.T) {
	now := time.Now()
	params := gin.Params{{Key: "id", Value: "7"}}

	t.Run("Returns the user's reviews with the cafe name", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT EXISTS`).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			m.ExpectQuery(`SELECT COUNT\(\*\) FROM reviews WHERE user_id`).WithArgs(7).WillReturnRows(countRow(1))
			m.ExpectQuery(`JOIN cafes c ON c.id = r.cafe_id`).WithArgs(7).
				WillReturnRows(sqlmock.NewRows([]string{"id", "cafe_id", "name", "user_id", "display_name", "rating", "body", "visited_on", "created_at", "updated_at"}).
					AddRow(10, 1, "Bean There", 7, "Demo User", 5, "Great.", now, now, now))
			m.ExpectQuery(`FROM review_images`).WillReturnRows(sqlmock.NewRows(imageCols))
			m.ExpectQuery(`FROM review_tags`).WillReturnRows(sqlmock.NewRows([]string{"review_id", "id", "name", "category", "scope"}))

			c, w := newTestContext(t, http.MethodGet, "/api/users/7/reviews", nil, params, 0, "")
			ListUserReviews(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "1", w.Header().Get("X-Total-Count"))
			var reviews []models.Review
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &reviews))
			require.Len(t, reviews, 1)
			assert.Equal(t, "Bean There", reviews[0].CafeName)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 404 when the user does not exist", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT EXISTS`).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

			c, w := newTestContext(t, http.MethodGet, "/api/users/7/reviews", nil, params, 0, "")
			ListUserReviews(c)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})
}
