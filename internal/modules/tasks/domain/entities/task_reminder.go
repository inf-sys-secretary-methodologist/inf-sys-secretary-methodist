// Package entities — TaskReminder aggregate for per-user task
// deadline reminders. Greenfield в v0.138.0. The TaskReminder entity
// is intentionally lightweight (no rich behavior) — it's a value
// captured at SetReminder time + dispatch-side fields (is_sent +
// sent_at) maintained by the scheduler. Cross-aggregate computation
// (remind_at = task.due_date - minutes_before) lives at the
// repository SQL level so that reminders auto-shift when task.
// due_date moves.
package entities

import (
	"errors"
	"time"
)

// ErrInvalidTaskID — task_id must be a positive integer. Mapped к
// HTTP 422 by the handler (domain validation error).
var ErrInvalidTaskID = errors.New("task_reminder: invalid task id")

// ErrInvalidUserID — user_id must be a positive integer.
var ErrInvalidUserID = errors.New("task_reminder: invalid user id")

// ErrInvalidReminderType — reminder_type must be one of the
// dispatch channels supported by the scheduler.
var ErrInvalidReminderType = errors.New("task_reminder: invalid reminder type")

// ErrInvalidMinutesBefore — minutes_before must be > 0. A zero or
// negative value would either fire immediately or never; both are
// useless states that should be rejected at the boundary.
var ErrInvalidMinutesBefore = errors.New("task_reminder: invalid minutes before")

// ReminderType enumerates the dispatch channels supported by the
// notifications scheduler. The set mirrors the event_reminders
// CHECK constraint (migration 014) so that the scheduler dispatch
// switch is identical across both reminder tables.
type ReminderType string

// ReminderType values — keep в sync с migration 038 CHECK.
const (
	ReminderTypeEmail    ReminderType = "email"
	ReminderTypePush     ReminderType = "push"
	ReminderTypeInApp    ReminderType = "in_app"
	ReminderTypeTelegram ReminderType = "telegram"
)

// validReminderTypes — the set of accepted ReminderType values.
// Kept private so callers must use the typed constants.
var validReminderTypes = map[ReminderType]struct{}{
	ReminderTypeEmail:    {},
	ReminderTypePush:     {},
	ReminderTypeInApp:    {},
	ReminderTypeTelegram: {},
}

// IsValid reports whether r is one of the recognized ReminderType
// constants. Reject unknown values to keep the dispatch switch
// exhaustive (default-deny).
func (r ReminderType) IsValid() bool {
	_, ok := validReminderTypes[r]
	return ok
}

// TaskReminder is the singleton-per-user reminder row tied to a
// project-mgmt Task by FK. Persisted in `task_reminders` (migration
// 038). Constructed via NewTaskReminder so all invariants run.
type TaskReminder struct {
	id            int64
	taskID        int64
	userID        int64
	reminderType  ReminderType
	minutesBefore int
	isSent        bool
	sentAt        *time.Time
	createdAt     time.Time
}

// NewTaskReminder builds a TaskReminder instance validating every
// invariant в одной точке. Returns the first violation as a typed
// sentinel for errors.Is matching. `now` is injected так что tests
// pin deterministic timestamps.
func NewTaskReminder(taskID, userID int64, reminderType ReminderType, minutesBefore int, now time.Time) (*TaskReminder, error) {
	return nil, errors.New("task_reminder: not implemented yet")
}

// ID returns the persistence-assigned identifier (0 until persisted).
func (t *TaskReminder) ID() int64 { return t.id }

// TaskID returns the parent task identifier.
func (t *TaskReminder) TaskID() int64 { return t.taskID }

// UserID returns the user this reminder fires for.
func (t *TaskReminder) UserID() int64 { return t.userID }

// ReminderType returns the dispatch channel.
func (t *TaskReminder) ReminderType() ReminderType { return t.reminderType }

// MinutesBefore returns the relative offset (in minutes) before the
// task's due_date at which the scheduler should dispatch.
func (t *TaskReminder) MinutesBefore() int { return t.minutesBefore }

// IsSent reports whether the scheduler has already dispatched this
// reminder. Set by the dispatch path; never directly mutated by
// domain callers outside the entity package.
func (t *TaskReminder) IsSent() bool { return t.isSent }

// SentAt returns the wall-clock timestamp of the dispatch, or nil
// if IsSent is false.
func (t *TaskReminder) SentAt() *time.Time { return t.sentAt }

// CreatedAt returns the creation timestamp.
func (t *TaskReminder) CreatedAt() time.Time { return t.createdAt }

// MarkSent flips the dispatch flags. Called by the scheduler after
// successful (or graceful-fallback) delivery.
func (t *TaskReminder) MarkSent(now time.Time) {
	t.isSent = true
	t.sentAt = &now
}

// HydrateFromPersistence rebuilds an entity from a persisted row.
// Bypasses invariant validation — the row was validated at insert
// time. Repository-only seam; do NOT call from use cases.
func HydrateFromPersistence(id, taskID, userID int64, reminderType ReminderType, minutesBefore int, isSent bool, sentAt *time.Time, createdAt time.Time) *TaskReminder {
	return &TaskReminder{
		id:            id,
		taskID:        taskID,
		userID:        userID,
		reminderType:  reminderType,
		minutesBefore: minutesBefore,
		isSent:        isSent,
		sentAt:        sentAt,
		createdAt:     createdAt,
	}
}
