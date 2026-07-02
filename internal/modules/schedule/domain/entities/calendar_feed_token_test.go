package entities

import (
	"encoding/hex"
	"errors"
	"testing"
	"time"
)

func TestNewCalendarFeedToken_Valid(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	tok, err := NewCalendarFeedToken(42, "abc123", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.UserID != 42 {
		t.Errorf("expected user id 42, got %d", tok.UserID)
	}
	if tok.Token != "abc123" {
		t.Errorf("expected token abc123, got %q", tok.Token)
	}
	if !tok.CreatedAt.Equal(now) {
		t.Errorf("expected created at %v, got %v", now, tok.CreatedAt)
	}
}

func TestNewCalendarFeedToken_Invalid(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		userID  int64
		token   string
		wantErr error
	}{
		{"zero user id", 0, "abc", ErrCalendarFeedTokenUserRequired},
		{"negative user id", -1, "abc", ErrCalendarFeedTokenUserRequired},
		{"empty token", 42, "", ErrCalendarFeedTokenEmpty},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewCalendarFeedToken(tc.userID, tc.token, now)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestGenerateCalendarFeedToken(t *testing.T) {
	a, err := GenerateCalendarFeedToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a) != 64 {
		t.Errorf("expected 64 hex chars (256 bits), got %d", len(a))
	}
	if _, err := hex.DecodeString(a); err != nil {
		t.Errorf("token is not valid hex: %v", err)
	}

	b, err := GenerateCalendarFeedToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == b {
		t.Errorf("two generated tokens must differ, both were %q", a)
	}
}
