package ratelimit

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
)

// GenerationLimiter is a Redis-backed per-user fixed-window limiter for
// LLM draft generation. STUB — behavior lands in the GREEN commit.
type GenerationLimiter struct{}

// compile-time check that the adapter satisfies the application port.
var _ usecases.GenerationRateLimiter = (*GenerationLimiter)(nil)

// NewGenerationLimiter is the STUB constructor.
func NewGenerationLimiter(client *redis.Client, limit int, window time.Duration) *GenerationLimiter {
	return &GenerationLimiter{}
}

// Allow is the STUB entrypoint — always returns not-implemented.
func (g *GenerationLimiter) Allow(ctx context.Context, userID int64) (bool, error) {
	return false, errors.New("work_program/ratelimit: Allow not implemented")
}
