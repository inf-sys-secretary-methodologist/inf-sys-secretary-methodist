package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/infrastructure/ratelimit"
)

func newClient(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	c := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = c.Close() })
	return c, mr
}

func TestGenerationLimiter_AllowsUpToLimitThenBlocks(t *testing.T) {
	c, _ := newClient(t)
	lim := ratelimit.NewGenerationLimiter(c, 3, time.Hour)
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		ok, err := lim.Allow(ctx, 7)
		require.NoError(t, err)
		assert.True(t, ok, "call %d must be within a budget of 3", i)
	}
	ok, err := lim.Allow(ctx, 7)
	require.NoError(t, err)
	assert.False(t, ok, "4th call exceeds the budget of 3")
}

func TestGenerationLimiter_PerUserIsolation(t *testing.T) {
	c, _ := newClient(t)
	lim := ratelimit.NewGenerationLimiter(c, 1, time.Hour)
	ctx := context.Background()

	ok, err := lim.Allow(ctx, 1)
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = lim.Allow(ctx, 1)
	require.NoError(t, err)
	assert.False(t, ok, "user 1 is exhausted")

	ok, err = lim.Allow(ctx, 2)
	require.NoError(t, err)
	assert.True(t, ok, "user 2 has an independent budget")
}

func TestGenerationLimiter_WindowResets(t *testing.T) {
	c, mr := newClient(t)
	lim := ratelimit.NewGenerationLimiter(c, 1, time.Hour)
	ctx := context.Background()

	ok, err := lim.Allow(ctx, 9)
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = lim.Allow(ctx, 9)
	require.NoError(t, err)
	assert.False(t, ok)

	mr.FastForward(time.Hour + time.Minute) // expire the window

	ok, err = lim.Allow(ctx, 9)
	require.NoError(t, err)
	assert.True(t, ok, "budget refills after the window elapses")
}

func TestGenerationLimiter_DefaultsOnNonPositive(t *testing.T) {
	c, _ := newClient(t)
	lim := ratelimit.NewGenerationLimiter(c, 0, 0) // falls back to 5/hour
	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		ok, err := lim.Allow(ctx, 3)
		require.NoError(t, err)
		assert.True(t, ok, "call %d must be within the default budget of 5", i)
	}
	ok, err := lim.Allow(ctx, 3)
	require.NoError(t, err)
	assert.False(t, ok, "6th call exceeds the default budget of 5")
}
