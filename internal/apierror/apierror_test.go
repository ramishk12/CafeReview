package apierror

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// captureLog redirects the standard logger for the duration of the test.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(previous) })
	return &buf
}

func TestInternalHidesErrorAndLogsIt(t *testing.T) {
	logs := captureLog(t)
	secret := errors.New(`pq: relation "users" does not exist`)

	r := gin.New()
	r.Use(RequestID())
	r.GET("/fail", func(c *gin.Context) { Internal(c, secret) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/fail", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NotContains(t, w.Body.String(), "users", "database details must not reach the client")
	assert.NotContains(t, w.Body.String(), "pq:")

	var body struct {
		Error     string `json:"error"`
		RequestID string `json:"request_id"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "Something went wrong", body.Error)
	assert.Equal(t, w.Header().Get(HeaderRequestID), body.RequestID, "body and header carry the same ID")

	assert.Contains(t, logs.String(), `error=pq: relation "users" does not exist`, "the real error is logged")
	assert.Contains(t, logs.String(), "request_id="+body.RequestID, "the log line carries the request ID")
	assert.Contains(t, logs.String(), "path=/fail")
}

func TestRequestIDGeneratesWhenMissing(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	var seen string
	r.GET("/", func(c *gin.Context) {
		seen = RequestIDFrom(c)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Len(t, seen, 16, "8 random bytes in hex")
	assert.Equal(t, seen, w.Header().Get(HeaderRequestID))
}

func TestRequestIDKeepsValidIncomingID(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderRequestID, "proxy-abc_123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, "proxy-abc_123", w.Header().Get(HeaderRequestID))
}

func TestRequestIDReplacesUnsafeIncomingID(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"newline (log injection)", "abc\nlevel=error"},
		{"spaces", "has spaces"},
		{"too long", strings.Repeat("a", 65)},
		{"markup", "<script>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(RequestID())
			r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(HeaderRequestID, tt.value)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			got := w.Header().Get(HeaderRequestID)
			assert.NotEqual(t, tt.value, got)
			assert.Len(t, got, 16)
		})
	}
}

func TestRequestIDFromWithoutMiddleware(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	assert.Equal(t, "", RequestIDFrom(c))
}
