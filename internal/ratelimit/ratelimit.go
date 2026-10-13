// Package ratelimit limits requests per key (for example, per client IP) and
// failed authentication attempts per account. Both store state in memory, so
// each process has its own counters. A shared store would be needed to enforce
// the limits across more than one instance.
package ratelimit

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Keyed limits requests per key using a token bucket. Entries idle for a while
// are swept, so the map does not grow without bound.
type Keyed struct {
	mu        sync.Mutex
	entries   map[string]*keyEntry
	limit     rate.Limit
	burst     int
	interval  time.Duration // time to refill one token, sent as Retry-After
	idleTTL   time.Duration
	lastSweep time.Time
	now       func() time.Time
}

type keyEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewKeyed allows limit requests per period for each key, with a burst equal to
// limit.
func NewKeyed(limit int, period time.Duration) *Keyed {
	return &Keyed{
		entries:  make(map[string]*keyEntry),
		limit:    rate.Limit(float64(limit) / period.Seconds()),
		burst:    limit,
		interval: period / time.Duration(limit),
		idleTTL:  10 * period,
		now:      time.Now,
	}
}

// Allow reports whether key may proceed now.
func (k *Keyed) Allow(key string) bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.sweepLocked()

	e := k.entries[key]
	if e == nil {
		e = &keyEntry{limiter: rate.NewLimiter(k.limit, k.burst)}
		k.entries[key] = e
	}
	e.lastSeen = k.now()
	return e.limiter.Allow()
}

func (k *Keyed) sweepLocked() {
	now := k.now()
	if now.Sub(k.lastSweep) < k.idleTTL {
		return
	}
	for key, e := range k.entries {
		if now.Sub(e.lastSeen) > k.idleTTL {
			delete(k.entries, key)
		}
	}
	k.lastSweep = now
}

// Middleware rejects requests over the limit with 429 and a Retry-After header.
func (k *Keyed) Middleware() gin.HandlerFunc {
	retry := strconv.Itoa(int(k.interval.Seconds()))
	return func(c *gin.Context) {
		if !k.Allow(c.ClientIP()) {
			c.Header("Retry-After", retry)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many attempts. Please try again later."})
			return
		}
		c.Next()
	}
}

// AccountGuard limits failed attempts per key (for example, per email) within a
// window, and locks the key out for a period once the limit is reached. A
// successful login clears the key.
type AccountGuard struct {
	mu          sync.Mutex
	entries     map[string]*guardEntry
	maxFailures int
	window      time.Duration
	lockout     time.Duration
	now         func() time.Time
}

type guardEntry struct {
	failures    []time.Time
	lockedUntil time.Time
}

// NewAccountGuard returns a guard that locks a key out for lockout once it has
// maxFailures failures within window.
func NewAccountGuard(maxFailures int, window, lockout time.Duration) *AccountGuard {
	return &AccountGuard{
		entries:     make(map[string]*guardEntry),
		maxFailures: maxFailures,
		window:      window,
		lockout:     lockout,
		now:         time.Now,
	}
}

// Blocked reports whether key is locked out and, if so, how long remains.
func (g *AccountGuard) Blocked(key string) (bool, time.Duration) {
	g.mu.Lock()
	defer g.mu.Unlock()
	e := g.entries[key]
	if e == nil {
		return false, 0
	}
	now := g.now()
	if now.Before(e.lockedUntil) {
		return true, e.lockedUntil.Sub(now)
	}
	return false, 0
}

// Fail records a failed attempt and locks the key out when the limit is reached.
func (g *AccountGuard) Fail(key string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.now()
	e := g.entries[key]
	if e == nil {
		e = &guardEntry{}
		g.entries[key] = e
	}

	// Drop failures outside the window, then record this one.
	cutoff := now.Add(-g.window)
	kept := e.failures[:0]
	for _, t := range e.failures {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	e.failures = append(kept, now)

	if len(e.failures) >= g.maxFailures {
		e.lockedUntil = now.Add(g.lockout)
		e.failures = nil
	}
}

// Success clears the key's failure history.
func (g *AccountGuard) Success(key string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.entries, key)
}
