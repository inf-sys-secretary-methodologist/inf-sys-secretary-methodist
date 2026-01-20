// Package entities contains domain entities for the notifications module.
package entities

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// TelegramVerificationCode represents a verification code for linking Telegram accounts
type TelegramVerificationCode struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	Code      string     `json:"code"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// TelegramConnection represents a user's Telegram connection
type TelegramConnection struct {
	UserID            int64     `json:"user_id"`
	TelegramChatID    int64     `json:"telegram_chat_id"`
	TelegramUsername  string    `json:"telegram_username,omitempty"`
	TelegramFirstName string    `json:"telegram_first_name,omitempty"`
	IsActive          bool      `json:"is_active"`
	ConnectedAt       time.Time `json:"connected_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// NewTelegramVerificationCode creates a new verification code for a user
func NewTelegramVerificationCode(userID int64, expiresIn time.Duration) (*TelegramVerificationCode, error) {
	code, err := generateSecureCode()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &TelegramVerificationCode{
		UserID:    userID,
		Code:      code,
		ExpiresAt: now.Add(expiresIn),
		CreatedAt: now,
	}, nil
}

// IsExpired checks if the verification code has expired
func (c *TelegramVerificationCode) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// IsUsed checks if the verification code has been used
func (c *TelegramVerificationCode) IsUsed() bool {
	return c.UsedAt != nil
}

// IsValid checks if the verification code is valid (not expired and not used)
func (c *TelegramVerificationCode) IsValid() bool {
	return !c.IsExpired() && !c.IsUsed()
}

// MarkAsUsed marks the verification code as used
func (c *TelegramVerificationCode) MarkAsUsed() {
	now := time.Now()
	c.UsedAt = &now
}

// generateSecureCode generates a cryptographically secure 8-character code
func generateSecureCode() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
