// Package persistence contains PostgreSQL implementations of repositories.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/repositories"
)

// PreferencesRepositoryPG is PostgreSQL implementation of PreferencesRepository
type PreferencesRepositoryPG struct {
	db *sql.DB
}

// NewPreferencesRepositoryPG creates a new PostgreSQL preferences repository
func NewPreferencesRepositoryPG(db *sql.DB) repositories.PreferencesRepository {
	return &PreferencesRepositoryPG{db: db}
}

// Create creates new notification preferences
func (r *PreferencesRepositoryPG) Create(ctx context.Context, preferences *entities.UserNotificationPreferences) error {
	query := `
		INSERT INTO notification_preferences (
			user_id, email_enabled, push_enabled, in_app_enabled,
			telegram_enabled, slack_enabled, quiet_hours_enabled,
			quiet_hours_start, quiet_hours_end, timezone,
			digest_enabled, digest_frequency, digest_time,
			type_preferences, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id`

	typePrefsJSON, err := json.Marshal(preferences.TypePreferences)
	if err != nil {
		return fmt.Errorf("failed to marshal type preferences: %w", err)
	}

	err = r.db.QueryRowContext(ctx, query,
		preferences.UserID,
		preferences.EmailEnabled,
		preferences.PushEnabled,
		preferences.InAppEnabled,
		preferences.TelegramEnabled,
		preferences.SlackEnabled,
		preferences.QuietHoursEnabled,
		preferences.QuietHoursStart,
		preferences.QuietHoursEnd,
		preferences.Timezone,
		preferences.DigestEnabled,
		preferences.DigestFrequency,
		preferences.DigestTime,
		typePrefsJSON,
		preferences.CreatedAt,
		preferences.UpdatedAt,
	).Scan(&preferences.ID)

	if err != nil {
		return fmt.Errorf("failed to create preferences: %w", err)
	}

	return nil
}

// Update updates existing notification preferences
func (r *PreferencesRepositoryPG) Update(ctx context.Context, preferences *entities.UserNotificationPreferences) error {
	query := `
		UPDATE notification_preferences SET
			email_enabled = $2, push_enabled = $3, in_app_enabled = $4,
			telegram_enabled = $5, slack_enabled = $6, quiet_hours_enabled = $7,
			quiet_hours_start = $8, quiet_hours_end = $9, timezone = $10,
			digest_enabled = $11, digest_frequency = $12, digest_time = $13,
			type_preferences = $14, updated_at = NOW()
		WHERE user_id = $1`

	typePrefsJSON, err := json.Marshal(preferences.TypePreferences)
	if err != nil {
		return fmt.Errorf("failed to marshal type preferences: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		preferences.UserID,
		preferences.EmailEnabled,
		preferences.PushEnabled,
		preferences.InAppEnabled,
		preferences.TelegramEnabled,
		preferences.SlackEnabled,
		preferences.QuietHoursEnabled,
		preferences.QuietHoursStart,
		preferences.QuietHoursEnd,
		preferences.Timezone,
		preferences.DigestEnabled,
		preferences.DigestFrequency,
		preferences.DigestTime,
		typePrefsJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to update preferences: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("preferences not found")
	}

	return nil
}

// Delete deletes notification preferences for a user
func (r *PreferencesRepositoryPG) Delete(ctx context.Context, userID int64) error {
	query := `DELETE FROM notification_preferences WHERE user_id = $1`
	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete preferences: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("preferences not found")
	}

	return nil
}

// GetByUserID retrieves notification preferences for a user
func (r *PreferencesRepositoryPG) GetByUserID(ctx context.Context, userID int64) (*entities.UserNotificationPreferences, error) {
	query := `
		SELECT id, user_id, email_enabled, push_enabled, in_app_enabled,
			   telegram_enabled, slack_enabled, quiet_hours_enabled,
			   quiet_hours_start, quiet_hours_end, timezone,
			   digest_enabled, digest_frequency, digest_time,
			   type_preferences, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1`

	preferences := &entities.UserNotificationPreferences{}
	var typePrefsJSON []byte

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&preferences.ID,
		&preferences.UserID,
		&preferences.EmailEnabled,
		&preferences.PushEnabled,
		&preferences.InAppEnabled,
		&preferences.TelegramEnabled,
		&preferences.SlackEnabled,
		&preferences.QuietHoursEnabled,
		&preferences.QuietHoursStart,
		&preferences.QuietHoursEnd,
		&preferences.Timezone,
		&preferences.DigestEnabled,
		&preferences.DigestFrequency,
		&preferences.DigestTime,
		&typePrefsJSON,
		&preferences.CreatedAt,
		&preferences.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get preferences: %w", err)
	}

	if len(typePrefsJSON) > 0 {
		preferences.TypePreferences = make(map[entities.NotificationType]entities.TypePreference)
		_ = json.Unmarshal(typePrefsJSON, &preferences.TypePreferences)
	}

	return preferences, nil
}

// GetOrCreate retrieves preferences or creates default if not exists
func (r *PreferencesRepositoryPG) GetOrCreate(ctx context.Context, userID int64) (*entities.UserNotificationPreferences, error) {
	preferences, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if preferences != nil {
		return preferences, nil
	}

	// Create default preferences
	preferences = entities.NewUserNotificationPreferences(userID)
	if err := r.Create(ctx, preferences); err != nil {
		return nil, err
	}

	return preferences, nil
}

// UpdateChannelEnabled updates a specific channel's enabled status
func (r *PreferencesRepositoryPG) UpdateChannelEnabled(ctx context.Context, userID int64, channel entities.NotificationChannel, enabled bool) error {
	var columnName string
	switch channel {
	case entities.ChannelEmail:
		columnName = "email_enabled"
	case entities.ChannelPush:
		columnName = "push_enabled"
	case entities.ChannelInApp:
		columnName = "in_app_enabled"
	case entities.ChannelTelegram:
		columnName = "telegram_enabled"
	case entities.ChannelSlack:
		columnName = "slack_enabled"
	default:
		return fmt.Errorf("unknown channel: %s", channel)
	}

	query := fmt.Sprintf(`UPDATE notification_preferences SET %s = $2, updated_at = NOW() WHERE user_id = $1`, columnName) // #nosec G201 -- column name from switch/case, not user input
	result, err := r.db.ExecContext(ctx, query, userID, enabled)
	if err != nil {
		return fmt.Errorf("failed to update channel enabled: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Create preferences with the channel enabled/disabled
		preferences := entities.NewUserNotificationPreferences(userID)
		switch channel {
		case entities.ChannelEmail:
			preferences.EmailEnabled = enabled
		case entities.ChannelPush:
			preferences.PushEnabled = enabled
		case entities.ChannelInApp:
			preferences.InAppEnabled = enabled
		case entities.ChannelTelegram:
			preferences.TelegramEnabled = enabled
		case entities.ChannelSlack:
			preferences.SlackEnabled = enabled
		}
		return r.Create(ctx, preferences)
	}

	return nil
}

// UpdateQuietHours updates quiet hours settings
func (r *PreferencesRepositoryPG) UpdateQuietHours(ctx context.Context, userID int64, enabled bool, start, end, timezone string) error {
	query := `
		UPDATE notification_preferences SET
			quiet_hours_enabled = $2,
			quiet_hours_start = $3,
			quiet_hours_end = $4,
			timezone = $5,
			updated_at = NOW()
		WHERE user_id = $1`

	result, err := r.db.ExecContext(ctx, query, userID, enabled, start, end, timezone)
	if err != nil {
		return fmt.Errorf("failed to update quiet hours: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Create preferences with quiet hours
		preferences := entities.NewUserNotificationPreferences(userID)
		preferences.QuietHoursEnabled = enabled
		preferences.QuietHoursStart = start
		preferences.QuietHoursEnd = end
		preferences.Timezone = timezone
		return r.Create(ctx, preferences)
	}

	return nil
}
