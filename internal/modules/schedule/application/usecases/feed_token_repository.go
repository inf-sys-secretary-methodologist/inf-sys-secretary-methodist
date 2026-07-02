package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// CalendarFeedTokenRepository persists per-user calendar feed tokens. A user
// has at most one token (user_id is unique); Save creates or rotates it.
type CalendarFeedTokenRepository interface {
	// Save creates the user's token or replaces it if one already exists.
	Save(ctx context.Context, token *entities.CalendarFeedToken) error
	// GetByUserID returns the user's token, or ErrCalendarFeedTokenNotFound.
	GetByUserID(ctx context.Context, userID int64) (*entities.CalendarFeedToken, error)
	// GetByToken resolves a token value to its owner, or ErrCalendarFeedTokenNotFound.
	GetByToken(ctx context.Context, token string) (*entities.CalendarFeedToken, error)
	// DeleteByUserID removes the user's token, disabling their feed. Deleting a
	// non-existent token is not an error.
	DeleteByUserID(ctx context.Context, userID int64) error
}
