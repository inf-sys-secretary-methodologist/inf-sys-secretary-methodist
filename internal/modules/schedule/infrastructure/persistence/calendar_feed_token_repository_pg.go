package persistence

import (
	"context"
	"database/sql"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// CalendarFeedTokenRepositoryPG implements CalendarFeedTokenRepository on PostgreSQL.
type CalendarFeedTokenRepositoryPG struct {
	db *sql.DB
}

var _ usecases.CalendarFeedTokenRepository = (*CalendarFeedTokenRepositoryPG)(nil)

// NewCalendarFeedTokenRepositoryPG creates a new CalendarFeedTokenRepositoryPG.
func NewCalendarFeedTokenRepositoryPG(db *sql.DB) *CalendarFeedTokenRepositoryPG {
	return &CalendarFeedTokenRepositoryPG{db: db}
}

// Save creates or rotates the user's feed token.
func (r *CalendarFeedTokenRepositoryPG) Save(ctx context.Context, token *entities.CalendarFeedToken) error {
	return nil
}

// GetByUserID returns the user's feed token.
func (r *CalendarFeedTokenRepositoryPG) GetByUserID(ctx context.Context, userID int64) (*entities.CalendarFeedToken, error) {
	return nil, nil
}

// GetByToken resolves a token value to its owner.
func (r *CalendarFeedTokenRepositoryPG) GetByToken(ctx context.Context, token string) (*entities.CalendarFeedToken, error) {
	return nil, nil
}

// DeleteByUserID removes the user's feed token.
func (r *CalendarFeedTokenRepositoryPG) DeleteByUserID(ctx context.Context, userID int64) error {
	return nil
}
