package auth

import (
	"net/url"
	"strings"
	"time"

	"cafe-review/internal/database"
	"cafe-review/internal/mail"
)

const (
	verifyTokenKind = "verify"
	verifyTokenTTL  = 24 * time.Hour
)

// Mailer sends transactional email. main sets it at startup. The default writes
// messages to the log, so local development works without a mail server.
var Mailer mail.Sender = mail.Log{}

// AppBaseURL is the public origin of the frontend, used to build links in emails.
var AppBaseURL = "http://localhost:3000"

// createEmailToken stores a single-use token and returns the raw value.
func createEmailToken(userID int, kind string, ttl time.Duration) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}
	expires := time.Now().UTC().Add(ttl)
	_, err = database.DB.Exec(
		`INSERT INTO email_tokens (user_id, token_hash, kind, expires_at) VALUES ($1, $2, $3, $4)`,
		userID, hashToken(token), kind, expires,
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

// consumeEmailToken marks a token used and returns its user ID. It returns
// sql.ErrNoRows when the token is unknown, already used, or expired.
func consumeEmailToken(token, kind string) (int, error) {
	var userID int
	err := database.DB.QueryRow(
		`UPDATE email_tokens SET used_at = (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		 WHERE token_hash = $1 AND kind = $2 AND used_at IS NULL
		   AND expires_at > (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
		 RETURNING user_id`,
		hashToken(token), kind,
	).Scan(&userID)
	return userID, err
}

// sendVerification creates a verification token and emails the link. The caller
// logs failures but keeps its response generic, so nothing is revealed.
func sendVerification(userID int, email string) error {
	token, err := createEmailToken(userID, verifyTokenKind, verifyTokenTTL)
	if err != nil {
		return err
	}
	link := strings.TrimRight(AppBaseURL, "/") + "/verify?token=" + url.QueryEscape(token)
	body := "Welcome to CafeReview.\r\n\r\n" +
		"Confirm your email address by opening this link:\r\n\r\n" +
		link + "\r\n\r\n" +
		"If you did not create an account, you can ignore this message.\r\n"
	return Mailer.Send(mail.Message{
		To:      email,
		Subject: "Confirm your CafeReview email",
		Body:    body,
	})
}
