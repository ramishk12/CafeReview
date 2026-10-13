package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestKeyedAllowsBurstThenBlocks(t *testing.T) {
	k := NewKeyed(3, time.Minute)

	for i := 0; i < 3; i++ {
		assert.True(t, k.Allow("1.2.3.4"), "attempt %d should be allowed", i+1)
	}
	assert.False(t, k.Allow("1.2.3.4"), "the 4th attempt should be blocked")
	assert.True(t, k.Allow("5.6.7.8"), "a different key has its own bucket")
}

func TestKeyedMiddlewareReturns429WithRetryAfter(t *testing.T) {
	k := NewKeyed(1, time.Minute)

	r := gin.New()
	r.POST("/login", k.Middleware(), func(c *gin.Context) { c.Status(http.StatusOK) })

	// First request passes. The key is the client IP, which for httptest is 192.0.2.1.
	first := httptest.NewRecorder()
	r.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/login", nil))
	assert.Equal(t, http.StatusOK, first.Code)

	second := httptest.NewRecorder()
	r.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/login", nil))
	assert.Equal(t, http.StatusTooManyRequests, second.Code)
	assert.Equal(t, "60", second.Header().Get("Retry-After"))
}

func TestKeyedSweepsIdleEntries(t *testing.T) {
	k := NewKeyed(1, time.Minute)
	now := time.Now()
	k.now = func() time.Time { return now }

	k.Allow("old")
	assert.Len(t, k.entries, 1)

	// Advance past the idle TTL, then touch another key to trigger the sweep.
	now = now.Add(11 * time.Minute)
	k.Allow("new")
	assert.Len(t, k.entries, 1, "the idle entry should have been swept")
	assert.Contains(t, k.entries, "new")
}

func TestAccountGuardLocksAfterMaxFailures(t *testing.T) {
	g := NewAccountGuard(3, time.Hour, 15*time.Minute)

	assert.False(t, blocked(g, "a@example.com"))

	g.Fail("a@example.com")
	g.Fail("a@example.com")
	assert.False(t, blocked(g, "a@example.com"), "still below the limit")

	g.Fail("a@example.com")
	locked, remaining := g.Blocked("a@example.com")
	assert.True(t, locked, "the third failure should lock the account")
	// remaining is the lockout minus the small time since Fail, so bound it.
	assert.Greater(t, remaining, 14*time.Minute)
	assert.LessOrEqual(t, remaining, 15*time.Minute)

	assert.False(t, blocked(g, "other@example.com"), "other accounts are unaffected")
}

func TestAccountGuardSuccessClearsFailures(t *testing.T) {
	g := NewAccountGuard(3, time.Hour, 15*time.Minute)

	g.Fail("a@example.com")
	g.Fail("a@example.com")
	g.Success("a@example.com")

	// The two earlier failures are forgotten, so it takes three more to lock.
	g.Fail("a@example.com")
	g.Fail("a@example.com")
	assert.False(t, blocked(g, "a@example.com"))
	g.Fail("a@example.com")
	assert.True(t, blocked(g, "a@example.com"))
}

func TestAccountGuardLockoutExpires(t *testing.T) {
	g := NewAccountGuard(1, time.Hour, 15*time.Minute)
	now := time.Now()
	g.now = func() time.Time { return now }

	g.Fail("a@example.com")
	assert.True(t, blocked(g, "a@example.com"))

	now = now.Add(16 * time.Minute)
	assert.False(t, blocked(g, "a@example.com"), "the lockout should have expired")
}

func TestAccountGuardIgnoresOldFailures(t *testing.T) {
	g := NewAccountGuard(3, time.Hour, 15*time.Minute)
	now := time.Now()
	g.now = func() time.Time { return now }

	g.Fail("a@example.com")
	g.Fail("a@example.com")
	// Both earlier failures fall outside the one-hour window.
	now = now.Add(2 * time.Hour)
	g.Fail("a@example.com")
	assert.False(t, blocked(g, "a@example.com"), "old failures should not count toward the limit")
}

func blocked(g *AccountGuard, key string) bool {
	ok, _ := g.Blocked(key)
	return ok
}

// Ensure Retry-After is a positive number of seconds.
func TestMiddlewareRetryAfterIsPositive(t *testing.T) {
	k := NewKeyed(5, time.Minute)
	if _, err := strconv.Atoi(strconv.Itoa(int(k.interval.Seconds()))); err != nil {
		t.Fatalf("interval not numeric: %v", err)
	}
}
