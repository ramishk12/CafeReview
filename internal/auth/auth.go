package auth

import (
	"crypto/subtle"
	"database/sql"
	"errors"
	"net/http"

	"cafe-review/internal/apierror"
	"cafe-review/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const (
	ctxUserID = "userID"
	ctxRole   = "role"
	ctxCSRF   = "csrf"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// RequireAuth authenticates a request from its session cookie. It loads the
// session and the user's current role from the database, so a role change takes
// effect on the next request, and stores the user ID, role, and the session's
// CSRF token in the context.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(SessionCookie)
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}
		sess, err := lookupSession(token)
		if errors.Is(err, sql.ErrNoRows) {
			clearSessionCookies(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
			return
		}
		if err != nil {
			apierror.Internal(c, err)
			return
		}
		c.Set(ctxUserID, sess.UserID)
		c.Set(ctxRole, sess.Role)
		c.Set(ctxCSRF, sess.CSRFToken)
		c.Next()
	}
}

// RequireCSRF must be placed after RequireAuth. For state-changing methods it
// compares the X-CSRF-Token header with the session's CSRF token. Safe methods
// pass through. The token is tied to the session and compared in constant time.
func RequireCSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}
		expected, _ := c.Get(ctxCSRF)
		want, _ := expected.(string)
		given := c.GetHeader("X-CSRF-Token")
		if want == "" || given == "" || subtle.ConstantTimeCompare([]byte(want), []byte(given)) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid CSRF token"})
			return
		}
		c.Next()
	}
}

// RequireAdmin must be placed after RequireAuth.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if CurrentRole(c) != models.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}
		c.Next()
	}
}

// CurrentUserID returns the authenticated user's ID. It is only valid behind RequireAuth.
func CurrentUserID(c *gin.Context) int {
	id, _ := c.Get(ctxUserID)
	v, _ := id.(int)
	return v
}

// CurrentRole returns the authenticated user's role. It is only valid behind RequireAuth.
func CurrentRole(c *gin.Context) string {
	role, _ := c.Get(ctxRole)
	v, _ := role.(string)
	return v
}
