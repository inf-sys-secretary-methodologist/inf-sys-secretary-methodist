package entities

import (
	"testing"
	"time"
)

func TestSession_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "expired session",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "valid session",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &Session{
				ID:           1,
				UserID:       42,
				RefreshToken: "token123",
				ExpiresAt:    tt.expiresAt,
			}

			got := session.IsExpired()
			if got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_IsActive(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "active session",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      true,
		},
		{
			name:      "expired session is not active",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &Session{
				ID:           1,
				UserID:       42,
				RefreshToken: "token123",
				ExpiresAt:    tt.expiresAt,
			}

			got := session.IsActive()
			if got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_Fields(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	session := &Session{
		ID:           1,
		UserID:       42,
		RefreshToken: "refresh_token_abc",
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "192.168.1.1",
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if session.ID != 1 {
		t.Errorf("expected ID 1, got %d", session.ID)
	}
	if session.UserID != 42 {
		t.Errorf("expected UserID 42, got %d", session.UserID)
	}
	if session.RefreshToken != "refresh_token_abc" {
		t.Errorf("expected RefreshToken %q, got %q", "refresh_token_abc", session.RefreshToken)
	}
	if session.UserAgent != "Mozilla/5.0" {
		t.Errorf("expected UserAgent %q, got %q", "Mozilla/5.0", session.UserAgent)
	}
	if session.IPAddress != "192.168.1.1" {
		t.Errorf("expected IPAddress %q, got %q", "192.168.1.1", session.IPAddress)
	}
}
