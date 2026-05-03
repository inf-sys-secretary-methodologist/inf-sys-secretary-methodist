package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
)

// setupResetTokenRedis spins a miniredis instance and a real go-redis
// client pointed at it. Same fixture style as the rest of the package.
func setupResetTokenRedis(t *testing.T) (*miniredis.Miniredis, *RedisPasswordResetTokenRepository) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return mr, NewRedisPasswordResetTokenRepository(client)
}

// TestRedisPasswordResetTokenRepository_StoreThenLookup verifies the
// happy path round-trip: a stored token returns its associated user id.
func TestRedisPasswordResetTokenRepository_StoreThenLookup(t *testing.T) {
	_, repo := setupResetTokenRedis(t)
	ctx := context.Background()

	require.NoError(t, repo.Store(ctx, "tok-1", int64(42), time.Minute))

	got, err := repo.LookupUser(ctx, "tok-1")
	require.NoError(t, err)
	assert.Equal(t, int64(42), got)
}

// TestRedisPasswordResetTokenRepository_LookupMissingReturnsSentinel —
// an unknown token must surface ErrPasswordResetTokenNotFound (errors.Is
// reachable) so usecase can map cleanly to ErrInvalidResetToken without
// string parsing.
func TestRedisPasswordResetTokenRepository_LookupMissingReturnsSentinel(t *testing.T) {
	_, repo := setupResetTokenRedis(t)

	_, err := repo.LookupUser(context.Background(), "never-stored")
	assert.ErrorIs(t, err, repositories.ErrPasswordResetTokenNotFound)
}

// TestRedisPasswordResetTokenRepository_TTLExpiry — a token past its TTL
// must look the same as a token that never existed (sentinel error).
// Uses miniredis's FastForward to advance virtual time without sleeping.
func TestRedisPasswordResetTokenRepository_TTLExpiry(t *testing.T) {
	mr, repo := setupResetTokenRedis(t)
	ctx := context.Background()

	require.NoError(t, repo.Store(ctx, "tok-ttl", int64(7), time.Second))
	mr.FastForward(2 * time.Second)

	_, err := repo.LookupUser(ctx, "tok-ttl")
	assert.ErrorIs(t, err, repositories.ErrPasswordResetTokenNotFound)
}

// TestRedisPasswordResetTokenRepository_DeleteEnforcesSingleUse — after
// Delete the token must look unknown. This is what makes a reset token
// single-use; otherwise a leaked / replayed link could change the
// password again.
func TestRedisPasswordResetTokenRepository_DeleteEnforcesSingleUse(t *testing.T) {
	_, repo := setupResetTokenRedis(t)
	ctx := context.Background()

	require.NoError(t, repo.Store(ctx, "tok-del", int64(99), time.Minute))
	require.NoError(t, repo.Delete(ctx, "tok-del"))

	_, err := repo.LookupUser(ctx, "tok-del")
	assert.ErrorIs(t, err, repositories.ErrPasswordResetTokenNotFound)
}

// TestRedisPasswordResetTokenRepository_RejectsEmptyOrNonPositiveInputs
// — defensive contract on bad input. Empty token or zero/negative TTL
// would let stale or eternal entries leak in; refuse loudly.
func TestRedisPasswordResetTokenRepository_RejectsEmptyOrNonPositiveInputs(t *testing.T) {
	_, repo := setupResetTokenRedis(t)
	ctx := context.Background()

	cases := []struct {
		name  string
		token string
		uid   int64
		ttl   time.Duration
	}{
		{"empty token", "", 1, time.Minute},
		{"zero ttl", "tok-x", 1, 0},
		{"negative ttl", "tok-x", 1, -time.Second},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Store(ctx, tc.token, tc.uid, tc.ttl)
			assert.Error(t, err)
		})
	}
}
