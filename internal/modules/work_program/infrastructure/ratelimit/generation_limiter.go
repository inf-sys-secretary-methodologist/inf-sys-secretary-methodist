package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
)

const (
	keyPrefix     = "rate_limit:wp_generate:user:"
	defaultLimit  = int64(5)
	defaultWindow = time.Hour
)

// allowScript atomically increments the per-user counter and sets the
// window TTL on the first increment, returning the post-increment count.
// Doing both in one Lua call avoids a race where two concurrent requests
// both see count==1 and neither (or both) set the expiry.
const allowScript = `
local c = redis.call("INCR", KEYS[1])
if c == 1 then
	redis.call("EXPIRE", KEYS[1], ARGV[1])
end
return c
`

// GenerationLimiter is a Redis-backed per-user fixed-window limiter for
// LLM draft generation (default 5/hour/user per ADR-7). It implements
// usecases.GenerationRateLimiter.
type GenerationLimiter struct {
	redis  *redis.Client
	limit  int64
	window time.Duration
}

// compile-time check that the adapter satisfies the application port.
var _ usecases.GenerationRateLimiter = (*GenerationLimiter)(nil)

// NewGenerationLimiter wires the limiter. Non-positive limit/window fall
// back to the ADR-7 defaults (5 per hour).
func NewGenerationLimiter(client *redis.Client, limit int, window time.Duration) *GenerationLimiter {
	l := int64(limit)
	if l <= 0 {
		l = defaultLimit
	}
	if window <= 0 {
		window = defaultWindow
	}
	return &GenerationLimiter{redis: client, limit: l, window: window}
}

// Allow atomically counts the call against the user's fixed window and
// reports whether it is within budget. A Redis error is surfaced to the
// caller (the use case decides how to degrade).
func (g *GenerationLimiter) Allow(ctx context.Context, userID int64) (bool, error) {
	key := fmt.Sprintf("%s%d", keyPrefix, userID)
	res, err := g.redis.Eval(ctx, allowScript, []string{key}, int64(g.window.Seconds())).Result()
	if err != nil {
		return false, fmt.Errorf("work_program/ratelimit: eval: %w", err)
	}
	count, ok := res.(int64)
	if !ok {
		return false, fmt.Errorf("work_program/ratelimit: unexpected eval result type %T", res)
	}
	return count <= g.limit, nil
}
