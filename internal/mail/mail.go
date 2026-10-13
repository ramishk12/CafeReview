// Package mail sends transactional email. One interface covers local
// development and production providers, so the sender can be swapped without
// changes elsewhere.
package mail

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

// Message is one outbound email.
type Message struct {
	To      string
	Subject string
	Body    string
}

// Sender sends email.
type Sender interface {
	Send(m Message) error
}

// Log writes messages to the server log. It is used when no SMTP host is
// configured, so local development works without a mail server.
type Log struct{}

func (Log) Send(m Message) error {
	log.Printf("mail (no SMTP host configured)\n  to: %s\n  subject: %s\n%s", m.To, m.Subject, m.Body)
	return nil
}

// SMTP sends through an SMTP server. Port 587 with STARTTLS is typical for
// providers. Mailpit on port 1025 needs no TLS or authentication.
type SMTP struct {
	Addr     string // host:port
	From     string // may include a display name, e.g. "CafeReview <no-reply@example.com>"
	Username string
	Password string
}

func (s SMTP) Send(m Message) error {
	var auth smtp.Auth
	if s.Username != "" {
		host := s.Addr
		if i := strings.LastIndex(host, ":"); i >= 0 {
			host = host[:i]
		}
		auth = smtp.PlainAuth("", s.Username, s.Password, host)
	}
	// The envelope sender must be a bare address, not the display-name form.
	return smtp.SendMail(s.Addr, auth, address(s.From), []string{m.To}, []byte(s.build(m)))
}

func (s SMTP) build(m Message) string {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", s.From)
	fmt.Fprintf(&b, "To: %s\r\n", m.To)
	fmt.Fprintf(&b, "Subject: %s\r\n", m.Subject)
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(m.Body)
	b.WriteString("\r\n")
	return b.String()
}

// address extracts the bare email address from a "Name <addr>" value.
func address(value string) string {
	if i := strings.LastIndex(value, "<"); i >= 0 {
		if j := strings.LastIndex(value, ">"); j > i {
			return strings.TrimSpace(value[i+1 : j])
		}
	}
	return strings.TrimSpace(value)
}
