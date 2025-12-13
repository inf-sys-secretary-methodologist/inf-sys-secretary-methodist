// Package entities provides domain entities for the tasks module.
package entities

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
)

// Task entity errors.
var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrTaskAlreadyCompleted    = errors.New("task is already completed")
	ErrTaskCancelled           = errors.New("task is cancelled")
)

// Task represents a task entity.
type Task struct {
	ID             int64               `db:"id" json:"id"`
	ProjectID      *int64              `db:"project_id" json:"project_id,omitempty"`
	Title          string              `db:"title" json:"title"`
	Description    *string             `db:"description" json:"description,omitempty"`
	DocumentID     *int64              `db:"document_id" json:"document_id,omitempty"`
	AuthorID       int64               `db:"author_id" json:"author_id"`
	AssigneeID     *int64              `db:"assignee_id" json:"assignee_id,omitempty"`
	Status         domain.TaskStatus   `db:"status" json:"status"`
	Priority       domain.TaskPriority `db:"priority" json:"priority"`
	DueDate        *time.Time          `db:"due_date" json:"due_date,omitempty"`
	StartDate      *time.Time          `db:"start_date" json:"start_date,omitempty"`
	CompletedAt    *time.Time          `db:"completed_at" json:"completed_at,omitempty"`
	Progress       int                 `db:"progress" json:"progress"`
	EstimatedHours *float64            `db:"estimated_hours" json:"estimated_hours,omitempty"`
	ActualHours    *float64            `db:"actual_hours" json:"actual_hours,omitempty"`
	Tags           []string            `db:"tags" json:"tags,omitempty"`
	Metadata       json.RawMessage     `db:"metadata" json:"metadata,omitempty"`
	CreatedAt      time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time           `db:"updated_at" json:"updated_at"`

	// Associations (not stored in DB, loaded separately)
	Project     *Project         `db:"-" json:"project,omitempty"`
	Assignee    *TaskAssignee    `db:"-" json:"assignee,omitempty"`
	Watchers    []TaskWatcher    `db:"-" json:"watchers,omitempty"`
	Comments    []TaskComment    `db:"-" json:"comments,omitempty"`
	Attachments []TaskAttachment `db:"-" json:"attachments,omitempty"`
	Checklists  []TaskChecklist  `db:"-" json:"checklists,omitempty"`
}

// TaskAssignee represents basic assignee info for task response.
type TaskAssignee struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// NewTask creates a new task with default values.
func NewTask(title string, authorID int64) *Task {
	now := time.Now()
	return &Task{
		Title:     title,
		AuthorID:  authorID,
		Status:    domain.TaskStatusNew,
		Priority:  domain.TaskPriorityNormal,
		Progress:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Assign assigns the task to a user.
func (t *Task) Assign(assigneeID int64) error {
	if t.Status == domain.TaskStatusCompleted {
		return ErrTaskAlreadyCompleted
	}
	if t.Status == domain.TaskStatusCancelled {
		return ErrTaskCancelled
	}

	t.AssigneeID = &assigneeID
	if t.Status == domain.TaskStatusNew {
		t.Status = domain.TaskStatusAssigned
	}
	t.UpdatedAt = time.Now()
	return nil
}

// Unassign removes the assignee from the task.
func (t *Task) Unassign() error {
	if t.Status == domain.TaskStatusCompleted {
		return ErrTaskAlreadyCompleted
	}
	if t.Status == domain.TaskStatusCancelled {
		return ErrTaskCancelled
	}

	t.AssigneeID = nil
	if t.Status == domain.TaskStatusAssigned {
		t.Status = domain.TaskStatusNew
	}
	t.UpdatedAt = time.Now()
	return nil
}

// StartWork transitions the task to in_progress status.
func (t *Task) StartWork() error {
	if t.Status == domain.TaskStatusCompleted {
		return ErrTaskAlreadyCompleted
	}
	if t.Status == domain.TaskStatusCancelled {
		return ErrTaskCancelled
	}
	if t.Status != domain.TaskStatusNew && t.Status != domain.TaskStatusAssigned && t.Status != domain.TaskStatusDeferred {
		return ErrInvalidStatusTransition
	}

	t.Status = domain.TaskStatusInProgress
	now := time.Now()
	if t.StartDate == nil {
		t.StartDate = &now
	}
	t.UpdatedAt = now
	return nil
}

// SubmitForReview transitions the task to review status.
func (t *Task) SubmitForReview() error {
	if t.Status != domain.TaskStatusInProgress {
		return ErrInvalidStatusTransition
	}

	t.Status = domain.TaskStatusReview
	t.UpdatedAt = time.Now()
	return nil
}

// Complete marks the task as completed.
func (t *Task) Complete() error {
	if t.Status == domain.TaskStatusCompleted {
		return ErrTaskAlreadyCompleted
	}
	if t.Status == domain.TaskStatusCancelled {
		return ErrTaskCancelled
	}
	if t.Status != domain.TaskStatusInProgress && t.Status != domain.TaskStatusReview {
		return ErrInvalidStatusTransition
	}

	t.Status = domain.TaskStatusCompleted
	now := time.Now()
	t.CompletedAt = &now
	t.Progress = 100
	t.UpdatedAt = now
	return nil
}

// Cancel cancels the task.
func (t *Task) Cancel() error {
	if t.Status == domain.TaskStatusCompleted {
		return ErrTaskAlreadyCompleted
	}
	if t.Status == domain.TaskStatusCancelled {
		return ErrTaskCancelled
	}

	t.Status = domain.TaskStatusCancelled
	t.UpdatedAt = time.Now()
	return nil
}

// Defer defers the task.
func (t *Task) Defer() error {
	if t.Status == domain.TaskStatusCompleted {
		return ErrTaskAlreadyCompleted
	}
	if t.Status == domain.TaskStatusCancelled {
		return ErrTaskCancelled
	}

	t.Status = domain.TaskStatusDeferred
	t.UpdatedAt = time.Now()
	return nil
}

// Reopen reopens a completed or cancelled task.
func (t *Task) Reopen() error {
	if t.Status != domain.TaskStatusCompleted && t.Status != domain.TaskStatusCancelled {
		return ErrInvalidStatusTransition
	}

	t.Status = domain.TaskStatusNew
	t.CompletedAt = nil
	t.UpdatedAt = time.Now()
	return nil
}

// SetProgress updates the task progress (0-100).
func (t *Task) SetProgress(progress int) error {
	if t.Status == domain.TaskStatusCompleted {
		return ErrTaskAlreadyCompleted
	}
	if t.Status == domain.TaskStatusCancelled {
		return ErrTaskCancelled
	}

	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	t.Progress = progress
	t.UpdatedAt = time.Now()
	return nil
}

// SetPriority updates the task priority.
func (t *Task) SetPriority(priority domain.TaskPriority) {
	t.Priority = priority
	t.UpdatedAt = time.Now()
}

// SetDueDate updates the task due date.
func (t *Task) SetDueDate(dueDate *time.Time) {
	t.DueDate = dueDate
	t.UpdatedAt = time.Now()
}

// IsOverdue checks if the task is overdue.
func (t *Task) IsOverdue() bool {
	if t.DueDate == nil {
		return false
	}
	if t.Status == domain.TaskStatusCompleted || t.Status == domain.TaskStatusCancelled {
		return false
	}
	return time.Now().After(*t.DueDate)
}

// CanEdit checks if the task can be edited.
func (t *Task) CanEdit() bool {
	return t.Status != domain.TaskStatusCompleted && t.Status != domain.TaskStatusCancelled
}
