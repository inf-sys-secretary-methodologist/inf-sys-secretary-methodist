// Package persistence contains PostgreSQL implementations of repositories.
package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/repositories"
)

// WebPushRepositoryPG is PostgreSQL implementation of WebPushRepository
type WebPushRepositoryPG struct {
	db *sql.DB
}

// NewWebPushRepositoryPG creates a new PostgreSQL Web Push repository
func NewWebPushRepositoryPG(db *sql.DB) repositories.WebPushRepository {
	return &WebPushRepositoryPG{db: db}
}

// Create creates a new web push subscription
func (r *WebPushRepositoryPG) Create(ctx context.Context, sub *entities.WebPushSubscription) error {
	query := `
		INSERT INTO webpush_subscriptions (user_id, endpoint, p256dh_key, auth_key, user_agent, device_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (endpoint) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			p256dh_key = EXCLUDED.p256dh_key,
			auth_key = EXCLUDED.auth_key,
			user_agent = EXCLUDED.user_agent,
			device_name = EXCLUDED.device_name,
			is_active = true,
			updated_at = NOW()
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		sub.UserID,
		sub.Endpoint,
		sub.P256dhKey,
		sub.AuthKey,
		nullString(sub.UserAgent),
		nullString(sub.DeviceName),
		sub.IsActive,
		sub.CreatedAt,
		sub.UpdatedAt,
	).Scan(&sub.ID)

	if err != nil {
		return fmt.Errorf("failed to create web push subscription: %w", err)
	}

	return nil
}

// GetByID retrieves a subscription by ID
func (r *WebPushRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.WebPushSubscription, error) {
	query := `
		SELECT id, user_id, endpoint, p256dh_key, auth_key, user_agent, device_name, is_active, last_used_at, created_at, updated_at
		FROM webpush_subscriptions
		WHERE id = $1`

	sub := &entities.WebPushSubscription{}
	var userAgent, deviceName sql.NullString
	var lastUsedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.Endpoint,
		&sub.P256dhKey,
		&sub.AuthKey,
		&userAgent,
		&deviceName,
		&sub.IsActive,
		&lastUsedAt,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get web push subscription: %w", err)
	}

	sub.UserAgent = userAgent.String
	sub.DeviceName = deviceName.String
	if lastUsedAt.Valid {
		sub.LastUsedAt = &lastUsedAt.Time
	}

	return sub, nil
}

// GetByEndpoint retrieves a subscription by endpoint
func (r *WebPushRepositoryPG) GetByEndpoint(ctx context.Context, endpoint string) (*entities.WebPushSubscription, error) {
	query := `
		SELECT id, user_id, endpoint, p256dh_key, auth_key, user_agent, device_name, is_active, last_used_at, created_at, updated_at
		FROM webpush_subscriptions
		WHERE endpoint = $1`

	sub := &entities.WebPushSubscription{}
	var userAgent, deviceName sql.NullString
	var lastUsedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, endpoint).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.Endpoint,
		&sub.P256dhKey,
		&sub.AuthKey,
		&userAgent,
		&deviceName,
		&sub.IsActive,
		&lastUsedAt,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get web push subscription by endpoint: %w", err)
	}

	sub.UserAgent = userAgent.String
	sub.DeviceName = deviceName.String
	if lastUsedAt.Valid {
		sub.LastUsedAt = &lastUsedAt.Time
	}

	return sub, nil
}

// GetByUserID retrieves all subscriptions for a user
func (r *WebPushRepositoryPG) GetByUserID(ctx context.Context, userID int64) ([]*entities.WebPushSubscription, error) {
	query := `
		SELECT id, user_id, endpoint, p256dh_key, auth_key, user_agent, device_name, is_active, last_used_at, created_at, updated_at
		FROM webpush_subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get web push subscriptions: %w", err)
	}
	defer rows.Close()

	return scanSubscriptions(rows)
}

// GetActiveByUserID retrieves all active subscriptions for a user
func (r *WebPushRepositoryPG) GetActiveByUserID(ctx context.Context, userID int64) ([]*entities.WebPushSubscription, error) {
	query := `
		SELECT id, user_id, endpoint, p256dh_key, auth_key, user_agent, device_name, is_active, last_used_at, created_at, updated_at
		FROM webpush_subscriptions
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active web push subscriptions: %w", err)
	}
	defer rows.Close()

	return scanSubscriptions(rows)
}

// Update updates an existing subscription
func (r *WebPushRepositoryPG) Update(ctx context.Context, sub *entities.WebPushSubscription) error {
	query := `
		UPDATE webpush_subscriptions SET
			p256dh_key = $2,
			auth_key = $3,
			user_agent = $4,
			device_name = $5,
			is_active = $6,
			updated_at = NOW()
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		sub.ID,
		sub.P256dhKey,
		sub.AuthKey,
		nullString(sub.UserAgent),
		nullString(sub.DeviceName),
		sub.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update web push subscription: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("web push subscription not found")
	}

	return nil
}

// UpdateLastUsed updates the last_used_at timestamp
func (r *WebPushRepositoryPG) UpdateLastUsed(ctx context.Context, id int64) error {
	query := `UPDATE webpush_subscriptions SET last_used_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}
	return nil
}

// Deactivate marks a subscription as inactive
func (r *WebPushRepositoryPG) Deactivate(ctx context.Context, id int64) error {
	query := `UPDATE webpush_subscriptions SET is_active = false, updated_at = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate subscription: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("web push subscription not found")
	}

	return nil
}

// Delete deletes a subscription by ID
func (r *WebPushRepositoryPG) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM webpush_subscriptions WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete web push subscription: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("web push subscription not found")
	}

	return nil
}

// DeleteByEndpoint deletes a subscription by endpoint
func (r *WebPushRepositoryPG) DeleteByEndpoint(ctx context.Context, endpoint string) error {
	query := `DELETE FROM webpush_subscriptions WHERE endpoint = $1`
	_, err := r.db.ExecContext(ctx, query, endpoint)
	if err != nil {
		return fmt.Errorf("failed to delete web push subscription: %w", err)
	}
	return nil
}

// DeleteByUserID deletes all subscriptions for a user
func (r *WebPushRepositoryPG) DeleteByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM webpush_subscriptions WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user's web push subscriptions: %w", err)
	}
	return nil
}

// CountByUserID counts the number of active subscriptions for a user
func (r *WebPushRepositoryPG) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	query := `SELECT COUNT(*) FROM webpush_subscriptions WHERE user_id = $1 AND is_active = true`
	var count int64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count web push subscriptions: %w", err)
	}
	return count, nil
}

// Helper functions

func scanSubscriptions(rows *sql.Rows) ([]*entities.WebPushSubscription, error) {
	var subscriptions []*entities.WebPushSubscription

	for rows.Next() {
		sub := &entities.WebPushSubscription{}
		var userAgent, deviceName sql.NullString
		var lastUsedAt sql.NullTime

		err := rows.Scan(
			&sub.ID,
			&sub.UserID,
			&sub.Endpoint,
			&sub.P256dhKey,
			&sub.AuthKey,
			&userAgent,
			&deviceName,
			&sub.IsActive,
			&lastUsedAt,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}

		sub.UserAgent = userAgent.String
		sub.DeviceName = deviceName.String
		if lastUsedAt.Valid {
			sub.LastUsedAt = &lastUsedAt.Time
		}

		subscriptions = append(subscriptions, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subscriptions: %w", err)
	}

	return subscriptions, nil
}

