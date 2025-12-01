package entities

import "time"

// TaskHistory represents a history entry for task changes.
type TaskHistory struct {
	ID        int64     `db:"id" json:"id"`
	TaskID    int64     `db:"task_id" json:"task_id"`
	UserID    *int64    `db:"user_id" json:"user_id,omitempty"`
	FieldName string    `db:"field_name" json:"field_name"`
	OldValue  *string   `db:"old_value" json:"old_value,omitempty"`
	NewValue  *string   `db:"new_value" json:"new_value,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`

	// Associations
	User *HistoryUser `db:"-" json:"user,omitempty"`
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
