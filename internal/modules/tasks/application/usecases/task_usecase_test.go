package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
)

const updatedTaskTitle = "Updated Task"

// MockTaskRepository implements TaskRepository for testing.
type MockTaskRepository struct {
	tasks      map[int64]*entities.Task
	comments   map[int64]*entities.TaskComment
	checklists map[int64]*entities.TaskChecklist
	items      map[int64]*entities.TaskChecklistItem
	watchers   map[int64][]*entities.TaskWatcher
	history    map[int64][]*entities.TaskHistory
	nextID     int64
}

func NewMockTaskRepository() *MockTaskRepository {
	return &MockTaskRepository{
		tasks:      make(map[int64]*entities.Task),
		comments:   make(map[int64]*entities.TaskComment),
		checklists: make(map[int64]*entities.TaskChecklist),
		items:      make(map[int64]*entities.TaskChecklistItem),
		watchers:   make(map[int64][]*entities.TaskWatcher),
		history:    make(map[int64][]*entities.TaskHistory),
		nextID:     1,
	}
}

func (m *MockTaskRepository) Create(_ context.Context, task *entities.Task) error {
	task.ID = m.nextID
	m.nextID++
	m.tasks[task.ID] = task
	return nil
}

func (m *MockTaskRepository) Save(_ context.Context, task *entities.Task) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *MockTaskRepository) GetByID(_ context.Context, id int64) (*entities.Task, error) {
	task, exists := m.tasks[id]
	if !exists {
		return nil, nil
	}
	return task, nil
}

func (m *MockTaskRepository) Delete(_ context.Context, id int64) error {
	delete(m.tasks, id)
	return nil
}

func (m *MockTaskRepository) List(_ context.Context, _ repositories.TaskFilter, _, _ int) ([]*entities.Task, error) {
	var tasks []*entities.Task
	for _, t := range m.tasks {
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (m *MockTaskRepository) Count(_ context.Context, _ repositories.TaskFilter) (int64, error) {
	return int64(len(m.tasks)), nil
}

func (m *MockTaskRepository) GetByProject(_ context.Context, _ int64, _, _ int) ([]*entities.Task, error) {
	return nil, nil
}

func (m *MockTaskRepository) GetByAuthor(_ context.Context, _ int64, _, _ int) ([]*entities.Task, error) {
	return nil, nil
}

func (m *MockTaskRepository) GetByAssignee(_ context.Context, _ int64, _, _ int) ([]*entities.Task, error) {
	return nil, nil
}

func (m *MockTaskRepository) GetByStatus(_ context.Context, _ domain.TaskStatus, _, _ int) ([]*entities.Task, error) {
	return nil, nil
}

func (m *MockTaskRepository) GetOverdueTasks(_ context.Context, _, _ int) ([]*entities.Task, error) {
	return nil, nil
}

func (m *MockTaskRepository) AddWatcher(_ context.Context, watcher *entities.TaskWatcher) error {
	m.watchers[watcher.TaskID] = append(m.watchers[watcher.TaskID], watcher)
	return nil
}

func (m *MockTaskRepository) RemoveWatcher(_ context.Context, taskID, userID int64) error {
	watchers := m.watchers[taskID]
	for i, w := range watchers {
		if w.UserID == userID {
			m.watchers[taskID] = append(watchers[:i], watchers[i+1:]...)
			break
		}
	}
	return nil
}

func (m *MockTaskRepository) GetWatchers(_ context.Context, taskID int64) ([]*entities.TaskWatcher, error) {
	return m.watchers[taskID], nil
}

func (m *MockTaskRepository) IsWatching(_ context.Context, taskID, userID int64) (bool, error) {
	for _, w := range m.watchers[taskID] {
		if w.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockTaskRepository) AddAttachment(_ context.Context, _ *entities.TaskAttachment) error {
	return nil
}

func (m *MockTaskRepository) RemoveAttachment(_ context.Context, _ int64) error {
	return nil
}

func (m *MockTaskRepository) GetAttachments(_ context.Context, _ int64) ([]*entities.TaskAttachment, error) {
	return nil, nil
}

func (m *MockTaskRepository) GetAttachmentByID(_ context.Context, _ int64) (*entities.TaskAttachment, error) {
	return nil, nil
}

func (m *MockTaskRepository) AddComment(_ context.Context, comment *entities.TaskComment) error {
	comment.ID = m.nextID
	m.nextID++
	m.comments[comment.ID] = comment
	return nil
}

func (m *MockTaskRepository) UpdateComment(_ context.Context, comment *entities.TaskComment) error {
	m.comments[comment.ID] = comment
	return nil
}

func (m *MockTaskRepository) DeleteComment(_ context.Context, commentID int64) error {
	delete(m.comments, commentID)
	return nil
}

func (m *MockTaskRepository) GetComments(_ context.Context, taskID int64) ([]*entities.TaskComment, error) {
	var comments []*entities.TaskComment
	for _, c := range m.comments {
		if c.TaskID == taskID {
			comments = append(comments, c)
		}
	}
	return comments, nil
}

func (m *MockTaskRepository) GetCommentByID(_ context.Context, commentID int64) (*entities.TaskComment, error) {
	return m.comments[commentID], nil
}

func (m *MockTaskRepository) AddChecklist(_ context.Context, checklist *entities.TaskChecklist) error {
	checklist.ID = m.nextID
	m.nextID++
	m.checklists[checklist.ID] = checklist
	return nil
}

func (m *MockTaskRepository) UpdateChecklist(_ context.Context, checklist *entities.TaskChecklist) error {
	m.checklists[checklist.ID] = checklist
	return nil
}

func (m *MockTaskRepository) DeleteChecklist(_ context.Context, checklistID int64) error {
	delete(m.checklists, checklistID)
	return nil
}

func (m *MockTaskRepository) GetChecklists(_ context.Context, taskID int64) ([]*entities.TaskChecklist, error) {
	var checklists []*entities.TaskChecklist
	for _, c := range m.checklists {
		if c.TaskID == taskID {
			checklists = append(checklists, c)
		}
	}
	return checklists, nil
}

func (m *MockTaskRepository) AddChecklistItem(_ context.Context, item *entities.TaskChecklistItem) error {
	item.ID = m.nextID
	m.nextID++
	m.items[item.ID] = item
	return nil
}

func (m *MockTaskRepository) UpdateChecklistItem(_ context.Context, item *entities.TaskChecklistItem) error {
	m.items[item.ID] = item
	return nil
}

func (m *MockTaskRepository) DeleteChecklistItem(_ context.Context, itemID int64) error {
	delete(m.items, itemID)
	return nil
}

func (m *MockTaskRepository) GetChecklistItems(_ context.Context, checklistID int64) ([]*entities.TaskChecklistItem, error) {
	var items []*entities.TaskChecklistItem
	for _, i := range m.items {
		if i.ChecklistID == checklistID {
			items = append(items, i)
		}
	}
	return items, nil
}

func (m *MockTaskRepository) AddHistory(_ context.Context, history *entities.TaskHistory) error {
	history.ID = m.nextID
	m.nextID++
	m.history[history.TaskID] = append(m.history[history.TaskID], history)
	return nil
}

func (m *MockTaskRepository) GetHistory(_ context.Context, taskID int64, _, _ int) ([]*entities.TaskHistory, error) {
	return m.history[taskID], nil
}

// MockProjectRepository implements ProjectRepository for testing.
type MockProjectRepository struct {
	projects map[int64]*entities.Project
	nextID   int64
}

func NewMockProjectRepository() *MockProjectRepository {
	return &MockProjectRepository{
		projects: make(map[int64]*entities.Project),
		nextID:   1,
	}
}

func (m *MockProjectRepository) Create(_ context.Context, project *entities.Project) error {
	project.ID = m.nextID
	m.nextID++
	m.projects[project.ID] = project
	return nil
}

func (m *MockProjectRepository) Save(_ context.Context, project *entities.Project) error {
	m.projects[project.ID] = project
	return nil
}

func (m *MockProjectRepository) GetByID(_ context.Context, id int64) (*entities.Project, error) {
	return m.projects[id], nil
}

func (m *MockProjectRepository) Delete(_ context.Context, id int64) error {
	delete(m.projects, id)
	return nil
}

func (m *MockProjectRepository) List(_ context.Context, _ repositories.ProjectFilter, _, _ int) ([]*entities.Project, error) {
	var projects []*entities.Project
	for _, p := range m.projects {
		projects = append(projects, p)
	}
	return projects, nil
}

func (m *MockProjectRepository) Count(_ context.Context, _ repositories.ProjectFilter) (int64, error) {
	return int64(len(m.projects)), nil
}

func (m *MockProjectRepository) GetByOwner(_ context.Context, _ int64, _, _ int) ([]*entities.Project, error) {
	return nil, nil
}

func (m *MockProjectRepository) GetByStatus(_ context.Context, _ domain.ProjectStatus, _, _ int) ([]*entities.Project, error) {
	return nil, nil
}

// Tests

func TestTaskUseCase_Create(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	projectRepo := NewMockProjectRepository()
	uc := NewTaskUseCase(taskRepo, projectRepo, nil, nil)

	ctx := context.Background()
	input := dto.CreateTaskInput{
		Title:       "Test Task",
		Description: strPtr("Test Description"),
	}

	task, err := uc.Create(ctx, 1, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if task.ID == 0 {
		t.Error("expected task ID to be set")
	}

	if task.Title != "Test Task" {
		t.Errorf("expected title 'Test Task', got '%s'", task.Title)
	}

	if task.Status != domain.TaskStatusNew {
		t.Errorf("expected status 'new', got '%s'", task.Status)
	}

	if task.AuthorID != 1 {
		t.Errorf("expected author_id 1, got %d", task.AuthorID)
	}
}

func TestTaskUseCase_Create_WithAssignee(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	projectRepo := NewMockProjectRepository()
	uc := NewTaskUseCase(taskRepo, projectRepo, nil, nil)

	ctx := context.Background()
	assigneeID := int64(2)
	input := dto.CreateTaskInput{
		Title:      "Test Task",
		AssigneeID: &assigneeID,
	}

	task, err := uc.Create(ctx, 1, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if task.AssigneeID == nil || *task.AssigneeID != 2 {
		t.Error("expected assignee_id to be 2")
	}

	if task.Status != domain.TaskStatusAssigned {
		t.Errorf("expected status 'assigned', got '%s'", task.Status)
	}
}

func TestTaskUseCase_GetByID(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	projectRepo := NewMockProjectRepository()
	uc := NewTaskUseCase(taskRepo, projectRepo, nil, nil)

	ctx := context.Background()

	// Create a task first
	input := dto.CreateTaskInput{Title: "Test Task"}
	created, _ := uc.Create(ctx, 1, input)

	// Get by ID
	task, err := uc.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if task.ID != created.ID {
		t.Errorf("expected task ID %d, got %d", created.ID, task.ID)
	}
}

func TestTaskUseCase_GetByID_NotFound(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	projectRepo := NewMockProjectRepository()
	uc := NewTaskUseCase(taskRepo, projectRepo, nil, nil)

	ctx := context.Background()

	_, err := uc.GetByID(ctx, 999)
	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskUseCase_Update(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	projectRepo := NewMockProjectRepository()
	uc := NewTaskUseCase(taskRepo, projectRepo, nil, nil)

	ctx := context.Background()

	// Create a task first
	input := dto.CreateTaskInput{Title: "Test Task"}
	created, _ := uc.Create(ctx, 1, input)

	// Update
	newTitle := updatedTaskTitle
	updateInput := dto.UpdateTaskInput{Title: &newTitle}

	updated, err := uc.Update(ctx, 1, created.ID, updateInput)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.Title != updatedTaskTitle {
		t.Errorf("expected title %q, got %q", updatedTaskTitle, updated.Title)
	}
}

func TestTaskUseCase_Delete(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	projectRepo := NewMockProjectRepository()
	uc := NewTaskUseCase(taskRepo, projectRepo, nil, nil)

	ctx := context.Background()

	// Create a task first
	input := dto.CreateTaskInput{Title: "Test Task"}
	created, _ := uc.Create(ctx, 1, input)

	// Delete (by author)
	err := uc.Delete(ctx, 1, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify deleted
	_, err = uc.GetByID(ctx, created.ID)
	if !errors.Is(err, ErrTaskNotFound) {
		t.Error("expected task to be deleted")
	}
}

func TestTaskUseCase_Delete_Unauthorized(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	projectRepo := NewMockProjectRepository()
	uc := NewTaskUseCase(taskRepo, projectRepo, nil, nil)

	ctx := context.Background()

	// Create a task first
	input := dto.CreateTaskInput{Title: "Test Task"}
	created, _ := uc.Create(ctx, 1, input)

	// Try to delete by non-author
	err := uc.Delete(ctx, 2, created.ID)
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestTaskUseCase_List(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	projectRepo := NewMockProjectRepository()
	uc := NewTaskUseCase(taskRepo, projectRepo, nil, nil)

	ctx := context.Background()

	// Create tasks
	_, _ = uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Task 1"})
	_, _ = uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Task 2"})

	// List
	input := dto.TaskFilterInput{Limit: 10}
	output, err := uc.List(ctx, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output.Total != 2 {
		t.Errorf("expected total 2, got %d", output.Total)
	}
}

func TestTaskUseCase_Workflow(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	projectRepo := NewMockProjectRepository()
	uc := NewTaskUseCase(taskRepo, projectRepo, nil, nil)

	ctx := context.Background()

	// Create task
	input := dto.CreateTaskInput{Title: "Test Task"}
	task, _ := uc.Create(ctx, 1, input)

	// Assign
	task, err := uc.Assign(ctx, 1, task.ID, dto.AssignTaskInput{AssigneeID: 2})
	if err != nil {
		t.Fatalf("assign error: %v", err)
	}
	if task.Status != domain.TaskStatusAssigned {
		t.Errorf("expected status 'assigned', got '%s'", task.Status)
	}

	// Start work
	task, err = uc.StartWork(ctx, 2, task.ID)
	if err != nil {
		t.Fatalf("start work error: %v", err)
	}
	if task.Status != domain.TaskStatusInProgress {
		t.Errorf("expected status 'in_progress', got '%s'", task.Status)
	}

	// Submit for review
	task, err = uc.SubmitForReview(ctx, 2, task.ID)
	if err != nil {
		t.Fatalf("submit for review error: %v", err)
	}
	if task.Status != domain.TaskStatusReview {
		t.Errorf("expected status 'review', got '%s'", task.Status)
	}

	// Complete
	task, err = uc.Complete(ctx, 1, task.ID)
	if err != nil {
		t.Fatalf("complete error: %v", err)
	}
	if task.Status != domain.TaskStatusCompleted {
		t.Errorf("expected status 'completed', got '%s'", task.Status)
	}
	if task.Progress != 100 {
		t.Errorf("expected progress 100, got %d", task.Progress)
	}
}

func TestTaskUseCase_Cancel(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	projectRepo := NewMockProjectRepository()
	uc := NewTaskUseCase(taskRepo, projectRepo, nil, nil)

	ctx := context.Background()

	// Create task
	input := dto.CreateTaskInput{Title: "Test Task"}
	task, _ := uc.Create(ctx, 1, input)

	// Cancel
	task, err := uc.Cancel(ctx, 1, task.ID)
	if err != nil {
		t.Fatalf("cancel error: %v", err)
	}
	if task.Status != domain.TaskStatusCancelled {
		t.Errorf("expected status 'canceled', got '%s'", task.Status)
	}
}

func TestTaskUseCase_Reopen(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	projectRepo := NewMockProjectRepository()
	uc := NewTaskUseCase(taskRepo, projectRepo, nil, nil)

	ctx := context.Background()

	// Create and cancel task
	input := dto.CreateTaskInput{Title: "Test Task"}
	task, _ := uc.Create(ctx, 1, input)
	task, _ = uc.Cancel(ctx, 1, task.ID)

	// Reopen
	task, err := uc.Reopen(ctx, 1, task.ID)
	if err != nil {
		t.Fatalf("reopen error: %v", err)
	}
	if task.Status != domain.TaskStatusNew {
		t.Errorf("expected status 'new', got '%s'", task.Status)
	}
}

func TestTaskUseCase_Comments(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	projectRepo := NewMockProjectRepository()
	uc := NewTaskUseCase(taskRepo, projectRepo, nil, nil)

	ctx := context.Background()

	// Create task
	task, _ := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test Task"})

	// Add comment
	comment, err := uc.AddComment(ctx, 1, task.ID, dto.AddCommentInput{Content: "Test comment"})
	if err != nil {
		t.Fatalf("add comment error: %v", err)
	}
	if comment.Content != "Test comment" {
		t.Errorf("expected content 'Test comment', got '%s'", comment.Content)
	}

	// Get comments
	comments, err := uc.GetComments(ctx, task.ID)
	if err != nil {
		t.Fatalf("get comments error: %v", err)
	}
	if len(comments) != 1 {
		t.Errorf("expected 1 comment, got %d", len(comments))
	}

	// Update comment
	updated, err := uc.UpdateComment(ctx, 1, comment.ID, dto.UpdateCommentInput{Content: "Updated comment"})
	if err != nil {
		t.Fatalf("update comment error: %v", err)
	}
	if updated.Content != "Updated comment" {
		t.Errorf("expected content 'Updated comment', got '%s'", updated.Content)
	}

	// Delete comment
	err = uc.DeleteComment(ctx, 1, comment.ID)
	if err != nil {
		t.Fatalf("delete comment error: %v", err)
	}
}

func TestTaskUseCase_Watchers(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	projectRepo := NewMockProjectRepository()
	uc := NewTaskUseCase(taskRepo, projectRepo, nil, nil)

	ctx := context.Background()

	// Create task
	task, _ := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test Task"})

	// Add watcher
	err := uc.AddWatcher(ctx, 1, task.ID, 2)
	if err != nil {
		t.Fatalf("add watcher error: %v", err)
	}

	// Get watchers
	watchers, err := uc.GetWatchers(ctx, task.ID)
	if err != nil {
		t.Fatalf("get watchers error: %v", err)
	}
	if len(watchers) != 1 {
		t.Errorf("expected 1 watcher, got %d", len(watchers))
	}

	// Remove watcher
	err = uc.RemoveWatcher(ctx, 1, task.ID, 2)
	if err != nil {
		t.Fatalf("remove watcher error: %v", err)
	}

	watchers, _ = uc.GetWatchers(ctx, task.ID)
	if len(watchers) != 0 {
		t.Errorf("expected 0 watchers, got %d", len(watchers))
	}
}

func TestTaskUseCase_Checklists(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	projectRepo := NewMockProjectRepository()
	uc := NewTaskUseCase(taskRepo, projectRepo, nil, nil)

	ctx := context.Background()

	// Create task
	task, _ := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test Task"})

	// Add checklist
	checklist, err := uc.AddChecklist(ctx, 1, task.ID, dto.AddChecklistInput{Title: "Test Checklist"})
	if err != nil {
		t.Fatalf("add checklist error: %v", err)
	}
	if checklist.Title != "Test Checklist" {
		t.Errorf("expected title 'Test Checklist', got '%s'", checklist.Title)
	}

	// Add checklist item
	item, err := uc.AddChecklistItem(ctx, 1, checklist.ID, dto.AddChecklistItemInput{Title: "Item 1"})
	if err != nil {
		t.Fatalf("add checklist item error: %v", err)
	}
	if item.Title != "Item 1" {
		t.Errorf("expected title 'Item 1', got '%s'", item.Title)
	}

	// Get checklists
	checklists, err := uc.GetChecklists(ctx, task.ID)
	if err != nil {
		t.Fatalf("get checklists error: %v", err)
	}
	if len(checklists) != 1 {
		t.Errorf("expected 1 checklist, got %d", len(checklists))
	}
}

func TestTaskUseCase_History(t *testing.T) {
	taskRepo := NewMockTaskRepository()
	projectRepo := NewMockProjectRepository()
	uc := NewTaskUseCase(taskRepo, projectRepo, nil, nil)

	ctx := context.Background()

	// Create task
	task, _ := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test Task"})

	// Update to generate history
	newTitle := updatedTaskTitle
	_, _ = uc.Update(ctx, 1, task.ID, dto.UpdateTaskInput{Title: &newTitle})

	// Get history
	history, err := uc.GetHistory(ctx, task.ID, 10, 0)
	if err != nil {
		t.Fatalf("get history error: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(history))
	}
}

// Test entity methods
func TestTask_IsOverdue(t *testing.T) {
	task := entities.NewTask("Test", 1)

	// No due date - not overdue
	if task.IsOverdue() {
		t.Error("task without due date should not be overdue")
	}

	// Future due date - not overdue
	future := time.Now().Add(24 * time.Hour)
	task.SetDueDate(&future)
	if task.IsOverdue() {
		t.Error("task with future due date should not be overdue")
	}

	// Past due date - overdue
	past := time.Now().Add(-24 * time.Hour)
	task.SetDueDate(&past)
	if !task.IsOverdue() {
		t.Error("task with past due date should be overdue")
	}

	// Completed task - not overdue
	task.Status = domain.TaskStatusCompleted
	if task.IsOverdue() {
		t.Error("completed task should not be overdue")
	}
}

func TestTask_CanEdit(t *testing.T) {
	task := entities.NewTask("Test", 1)

	if !task.CanEdit() {
		t.Error("new task should be editable")
	}

	task.Status = domain.TaskStatusCompleted
	if task.CanEdit() {
		t.Error("completed task should not be editable")
	}

	task.Status = domain.TaskStatusCancelled
	if task.CanEdit() {
		t.Error("canceled task should not be editable")
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}
