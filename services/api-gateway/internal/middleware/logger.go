package middleware

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger logs each incoming request with method, path, status and latency.
// It preserves the request body for downstream handlers.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Buffer the request body to preserve it
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			// Restore body for downstream handlers
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fmt.Printf("[GATEWAY] %s | %d | %v | %s %s\n",
			time.Now().Format("15:04:05"),
			status,
			latency,
			method,
			path,
		)
	}
}
