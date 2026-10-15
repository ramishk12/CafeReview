package handlers

import (
	"bytes"
	"encoding/binary"
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

// The fixtures are minimal but structurally valid, so the metadata sanitizer
// accepts them.

func jfifFixture() []byte {
	return []byte{0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00}
}

func jpegFixture() []byte {
	out := []byte{0xFF, 0xD8}
	out = append(out, jfifFixture()...)
	out = append(out, 0xFF, 0xDA, 0x00, 0x08, 0x01, 0x01, 0x00, 0x00, 0x3F, 0x00) // SOS header
	out = append(out, 0x12, 0x34, 0xFF, 0xD9)                                     // scan data + EOI
	return out
}

func pngChunkFixture(typ string, data []byte) []byte {
	out := make([]byte, 0, len(data)+12)
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(data)))
	out = append(out, length[:]...)
	out = append(out, typ...)
	out = append(out, data...)
	return append(out, 0, 0, 0, 0) // CRC is not verified by the sanitizer
}

func pngFixture() []byte {
	ihdr := []byte{0, 0, 0, 1, 0, 0, 0, 1, 8, 2, 0, 0, 0}
	out := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	out = append(out, pngChunkFixture("IHDR", ihdr)...)
	out = append(out, pngChunkFixture("IDAT", []byte{0x08, 0xD7})...)
	return append(out, pngChunkFixture("IEND", nil)...)
}

func webpFixture() []byte {
	vp8x := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	chunk := append([]byte("VP8X"), make([]byte, 4)...)
	binary.LittleEndian.PutUint32(chunk[4:8], uint32(len(vp8x)))
	chunk = append(chunk, vp8x...)
	body := append([]byte("WEBP"), chunk...)

	out := append([]byte("RIFF"), make([]byte, 4)...)
	binary.LittleEndian.PutUint32(out[4:8], uint32(len(body)))
	return append(out, body...)
}

var (
	pngBytes  = pngFixture()
	jpegBytes = jpegFixture()
	webpBytes = webpFixture()
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
