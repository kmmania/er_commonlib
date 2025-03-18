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

const (
	// CachedLifetime represents the duration for which a cached item remains valid.
	// Set to 1 hour.
	CachedLifetime = 1 * time.Hour

	// CachedTimeout represents the maximum time to wait for a cache operation (e.g., retrieving an item)
	// Set to 2 seconds.
	CachedTimeout = 2 * time.Second
)

// RedisCache defines the contract for interacting with a Redis-based cache.
// It provides methods to perform common cache operations like retrieving, storing, and invalidating data.
type RedisCache interface {

	// Get retrieves a value from the cache by its key, applying a timeout.
	Get(ctx context.Context, key string, dest interface{}, timeout time.Duration) error

	// Set sets a value in the cache with a specified TTL, applying a timeout.
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration, timeout time.Duration)

	// Delete removes a value from the cache by its key, applying a timeout.
	Delete(ctx context.Context, key string, timeout time.Duration) error

	// GetFromCache retrieves data from the cache if it exists.
	GetFromCache(ctx context.Context, key string, logger logger.Logger, target interface{}) (bool, error)

	// SetCache stores data in the cache with the specified key.
	SetCache(ctx context.Context, key string, data interface{}, logger logger.Logger)

	// InvalidateCache removes cached data for a specific key.
	InvalidateCache(ctx context.Context, key string, logger logger.Logger) error
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

// GetFromCache retrieves data from the cache if it exists.
// It checks the cache for the given key and populates the target object if the data is found.
// If the data is not found (cache miss), it returns false with no error.
// If a Redis error occurs, it returns false with the error.
//
// Parameters:
//   - ctx (context.Context): The context for the cache operation.
//   - key (string): The key used to retrieve the data from the cache.
//   - logger (logger.Logger): The logger instance for logging cache operations.
//   - target (interface{}): A pointer to the object where the cached data will be stored.
//
// Returns:
//   - bool: True if the data was found in the cache, false otherwise.
//   - error: An error if the cache operation fails, or nil if successful.
func (cm *RedisCacheManager) GetFromCache(
	ctx context.Context,
	key string,
	logger logger.Logger,
	target interface{},
) (bool, error) {
	err := cm.Get(ctx, key, target, CachedTimeout)
	if err == nil {
		logger.Info("Response from cache", zap.String("cacheKey", key))
		return true, nil
	} else if errors.Is(err, ErrCacheMiss) {
		logger.Info("Cache miss", zap.String("cacheKey", key))
		return false, nil
	} else {
		logger.Error("Cache access error", zap.Error(err))
		return false, err
	}
}

// SetCache stores data in the cache with the specified key.
// It logs the operation and does not return an error, as the underlying cache manager's Set method does not return one.
//
// Parameters:
//   - ctx (context.Context): The context for the cache operation.
//   - key (string): The key under which the data will be stored in the cache.
//   - data (interface{}): The data to be stored in the cache.
//   - logger (logger.Logger): The logger instance for logging cache operations.
func (cm *RedisCacheManager) SetCache(ctx context.Context, key string, data interface{}, logger logger.Logger) {
	cm.Set(ctx, key, data, CachedLifetime, CachedTimeout)
	logger.Info("Data cached successfully", zap.String("cacheKey", key))
}

// InvalidateCache removes cached data for a specific key.
// It logs the operation and returns an error if the cache invalidation fails.
//
// Parameters:
//   - ctx (context.Context): The context for the cache operation.
//   - key (string): The key for which the cached data should be invalidated.
//   - logger (logger.Logger): The logger instance for logging cache operations.
//
// Returns:
//   - error: An error if the cache invalidation fails, or nil if successful.
func (cm *RedisCacheManager) InvalidateCache(ctx context.Context, key string, logger logger.Logger) error {
	err := cm.Delete(ctx, key, CachedTimeout)
	if err != nil {
		logger.Error("Failed to invalidate cache", zap.Error(err))
		return err
	}
	logger.Info("Cache invalidated successfully", zap.String("cacheKey", key))
	return nil
}
