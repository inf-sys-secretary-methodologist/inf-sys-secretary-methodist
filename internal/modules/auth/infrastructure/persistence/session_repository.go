package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/database"
)

// SessionRepositoryPG implements PostgreSQL session repository
type SessionRepositoryPG struct {
	db *sql.DB
}

// NewSessionRepositoryPG creates a new PostgreSQL session repository
func NewSessionRepositoryPG(db *sql.DB) repositories.SessionRepository {
	return &SessionRepositoryPG{db: db}
}

// Create creates a new session
func (r *SessionRepositoryPG) Create(ctx context.Context, session *entities.Session) error {
	query := `
		INSERT INTO sessions (user_id, refresh_token, user_agent, ip_address, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		session.UserID,
		session.RefreshToken,
		session.UserAgent,
		session.IPAddress,
		session.ExpiresAt,
		session.CreatedAt,
		session.UpdatedAt,
	).Scan(&session.ID)

	if err != nil {
		return database.MapPostgresError(err)
	}
	return nil
}

// GetByRefreshToken retrieves a session by refresh token
func (r *SessionRepositoryPG) GetByRefreshToken(ctx context.Context, refreshToken string) (*entities.Session, error) {
	session := &entities.Session{}
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at, updated_at
		FROM sessions
		WHERE refresh_token = $1
	`
	err := r.db.QueryRowContext(ctx, query, refreshToken).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		return nil, database.MapPostgresError(err)
	}

	return session, nil
}

// Delete deletes a session by refresh token
func (r *SessionRepositoryPG) Delete(ctx context.Context, refreshToken string) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE refresh_token = $1`,
		refreshToken,
	)
	if err != nil {
		return database.MapPostgresError(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return database.MapPostgresError(err)
	}

	if rows == 0 {
		return database.MapPostgresError(sql.ErrNoRows)
	}

	return nil
}

// DeleteByUserID deletes all sessions for a user
func (r *SessionRepositoryPG) DeleteByUserID(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return database.MapPostgresError(err)
	}

	return nil
}

// DeleteExpired deletes all expired sessions
func (r *SessionRepositoryPG) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE expires_at < $1`,
		time.Now(),
	)
	if err != nil {
		return database.MapPostgresError(err)
	}

	return nil
}

// GetActiveByUserID retrieves all active sessions for a user
func (r *SessionRepositoryPG) GetActiveByUserID(ctx context.Context, userID int64) ([]*entities.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at, updated_at
		FROM sessions
		WHERE user_id = $1 AND expires_at > $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID, time.Now())
	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	defer rows.Close()

	sessions := []*entities.Session{}
	for rows.Next() {
		session := &entities.Session{}
		if err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.RefreshToken,
			&session.UserAgent,
			&session.IPAddress,
			&session.ExpiresAt,
			&session.CreatedAt,
			&session.UpdatedAt,
		); err != nil {
			return nil, database.MapPostgresError(err)
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, database.MapPostgresError(err)
	}

	return sessions, nil
}
