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

func setupTracker(t *testing.T) (*miniredis.Miniredis, *RedisLoginAttemptTracker) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return mr, NewRedisLoginAttemptTracker(client, 5, 15*time.Minute)
}

func TestRedisLoginAttemptTracker_LockoutAtThreshold(t *testing.T) {
	_, tr := setupTracker(t)
	ctx := context.Background()

	// 4 failures — under threshold, not locked.
	for i := 0; i < 4; i++ {
		n, err := tr.RegisterFailure(ctx, "alice@x")
		require.NoError(t, err)
		assert.Equal(t, i+1, n)
	}
	locked, err := tr.IsLocked(ctx, "alice@x")
	require.NoError(t, err)
	assert.False(t, locked, "4 failures < threshold 5 — must not be locked")

	// 5th failure flips the lockout.
	n, err := tr.RegisterFailure(ctx, "alice@x")
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	locked, err = tr.IsLocked(ctx, "alice@x")
	require.NoError(t, err)
	assert.True(t, locked, "5 failures = threshold — must be locked")
}

func TestRedisLoginAttemptTracker_ResetClearsCounter(t *testing.T) {
	_, tr := setupTracker(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_, _ = tr.RegisterFailure(ctx, "bob@x")
	}
	locked, _ := tr.IsLocked(ctx, "bob@x")
	require.True(t, locked)

	require.NoError(t, tr.Reset(ctx, "bob@x"))
	locked, err := tr.IsLocked(ctx, "bob@x")
	require.NoError(t, err)
	assert.False(t, locked, "Reset must drop the counter back to zero")
}

func TestRedisLoginAttemptTracker_IdentifierCaseInsensitive(t *testing.T) {
	_, tr := setupTracker(t)
	ctx := context.Background()

	// Same logical account, three case variants.
	_, _ = tr.RegisterFailure(ctx, "Carol@X.com")
	_, _ = tr.RegisterFailure(ctx, "carol@x.com")
	n, err := tr.RegisterFailure(ctx, "  CAROL@X.COM  ")
	require.NoError(t, err)
	assert.Equal(t, 3, n, "case + whitespace differences must share the same counter")
}

func TestRedisLoginAttemptTracker_TTLExpiresCounter(t *testing.T) {
	mr, tr := setupTracker(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_, _ = tr.RegisterFailure(ctx, "dave@x")
	}
	locked, _ := tr.IsLocked(ctx, "dave@x")
	require.True(t, locked)

	// Advance virtual clock past the 15-minute window.
	mr.FastForward(16 * time.Minute)

	locked, err := tr.IsLocked(ctx, "dave@x")
	require.NoError(t, err)
	assert.False(t, locked, "TTL must self-clean the counter so the user is not permanently locked out")
}
