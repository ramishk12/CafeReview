package auth

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cafe-review/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Secret signs and verifies JWTs. It is set once at startup from configuration.
var Secret []byte

const tokenTTL = 24 * time.Hour

const (
	ctxUserID = "userID"
	ctxRole   = "role"
)

type claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func GenerateToken(userID int, role string) (string, error) {
	if len(Secret) == 0 {
		return "", errors.New("JWT secret is not configured")
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	})
	return token.SignedString(Secret)
}

func parseToken(raw string) (userID int, role string, err error) {
	var c claims
	token, err := jwt.ParseWithClaims(raw, &c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return Secret, nil
	})
	if err != nil || !token.Valid {
		return 0, "", errors.New("invalid token")
	}
	userID, err = strconv.Atoi(c.Subject)
	if err != nil {
		return 0, "", errors.New("invalid token subject")
	}
	return userID, c.Role, nil
}

// RequireAuth rejects requests without a valid Bearer token and stores the
// user ID and role in the context.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		raw, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}
		userID, role, err := parseToken(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}
		c.Set(ctxUserID, userID)
		c.Set(ctxRole, role)
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
