package handlers

import (
	"net/http"
	"testing"
	"time"

	"cafe-review/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var listCols = []string{"id", "name", "is_public", "created_at", "count"}

func TestListMyLists(t *testing.T) {
	now := time.Now()

	t.Run("Returns the user's lists with counts", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM lists l`).WithArgs(7).
				WillReturnRows(sqlmock.NewRows(listCols).AddRow(1, "Visited", false, now, 3))

			c, w := newTestContext(t, http.MethodGet, "/api/me/lists", nil, nil, 7, models.RoleUser)
			ListMyLists(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "Visited")
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Reports whether a cafe is in each list", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM lists l`).WithArgs(7).
				WillReturnRows(sqlmock.NewRows(listCols).AddRow(1, "Visited", false, now, 3))
			m.ExpectQuery(`SELECT list_id FROM list_cafes WHERE cafe_id`).WithArgs(5).
				WillReturnRows(sqlmock.NewRows([]string{"list_id"}).AddRow(1))

			c, w := newTestContext(t, http.MethodGet, "/api/me/lists?cafe=5", nil, nil, 7, models.RoleUser)
			ListMyLists(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), `"contains_cafe":true`)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})
}

func TestCreateList(t *testing.T) {
	now := time.Now()

	t.Run("Creates a list", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT COUNT\(\*\) FROM lists WHERE user_id`).WithArgs(7).WillReturnRows(countRow(1))
			m.ExpectQuery(`INSERT INTO lists`).WithArgs(7, "Visited").
				WillReturnRows(sqlmock.NewRows([]string{"id", "name", "is_public", "created_at"}).AddRow(2, "Visited", false, now))

			c, w := newTestContext(t, http.MethodPost, "/api/me/lists", map[string]any{"name": "Visited"}, nil, 7, models.RoleUser)
			CreateList(c)

			assert.Equal(t, http.StatusCreated, w.Code)
			assert.Contains(t, w.Body.String(), "Visited")
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Rejects an empty name without touching the database", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			c, w := newTestContext(t, http.MethodPost, "/api/me/lists", map[string]any{"name": "   "}, nil, 7, models.RoleUser)
			CreateList(c)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Rejects a new list when the cap is reached", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT COUNT`).WithArgs(7).WillReturnRows(countRow(maxListsPerUser))

			c, w := newTestContext(t, http.MethodPost, "/api/me/lists", map[string]any{"name": "More"}, nil, 7, models.RoleUser)
			CreateList(c)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})
}

func TestGetList(t *testing.T) {
	now := time.Now()
	params := gin.Params{{Key: "id", Value: "1"}}

	withMockDB(t, func(m sqlmock.Sqlmock) {
		m.ExpectQuery(`SELECT user_id FROM lists WHERE id`).WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
		m.ExpectQuery(`SELECT l.id, l.name, l.is_public, l.created_at, COUNT`).WithArgs(1).
			WillReturnRows(sqlmock.NewRows(listCols).AddRow(1, "Visited", false, now, 1))
		m.ExpectQuery(`FROM list_cafes lc`).WithArgs(1).
			WillReturnRows(sqlmock.NewRows(cafeCols).AddRow(1, "Bean There", "12 Market St", "", now, nil, nil, nil, nil, 4.5, 2, 2))
		m.ExpectQuery(`FROM cafe_tags`).WillReturnRows(tagRows())
		m.ExpectQuery(`FROM list_users lu`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "display_name"}))
		m.ExpectQuery(`FROM list_reviews lr`).WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "cafe_id", "name", "user_id", "display_name", "rating", "body", "visited_on", "created_at", "updated_at"}))

		c, w := newTestContext(t, http.MethodGet, "/api/lists/1", nil, params, 7, models.RoleUser)
		GetList(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Bean There")
		assert.NoError(t, m.ExpectationsWereMet())
	})
}

func TestUpdateList(t *testing.T) {
	now := time.Now()
	params := gin.Params{{Key: "id", Value: "1"}}

	withMockDB(t, func(m sqlmock.Sqlmock) {
		m.ExpectQuery(`SELECT user_id FROM lists WHERE id`).WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
		m.ExpectExec(`UPDATE lists SET name = \$1 WHERE id = \$2`).WithArgs("Favourites", 1).WillReturnResult(sqlmock.NewResult(0, 1))
		m.ExpectQuery(`SELECT l.id, l.name, l.is_public, l.created_at, COUNT`).WithArgs(1).
			WillReturnRows(sqlmock.NewRows(listCols).AddRow(1, "Favourites", false, now, 0))

		c, w := newTestContext(t, http.MethodPut, "/api/lists/1", map[string]any{"name": "Favourites"}, params, 7, models.RoleUser)
		UpdateList(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Favourites")
		assert.NoError(t, m.ExpectationsWereMet())
	})
}

func TestDeleteList(t *testing.T) {
	params := gin.Params{{Key: "id", Value: "1"}}

	withMockDB(t, func(m sqlmock.Sqlmock) {
		m.ExpectQuery(`SELECT user_id FROM lists WHERE id`).WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
		m.ExpectExec(`DELETE FROM lists WHERE id = \$1`).WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))

		c, w := newTestContext(t, http.MethodDelete, "/api/lists/1", nil, params, 7, models.RoleUser)
		DeleteList(c)
		c.Writer.WriteHeaderNow()

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.NoError(t, m.ExpectationsWereMet())
	})
}

func TestAddCafeToList(t *testing.T) {
	params := gin.Params{{Key: "id", Value: "1"}}

	t.Run("Adds a cafe", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM lists WHERE id`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
			m.ExpectQuery(`SELECT COUNT\(\*\) FROM list_cafes WHERE list_id`).WithArgs(1).WillReturnRows(countRow(0))
			m.ExpectExec(`INSERT INTO list_cafes`).WithArgs(1, 5).WillReturnResult(sqlmock.NewResult(0, 1))

			c, w := newTestContext(t, http.MethodPost, "/api/lists/1/cafes", map[string]any{"cafe_id": 5}, params, 7, models.RoleUser)
			AddCafeToList(c)
			c.Writer.WriteHeaderNow()

			assert.Equal(t, http.StatusNoContent, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 409 when the cafe is already in the list", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM lists WHERE id`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
			m.ExpectQuery(`SELECT COUNT`).WithArgs(1).WillReturnRows(countRow(1))
			m.ExpectExec(`INSERT INTO list_cafes`).WillReturnError(&pq.Error{Code: "23505"})

			c, w := newTestContext(t, http.MethodPost, "/api/lists/1/cafes", map[string]any{"cafe_id": 5}, params, 7, models.RoleUser)
			AddCafeToList(c)
			assert.Equal(t, http.StatusConflict, w.Code)
		})
	})

	t.Run("Returns 404 when the cafe does not exist", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM lists WHERE id`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
			m.ExpectQuery(`SELECT COUNT`).WithArgs(1).WillReturnRows(countRow(0))
			m.ExpectExec(`INSERT INTO list_cafes`).WillReturnError(&pq.Error{Code: "23503"})

			c, w := newTestContext(t, http.MethodPost, "/api/lists/1/cafes", map[string]any{"cafe_id": 5}, params, 7, models.RoleUser)
			AddCafeToList(c)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})
}

func TestRemoveCafeFromList(t *testing.T) {
	params := gin.Params{{Key: "id", Value: "1"}, {Key: "cafeId", Value: "5"}}

	t.Run("Removes a cafe", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM lists WHERE id`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
			m.ExpectExec(`DELETE FROM list_cafes WHERE list_id = \$1 AND cafe_id = \$2`).WithArgs(1, 5).
				WillReturnResult(sqlmock.NewResult(0, 1))

			c, w := newTestContext(t, http.MethodDelete, "/api/lists/1/cafes/5", nil, params, 7, models.RoleUser)
			RemoveCafeFromList(c)
			c.Writer.WriteHeaderNow()

			assert.Equal(t, http.StatusNoContent, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 404 when the cafe is not in the list", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM lists WHERE id`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
			m.ExpectExec(`DELETE FROM list_cafes`).WithArgs(1, 5).WillReturnResult(sqlmock.NewResult(0, 0))

			c, w := newTestContext(t, http.MethodDelete, "/api/lists/1/cafes/5", nil, params, 7, models.RoleUser)
			RemoveCafeFromList(c)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})
}

func TestListOwnership(t *testing.T) {
	params := gin.Params{{Key: "id", Value: "1"}}

	t.Run("Returns 403 for another user's list", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM lists WHERE id`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(9))

			c, w := newTestContext(t, http.MethodGet, "/api/lists/1", nil, params, 7, models.RoleUser)
			GetList(c)
			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	})

	t.Run("Returns 404 when the list does not exist", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM lists WHERE id`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}))

			c, w := newTestContext(t, http.MethodGet, "/api/lists/1", nil, params, 7, models.RoleUser)
			GetList(c)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})
}

func TestAddUserToList(t *testing.T) {
	params := gin.Params{{Key: "id", Value: "1"}}

	t.Run("Adds a reviewer", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM lists WHERE id`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
			m.ExpectQuery(`SELECT COUNT\(\*\) FROM list_users WHERE list_id`).WithArgs(1).WillReturnRows(countRow(0))
			m.ExpectExec(`INSERT INTO list_users`).WithArgs(1, 9).WillReturnResult(sqlmock.NewResult(0, 1))

			c, w := newTestContext(t, http.MethodPost, "/api/lists/1/users", map[string]any{"user_id": 9}, params, 7, models.RoleUser)
			AddUserToList(c)
			c.Writer.WriteHeaderNow()

			assert.Equal(t, http.StatusNoContent, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 409 when already in the list", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM lists WHERE id`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
			m.ExpectQuery(`SELECT COUNT`).WithArgs(1).WillReturnRows(countRow(0))
			m.ExpectExec(`INSERT INTO list_users`).WillReturnError(&pq.Error{Code: "23505"})

			c, w := newTestContext(t, http.MethodPost, "/api/lists/1/users", map[string]any{"user_id": 9}, params, 7, models.RoleUser)
			AddUserToList(c)
			assert.Equal(t, http.StatusConflict, w.Code)
		})
	})
}

func TestRemoveUserFromList(t *testing.T) {
	params := gin.Params{{Key: "id", Value: "1"}, {Key: "userId", Value: "9"}}

	withMockDB(t, func(m sqlmock.Sqlmock) {
		m.ExpectQuery(`SELECT user_id FROM lists WHERE id`).WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
		m.ExpectExec(`DELETE FROM list_users WHERE list_id = \$1 AND user_id = \$2`).WithArgs(1, 9).
			WillReturnResult(sqlmock.NewResult(0, 1))

		c, w := newTestContext(t, http.MethodDelete, "/api/lists/1/users/9", nil, params, 7, models.RoleUser)
		RemoveUserFromList(c)
		c.Writer.WriteHeaderNow()

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.NoError(t, m.ExpectationsWereMet())
	})
}

func TestAddReviewToList(t *testing.T) {
	params := gin.Params{{Key: "id", Value: "1"}}

	withMockDB(t, func(m sqlmock.Sqlmock) {
		m.ExpectQuery(`SELECT user_id FROM lists WHERE id`).WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
		m.ExpectQuery(`SELECT COUNT\(\*\) FROM list_reviews WHERE list_id`).WithArgs(1).WillReturnRows(countRow(0))
		m.ExpectExec(`INSERT INTO list_reviews`).WithArgs(1, 10).WillReturnResult(sqlmock.NewResult(0, 1))

		c, w := newTestContext(t, http.MethodPost, "/api/lists/1/reviews", map[string]any{"review_id": 10}, params, 7, models.RoleUser)
		AddReviewToList(c)
		c.Writer.WriteHeaderNow()

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.NoError(t, m.ExpectationsWereMet())
	})
}

func TestRemoveReviewFromList(t *testing.T) {
	params := gin.Params{{Key: "id", Value: "1"}, {Key: "reviewId", Value: "10"}}

	withMockDB(t, func(m sqlmock.Sqlmock) {
		m.ExpectQuery(`SELECT user_id FROM lists WHERE id`).WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
		m.ExpectExec(`DELETE FROM list_reviews WHERE list_id = \$1 AND review_id = \$2`).WithArgs(1, 10).
			WillReturnResult(sqlmock.NewResult(0, 1))

		c, w := newTestContext(t, http.MethodDelete, "/api/lists/1/reviews/10", nil, params, 7, models.RoleUser)
		RemoveReviewFromList(c)
		c.Writer.WriteHeaderNow()

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.NoError(t, m.ExpectationsWereMet())
	})
}
