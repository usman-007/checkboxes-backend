package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns a middleware that logs request details
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Log request details
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		// Log using Gin's logger
		gin.DefaultWriter.Write([]byte(
			"[GIN] " + method + " " + path + " " +
				time.Now().Format("2006/01/02 - 15:04:05") + " " +
				"| " + string(rune(status)) + " | " +
				latency.String() + "\n"))
	}
} 