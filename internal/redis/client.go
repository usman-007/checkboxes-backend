package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/usman-007/checkbox-backend/config"
)

// Client wraps the Redis client
type Client struct {
	*redis.Client
}

// NewClient creates a new Redis client
func NewClient(cfg *config.RedisConfig) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Client{client}, nil
}

func (c *Client) InitializeGridState(ctx context.Context, rows, cols int) error {
	// Create a pipeline
	pipe := c.Client.Pipeline()

	// Iterate through each cell of the grid
	for r := 0; r <= rows; r++ {
		for col := 0; col <= cols; col++ {
			// Format the key according to your specified format
			key := fmt.Sprintf("states:(%d,%d)", r, col)

			// Queue the SET command in the pipeline.
			// We set the value to 0 (representing false).
			// We use 0 for expiration, meaning the key won't expire.
			pipe.Set(ctx, key, 0, 0)
		}
	}

	// Execute all commands in the pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute grid initialization pipeline: %w", err)
	}

	fmt.Printf("Successfully initialized %d x %d grid states to 0.\n", rows, cols)
	return nil
}


// Set stores a key-value pair with expiration
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.Client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// Delete removes a key
func (c *Client) Delete(ctx context.Context, key string) error {
	return c.Client.Del(ctx, key).Err()
}

// HashSet sets a field in a hash stored at key
func (c *Client) HashSet(ctx context.Context, key, field string, value interface{}) error {
	return c.Client.HSet(ctx, key, field, value).Err()
}

// HashGet gets a field from a hash stored at key
func (c *Client) HashGet(ctx context.Context, key, field string) (string, error) {
	return c.Client.HGet(ctx, key, field).Result()
}

// HashGetAll gets all fields from a hash stored at key
func (c *Client) HashGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.Client.HGetAll(ctx, key).Result()
} 