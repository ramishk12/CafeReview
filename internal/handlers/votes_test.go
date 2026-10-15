package handlers

import (
	"net/http"
	"testing"

	"cafe-review/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestVoteReview(t *testing.T) {
	params := gin.Params{{Key: "id", Value: "10"}}

	t.Run("Marks a review helpful", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews WHERE id = \$1`).WithArgs(10).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(9))
			m.ExpectExec(`INSERT INTO review_votes`).WithArgs(10, 7).WillReturnResult(sqlmock.NewResult(0, 1))
			m.ExpectQuery(`SELECT COUNT\(\*\) FROM review_votes WHERE review_id`).WithArgs(10).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

			c, w := newTestContext(t, http.MethodPost, "/api/reviews/10/votes", nil, params, 7, models.RoleUser)
			VoteReview(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), `"helpful_count":3`)
			assert.Contains(t, w.Body.String(), `"voted":true`)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Rejects voting on your own review", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews WHERE id = \$1`).WithArgs(10).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))

			c, w := newTestContext(t, http.MethodPost, "/api/reviews/10/votes", nil, params, 7, models.RoleUser)
			VoteReview(c)
			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 404 when the review does not exist", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews WHERE id = \$1`).WithArgs(10).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}))

			c, w := newTestContext(t, http.MethodPost, "/api/reviews/10/votes", nil, params, 7, models.RoleUser)
			VoteReview(c)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})
}

func TestUnvoteReview(t *testing.T) {
	params := gin.Params{{Key: "id", Value: "10"}}

	withMockDB(t, func(m sqlmock.Sqlmock) {
		m.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM reviews WHERE id`).WithArgs(10).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
		m.ExpectExec(`DELETE FROM review_votes WHERE review_id = \$1 AND user_id = \$2`).WithArgs(10, 7).
			WillReturnResult(sqlmock.NewResult(0, 1))
		m.ExpectQuery(`SELECT COUNT\(\*\) FROM review_votes`).WithArgs(10).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		c, w := newTestContext(t, http.MethodDelete, "/api/reviews/10/votes", nil, params, 7, models.RoleUser)
		UnvoteReview(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"helpful_count":2`)
		assert.Contains(t, w.Body.String(), `"voted":false`)
		assert.NoError(t, m.ExpectationsWereMet())
	})
}

func TestListMyReviewVotes(t *testing.T) {
	t.Run("Returns the review IDs the user voted on", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM review_votes rv`).WithArgs(7).
				WillReturnRows(sqlmock.NewRows([]string{"review_id"}).AddRow(10).AddRow(11))

			c, w := newTestContext(t, http.MethodGet, "/api/me/review-votes", nil, nil, 7, models.RoleUser)
			ListMyReviewVotes(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "10")
			assert.Contains(t, w.Body.String(), "11")
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Filters by cafe", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`r\.cafe_id = \$2`).WithArgs(7, 1).
				WillReturnRows(sqlmock.NewRows([]string{"review_id"}).AddRow(10))

			c, w := newTestContext(t, http.MethodGet, "/api/me/review-votes?cafe=1", nil, nil, 7, models.RoleUser)
			ListMyReviewVotes(c)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Rejects an invalid cafe", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			c, w := newTestContext(t, http.MethodGet, "/api/me/review-votes?cafe=x", nil, nil, 7, models.RoleUser)
			ListMyReviewVotes(c)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})
}
