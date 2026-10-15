package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"cafe-review/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCafe(t *testing.T) {
	now := time.Now()
	params := gin.Params{{Key: "id", Value: "1"}}

	t.Run("Returns the cafe with its rating summary", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`WHERE c\.id = \$1`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows(cafeCols).AddRow(1, "Bean There", "12 Market St", "Pour-overs.", now, nil, 4.5, 2, 2))
			m.ExpectQuery(`FROM cafe_tags`).WithArgs(1).WillReturnRows(tagRows())

			c, w := newTestContext(t, http.MethodGet, "/api/cafes/1", nil, params, 0, "")
			GetCafe(c)

			assert.Equal(t, http.StatusOK, w.Code)
			var cafe models.Cafe
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cafe))
			assert.Equal(t, "Bean There", cafe.Name)
			assert.Equal(t, 2, cafe.ReviewCount)
			assert.InDelta(t, 4.5, cafe.AvgRating, 0.001)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 404 when the cafe does not exist", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`WHERE c\.id = \$1`).WithArgs(1).WillReturnError(sql.ErrNoRows)

			c, w := newTestContext(t, http.MethodGet, "/api/cafes/1", nil, params, 0, "")
			GetCafe(c)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})

	t.Run("Returns 500 when the database fails", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`WHERE c\.id = \$1`).WithArgs(1).WillReturnError(errors.New("boom"))

			c, w := newTestContext(t, http.MethodGet, "/api/cafes/1", nil, params, 0, "")
			GetCafe(c)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})

	t.Run("Returns 400 for a non-numeric ID", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			c, w := newTestContext(t, http.MethodGet, "/api/cafes/abc", nil, gin.Params{{Key: "id", Value: "abc"}}, 0, "")
			GetCafe(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})
}

func TestCreateCafe(t *testing.T) {
	now := time.Now()

	t.Run("Inserts the cafe and returns the stored record", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`INSERT INTO cafes`).WithArgs("Bean There", "12 Market St", "Pour-overs.").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
			m.ExpectQuery(`WHERE c\.id = \$1`).WithArgs(3).
				WillReturnRows(sqlmock.NewRows(cafeCols).AddRow(3, "Bean There", "12 Market St", "Pour-overs.", now, nil, 0.0, 0, 0))

			body := map[string]string{"name": "  Bean There  ", "address": "12 Market St", "description": "Pour-overs."}
			c, w := newTestContext(t, http.MethodPost, "/api/cafes", body, nil, 1, models.RoleAdmin)
			CreateCafe(c)

			assert.Equal(t, http.StatusCreated, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 400 for a name longer than 255 characters", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			body := map[string]string{"name": strings.Repeat("a", 256)}
			c, w := newTestContext(t, http.MethodPost, "/api/cafes", body, nil, 1, models.RoleAdmin)
			CreateCafe(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 400 for malformed JSON", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			c, w := newTestContext(t, http.MethodPost, "/api/cafes", nil, nil, 1, models.RoleAdmin)
			c.Request.Body = http.NoBody
			CreateCafe(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})
}

func TestCloseAndReopenCafe(t *testing.T) {
	now := time.Now()
	params := gin.Params{{Key: "id", Value: "1"}}

	t.Run("Close sets closed_at and returns the cafe", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectExec(`UPDATE cafes SET closed_at = \(CURRENT_TIMESTAMP AT TIME ZONE 'UTC'\) WHERE id = \$1`).WithArgs(1).
				WillReturnResult(sqlmock.NewResult(0, 1))
			m.ExpectQuery(`WHERE c\.id = \$1`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows(cafeCols).AddRow(1, "Bean There", "12 Market St", "", now, now, 4.5, 2, 2))

			c, w := newTestContext(t, http.MethodPost, "/api/cafes/1/close", nil, params, 1, models.RoleAdmin)
			CloseCafe(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), `"closed":true`)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Reopen clears closed_at", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectExec(`UPDATE cafes SET closed_at = NULL WHERE id = \$1`).WithArgs(1).
				WillReturnResult(sqlmock.NewResult(0, 1))
			m.ExpectQuery(`WHERE c\.id = \$1`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows(cafeCols).AddRow(1, "Bean There", "12 Market St", "", now, nil, 4.5, 2, 2))

			c, w := newTestContext(t, http.MethodPost, "/api/cafes/1/reopen", nil, params, 1, models.RoleAdmin)
			ReopenCafe(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), `"closed":false`)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 404 when the cafe does not exist", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectExec(`UPDATE cafes`).WillReturnResult(sqlmock.NewResult(0, 0))

			c, w := newTestContext(t, http.MethodPost, "/api/cafes/1/close", nil, params, 1, models.RoleAdmin)
			CloseCafe(c)

			assert.Equal(t, http.StatusNotFound, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})
}
