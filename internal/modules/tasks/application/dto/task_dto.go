// Package dto provides data transfer objects for the tasks module.
package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
)

// CreateTaskInput represents input for creating a task.
type CreateTaskInput struct {
	ProjectID      *int64            `json:"project_id,omitempty"`
	Title          string            `json:"title" validate:"required,min=1,max=500"`
	Description    *string           `json:"description,omitempty"`
	DocumentID     *int64            `json:"document_id,omitempty"`
	AssigneeID     *int64            `json:"assignee_id,omitempty"`
	Priority       *string           `json:"priority,omitempty"`
	DueDate        *time.Time        `json:"due_date,omitempty"`
	StartDate      *time.Time        `json:"start_date,omitempty"`
	EstimatedHours *float64          `json:"estimated_hours,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	Metadata       map[string]any    `json:"metadata,omitempty"`
}

// UpdateTaskInput represents input for updating a task.
type UpdateTaskInput struct {
	Title          *string           `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
	Description    *string           `json:"description,omitempty"`
	ProjectID      *int64            `json:"project_id,omitempty"`
	AssigneeID     *int64            `json:"assignee_id,omitempty"`
	Priority       *string           `json:"priority,omitempty"`
	DueDate        *time.Time        `json:"due_date,omitempty"`
	StartDate      *time.Time        `json:"start_date,omitempty"`
	Progress       *int              `json:"progress,omitempty" validate:"omitempty,min=0,max=100"`
	EstimatedHours *float64          `json:"estimated_hours,omitempty"`
	ActualHours    *float64          `json:"actual_hours,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	Metadata       map[string]any    `json:"metadata,omitempty"`
}

// AssignTaskInput represents input for assigning a task.
type AssignTaskInput struct {
	AssigneeID int64 `json:"assignee_id" validate:"required"`
}

// TaskFilterInput represents input for filtering tasks.
type TaskFilterInput struct {
	ProjectID  *int64  `form:"project_id"`
	AuthorID   *int64  `form:"author_id"`
	AssigneeID *int64  `form:"assignee_id"`
	Status     *string `form:"status"`
	Priority   *string `form:"priority"`
	IsOverdue  *bool   `form:"is_overdue"`
	Search     *string `form:"search"`
	Tags       []string `form:"tags"`
	Limit      int     `form:"limit,default=20"`
	Offset     int     `form:"offset,default=0"`
}

// ToTaskFilter converts TaskFilterInput to domain TaskFilter.
func (f *TaskFilterInput) ToTaskFilter() repositories.TaskFilter {
	filter := repositories.TaskFilter{
		ProjectID:  f.ProjectID,
		AuthorID:   f.AuthorID,
		AssigneeID: f.AssigneeID,
		IsOverdue:  f.IsOverdue,
		Search:     f.Search,
		Tags:       f.Tags,
	}
	if f.Status != nil {
		status := domain.TaskStatus(*f.Status)
		filter.Status = &status
	}
	if f.Priority != nil {
		priority := domain.TaskPriority(*f.Priority)
		filter.Priority = &priority
	}
	return filter
}

// TaskOutput represents the output for a task.
type TaskOutput struct {
	ID             int64                  `json:"id"`
	ProjectID      *int64                 `json:"project_id,omitempty"`
	Title          string                 `json:"title"`
	Description    *string                `json:"description,omitempty"`
	DocumentID     *int64                 `json:"document_id,omitempty"`
	AuthorID       int64                  `json:"author_id"`
	AssigneeID     *int64                 `json:"assignee_id,omitempty"`
	Status         string                 `json:"status"`
	Priority       string                 `json:"priority"`
	DueDate        *time.Time             `json:"due_date,omitempty"`
	StartDate      *time.Time             `json:"start_date,omitempty"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	Progress       int                    `json:"progress"`
	EstimatedHours *float64               `json:"estimated_hours,omitempty"`
	ActualHours    *float64               `json:"actual_hours,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	Metadata       map[string]any         `json:"metadata,omitempty"`
	IsOverdue      bool                   `json:"is_overdue"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	Project        *ProjectOutput         `json:"project,omitempty"`
	Assignee       *UserOutput            `json:"assignee,omitempty"`
	Watchers       []UserOutput           `json:"watchers,omitempty"`
	Checklists     []TaskChecklistOutput  `json:"checklists,omitempty"`
}

// TaskListOutput represents the output for a list of tasks.
type TaskListOutput struct {
	Tasks      []TaskOutput `json:"tasks"`
	Total      int64        `json:"total"`
	Limit      int          `json:"limit"`
	Offset     int          `json:"offset"`
}

// UserOutput represents basic user info in responses.
type UserOutput struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ToTaskOutput converts a Task entity to TaskOutput.
func ToTaskOutput(task *entities.Task) TaskOutput {
	output := TaskOutput{
		ID:             task.ID,
		ProjectID:      task.ProjectID,
		Title:          task.Title,
		Description:    task.Description,
		DocumentID:     task.DocumentID,
		AuthorID:       task.AuthorID,
		AssigneeID:     task.AssigneeID,
		Status:         string(task.Status),
		Priority:       string(task.Priority),
		DueDate:        task.DueDate,
		StartDate:      task.StartDate,
		CompletedAt:    task.CompletedAt,
		Progress:       task.Progress,
		EstimatedHours: task.EstimatedHours,
		ActualHours:    task.ActualHours,
		Tags:           task.Tags,
		IsOverdue:      task.IsOverdue(),
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
	}

	if task.Metadata != nil {
		_ = task.Metadata // metadata is already json.RawMessage, handle in handler if needed
	}

	if task.Project != nil {
		projectOutput := ToProjectOutput(task.Project)
		output.Project = &projectOutput
	}

	if task.Assignee != nil {
		output.Assignee = &UserOutput{
			ID:    task.Assignee.ID,
			Name:  task.Assignee.Name,
			Email: task.Assignee.Email,
		}
	}

	for _, w := range task.Watchers {
		if w.User != nil {
			output.Watchers = append(output.Watchers, UserOutput{
				ID:    w.User.ID,
				Name:  w.User.Name,
				Email: w.User.Email,
			})
		}
	}

	for _, c := range task.Checklists {
		output.Checklists = append(output.Checklists, ToTaskChecklistOutput(&c))
	}

	return output
}

// AddCommentInput represents input for adding a comment.
type AddCommentInput struct {
	Content         string `json:"content" validate:"required,min=1"`
	ParentCommentID *int64 `json:"parent_comment_id,omitempty"`
}

// UpdateCommentInput represents input for updating a comment.
type UpdateCommentInput struct {
	Content string `json:"content" validate:"required,min=1"`
}

// TaskCommentOutput represents the output for a task comment.
type TaskCommentOutput struct {
	ID              int64               `json:"id"`
	TaskID          int64               `json:"task_id"`
	AuthorID        int64               `json:"author_id"`
	Content         string              `json:"content"`
	ParentCommentID *int64              `json:"parent_comment_id,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	Author          *UserOutput         `json:"author,omitempty"`
	Replies         []TaskCommentOutput `json:"replies,omitempty"`
}

// ToTaskCommentOutput converts a TaskComment entity to TaskCommentOutput.
func ToTaskCommentOutput(comment *entities.TaskComment) TaskCommentOutput {
	output := TaskCommentOutput{
		ID:              comment.ID,
		TaskID:          comment.TaskID,
		AuthorID:        comment.AuthorID,
		Content:         comment.Content,
		ParentCommentID: comment.ParentCommentID,
		CreatedAt:       comment.CreatedAt,
		UpdatedAt:       comment.UpdatedAt,
	}

	if comment.Author != nil {
		output.Author = &UserOutput{
			ID:    comment.Author.ID,
			Name:  comment.Author.Name,
			Email: comment.Author.Email,
		}
	}

	for _, r := range comment.Replies {
		output.Replies = append(output.Replies, ToTaskCommentOutput(&r))
	}

	return output
}

// TaskAttachmentOutput represents the output for a task attachment.
type TaskAttachmentOutput struct {
	ID         int64     `json:"id"`
	TaskID     int64     `json:"task_id"`
	FileName   string    `json:"file_name"`
	FilePath   string    `json:"file_path"`
	FileSize   int64     `json:"file_size"`
	MimeType   *string   `json:"mime_type,omitempty"`
	UploadedBy int64     `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
}

// ToTaskAttachmentOutput converts a TaskAttachment entity to TaskAttachmentOutput.
func ToTaskAttachmentOutput(attachment *entities.TaskAttachment) TaskAttachmentOutput {
	return TaskAttachmentOutput{
		ID:         attachment.ID,
		TaskID:     attachment.TaskID,
		FileName:   attachment.FileName,
		FilePath:   attachment.FilePath,
		FileSize:   attachment.FileSize,
		MimeType:   attachment.MimeType,
		UploadedBy: attachment.UploadedBy,
		CreatedAt:  attachment.CreatedAt,
	}
}

// AddChecklistInput represents input for adding a checklist.
type AddChecklistInput struct {
	Title string `json:"title" validate:"required,min=1,max=500"`
}

// AddChecklistItemInput represents input for adding a checklist item.
type AddChecklistItemInput struct {
	Title string `json:"title" validate:"required,min=1,max=500"`
}

// TaskChecklistOutput represents the output for a task checklist.
type TaskChecklistOutput struct {
	ID                   int64                     `json:"id"`
	TaskID               int64                     `json:"task_id"`
	Title                string                    `json:"title"`
	Position             int                       `json:"position"`
	CompletionPercentage int                       `json:"completion_percentage"`
	CreatedAt            time.Time                 `json:"created_at"`
	Items                []TaskChecklistItemOutput `json:"items,omitempty"`
}

// TaskChecklistItemOutput represents the output for a checklist item.
type TaskChecklistItemOutput struct {
	ID          int64      `json:"id"`
	ChecklistID int64      `json:"checklist_id"`
	Title       string     `json:"title"`
	IsCompleted bool       `json:"is_completed"`
	Position    int        `json:"position"`
	CompletedBy *int64     `json:"completed_by,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ToTaskChecklistOutput converts a TaskChecklist entity to TaskChecklistOutput.
func ToTaskChecklistOutput(checklist *entities.TaskChecklist) TaskChecklistOutput {
	output := TaskChecklistOutput{
		ID:                   checklist.ID,
		TaskID:               checklist.TaskID,
		Title:                checklist.Title,
		Position:             checklist.Position,
		CompletionPercentage: checklist.CompletionPercentage(),
		CreatedAt:            checklist.CreatedAt,
	}

	for _, item := range checklist.Items {
		output.Items = append(output.Items, TaskChecklistItemOutput{
			ID:          item.ID,
			ChecklistID: item.ChecklistID,
			Title:       item.Title,
			IsCompleted: item.IsCompleted,
			Position:    item.Position,
			CompletedBy: item.CompletedBy,
			CompletedAt: item.CompletedAt,
			CreatedAt:   item.CreatedAt,
		})
	}

	return output
}

// TaskHistoryOutput represents the output for a task history entry.
type TaskHistoryOutput struct {
	ID        int64       `json:"id"`
	TaskID    int64       `json:"task_id"`
	UserID    *int64      `json:"user_id,omitempty"`
	FieldName string      `json:"field_name"`
	OldValue  *string     `json:"old_value,omitempty"`
	NewValue  *string     `json:"new_value,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	User      *UserOutput `json:"user,omitempty"`
}

// ToTaskHistoryOutput converts a TaskHistory entity to TaskHistoryOutput.
func ToTaskHistoryOutput(history *entities.TaskHistory) TaskHistoryOutput {
	output := TaskHistoryOutput{
		ID:        history.ID,
		TaskID:    history.TaskID,
		UserID:    history.UserID,
		FieldName: history.FieldName,
		OldValue:  history.OldValue,
		NewValue:  history.NewValue,
		CreatedAt: history.CreatedAt,
	}

	if history.User != nil {
		output.User = &UserOutput{
			ID:    history.User.ID,
			Name:  history.User.Name,
			Email: history.User.Email,
		}
	}

	return output
}
