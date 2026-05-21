package usecases

import (
	"context"
	"time"
)

// PasswordResetTokenRepository persists short-lived single-use tokens
// that authenticate the holder to set a new password for a specific user.
//
// Implementations are expected to auto-expire entries after their TTL
// (typical Redis SET … EX <ttl>). The repository does not judge token
// validity beyond presence/absence; the usecase treats an absent entry
// as an invalid token.
type PasswordResetTokenRepository interface {
	// Store records that token grants reset permission for userID, valid
	// for ttl. After ttl elapses the entry must vanish on its own.
	Store(ctx context.Context, token string, userID int64, ttl time.Duration) error

	// LookupUser returns the userID associated with a stored token.
	// Returns domain.ErrPasswordResetTokenNotFound when the token is absent
	// or expired.
	LookupUser(ctx context.Context, token string) (int64, error)

	// Delete removes the token unconditionally — used after a successful
	// password reset to enforce single-use.
	Delete(ctx context.Context, token string) error

	// LookupUserAndConsume atomically returns the userID associated with
	// the token AND removes the token in one round-trip (Redis GETDEL).
	// After a successful call the token is gone irrespective of what the
	// caller does next — replay-by-Save-failure is structurally
	// impossible. Returns domain.ErrPasswordResetTokenNotFound when the
	// token is absent or already consumed. Issue #279 ADR-6.
	LookupUserAndConsume(ctx context.Context, token string) (int64, error)
}
