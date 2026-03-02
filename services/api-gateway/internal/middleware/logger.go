package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger logs each incoming request with method, path, status and latency.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fmt.Printf("[GATEWAY] %s | %d | %v | %s %s\n",
			time.Now().Format("15:04:05"),
			status,
			latency,
			c.Request.Method,
			path,
		)
	}
}
