package auth

import (
	"time"

	"cafe-review/internal/ratelimit"
)

// Limiters and guards for the authentication endpoints. They are variables so
// they can be tuned in one place and replaced in tests. They hold state in
// memory, so each process enforces its own limits.
var (
	// LoginLimiter allows 5 login requests per minute per IP.
	LoginLimiter = ratelimit.NewKeyed(5, time.Minute)
	// RegisterLimiter allows 5 registrations per hour per IP.
	RegisterLimiter = ratelimit.NewKeyed(5, time.Hour)
	// LoginGuard locks an account for 15 minutes after 10 failed logins within an hour.
	LoginGuard = ratelimit.NewAccountGuard(10, time.Hour, 15*time.Minute)
)
