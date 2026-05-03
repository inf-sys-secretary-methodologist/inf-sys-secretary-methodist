package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
)

var (
	// ErrTokenAlreadyExpired is returned when caller hands a token whose
	// exp claim is already in the past — there is nothing left to revoke.
	ErrTokenAlreadyExpired = errors.New("token already expired")

	// ErrTokenMissingJTI is returned when the token has no jti claim. JTI
	// is the unique handle a revocation entry hangs on; without it we
	// cannot blacklist a single token without blacklisting the user.
	ErrTokenMissingJTI = errors.New("token missing jti claim")
)

// LogoutUseCase invalidates an access token by adding its JTI to the
// revoked-token set with TTL bounded by the token's own remaining lifetime.
// Refresh tokens are not handled here — the client is expected to discard
// them; revocation of refresh tokens belongs to SessionRepository (separate
// concern).
type LogoutUseCase struct {
	revokedRepo repositories.RevokedTokenRepository
	jwtSecret   []byte
}

// NewLogoutUseCase wires a LogoutUseCase. jwtSecret must match the secret
// used to sign access tokens (see AuthUseCase.generateTokens).
func NewLogoutUseCase(revokedRepo repositories.RevokedTokenRepository, jwtSecret []byte) *LogoutUseCase {
	return &LogoutUseCase{
		revokedRepo: revokedRepo,
		jwtSecret:   jwtSecret,
	}
}

// Logout parses, validates the access token and stores its JTI in the
// revoked-token repository with TTL = (exp − now). Returns an error for
// malformed/expired tokens or tokens without a JTI claim.
func (u *LogoutUseCase) Logout(ctx context.Context, accessToken string) error {
	parsed, err := jwt.Parse(accessToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return u.jwtSecret, nil
	})
	if err != nil || !parsed.Valid {
		return fmt.Errorf("invalid access token: %w", ErrInvalidToken)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("unexpected claims shape: %w", ErrInvalidToken)
	}

	jti, _ := claims["jti"].(string)
	if jti == "" {
		return ErrTokenMissingJTI
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("missing exp claim: %w", ErrInvalidToken)
	}
	expiresAt := time.Unix(int64(expFloat), 0)
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return ErrTokenAlreadyExpired
	}

	if err := u.revokedRepo.Revoke(ctx, jti, ttl); err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}
	return nil
}
