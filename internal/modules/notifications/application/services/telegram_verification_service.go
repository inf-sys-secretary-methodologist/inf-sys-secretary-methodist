// Package services contains application services for the notifications module.
package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

const (
	// VerificationCodeExpiry is the duration after which a verification code expires
	VerificationCodeExpiry = 15 * time.Minute
)

// TelegramVerificationService handles Telegram account verification and linking
type TelegramVerificationService struct {
	telegramRepo    repositories.TelegramRepository
	prefsRepo       repositories.PreferencesRepository
	telegramService services.TelegramService // Composio for sending messages
	auditLog        *logging.AuditLogger
	botUsername     string
}

// NewTelegramVerificationService creates a new Telegram verification service
func NewTelegramVerificationService(
	telegramRepo repositories.TelegramRepository,
	prefsRepo repositories.PreferencesRepository,
	telegramService services.TelegramService,
	auditLog *logging.AuditLogger,
	botUsername string,
) *TelegramVerificationService {
	return &TelegramVerificationService{
		telegramRepo:    telegramRepo,
		prefsRepo:       prefsRepo,
		telegramService: telegramService,
		auditLog:        auditLog,
		botUsername:     botUsername,
	}
}

// GenerateVerificationCodeResponse contains the response for code generation
type GenerateVerificationCodeResponse struct {
	Code        string    `json:"code"`
	ExpiresAt   time.Time `json:"expires_at"`
	BotUsername string    `json:"bot_username"`
	BotLink     string    `json:"bot_link"`
}

// GenerateVerificationCode generates a new verification code for a user
func (s *TelegramVerificationService) GenerateVerificationCode(ctx context.Context, userID int64) (*GenerateVerificationCodeResponse, error) {
	// Check if there's an existing active code
	existingCode, err := s.telegramRepo.GetActiveVerificationCodeByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing code: %w", err)
	}

	// If a valid code exists, return it
	if existingCode != nil && existingCode.IsValid() {
		return &GenerateVerificationCodeResponse{
			Code:        existingCode.Code,
			ExpiresAt:   existingCode.ExpiresAt,
			BotUsername: s.botUsername,
			BotLink:     fmt.Sprintf("https://t.me/%s?start=%s", s.botUsername, existingCode.Code),
		}, nil
	}

	// Generate a new code
	code, err := entities.NewTelegramVerificationCode(userID, VerificationCodeExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate code: %w", err)
	}

	if err := s.telegramRepo.CreateVerificationCode(ctx, code); err != nil {
		return nil, fmt.Errorf("failed to save code: %w", err)
	}

	s.logAudit(ctx, "telegram_verification_code_generated", "telegram", map[string]any{
		"user_id":    userID,
		"expires_at": code.ExpiresAt,
	})

	return &GenerateVerificationCodeResponse{
		Code:        code.Code,
		ExpiresAt:   code.ExpiresAt,
		BotUsername: s.botUsername,
		BotLink:     fmt.Sprintf("https://t.me/%s?start=%s", s.botUsername, code.Code),
	}, nil
}

// VerifyCodeRequest contains the request for code verification
type VerifyCodeRequest struct {
	Code              string
	TelegramChatID    int64
	TelegramUsername  string
	TelegramFirstName string
}

// VerifyCodeResponse contains the response for code verification
type VerifyCodeResponse struct {
	Success bool   `json:"success"`
	UserID  int64  `json:"user_id,omitempty"`
	Message string `json:"message"`
}

// VerifyCode verifies a code and links the Telegram account
func (s *TelegramVerificationService) VerifyCode(ctx context.Context, req *VerifyCodeRequest) (*VerifyCodeResponse, error) {
	// Find the verification code
	code, err := s.telegramRepo.GetVerificationCodeByCode(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to get code: %w", err)
	}

	if code == nil {
		return &VerifyCodeResponse{
			Success: false,
			Message: "Код не найден. Пожалуйста, получите новый код в настройках уведомлений.",
		}, nil
	}

	if code.IsUsed() {
		return &VerifyCodeResponse{
			Success: false,
			Message: "Этот код уже был использован. Пожалуйста, получите новый код.",
		}, nil
	}

	if code.IsExpired() {
		return &VerifyCodeResponse{
			Success: false,
			Message: "Код истёк. Пожалуйста, получите новый код.",
		}, nil
	}

	// Check if this chat ID is already linked to another user
	existingConn, err := s.telegramRepo.GetConnectionByChatID(ctx, req.TelegramChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing connection: %w", err)
	}

	if existingConn != nil && existingConn.UserID != code.UserID {
		return &VerifyCodeResponse{
			Success: false,
			Message: "Этот Telegram аккаунт уже привязан к другому пользователю.",
		}, nil
	}

	// Mark the code as used
	if err := s.telegramRepo.MarkCodeAsUsed(ctx, code.ID); err != nil {
		return nil, fmt.Errorf("failed to mark code as used: %w", err)
	}

	// Create or update the Telegram connection
	conn := &entities.TelegramConnection{
		UserID:            code.UserID,
		TelegramChatID:    req.TelegramChatID,
		TelegramUsername:  req.TelegramUsername,
		TelegramFirstName: req.TelegramFirstName,
		IsActive:          true,
		ConnectedAt:       time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.telegramRepo.CreateConnection(ctx, conn); err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	// Enable Telegram notifications in preferences
	if err := s.prefsRepo.UpdateChannelEnabled(ctx, code.UserID, entities.ChannelTelegram, true); err != nil {
		// Log but don't fail - the connection was created
		s.logAudit(ctx, "telegram_preferences_update_failed", "telegram", map[string]any{
			"user_id": code.UserID,
			"error":   err.Error(),
		})
	}

	s.logAudit(ctx, "telegram_account_linked", "telegram", map[string]any{
		"user_id":    code.UserID,
		"chat_id":    req.TelegramChatID,
		"username":   req.TelegramUsername,
		"first_name": req.TelegramFirstName,
	})

	return &VerifyCodeResponse{
		Success: true,
		UserID:  code.UserID,
		Message: "Telegram аккаунт успешно привязан! Теперь вы будете получать уведомления.",
	}, nil
}

// GetConnection returns the Telegram connection for a user
func (s *TelegramVerificationService) GetConnection(ctx context.Context, userID int64) (*entities.TelegramConnection, error) {
	return s.telegramRepo.GetConnectionByUserID(ctx, userID)
}

// DisconnectTelegram removes the Telegram connection for a user
func (s *TelegramVerificationService) DisconnectTelegram(ctx context.Context, userID int64) error {
	// Delete the connection
	if err := s.telegramRepo.DeleteConnection(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	// Disable Telegram notifications
	if err := s.prefsRepo.UpdateChannelEnabled(ctx, userID, entities.ChannelTelegram, false); err != nil {
		// Log but don't fail
		s.logAudit(ctx, "telegram_preferences_update_failed", "telegram", map[string]any{
			"user_id": userID,
			"error":   err.Error(),
		})
	}

	s.logAudit(ctx, "telegram_account_disconnected", "telegram", map[string]any{
		"user_id": userID,
	})

	return nil
}

// CleanupExpiredCodes removes expired verification codes
func (s *TelegramVerificationService) CleanupExpiredCodes(ctx context.Context) error {
	return s.telegramRepo.DeleteExpiredCodes(ctx)
}

// SendWelcomeMessage sends a welcome message to a newly connected user
func (s *TelegramVerificationService) SendWelcomeMessage(ctx context.Context, chatID int64, firstName string) error {
	message := fmt.Sprintf(
		"Привет, %s! 🎉\n\n"+
			"Ваш Telegram аккаунт успешно привязан к системе.\n\n"+
			"Теперь вы будете получать уведомления прямо сюда.\n\n"+
			"Управлять настройками уведомлений можно в разделе «Настройки» → «Уведомления».",
		firstName,
	)

	// Use Composio TelegramService for sending (chatID as string)
	return s.telegramService.SendNotification(ctx, strconv.FormatInt(chatID, 10), "Добро пожаловать!", message, "normal")
}

// logAudit safely logs an audit event with nil check
func (s *TelegramVerificationService) logAudit(ctx context.Context, action, resourceType string, details map[string]any) {
	if s.auditLog != nil {
		s.auditLog.LogAuditEvent(ctx, action, resourceType, details)
	}
}
