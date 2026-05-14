package dto

import "time"

// CreateTaskReminderRequest is the body accepted by POST
// /api/tasks/:id/reminders. ActorUserID is NOT in the body — it
// derives from the JWT context inside the handler.
type CreateTaskReminderRequest struct {
	ReminderType  string `json:"reminder_type"`
	MinutesBefore int    `json:"minutes_before"`
}

// TaskReminderResponse is the JSON projection of a TaskReminder
// entity. Used by POST (response) and GET (list element).
type TaskReminderResponse struct {
	ID            int64      `json:"id"`
	TaskID        int64      `json:"task_id"`
	UserID        int64      `json:"user_id"`
	ReminderType  string     `json:"reminder_type"`
	MinutesBefore int        `json:"minutes_before"`
	IsSent        bool       `json:"is_sent"`
	SentAt        *time.Time `json:"sent_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
