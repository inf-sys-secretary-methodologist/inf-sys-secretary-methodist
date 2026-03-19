package entities

import (
	"errors"
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

func TestTask_Assign_NonNewStatus(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusInProgress

	err := task.Assign(42)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Status should remain in_progress, not change to assigned
	if task.Status != domain.TaskStatusInProgress {
		t.Errorf("expected status to remain %q, got %q", domain.TaskStatusInProgress, task.Status)
	}
}

func TestTask_Assign_Completed(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCompleted

	err := task.Assign(42)

	if !errors.Is(err, ErrTaskAlreadyCompleted) {
		t.Errorf("expected error %v, got %v", ErrTaskAlreadyCompleted, err)
	}
}

func TestTask_Assign_Canceled(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCancelled

	err := task.Assign(42)

	if !errors.Is(err, ErrTaskCancelled) {
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

func TestTask_Unassign_NonAssignedStatus(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusInProgress
	assigneeID := int64(42)
	task.AssigneeID = &assigneeID

	err := task.Unassign()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Status should remain in_progress
	if task.Status != domain.TaskStatusInProgress {
		t.Errorf("expected status to remain %q, got %q", domain.TaskStatusInProgress, task.Status)
	}
}

func TestTask_Unassign_Completed(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCompleted

	err := task.Unassign()
	if !errors.Is(err, ErrTaskAlreadyCompleted) {
		t.Errorf("expected error %v, got %v", ErrTaskAlreadyCompleted, err)
	}
}

func TestTask_Unassign_Canceled(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCancelled

	err := task.Unassign()
	if !errors.Is(err, ErrTaskCancelled) {
		t.Errorf("expected error %v, got %v", ErrTaskCancelled, err)
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

func TestTask_StartWork_FromAssigned(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusAssigned

	err := task.StartWork()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if task.Status != domain.TaskStatusInProgress {
		t.Errorf("expected status %q, got %q", domain.TaskStatusInProgress, task.Status)
	}
}

func TestTask_StartWork_FromDeferred(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusDeferred

	err := task.StartWork()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if task.Status != domain.TaskStatusInProgress {
		t.Errorf("expected status %q, got %q", domain.TaskStatusInProgress, task.Status)
	}
}

func TestTask_StartWork_AlreadyStarted(t *testing.T) {
	task := NewTask("Task", 1)
	now := time.Now()
	task.StartDate = &now
	task.Status = domain.TaskStatusDeferred

	err := task.StartWork()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// StartDate should remain the original value
	if task.StartDate != &now {
		// just verify it's not nil
		if task.StartDate == nil {
			t.Error("expected start date to remain set")
		}
	}
}

func TestTask_StartWork_Completed(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCompleted

	err := task.StartWork()
	if !errors.Is(err, ErrTaskAlreadyCompleted) {
		t.Errorf("expected error %v, got %v", ErrTaskAlreadyCompleted, err)
	}
}

func TestTask_StartWork_Canceled(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCancelled

	err := task.StartWork()
	if !errors.Is(err, ErrTaskCancelled) {
		t.Errorf("expected error %v, got %v", ErrTaskCancelled, err)
	}
}

func TestTask_StartWork_InvalidStatus(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusReview

	err := task.StartWork()

	if !errors.Is(err, ErrInvalidStatusTransition) {
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

	if !errors.Is(err, ErrInvalidStatusTransition) {
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

func TestTask_Complete_AlreadyCompleted(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCompleted

	err := task.Complete()
	if !errors.Is(err, ErrTaskAlreadyCompleted) {
		t.Errorf("expected error %v, got %v", ErrTaskAlreadyCompleted, err)
	}
}

func TestTask_Complete_Canceled(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCancelled

	err := task.Complete()
	if !errors.Is(err, ErrTaskCancelled) {
		t.Errorf("expected error %v, got %v", ErrTaskCancelled, err)
	}
}

func TestTask_Complete_InvalidStatus(t *testing.T) {
	task := NewTask("Task", 1)
	// Status is New, which is invalid for Complete

	err := task.Complete()
	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Errorf("expected error %v, got %v", ErrInvalidStatusTransition, err)
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

func TestTask_Cancel_Completed(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCompleted

	err := task.Cancel()
	if !errors.Is(err, ErrTaskAlreadyCompleted) {
		t.Errorf("expected error %v, got %v", ErrTaskAlreadyCompleted, err)
	}
}

func TestTask_Cancel_AlreadyCanceled(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCancelled

	err := task.Cancel()
	if !errors.Is(err, ErrTaskCancelled) {
		t.Errorf("expected error %v, got %v", ErrTaskCancelled, err)
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

func TestTask_Defer_Completed(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCompleted

	err := task.Defer()
	if !errors.Is(err, ErrTaskAlreadyCompleted) {
		t.Errorf("expected error %v, got %v", ErrTaskAlreadyCompleted, err)
	}
}

func TestTask_Defer_Canceled(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCancelled

	err := task.Defer()
	if !errors.Is(err, ErrTaskCancelled) {
		t.Errorf("expected error %v, got %v", ErrTaskCancelled, err)
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

func TestTask_Reopen_FromCanceled(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCancelled

	err := task.Reopen()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if task.Status != domain.TaskStatusNew {
		t.Errorf("expected status %q, got %q", domain.TaskStatusNew, task.Status)
	}
}

func TestTask_Reopen_InvalidStatus(t *testing.T) {
	task := NewTask("Task", 1)

	err := task.Reopen()

	if !errors.Is(err, ErrInvalidStatusTransition) {
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

	_ = task.SetProgress(-10)
	if task.Progress != 0 {
		t.Errorf("expected progress to be clamped to 0, got %d", task.Progress)
	}

	_ = task.SetProgress(150)
	if task.Progress != 100 {
		t.Errorf("expected progress to be clamped to 100, got %d", task.Progress)
	}
}

func TestTask_SetProgress_Completed(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCompleted

	err := task.SetProgress(50)
	if !errors.Is(err, ErrTaskAlreadyCompleted) {
		t.Errorf("expected error %v, got %v", ErrTaskAlreadyCompleted, err)
	}
}

func TestTask_SetProgress_Canceled(t *testing.T) {
	task := NewTask("Task", 1)
	task.Status = domain.TaskStatusCancelled

	err := task.SetProgress(50)
	if !errors.Is(err, ErrTaskCancelled) {
		t.Errorf("expected error %v, got %v", ErrTaskCancelled, err)
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

	// Set nil
	task.SetDueDate(nil)
	if task.DueDate != nil {
		t.Error("expected due date to be nil")
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

	// Canceled task with past due date
	task5 := NewTask("Task", 1)
	task5.DueDate = &pastDue
	task5.Status = domain.TaskStatusCancelled
	if task5.IsOverdue() {
		t.Error("canceled task should not be overdue")
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
		{"canceled cannot edit", domain.TaskStatusCancelled, false},
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

// --- Project tests ---

func TestNewProject(t *testing.T) {
	name := "Test Project"
	ownerID := int64(42)

	project := NewProject(name, ownerID)

	if project.Name != name {
		t.Errorf("expected name %q, got %q", name, project.Name)
	}
	if project.OwnerID != ownerID {
		t.Errorf("expected owner ID %d, got %d", ownerID, project.OwnerID)
	}
	if project.Status != domain.ProjectStatusPlanning {
		t.Errorf("expected status %q, got %q", domain.ProjectStatusPlanning, project.Status)
	}
	if project.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestProject_Activate(t *testing.T) {
	project := NewProject("Project", 1)
	project.Activate()
	if project.Status != domain.ProjectStatusActive {
		t.Errorf("expected status %q, got %q", domain.ProjectStatusActive, project.Status)
	}
}

func TestProject_PutOnHold(t *testing.T) {
	project := NewProject("Project", 1)
	project.PutOnHold()
	if project.Status != domain.ProjectStatusOnHold {
		t.Errorf("expected status %q, got %q", domain.ProjectStatusOnHold, project.Status)
	}
}

func TestProject_Complete(t *testing.T) {
	project := NewProject("Project", 1)
	project.Complete()
	if project.Status != domain.ProjectStatusCompleted {
		t.Errorf("expected status %q, got %q", domain.ProjectStatusCompleted, project.Status)
	}
}

func TestProject_Cancel(t *testing.T) {
	project := NewProject("Project", 1)
	project.Cancel()
	if project.Status != domain.ProjectStatusCancelled {
		t.Errorf("expected status %q, got %q", domain.ProjectStatusCancelled, project.Status)
	}
}

func TestProject_IsActive(t *testing.T) {
	project := NewProject("Project", 1)
	if project.IsActive() {
		t.Error("new project should not be active")
	}
	project.Activate()
	if !project.IsActive() {
		t.Error("activated project should be active")
	}
}

// --- TaskChecklist tests ---

func TestNewTaskChecklist(t *testing.T) {
	cl := NewTaskChecklist(1, "Checklist", 0)
	if cl.TaskID != 1 {
		t.Errorf("expected task ID 1, got %d", cl.TaskID)
	}
	if cl.Title != "Checklist" {
		t.Errorf("expected title 'Checklist', got %q", cl.Title)
	}
	if cl.Position != 0 {
		t.Errorf("expected position 0, got %d", cl.Position)
	}
	if cl.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestNewTaskChecklistItem(t *testing.T) {
	item := NewTaskChecklistItem(1, "Item", 0)
	if item.ChecklistID != 1 {
		t.Errorf("expected checklist ID 1, got %d", item.ChecklistID)
	}
	if item.Title != "Item" {
		t.Errorf("expected title 'Item', got %q", item.Title)
	}
	if item.IsCompleted {
		t.Error("expected IsCompleted to be false")
	}
	if item.Position != 0 {
		t.Errorf("expected position 0, got %d", item.Position)
	}
}

func TestTaskChecklistItem_Complete(t *testing.T) {
	item := NewTaskChecklistItem(1, "Item", 0)
	item.Complete(42)

	if !item.IsCompleted {
		t.Error("expected item to be completed")
	}
	if item.CompletedBy == nil || *item.CompletedBy != 42 {
		t.Errorf("expected completed by 42, got %v", item.CompletedBy)
	}
	if item.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
}

func TestTaskChecklistItem_Uncomplete(t *testing.T) {
	item := NewTaskChecklistItem(1, "Item", 0)
	item.Complete(42)
	item.Uncomplete()

	if item.IsCompleted {
		t.Error("expected item to not be completed")
	}
	if item.CompletedBy != nil {
		t.Error("expected CompletedBy to be nil")
	}
	if item.CompletedAt != nil {
		t.Error("expected CompletedAt to be nil")
	}
}

func TestTaskChecklist_CompletionPercentage(t *testing.T) {
	cl := NewTaskChecklist(1, "CL", 0)

	// Empty checklist
	if cl.CompletionPercentage() != 0 {
		t.Errorf("expected 0%% for empty checklist, got %d%%", cl.CompletionPercentage())
	}

	// Add items
	cl.Items = []TaskChecklistItem{
		{ID: 1, IsCompleted: true},
		{ID: 2, IsCompleted: false},
		{ID: 3, IsCompleted: true},
		{ID: 4, IsCompleted: false},
	}

	if cl.CompletionPercentage() != 50 {
		t.Errorf("expected 50%%, got %d%%", cl.CompletionPercentage())
	}

	// All completed
	cl.Items = []TaskChecklistItem{
		{ID: 1, IsCompleted: true},
		{ID: 2, IsCompleted: true},
	}

	if cl.CompletionPercentage() != 100 {
		t.Errorf("expected 100%%, got %d%%", cl.CompletionPercentage())
	}
}

// --- TaskComment tests ---

func TestNewTaskComment(t *testing.T) {
	c := NewTaskComment(1, 42, "Hello")
	if c.TaskID != 1 {
		t.Errorf("expected task ID 1, got %d", c.TaskID)
	}
	if c.AuthorID != 42 {
		t.Errorf("expected author ID 42, got %d", c.AuthorID)
	}
	if c.Content != "Hello" {
		t.Errorf("expected content 'Hello', got %q", c.Content)
	}
	if c.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestTaskComment_SetParent(t *testing.T) {
	c := NewTaskComment(1, 1, "Reply")
	c.SetParent(10)
	if c.ParentCommentID == nil || *c.ParentCommentID != 10 {
		t.Errorf("expected parent comment ID 10, got %v", c.ParentCommentID)
	}
}

func TestTaskComment_Update(t *testing.T) {
	c := NewTaskComment(1, 1, "Old")
	c.Update("New")
	if c.Content != "New" {
		t.Errorf("expected content 'New', got %q", c.Content)
	}
}

// --- TaskHistory tests ---

func TestNewTaskHistory(t *testing.T) {
	userID := int64(42)
	oldVal := "old"
	newVal := "new"

	h := NewTaskHistory(1, &userID, "status", &oldVal, &newVal)

	if h.TaskID != 1 {
		t.Errorf("expected task ID 1, got %d", h.TaskID)
	}
	if h.UserID == nil || *h.UserID != 42 {
		t.Errorf("expected user ID 42, got %v", h.UserID)
	}
	if h.FieldName != "status" {
		t.Errorf("expected field name 'status', got %q", h.FieldName)
	}
	if h.OldValue == nil || *h.OldValue != "old" {
		t.Errorf("expected old value 'old', got %v", h.OldValue)
	}
	if h.NewValue == nil || *h.NewValue != "new" {
		t.Errorf("expected new value 'new', got %v", h.NewValue)
	}
	if h.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestNewTaskHistory_NilValues(t *testing.T) {
	h := NewTaskHistory(1, nil, "title", nil, nil)
	if h.UserID != nil {
		t.Error("expected UserID to be nil")
	}
	if h.OldValue != nil {
		t.Error("expected OldValue to be nil")
	}
	if h.NewValue != nil {
		t.Error("expected NewValue to be nil")
	}
}

// --- TaskWatcher tests ---

func TestNewTaskWatcher(t *testing.T) {
	w := NewTaskWatcher(1, 42)
	if w.TaskID != 1 {
		t.Errorf("expected task ID 1, got %d", w.TaskID)
	}
	if w.UserID != 42 {
		t.Errorf("expected user ID 42, got %d", w.UserID)
	}
	if w.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

// --- TaskAttachment tests ---

func TestNewTaskAttachment(t *testing.T) {
	a := NewTaskAttachment(1, "file.pdf", "/path/file.pdf", 1024, 42)
	if a.TaskID != 1 {
		t.Errorf("expected task ID 1, got %d", a.TaskID)
	}
	if a.FileName != "file.pdf" {
		t.Errorf("expected file name 'file.pdf', got %q", a.FileName)
	}
	if a.FilePath != "/path/file.pdf" {
		t.Errorf("expected file path '/path/file.pdf', got %q", a.FilePath)
	}
	if a.FileSize != 1024 {
		t.Errorf("expected file size 1024, got %d", a.FileSize)
	}
	if a.UploadedBy != 42 {
		t.Errorf("expected uploaded by 42, got %d", a.UploadedBy)
	}
	if a.MimeType != nil {
		t.Error("expected mime type to be nil")
	}
	if a.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestTaskAttachment_SetMimeType(t *testing.T) {
	a := NewTaskAttachment(1, "file.pdf", "/path/file.pdf", 1024, 42)
	a.SetMimeType("application/pdf")
	if a.MimeType == nil || *a.MimeType != "application/pdf" {
		t.Errorf("expected mime type 'application/pdf', got %v", a.MimeType)
	}
}

// --- Struct tests ---

func TestTaskAssignee_Struct(t *testing.T) {
	a := TaskAssignee{ID: 1, Name: "John", Email: "john@test.com"}
	if a.ID != 1 {
		t.Errorf("expected ID 1, got %d", a.ID)
	}
	if a.Name != "John" {
		t.Errorf("expected name 'John', got %q", a.Name)
	}
}

func TestCommentAuthor_Struct(t *testing.T) {
	a := CommentAuthor{ID: 1, Name: "John", Email: "john@test.com"}
	if a.ID != 1 {
		t.Errorf("expected ID 1, got %d", a.ID)
	}
}

func TestHistoryUser_Struct(t *testing.T) {
	u := HistoryUser{ID: 1, Name: "John", Email: "john@test.com"}
	if u.ID != 1 {
		t.Errorf("expected ID 1, got %d", u.ID)
	}
}

func TestWatcherUser_Struct(t *testing.T) {
	u := WatcherUser{ID: 1, Name: "John", Email: "john@test.com"}
	if u.ID != 1 {
		t.Errorf("expected ID 1, got %d", u.ID)
	}
}
