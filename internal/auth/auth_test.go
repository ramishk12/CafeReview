package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cafe-review/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// useSecret sets the signing secret for one test and restores the previous value afterwards.
func useSecret(t *testing.T, secret string) {
	t.Helper()
	previous := Secret
	Secret = []byte(secret)
	t.Cleanup(func() { Secret = previous })
}

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("correct horse")
	require.NoError(t, err)

	assert.NotEqual(t, "correct horse", hash, "password must not be stored in plain text")
	assert.True(t, CheckPassword(hash, "correct horse"))
	assert.False(t, CheckPassword(hash, "wrong password"))
	assert.False(t, CheckPassword("not-a-bcrypt-hash", "correct horse"))
}

func TestGenerateTokenRequiresSecret(t *testing.T) {
	useSecret(t, "")

	_, err := GenerateToken(1, models.RoleUser)
	assert.Error(t, err)
}

func TestTokenRoundTrip(t *testing.T) {
	useSecret(t, "test-secret-for-round-trip")

	token, err := GenerateToken(42, models.RoleAdmin)
	require.NoError(t, err)

	userID, role, err := parseToken(token)
	require.NoError(t, err)
	assert.Equal(t, 42, userID)
	assert.Equal(t, models.RoleAdmin, role)
}

func TestParseTokenRejectsInvalidTokens(t *testing.T) {
	useSecret(t, "test-secret-for-rejection")

	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Role: models.RoleUser,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	})
	expiredToken, err := expired.SignedString(Secret)
	require.NoError(t, err)

	otherSecret, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Role:             models.RoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{Subject: "1"},
	}).SignedString([]byte("attacker-secret"))
	require.NoError(t, err)

	noneToken, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims{
		Role:             models.RoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{Subject: "1"},
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	nonNumericSubject, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Role:             models.RoleUser,
		RegisteredClaims: jwt.RegisteredClaims{Subject: "not-a-number"},
	}).SignedString(Secret)
	require.NoError(t, err)

	tests := []struct {
		name  string
		token string
	}{
		{"malformed", "not.a.jwt"},
		{"expired", expiredToken},
		{"signed with another secret", otherSecret},
		{"unsigned alg none", noneToken},
		{"non-numeric subject", nonNumericSubject},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parseToken(tt.token)
			assert.Error(t, err)
		})
	}
}

func TestRequireAuthAndAdmin(t *testing.T) {
	useSecret(t, "test-secret-for-middleware")

	userToken, err := GenerateToken(7, models.RoleUser)
	require.NoError(t, err)
	adminToken, err := GenerateToken(1, models.RoleAdmin)
	require.NoError(t, err)

	r := gin.New()
	r.GET("/me", RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"id": CurrentUserID(c), "role": CurrentRole(c)})
	})
	r.GET("/admin", RequireAuth(), RequireAdmin(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	tests := []struct {
		name   string
		path   string
		header string
		want   int
	}{
		{"no header", "/me", "", http.StatusUnauthorized},
		{"wrong scheme", "/me", "Basic " + userToken, http.StatusUnauthorized},
		{"empty bearer", "/me", "Bearer ", http.StatusUnauthorized},
		{"garbage token", "/me", "Bearer nope", http.StatusUnauthorized},
		{"valid user token", "/me", "Bearer " + userToken, http.StatusOK},
		{"user on admin route", "/admin", "Bearer " + userToken, http.StatusForbidden},
		{"admin on admin route", "/admin", "Bearer " + adminToken, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}

func TestRequireAuthSetsUserContext(t *testing.T) {
	useSecret(t, "test-secret-for-context")
	token, err := GenerateToken(7, models.RoleUser)
	require.NoError(t, err)

	r := gin.New()
	r.GET("/me", RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"id": CurrentUserID(c), "role": CurrentRole(c)})
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"id":7,"role":"user"}`, w.Body.String())
}
