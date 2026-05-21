package persistence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
)

// passwordResetKeyPrefix isolates the keyspace from other Redis users.
// Mirrors the revokedTokenKeyPrefix convention.
const passwordResetKeyPrefix = "pwreset:"

// hashResetToken returns the canonical Redis key for the supplied raw
// reset token. v0.159.0 ADR-5 — the raw token IS the secret in the
// reset email, so it must never appear in Redis. Hashing with sha256
// gives a fixed-length opaque key that an attacker who reads Redis
// cannot reverse to the original link. The transform is applied
// symmetrically on Store / LookupUser / Delete so callers stay
// unaware. Issue #279.
func hashResetToken(rawToken string) string {
	digest := sha256.Sum256([]byte(rawToken))
	return passwordResetKeyPrefix + hex.EncodeToString(digest[:])
}

// RedisPasswordResetTokenRepository persists short-lived password reset
// tokens in Redis with auto-expiry. Each entry maps a token string to
// the user id it grants reset rights for; TTL bounds the link's life
// without an explicit cleanup cron.
type RedisPasswordResetTokenRepository struct {
	client *redis.Client
}

// NewRedisPasswordResetTokenRepository wires a repository onto an
// existing Redis client. The caller owns client lifecycle (ping, close).
func NewRedisPasswordResetTokenRepository(client *redis.Client) *RedisPasswordResetTokenRepository {
	return &RedisPasswordResetTokenRepository{client: client}
}

// Store records that token grants reset rights for userID, valid for
// ttl. Empty token or non-positive ttl are refused — both would let
// stale or eternal entries leak in. v0.159.0 ADR-5: the raw token is
// sha256-hashed before being used as the Redis key.
func (r *RedisPasswordResetTokenRepository) Store(ctx context.Context, token string, userID int64, ttl time.Duration) error {
	if token == "" {
		return errors.New("empty token")
	}
	if ttl <= 0 {
		return errors.New("non-positive ttl")
	}
	key := hashResetToken(token)
	if err := r.client.Set(ctx, key, strconv.FormatInt(userID, 10), ttl).Err(); err != nil {
		return fmt.Errorf("redis SET %s: %w", key, err)
	}
	return nil
}

// LookupUser returns the userID stored under the token, or
// ErrPasswordResetTokenNotFound (from the domain) when the entry is
// absent or has expired. redis.Nil is mapped to the domain sentinel so
// callers do not need to import the redis package. v0.159.0 ADR-5: the
// raw token is sha256-hashed before lookup, symmetric with Store.
func (r *RedisPasswordResetTokenRepository) LookupUser(ctx context.Context, token string) (int64, error) {
	if token == "" {
		return 0, domain.ErrPasswordResetTokenNotFound
	}
	key := hashResetToken(token)
	raw, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return 0, domain.ErrPasswordResetTokenNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("redis GET %s: %w", key, err)
	}
	uid, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse userID from %s: %w", key, err)
	}
	return uid, nil
}

// Delete removes the token. Deleting a missing key is not an error —
// the postcondition (token gone) holds either way. v0.159.0 ADR-5:
// the raw token is sha256-hashed before deletion, symmetric with Store.
func (r *RedisPasswordResetTokenRepository) Delete(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	key := hashResetToken(token)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis DEL %s: %w", key, err)
	}
	return nil
}

// LookupUserAndConsume atomically reads the user ID and removes the
// token via Redis GETDEL (single round-trip, atomic from Redis's
// perspective). v0.159.0 ADR-6 — eliminates the Save-then-Delete
// replay window: after this call the token is gone, so a subsequent
// Save failure cannot leave a usable token behind for a replay.
func (r *RedisPasswordResetTokenRepository) LookupUserAndConsume(ctx context.Context, token string) (int64, error) {
	if token == "" {
		return 0, domain.ErrPasswordResetTokenNotFound
	}
	key := hashResetToken(token)
	raw, err := r.client.GetDel(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return 0, domain.ErrPasswordResetTokenNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("redis GETDEL %s: %w", key, err)
	}
	uid, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse userID from %s: %w", key, err)
	}
	return uid, nil
}
