package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"net/mail"
	"strings"

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

type authResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create account"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, err := GenerateToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create session"})
		return
	}
	c.JSON(http.StatusCreated, authResponse{Token: token, User: user})
}

// Login verifies credentials and returns a token.
func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	row := database.DB.QueryRow(`SELECT `+userColumns+` FROM users WHERE email = $1`, email)
	user, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !CheckPassword(user.PasswordHash, req.Password)) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, err := GenerateToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create session"})
		return
	}
	c.JSON(http.StatusOK, authResponse{Token: token, User: user})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}
