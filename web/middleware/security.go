package middleware

import (
	"github.com/gin-gonic/gin"
)

func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' data:; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'"
		c.Header("Content-Security-Policy", csp)
		c.Header("Server", "")
		
		c.Next()
	}
}

func SessionSecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.TLS != nil {
			c.Header("Set-Cookie", "HttpOnly; Secure; SameSite=Strict")
		} else {
			c.Header("Set-Cookie", "HttpOnly; SameSite=Strict")
		}
		c.Next()
	}
}