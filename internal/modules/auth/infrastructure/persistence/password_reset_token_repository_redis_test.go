package persistence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
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
	assert.ErrorIs(t, err, domain.ErrPasswordResetTokenNotFound)
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
	assert.ErrorIs(t, err, domain.ErrPasswordResetTokenNotFound)
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
	assert.ErrorIs(t, err, domain.ErrPasswordResetTokenNotFound)
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

// TestRedisPasswordResetTokenRepository_StoresHashedKey pins v0.159.0
// ADR-5: the raw token must never appear in Redis as a key. Storage
// hashes the token with sha256 before SET; lookup hashes the incoming
// token before GET. A Redis read-access attacker holding the dump
// sees only opaque digests and cannot forge a reset link. Issue #279.
func TestRedisPasswordResetTokenRepository_StoresHashedKey(t *testing.T) {
	mr, repo := setupResetTokenRedis(t)
	ctx := context.Background()

	rawToken := "tok-secret-159"
	require.NoError(t, repo.Store(ctx, rawToken, int64(101), time.Minute))

	// The raw token MUST NOT appear as a Redis key — it is the secret
	// being protected. The hashed form (hex sha256) IS what Redis sees.
	rawKey := "pwreset:" + rawToken
	assert.False(t, mr.Exists(rawKey), "raw token must not be stored as a Redis key — sha256 first")

	digest := sha256.Sum256([]byte(rawToken))
	hashedKey := "pwreset:" + hex.EncodeToString(digest[:])
	assert.True(t, mr.Exists(hashedKey), "Redis must store the sha256(token) hex digest as the key")

	// Lookup with the raw token must transparently find the entry by
	// hashing on the read path — callers (handler / use case) must not
	// have to know about the hashing transform.
	uid, err := repo.LookupUser(ctx, rawToken)
	require.NoError(t, err)
	assert.Equal(t, int64(101), uid)
}

// TestRedisPasswordResetTokenRepository_LookupHashesIncoming pins the
// symmetric read-path transform: looking up by the pre-hashed digest
// must NOT find the entry (it is hashed again on lookup, double-hash
// mismatch). Defensive — guards against accidentally bypassing the
// transform from a caller that already has the hash in hand. Issue #279.
func TestRedisPasswordResetTokenRepository_LookupHashesIncoming(t *testing.T) {
	_, repo := setupResetTokenRedis(t)
	ctx := context.Background()

	rawToken := "raw-tok-159"
	require.NoError(t, repo.Store(ctx, rawToken, int64(202), time.Minute))

	digest := sha256.Sum256([]byte(rawToken))
	hashedToken := hex.EncodeToString(digest[:])

	_, err := repo.LookupUser(ctx, hashedToken)
	assert.ErrorIs(t, err, domain.ErrPasswordResetTokenNotFound,
		"lookup by pre-hashed token must be a miss — the read path hashes again, defensive")
}
