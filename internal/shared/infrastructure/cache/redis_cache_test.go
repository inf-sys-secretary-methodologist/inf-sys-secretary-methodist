package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMiniredis(t *testing.T) (*miniredis.Miniredis, *RedisCache) {
	t.Helper()
	mr := miniredis.RunT(t)
	rc, err := NewRedisCache(mr.Addr(), "", 0)
	require.NoError(t, err)
	t.Cleanup(func() { rc.Close() })
	return mr, rc
}

func TestNewRedisCache(t *testing.T) {
	_, rc := setupMiniredis(t)
	assert.NotNil(t, rc)
	assert.NotNil(t, rc.Client())
}

func TestNewRedisCache_ConnectionFailure(t *testing.T) {
	_, err := NewRedisCache("localhost:1", "", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to Redis")
}

func TestRedisCache_SetAndGet(t *testing.T) {
	_, rc := setupMiniredis(t)
	ctx := context.Background()

	type testData struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	err := rc.Set(ctx, "test:key", testData{Name: "Alice", Age: 30}, 5*time.Minute)
	require.NoError(t, err)

	var result testData
	found, err := rc.Get(ctx, "test:key", &result)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "Alice", result.Name)
	assert.Equal(t, 30, result.Age)
}

func TestRedisCache_Get_CacheMiss(t *testing.T) {
	_, rc := setupMiniredis(t)
	ctx := context.Background()

	var result string
	found, err := rc.Get(ctx, "nonexistent", &result)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestRedisCache_Get_UnmarshalError(t *testing.T) {
	mr, rc := setupMiniredis(t)
	ctx := context.Background()

	// Set raw invalid JSON directly in miniredis
	mr.Set("bad:json", "not-valid-json{{{")

	var result map[string]string
	found, err := rc.Get(ctx, "bad:json", &result)
	assert.False(t, found)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cache unmarshal error")
}

func TestRedisCache_Delete(t *testing.T) {
	_, rc := setupMiniredis(t)
	ctx := context.Background()

	err := rc.Set(ctx, "del:key", "value", 5*time.Minute)
	require.NoError(t, err)

	err = rc.Delete(ctx, "del:key")
	require.NoError(t, err)

	var result string
	found, err := rc.Get(ctx, "del:key", &result)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestRedisCache_DeletePattern(t *testing.T) {
	_, rc := setupMiniredis(t)
	ctx := context.Background()

	for i := range 3 {
		key := "pattern:" + string(rune('a'+i))
		err := rc.Set(ctx, key, "val", 5*time.Minute)
		require.NoError(t, err)
	}

	err := rc.DeletePattern(ctx, "pattern:*")
	require.NoError(t, err)

	for i := range 3 {
		key := "pattern:" + string(rune('a'+i))
		var result string
		found, err := rc.Get(ctx, key, &result)
		require.NoError(t, err)
		assert.False(t, found, "key %s should be deleted", key)
	}
}

func TestRedisCache_Ping(t *testing.T) {
	_, rc := setupMiniredis(t)
	ctx := context.Background()

	err := rc.Ping(ctx)
	assert.NoError(t, err)
}

func TestRedisCache_Close(t *testing.T) {
	mr := miniredis.RunT(t)
	rc, err := NewRedisCache(mr.Addr(), "", 0)
	require.NoError(t, err)

	err = rc.Close()
	assert.NoError(t, err)
}

func TestUserCache(t *testing.T) {
	_, rc := setupMiniredis(t)
	ctx := context.Background()

	uc := NewUserCache(rc, 10*time.Minute)
	assert.NotNil(t, uc)

	type user struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
	}

	u := user{ID: 42, Email: "test@example.com"}

	// Set and get by ID
	err := uc.SetUser(ctx, 42, u)
	require.NoError(t, err)

	var got user
	found, err := uc.GetUser(ctx, 42, &got)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, u, got)

	// Invalidate
	err = uc.InvalidateUser(ctx, 42)
	require.NoError(t, err)

	found, err = uc.GetUser(ctx, 42, &got)
	require.NoError(t, err)
	assert.False(t, found)

	// Set and get by email
	err = uc.SetUserByEmail(ctx, "test@example.com", u)
	require.NoError(t, err)

	var gotByEmail user
	found, err = uc.GetUserByEmail(ctx, "test@example.com", &gotByEmail)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, u, gotByEmail)
}

func TestTokenBlacklist(t *testing.T) {
	_, rc := setupMiniredis(t)
	ctx := context.Background()

	tb := NewTokenBlacklist(rc)
	assert.NotNil(t, tb)

	// Token not blacklisted
	blacklisted, err := tb.IsBlacklisted(ctx, "token-123")
	require.NoError(t, err)
	assert.False(t, blacklisted)

	// Add to blacklist with future expiry
	err = tb.Add(ctx, "token-123", time.Now().Add(10*time.Minute))
	require.NoError(t, err)

	blacklisted, err = tb.IsBlacklisted(ctx, "token-123")
	require.NoError(t, err)
	assert.True(t, blacklisted)

	// Already-expired token should be a no-op
	err = tb.Add(ctx, "token-expired", time.Now().Add(-1*time.Minute))
	require.NoError(t, err)

	blacklisted, err = tb.IsBlacklisted(ctx, "token-expired")
	require.NoError(t, err)
	assert.False(t, blacklisted)
}

func TestNewRedisCacheWithTracing_NoTracing(t *testing.T) {
	mr := miniredis.RunT(t)
	rc, err := NewRedisCacheWithTracing(mr.Addr(), "", 0, false)
	require.NoError(t, err)
	assert.NotNil(t, rc)
	rc.Close()
}
