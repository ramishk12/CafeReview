package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cafe-review/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	pngBytes  = append([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, make([]byte, 64)...)
	jpegBytes = append([]byte{0xff, 0xd8, 0xff, 0xe0}, make([]byte, 64)...)
	webpBytes = append([]byte("RIFF\x00\x00\x00\x00WEBPVP8 "), make([]byte, 64)...)
)

// multipartImage builds a multipart request with the file in the "image" field.
func multipartImage(t *testing.T, filename string, data []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("image", filename)
	require.NoError(t, err)
	_, err = part.Write(data)
	require.NoError(t, err)
	require.NoError(t, mw.Close())
	return &buf, mw.FormDataContentType()
}

// uploadContext builds a POST /api/reviews/:id/images request with a multipart body.
func uploadContext(t *testing.T, body *bytes.Buffer, contentType string, userID int) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/reviews/10/images", body)
	c.Request.Header.Set("Content-Type", contentType)
	c.Params = gin.Params{{Key: "id", Value: "10"}}
	c.Set("userID", userID)
	c.Set("role", models.RoleUser)
	return c, w
}

func TestUploadReviewImage(t *testing.T) {
	now := time.Now()

	t.Run("Accepts a PNG, stores it under a random name, and records the row", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews WHERE id = \$1`).WithArgs(10).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
			m.ExpectQuery(`INSERT INTO review_images`).WithArgs(10, sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(55, now))

			body, ct := multipartImage(t, "cup.png", pngBytes)
			c, w := uploadContext(t, body, ct, 7)
			UploadReviewImage(c)

			require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
			var img models.ReviewImage
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &img))
			assert.Equal(t, 55, img.ID)
			assert.True(t, strings.HasPrefix(img.URL, "/uploads/"))
			assert.True(t, strings.HasSuffix(img.URL, ".png"), "extension comes from the sniffed type")
			assert.NotContains(t, img.URL, "cup", "original filename must not be used")

			stored := filepath.Join(uploadsDir(t), filepath.Base(img.URL))
			_, err := os.Stat(stored)
			assert.NoError(t, err, "file should be stored in the uploads folder")
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Accepts JPEG and WebP by sniffed content, ignoring a misleading file name", func(t *testing.T) {
		tests := []struct {
			name string
			data []byte
			ext  string
		}{
			{"jpeg", jpegBytes, ".jpg"},
			{"webp", webpBytes, ".webp"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				withMockDB(t, func(m sqlmock.Sqlmock) {
					m.ExpectQuery(`SELECT user_id FROM reviews`).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
					m.ExpectQuery(`INSERT INTO review_images`).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now))

					body, ct := multipartImage(t, "looks-like-a-gif.gif", tt.data)
					c, w := uploadContext(t, body, ct, 7)
					UploadReviewImage(c)

					require.Equal(t, http.StatusCreated, w.Code)
					var img models.ReviewImage
					require.NoError(t, json.Unmarshal(w.Body.Bytes(), &img))
					assert.True(t, strings.HasSuffix(img.URL, tt.ext))
				})
			})
		}
	})

	t.Run("Rejects a file whose content is not an allowed image, even if named .png", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews`).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))

			body, ct := multipartImage(t, "evil.png", []byte("<html><script>alert(1)</script></html>"))
			c, w := uploadContext(t, body, ct, 7)
			UploadReviewImage(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.NoError(t, m.ExpectationsWereMet(), "nothing should be inserted")
		})
	})

	t.Run("Rejects files larger than 5 MB", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews`).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))

			big := append(append([]byte{}, pngBytes...), make([]byte, maxImageBytes)...)
			body, ct := multipartImage(t, "big.png", big)
			c, w := uploadContext(t, body, ct, 7)
			UploadReviewImage(c)

			assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
		})
	})

	t.Run("Returns 400 when the image field is missing", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews`).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))

			var buf bytes.Buffer
			mw := multipart.NewWriter(&buf)
			require.NoError(t, mw.WriteField("other", "x"))
			require.NoError(t, mw.Close())

			c, w := uploadContext(t, &buf, mw.FormDataContentType(), 7)
			UploadReviewImage(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})

	t.Run("Returns 403 when the review belongs to another user", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews`).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(9))

			body, ct := multipartImage(t, "cup.png", pngBytes)
			c, w := uploadContext(t, body, ct, 7)
			UploadReviewImage(c)

			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Deletes the stored file when the database insert fails", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT user_id FROM reviews`).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
			m.ExpectQuery(`INSERT INTO review_images`).WillReturnError(errors.New("insert failed"))

			before, err := os.ReadDir(uploadsDir(t))
			require.NoError(t, err)

			body, ct := multipartImage(t, "cup.png", pngBytes)
			c, w := uploadContext(t, body, ct, 7)
			UploadReviewImage(c)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
			after, err := os.ReadDir(uploadsDir(t))
			require.NoError(t, err)
			assert.Len(t, after, len(before), "orphaned file should be removed")
		})
	})
}
