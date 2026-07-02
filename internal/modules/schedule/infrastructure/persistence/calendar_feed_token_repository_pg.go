package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

// Save creates or rotates the user's feed token. user_id is unique, so an
// existing token is replaced (upsert), which invalidates the old feed URL.
func (r *CalendarFeedTokenRepositoryPG) Save(ctx context.Context, token *entities.CalendarFeedToken) error {
	const query = `
		INSERT INTO calendar_feed_tokens (user_id, token, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE
			SET token = EXCLUDED.token, created_at = EXCLUDED.created_at
		RETURNING id`

	if err := r.db.QueryRowContext(ctx, query, token.UserID, token.Token, token.CreatedAt).Scan(&token.ID); err != nil {
		return fmt.Errorf("failed to save calendar feed token: %w", err)
	}
	return nil
}

// GetByUserID returns the user's feed token, or ErrCalendarFeedTokenNotFound.
func (r *CalendarFeedTokenRepositoryPG) GetByUserID(ctx context.Context, userID int64) (*entities.CalendarFeedToken, error) {
	const query = `SELECT id, user_id, token, created_at FROM calendar_feed_tokens WHERE user_id = $1`
	return scanFeedToken(r.db.QueryRowContext(ctx, query, userID))
}

// GetByToken resolves a token value to its owner, or ErrCalendarFeedTokenNotFound.
func (r *CalendarFeedTokenRepositoryPG) GetByToken(ctx context.Context, token string) (*entities.CalendarFeedToken, error) {
	const query = `SELECT id, user_id, token, created_at FROM calendar_feed_tokens WHERE token = $1`
	return scanFeedToken(r.db.QueryRowContext(ctx, query, token))
}

// DeleteByUserID removes the user's feed token; deleting a missing token is a no-op.
func (r *CalendarFeedTokenRepositoryPG) DeleteByUserID(ctx context.Context, userID int64) error {
	const query = `DELETE FROM calendar_feed_tokens WHERE user_id = $1`
	if _, err := r.db.ExecContext(ctx, query, userID); err != nil {
		return fmt.Errorf("failed to delete calendar feed token: %w", err)
	}
	return nil
}

// scanFeedToken maps a single row into a CalendarFeedToken, translating a
// missing row into the domain ErrCalendarFeedTokenNotFound sentinel.
func scanFeedToken(row *sql.Row) (*entities.CalendarFeedToken, error) {
	t := &entities.CalendarFeedToken{}
	err := row.Scan(&t.ID, &t.UserID, &t.Token, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, entities.ErrCalendarFeedTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar feed token: %w", err)
	}
	return t, nil
}
