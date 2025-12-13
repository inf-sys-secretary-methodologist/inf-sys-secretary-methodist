// Package repositories defines repository interfaces for the notifications module.
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
)

// TelegramRepository defines the interface for Telegram-related persistence operations
type TelegramRepository interface {
	// Verification codes
	CreateVerificationCode(ctx context.Context, code *entities.TelegramVerificationCode) error
	GetVerificationCodeByCode(ctx context.Context, code string) (*entities.TelegramVerificationCode, error)
	GetActiveVerificationCodeByUserID(ctx context.Context, userID int64) (*entities.TelegramVerificationCode, error)
	MarkCodeAsUsed(ctx context.Context, codeID int64) error
	DeleteExpiredCodes(ctx context.Context) error

	// Telegram connections
	CreateConnection(ctx context.Context, conn *entities.TelegramConnection) error
	GetConnectionByUserID(ctx context.Context, userID int64) (*entities.TelegramConnection, error)
	GetConnectionByChatID(ctx context.Context, chatID int64) (*entities.TelegramConnection, error)
	UpdateConnection(ctx context.Context, conn *entities.TelegramConnection) error
	DeleteConnection(ctx context.Context, userID int64) error
}
