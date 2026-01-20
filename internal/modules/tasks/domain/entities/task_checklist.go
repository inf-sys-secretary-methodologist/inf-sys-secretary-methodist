package entities

import "time"

// TaskChecklist represents a checklist for a task.
type TaskChecklist struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	Title     string    `json:"title"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`

	// Associations
	Items []TaskChecklistItem `json:"items,omitempty"`
}

// TaskChecklistItem represents an item in a checklist.
type TaskChecklistItem struct {
	ID          int64      `json:"id"`
	ChecklistID int64      `json:"checklist_id"`
	Title       string     `json:"title"`
	IsCompleted bool       `json:"is_completed"`
	Position    int        `json:"position"`
	CompletedBy *int64     `json:"completed_by,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// NewTaskChecklist creates a new task checklist.
func NewTaskChecklist(taskID int64, title string, position int) *TaskChecklist {
	return &TaskChecklist{
		TaskID:    taskID,
		Title:     title,
		Position:  position,
		CreatedAt: time.Now(),
	}
}

// NewTaskChecklistItem creates a new checklist item.
func NewTaskChecklistItem(checklistID int64, title string, position int) *TaskChecklistItem {
	return &TaskChecklistItem{
		ChecklistID: checklistID,
		Title:       title,
		IsCompleted: false,
		Position:    position,
		CreatedAt:   time.Now(),
	}
}

// Complete marks the checklist item as completed.
func (i *TaskChecklistItem) Complete(userID int64) {
	i.IsCompleted = true
	i.CompletedBy = &userID
	now := time.Now()
	i.CompletedAt = &now
}

// Uncomplete marks the checklist item as not completed.
func (i *TaskChecklistItem) Uncomplete() {
	i.IsCompleted = false
	i.CompletedBy = nil
	i.CompletedAt = nil
}

// CompletionPercentage calculates the completion percentage of the checklist.
func (c *TaskChecklist) CompletionPercentage() int {
	if len(c.Items) == 0 {
		return 0
	}
	completed := 0
	for _, item := range c.Items {
		if item.IsCompleted {
			completed++
		}
	}
	return (completed * 100) / len(c.Items)
}
