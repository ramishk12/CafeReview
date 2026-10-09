package handlers

import (
	"net/http"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDeleteReviewImage(t *testing.T) {
	params := gin.Params{{Key: "id", Value: "3"}, {Key: "imageId", Value: "5"}}

	t.Run("Removes a photo from the current user's review and deletes the file", func(t *testing.T) {
		name := "review-photo-delete.jpg"
		path := writeUpload(t, name)

		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews WHERE id = \$1`).WithArgs(3).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
			m.ExpectQuery(`DELETE FROM review_images WHERE id = \$1 AND review_id = \$2 RETURNING filename`).
				WithArgs(5, 3).
				WillReturnRows(sqlmock.NewRows([]string{"filename"}).AddRow(name))

			c, w := newTestContext(t, http.MethodDelete, "/api/reviews/3/images/5", nil, params, 7, "user")
			DeleteReviewImage(c)

			c.Writer.WriteHeaderNow() // gin writes the status only when the response is finalized
			assert.Equal(t, http.StatusNoContent, w.Code)
			_, err := os.Stat(path)
			assert.True(t, os.IsNotExist(err), "image file should be removed")
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 403 when the review belongs to another user", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews WHERE id = \$1`).WithArgs(3).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(9))

			c, w := newTestContext(t, http.MethodDelete, "/api/reviews/3/images/5", nil, params, 7, "user")
			DeleteReviewImage(c)

			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 404 when the review does not exist", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews WHERE id = \$1`).WithArgs(3).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}))

			c, w := newTestContext(t, http.MethodDelete, "/api/reviews/3/images/5", nil, params, 7, "user")
			DeleteReviewImage(c)

			assert.Equal(t, http.StatusNotFound, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 404 when the photo is not part of the review", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews WHERE id = \$1`).WithArgs(3).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
			m.ExpectQuery(`DELETE FROM review_images`).WithArgs(5, 3).
				WillReturnRows(sqlmock.NewRows([]string{"filename"}))

			c, w := newTestContext(t, http.MethodDelete, "/api/reviews/3/images/5", nil, params, 7, "user")
			DeleteReviewImage(c)

			assert.Equal(t, http.StatusNotFound, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 400 for an invalid image ID", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			badParams := gin.Params{{Key: "id", Value: "3"}, {Key: "imageId", Value: "x"}}
			c, w := newTestContext(t, http.MethodDelete, "/api/reviews/3/images/x", nil, badParams, 7, "user")
			DeleteReviewImage(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})
}
