package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"cafe-review/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTags(t *testing.T) {
	t.Run("Returns approved tags", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT id, name, category, scope FROM tags WHERE status = 'approved'`).
				WillReturnRows(sqlmock.NewRows([]string{"id", "name", "category", "scope"}).
					AddRow(1, "Quiet", "vibe", "cafe"))

			c, w := newTestContext(t, http.MethodGet, "/api/tags", nil, nil, 0, "")
			ListTags(c)

			assert.Equal(t, http.StatusOK, w.Code)
			var tags []models.Tag
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tags))
			require.Len(t, tags, 1)
			assert.Equal(t, "Quiet", tags[0].Name)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Filters by scope", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`AND scope = \$1`).WithArgs("cafe").
				WillReturnRows(sqlmock.NewRows([]string{"id", "name", "category", "scope"}))

			c, w := newTestContext(t, http.MethodGet, "/api/tags?scope=cafe", nil, nil, 0, "")
			ListTags(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Rejects an unknown scope", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			c, w := newTestContext(t, http.MethodGet, "/api/tags?scope=nope", nil, nil, 0, "")
			ListTags(c)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 500 on a query error", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM tags`).WillReturnError(errors.New("boom"))

			c, w := newTestContext(t, http.MethodGet, "/api/tags", nil, nil, 0, "")
			ListTags(c)
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})
}

func TestSetCafeTags(t *testing.T) {
	now := time.Now()
	params := gin.Params{{Key: "id", Value: "1"}}

	t.Run("Replaces the cafe's tags", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT EXISTS`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			m.ExpectQuery(`SELECT COUNT\(\*\) FROM tags WHERE scope = 'cafe'`).WithArgs(1, 2).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
			m.ExpectBegin()
			m.ExpectExec(`DELETE FROM cafe_tags WHERE cafe_id = \$1`).WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
			m.ExpectExec(`INSERT INTO cafe_tags`).WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 1))
			m.ExpectExec(`INSERT INTO cafe_tags`).WithArgs(1, 2).WillReturnResult(sqlmock.NewResult(0, 1))
			m.ExpectCommit()
			m.ExpectQuery(`WHERE c\.id = \$1`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows(cafeCols).AddRow(1, "Bean There", "12 Market St", "", now, nil, 4.5, 2, 2))
			m.ExpectQuery(`FROM cafe_tags`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"cafe_id", "id", "name", "category", "scope"}).
					AddRow(1, 1, "Quiet", "vibe", "cafe"))

			c, w := newTestContext(t, http.MethodPut, "/api/cafes/1/tags", map[string]any{"tag_ids": []int{1, 2}}, params, 1, models.RoleAdmin)
			SetCafeTags(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), `"Quiet"`)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 400 for an invalid tag", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT EXISTS`).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			m.ExpectQuery(`SELECT COUNT\(\*\) FROM tags`).WithArgs(9).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

			c, w := newTestContext(t, http.MethodPut, "/api/cafes/1/tags", map[string]any{"tag_ids": []int{9}}, params, 1, models.RoleAdmin)
			SetCafeTags(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 404 when the cafe does not exist", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT EXISTS`).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

			c, w := newTestContext(t, http.MethodPut, "/api/cafes/1/tags", map[string]any{"tag_ids": []int{}}, params, 1, models.RoleAdmin)
			SetCafeTags(c)

			assert.Equal(t, http.StatusNotFound, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})
}
