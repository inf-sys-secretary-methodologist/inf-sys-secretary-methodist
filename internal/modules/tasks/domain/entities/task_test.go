package entities

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
)

func TestNewTask(t *testing.T) {
	title := "Test Task"
	authorID := int64(42)

	task := NewTask(title, authorID)

	if task.Title != title {
		t.Errorf("expected title %q, got %q", title, task.Title)
	}
	if task.AuthorID != authorID {
		t.Errorf("expected author ID %d, got %d", authorID, task.AuthorID)
	}
	if task.Status != domain.TaskStatusNew {
		t.Errorf("expected status %q, got %q", domain.TaskStatusNew, task.Status)
	}
	if task.Priority != domain.TaskPriorityNormal {
		t.Errorf("expected priority %q, got %q", domain.TaskPriorityNormal, task.Priority)
	}
	if task.Progress != 0 {
		t.Errorf("expected progress 0, got %d", task.Progress)
	}
	if task.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestTask_Assign(t *testing.T) {
	task := NewTask("Task", 1)
	assigneeID := int64(42)

	err := task.Assign(assigneeID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if task.AssigneeID == nil || *task.AssigneeID != assigneeID {
		t.Errorf("expected assignee ID %d, got %v", assigneeID, task.AssigneeID)
	}
	if task.Status != domain.TaskStatusAssigned {
		t.Errorf("expected status %q, got %q", domain.TaskStatusAssigned, task.Status)
	}
}

func TestTask_Assign_Completed(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCompleted

	err := task.Assign(42)

	if err != ErrTaskAlreadyCompleted {
		t.Errorf("expected error %v, got %v", ErrTaskAlreadyCompleted, err)
	}
}

func TestTask_Assign_Cancelled(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCancelled

	err := task.Assign(42)

	if err != ErrTaskCancelled {
		t.Errorf("expected error %v, got %v", ErrTaskCancelled, err)
	}
}

func TestTask_Unassign(t *testing.T) {
	task := NewTask("Task", 1)
	assigneeID := int64(42)
	task.AssigneeID = &assigneeID
	task.Status = domain.TaskStatusAssigned

	err := task.Unassign()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if task.AssigneeID != nil {
		t.Error("expected assignee ID to be nil")
	}
	if task.Status != domain.TaskStatusNew {
		t.Errorf("expected status %q, got %q", domain.TaskStatusNew, task.Status)
	}
}

func TestTask_StartWork(t *testing.T) {
	task := NewTask("Task", 1)

	err := task.StartWork()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if task.Status != domain.TaskStatusInProgress {
		t.Errorf("expected status %q, got %q", domain.TaskStatusInProgress, task.Status)
	}
	if task.StartDate == nil {
		t.Error("expected start date to be set")
	}
}

func TestTask_StartWork_InvalidStatus(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusReview

	err := task.StartWork()

	if err != ErrInvalidStatusTransition {
		t.Errorf("expected error %v, got %v", ErrInvalidStatusTransition, err)
	}
}

func TestTask_SubmitForReview(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusInProgress

	err := task.SubmitForReview()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if task.Status != domain.TaskStatusReview {
		t.Errorf("expected status %q, got %q", domain.TaskStatusReview, task.Status)
	}
}

func TestTask_SubmitForReview_InvalidStatus(t *testing.T) {
	task := NewTask("Task", 1)

	err := task.SubmitForReview()

	if err != ErrInvalidStatusTransition {
		t.Errorf("expected error %v, got %v", ErrInvalidStatusTransition, err)
	}
}

func TestTask_Complete(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusInProgress

	err := task.Complete()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if task.Status != domain.TaskStatusCompleted {
		t.Errorf("expected status %q, got %q", domain.TaskStatusCompleted, task.Status)
	}
	if task.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
	if task.Progress != 100 {
		t.Errorf("expected progress 100, got %d", task.Progress)
	}
}

func TestTask_Complete_FromReview(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusReview

	err := task.Complete()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if task.Status != domain.TaskStatusCompleted {
		t.Errorf("expected status %q, got %q", domain.TaskStatusCompleted, task.Status)
	}
}

func TestTask_Cancel(t *testing.T) {
	task := NewTask("Task", 1)

	err := task.Cancel()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if task.Status != domain.TaskStatusCancelled {
		t.Errorf("expected status %q, got %q", domain.TaskStatusCancelled, task.Status)
	}
}

func TestTask_Defer(t *testing.T) {
	task := NewTask("Task", 1)

	err := task.Defer()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if task.Status != domain.TaskStatusDeferred {
		t.Errorf("expected status %q, got %q", domain.TaskStatusDeferred, task.Status)
	}
}

func TestTask_Reopen(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCompleted
	now := time.Now()
	task.CompletedAt = &now

	err := task.Reopen()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if task.Status != domain.TaskStatusNew {
		t.Errorf("expected status %q, got %q", domain.TaskStatusNew, task.Status)
	}
	if task.CompletedAt != nil {
		t.Error("expected completed_at to be nil")
	}
}

func TestTask_Reopen_InvalidStatus(t *testing.T) {
	task := NewTask("Task", 1)

	err := task.Reopen()

	if err != ErrInvalidStatusTransition {
		t.Errorf("expected error %v, got %v", ErrInvalidStatusTransition, err)
	}
}

func TestTask_SetProgress(t *testing.T) {
	task := NewTask("Task", 1)

	err := task.SetProgress(50)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if task.Progress != 50 {
		t.Errorf("expected progress 50, got %d", task.Progress)
	}
}

func TestTask_SetProgress_Clamping(t *testing.T) {
	task := NewTask("Task", 1)

	task.SetProgress(-10)
	if task.Progress != 0 {
		t.Errorf("expected progress to be clamped to 0, got %d", task.Progress)
	}

	task.SetProgress(150)
	if task.Progress != 100 {
		t.Errorf("expected progress to be clamped to 100, got %d", task.Progress)
	}
}

func TestTask_SetPriority(t *testing.T) {
	task := NewTask("Task", 1)

	task.SetPriority(domain.TaskPriorityHigh)

	if task.Priority != domain.TaskPriorityHigh {
		t.Errorf("expected priority %q, got %q", domain.TaskPriorityHigh, task.Priority)
	}
}

func TestTask_SetDueDate(t *testing.T) {
	task := NewTask("Task", 1)
	dueDate := time.Now().Add(24 * time.Hour)

	task.SetDueDate(&dueDate)

	if task.DueDate == nil || !task.DueDate.Equal(dueDate) {
		t.Errorf("expected due date %v, got %v", dueDate, task.DueDate)
	}
}

func TestTask_IsOverdue(t *testing.T) {
	// Task without due date
	task1 := NewTask("Task", 1)
	if task1.IsOverdue() {
		t.Error("task without due date should not be overdue")
	}

	// Task with future due date
	futureDue := time.Now().Add(24 * time.Hour)
	task2 := NewTask("Task", 1)
	task2.DueDate = &futureDue
	if task2.IsOverdue() {
		t.Error("task with future due date should not be overdue")
	}

	// Task with past due date
	pastDue := time.Now().Add(-24 * time.Hour)
	task3 := NewTask("Task", 1)
	task3.DueDate = &pastDue
	if !task3.IsOverdue() {
		t.Error("task with past due date should be overdue")
	}

	// Completed task with past due date
	task4 := NewTask("Task", 1)
	task4.DueDate = &pastDue
	task4.Status = domain.TaskStatusCompleted
	if task4.IsOverdue() {
		t.Error("completed task should not be overdue")
	}

	// Cancelled task with past due date
	task5 := NewTask("Task", 1)
	task5.DueDate = &pastDue
	task5.Status = domain.TaskStatusCancelled
	if task5.IsOverdue() {
		t.Error("cancelled task should not be overdue")
	}
}

func TestTask_CanEdit(t *testing.T) {
	tests := []struct {
		name   string
		status domain.TaskStatus
		want   bool
	}{
		{"new can edit", domain.TaskStatusNew, true},
		{"assigned can edit", domain.TaskStatusAssigned, true},
		{"in_progress can edit", domain.TaskStatusInProgress, true},
		{"review can edit", domain.TaskStatusReview, true},
		{"deferred can edit", domain.TaskStatusDeferred, true},
		{"completed cannot edit", domain.TaskStatusCompleted, false},
		{"cancelled cannot edit", domain.TaskStatusCancelled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask("Task", 1)
			task.Status = tt.status

			got := task.CanEdit()
			if got != tt.want {
				t.Errorf("CanEdit() = %v, want %v", got, tt.want)
			}
		})
	}
}
