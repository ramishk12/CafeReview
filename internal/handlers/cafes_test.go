package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"cafe-review/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var cafeCols = []string{"id", "name", "address", "description", "created_at", "avg", "count"}

func cafeListRows(now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows(cafeCols).
		AddRow(1, "Bean There", "12 Market St", "Pour-overs.", now, 4.5, 2).
		AddRow(2, "Morning Crumb", "48 Elm Ave", "Pastries.", now, 0.0, 0)
}

func countRow(n int) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"count"}).AddRow(n)
}

func TestListCafes(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		target         string
		setupMock      func(sqlmock.Sqlmock)
		expectedStatus int
		expectedTotal  string
		expectedCount  int
	}{
		{
			name:   "Returns every cafe and the total when no paging is requested",
			target: "/api/cafes",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT\(\*\) FROM cafes c`).WillReturnRows(countRow(2))
				m.ExpectQuery(`ORDER BY c\.name ASC, c\.id ASC`).WillReturnRows(cafeListRows(now))
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  "2",
			expectedCount:  2,
		},
		{
			name:   "Searches name and address case-insensitively",
			target: "/api/cafes?q=bean",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`WHERE c\.name ILIKE \$1 OR c\.address ILIKE \$1`).
					WithArgs("%bean%").
					WillReturnRows(countRow(1))
				m.ExpectQuery(`WHERE c\.name ILIKE \$1 OR c\.address ILIKE \$1`).
					WithArgs("%bean%").
					WillReturnRows(sqlmock.NewRows(cafeCols).AddRow(1, "Bean There", "12 Market St", "", now, 4.5, 2))
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  "1",
			expectedCount:  1,
		},
		{
			name:   "Escapes LIKE wildcards in the search term",
			target: "/api/cafes?q=" + url.QueryEscape("50%_off"),
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`COUNT`).WithArgs(`%50\%\_off%`).WillReturnRows(countRow(0))
				m.ExpectQuery(`ILIKE`).WithArgs(`%50\%\_off%`).WillReturnRows(sqlmock.NewRows(cafeCols))
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  "0",
			expectedCount:  0,
		},
		{
			name:   "Sorts by newest",
			target: "/api/cafes?sort=newest",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`COUNT`).WillReturnRows(countRow(2))
				m.ExpectQuery(`ORDER BY c\.created_at DESC, c\.id DESC`).WillReturnRows(cafeListRows(now))
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  "2",
			expectedCount:  2,
		},
		{
			name:   "Pages results with limit and offset",
			target: "/api/cafes?limit=2&offset=4",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`COUNT`).WillReturnRows(countRow(5))
				m.ExpectQuery(`LIMIT \$1 OFFSET \$2`).WithArgs(2, 4).WillReturnRows(cafeListRows(now))
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  "5",
			expectedCount:  2,
		},
		{
			name:           "Rejects an unknown sort",
			target:         "/api/cafes?sort=random",
			setupMock:      func(sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Rejects a limit above 100",
			target:         "/api/cafes?limit=500",
			setupMock:      func(sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Rejects a negative offset",
			target:         "/api/cafes?limit=10&offset=-1",
			setupMock:      func(sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Rejects a search longer than 100 characters",
			target:         "/api/cafes?q=" + strings.Repeat("a", 101),
			setupMock:      func(sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Returns 500 when the count query fails",
			target: "/api/cafes",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`COUNT`).WillReturnError(errors.New("connection lost"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockDB(t, func(mock sqlmock.Sqlmock) {
				tt.setupMock(mock)
				c, w := newTestContext(t, http.MethodGet, tt.target, nil, nil, 0, "")
				ListCafes(c)

				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.expectedTotal != "" {
					assert.Equal(t, tt.expectedTotal, w.Header().Get("X-Total-Count"))
				}
				if tt.expectedStatus == http.StatusOK {
					var cafes []models.Cafe
					assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &cafes))
					assert.Len(t, cafes, tt.expectedCount)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		})
	}
}

func TestUpdateCafe(t *testing.T) {
	now := time.Now()
	params := gin.Params{{Key: "id", Value: "1"}}

	t.Run("Updates a cafe and returns it", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectExec(`UPDATE cafes SET name = \$1, address = \$2, description = \$3 WHERE id = \$4`).
				WithArgs("Bean There", "12 Market St", "Now with seats.", 1).
				WillReturnResult(sqlmock.NewResult(0, 1))
			m.ExpectQuery(`WHERE c\.id = \$1`).WithArgs(1).WillReturnRows(
				sqlmock.NewRows(cafeCols).AddRow(1, "Bean There", "12 Market St", "Now with seats.", now, 4.5, 2),
			)

			body := map[string]string{"name": "Bean There", "address": "12 Market St", "description": "Now with seats."}
			c, w := newTestContext(t, http.MethodPut, "/api/cafes/1", body, params, 1, "admin")
			UpdateCafe(c)

			assert.Equal(t, http.StatusOK, w.Code)
			var cafe models.Cafe
			assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &cafe))
			assert.Equal(t, "Now with seats.", cafe.Description)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 404 when the cafe does not exist", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectExec(`UPDATE cafes`).WillReturnResult(sqlmock.NewResult(0, 0))

			body := map[string]string{"name": "Ghost Cafe"}
			c, w := newTestContext(t, http.MethodPut, "/api/cafes/1", body, params, 1, "admin")
			UpdateCafe(c)

			assert.Equal(t, http.StatusNotFound, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 400 for an empty name without touching the database", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			body := map[string]string{"name": "   "}
			c, w := newTestContext(t, http.MethodPut, "/api/cafes/1", body, params, 1, "admin")
			UpdateCafe(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 400 for an invalid ID", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			body := map[string]string{"name": "Bean There"}
			c, w := newTestContext(t, http.MethodPut, "/api/cafes/0", body, gin.Params{{Key: "id", Value: "0"}}, 1, "admin")
			UpdateCafe(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})
}

func TestDeleteCafe(t *testing.T) {
	params := gin.Params{{Key: "id", Value: "1"}}

	t.Run("Deletes the cafe and its photo files", func(t *testing.T) {
		name := "cafe-delete-photo.jpg"
		path := writeUpload(t, name)

		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT ri\.filename`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"filename"}).AddRow(name))
			m.ExpectExec(`DELETE FROM cafes WHERE id = \$1`).WithArgs(1).
				WillReturnResult(sqlmock.NewResult(0, 1))

			c, w := newTestContext(t, http.MethodDelete, "/api/cafes/1", nil, params, 1, "admin")
			DeleteCafe(c)

			c.Writer.WriteHeaderNow() // gin writes the status only when the response is finalized
			assert.Equal(t, http.StatusNoContent, w.Code)
			_, err := os.Stat(path)
			assert.True(t, os.IsNotExist(err), "photo file should be removed")
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 404 when the cafe does not exist", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT ri\.filename`).WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"filename"}))
			m.ExpectExec(`DELETE FROM cafes WHERE id = \$1`).WithArgs(1).
				WillReturnResult(sqlmock.NewResult(0, 0))

			c, w := newTestContext(t, http.MethodDelete, "/api/cafes/1", nil, params, 1, "admin")
			DeleteCafe(c)

			assert.Equal(t, http.StatusNotFound, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 400 for an invalid ID", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			c, w := newTestContext(t, http.MethodDelete, "/api/cafes/abc", nil, gin.Params{{Key: "id", Value: "abc"}}, 1, "admin")
			DeleteCafe(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})
}
