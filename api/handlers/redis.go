package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/usman-007/checkbox-backend/internal/redis"
)

// RedisTestHandler handles Redis test requests
type RedisTestHandler struct {
	redisClient *redis.Client
}

// NewRedisTestHandler creates a new Redis test handler
func NewRedisTestHandler(redisClient *redis.Client) *RedisTestHandler {
	return &RedisTestHandler{
		redisClient: redisClient,
	}
}

// TestRedis handles the request to test Redis
func (h *RedisTestHandler) TestRedis(c *gin.Context) {
	ctx := c.Request.Context()
	key := "test:key"
	value := time.Now().String()

	// Set a value in Redis
	err := h.redisClient.Set(ctx, key, value, 1*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to set value in Redis",
			"error":   err.Error(),
		})
		return
	}

	// Get the value from Redis
	retrievedValue, err := h.redisClient.Get(ctx, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get value from Redis",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          "ok",
		"message":         "Redis is working properly",
		"stored_value":    value,
		"retrieved_value": retrievedValue,
	})
}

// implent the clear redis function here
func (h *RedisTestHandler) ClearRedis(c *gin.Context) {
	ctx := c.Request.Context()
	err := h.redisClient.FlushAll(ctx).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to clear Redis database: " + err.Error(),
		})
	}
}