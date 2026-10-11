package auth

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cafe-review/internal/database"
	"cafe-review/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var userCols = []string{"id", "email", "display_name", "password_hash", "role", "created_at"}

// withMockDB swaps database.DB for a sqlmock connection while f runs.
func withMockDB(t *testing.T, f func(mock sqlmock.Sqlmock)) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	original := database.DB
	database.DB = db
	defer func() { database.DB = original }()

	f(mock)
}

// postJSON runs handler with a JSON body and returns the recorded response.
func postJSON(t *testing.T, handler gin.HandlerFunc, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	data, err := json.Marshal(body)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, path, bytes.NewReader(data))
	c.Request.Header.Set("Content-Type", "application/json")
	handler(c)
	return w
}

func TestRegister(t *testing.T) {
	useSecret(t, "test-secret-for-register")
	now := time.Now()

	t.Run("Creates a user, normalises the email, and returns a token without the hash", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`INSERT INTO users`).
				WithArgs("new@example.com", "Ann", sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows(userCols).AddRow(5, "new@example.com", "Ann", "hash", "user", now))

			w := postJSON(t, Register, "/api/auth/register", map[string]string{
				"email": "  New@Example.com ", "password": "longenough", "display_name": " Ann ",
			})

			assert.Equal(t, http.StatusCreated, w.Code)
			assert.NotContains(t, w.Body.String(), "password_hash")
			assert.NotContains(t, w.Body.String(), "hash\"")
			var resp authResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.NotEmpty(t, resp.Token)
			assert.Equal(t, "new@example.com", resp.User.Email)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 409 when the email is already registered", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`INSERT INTO users`).WillReturnError(&pq.Error{Code: "23505"})

			w := postJSON(t, Register, "/api/auth/register", map[string]string{
				"email": "taken@example.com", "password": "longenough", "display_name": "Ann",
			})

			assert.Equal(t, http.StatusConflict, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Rejects invalid input without touching the database", func(t *testing.T) {
		tests := []struct {
			name string
			body map[string]string
		}{
			{"bad email", map[string]string{"email": "not-an-email", "password": "longenough", "display_name": "Ann"}},
			{"short password", map[string]string{"email": "a@example.com", "password": "short", "display_name": "Ann"}},
			{"blank display name", map[string]string{"email": "a@example.com", "password": "longenough", "display_name": "   "}},
			{"display name too long", map[string]string{"email": "a@example.com", "password": "longenough", "display_name": strings.Repeat("a", 101)}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				withMockDB(t, func(m sqlmock.Sqlmock) {
					w := postJSON(t, Register, "/api/auth/register", tt.body)
					assert.Equal(t, http.StatusBadRequest, w.Code)
					assert.NoError(t, m.ExpectationsWereMet())
				})
			})
		}
	})

	t.Run("Returns 400 for a malformed JSON body", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString("{"))
			Register(c)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})
}

func TestLogin(t *testing.T) {
	useSecret(t, "test-secret-for-login")
	now := time.Now()
	hash, err := HashPassword("correct-password")
	require.NoError(t, err)

	t.Run("Returns a token for correct credentials", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM users WHERE email = \$1`).WithArgs("demo@example.com").
				WillReturnRows(sqlmock.NewRows(userCols).AddRow(2, "demo@example.com", "Demo", hash, "user", now))

			w := postJSON(t, Login, "/api/auth/login", map[string]string{"email": "Demo@Example.com", "password": "correct-password"})

			assert.Equal(t, http.StatusOK, w.Code)
			var resp authResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.NotEmpty(t, resp.Token)
			assert.Equal(t, 2, resp.User.ID)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns the same 401 for a wrong password and an unknown email", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM users WHERE email`).WillReturnRows(sqlmock.NewRows(userCols).AddRow(2, "demo@example.com", "Demo", hash, "user", now))
			wrong := postJSON(t, Login, "/api/auth/login", map[string]string{"email": "demo@example.com", "password": "wrong-password"})

			m.ExpectQuery(`FROM users WHERE email`).WillReturnError(sql.ErrNoRows)
			unknown := postJSON(t, Login, "/api/auth/login", map[string]string{"email": "nobody@example.com", "password": "x"})

			assert.Equal(t, http.StatusUnauthorized, wrong.Code)
			assert.Equal(t, http.StatusUnauthorized, unknown.Code)
			assert.Equal(t, wrong.Body.String(), unknown.Body.String(), "responses must not reveal which check failed")
		})
	})

	t.Run("Returns 500 when the database fails", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM users WHERE email`).WillReturnError(errors.New("connection reset"))

			w := postJSON(t, Login, "/api/auth/login", map[string]string{"email": "a@example.com", "password": "x"})
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})
}

func TestMe(t *testing.T) {
	now := time.Now()

	t.Run("Returns the current user", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM users WHERE id = \$1`).WithArgs(7).
				WillReturnRows(sqlmock.NewRows(userCols).AddRow(7, "me@example.com", "Me", "hash", models.RoleUser, now))

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
			c.Set(ctxUserID, 7)
			Me(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.NotContains(t, w.Body.String(), "hash")
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns 404 when the user no longer exists", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM users WHERE id = \$1`).WillReturnError(sql.ErrNoRows)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
			c.Set(ctxUserID, 7)
			Me(c)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})
}
