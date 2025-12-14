package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/smilemakc/auth-gateway/internal/config"
	"github.com/smilemakc/auth-gateway/internal/models"
)

// RedisService provides Redis operations
type RedisService struct {
	client *redis.Client
}

// NewRedisService creates a new Redis service
func NewRedisService(cfg *config.RedisConfig) (*RedisService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisService{client: client}, nil
}

// Close closes the Redis connection
func (r *RedisService) Close() error {
	return r.client.Close()
}

// Set sets a key-value pair with expiration
func (r *RedisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get gets a value by key
func (r *RedisService) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Delete deletes a key
func (r *RedisService) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (r *RedisService) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	return result > 0, err
}

// Increment increments a counter
func (r *RedisService) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// SetNX sets a key only if it doesn't exist
func (r *RedisService) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

// Expire sets an expiration on a key
func (r *RedisService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// Health checks the Redis health
func (r *RedisService) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// AddToBlacklist adds a token to the blacklist
func (r *RedisService) AddToBlacklist(ctx context.Context, tokenHash string, expiration time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", tokenHash)
	return r.Set(ctx, key, "1", expiration)
}

// IsBlacklisted checks if a token is blacklisted
func (r *RedisService) IsBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", tokenHash)
	return r.Exists(ctx, key)
}

// IncrementRateLimit increments the rate limit counter
func (r *RedisService) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	count, err := r.Increment(ctx, key)
	if err != nil {
		return 0, err
	}

	// Set expiration if this is the first increment
	if count == 1 {
		if err := r.Expire(ctx, key, window); err != nil {
			return 0, err
		}
	}

	return count, nil
}

// StorePendingRegistration stores pending registration data in Redis
func (r *RedisService) StorePendingRegistration(ctx context.Context, identifier string, data *models.PendingRegistration, expiration time.Duration) error {
	key := fmt.Sprintf("pending:registration:%s", identifier)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal pending registration: %w", err)
	}
	return r.Set(ctx, key, jsonData, expiration)
}

// GetPendingRegistration retrieves pending registration data from Redis
func (r *RedisService) GetPendingRegistration(ctx context.Context, identifier string) (*models.PendingRegistration, error) {
	key := fmt.Sprintf("pending:registration:%s", identifier)
	data, err := r.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	var pending models.PendingRegistration
	if err := json.Unmarshal([]byte(data), &pending); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pending registration: %w", err)
	}
	return &pending, nil
}

// DeletePendingRegistration deletes pending registration data from Redis
func (r *RedisService) DeletePendingRegistration(ctx context.Context, identifier string) error {
	key := fmt.Sprintf("pending:registration:%s", identifier)
	return r.Delete(ctx, key)
}
