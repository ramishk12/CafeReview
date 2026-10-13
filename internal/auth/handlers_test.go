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
	"cafe-review/internal/mail"
	"cafe-review/internal/models"
	"cafe-review/internal/ratelimit"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var userCols = []string{"id", "email", "display_name", "password_hash", "role", "created_at", "email_verified_at"}

// fakeMailer records sent messages.
type fakeMailer struct {
	sent []mail.Message
	err  error
}

func (f *fakeMailer) Send(m mail.Message) error {
	f.sent = append(f.sent, m)
	return f.err
}

// useMailer replaces the package mailer for one test.
func useMailer(t *testing.T, f *fakeMailer) {
	t.Helper()
	previous := Mailer
	Mailer = f
	t.Cleanup(func() { Mailer = previous })
}

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
	now := time.Now()

	t.Run("Creates an unverified user, sends the email, and returns a generic message", func(t *testing.T) {
		mailer := &fakeMailer{}
		useMailer(t, mailer)

		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`INSERT INTO users`).
				WithArgs("new@example.com", "Ann", sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows(userCols).AddRow(5, "new@example.com", "Ann", "hash", "user", now, nil))
			m.ExpectExec(`INSERT INTO email_tokens`).WillReturnResult(sqlmock.NewResult(1, 1))

			w := postJSON(t, Register, "/api/auth/register", map[string]string{
				"email": "  New@Example.com ", "password": "longenough", "display_name": " Ann ",
			})

			assert.Equal(t, http.StatusCreated, w.Code)
			assert.Contains(t, w.Body.String(), "verification link")
			assert.NotContains(t, w.Body.String(), "password_hash")
			require.Len(t, mailer.sent, 1)
			assert.Equal(t, "new@example.com", mailer.sent[0].To)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Returns the same response for an existing email and sends nothing", func(t *testing.T) {
		mailer := &fakeMailer{}
		useMailer(t, mailer)

		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`INSERT INTO users`).WillReturnError(&pq.Error{Code: "23505"})

			w := postJSON(t, Register, "/api/auth/register", map[string]string{
				"email": "taken@example.com", "password": "longenough", "display_name": "Ann",
			})

			assert.Equal(t, http.StatusCreated, w.Code)
			assert.Contains(t, w.Body.String(), "verification link")
			assert.Empty(t, mailer.sent)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Still succeeds when the email cannot be sent", func(t *testing.T) {
		useMailer(t, &fakeMailer{err: errors.New("smtp down")})

		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`INSERT INTO users`).
				WillReturnRows(sqlmock.NewRows(userCols).AddRow(5, "new@example.com", "Ann", "hash", "user", now, nil))
			m.ExpectExec(`INSERT INTO email_tokens`).WillReturnResult(sqlmock.NewResult(1, 1))

			w := postJSON(t, Register, "/api/auth/register", map[string]string{
				"email": "new@example.com", "password": "longenough", "display_name": "Ann",
			})
			assert.Equal(t, http.StatusCreated, w.Code)
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
}

func TestLogin(t *testing.T) {
	now := time.Now()
	hash, err := HashPassword("correct-password")
	require.NoError(t, err)

	t.Run("Starts a session for a verified account", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM users WHERE email = \$1`).WithArgs("demo@example.com").
				WillReturnRows(sqlmock.NewRows(userCols).AddRow(2, "demo@example.com", "Demo", hash, "user", now, now))
			m.ExpectExec(`INSERT INTO sessions`).WillReturnResult(sqlmock.NewResult(1, 1))

			w := postJSON(t, Login, "/api/auth/login", map[string]string{"email": "Demo@Example.com", "password": "correct-password"})

			assert.Equal(t, http.StatusOK, w.Code)
			var u models.User
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &u))
			assert.Equal(t, 2, u.ID)
			assertHasSessionCookie(t, w)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Refuses an unverified account without a session", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM users WHERE email`).WithArgs("new@example.com").
				WillReturnRows(sqlmock.NewRows(userCols).AddRow(2, "new@example.com", "New", hash, "user", now, nil))

			w := postJSON(t, Login, "/api/auth/login", map[string]string{"email": "new@example.com", "password": "correct-password"})

			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.Contains(t, w.Body.String(), "verify your email")
			assert.NoError(t, m.ExpectationsWereMet(), "no session should be created")
		})
	})

	t.Run("Returns the same 401 for a wrong password and an unknown email", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM users WHERE email`).WillReturnRows(sqlmock.NewRows(userCols).AddRow(2, "demo@example.com", "Demo", hash, "user", now, now))
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

func TestVerify(t *testing.T) {
	t.Run("Confirms the email for a valid token", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`UPDATE email_tokens SET used_at`).WithArgs(hashToken("tok"), verifyTokenKind).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(5))
			m.ExpectExec(`UPDATE users SET email_verified_at`).WithArgs(5).
				WillReturnResult(sqlmock.NewResult(0, 1))

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/auth/verify?token=tok", nil)
			Verify(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Rejects an unknown, used, or expired token", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`UPDATE email_tokens`).WillReturnError(sql.ErrNoRows)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/auth/verify?token=bad", nil)
			Verify(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})

	t.Run("Rejects a missing token without touching the database", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/auth/verify", nil)
			Verify(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})
}

func TestResendVerification(t *testing.T) {
	now := time.Now()

	t.Run("Sends a new link for an unverified account", func(t *testing.T) {
		mailer := &fakeMailer{}
		useMailer(t, mailer)

		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT id, email_verified_at FROM users WHERE email`).WithArgs("new@example.com").
				WillReturnRows(sqlmock.NewRows([]string{"id", "email_verified_at"}).AddRow(5, nil))
			m.ExpectExec(`INSERT INTO email_tokens`).WillReturnResult(sqlmock.NewResult(1, 1))

			w := postJSON(t, ResendVerification, "/api/auth/verify/resend", map[string]string{"email": "New@Example.com"})

			assert.Equal(t, http.StatusOK, w.Code)
			require.Len(t, mailer.sent, 1)
			assert.Equal(t, "new@example.com", mailer.sent[0].To)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Sends nothing for an already verified account but replies the same", func(t *testing.T) {
		mailer := &fakeMailer{}
		useMailer(t, mailer)

		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT id, email_verified_at`).
				WillReturnRows(sqlmock.NewRows([]string{"id", "email_verified_at"}).AddRow(5, now))

			w := postJSON(t, ResendVerification, "/api/auth/verify/resend", map[string]string{"email": "done@example.com"})
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Empty(t, mailer.sent)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Replies the same for an unknown email", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT id, email_verified_at`).WillReturnError(sql.ErrNoRows)

			w := postJSON(t, ResendVerification, "/api/auth/verify/resend", map[string]string{"email": "nobody@example.com"})
			assert.Equal(t, http.StatusOK, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})
}

func TestMe(t *testing.T) {
	now := time.Now()

	t.Run("Returns the current user", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM users WHERE id = \$1`).WithArgs(7).
				WillReturnRows(sqlmock.NewRows(userCols).AddRow(7, "me@example.com", "Me", "hash", models.RoleUser, now, now))

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

func TestLogout(t *testing.T) {
	t.Run("Deletes the session and clears the cookies", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectExec(`DELETE FROM sessions WHERE token_hash = \$1`).WithArgs(hashToken("tok")).
				WillReturnResult(sqlmock.NewResult(0, 1))

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
			c.Request.AddCookie(&http.Cookie{Name: SessionCookie, Value: "tok"})
			Logout(c)
			c.Writer.WriteHeaderNow()

			assert.Equal(t, http.StatusNoContent, w.Code)
			assert.NotEmpty(t, w.Result().Cookies(), "cookies should be cleared")
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Succeeds without a session cookie", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
			Logout(c)
			c.Writer.WriteHeaderNow()

			assert.Equal(t, http.StatusNoContent, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})
}

func TestLoginAccountLockout(t *testing.T) {
	previous := LoginGuard
	LoginGuard = ratelimit.NewAccountGuard(2, time.Hour, 15*time.Minute)
	t.Cleanup(func() { LoginGuard = previous })

	hash, err := HashPassword("correct-password")
	require.NoError(t, err)
	now := time.Now()

	// Two failed attempts lock the account.
	for i := 0; i < 2; i++ {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM users WHERE email`).WillReturnRows(
				sqlmock.NewRows(userCols).AddRow(2, "lock@example.com", "Lock", hash, "user", now, now))
			w := postJSON(t, Login, "/api/auth/login", map[string]string{"email": "lock@example.com", "password": "wrong"})
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}

	// The next attempt is refused before the database is used, even with the right password.
	for _, password := range []string{"wrong", "correct-password"} {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			w := postJSON(t, Login, "/api/auth/login", map[string]string{"email": "lock@example.com", "password": password})
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
			assert.NotEmpty(t, w.Header().Get("Retry-After"))
			assert.NoError(t, m.ExpectationsWereMet(), "no database call should be made while locked")
		})
	}
}
