package entities

import "time"

// TaskWatcher represents a user watching a task.
type TaskWatcher struct {
	TaskID    int64     `db:"task_id" json:"task_id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`

	// Associations
	User *WatcherUser `db:"-" json:"user,omitempty"`
}

// WatcherUser represents basic user info for watcher response.
type WatcherUser struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// NewTaskWatcher creates a new task watcher.
func NewTaskWatcher(taskID, userID int64) *TaskWatcher {
	return &TaskWatcher{
		TaskID:    taskID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
}
