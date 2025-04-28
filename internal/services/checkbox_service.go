package services

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// CheckboxService handles operations related to checkboxes
type CheckboxService struct {
	RedisClient *redis.Client
}

// NewCheckboxService creates a new instance of CheckboxService
func NewCheckboxService(redisClient *redis.Client) *CheckboxService {
	return &CheckboxService{
		RedisClient: redisClient,
	}
}

// GetAllCheckboxes retrieves all checkboxes with their states from Redis
// Returns a map where keys are checkbox coordinates and values are their states (true/false)
func (s *CheckboxService) GetAllCheckboxes() (map[string]bool, error) {
	ctx := context.Background()
	keys, err := s.RedisClient.Keys(ctx, "states:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys from Redis: %w", err)
	}
	
	result := make(map[string]bool)
	
	// For each key, get its value
	for _, key := range keys {
		// Get the bit value (0 or 1)
		bitValue, err := s.RedisClient.GetBit(ctx, key, 0).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get value for key %s: %w", key, err)
		}
		
		// Convert bit to boolean (0 -> false, 1 -> true)
		result[key] = bitValue == 1
	}
	
	return result, nil
}

// UpdateCheckboxState updates the state of a checkbox in Redis
func (s *CheckboxService) UpdateCheckboxState(row uint32, column uint32, value bool) error {
	ctx := context.Background()
	
	// Create a key in the format "(row,column)"
	key := fmt.Sprintf("states:(%d,%d)", row, column)
	
    // Convert boolean to bit (0 for false, 1 for true)
    bitValue := 0
    if value {
        bitValue = 1
    }
    // err:= s.RedisClient.HSet(ctx, "states", key, bitValue)
	
    // Set the bit in Redis - using offset 0 since we're storing a single bit per key
    err := s.RedisClient.SetBit(ctx, key, 0, bitValue).Err()
    if err != nil {
		return fmt.Errorf("failed to set bit in Redis: %w", err)
	}

    // Publish a message to notify about the update
    message := fmt.Sprintf("(%d,%d):%t", row, column, value)
    err = s.RedisClient.Publish(ctx, "checkbox_updates", message).Err()
    if err != nil {
        return fmt.Errorf("failed to publish update notification: %w", err)
    }

	return nil
}
