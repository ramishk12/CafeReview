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
)

var (
	reviewCols = []string{"id", "cafe_id", "user_id", "display_name", "rating", "body", "created_at", "updated_at"}
	imageCols  = []string{"id", "review_id", "filename", "created_at"}
)

func reviewListRows(now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows(reviewCols).
		AddRow(10, 1, 7, "Demo User", 5, "Excellent.", now, now).
		AddRow(11, 1, 8, "Sam Lee", 3, "Fine.", now, now)
}

func TestListReviews(t *testing.T) {
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
			name:   "Lists reviews newest first with the total",
			target: "/api/cafes/1/reviews",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT\(\*\) FROM reviews WHERE cafe_id = \$1`).WithArgs(1).WillReturnRows(countRow(2))
				m.ExpectQuery(`ORDER BY r\.created_at DESC, r\.id DESC`).WithArgs(1).WillReturnRows(reviewListRows(now))
				m.ExpectQuery(`FROM review_images`).WillReturnRows(sqlmock.NewRows(imageCols))
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  "2",
			expectedCount:  2,
		},
		{
			name:   "Sorts by highest rating",
			target: "/api/cafes/1/reviews?sort=highest",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`COUNT`).WithArgs(1).WillReturnRows(countRow(2))
				m.ExpectQuery(`ORDER BY r\.rating DESC, r\.created_at DESC`).WithArgs(1).WillReturnRows(reviewListRows(now))
				m.ExpectQuery(`FROM review_images`).WillReturnRows(sqlmock.NewRows(imageCols))
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  "2",
			expectedCount:  2,
		},
		{
			name:   "Sorts by lowest rating",
			target: "/api/cafes/1/reviews?sort=lowest",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`COUNT`).WithArgs(1).WillReturnRows(countRow(2))
				m.ExpectQuery(`ORDER BY r\.rating ASC, r\.created_at DESC`).WithArgs(1).WillReturnRows(reviewListRows(now))
				m.ExpectQuery(`FROM review_images`).WillReturnRows(sqlmock.NewRows(imageCols))
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  "2",
			expectedCount:  2,
		},
		{
			name:   "Pages reviews with limit and offset",
			target: "/api/cafes/1/reviews?limit=1&offset=1",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`COUNT`).WithArgs(1).WillReturnRows(countRow(2))
				m.ExpectQuery(`LIMIT \$2 OFFSET \$3`).WithArgs(1, 1, 1).
					WillReturnRows(sqlmock.NewRows(reviewCols).AddRow(11, 1, 8, "Sam Lee", 3, "Fine.", now, now))
				m.ExpectQuery(`FROM review_images`).WillReturnRows(sqlmock.NewRows(imageCols))
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  "2",
			expectedCount:  1,
		},
		{
			name:           "Rejects an unknown sort",
			target:         "/api/cafes/1/reviews?sort=cheapest",
			setupMock:      func(sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Rejects a limit of zero",
			target:         "/api/cafes/1/reviews?limit=0",
			setupMock:      func(sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockDB(t, func(mock sqlmock.Sqlmock) {
				tt.setupMock(mock)
				c, w := newTestContext(t, http.MethodGet, tt.target, nil, gin.Params{{Key: "id", Value: "1"}}, 0, "")
				ListReviews(c)

				assert.Equal(t, tt.expectedStatus, w.Code)
				if tt.expectedTotal != "" {
					assert.Equal(t, tt.expectedTotal, w.Header().Get("X-Total-Count"))
				}
				if tt.expectedStatus == http.StatusOK {
					var reviews []models.Review
					assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &reviews))
					assert.Len(t, reviews, tt.expectedCount)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		})
	}
}
