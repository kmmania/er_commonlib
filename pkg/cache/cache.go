/*
Package cache provides a Redis-based caching mechanism for managing temporary data storage.
It includes functionality for retrieving, storing, and invalidating cached data, helping
to reduce database load and improve application performance.
*/
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/kmmania/er_commonlib/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ErrCacheMiss indicates that the requested key was not found in the cache.
var ErrCacheMiss = errors.New("cache miss")

// RedisCache defines the contract for interacting with a Redis-based cache.
// It provides methods to perform common cache operations like retrieving, storing, and invalidating data.
type RedisCache interface {

	// Get retrieves a value from the cache by its key, applying a timeout.
	Get(ctx context.Context, key string, dest interface{}, timeout time.Duration) error

	// Set sets a value in the cache with a specified TTL, applying a timeout.
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration, timeout time.Duration)

	// Delete removes a value from the cache by its key, applying a timeout.
	Delete(ctx context.Context, key string, timeout time.Duration) error
}

// RedisCacheManager is a struct that manages interactions with a Redis cache instance.
// It provides methods to retrieve, store, and delete data in the cache.
type RedisCacheManager struct {
	// client is a pointer to the Redis client used for executing cache operations.
	client *redis.Client
	// logger provides structured logging for this RedisCacheManager's operations.
	logger logger.Logger
}

// New creates and returns a new RedisCacheManager instance.
//
// Parameters:
// - client (*redis.Client): The Redis client used to interact with the Redis server.
// - logger (*zap.Logger): A logger instance for logging server activities and errors.
//
// Returns:
// - *RedisCacheManager: An initialized RedisCacheManager.
func New(client *redis.Client, logger logger.Logger) *RedisCacheManager {
	return &RedisCacheManager{
		client: client,
		logger: logger,
	}
}

// Get retrieves a value from the cache by its key, applying a timeout.
// If the operation exceeds the specified timeout duration, it is cancelled automatically.
//
// Parameters:
// - ctx (context.Context): The base context for the operation.
// - key (string): The key of the cache entry to retrieve.
// - dest (interface{}): A pointer to the variable where the retrieved value should be unmarshalled.
// - timeout (time.Duration): The maximum time allowed for the operation.
//
// Returns:
// - error: An error if the operation fails, times out, or the key is not found.
func (cm *RedisCacheManager) Get(ctx context.Context, key string, dest interface{}, timeout time.Duration) error {
	// Create a new context with timeout.
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel() // Ensure the context is cancelled after the operation

	// Try to retrieve the value from Redis.
	data, err := cm.client.Get(ctxWithTimeout, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			cm.logger.Debug("cache miss", zap.String("key", key), zap.Duration("timeout", timeout))
			return ErrCacheMiss
		}
		cm.logger.Error("Error accessing Redis cache", zap.Error(err))
		return err
	}

	// Attempt to unmarshal the JSON data into the destination variable.
	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		cm.logger.Error("Error unmarshalling cache data for key", zap.Error(err))
		return err
	}

	return nil
}

// Set sets a value in the cache with a specified TTL, applying a timeout.
//
// Parameters:
// - ctx (context.Context): The base context for the operation.
// - key (string): The cache key.
// - value (interface{}): The value to cache.
// - ttl (time.Duration): The time-to-live for the cached value.
// - timeout (time.Duration): The maximum time allowed for the operation.
//
// Returns:
// - error: An error if the operation fails or times out.
func (cm *RedisCacheManager) Set(
	ctx context.Context,
	key string,
	value interface{},
	ttl time.Duration,
	timeout time.Duration,
) {
	// Create a new context with timeout.
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Marshal the value to JSON.
	data, err := json.Marshal(value)
	if err != nil {
		cm.logger.Error("Error marshalling data for cache key", zap.String("key", key), zap.Error(err))
		return
	}

	// Perform the Redis Set operation.
	err = cm.client.Set(ctxWithTimeout, key, data, ttl).Err()
	if err != nil {
		cm.logger.Error("Error setting cache for key", zap.String("key", key), zap.Error(err))
	}
}

// Delete removes a value from the cache by its key, applying a timeout.
//
// Parameters:
// - ctx (context.Context): The base context for the operation.
// - key (string): The cache key to delete.
// - timeout (time.Duration): The maximum time allowed for the operation.
//
// Returns:
// - error: An error if the operation fails or times out.
func (cm *RedisCacheManager) Delete(ctx context.Context, key string, timeout time.Duration) error {
	// Create a new context with timeout.
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Perform the Redis Delete operation.
	err := cm.client.Del(ctxWithTimeout, key).Err()
	if err != nil {
		cm.logger.Error("Error deleting cache for key", zap.String("key", key), zap.Error(err))
	} else {
		cm.logger.Info("Cache invalidated for key", zap.String("key", key))
	}

	return nil
}
