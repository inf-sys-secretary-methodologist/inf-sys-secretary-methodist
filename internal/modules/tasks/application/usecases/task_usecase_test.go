package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

const (
	updatedTaskTitle = "Updated Task"
	testNewTitle     = "New Title"
	testUpdated      = "Updated"
)

// --- Error-returning mock task repository ---

type ErrorMockTaskRepository struct {
	MockTaskRepository
	createErr              error
	saveErr                error
	getByIDErr             error
	deleteErr              error
	listErr                error
	countErr               error
	addWatcherErr          error
	removeWatcherErr       error
	isWatchingErr          error
	addCommentErr          error
	getCommentByIDErr      error
	updateCommentErr       error
	deleteCommentErr       error
	addChecklistErr        error
	deleteChecklistErr     error
	getChecklistsErr       error
	getChecklistItemErr    error
	addChecklistItemErr    error
	deleteChecklistItemErr error
}

func (m *ErrorMockTaskRepository) Create(ctx context.Context, task *entities.Task) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.MockTaskRepository.Create(ctx, task)
}

func (m *ErrorMockTaskRepository) Save(ctx context.Context, task *entities.Task) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	return m.MockTaskRepository.Save(ctx, task)
}

func (m *ErrorMockTaskRepository) GetByID(ctx context.Context, id int64) (*entities.Task, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.MockTaskRepository.GetByID(ctx, id)
}

func (m *ErrorMockTaskRepository) Delete(ctx context.Context, id int64) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return m.MockTaskRepository.Delete(ctx, id)
}

func (m *ErrorMockTaskRepository) List(ctx context.Context, f repositories.TaskFilter, limit, offset int) ([]*entities.Task, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.MockTaskRepository.List(ctx, f, limit, offset)
}

func (m *ErrorMockTaskRepository) Count(ctx context.Context, f repositories.TaskFilter) (int64, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	return m.MockTaskRepository.Count(ctx, f)
}

func (m *ErrorMockTaskRepository) AddWatcher(ctx context.Context, w *entities.TaskWatcher) error {
	if m.addWatcherErr != nil {
		return m.addWatcherErr
	}
	return m.MockTaskRepository.AddWatcher(ctx, w)
}

func (m *ErrorMockTaskRepository) RemoveWatcher(ctx context.Context, taskID, userID int64) error {
	if m.removeWatcherErr != nil {
		return m.removeWatcherErr
	}
	return m.MockTaskRepository.RemoveWatcher(ctx, taskID, userID)
}

func (m *ErrorMockTaskRepository) IsWatching(ctx context.Context, taskID, userID int64) (bool, error) {
	if m.isWatchingErr != nil {
		return false, m.isWatchingErr
	}
	return m.MockTaskRepository.IsWatching(ctx, taskID, userID)
}

func (m *ErrorMockTaskRepository) AddComment(ctx context.Context, c *entities.TaskComment) error {
	if m.addCommentErr != nil {
		return m.addCommentErr
	}
	return m.MockTaskRepository.AddComment(ctx, c)
}

func (m *ErrorMockTaskRepository) GetCommentByID(ctx context.Context, commentID int64) (*entities.TaskComment, error) {
	if m.getCommentByIDErr != nil {
		return nil, m.getCommentByIDErr
	}
	return m.MockTaskRepository.GetCommentByID(ctx, commentID)
}

func (m *ErrorMockTaskRepository) UpdateComment(ctx context.Context, c *entities.TaskComment) error {
	if m.updateCommentErr != nil {
		return m.updateCommentErr
	}
	return m.MockTaskRepository.UpdateComment(ctx, c)
}

func (m *ErrorMockTaskRepository) DeleteComment(ctx context.Context, commentID int64) error {
	if m.deleteCommentErr != nil {
		return m.deleteCommentErr
	}
	return m.MockTaskRepository.DeleteComment(ctx, commentID)
}

func (m *ErrorMockTaskRepository) AddChecklist(ctx context.Context, c *entities.TaskChecklist) error {
	if m.addChecklistErr != nil {
		return m.addChecklistErr
	}
	return m.MockTaskRepository.AddChecklist(ctx, c)
}

func (m *ErrorMockTaskRepository) DeleteChecklist(ctx context.Context, id int64) error {
	if m.deleteChecklistErr != nil {
		return m.deleteChecklistErr
	}
	return m.MockTaskRepository.DeleteChecklist(ctx, id)
}

func (m *ErrorMockTaskRepository) GetChecklists(ctx context.Context, taskID int64) ([]*entities.TaskChecklist, error) {
	if m.getChecklistsErr != nil {
		return nil, m.getChecklistsErr
	}
	return m.MockTaskRepository.GetChecklists(ctx, taskID)
}

func (m *ErrorMockTaskRepository) GetChecklistItems(ctx context.Context, checklistID int64) ([]*entities.TaskChecklistItem, error) {
	if m.getChecklistItemErr != nil {
		return nil, m.getChecklistItemErr
	}
	return m.MockTaskRepository.GetChecklistItems(ctx, checklistID)
}

func (m *ErrorMockTaskRepository) AddChecklistItem(ctx context.Context, item *entities.TaskChecklistItem) error {
	if m.addChecklistItemErr != nil {
		return m.addChecklistItemErr
	}
	return m.MockTaskRepository.AddChecklistItem(ctx, item)
}

func (m *ErrorMockTaskRepository) DeleteChecklistItem(ctx context.Context, itemID int64) error {
	if m.deleteChecklistItemErr != nil {
		return m.deleteChecklistItemErr
	}
	return m.MockTaskRepository.DeleteChecklistItem(ctx, itemID)
}

func newErrorMockTaskRepo() *ErrorMockTaskRepository {
	return &ErrorMockTaskRepository{
		MockTaskRepository: *NewMockTaskRepository(),
	}
}

// --- Helpers ---

func createTestAuditLogger() *logging.AuditLogger {
	logger := logging.NewLogger("error")
	return logging.NewAuditLogger(logger)
}

func strPtr(s string) *string       { return &s }
func intPtr(i int) *int             { return &i }
func int64Ptr(i int64) *int64       { return &i }
func float64Ptr(f float64) *float64 { return &f }

func setupTaskUseCase() (*TaskUseCase, *MockTaskRepository) {
	taskRepo := NewMockTaskRepository()
	projectRepo := NewMockProjectRepository()
	uc := NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	return uc, taskRepo
}

func setupTaskUseCaseWithAudit() (*TaskUseCase, *MockTaskRepository) {
	taskRepo := NewMockTaskRepository()
	projectRepo := NewMockProjectRepository()
	audit := createTestAuditLogger()
	uc := NewTaskUseCase(taskRepo, projectRepo, audit, nil)
	return uc, taskRepo
}

func createTask(t *testing.T, uc *TaskUseCase, title string, authorID int64) *entities.Task {
	t.Helper()
	task, err := uc.Create(context.Background(), authorID, dto.CreateTaskInput{Title: title})
	require.NoError(t, err)
	return task
}

// ===================== Create =====================

func TestTaskUseCase_Create(t *testing.T) {
	uc, _ := setupTaskUseCase()
	ctx := context.Background()

	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{
		Title:       "Test Task",
		Description: strPtr("Test Description"),
	})

	require.NoError(t, err)
	assert.NotZero(t, task.ID)
	assert.Equal(t, "Test Task", task.Title)
	assert.Equal(t, domain.TaskStatusNew, task.Status)
	assert.Equal(t, int64(1), task.AuthorID)
}

func TestTaskUseCase_Create_WithAssignee(t *testing.T) {
	uc, _ := setupTaskUseCase()
	ctx := context.Background()

	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{
		Title:      "Test Task",
		AssigneeID: int64Ptr(2),
	})

	require.NoError(t, err)
	require.NotNil(t, task.AssigneeID)
	assert.Equal(t, int64(2), *task.AssigneeID)
	assert.Equal(t, domain.TaskStatusAssigned, task.Status)
}

func TestTaskUseCase_Create_WithPriority(t *testing.T) {
	uc, _ := setupTaskUseCase()
	ctx := context.Background()

	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{
		Title:    "Test Task",
		Priority: strPtr("high"),
	})

	require.NoError(t, err)
	assert.Equal(t, domain.TaskPriorityHigh, task.Priority)
}

func TestTaskUseCase_Create_WithInvalidPriority(t *testing.T) {
	uc, _ := setupTaskUseCase()
	ctx := context.Background()

	_, err := uc.Create(ctx, 1, dto.CreateTaskInput{
		Title:    "Test Task",
		Priority: strPtr("invalid"),
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestTaskUseCase_Create_WithMetadata(t *testing.T) {
	uc, _ := setupTaskUseCase()
	ctx := context.Background()

	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{
		Title:    "Test Task",
		Metadata: map[string]any{"key": "value"},
	})

	require.NoError(t, err)
	assert.NotNil(t, task.Metadata)
}

func TestTaskUseCase_Create_WithAllFields(t *testing.T) {
	uc, _ := setupTaskUseCase()
	ctx := context.Background()
	now := time.Now()

	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{
		Title:          "Full Task",
		Description:    strPtr("desc"),
		ProjectID:      int64Ptr(10),
		DocumentID:     int64Ptr(20),
		DueDate:        &now,
		StartDate:      &now,
		EstimatedHours: float64Ptr(5.0),
		Tags:           []string{"tag1", "tag2"},
	})

	require.NoError(t, err)
	assert.Equal(t, "Full Task", task.Title)
	assert.Equal(t, int64Ptr(10), task.ProjectID)
	assert.Equal(t, int64Ptr(20), task.DocumentID)
	assert.NotNil(t, task.DueDate)
	assert.NotNil(t, task.StartDate)
	assert.Equal(t, float64Ptr(5.0), task.EstimatedHours)
	assert.Equal(t, []string{"tag1", "tag2"}, task.Tags)
}

func TestTaskUseCase_Create_RepoError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	repo.createErr = errors.New("db error")
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)

	_, err := uc.Create(context.Background(), 1, dto.CreateTaskInput{Title: "Test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create task")
}

func TestTaskUseCase_Create_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()

	task, err := uc.Create(context.Background(), 1, dto.CreateTaskInput{Title: "Test"})
	require.NoError(t, err)
	assert.NotZero(t, task.ID)
}

// ===================== GetByID =====================

func TestTaskUseCase_GetByID(t *testing.T) {
	uc, _ := setupTaskUseCase()
	created := createTask(t, uc, "Test Task", 1)

	task, err := uc.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, task.ID)
}

func TestTaskUseCase_GetByID_NotFound(t *testing.T) {
	uc, _ := setupTaskUseCase()

	_, err := uc.GetByID(context.Background(), 999)
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskUseCase_GetByID_RepoError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	repo.getByIDErr = errors.New("db error")
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)

	_, err := uc.GetByID(context.Background(), 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get task")
}

// ===================== Update =====================

func TestTaskUseCase_Update(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test Task", 1)

	newTitle := updatedTaskTitle
	updated, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{Title: &newTitle})

	require.NoError(t, err)
	assert.Equal(t, updatedTaskTitle, updated.Title)
}

func TestTaskUseCase_Update_NotFound(t *testing.T) {
	uc, _ := setupTaskUseCase()

	newTitle := testNewTitle
	_, err := uc.Update(context.Background(), 1, 999, dto.UpdateTaskInput{Title: &newTitle})
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskUseCase_Update_CannotModifyCompleted(t *testing.T) {
	uc, taskRepo := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	task.Status = domain.TaskStatusCompleted
	taskRepo.tasks[task.ID] = task

	newTitle := testNewTitle
	_, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{Title: &newTitle})
	assert.ErrorIs(t, err, ErrCannotModifyTask)
}

func TestTaskUseCase_Update_CannotModifyCancelled(t *testing.T) {
	uc, taskRepo := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	task.Status = domain.TaskStatusCancelled
	taskRepo.tasks[task.ID] = task

	newTitle := testNewTitle
	_, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{Title: &newTitle})
	assert.ErrorIs(t, err, ErrCannotModifyTask)
}

func TestTaskUseCase_Update_Description(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	// Update with nil description initially
	newDesc := "new description"
	updated, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{Description: &newDesc})
	require.NoError(t, err)
	require.NotNil(t, updated.Description)
	assert.Equal(t, "new description", *updated.Description)

	// Update again with existing description
	newDesc2 := "newer description"
	updated2, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{Description: &newDesc2})
	require.NoError(t, err)
	assert.Equal(t, "newer description", *updated2.Description)
}

func TestTaskUseCase_Update_ProjectID(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	updated, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{ProjectID: int64Ptr(5)})
	require.NoError(t, err)
	require.NotNil(t, updated.ProjectID)
	assert.Equal(t, int64(5), *updated.ProjectID)
}

func TestTaskUseCase_Update_Priority(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	updated, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{Priority: strPtr("urgent")})
	require.NoError(t, err)
	assert.Equal(t, domain.TaskPriorityUrgent, updated.Priority)
}

func TestTaskUseCase_Update_InvalidPriority(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	_, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{Priority: strPtr("invalid")})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestTaskUseCase_Update_DueDate(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	due := time.Now().Add(48 * time.Hour)

	updated, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{DueDate: &due})
	require.NoError(t, err)
	assert.NotNil(t, updated.DueDate)
}

func TestTaskUseCase_Update_StartDate(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	start := time.Now()

	updated, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{StartDate: &start})
	require.NoError(t, err)
	assert.NotNil(t, updated.StartDate)
}

func TestTaskUseCase_Update_Progress(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	updated, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{Progress: intPtr(50)})
	require.NoError(t, err)
	assert.Equal(t, 50, updated.Progress)
}

func TestTaskUseCase_Update_Progress_OnCompletedTask(t *testing.T) {
	uc, taskRepo := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	task.Status = domain.TaskStatusCompleted
	taskRepo.tasks[task.ID] = task

	_, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{Progress: intPtr(50)})
	assert.ErrorIs(t, err, ErrCannotModifyTask)
}

func TestTaskUseCase_Update_EstimatedHours(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	updated, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{EstimatedHours: float64Ptr(10.5)})
	require.NoError(t, err)
	require.NotNil(t, updated.EstimatedHours)
	assert.InDelta(t, 10.5, *updated.EstimatedHours, 0.01)
}

func TestTaskUseCase_Update_ActualHours(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	updated, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{ActualHours: float64Ptr(8.0)})
	require.NoError(t, err)
	require.NotNil(t, updated.ActualHours)
	assert.InDelta(t, 8.0, *updated.ActualHours, 0.01)
}

func TestTaskUseCase_Update_Tags(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	updated, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{Tags: []string{"go", "test"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"go", "test"}, updated.Tags)
}

func TestTaskUseCase_Update_Metadata(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	updated, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{Metadata: map[string]any{"key": "val"}})
	require.NoError(t, err)
	assert.NotNil(t, updated.Metadata)
}

func TestTaskUseCase_Update_TitleUnchanged(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	// Updating with same title should not record change
	sameTitle := "Test"
	updated, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{Title: &sameTitle})
	require.NoError(t, err)
	assert.Equal(t, "Test", updated.Title)
}

func TestTaskUseCase_Update_SaveError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})
	require.NoError(t, err)

	repo.saveErr = errors.New("save error")
	newTitle := testUpdated
	_, err = uc.Update(ctx, 1, task.ID, dto.UpdateTaskInput{Title: &newTitle})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update task")
}

func TestTaskUseCase_Update_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)

	newTitle := testUpdated
	updated, err := uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{Title: &newTitle})
	require.NoError(t, err)
	assert.Equal(t, testUpdated, updated.Title)
}

// ===================== Delete =====================

func TestTaskUseCase_Delete(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	err := uc.Delete(context.Background(), 1, task.ID)
	require.NoError(t, err)

	_, err = uc.GetByID(context.Background(), task.ID)
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskUseCase_Delete_Unauthorized(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	err := uc.Delete(context.Background(), 2, task.ID)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestTaskUseCase_Delete_NotFound(t *testing.T) {
	uc, _ := setupTaskUseCase()

	err := uc.Delete(context.Background(), 1, 999)
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskUseCase_Delete_RepoError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})
	require.NoError(t, err)

	repo.deleteErr = errors.New("delete error")
	err = uc.Delete(ctx, 1, task.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete task")
}

func TestTaskUseCase_Delete_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)

	err := uc.Delete(context.Background(), 1, task.ID)
	require.NoError(t, err)
}

// ===================== List =====================

func TestTaskUseCase_List(t *testing.T) {
	uc, _ := setupTaskUseCase()
	ctx := context.Background()

	createTask(t, uc, "Task 1", 1)
	createTask(t, uc, "Task 2", 1)

	output, err := uc.List(ctx, dto.TaskFilterInput{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(2), output.Total)
	assert.Len(t, output.Tasks, 2)
}

func TestTaskUseCase_List_ListError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	repo.listErr = errors.New("list error")
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)

	_, err := uc.List(context.Background(), dto.TaskFilterInput{Limit: 10})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list tasks")
}

func TestTaskUseCase_List_CountError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	repo.countErr = errors.New("count error")
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)

	_, err := uc.List(context.Background(), dto.TaskFilterInput{Limit: 10})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to count tasks")
}

func TestTaskUseCase_List_Empty(t *testing.T) {
	uc, _ := setupTaskUseCase()

	output, err := uc.List(context.Background(), dto.TaskFilterInput{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(0), output.Total)
	assert.Empty(t, output.Tasks)
}

// ===================== Assign =====================

func TestTaskUseCase_Assign(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	result, err := uc.Assign(context.Background(), 1, task.ID, dto.AssignTaskInput{AssigneeID: 2})
	require.NoError(t, err)
	require.NotNil(t, result.AssigneeID)
	assert.Equal(t, int64(2), *result.AssigneeID)
	assert.Equal(t, domain.TaskStatusAssigned, result.Status)
}

func TestTaskUseCase_Assign_NotFound(t *testing.T) {
	uc, _ := setupTaskUseCase()

	_, err := uc.Assign(context.Background(), 1, 999, dto.AssignTaskInput{AssigneeID: 2})
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskUseCase_Assign_Reassign(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	// First assign
	task, err := uc.Assign(context.Background(), 1, task.ID, dto.AssignTaskInput{AssigneeID: 2})
	require.NoError(t, err)

	// Reassign - old assignee is non-nil
	task, err = uc.Assign(context.Background(), 1, task.ID, dto.AssignTaskInput{AssigneeID: 3})
	require.NoError(t, err)
	assert.Equal(t, int64(3), *task.AssigneeID)
}

func TestTaskUseCase_Assign_SaveError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})
	require.NoError(t, err)

	repo.saveErr = errors.New("save error")
	_, err = uc.Assign(ctx, 1, task.ID, dto.AssignTaskInput{AssigneeID: 2})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to assign task")
}

func TestTaskUseCase_Assign_CompletedTask(t *testing.T) {
	uc, taskRepo := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	task.Status = domain.TaskStatusCompleted
	taskRepo.tasks[task.ID] = task

	_, err := uc.Assign(context.Background(), 1, task.ID, dto.AssignTaskInput{AssigneeID: 2})
	require.Error(t, err)
}

func TestTaskUseCase_Assign_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)

	result, err := uc.Assign(context.Background(), 1, task.ID, dto.AssignTaskInput{AssigneeID: 2})
	require.NoError(t, err)
	assert.Equal(t, int64(2), *result.AssigneeID)
}

// ===================== Unassign =====================

func TestTaskUseCase_Unassign(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	// Assign first
	task, err := uc.Assign(context.Background(), 1, task.ID, dto.AssignTaskInput{AssigneeID: 2})
	require.NoError(t, err)
	assert.Equal(t, domain.TaskStatusAssigned, task.Status)

	// Unassign
	result, err := uc.Unassign(context.Background(), 1, task.ID)
	require.NoError(t, err)
	assert.Nil(t, result.AssigneeID)
	assert.Equal(t, domain.TaskStatusNew, result.Status)
}

func TestTaskUseCase_Unassign_NotFound(t *testing.T) {
	uc, _ := setupTaskUseCase()

	_, err := uc.Unassign(context.Background(), 1, 999)
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskUseCase_Unassign_CompletedTask(t *testing.T) {
	uc, taskRepo := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	task.Status = domain.TaskStatusCompleted
	taskRepo.tasks[task.ID] = task

	_, err := uc.Unassign(context.Background(), 1, task.ID)
	require.Error(t, err)
}

func TestTaskUseCase_Unassign_CancelledTask(t *testing.T) {
	uc, taskRepo := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	task.Status = domain.TaskStatusCancelled
	taskRepo.tasks[task.ID] = task

	_, err := uc.Unassign(context.Background(), 1, task.ID)
	require.Error(t, err)
}

func TestTaskUseCase_Unassign_SaveError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})
	require.NoError(t, err)

	// Assign first
	task, err = uc.Assign(ctx, 1, task.ID, dto.AssignTaskInput{AssigneeID: 2})
	require.NoError(t, err)

	repo.saveErr = errors.New("save error")
	_, err = uc.Unassign(ctx, 1, task.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unassign task")
}

func TestTaskUseCase_Unassign_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)
	task, _ = uc.Assign(context.Background(), 1, task.ID, dto.AssignTaskInput{AssigneeID: 2})

	result, err := uc.Unassign(context.Background(), 1, task.ID)
	require.NoError(t, err)
	assert.Nil(t, result.AssigneeID)
}

// ===================== StartWork =====================

func TestTaskUseCase_StartWork(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	result, err := uc.StartWork(context.Background(), 1, task.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.TaskStatusInProgress, result.Status)
}

func TestTaskUseCase_StartWork_NotFound(t *testing.T) {
	uc, _ := setupTaskUseCase()

	_, err := uc.StartWork(context.Background(), 1, 999)
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskUseCase_StartWork_InvalidTransition(t *testing.T) {
	uc, taskRepo := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	task.Status = domain.TaskStatusReview
	taskRepo.tasks[task.ID] = task

	_, err := uc.StartWork(context.Background(), 1, task.ID)
	require.Error(t, err)
}

func TestTaskUseCase_StartWork_SaveError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})
	require.NoError(t, err)

	repo.saveErr = errors.New("save error")
	_, err = uc.StartWork(ctx, 1, task.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start work")
}

func TestTaskUseCase_StartWork_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)

	result, err := uc.StartWork(context.Background(), 1, task.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.TaskStatusInProgress, result.Status)
}

// ===================== SubmitForReview =====================

func TestTaskUseCase_SubmitForReview(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	task, err := uc.StartWork(context.Background(), 1, task.ID)
	require.NoError(t, err)

	result, err := uc.SubmitForReview(context.Background(), 1, task.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.TaskStatusReview, result.Status)
}

func TestTaskUseCase_SubmitForReview_NotFound(t *testing.T) {
	uc, _ := setupTaskUseCase()

	_, err := uc.SubmitForReview(context.Background(), 1, 999)
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskUseCase_SubmitForReview_InvalidTransition(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1) // status is "new"

	_, err := uc.SubmitForReview(context.Background(), 1, task.ID)
	require.Error(t, err)
}

func TestTaskUseCase_SubmitForReview_SaveError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})
	require.NoError(t, err)
	task, err = uc.StartWork(ctx, 1, task.ID)
	require.NoError(t, err)

	repo.saveErr = errors.New("save error")
	_, err = uc.SubmitForReview(ctx, 1, task.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to submit for review")
}

func TestTaskUseCase_SubmitForReview_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)
	task, _ = uc.StartWork(context.Background(), 1, task.ID)

	result, err := uc.SubmitForReview(context.Background(), 2, task.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.TaskStatusReview, result.Status)
}

// ===================== Complete =====================

func TestTaskUseCase_Complete(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	task, _ = uc.StartWork(context.Background(), 1, task.ID)

	result, err := uc.Complete(context.Background(), 1, task.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.TaskStatusCompleted, result.Status)
	assert.Equal(t, 100, result.Progress)
}

func TestTaskUseCase_Complete_NotFound(t *testing.T) {
	uc, _ := setupTaskUseCase()

	_, err := uc.Complete(context.Background(), 1, 999)
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskUseCase_Complete_InvalidTransition(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	_, err := uc.Complete(context.Background(), 1, task.ID)
	require.Error(t, err)
}

func TestTaskUseCase_Complete_SaveError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})
	require.NoError(t, err)
	task, err = uc.StartWork(ctx, 1, task.ID)
	require.NoError(t, err)

	repo.saveErr = errors.New("save error")
	_, err = uc.Complete(ctx, 1, task.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to complete task")
}

func TestTaskUseCase_Complete_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)
	task, _ = uc.StartWork(context.Background(), 1, task.ID)

	// Complete by different user to hit notification path
	result, err := uc.Complete(context.Background(), 2, task.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.TaskStatusCompleted, result.Status)
}

// ===================== Cancel =====================

func TestTaskUseCase_Cancel(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	result, err := uc.Cancel(context.Background(), 1, task.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.TaskStatusCancelled, result.Status)
}

func TestTaskUseCase_Cancel_NotFound(t *testing.T) {
	uc, _ := setupTaskUseCase()

	_, err := uc.Cancel(context.Background(), 1, 999)
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskUseCase_Cancel_AlreadyCompleted(t *testing.T) {
	uc, taskRepo := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	task.Status = domain.TaskStatusCompleted
	taskRepo.tasks[task.ID] = task

	_, err := uc.Cancel(context.Background(), 1, task.ID)
	require.Error(t, err)
}

func TestTaskUseCase_Cancel_SaveError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})
	require.NoError(t, err)

	repo.saveErr = errors.New("save error")
	_, err = uc.Cancel(ctx, 1, task.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to cancel task")
}

func TestTaskUseCase_Cancel_WithAssignee(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)
	task, _ = uc.Assign(context.Background(), 1, task.ID, dto.AssignTaskInput{AssigneeID: 2})

	// Cancel by different user than assignee
	result, err := uc.Cancel(context.Background(), 1, task.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.TaskStatusCancelled, result.Status)
}

// ===================== Reopen =====================

func TestTaskUseCase_Reopen(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	task, err := uc.Cancel(context.Background(), 1, task.ID)
	require.NoError(t, err)

	result, err := uc.Reopen(context.Background(), 1, task.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.TaskStatusNew, result.Status)
}

func TestTaskUseCase_Reopen_NotFound(t *testing.T) {
	uc, _ := setupTaskUseCase()

	_, err := uc.Reopen(context.Background(), 1, 999)
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskUseCase_Reopen_InvalidTransition(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1) // status is "new"

	_, err := uc.Reopen(context.Background(), 1, task.ID)
	require.Error(t, err)
}

func TestTaskUseCase_Reopen_SaveError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})
	require.NoError(t, err)
	task, err = uc.Cancel(ctx, 1, task.ID)
	require.NoError(t, err)

	repo.saveErr = errors.New("save error")
	_, err = uc.Reopen(ctx, 1, task.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to reopen task")
}

func TestTaskUseCase_Reopen_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)
	task, _ = uc.Cancel(context.Background(), 1, task.ID)

	result, err := uc.Reopen(context.Background(), 1, task.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.TaskStatusNew, result.Status)
}

// ===================== Watchers =====================

func TestTaskUseCase_AddWatcher(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	err := uc.AddWatcher(context.Background(), 1, task.ID, 2)
	require.NoError(t, err)

	watchers, err := uc.GetWatchers(context.Background(), task.ID)
	require.NoError(t, err)
	assert.Len(t, watchers, 1)
	assert.Equal(t, int64(2), watchers[0].UserID)
}

func TestTaskUseCase_AddWatcher_AlreadyWatching(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	err := uc.AddWatcher(context.Background(), 1, task.ID, 2)
	require.NoError(t, err)

	// Add same watcher again - should be no-op
	err = uc.AddWatcher(context.Background(), 1, task.ID, 2)
	require.NoError(t, err)

	watchers, _ := uc.GetWatchers(context.Background(), task.ID)
	assert.Len(t, watchers, 1)
}

func TestTaskUseCase_AddWatcher_NotFound(t *testing.T) {
	uc, _ := setupTaskUseCase()

	err := uc.AddWatcher(context.Background(), 1, 999, 2)
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskUseCase_AddWatcher_IsWatchingError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})
	require.NoError(t, err)

	repo.isWatchingErr = errors.New("check error")
	err = uc.AddWatcher(ctx, 1, task.ID, 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check watcher")
}

func TestTaskUseCase_AddWatcher_AddError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})
	require.NoError(t, err)

	repo.addWatcherErr = errors.New("add error")
	err = uc.AddWatcher(ctx, 1, task.ID, 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add watcher")
}

func TestTaskUseCase_AddWatcher_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)

	err := uc.AddWatcher(context.Background(), 1, task.ID, 2)
	require.NoError(t, err)
}

func TestTaskUseCase_RemoveWatcher(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	_ = uc.AddWatcher(context.Background(), 1, task.ID, 2)
	err := uc.RemoveWatcher(context.Background(), 1, task.ID, 2)
	require.NoError(t, err)

	watchers, _ := uc.GetWatchers(context.Background(), task.ID)
	assert.Len(t, watchers, 0)
}

func TestTaskUseCase_RemoveWatcher_Error(t *testing.T) {
	repo := newErrorMockTaskRepo()
	repo.removeWatcherErr = errors.New("remove error")
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)

	err := uc.RemoveWatcher(context.Background(), 1, 1, 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove watcher")
}

func TestTaskUseCase_RemoveWatcher_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)
	_ = uc.AddWatcher(context.Background(), 1, task.ID, 2)

	err := uc.RemoveWatcher(context.Background(), 1, task.ID, 2)
	require.NoError(t, err)
}

func TestTaskUseCase_GetWatchers(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	watchers, err := uc.GetWatchers(context.Background(), task.ID)
	require.NoError(t, err)
	assert.Empty(t, watchers)
}

// ===================== Comments =====================

func TestTaskUseCase_AddComment(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	comment, err := uc.AddComment(context.Background(), 1, task.ID, dto.AddCommentInput{Content: "Hello"})
	require.NoError(t, err)
	assert.Equal(t, "Hello", comment.Content)
	assert.Equal(t, task.ID, comment.TaskID)
	assert.Equal(t, int64(1), comment.AuthorID)
}

func TestTaskUseCase_AddComment_WithParent(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	parent, err := uc.AddComment(context.Background(), 1, task.ID, dto.AddCommentInput{Content: "Parent"})
	require.NoError(t, err)

	child, err := uc.AddComment(context.Background(), 2, task.ID, dto.AddCommentInput{
		Content:         "Reply",
		ParentCommentID: &parent.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, child.ParentCommentID)
	assert.Equal(t, parent.ID, *child.ParentCommentID)
}

func TestTaskUseCase_AddComment_NotFound(t *testing.T) {
	uc, _ := setupTaskUseCase()

	_, err := uc.AddComment(context.Background(), 1, 999, dto.AddCommentInput{Content: "Hello"})
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskUseCase_AddComment_RepoError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})
	require.NoError(t, err)

	repo.addCommentErr = errors.New("comment error")
	_, err = uc.AddComment(ctx, 1, task.ID, dto.AddCommentInput{Content: "Hello"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add comment")
}

func TestTaskUseCase_AddComment_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)

	comment, err := uc.AddComment(context.Background(), 1, task.ID, dto.AddCommentInput{Content: "Hello"})
	require.NoError(t, err)
	assert.Equal(t, "Hello", comment.Content)
}

func TestTaskUseCase_UpdateComment(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	comment, _ := uc.AddComment(context.Background(), 1, task.ID, dto.AddCommentInput{Content: "Hello"})

	updated, err := uc.UpdateComment(context.Background(), 1, comment.ID, dto.UpdateCommentInput{Content: testUpdated})
	require.NoError(t, err)
	assert.Equal(t, testUpdated, updated.Content)
}

func TestTaskUseCase_UpdateComment_NotFound(t *testing.T) {
	uc, _ := setupTaskUseCase()

	_, err := uc.UpdateComment(context.Background(), 1, 999, dto.UpdateCommentInput{Content: testUpdated})
	assert.ErrorIs(t, err, ErrCommentNotFound)
}

func TestTaskUseCase_UpdateComment_Unauthorized(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	comment, _ := uc.AddComment(context.Background(), 1, task.ID, dto.AddCommentInput{Content: "Hello"})

	_, err := uc.UpdateComment(context.Background(), 2, comment.ID, dto.UpdateCommentInput{Content: testUpdated})
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestTaskUseCase_UpdateComment_GetError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	repo.getCommentByIDErr = errors.New("get error")
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)

	_, err := uc.UpdateComment(context.Background(), 1, 1, dto.UpdateCommentInput{Content: testUpdated})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get comment")
}

func TestTaskUseCase_UpdateComment_SaveError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, _ := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})
	comment, _ := uc.AddComment(ctx, 1, task.ID, dto.AddCommentInput{Content: "Hello"})

	repo.updateCommentErr = errors.New("update error")
	_, err := uc.UpdateComment(ctx, 1, comment.ID, dto.UpdateCommentInput{Content: testUpdated})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update comment")
}

func TestTaskUseCase_UpdateComment_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)
	comment, _ := uc.AddComment(context.Background(), 1, task.ID, dto.AddCommentInput{Content: "Hello"})

	updated, err := uc.UpdateComment(context.Background(), 1, comment.ID, dto.UpdateCommentInput{Content: testUpdated})
	require.NoError(t, err)
	assert.Equal(t, testUpdated, updated.Content)
}

func TestTaskUseCase_DeleteComment(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	comment, _ := uc.AddComment(context.Background(), 1, task.ID, dto.AddCommentInput{Content: "Hello"})

	err := uc.DeleteComment(context.Background(), 1, comment.ID)
	require.NoError(t, err)
}

func TestTaskUseCase_DeleteComment_NotFound(t *testing.T) {
	uc, _ := setupTaskUseCase()

	err := uc.DeleteComment(context.Background(), 1, 999)
	assert.ErrorIs(t, err, ErrCommentNotFound)
}

func TestTaskUseCase_DeleteComment_Unauthorized(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	comment, _ := uc.AddComment(context.Background(), 1, task.ID, dto.AddCommentInput{Content: "Hello"})

	err := uc.DeleteComment(context.Background(), 2, comment.ID)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestTaskUseCase_DeleteComment_GetError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	repo.getCommentByIDErr = errors.New("get error")
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)

	err := uc.DeleteComment(context.Background(), 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get comment")
}

func TestTaskUseCase_DeleteComment_RepoError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, _ := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})
	comment, _ := uc.AddComment(ctx, 1, task.ID, dto.AddCommentInput{Content: "Hello"})

	repo.deleteCommentErr = errors.New("delete error")
	err := uc.DeleteComment(ctx, 1, comment.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete comment")
}

func TestTaskUseCase_DeleteComment_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)
	comment, _ := uc.AddComment(context.Background(), 1, task.ID, dto.AddCommentInput{Content: "Hello"})

	err := uc.DeleteComment(context.Background(), 1, comment.ID)
	require.NoError(t, err)
}

func TestTaskUseCase_GetComments(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	ctx := context.Background()

	_, _ = uc.AddComment(ctx, 1, task.ID, dto.AddCommentInput{Content: "Comment 1"})
	_, _ = uc.AddComment(ctx, 1, task.ID, dto.AddCommentInput{Content: "Comment 2"})

	comments, err := uc.GetComments(ctx, task.ID)
	require.NoError(t, err)
	assert.Len(t, comments, 2)
}

// ===================== Checklists =====================

func TestTaskUseCase_AddChecklist(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	checklist, err := uc.AddChecklist(context.Background(), 1, task.ID, dto.AddChecklistInput{Title: "Checklist 1"})
	require.NoError(t, err)
	assert.Equal(t, "Checklist 1", checklist.Title)
	assert.Equal(t, task.ID, checklist.TaskID)
}

func TestTaskUseCase_AddChecklist_NotFound(t *testing.T) {
	uc, _ := setupTaskUseCase()

	_, err := uc.AddChecklist(context.Background(), 1, 999, dto.AddChecklistInput{Title: "Checklist"})
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskUseCase_AddChecklist_GetChecklistsError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, _ := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})

	repo.getChecklistsErr = errors.New("get error")
	_, err := uc.AddChecklist(ctx, 1, task.ID, dto.AddChecklistInput{Title: "CL"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get checklists")
}

func TestTaskUseCase_AddChecklist_AddError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, _ := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})

	repo.addChecklistErr = errors.New("add error")
	_, err := uc.AddChecklist(ctx, 1, task.ID, dto.AddChecklistInput{Title: "CL"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add checklist")
}

func TestTaskUseCase_AddChecklist_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)

	checklist, err := uc.AddChecklist(context.Background(), 1, task.ID, dto.AddChecklistInput{Title: "CL"})
	require.NoError(t, err)
	assert.Equal(t, "CL", checklist.Title)
}

func TestTaskUseCase_DeleteChecklist(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	checklist, _ := uc.AddChecklist(context.Background(), 1, task.ID, dto.AddChecklistInput{Title: "CL"})

	err := uc.DeleteChecklist(context.Background(), 1, checklist.ID)
	require.NoError(t, err)
}

func TestTaskUseCase_DeleteChecklist_Error(t *testing.T) {
	repo := newErrorMockTaskRepo()
	repo.deleteChecklistErr = errors.New("delete error")
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)

	err := uc.DeleteChecklist(context.Background(), 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete checklist")
}

func TestTaskUseCase_DeleteChecklist_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)
	checklist, _ := uc.AddChecklist(context.Background(), 1, task.ID, dto.AddChecklistInput{Title: "CL"})

	err := uc.DeleteChecklist(context.Background(), 1, checklist.ID)
	require.NoError(t, err)
}

func TestTaskUseCase_GetChecklists(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	ctx := context.Background()

	checklist, _ := uc.AddChecklist(ctx, 1, task.ID, dto.AddChecklistInput{Title: "CL"})
	_, _ = uc.AddChecklistItem(ctx, 1, checklist.ID, dto.AddChecklistItemInput{Title: "Item 1"})

	checklists, err := uc.GetChecklists(ctx, task.ID)
	require.NoError(t, err)
	assert.Len(t, checklists, 1)
	assert.Len(t, checklists[0].Items, 1)
}

func TestTaskUseCase_GetChecklists_GetChecklistsError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	_, _ = uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})

	repo.getChecklistsErr = errors.New("error")
	_, err := uc.GetChecklists(ctx, 1)
	require.Error(t, err)
}

func TestTaskUseCase_GetChecklists_GetItemsError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, _ := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})
	_, _ = uc.AddChecklist(ctx, 1, task.ID, dto.AddChecklistInput{Title: "CL"})

	repo.getChecklistItemErr = errors.New("items error")
	_, err := uc.GetChecklists(ctx, task.ID)
	require.Error(t, err)
}

// ===================== Checklist Items =====================

func TestTaskUseCase_AddChecklistItem(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	checklist, _ := uc.AddChecklist(context.Background(), 1, task.ID, dto.AddChecklistInput{Title: "CL"})

	item, err := uc.AddChecklistItem(context.Background(), 1, checklist.ID, dto.AddChecklistItemInput{Title: "Item 1"})
	require.NoError(t, err)
	assert.Equal(t, "Item 1", item.Title)
	assert.Equal(t, checklist.ID, item.ChecklistID)
}

func TestTaskUseCase_AddChecklistItem_GetItemsError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	repo.getChecklistItemErr = errors.New("get error")
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)

	_, err := uc.AddChecklistItem(context.Background(), 1, 1, dto.AddChecklistItemInput{Title: "Item"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get checklist items")
}

func TestTaskUseCase_AddChecklistItem_AddError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)
	ctx := context.Background()

	task, _ := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Test"})
	checklist, _ := uc.AddChecklist(ctx, 1, task.ID, dto.AddChecklistInput{Title: "CL"})

	repo.addChecklistItemErr = errors.New("add error")
	_, err := uc.AddChecklistItem(ctx, 1, checklist.ID, dto.AddChecklistItemInput{Title: "Item"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add checklist item")
}

func TestTaskUseCase_AddChecklistItem_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)
	checklist, _ := uc.AddChecklist(context.Background(), 1, task.ID, dto.AddChecklistInput{Title: "CL"})

	item, err := uc.AddChecklistItem(context.Background(), 1, checklist.ID, dto.AddChecklistItemInput{Title: "Item"})
	require.NoError(t, err)
	assert.Equal(t, "Item", item.Title)
}

func TestTaskUseCase_ToggleChecklistItem(t *testing.T) {
	uc, _ := setupTaskUseCase()

	// ToggleChecklistItem fails when no items found for given ID
	_, err := uc.ToggleChecklistItem(context.Background(), 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get checklist item")
}

func TestTaskUseCase_ToggleChecklistItem_GetItemsError(t *testing.T) {
	repo := newErrorMockTaskRepo()
	repo.getChecklistItemErr = errors.New("get error")
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)

	_, err := uc.ToggleChecklistItem(context.Background(), 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get checklist item")
}

func TestTaskUseCase_DeleteChecklistItem(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)
	checklist, _ := uc.AddChecklist(context.Background(), 1, task.ID, dto.AddChecklistInput{Title: "CL"})
	item, _ := uc.AddChecklistItem(context.Background(), 1, checklist.ID, dto.AddChecklistItemInput{Title: "Item"})

	err := uc.DeleteChecklistItem(context.Background(), 1, item.ID)
	require.NoError(t, err)
}

func TestTaskUseCase_DeleteChecklistItem_Error(t *testing.T) {
	repo := newErrorMockTaskRepo()
	repo.deleteChecklistItemErr = errors.New("delete error")
	uc := NewTaskUseCase(repo, NewMockProjectRepository(), nil, nil)

	err := uc.DeleteChecklistItem(context.Background(), 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete checklist item")
}

func TestTaskUseCase_DeleteChecklistItem_WithAuditLogger(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	task := createTask(t, uc, "Test", 1)
	checklist, _ := uc.AddChecklist(context.Background(), 1, task.ID, dto.AddChecklistInput{Title: "CL"})
	item, _ := uc.AddChecklistItem(context.Background(), 1, checklist.ID, dto.AddChecklistItemInput{Title: "Item"})

	err := uc.DeleteChecklistItem(context.Background(), 1, item.ID)
	require.NoError(t, err)
}

// ===================== History =====================

func TestTaskUseCase_GetHistory(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	newTitle := updatedTaskTitle
	_, _ = uc.Update(context.Background(), 1, task.ID, dto.UpdateTaskInput{Title: &newTitle})

	history, err := uc.GetHistory(context.Background(), task.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, history, 1)
}

func TestTaskUseCase_GetHistory_Empty(t *testing.T) {
	uc, _ := setupTaskUseCase()
	task := createTask(t, uc, "Test", 1)

	history, err := uc.GetHistory(context.Background(), task.ID, 10, 0)
	require.NoError(t, err)
	assert.Empty(t, history)
}

// ===================== Full Workflow =====================

func TestTaskUseCase_FullWorkflow(t *testing.T) {
	uc, _ := setupTaskUseCaseWithAudit()
	ctx := context.Background()

	// Create
	task, err := uc.Create(ctx, 1, dto.CreateTaskInput{Title: "Full Workflow"})
	require.NoError(t, err)
	assert.Equal(t, domain.TaskStatusNew, task.Status)

	// Assign
	task, err = uc.Assign(ctx, 1, task.ID, dto.AssignTaskInput{AssigneeID: 2})
	require.NoError(t, err)
	assert.Equal(t, domain.TaskStatusAssigned, task.Status)

	// Start work
	task, err = uc.StartWork(ctx, 2, task.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.TaskStatusInProgress, task.Status)

	// Submit for review
	task, err = uc.SubmitForReview(ctx, 2, task.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.TaskStatusReview, task.Status)

	// Complete
	task, err = uc.Complete(ctx, 1, task.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.TaskStatusCompleted, task.Status)
	assert.Equal(t, 100, task.Progress)

	// Reopen
	task, err = uc.Reopen(ctx, 1, task.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.TaskStatusNew, task.Status)
}

// ===================== Entity Tests =====================

func TestTask_IsOverdue(t *testing.T) {
	task := entities.NewTask("Test", 1)

	assert.False(t, task.IsOverdue(), "task without due date should not be overdue")

	future := time.Now().Add(24 * time.Hour)
	task.SetDueDate(&future)
	assert.False(t, task.IsOverdue(), "task with future due date should not be overdue")

	past := time.Now().Add(-24 * time.Hour)
	task.SetDueDate(&past)
	assert.True(t, task.IsOverdue(), "task with past due date should be overdue")

	task.Status = domain.TaskStatusCompleted
	assert.False(t, task.IsOverdue(), "completed task should not be overdue")

	task.Status = domain.TaskStatusCancelled
	assert.False(t, task.IsOverdue(), "cancelled task should not be overdue")
}

func TestTask_CanEdit(t *testing.T) {
	task := entities.NewTask("Test", 1)
	assert.True(t, task.CanEdit())

	task.Status = domain.TaskStatusCompleted
	assert.False(t, task.CanEdit())

	task.Status = domain.TaskStatusCancelled
	assert.False(t, task.CanEdit())

	task.Status = domain.TaskStatusInProgress
	assert.True(t, task.CanEdit())
}
