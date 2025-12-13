// Package persistence contains PostgreSQL implementations of repositories.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/repositories"
)

// NotificationRepositoryPG is PostgreSQL implementation of NotificationRepository
type NotificationRepositoryPG struct {
	db *sql.DB
}

// NewNotificationRepositoryPG creates a new PostgreSQL notification repository
func NewNotificationRepositoryPG(db *sql.DB) repositories.NotificationRepository {
	return &NotificationRepositoryPG{db: db}
}

// Create creates a new notification
func (r *NotificationRepositoryPG) Create(ctx context.Context, notification *entities.Notification) error {
	query := `
		INSERT INTO notifications (
			user_id, type, priority, title, message, link, image_url,
			is_read, read_at, expires_at, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`

	metadataJSON, err := json.Marshal(notification.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = r.db.QueryRowContext(ctx, query,
		notification.UserID,
		notification.Type,
		notification.Priority,
		notification.Title,
		notification.Message,
		nullString(notification.Link),
		nullString(notification.ImageURL),
		notification.IsRead,
		notification.ReadAt,
		notification.ExpiresAt,
		metadataJSON,
		notification.CreatedAt,
		notification.UpdatedAt,
	).Scan(&notification.ID)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// Update updates an existing notification
func (r *NotificationRepositoryPG) Update(ctx context.Context, notification *entities.Notification) error {
	query := `
		UPDATE notifications SET
			type = $2, priority = $3, title = $4, message = $5,
			link = $6, image_url = $7, is_read = $8, read_at = $9,
			expires_at = $10, metadata = $11, updated_at = NOW()
		WHERE id = $1`

	metadataJSON, err := json.Marshal(notification.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		notification.ID,
		notification.Type,
		notification.Priority,
		notification.Title,
		notification.Message,
		nullString(notification.Link),
		nullString(notification.ImageURL),
		notification.IsRead,
		notification.ReadAt,
		notification.ExpiresAt,
		metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// Delete deletes a notification by ID
func (r *NotificationRepositoryPG) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM notifications WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// GetByID retrieves a notification by ID
func (r *NotificationRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.Notification, error) {
	query := `
		SELECT id, user_id, type, priority, title, message, link, image_url,
			   is_read, read_at, expires_at, metadata, created_at, updated_at
		FROM notifications
		WHERE id = $1`

	notification := &entities.Notification{}
	var link, imageURL sql.NullString
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Priority,
		&notification.Title,
		&notification.Message,
		&link,
		&imageURL,
		&notification.IsRead,
		&notification.ReadAt,
		&notification.ExpiresAt,
		&metadataJSON,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	notification.Link = link.String
	notification.ImageURL = imageURL.String

	if len(metadataJSON) > 0 {
		_ = json.Unmarshal(metadataJSON, &notification.Metadata)
	}

	return notification, nil
}

// List retrieves notifications based on filter criteria
func (r *NotificationRepositoryPG) List(ctx context.Context, filter *entities.NotificationFilter) ([]*entities.Notification, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	if filter.UserID > 0 {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argNum))
		args = append(args, filter.UserID)
		argNum++
	}

	if filter.Type != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argNum))
		args = append(args, filter.Type)
		argNum++
	}

	if filter.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("priority = $%d", argNum))
		args = append(args, filter.Priority)
		argNum++
	}

	if filter.IsRead != nil {
		conditions = append(conditions, fmt.Sprintf("is_read = $%d", argNum))
		args = append(args, *filter.IsRead)
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, type, priority, title, message, link, image_url,
			   is_read, read_at, expires_at, metadata, created_at, updated_at
		FROM notifications
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argNum, argNum+1)

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	args = append(args, limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	return r.scanNotifications(rows)
}

// GetByUserID retrieves notifications for a user
func (r *NotificationRepositoryPG) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*entities.Notification, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, user_id, type, priority, title, message, link, image_url,
			   is_read, read_at, expires_at, metadata, created_at, updated_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications by user: %w", err)
	}
	defer rows.Close()

	return r.scanNotifications(rows)
}

// GetUnreadByUserID retrieves unread notifications for a user
func (r *NotificationRepositoryPG) GetUnreadByUserID(ctx context.Context, userID int64) ([]*entities.Notification, error) {
	query := `
		SELECT id, user_id, type, priority, title, message, link, image_url,
			   is_read, read_at, expires_at, metadata, created_at, updated_at
		FROM notifications
		WHERE user_id = $1 AND is_read = false
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread notifications: %w", err)
	}
	defer rows.Close()

	return r.scanNotifications(rows)
}

// MarkAsRead marks a notification as read
func (r *NotificationRepositoryPG) MarkAsRead(ctx context.Context, id int64) error {
	query := `UPDATE notifications SET is_read = true, read_at = NOW(), updated_at = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (r *NotificationRepositoryPG) MarkAllAsRead(ctx context.Context, userID int64) error {
	query := `UPDATE notifications SET is_read = true, read_at = NOW(), updated_at = NOW() WHERE user_id = $1 AND is_read = false`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// DeleteByUserID deletes all notifications for a user
func (r *NotificationRepositoryPG) DeleteByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM notifications WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user notifications: %w", err)
	}

	return nil
}

// DeleteExpired deletes expired notifications
func (r *NotificationRepositoryPG) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM notifications WHERE expires_at IS NOT NULL AND expires_at < NOW()`
	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired notifications: %w", err)
	}

	return result.RowsAffected()
}

// GetUnreadCount returns the count of unread notifications for a user
func (r *NotificationRepositoryPG) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`
	var count int64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// GetStats returns notification statistics for a user
func (r *NotificationRepositoryPG) GetStats(ctx context.Context, userID int64) (*entities.NotificationStats, error) {
	query := `
		SELECT
			COUNT(*) as total_count,
			COUNT(*) FILTER (WHERE is_read = false) as unread_count,
			COUNT(*) FILTER (WHERE created_at >= CURRENT_DATE) as today_count,
			COUNT(*) FILTER (WHERE priority = 'urgent' AND is_read = false) as urgent_count,
			COUNT(*) FILTER (WHERE expires_at IS NOT NULL AND expires_at < NOW()) as expired_count
		FROM notifications
		WHERE user_id = $1`

	stats := &entities.NotificationStats{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.TotalCount,
		&stats.UnreadCount,
		&stats.TodayCount,
		&stats.UrgentCount,
		&stats.ExpiredCount,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get notification stats: %w", err)
	}

	return stats, nil
}

// CreateBulk creates multiple notifications in a single transaction
func (r *NotificationRepositoryPG) CreateBulk(ctx context.Context, notifications []*entities.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO notifications (
			user_id, type, priority, title, message, link, image_url,
			is_read, expires_at, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, notification := range notifications {
		metadataJSON, err := json.Marshal(notification.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		err = stmt.QueryRowContext(ctx,
			notification.UserID,
			notification.Type,
			notification.Priority,
			notification.Title,
			notification.Message,
			nullString(notification.Link),
			nullString(notification.ImageURL),
			notification.IsRead,
			notification.ExpiresAt,
			metadataJSON,
			notification.CreatedAt,
			notification.UpdatedAt,
		).Scan(&notification.ID)

		if err != nil {
			return fmt.Errorf("failed to create notification: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// scanNotifications scans rows into notification slice
func (r *NotificationRepositoryPG) scanNotifications(rows *sql.Rows) ([]*entities.Notification, error) {
	var notifications []*entities.Notification

	for rows.Next() {
		notification := &entities.Notification{}
		var link, imageURL sql.NullString
		var metadataJSON []byte

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Priority,
			&notification.Title,
			&notification.Message,
			&link,
			&imageURL,
			&notification.IsRead,
			&notification.ReadAt,
			&notification.ExpiresAt,
			&metadataJSON,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		notification.Link = link.String
		notification.ImageURL = imageURL.String

		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &notification.Metadata)
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return notifications, nil
}

// nullString converts a string to sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
