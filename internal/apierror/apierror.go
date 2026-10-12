// Package apierror writes server error responses that are safe to show clients.
// The real error is logged on the server with a request ID, and the client
// receives only a generic message and that same ID.
package apierror

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

const (
	// HeaderRequestID is set on every response and can be sent by a proxy to keep an ID across hops.
	HeaderRequestID = "X-Request-ID"

	genericMessage = "Something went wrong"
	ctxRequestID   = "requestID"
)

// validRequestID limits incoming IDs to safe characters, so they cannot inject
// content into logs or responses.
var validRequestID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// RequestID gives each request an ID. A valid incoming X-Request-ID is kept;
// anything else is replaced with a random one.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderRequestID)
		if !validRequestID.MatchString(id) {
			id = newID()
		}
		c.Set(ctxRequestID, id)
		c.Header(HeaderRequestID, id)
		c.Next()
	}
}

// RequestIDFrom returns the request's ID, or "" if RequestID has not run.
func RequestIDFrom(c *gin.Context) string {
	v, _ := c.Get(ctxRequestID)
	id, _ := v.(string)
	return id
}

// Internal logs err with the request ID and responds 500 with a generic message.
// The error text is never sent to the client.
func Internal(c *gin.Context, err error) {
	id := RequestIDFrom(c)
	log.Printf("request_id=%s method=%s path=%s error=%v", id, c.Request.Method, c.Request.URL.Path, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": genericMessage, "request_id": id})
}

func newID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "unavailable"
	}
	return hex.EncodeToString(b)
}
