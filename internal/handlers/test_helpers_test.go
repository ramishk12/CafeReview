package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"cafe-review/internal/database"
	"cafe-review/internal/storage"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

// TestMain points uploads at a temporary folder so tests never touch the real uploads directory.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	dir, err := os.MkdirTemp("", "cafe-review-test-uploads-")
	if err != nil {
		panic(err)
	}
	local, err := storage.NewLocal(dir, "/uploads")
	if err != nil {
		panic(err)
	}
	Uploads = local

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// withMockDB swaps database.DB for a sqlmock connection while f runs.
func withMockDB(t *testing.T, f func(mock sqlmock.Sqlmock)) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create mock db: %v", err)
	}
	defer db.Close()

	original := database.DB
	database.DB = db
	defer func() { database.DB = original }()

	f(mock)
}

// newTestContext builds a gin context for one request. The user ID and role are
// set the same way auth.RequireAuth sets them.
func newTestContext(t *testing.T, method, target string, body any, params gin.Params, userID int, role string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(data)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, reader)
	c.Params = params
	if userID > 0 {
		c.Set("userID", userID)
	}
	if role != "" {
		c.Set("role", role)
	}
	return c, w
}

// writeUpload creates a file in the test uploads folder and returns its path.
func writeUpload(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(uploadsDir(t), name)
	if err := os.WriteFile(path, []byte("test image"), 0o644); err != nil {
		t.Fatalf("write upload: %v", err)
	}
	return path
}

// uploadsDir returns the test uploads folder.
func uploadsDir(t *testing.T) string {
	t.Helper()
	local, ok := Uploads.(*storage.Local)
	if !ok {
		t.Fatalf("Uploads is %T, want *storage.Local", Uploads)
	}
	return local.Dir
}
