// Package persistence contains PostgreSQL implementations of repositories.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/repositories"
)

// TelegramRepositoryPG is PostgreSQL implementation of TelegramRepository
type TelegramRepositoryPG struct {
	db *sql.DB
}

// NewTelegramRepositoryPG creates a new PostgreSQL Telegram repository
func NewTelegramRepositoryPG(db *sql.DB) repositories.TelegramRepository {
	return &TelegramRepositoryPG{db: db}
}

// CreateVerificationCode creates a new verification code
func (r *TelegramRepositoryPG) CreateVerificationCode(ctx context.Context, code *entities.TelegramVerificationCode) error {
	query := `
		INSERT INTO telegram_verification_codes (user_id, code, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		code.UserID,
		code.Code,
		code.ExpiresAt,
		code.CreatedAt,
	).Scan(&code.ID)

	if err != nil {
		return fmt.Errorf("failed to create verification code: %w", err)
	}

	return nil
}

// GetVerificationCodeByCode retrieves a verification code by its code value
func (r *TelegramRepositoryPG) GetVerificationCodeByCode(ctx context.Context, code string) (*entities.TelegramVerificationCode, error) {
	query := `
		SELECT id, user_id, code, expires_at, used_at, created_at
		FROM telegram_verification_codes
		WHERE code = $1`

	verificationCode := &entities.TelegramVerificationCode{}
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&verificationCode.ID,
		&verificationCode.UserID,
		&verificationCode.Code,
		&verificationCode.ExpiresAt,
		&verificationCode.UsedAt,
		&verificationCode.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get verification code: %w", err)
	}

	return verificationCode, nil
}

// GetActiveVerificationCodeByUserID retrieves the active verification code for a user
func (r *TelegramRepositoryPG) GetActiveVerificationCodeByUserID(ctx context.Context, userID int64) (*entities.TelegramVerificationCode, error) {
	query := `
		SELECT id, user_id, code, expires_at, used_at, created_at
		FROM telegram_verification_codes
		WHERE user_id = $1 AND used_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1`

	verificationCode := &entities.TelegramVerificationCode{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&verificationCode.ID,
		&verificationCode.UserID,
		&verificationCode.Code,
		&verificationCode.ExpiresAt,
		&verificationCode.UsedAt,
		&verificationCode.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active verification code: %w", err)
	}

	return verificationCode, nil
}

// MarkCodeAsUsed marks a verification code as used
func (r *TelegramRepositoryPG) MarkCodeAsUsed(ctx context.Context, codeID int64) error {
	query := `UPDATE telegram_verification_codes SET used_at = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, codeID)
	if err != nil {
		return fmt.Errorf("failed to mark code as used: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("verification code not found")
	}

	return nil
}

// DeleteExpiredCodes deletes expired and used verification codes
func (r *TelegramRepositoryPG) DeleteExpiredCodes(ctx context.Context) error {
	query := `DELETE FROM telegram_verification_codes WHERE expires_at < NOW() OR used_at IS NOT NULL`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired codes: %w", err)
	}
	return nil
}

// CreateConnection creates a new Telegram connection
func (r *TelegramRepositoryPG) CreateConnection(ctx context.Context, conn *entities.TelegramConnection) error {
	query := `
		INSERT INTO user_telegram_connections (user_id, telegram_chat_id, telegram_username, telegram_first_name, is_active, connected_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			telegram_chat_id = EXCLUDED.telegram_chat_id,
			telegram_username = EXCLUDED.telegram_username,
			telegram_first_name = EXCLUDED.telegram_first_name,
			is_active = EXCLUDED.is_active,
			updated_at = NOW()`

	_, err := r.db.ExecContext(ctx, query,
		conn.UserID,
		conn.TelegramChatID,
		conn.TelegramUsername,
		conn.TelegramFirstName,
		conn.IsActive,
		conn.ConnectedAt,
		conn.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create connection: %w", err)
	}

	return nil
}

// GetConnectionByUserID retrieves a Telegram connection by user ID
func (r *TelegramRepositoryPG) GetConnectionByUserID(ctx context.Context, userID int64) (*entities.TelegramConnection, error) {
	query := `
		SELECT user_id, telegram_chat_id, telegram_username, telegram_first_name, is_active, connected_at, updated_at
		FROM user_telegram_connections
		WHERE user_id = $1`

	conn := &entities.TelegramConnection{}
	var username, firstName sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&conn.UserID,
		&conn.TelegramChatID,
		&username,
		&firstName,
		&conn.IsActive,
		&conn.ConnectedAt,
		&conn.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	conn.TelegramUsername = username.String
	conn.TelegramFirstName = firstName.String

	return conn, nil
}

// GetConnectionByChatID retrieves a Telegram connection by chat ID
func (r *TelegramRepositoryPG) GetConnectionByChatID(ctx context.Context, chatID int64) (*entities.TelegramConnection, error) {
	query := `
		SELECT user_id, telegram_chat_id, telegram_username, telegram_first_name, is_active, connected_at, updated_at
		FROM user_telegram_connections
		WHERE telegram_chat_id = $1`

	conn := &entities.TelegramConnection{}
	var username, firstName sql.NullString

	err := r.db.QueryRowContext(ctx, query, chatID).Scan(
		&conn.UserID,
		&conn.TelegramChatID,
		&username,
		&firstName,
		&conn.IsActive,
		&conn.ConnectedAt,
		&conn.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get connection by chat ID: %w", err)
	}

	conn.TelegramUsername = username.String
	conn.TelegramFirstName = firstName.String

	return conn, nil
}

// GetActiveConnections retrieves all active Telegram connections
func (r *TelegramRepositoryPG) GetActiveConnections(ctx context.Context) ([]entities.TelegramConnection, error) {
	query := `
		SELECT user_id, telegram_chat_id, telegram_username, telegram_first_name, is_active, connected_at, updated_at
		FROM user_telegram_connections
		WHERE is_active = true`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active connections: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var connections []entities.TelegramConnection
	for rows.Next() {
		conn := entities.TelegramConnection{}
		var username, firstName sql.NullString

		if err := rows.Scan(
			&conn.UserID,
			&conn.TelegramChatID,
			&username,
			&firstName,
			&conn.IsActive,
			&conn.ConnectedAt,
			&conn.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan connection: %w", err)
		}

		conn.TelegramUsername = username.String
		conn.TelegramFirstName = firstName.String
		connections = append(connections, conn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating connections: %w", err)
	}

	return connections, nil
}

// UpdateConnection updates an existing Telegram connection
func (r *TelegramRepositoryPG) UpdateConnection(ctx context.Context, conn *entities.TelegramConnection) error {
	query := `
		UPDATE user_telegram_connections SET
			telegram_chat_id = $2,
			telegram_username = $3,
			telegram_first_name = $4,
			is_active = $5,
			updated_at = NOW()
		WHERE user_id = $1`

	result, err := r.db.ExecContext(ctx, query,
		conn.UserID,
		conn.TelegramChatID,
		conn.TelegramUsername,
		conn.TelegramFirstName,
		conn.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update connection: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("connection not found")
	}

	return nil
}

// DeleteConnection deletes a Telegram connection
func (r *TelegramRepositoryPG) DeleteConnection(ctx context.Context, userID int64) error {
	query := `DELETE FROM user_telegram_connections WHERE user_id = $1`
	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("connection not found")
	}

	return nil
}
