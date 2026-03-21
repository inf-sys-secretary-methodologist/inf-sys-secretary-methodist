package handlers

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
)

type mockArgs = mock.Arguments

var (
	anyCtx           = mock.Anything
	anyTask          = mock.AnythingOfType("*entities.Task")
	anyFilter        = mock.Anything
	anyHistory       = mock.AnythingOfType("*entities.TaskHistory")
	anyWatcher       = mock.AnythingOfType("*entities.TaskWatcher")
	anyComment       = mock.AnythingOfType("*entities.TaskComment")
	anyChecklist     = mock.AnythingOfType("*entities.TaskChecklist")
	anyChecklistItem = mock.AnythingOfType("*entities.TaskChecklistItem")
	anyProject       = mock.AnythingOfType("*entities.Project")
)

// mockTaskRepo implements repositories.TaskRepository
type mockTaskRepo struct {
	mock.Mock
}

func (m *mockTaskRepo) Create(ctx context.Context, task *entities.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *mockTaskRepo) Save(ctx context.Context, task *entities.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *mockTaskRepo) GetByID(ctx context.Context, id int64) (*entities.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Task), args.Error(1)
}

func (m *mockTaskRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockTaskRepo) List(ctx context.Context, filter repositories.TaskFilter, limit, offset int) ([]*entities.Task, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Task), args.Error(1)
}

func (m *mockTaskRepo) Count(ctx context.Context, filter repositories.TaskFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockTaskRepo) GetByProject(ctx context.Context, projectID int64, limit, offset int) ([]*entities.Task, error) {
	args := m.Called(ctx, projectID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Task), args.Error(1)
}

func (m *mockTaskRepo) GetByAuthor(ctx context.Context, authorID int64, limit, offset int) ([]*entities.Task, error) {
	args := m.Called(ctx, authorID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Task), args.Error(1)
}

func (m *mockTaskRepo) GetByAssignee(ctx context.Context, assigneeID int64, limit, offset int) ([]*entities.Task, error) {
	args := m.Called(ctx, assigneeID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Task), args.Error(1)
}

func (m *mockTaskRepo) GetByStatus(ctx context.Context, status domain.TaskStatus, limit, offset int) ([]*entities.Task, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Task), args.Error(1)
}

func (m *mockTaskRepo) GetOverdueTasks(ctx context.Context, limit, offset int) ([]*entities.Task, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Task), args.Error(1)
}

func (m *mockTaskRepo) AddHistory(ctx context.Context, history *entities.TaskHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *mockTaskRepo) GetHistory(ctx context.Context, taskID int64, limit, offset int) ([]*entities.TaskHistory, error) {
	args := m.Called(ctx, taskID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.TaskHistory), args.Error(1)
}

func (m *mockTaskRepo) AddWatcher(ctx context.Context, watcher *entities.TaskWatcher) error {
	args := m.Called(ctx, watcher)
	return args.Error(0)
}

func (m *mockTaskRepo) RemoveWatcher(ctx context.Context, taskID, userID int64) error {
	args := m.Called(ctx, taskID, userID)
	return args.Error(0)
}

func (m *mockTaskRepo) GetWatchers(ctx context.Context, taskID int64) ([]*entities.TaskWatcher, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.TaskWatcher), args.Error(1)
}

func (m *mockTaskRepo) IsWatching(ctx context.Context, taskID, userID int64) (bool, error) {
	args := m.Called(ctx, taskID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *mockTaskRepo) AddComment(ctx context.Context, comment *entities.TaskComment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *mockTaskRepo) UpdateComment(ctx context.Context, comment *entities.TaskComment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *mockTaskRepo) DeleteComment(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockTaskRepo) GetComments(ctx context.Context, taskID int64) ([]*entities.TaskComment, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.TaskComment), args.Error(1)
}

func (m *mockTaskRepo) GetCommentByID(ctx context.Context, id int64) (*entities.TaskComment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TaskComment), args.Error(1)
}

func (m *mockTaskRepo) AddChecklist(ctx context.Context, checklist *entities.TaskChecklist) error {
	args := m.Called(ctx, checklist)
	return args.Error(0)
}

func (m *mockTaskRepo) UpdateChecklist(ctx context.Context, checklist *entities.TaskChecklist) error {
	args := m.Called(ctx, checklist)
	return args.Error(0)
}

func (m *mockTaskRepo) DeleteChecklist(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockTaskRepo) GetChecklists(ctx context.Context, taskID int64) ([]*entities.TaskChecklist, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.TaskChecklist), args.Error(1)
}

func (m *mockTaskRepo) AddChecklistItem(ctx context.Context, item *entities.TaskChecklistItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *mockTaskRepo) UpdateChecklistItem(ctx context.Context, item *entities.TaskChecklistItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *mockTaskRepo) DeleteChecklistItem(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockTaskRepo) GetChecklistItems(ctx context.Context, checklistID int64) ([]*entities.TaskChecklistItem, error) {
	args := m.Called(ctx, checklistID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.TaskChecklistItem), args.Error(1)
}

func (m *mockTaskRepo) AddAttachment(ctx context.Context, attachment *entities.TaskAttachment) error {
	args := m.Called(ctx, attachment)
	return args.Error(0)
}

func (m *mockTaskRepo) RemoveAttachment(ctx context.Context, attachmentID int64) error {
	args := m.Called(ctx, attachmentID)
	return args.Error(0)
}

func (m *mockTaskRepo) GetAttachments(ctx context.Context, taskID int64) ([]*entities.TaskAttachment, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.TaskAttachment), args.Error(1)
}

func (m *mockTaskRepo) GetAttachmentByID(ctx context.Context, attachmentID int64) (*entities.TaskAttachment, error) {
	args := m.Called(ctx, attachmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TaskAttachment), args.Error(1)
}

// mockProjectRepo implements repositories.ProjectRepository
type mockProjectRepo struct {
	mock.Mock
}

func (m *mockProjectRepo) Create(ctx context.Context, project *entities.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *mockProjectRepo) Save(ctx context.Context, project *entities.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *mockProjectRepo) GetByID(ctx context.Context, id int64) (*entities.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Project), args.Error(1)
}

func (m *mockProjectRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockProjectRepo) List(ctx context.Context, filter repositories.ProjectFilter, limit, offset int) ([]*entities.Project, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Project), args.Error(1)
}

func (m *mockProjectRepo) Count(ctx context.Context, filter repositories.ProjectFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockProjectRepo) GetByOwner(ctx context.Context, ownerID int64, limit, offset int) ([]*entities.Project, error) {
	args := m.Called(ctx, ownerID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Project), args.Error(1)
}

func (m *mockProjectRepo) GetByStatus(ctx context.Context, status domain.ProjectStatus, limit, offset int) ([]*entities.Project, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Project), args.Error(1)
}
