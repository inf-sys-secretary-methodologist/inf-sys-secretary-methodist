// Package persistence contains repository implementations for the auth module.
package persistence

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// revokedTokenKeyPrefix isolates the keyspace from other Redis users.
const revokedTokenKeyPrefix = "jwt:revoked:" // #nosec G101 -- Redis key namespace, not a credential

// userRevokeEpochKeyPrefix stores the per-user revocation epoch used by
// RevokeAllForUser / IsRevokedForUser (RFC 6749 §10.4 token-family
// revocation). Issue #279 ADR-2.
const userRevokeEpochKeyPrefix = "jwt:user_revoke_epoch:" // #nosec G101 -- Redis key namespace, not a credential

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

// RevokeIfAbsent atomically claims the JTI via Redis SET NX. claimed=true
// when THIS call performed the revocation; claimed=false when the JTI was
// already revoked (another caller / a prior reuse-detection got there
// first). Used by refresh-token rotation to close the IsRevoked → Revoke
// race window. Issue #279 ADR-2.
func (r *RedisRevokedTokenRepository) RevokeIfAbsent(ctx context.Context, jti string, ttl time.Duration) (bool, error) {
	if jti == "" {
		return false, errors.New("empty jti")
	}
	if ttl <= 0 {
		return false, errors.New("non-positive ttl")
	}
	key := revokedTokenKeyPrefix + jti
	ok, err := r.client.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis SETNX %s: %w", key, err)
	}
	return ok, nil
}

// RevokeAllForUser stores a per-user revocation epoch — every token
// whose iat is at or before issuedAtUnix is considered revoked for the
// given userID. RFC 6749 §10.4 token-family revocation. The epoch
// monotonically advances: if a higher epoch is already present, this
// call leaves it alone (a later reuse-detection event must not weaken
// an earlier one). Issue #279 ADR-2.
func (r *RedisRevokedTokenRepository) RevokeAllForUser(ctx context.Context, userID int64, issuedAtUnix int64, ttl time.Duration) error {
	if ttl <= 0 {
		return errors.New("non-positive ttl")
	}
	key := fmt.Sprintf("%s%d", userRevokeEpochKeyPrefix, userID)
	// Lua script: max(currentEpoch, newEpoch), set TTL afresh so the
	// entry doesn't expire while related refresh tokens are still alive.
	luaScript := `
		local key = KEYS[1]
		local epoch = tonumber(ARGV[1])
		local ttl = tonumber(ARGV[2])
		local cur = tonumber(redis.call("GET", key) or "0")
		if epoch > cur then
			redis.call("SET", key, epoch)
		end
		redis.call("EXPIRE", key, ttl)
		return 1
	`
	if _, err := r.client.Eval(ctx, luaScript, []string{key}, issuedAtUnix, int64(ttl.Seconds())).Result(); err != nil {
		return fmt.Errorf("redis EVAL %s: %w", key, err)
	}
	return nil
}

// IsRevokedForUser reports whether the user's revocation epoch is
// later than issuedAtUnix — i.e., a RevokeAllForUser call has been
// made since this token was minted. Used by RefreshToken / access-
// token validators to honor cascade revokes.
func (r *RedisRevokedTokenRepository) IsRevokedForUser(ctx context.Context, userID int64, issuedAtUnix int64) (bool, error) {
	key := fmt.Sprintf("%s%d", userRevokeEpochKeyPrefix, userID)
	raw, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("redis GET %s: %w", key, err)
	}
	epoch, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return false, fmt.Errorf("parse epoch from %s: %w", key, err)
	}
	// Strict greater-than (not >=): tokens with iat == epoch survive.
	// This matters for concurrent-refresh races where multiple callers
	// share the same iat — the loser's cascade sets epoch = iat, but
	// the winner who already claimed the JTI atomically must not be
	// caught by their own peer's cascade. The reused token's JTI is
	// already in the per-JTI blacklist independently, so per-token
	// safety is unaffected.
	return epoch > issuedAtUnix, nil
}
