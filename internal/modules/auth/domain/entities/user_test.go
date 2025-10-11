package entities

import (
	"testing"
)

func TestNewUser(t *testing.T) {
	user := NewUser("test-id", "test@example.com", "hashedpass", "Test User", RoleStudent)

	if user.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got '%s'", user.ID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", user.Email)
	}
	if user.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", user.Name)
	}
	if user.Role != RoleStudent {
		t.Errorf("expected role RoleStudent, got '%s'", user.Role)
	}
	if user.Status != UserStatusActive {
		t.Errorf("expected status UserStatusActive, got '%s'", user.Status)
	}
}

func TestUser_Activate(t *testing.T) {
	user := NewUser("test-id", "test@example.com", "hashedpass", "Test User", RoleStudent)
	user.Status = UserStatusInactive

	user.Activate()

	if user.Status != UserStatusActive {
		t.Errorf("expected status UserStatusActive after Activate, got '%s'", user.Status)
	}
}

func TestUser_Deactivate(t *testing.T) {
	user := NewUser("test-id", "test@example.com", "hashedpass", "Test User", RoleStudent)

	user.Deactivate()

	if user.Status != UserStatusInactive {
		t.Errorf("expected status UserStatusInactive after Deactivate, got '%s'", user.Status)
	}
}

func TestUser_Block(t *testing.T) {
	user := NewUser("test-id", "test@example.com", "hashedpass", "Test User", RoleStudent)

	user.Block()

	if user.Status != UserStatusBlocked {
		t.Errorf("expected status UserStatusBlocked after Block, got '%s'", user.Status)
	}
}

func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   UserStatus
		expected bool
	}{
		{"active user", UserStatusActive, true},
		{"inactive user", UserStatusInactive, false},
		{"blocked user", UserStatusBlocked, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser("test-id", "test@example.com", "hashedpass", "Test User", RoleStudent)
			user.Status = tt.status

			if got := user.IsActive(); got != tt.expected {
				t.Errorf("IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}
