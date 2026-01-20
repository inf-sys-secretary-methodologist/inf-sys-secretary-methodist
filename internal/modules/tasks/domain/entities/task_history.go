package entities

import "time"

// TaskHistory represents a history entry for task changes.
type TaskHistory struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	UserID    *int64    `json:"user_id,omitempty"`
	FieldName string    `json:"field_name"`
	OldValue  *string   `json:"old_value,omitempty"`
	NewValue  *string   `json:"new_value,omitempty"`
	CreatedAt time.Time `json:"created_at"`

	// Associations
	User *HistoryUser `json:"user,omitempty"`
}

// HistoryUser represents basic user info for history response.
type HistoryUser struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// NewTaskHistory creates a new task history entry.
func NewTaskHistory(taskID int64, userID *int64, fieldName string, oldValue, newValue *string) *TaskHistory {
	return &TaskHistory{
		TaskID:    taskID,
		UserID:    userID,
		FieldName: fieldName,
		OldValue:  oldValue,
		NewValue:  newValue,
		CreatedAt: time.Now(),
	}
}
