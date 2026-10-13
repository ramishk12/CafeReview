package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cafe-review/internal/auth"
	"cafe-review/internal/ratelimit"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func hasRoute(r *gin.Engine, method, path string) bool {
	for _, route := range r.Routes() {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}

func TestRouterRegistersNewRoutes(t *testing.T) {
	r := NewRouter(t.TempDir())

	assert.True(t, hasRoute(r, http.MethodPut, "/api/cafes/:id"), "PUT /api/cafes/:id")
	assert.True(t, hasRoute(r, http.MethodDelete, "/api/cafes/:id"), "DELETE /api/cafes/:id")
	assert.True(t, hasRoute(r, http.MethodDelete, "/api/reviews/:id/images/:imageId"), "DELETE /api/reviews/:id/images/:imageId")
	assert.True(t, hasRoute(r, http.MethodPost, "/api/auth/logout"), "POST /api/auth/logout")
}

func TestPublicAndReviewRoutes(t *testing.T) {
	r := NewRouter(t.TempDir())

	for _, path := range []string{"/api/health", "/api/cafes", "/api/cafes/:id", "/api/cafes/:id/reviews"} {
		assert.True(t, hasRoute(r, http.MethodGet, path), "GET "+path)
	}
	assert.True(t, hasRoute(r, http.MethodPost, "/api/auth/register"))
	assert.True(t, hasRoute(r, http.MethodPost, "/api/auth/login"))
	assert.True(t, hasRoute(r, http.MethodPost, "/api/cafes/:id/reviews"))
	assert.True(t, hasRoute(r, http.MethodPut, "/api/reviews/:id"))
	assert.True(t, hasRoute(r, http.MethodPost, "/api/reviews/:id/images"))
}

func TestWritesRequireASession(t *testing.T) {
	r := NewRouter(t.TempDir())

	for _, tt := range []struct{ method, path string }{
		{http.MethodPost, "/api/cafes/1/reviews"},
		{http.MethodPut, "/api/reviews/1"},
		{http.MethodDelete, "/api/reviews/1"},
		{http.MethodPost, "/api/reviews/1/images"},
		{http.MethodPost, "/api/auth/logout"},
		{http.MethodPut, "/api/cafes/1"},
		{http.MethodDelete, "/api/cafes/1"},
	} {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestEveryResponseHasRequestID(t *testing.T) {
	r := NewRouter(t.TempDir())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/health", nil))

	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestSecurityHeaders(t *testing.T) {
	r := NewRouter(t.TempDir())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/health", nil))

	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "no-referrer", w.Header().Get("Referrer-Policy"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "same-origin", w.Header().Get("Cross-Origin-Resource-Policy"))
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src 'none'")
}

func TestLoginRateLimitAppliesInRouter(t *testing.T) {
	previous := auth.LoginLimiter
	auth.LoginLimiter = ratelimit.NewKeyed(1, time.Minute)
	t.Cleanup(func() { auth.LoginLimiter = previous })

	r := NewRouter(t.TempDir())
	// Use up the single token for the test client's IP.
	auth.LoginLimiter.Allow("192.0.2.1")

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.NotEmpty(t, w.Header().Get("Retry-After"))
}
