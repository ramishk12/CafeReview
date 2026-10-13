package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"time"

	"cafe-review/internal/database"

	"github.com/gin-gonic/gin"
)

const (
	// SessionCookie holds the opaque session token. It is HttpOnly, so scripts cannot read it.
	SessionCookie = "cafe_session"
	// CSRFCookie holds the CSRF token. It is readable by scripts so the client can echo it in a header.
	CSRFCookie = "cafe_csrf"

	sessionTTL = 30 * 24 * time.Hour
)

// CookieSecure sets the Secure flag on the session and CSRF cookies. main sets it
// from config. Local http development sets it false; production must be true.
var CookieSecure = true

type session struct {
	UserID    int
	Role      string
	CSRFToken string
}

// newToken returns a random, URL-safe token.
func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// hashToken returns the value stored in the database. Only the hash is kept, so
// a database leak does not expose usable session tokens.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// createSession stores a new session for the user and returns the raw session
// token and the CSRF token.
func createSession(userID int) (token, csrf string, err error) {
	if token, err = newToken(); err != nil {
		return "", "", err
	}
	if csrf, err = newToken(); err != nil {
		return "", "", err
	}
	expires := time.Now().UTC().Add(sessionTTL)
	_, err = database.DB.Exec(
		`INSERT INTO sessions (token_hash, user_id, csrf_token, expires_at) VALUES ($1, $2, $3, $4)`,
		hashToken(token), userID, csrf, expires,
	)
	if err != nil {
		return "", "", err
	}
	return token, csrf, nil
}

// lookupSession returns the session and its user's current role. Expired
// sessions are treated as missing, so the caller returns sql.ErrNoRows.
func lookupSession(token string) (session, error) {
	var s session
	err := database.DB.QueryRow(
		`SELECT s.user_id, s.csrf_token, u.role
		 FROM sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.token_hash = $1 AND s.expires_at > (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')`,
		hashToken(token),
	).Scan(&s.UserID, &s.CSRFToken, &s.Role)
	return s, err
}

// deleteSession removes the session for the given raw token.
func deleteSession(token string) error {
	_, err := database.DB.Exec(`DELETE FROM sessions WHERE token_hash = $1`, hashToken(token))
	return err
}

func setSessionCookies(c *gin.Context, token, csrf string) {
	maxAge := int(sessionTTL.Seconds())
	http.SetCookie(c.Writer, &http.Cookie{
		Name: SessionCookie, Value: token, Path: "/", MaxAge: maxAge,
		HttpOnly: true, Secure: CookieSecure, SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name: CSRFCookie, Value: csrf, Path: "/", MaxAge: maxAge,
		HttpOnly: false, Secure: CookieSecure, SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookies(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name: SessionCookie, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: CookieSecure, SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name: CSRFCookie, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: false, Secure: CookieSecure, SameSite: http.SameSiteLaxMode,
	})
}
