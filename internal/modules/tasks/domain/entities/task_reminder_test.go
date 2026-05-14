package entities

import (
	"errors"
	"testing"
	"time"
)

// TestReminderType_IsValid pins the enumeration boundary. Mirror
// к migration 038 CHECK constraint — any new dispatch channel
// must be added in both places (sync trap intentional: forces
// reviewer to remember the SQL update).
func TestReminderType_IsValid(t *testing.T) {
	cases := []struct {
		name  string
		input ReminderType
		want  bool
	}{
		{"email", ReminderTypeEmail, true},
		{"push", ReminderTypePush, true},
		{"in_app", ReminderTypeInApp, true},
		{"telegram", ReminderTypeTelegram, true},
		{"unknown_lowercase", ReminderType("slack"), false},
		{"empty", ReminderType(""), false},
		{"case_mismatch", ReminderType("Email"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.input.IsValid(); got != tc.want {
				t.Fatalf("IsValid(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

// TestNewTaskReminder_HappyPath verifies the constructor accepts
// valid input and exposes every field through getters.
func TestNewTaskReminder_HappyPath(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	tr, err := NewTaskReminder(42, 7, ReminderTypeTelegram, 15, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr == nil {
		t.Fatal("expected non-nil reminder")
	}
	if got := tr.TaskID(); got != 42 {
		t.Errorf("TaskID() = %d, want 42", got)
	}
	if got := tr.UserID(); got != 7 {
		t.Errorf("UserID() = %d, want 7", got)
	}
	if got := tr.ReminderType(); got != ReminderTypeTelegram {
		t.Errorf("ReminderType() = %q, want %q", got, ReminderTypeTelegram)
	}
	if got := tr.MinutesBefore(); got != 15 {
		t.Errorf("MinutesBefore() = %d, want 15", got)
	}
	if got := tr.IsSent(); got {
		t.Error("IsSent() = true, want false at construction")
	}
	if got := tr.SentAt(); got != nil {
		t.Errorf("SentAt() = %v, want nil at construction", got)
	}
	if got := tr.CreatedAt(); !got.Equal(now) {
		t.Errorf("CreatedAt() = %v, want %v", got, now)
	}
	if got := tr.ID(); got != 0 {
		t.Errorf("ID() = %d, want 0 before persistence", got)
	}
}

// TestNewTaskReminder_InvalidInput is table-driven (CLAUDE.md ≥3
// variants gate) — covers all four sentinel-return paths.
func TestNewTaskReminder_InvalidInput(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name          string
		taskID        int64
		userID        int64
		reminderType  ReminderType
		minutesBefore int
		wantErr       error
	}{
		{"task_id_zero", 0, 7, ReminderTypeEmail, 15, ErrInvalidTaskID},
		{"task_id_negative", -1, 7, ReminderTypeEmail, 15, ErrInvalidTaskID},
		{"user_id_zero", 42, 0, ReminderTypeEmail, 15, ErrInvalidUserID},
		{"user_id_negative", 42, -7, ReminderTypeEmail, 15, ErrInvalidUserID},
		{"reminder_type_empty", 42, 7, ReminderType(""), 15, ErrInvalidReminderType},
		{"reminder_type_unknown", 42, 7, ReminderType("slack"), 15, ErrInvalidReminderType},
		{"minutes_before_zero", 42, 7, ReminderTypeEmail, 0, ErrInvalidMinutesBefore},
		{"minutes_before_negative", 42, 7, ReminderTypeEmail, -5, ErrInvalidMinutesBefore},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tr, err := NewTaskReminder(tc.taskID, tc.userID, tc.reminderType, tc.minutesBefore, now)
			if tr != nil {
				t.Errorf("expected nil reminder on invalid input, got %+v", tr)
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want sentinel %v (errors.Is must match)", err, tc.wantErr)
			}
		})
	}
}

// TestTaskReminder_MarkSent verifies the dispatch flip — IsSent
// becomes true and SentAt holds the dispatch timestamp.
func TestTaskReminder_MarkSent(t *testing.T) {
	created := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	tr, err := NewTaskReminder(42, 7, ReminderTypeInApp, 30, created)
	if err != nil {
		t.Fatalf("NewTaskReminder: %v", err)
	}
	dispatched := time.Date(2026, 5, 14, 14, 30, 0, 0, time.UTC)
	tr.MarkSent(dispatched)
	if !tr.IsSent() {
		t.Error("IsSent() = false after MarkSent")
	}
	if got := tr.SentAt(); got == nil || !got.Equal(dispatched) {
		t.Errorf("SentAt() = %v, want %v", got, dispatched)
	}
}

// TestHydrateFromPersistence_BypassesValidation confirms the repo
// seam reconstructs entities without going through invariants — a
// persisted row was validated at insert time, не нужно re-validate.
func TestHydrateFromPersistence_BypassesValidation(t *testing.T) {
	created := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	sentAt := time.Date(2026, 5, 14, 14, 30, 0, 0, time.UTC)
	// Construct with values that would FAIL NewTaskReminder (e.g.,
	// minutes_before=0) — hydrate must accept them since the row
	// was inserted historically.
	tr := HydrateFromPersistence(101, 42, 7, ReminderTypeTelegram, 0, true, &sentAt, created)
	if tr.ID() != 101 {
		t.Errorf("ID() = %d, want 101", tr.ID())
	}
	if tr.MinutesBefore() != 0 {
		t.Errorf("MinutesBefore() = %d, want 0 (bypassed validation)", tr.MinutesBefore())
	}
	if !tr.IsSent() {
		t.Error("IsSent() = false, want true")
	}
}
