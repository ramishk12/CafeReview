package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"cafe-review/internal/apierror"
	"cafe-review/internal/database"
	"cafe-review/internal/models"

	"github.com/gin-gonic/gin"
)

const minPasswordLength = 8

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

const userColumns = `id, email, display_name, password_hash, role, created_at`

func scanUser(row interface{ Scan(...any) error }) (models.User, error) {
	var u models.User
	err := row.Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.Role, &u.CreatedAt)
	return u, err
}

// Register creates a regular user account and returns a token.
func Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.DisplayName = strings.TrimSpace(req.DisplayName)

	if _, err := mail.ParseAddress(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address"})
		return
	}
	if req.DisplayName == "" || len(req.DisplayName) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Display name must be 1-100 characters"})
		return
	}
	if len(req.Password) < minPasswordLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters"})
		return
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		apierror.Internal(c, err)
		return
	}

	row := database.DB.QueryRow(
		`INSERT INTO users (email, display_name, password_hash, role)
		 VALUES ($1, $2, $3, 'user')
		 RETURNING `+userColumns,
		req.Email, req.DisplayName, hash,
	)
	user, err := scanUser(row)
	if database.IsUniqueViolation(err) {
		c.JSON(http.StatusConflict, gin.H{"error": "Email is already registered"})
		return
	}
	if err != nil {
		apierror.Internal(c, err)
		return
	}

	startSession(c, user)
	c.JSON(http.StatusCreated, user)
}

// Login verifies credentials and returns a token.
func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	// A locked account stops here, before the password is checked, so repeated
	// guessing cannot continue. The lockout is short to limit denial of service.
	if blocked, retryAfter := LoginGuard.Blocked(email); blocked {
		c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many failed attempts. Please try again later."})
		return
	}

	row := database.DB.QueryRow(`SELECT `+userColumns+` FROM users WHERE email = $1`, email)
	user, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !CheckPassword(user.PasswordHash, req.Password)) {
		LoginGuard.Fail(email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if err != nil {
		apierror.Internal(c, err)
		return
	}

	LoginGuard.Success(email)

	startSession(c, user)
	c.JSON(http.StatusOK, user)
}

// startSession creates a session and sets the session and CSRF cookies.
func startSession(c *gin.Context, user models.User) {
	token, csrf, err := createSession(user.ID)
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	setSessionCookies(c, token, csrf)
}

// Logout deletes the current session and clears the cookies.
func Logout(c *gin.Context) {
	if token, err := c.Cookie(SessionCookie); err == nil && token != "" {
		if err := deleteSession(token); err != nil {
			apierror.Internal(c, err)
			return
		}
	}
	clearSessionCookies(c)
	c.Status(http.StatusNoContent)
}

// Me returns the currently authenticated user.
func Me(c *gin.Context) {
	row := database.DB.QueryRow(`SELECT `+userColumns+` FROM users WHERE id = $1`, CurrentUserID(c))
	user, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		apierror.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}
