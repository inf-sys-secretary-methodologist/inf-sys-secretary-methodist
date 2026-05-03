package repositories

import (
	"context"
	"errors"
	"time"
)

// ErrPasswordResetTokenNotFound is returned by LookupUser when the token
// is absent or has expired. Exposed as a domain sentinel so callers can
// distinguish "invalid/expired token" from a transport/storage failure
// via errors.Is, without parsing strings.
var ErrPasswordResetTokenNotFound = errors.New("password reset token not found")

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
	// Returns ErrPasswordResetTokenNotFound when the token is absent or
	// expired.
	LookupUser(ctx context.Context, token string) (int64, error)

	// Delete removes the token unconditionally — used after a successful
	// password reset to enforce single-use.
	Delete(ctx context.Context, token string) error
}
