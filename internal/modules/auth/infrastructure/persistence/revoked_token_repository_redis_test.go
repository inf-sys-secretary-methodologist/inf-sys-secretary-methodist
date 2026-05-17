package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// v0.153.1 #196 coverage push: revoked_token_repository_redis.go was at
// 0% covered. Tests via miniredis (in-process redis stub) — same
// pattern as cached_user_repository_test.go.

func newRevokedRedisRepo(t *testing.T) (*RedisRevokedTokenRepository, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return NewRedisRevokedTokenRepository(client), mr
}

// --- Constructor ---

func TestNewRedisRevokedTokenRepository_StoresClient(t *testing.T) {
	repo, _ := newRevokedRedisRepo(t)
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.client)
}

// --- Revoke ---

func TestRedisRevokedTokenRepository_Revoke_HappyPath(t *testing.T) {
	repo, mr := newRevokedRedisRepo(t)
	ctx := context.Background()

	err := repo.Revoke(ctx, "jti-abc-123", 5*time.Minute)
	require.NoError(t, err)

	// Verify key exists with TTL roughly in expected window.
	val, ok := mr.Keys(), false
	for _, k := range val {
		if k == revokedTokenKeyPrefix+"jti-abc-123" {
			ok = true
			break
		}
	}
	assert.True(t, ok, "key must be set after Revoke")
	ttl := mr.TTL(revokedTokenKeyPrefix + "jti-abc-123")
	assert.True(t, ttl > 0 && ttl <= 5*time.Minute, "TTL within expected window, got %v", ttl)
}

func TestRedisRevokedTokenRepository_Revoke_EmptyJTIRejected(t *testing.T) {
	repo, _ := newRevokedRedisRepo(t)
	err := repo.Revoke(context.Background(), "", time.Minute)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty jti")
}

func TestRedisRevokedTokenRepository_Revoke_ZeroTTLRejected(t *testing.T) {
	repo, _ := newRevokedRedisRepo(t)
	err := repo.Revoke(context.Background(), "jti-x", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-positive ttl")
}

func TestRedisRevokedTokenRepository_Revoke_NegativeTTLRejected(t *testing.T) {
	repo, _ := newRevokedRedisRepo(t)
	err := repo.Revoke(context.Background(), "jti-x", -1*time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-positive ttl")
}

func TestRedisRevokedTokenRepository_Revoke_RedisError(t *testing.T) {
	repo, mr := newRevokedRedisRepo(t)
	mr.Close() // simulate broken redis — SET will fail
	err := repo.Revoke(context.Background(), "jti-fail", time.Minute)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redis SET")
}

// --- IsRevoked ---

func TestRedisRevokedTokenRepository_IsRevoked_PresentReturnsTrue(t *testing.T) {
	repo, _ := newRevokedRedisRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.Revoke(ctx, "jti-present", time.Hour))

	revoked, err := repo.IsRevoked(ctx, "jti-present")
	require.NoError(t, err)
	assert.True(t, revoked)
}

func TestRedisRevokedTokenRepository_IsRevoked_AbsentReturnsFalse(t *testing.T) {
	repo, _ := newRevokedRedisRepo(t)

	revoked, err := repo.IsRevoked(context.Background(), "jti-never-set")
	require.NoError(t, err)
	assert.False(t, revoked)
}

func TestRedisRevokedTokenRepository_IsRevoked_EmptyJTIReturnsFalseNoError(t *testing.T) {
	repo, _ := newRevokedRedisRepo(t)

	revoked, err := repo.IsRevoked(context.Background(), "")
	require.NoError(t, err, "empty jti must not error — defense-in-depth")
	assert.False(t, revoked)
}

func TestRedisRevokedTokenRepository_IsRevoked_ExpiredReturnsFalse(t *testing.T) {
	repo, mr := newRevokedRedisRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.Revoke(ctx, "jti-short", time.Minute))
	// Fast-forward beyond TTL.
	mr.FastForward(2 * time.Minute)

	revoked, err := repo.IsRevoked(ctx, "jti-short")
	require.NoError(t, err)
	assert.False(t, revoked, "expired key must not count as revoked")
}

func TestRedisRevokedTokenRepository_IsRevoked_RedisError(t *testing.T) {
	repo, mr := newRevokedRedisRepo(t)
	mr.Close()

	revoked, err := repo.IsRevoked(context.Background(), "jti-x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redis EXISTS")
	assert.False(t, revoked)
}
