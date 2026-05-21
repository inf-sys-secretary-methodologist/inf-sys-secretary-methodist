package usecases

import (
	"context"
	"time"
)

// RevokedTokenRepository stores JTIs of access tokens that have been
// invalidated by an explicit logout. Implementations are expected to
// auto-expire entries after their TTL — a typical Redis SET … EX <ttl>.
//
// This is intentionally narrow: revocation is an off-band signal, not a
// general session store. Refresh-token rotation (v0.159.0 ADR-2) reuses
// the same backing store via the RevokeIfAbsent / RevokeAllForUser
// methods below so a stolen refresh JTI cannot be replayed.
type RevokedTokenRepository interface {
	// Revoke records the JTI as revoked with the given TTL. After ttl
	// elapses the entry must vanish on its own.
	Revoke(ctx context.Context, jti string, ttl time.Duration) error

	// IsRevoked reports whether the JTI is currently in the revocation set.
	IsRevoked(ctx context.Context, jti string) (bool, error)

	// RevokeIfAbsent atomically marks the JTI as revoked ONLY if it
	// is not already in the revocation set. Returns claimed=true when
	// THIS call performed the revocation, claimed=false when the JTI
	// was already revoked. Implementations MUST use a compare-and-set
	// primitive (Redis SET NX) so two concurrent callers cannot both
	// observe "absent" and both win — closes the refresh-rotation
	// race that v0.159.0 ADR-2 would otherwise have. Issue #279.
	RevokeIfAbsent(ctx context.Context, jti string, ttl time.Duration) (claimed bool, err error)

	// RevokeAllForUser triggers RFC 6749 §10.4 token-family revocation:
	// when refresh-token reuse is detected, every still-valid token
	// issued for the user must be invalidated, forcing a fresh login.
	// Implementations record a per-user revocation epoch; downstream
	// IsRevokedForUser checks the epoch against the token's iat to
	// decide validity. Issue #279.
	RevokeAllForUser(ctx context.Context, userID int64, issuedAtUnix int64, ttl time.Duration) error

	// IsRevokedForUser reports whether a token issued at issuedAtUnix
	// for userID falls under an outstanding user-level revocation
	// (RevokeAllForUser entry whose epoch >= issuedAtUnix). Used by
	// RefreshToken / access-token validators to honor cascade revokes.
	IsRevokedForUser(ctx context.Context, userID int64, issuedAtUnix int64) (bool, error)
}
