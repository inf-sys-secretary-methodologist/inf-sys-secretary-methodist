// Package cache provides caching implementations using Redis.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache provides caching layer for performance optimization
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(addr string, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

// Get retrieves a value from cache
func (rc *RedisCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	val, err := rc.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil // Cache miss
	}
	if err != nil {
		return false, fmt.Errorf("cache get error: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false, fmt.Errorf("cache unmarshal error: %w", err)
	}

	return true, nil
}

// Set stores a value in cache with TTL
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal error: %w", err)
	}

	if err := rc.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("cache set error: %w", err)
	}

	return nil
}

// Delete removes a value from cache
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	if err := rc.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("cache delete error: %w", err)
	}
	return nil
}

// DeletePattern deletes all keys matching a pattern
func (rc *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	iter := rc.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := rc.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("cache delete pattern error: %w", err)
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("cache scan error: %w", err)
	}
	return nil
}

// Close closes the Redis connection
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

// Ping checks if Redis is reachable
func (rc *RedisCache) Ping(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}

// UserCache provides user-specific caching with DDD patterns
type UserCache struct {
	cache *RedisCache
	ttl   time.Duration
}

// NewUserCache creates a new user cache
func NewUserCache(cache *RedisCache, ttl time.Duration) *UserCache {
	return &UserCache{
		cache: cache,
		ttl:   ttl,
	}
}

// GetUser retrieves user from cache
func (uc *UserCache) GetUser(ctx context.Context, userID int64, dest interface{}) (bool, error) {
	key := fmt.Sprintf("user:%d", userID)
	return uc.cache.Get(ctx, key, dest)
}

// SetUser stores user in cache
func (uc *UserCache) SetUser(ctx context.Context, userID int64, user interface{}) error {
	key := fmt.Sprintf("user:%d", userID)
	return uc.cache.Set(ctx, key, user, uc.ttl)
}

// InvalidateUser removes user from cache
func (uc *UserCache) InvalidateUser(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("user:%d", userID)
	return uc.cache.Delete(ctx, key)
}

// GetUserByEmail retrieves user by email from cache
func (uc *UserCache) GetUserByEmail(ctx context.Context, email string, dest interface{}) (bool, error) {
	key := fmt.Sprintf("user:email:%s", email)
	return uc.cache.Get(ctx, key, dest)
}

// SetUserByEmail stores user by email in cache
func (uc *UserCache) SetUserByEmail(ctx context.Context, email string, user interface{}) error {
	key := fmt.Sprintf("user:email:%s", email)
	return uc.cache.Set(ctx, key, user, uc.ttl)
}

// TokenBlacklist manages invalidated tokens
type TokenBlacklist struct {
	cache *RedisCache
}

// NewTokenBlacklist creates a new token blacklist
func NewTokenBlacklist(cache *RedisCache) *TokenBlacklist {
	return &TokenBlacklist{cache: cache}
}

// Add adds a token to blacklist with TTL matching token expiry
func (tb *TokenBlacklist) Add(ctx context.Context, tokenID string, expiresAt time.Time) error {
	key := fmt.Sprintf("token:blacklist:%s", tokenID)
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil // Token already expired
	}
	return tb.cache.Set(ctx, key, true, ttl)
}

// IsBlacklisted checks if token is blacklisted
func (tb *TokenBlacklist) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := fmt.Sprintf("token:blacklist:%s", tokenID)
	var exists bool
	found, err := tb.cache.Get(ctx, key, &exists)
	return found && exists, err
}

// Client returns the underlying Redis client for advanced operations.
func (rc *RedisCache) Client() *redis.Client {
	return rc.client
}
