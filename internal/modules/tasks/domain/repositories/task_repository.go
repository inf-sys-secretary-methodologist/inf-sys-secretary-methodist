// Package repositories provides repository interfaces for the tasks module.
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
)

// TaskFilter defines filtering options for task queries.
type TaskFilter struct {
	ProjectID  *int64
	AuthorID   *int64
	AssigneeID *int64
	Status     *domain.TaskStatus
	Priority   *domain.TaskPriority
	IsOverdue  *bool
	Search     *string
	Tags       []string
}

// TaskRepository defines the interface for task data access.
type TaskRepository interface {
	// CRUD operations
	Create(ctx context.Context, task *entities.Task) error
	Save(ctx context.Context, task *entities.Task) error
	GetByID(ctx context.Context, id int64) (*entities.Task, error)
	Delete(ctx context.Context, id int64) error

	// Query operations
	List(ctx context.Context, filter TaskFilter, limit, offset int) ([]*entities.Task, error)
	Count(ctx context.Context, filter TaskFilter) (int64, error)
	GetByProject(ctx context.Context, projectID int64, limit, offset int) ([]*entities.Task, error)
	GetByAuthor(ctx context.Context, authorID int64, limit, offset int) ([]*entities.Task, error)
	GetByAssignee(ctx context.Context, assigneeID int64, limit, offset int) ([]*entities.Task, error)
	GetByStatus(ctx context.Context, status domain.TaskStatus, limit, offset int) ([]*entities.Task, error)
	GetOverdueTasks(ctx context.Context, limit, offset int) ([]*entities.Task, error)

	// Watchers
	AddWatcher(ctx context.Context, watcher *entities.TaskWatcher) error
	RemoveWatcher(ctx context.Context, taskID, userID int64) error
	GetWatchers(ctx context.Context, taskID int64) ([]*entities.TaskWatcher, error)
	IsWatching(ctx context.Context, taskID, userID int64) (bool, error)

	// Attachments
	AddAttachment(ctx context.Context, attachment *entities.TaskAttachment) error
	RemoveAttachment(ctx context.Context, attachmentID int64) error
	GetAttachments(ctx context.Context, taskID int64) ([]*entities.TaskAttachment, error)
	GetAttachmentByID(ctx context.Context, attachmentID int64) (*entities.TaskAttachment, error)

	// Comments
	AddComment(ctx context.Context, comment *entities.TaskComment) error
	UpdateComment(ctx context.Context, comment *entities.TaskComment) error
	DeleteComment(ctx context.Context, commentID int64) error
	GetComments(ctx context.Context, taskID int64) ([]*entities.TaskComment, error)
	GetCommentByID(ctx context.Context, commentID int64) (*entities.TaskComment, error)

	// Checklists
	AddChecklist(ctx context.Context, checklist *entities.TaskChecklist) error
	UpdateChecklist(ctx context.Context, checklist *entities.TaskChecklist) error
	DeleteChecklist(ctx context.Context, checklistID int64) error
	GetChecklists(ctx context.Context, taskID int64) ([]*entities.TaskChecklist, error)

	// Checklist items
	AddChecklistItem(ctx context.Context, item *entities.TaskChecklistItem) error
	UpdateChecklistItem(ctx context.Context, item *entities.TaskChecklistItem) error
	DeleteChecklistItem(ctx context.Context, itemID int64) error
	GetChecklistItems(ctx context.Context, checklistID int64) ([]*entities.TaskChecklistItem, error)

	// History
	AddHistory(ctx context.Context, history *entities.TaskHistory) error
	GetHistory(ctx context.Context, taskID int64, limit, offset int) ([]*entities.TaskHistory, error)
}
