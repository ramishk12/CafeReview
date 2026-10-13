package auth

import (
	"database/sql"
	"errors"
	"log"
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

const userColumns = `id, email, display_name, password_hash, role, created_at, email_verified_at`

// registerMessage is the same response for a new and an existing email, so
// registration cannot be used to check whether an address has an account.
const registerMessage = "If this email can be registered, we've sent a verification link. Check your inbox."

func scanUser(row interface{ Scan(...any) error }) (models.User, error) {
	var u models.User
	var verified sql.NullTime
	err := row.Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.Role, &u.CreatedAt, &verified)
	if verified.Valid {
		u.EmailVerifiedAt = &verified.Time
	}
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
		// Do not reveal that the address is taken: reply exactly as for a new signup.
		c.JSON(http.StatusCreated, gin.H{"message": registerMessage})
		return
	}
	if err != nil {
		apierror.Internal(c, err)
		return
	}

	if err := sendVerification(user.ID, user.Email); err != nil {
		// Logged only; the response stays generic so it does not leak state.
		log.Printf("send verification email to %s: %v", user.Email, err)
	}
	c.JSON(http.StatusCreated, gin.H{"message": registerMessage})
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

	// The password is correct, so revealing that the email is unverified does not
	// help enumeration. Only then do we refuse the session.
	if user.EmailVerifiedAt == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Please verify your email before logging in. Check your inbox for the confirmation link."})
		return
	}

	LoginGuard.Success(email)

	startSession(c, user)
	c.JSON(http.StatusOK, user)
}

// Verify consumes a verification token and marks the user's email confirmed.
func Verify(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A verification token is required"})
		return
	}

	userID, err := consumeEmailToken(token, verifyTokenKind)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This verification link is invalid or has expired."})
		return
	}
	if err != nil {
		apierror.Internal(c, err)
		return
	}

	if _, err := database.DB.Exec(
		`UPDATE users SET email_verified_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') WHERE id = $1`, userID,
	); err != nil {
		apierror.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Your email is confirmed. You can log in."})
}

type resendRequest struct {
	Email string `json:"email"`
}

// ResendVerification sends a new verification link when the address needs one.
// The response is always the same, so it cannot be used to check an email.
func ResendVerification(c *gin.Context) {
	var req resendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	const message = "If that account still needs confirmation, we've sent a new link."

	email := strings.ToLower(strings.TrimSpace(req.Email))
	var (
		id       int
		verified sql.NullTime
	)
	err := database.DB.QueryRow(`SELECT id, email_verified_at FROM users WHERE email = $1`, email).Scan(&id, &verified)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		// Unknown address: reply the same as for a known one.
	case err != nil:
		apierror.Internal(c, err)
		return
	case verified.Valid:
		// Already confirmed: nothing to send, same reply.
	default:
		if err := sendVerification(id, email); err != nil {
			log.Printf("resend verification email to %s: %v", email, err)
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": message})
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
