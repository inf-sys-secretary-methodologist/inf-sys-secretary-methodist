package entities

import (
	"testing"
	"time"
)

func TestNewTelegramVerificationCode(t *testing.T) {
	userID := int64(42)
	expiresIn := 15 * time.Minute

	code, err := NewTelegramVerificationCode(userID, expiresIn)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if code.UserID != userID {
		t.Errorf("expected user ID %d, got %d", userID, code.UserID)
	}

	if code.Code == "" {
		t.Error("expected non-empty code")
	}

	if len(code.Code) != 8 {
		t.Errorf("expected code length 8, got %d", len(code.Code))
	}

	if code.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if code.ExpiresAt.IsZero() {
		t.Error("expected ExpiresAt to be set")
	}

	if code.UsedAt != nil {
		t.Error("expected UsedAt to be nil")
	}

	// Check that ExpiresAt is approximately createdAt + expiresIn
	expectedExpiry := code.CreatedAt.Add(expiresIn)
	if code.ExpiresAt.Sub(expectedExpiry) > time.Second {
		t.Errorf("expected ExpiresAt close to %v, got %v", expectedExpiry, code.ExpiresAt)
	}
}

func TestNewTelegramVerificationCode_UniqueCode(t *testing.T) {
	userID := int64(42)
	expiresIn := 15 * time.Minute

	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, err := NewTelegramVerificationCode(userID, expiresIn)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if codes[code.Code] {
			t.Errorf("duplicate code generated: %s", code.Code)
		}
		codes[code.Code] = true
	}
}

func TestTelegramVerificationCode_IsExpired_NotExpired(t *testing.T) {
	code := &TelegramVerificationCode{
		UserID:    42,
		Code:      "test1234",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
	}

	if code.IsExpired() {
		t.Error("expected code to not be expired")
	}
}

func TestTelegramVerificationCode_IsExpired_Expired(t *testing.T) {
	code := &TelegramVerificationCode{
		UserID:    42,
		Code:      "test1234",
		ExpiresAt: time.Now().Add(-1 * time.Minute),
		CreatedAt: time.Now().Add(-16 * time.Minute),
	}

	if !code.IsExpired() {
		t.Error("expected code to be expired")
	}
}

func TestTelegramVerificationCode_IsExpired_JustExpired(t *testing.T) {
	code := &TelegramVerificationCode{
		UserID:    42,
		Code:      "test1234",
		ExpiresAt: time.Now().Add(-1 * time.Millisecond),
		CreatedAt: time.Now().Add(-15 * time.Minute),
	}

	if !code.IsExpired() {
		t.Error("expected code to be expired")
	}
}

func TestTelegramVerificationCode_IsUsed_NotUsed(t *testing.T) {
	code := &TelegramVerificationCode{
		UserID:    42,
		Code:      "test1234",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
		UsedAt:    nil,
	}

	if code.IsUsed() {
		t.Error("expected code to not be used")
	}
}

func TestTelegramVerificationCode_IsUsed_Used(t *testing.T) {
	usedTime := time.Now()
	code := &TelegramVerificationCode{
		UserID:    42,
		Code:      "test1234",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now().Add(-5 * time.Minute),
		UsedAt:    &usedTime,
	}

	if !code.IsUsed() {
		t.Error("expected code to be used")
	}
}

func TestTelegramVerificationCode_IsValid_Valid(t *testing.T) {
	code := &TelegramVerificationCode{
		UserID:    42,
		Code:      "test1234",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
		UsedAt:    nil,
	}

	if !code.IsValid() {
		t.Error("expected code to be valid")
	}
}

func TestTelegramVerificationCode_IsValid_Expired(t *testing.T) {
	code := &TelegramVerificationCode{
		UserID:    42,
		Code:      "test1234",
		ExpiresAt: time.Now().Add(-1 * time.Minute),
		CreatedAt: time.Now().Add(-16 * time.Minute),
		UsedAt:    nil,
	}

	if code.IsValid() {
		t.Error("expected expired code to be invalid")
	}
}

func TestTelegramVerificationCode_IsValid_Used(t *testing.T) {
	usedTime := time.Now()
	code := &TelegramVerificationCode{
		UserID:    42,
		Code:      "test1234",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now().Add(-5 * time.Minute),
		UsedAt:    &usedTime,
	}

	if code.IsValid() {
		t.Error("expected used code to be invalid")
	}
}

func TestTelegramVerificationCode_IsValid_ExpiredAndUsed(t *testing.T) {
	usedTime := time.Now().Add(-10 * time.Minute)
	code := &TelegramVerificationCode{
		UserID:    42,
		Code:      "test1234",
		ExpiresAt: time.Now().Add(-1 * time.Minute),
		CreatedAt: time.Now().Add(-16 * time.Minute),
		UsedAt:    &usedTime,
	}

	if code.IsValid() {
		t.Error("expected expired and used code to be invalid")
	}
}

func TestTelegramVerificationCode_MarkAsUsed(t *testing.T) {
	code := &TelegramVerificationCode{
		UserID:    42,
		Code:      "test1234",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
		UsedAt:    nil,
	}

	if code.IsUsed() {
		t.Error("expected code to not be used before marking")
	}

	beforeMark := time.Now()
	code.MarkAsUsed()
	afterMark := time.Now()

	if !code.IsUsed() {
		t.Error("expected code to be used after marking")
	}

	if code.UsedAt == nil {
		t.Fatal("expected UsedAt to be set")
	}

	if code.UsedAt.Before(beforeMark) || code.UsedAt.After(afterMark) {
		t.Errorf("expected UsedAt %v to be between %v and %v", code.UsedAt, beforeMark, afterMark)
	}
}

func TestTelegramVerificationCode_MarkAsUsed_InvalidatesCode(t *testing.T) {
	code := &TelegramVerificationCode{
		UserID:    42,
		Code:      "test1234",
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
		UsedAt:    nil,
	}

	if !code.IsValid() {
		t.Error("expected code to be valid before marking")
	}

	code.MarkAsUsed()

	if code.IsValid() {
		t.Error("expected code to be invalid after marking as used")
	}
}

func TestTelegramConnectionStruct(t *testing.T) {
	now := time.Now()
	conn := TelegramConnection{
		UserID:            42,
		TelegramChatID:    123456789,
		TelegramUsername:  "testuser",
		TelegramFirstName: "Test",
		IsActive:          true,
		ConnectedAt:       now,
		UpdatedAt:         now,
	}

	if conn.UserID != 42 {
		t.Errorf("expected user ID 42, got %d", conn.UserID)
	}
	if conn.TelegramChatID != 123456789 {
		t.Errorf("expected chat ID 123456789, got %d", conn.TelegramChatID)
	}
	if conn.TelegramUsername != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", conn.TelegramUsername)
	}
	if conn.TelegramFirstName != "Test" {
		t.Errorf("expected first name 'Test', got '%s'", conn.TelegramFirstName)
	}
	if !conn.IsActive {
		t.Error("expected IsActive to be true")
	}
	if conn.ConnectedAt.IsZero() {
		t.Error("expected ConnectedAt to be set")
	}
	if conn.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}
