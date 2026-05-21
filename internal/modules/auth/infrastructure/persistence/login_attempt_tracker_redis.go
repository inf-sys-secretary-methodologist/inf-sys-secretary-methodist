package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
)

// loginAttemptKeyPrefix isolates the failure-counter keyspace from
// other Redis users. Mirrors the pwreset: / revoked: conventions.
const loginAttemptKeyPrefix = "login_fail:"

// RedisLoginAttemptTracker is a Redis-backed counter that fronts the
// per-account brute-force lockout. Failures accumulate under a short
// TTL; once the threshold is crossed IsLocked returns true until the
// TTL expires. Successful logins call Reset to clear the counter.
//
// The identifier is normalised to lowercase + trimmed so casing
// differences (Alice@x vs alice@x) share the same bucket.
type RedisLoginAttemptTracker struct {
	client    *redis.Client
	threshold int
	window    time.Duration
}

// Ensure the type satisfies the use-case interface at compile time so
// drift between the two surfaces fails the build, not silently at runtime.
var _ usecases.LoginAttemptTracker = (*RedisLoginAttemptTracker)(nil)

// NewRedisLoginAttemptTracker constructs the tracker with the lockout
// threshold (failures within the window) and the window duration
// itself (counter TTL). Callers own the redis client lifecycle.
func NewRedisLoginAttemptTracker(client *redis.Client, threshold int, window time.Duration) *RedisLoginAttemptTracker {
	return &RedisLoginAttemptTracker{
		client:    client,
		threshold: threshold,
		window:    window,
	}
}

func (t *RedisLoginAttemptTracker) keyFor(identifier string) string {
	return loginAttemptKeyPrefix + strings.ToLower(strings.TrimSpace(identifier))
}

// RegisterFailure atomically increments the counter and refreshes the
// TTL on every write. The Lua script keeps the INCR + EXPIRE atomic so
// a race between two concurrent failures cannot leave the entry
// without a TTL.
func (t *RedisLoginAttemptTracker) RegisterFailure(ctx context.Context, identifier string) (int, error) {
	key := t.keyFor(identifier)
	luaScript := `
		local key = KEYS[1]
		local window = tonumber(ARGV[1])
		local n = redis.call("INCR", key)
		redis.call("EXPIRE", key, window)
		return n
	`
	res, err := t.client.Eval(ctx, luaScript, []string{key}, int64(t.window.Seconds())).Result()
	if err != nil {
		return 0, fmt.Errorf("redis EVAL %s: %w", key, err)
	}
	n, ok := res.(int64)
	if !ok {
		return 0, fmt.Errorf("redis EVAL %s: unexpected result type %T", key, res)
	}
	return int(n), nil
}

// IsLocked reports whether the current counter is at or above the
// lockout threshold. A missing key (TTL expired) reads as "not locked"
// so abandoned counters do not permanently keep a user out.
func (t *RedisLoginAttemptTracker) IsLocked(ctx context.Context, identifier string) (bool, error) {
	key := t.keyFor(identifier)
	raw, err := t.client.Get(ctx, key).Int()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("redis GET %s: %w", key, err)
	}
	return raw >= t.threshold, nil
}

// Reset clears the counter. Deleting a missing key is not an error.
func (t *RedisLoginAttemptTracker) Reset(ctx context.Context, identifier string) error {
	key := t.keyFor(identifier)
	if err := t.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis DEL %s: %w", key, err)
	}
	return nil
}
