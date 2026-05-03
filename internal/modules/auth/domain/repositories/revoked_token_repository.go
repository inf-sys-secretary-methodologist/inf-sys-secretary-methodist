package repositories

import (
	"context"
	"time"
)

// RevokedTokenRepository stores JTIs of access tokens that have been
// invalidated by an explicit logout. Implementations are expected to
// auto-expire entries after their TTL — a typical Redis SET … EX <ttl>.
//
// This is intentionally narrow: revocation is an off-band signal, not a
// general session store. Refresh tokens are tracked separately via
// SessionRepository.
type RevokedTokenRepository interface {
	// Revoke records the JTI as revoked with the given TTL. After ttl
	// elapses the entry must vanish on its own.
	Revoke(ctx context.Context, jti string, ttl time.Duration) error

	// IsRevoked reports whether the JTI is currently in the revocation set.
	IsRevoked(ctx context.Context, jti string) (bool, error)
}
