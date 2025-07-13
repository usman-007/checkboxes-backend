package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/usman-007/checkbox-backend/internal/monitoring"
)

// Logger returns a middleware that logs request details and records metrics
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate metrics
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		// Record Prometheus metrics
		statusStr := strconv.Itoa(status)
		monitoring.HttpRequestsTotal.WithLabelValues(method, path, statusStr).Inc()
		monitoring.HttpRequestDuration.WithLabelValues(method, path).Observe(latency.Seconds())

		// Log using Gin's logger
		gin.DefaultWriter.Write([]byte(
			"[GIN] " + method + " " + path + " " +
				time.Now().Format("2006/01/02 - 15:04:05") + " " +
				"| " + string(rune(status)) + " | " +
				latency.String() + "\n"))
	}
} 