package entities

import "time"

// TaskWatcher represents a user watching a task.
type TaskWatcher struct {
	TaskID    int64     `json:"task_id"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`

	// Associations
	User *WatcherUser `json:"user,omitempty"`
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
