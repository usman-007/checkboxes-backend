package config

import (
	"os"
)

// Config holds all configuration for the application
type Config struct {
	Environment   string
	ServerAddress string
	Redis         RedisConfig
	// Add more configuration fields as needed (database, etc.)
}

// RedisConfig holds Redis-specific configuration
type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	// Redis DB is typically a number from 0-15
	redisDB := 0

	return &Config{
		Environment:   env,
		ServerAddress: addr,
		Redis: RedisConfig{
			Address:  redisAddr,
			Password: redisPassword,
			DB:       redisDB,
		},
	}, nil
} 