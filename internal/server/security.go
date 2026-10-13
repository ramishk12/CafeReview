package server

import "github.com/gin-gonic/gin"

// apiCSP locks down API and upload responses. They are JSON or image bytes, so
// nothing needs to load. This also protects the /uploads route, where an
// uploaded file is served from our origin.
const apiCSP = "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'; sandbox"

// securityHeaders sets hardening headers on every API response.
func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy", apiCSP)
		// Keep other origins from embedding our API or uploads.
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		c.Next()
	}
}
