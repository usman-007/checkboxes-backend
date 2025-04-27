package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/usman-007/checkbox-backend/api/middleware"
	"github.com/usman-007/checkbox-backend/api/routes"
	"github.com/usman-007/checkbox-backend/config"
	"github.com/usman-007/checkbox-backend/internal/redis"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Redis client
	redisClient, err := redis.NewClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

		// Define grid dimensions
	gridRows := 20
	gridCols := 20

	// Create a context
	ctx := context.Background()

	// Initialize the grid state
	err = redisClient.InitializeGridState(ctx, gridRows, gridCols)
	if err != nil {
		log.Println("Error initializing grid state:", err)
	}


	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := gin.Default()

	// Apply global middleware
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// Register routes
	routes.Setup(router, redisClient)

	// Start server
	if err := router.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 