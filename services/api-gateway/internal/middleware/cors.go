package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS handles cross-origin requests from the frontend.
// The frontend runs on localhost:3000, the gateway on localhost:8080.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowedMap := make(map[string]bool)
	for _, o := range allowedOrigins {
		allowedMap[o] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if allowedMap[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// Respond to preflight OPTIONS requests immediately
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
