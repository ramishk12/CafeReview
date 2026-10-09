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
