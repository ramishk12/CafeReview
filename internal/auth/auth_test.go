package auth

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("correct horse")
	require.NoError(t, err)

	assert.NotEqual(t, "correct horse", hash, "password must not be stored in plain text")
	assert.True(t, CheckPassword(hash, "correct horse"))
	assert.False(t, CheckPassword(hash, "wrong password"))
	assert.False(t, CheckPassword("not-a-bcrypt-hash", "correct horse"))
}

func TestNewTokenIsRandomAndHashIsStable(t *testing.T) {
	a, err := newToken()
	require.NoError(t, err)
	b, err := newToken()
	require.NoError(t, err)

	assert.NotEqual(t, a, b)
	assert.Len(t, hashToken(a), 64, "sha256 hex is 64 characters")
	assert.Equal(t, hashToken(a), hashToken(a))
	assert.NotContains(t, hashToken(a), a, "the stored value is a hash, not the token")
}

func TestCreateSessionStoresOnlyTheHash(t *testing.T) {
	withMockDB(t, func(m sqlmock.Sqlmock) {
		m.ExpectExec(`INSERT INTO sessions \(token_hash, user_id, csrf_token, expires_at\)`).
			WithArgs(sqlmock.AnyArg(), 7, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		token, csrf, err := createSession(7)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotEmpty(t, csrf)
		assert.NotEqual(t, token, csrf)
		assert.Len(t, hashToken(token), 64)
		assert.NoError(t, m.ExpectationsWereMet())
	})
}

func TestLookupSession(t *testing.T) {
	t.Run("Returns the session and role", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			expires := time.Now().UTC().Add(sessionTTL)
			m.ExpectQuery(`FROM sessions s JOIN users u`).WithArgs(hashToken("tok")).
				WillReturnRows(sqlmock.NewRows([]string{"user_id", "csrf_token", "expires_at", "role"}).AddRow(7, "csrf-1", expires, "admin"))

			s, err := lookupSession("tok")
			require.NoError(t, err)
			assert.Equal(t, 7, s.UserID)
			assert.Equal(t, "csrf-1", s.CSRFToken)
			assert.Equal(t, "admin", s.Role)
			assert.WithinDuration(t, expires, s.ExpiresAt, time.Second)
		})
	})

	t.Run("Returns ErrNoRows for an unknown or expired token", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM sessions s JOIN users u`).WithArgs(hashToken("gone")).
				WillReturnError(sql.ErrNoRows)

			_, err := lookupSession("gone")
			assert.ErrorIs(t, err, sql.ErrNoRows)
		})
	})
}

// contextRecorder captures what the auth middleware put in the context.
func contextRecorder() (gin.HandlerFunc, *struct {
	id   int
	role string
}) {
	out := &struct {
		id   int
		role string
	}{}
	return func(c *gin.Context) {
		out.id = CurrentUserID(c)
		out.role = CurrentRole(c)
		c.Status(http.StatusOK)
	}, out
}

func TestRequireAuth(t *testing.T) {
	t.Run("Rejects a request without a session cookie", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			r := gin.New()
			h, _ := contextRecorder()
			r.GET("/x", RequireAuth(), h)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.NoError(t, m.ExpectationsWereMet())
		})
	})

	t.Run("Rejects an unknown session and clears the cookies", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM sessions`).WillReturnError(sql.ErrNoRows)

			r := gin.New()
			h, _ := contextRecorder()
			r.GET("/x", RequireAuth(), h)

			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			req.AddCookie(&http.Cookie{Name: SessionCookie, Value: "stale"})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			cookies := strings.Join(w.Header().Values("Set-Cookie"), "\n")
			assert.Contains(t, cookies, SessionCookie+"=", "cookies should be cleared")
		})
	})

	t.Run("Accepts a valid session and sets the context", func(t *testing.T) {
		withMockDB(t, func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM sessions`).WithArgs(hashToken("good")).
				WillReturnRows(sqlmock.NewRows([]string{"user_id", "csrf_token", "expires_at", "role"}).
					AddRow(7, "csrf-1", time.Now().UTC().Add(sessionTTL), "user"))

			r := gin.New()
			h, out := contextRecorder()
			r.GET("/x", RequireAuth(), h)

			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			req.AddCookie(&http.Cookie{Name: SessionCookie, Value: "good"})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, 7, out.id)
			assert.Equal(t, "user", out.role)
		})
	})
}

func TestRequireCSRF(t *testing.T) {
	build := func(method string, header string) *httptest.ResponseRecorder {
		r := gin.New()
		r.Use(func(c *gin.Context) { c.Set(ctxCSRF, "expected-token") })
		r.Handle(method, "/x", RequireCSRF(), func(c *gin.Context) { c.Status(http.StatusOK) })

		req := httptest.NewRequest(method, "/x", nil)
		if header != "" {
			req.Header.Set("X-CSRF-Token", header)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	assert.Equal(t, http.StatusOK, build(http.MethodGet, "").Code, "safe methods do not need a token")
	assert.Equal(t, http.StatusOK, build(http.MethodPost, "expected-token").Code)
	assert.Equal(t, http.StatusForbidden, build(http.MethodPost, "wrong").Code)
	assert.Equal(t, http.StatusForbidden, build(http.MethodPost, "").Code)
	assert.Equal(t, http.StatusForbidden, build(http.MethodDelete, "wrong").Code)
}

func TestRequireAdmin(t *testing.T) {
	build := func(role string) int {
		r := gin.New()
		r.Use(func(c *gin.Context) { c.Set(ctxRole, role) })
		r.GET("/admin", RequireAdmin(), func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin", nil))
		return w.Code
	}

	assert.Equal(t, http.StatusOK, build("admin"))
	assert.Equal(t, http.StatusForbidden, build("user"))
	assert.Equal(t, http.StatusForbidden, build(""))
}

func TestSessionCookiesFlags(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setSessionCookies(c, "sess", "csrf")

	cookies := w.Result().Cookies()
	byName := map[string]*http.Cookie{}
	for _, ck := range cookies {
		byName[ck.Name] = ck
	}

	require.Contains(t, byName, SessionCookie)
	assert.True(t, byName[SessionCookie].HttpOnly, "the session cookie must be HttpOnly")
	assert.False(t, byName[CSRFCookie].HttpOnly, "the CSRF cookie must be readable by scripts")
	assert.Equal(t, http.SameSiteLaxMode, byName[SessionCookie].SameSite)
	assert.Greater(t, byName[SessionCookie].MaxAge, 0)
	assert.Equal(t, int(sessionTTL.Seconds()), byName[SessionCookie].MaxAge)
}

func assertHasSessionCookie(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	for _, ck := range w.Result().Cookies() {
		if ck.Name == SessionCookie {
			assert.NotEmpty(t, ck.Value)
			assert.True(t, ck.HttpOnly)
			return
		}
	}
	t.Fatal("expected a session cookie to be set")
}

func TestShouldExtend(t *testing.T) {
	now := time.Now().UTC()
	assert.False(t, shouldExtend(now.Add(sessionTTL), now), "a fresh session is not extended yet")
	assert.False(t, shouldExtend(now.Add(sessionTTL-sessionSlideInterval+time.Minute), now), "not due yet")
	assert.True(t, shouldExtend(now.Add(sessionTTL-sessionSlideInterval), now), "due at the slide interval")
	assert.True(t, shouldExtend(now.Add(sessionTTL-sessionSlideInterval-time.Hour), now), "overdue")
	assert.True(t, shouldExtend(now.Add(-time.Minute), now), "already past")
}

func TestRequireAuthSlidesExpiration(t *testing.T) {
	withMockDB(t, func(m sqlmock.Sqlmock) {
		// The stored expiry is older than the slide interval, so it should be pushed forward.
		m.ExpectQuery(`FROM sessions`).WithArgs(hashToken("good")).
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "csrf_token", "expires_at", "role"}).
				AddRow(7, "csrf-1", time.Now().UTC().Add(sessionTTL-sessionSlideInterval-time.Hour), "user"))
		m.ExpectExec(`UPDATE sessions SET expires_at = \$1 WHERE token_hash = \$2`).
			WithArgs(sqlmock.AnyArg(), hashToken("good")).
			WillReturnResult(sqlmock.NewResult(0, 1))

		r := gin.New()
		h, _ := contextRecorder()
		r.GET("/x", RequireAuth(), h)

		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.AddCookie(&http.Cookie{Name: SessionCookie, Value: "good"})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Result().Cookies(), "the cookie should be refreshed with the new expiry")
		assert.NoError(t, m.ExpectationsWereMet())
	})
}

func TestRequireAuthDoesNotSlideFreshSession(t *testing.T) {
	withMockDB(t, func(m sqlmock.Sqlmock) {
		// A fresh session is not due, so no UPDATE should happen.
		m.ExpectQuery(`FROM sessions`).WithArgs(hashToken("good")).
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "csrf_token", "expires_at", "role"}).
				AddRow(7, "csrf-1", time.Now().UTC().Add(sessionTTL), "user"))

		r := gin.New()
		h, _ := contextRecorder()
		r.GET("/x", RequireAuth(), h)

		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.AddCookie(&http.Cookie{Name: SessionCookie, Value: "good"})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, m.ExpectationsWereMet(), "no UPDATE expected for a fresh session")
	})
}

func TestDeleteExpiredSessions(t *testing.T) {
	withMockDB(t, func(m sqlmock.Sqlmock) {
		m.ExpectExec(`DELETE FROM sessions WHERE expires_at <`).
			WillReturnResult(sqlmock.NewResult(0, 3))

		n, err := DeleteExpiredSessions()
		require.NoError(t, err)
		assert.Equal(t, int64(3), n)
		assert.NoError(t, m.ExpectationsWereMet())
	})
}

func TestDeleteExpiredSessionsReportsErrors(t *testing.T) {
	withMockDB(t, func(m sqlmock.Sqlmock) {
		m.ExpectExec(`DELETE FROM sessions WHERE expires_at <`).WillReturnError(sql.ErrConnDone)

		_, err := DeleteExpiredSessions()
		assert.Error(t, err)
	})
}
