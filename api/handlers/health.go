package handlers

import (
	"net/http"
	"github.com/usman-007/checkbox-backend/internal/monitoring"
	"github.com/gin-gonic/gin"
)

// HealthCheck handles the health check endpoint
func HealthCheck(c *gin.Context) {
	monitoring.HealthCounter.Inc()
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
} 