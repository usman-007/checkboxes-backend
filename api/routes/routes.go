package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/usman-007/checkbox-backend/api/handlers"
	"github.com/usman-007/checkbox-backend/internal/redis"
	"github.com/usman-007/checkbox-backend/internal/services"
)

// Setup configures all routes for the application
func Setup(router *gin.Engine, redisClient *redis.Client) {
	// Health check
	router.GET("/health", handlers.HealthCheck)
	
	// Initialize services
	checkboxService := services.NewCheckboxService(redisClient.Client)
	
	// Initialize handlers
	checkboxHandler := handlers.NewCheckboxHandler(checkboxService)
	redisTestHandler := handlers.NewRedisTestHandler(redisClient)
	websocketHandler := handlers.NewWebSocketHandler(checkboxService, redisClient.Client)
	
	// Start Redis subscription for WebSocket updates in a goroutine
	go websocketHandler.StartRedisSubscription()

	redis := router.Group("/redis")
	{
		redis.DELETE("", redisTestHandler.ClearRedis)
		redis.GET("", redisTestHandler.TestRedis)
	}
	
	// WebSocket endpoint
	router.GET("/ws", websocketHandler.HandleWebSocket)
	
	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// redis routes
		redis := v1.Group("/redis")
		{
			redis.DELETE("", redisTestHandler.ClearRedis)
			redis.GET("", redisTestHandler.TestRedis)
		}
		
		// Checkbox routes
		checkbox := v1.Group("/checkbox")
		{
			checkbox.GET("", checkboxHandler.GetAllCheckboxes)
			checkbox.PATCH("", checkboxHandler.UpdateCheckbox)
		}
		
		// WebSocket endpoint in API v1
		v1.GET("/ws", websocketHandler.HandleWebSocket)
	}
}

/*
redis 
curl http://localhost:8080/api/v1/redis // TEST REDIS
curl -X DELETE http://localhost:8080/api/v1/redis // CLEAR REDIS

checkbox
curl http://localhost:8080/api/v1/checkbox // GET ALL CHECKBOXES
curl -X PATCH http://localhost:8080/api/v1/checkbox?row=1$column=2&value=true
*/
