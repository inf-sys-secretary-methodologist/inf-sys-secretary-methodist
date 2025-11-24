// Package repositories defines repository interfaces for the auth module.
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
)

// SessionRepository defines the interface for session persistence
type SessionRepository interface {
	// Create creates a new session
	Create(ctx context.Context, session *entities.Session) error

	// GetByRefreshToken retrieves a session by refresh token
	GetByRefreshToken(ctx context.Context, refreshToken string) (*entities.Session, error)

	// Delete deletes a session by refresh token
	Delete(ctx context.Context, refreshToken string) error

	// DeleteByUserID deletes all sessions for a user
	DeleteByUserID(ctx context.Context, userID int64) error

	// DeleteExpired deletes all expired sessions
	DeleteExpired(ctx context.Context) error

	// GetActiveByUserID retrieves all active sessions for a user
	GetActiveByUserID(ctx context.Context, userID int64) ([]*entities.Session, error)
}
