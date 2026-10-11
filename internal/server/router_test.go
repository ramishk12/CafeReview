package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"cafe-review/internal/auth"

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
}

func TestCafeWritesRequireAdmin(t *testing.T) {
	auth.Secret = []byte("router-test-secret")
	userToken, err := auth.GenerateToken(7, "user")
	assert.NoError(t, err)

	r := NewRouter(t.TempDir())

	tests := []struct {
		name   string
		method string
		path   string
		token  string
		want   int
	}{
		{"Update without a token is rejected", http.MethodPut, "/api/cafes/1", "", http.StatusUnauthorized},
		{"Update by a regular user is forbidden", http.MethodPut, "/api/cafes/1", userToken, http.StatusForbidden},
		{"Delete without a token is rejected", http.MethodDelete, "/api/cafes/1", "", http.StatusUnauthorized},
		{"Delete by a regular user is forbidden", http.MethodDelete, "/api/cafes/1", userToken, http.StatusForbidden},
		{"Deleting a review photo needs a token", http.MethodDelete, "/api/reviews/1/images/2", "", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}

func TestPublicAndReviewRoutes(t *testing.T) {
	auth.Secret = []byte("router-test-secret")
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

func TestReviewWritesRequireAuth(t *testing.T) {
	r := NewRouter(t.TempDir())

	for _, tt := range []struct{ method, path string }{
		{http.MethodPost, "/api/cafes/1/reviews"},
		{http.MethodPut, "/api/reviews/1"},
		{http.MethodDelete, "/api/reviews/1"},
		{http.MethodPost, "/api/reviews/1/images"},
	} {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}
