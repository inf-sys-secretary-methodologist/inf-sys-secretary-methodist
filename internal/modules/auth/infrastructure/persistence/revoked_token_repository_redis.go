// Package persistence contains repository implementations for the auth module.
package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// revokedTokenKeyPrefix isolates the keyspace from other Redis users.
const revokedTokenKeyPrefix = "jwt:revoked:"

// RedisRevokedTokenRepository stores revoked JTIs in Redis with auto-expiry.
// Each entry's TTL equals the token's remaining lifetime, so the keyspace
// stays bounded without an explicit cleanup cron.
type RedisRevokedTokenRepository struct {
	client *redis.Client
}

// NewRedisRevokedTokenRepository wires a repository onto an existing Redis
// client. The caller is responsible for client lifecycle (ping, close).
func NewRedisRevokedTokenRepository(client *redis.Client) *RedisRevokedTokenRepository {
	return &RedisRevokedTokenRepository{client: client}
}

// Revoke marks the JTI as revoked with the given TTL.
// A zero or negative TTL is rejected — the caller must filter expired
// tokens before reaching here (LogoutUseCase already does).
func (r *RedisRevokedTokenRepository) Revoke(ctx context.Context, jti string, ttl time.Duration) error {
	if jti == "" {
		return errors.New("empty jti")
	}
	if ttl <= 0 {
		return errors.New("non-positive ttl")
	}
	key := revokedTokenKeyPrefix + jti
	if err := r.client.Set(ctx, key, "1", ttl).Err(); err != nil {
		return fmt.Errorf("redis SET %s: %w", key, err)
	}
	return nil
}

// IsRevoked reports whether the JTI is currently in the revocation set.
// EXISTS returns 0 for absent or expired keys, so the caller does not need
// to distinguish "never revoked" from "revoked-but-now-expired".
func (r *RedisRevokedTokenRepository) IsRevoked(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, nil
	}
	key := revokedTokenKeyPrefix + jti
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis EXISTS %s: %w", key, err)
	}
	return n > 0, nil
}
