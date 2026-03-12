// Package persistence provides PostgreSQL repository implementations for the tasks module.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
)

// TaskRepositoryPG implements TaskRepository using PostgreSQL.
type TaskRepositoryPG struct {
	db *sql.DB
}

// NewTaskRepositoryPG creates a new TaskRepositoryPG.
func NewTaskRepositoryPG(db *sql.DB) *TaskRepositoryPG {
	return &TaskRepositoryPG{db: db}
}

// Create creates a new task.
func (r *TaskRepositoryPG) Create(ctx context.Context, task *entities.Task) error {
	query := `
		INSERT INTO tasks (
			project_id, title, description, document_id, author_id, assignee_id,
			status, priority, due_date, start_date, progress, estimated_hours,
			tags, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		task.ProjectID, task.Title, task.Description, task.DocumentID,
		task.AuthorID, task.AssigneeID, task.Status, task.Priority,
		task.DueDate, task.StartDate, task.Progress, task.EstimatedHours,
		pq.Array(task.Tags), task.Metadata, task.CreatedAt, task.UpdatedAt,
	).Scan(&task.ID)
}

// Save updates an existing task.
func (r *TaskRepositoryPG) Save(ctx context.Context, task *entities.Task) error {
	query := `
		UPDATE tasks SET
			project_id = $1, title = $2, description = $3, document_id = $4,
			assignee_id = $5, status = $6, priority = $7, due_date = $8,
			start_date = $9, completed_at = $10, progress = $11,
			estimated_hours = $12, actual_hours = $13, tags = $14,
			metadata = $15, updated_at = $16
		WHERE id = $17`

	_, err := r.db.ExecContext(ctx, query,
		task.ProjectID, task.Title, task.Description, task.DocumentID,
		task.AssigneeID, task.Status, task.Priority, task.DueDate,
		task.StartDate, task.CompletedAt, task.Progress,
		task.EstimatedHours, task.ActualHours, pq.Array(task.Tags),
		task.Metadata, task.UpdatedAt, task.ID,
	)
	return err
}

// GetByID retrieves a task by ID.
func (r *TaskRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.Task, error) {
	query := `
		SELECT id, project_id, title, description, document_id, author_id,
			assignee_id, status, priority, due_date, start_date, completed_at,
			progress, estimated_hours, actual_hours, tags, metadata,
			created_at, updated_at
		FROM tasks WHERE id = $1`

	task := &entities.Task{}
	var tags pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.ProjectID, &task.Title, &task.Description,
		&task.DocumentID, &task.AuthorID, &task.AssigneeID, &task.Status,
		&task.Priority, &task.DueDate, &task.StartDate, &task.CompletedAt,
		&task.Progress, &task.EstimatedHours, &task.ActualHours,
		&tags, &task.Metadata, &task.CreatedAt, &task.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	task.Tags = tags
	return task, nil
}

// Delete deletes a task.
func (r *TaskRepositoryPG) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1", id)
	return err
}

// List lists tasks with filters.
func (r *TaskRepositoryPG) List(ctx context.Context, filter repositories.TaskFilter, limit, offset int) ([]*entities.Task, error) {
	query, args := r.buildListQuery(filter, limit, offset, false)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanTasks(rows)
}

// Count counts tasks with filters.
func (r *TaskRepositoryPG) Count(ctx context.Context, filter repositories.TaskFilter) (int64, error) {
	query, args := r.buildListQuery(filter, 0, 0, true)

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *TaskRepositoryPG) buildListQuery(filter repositories.TaskFilter, limit, offset int, countOnly bool) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argNum := 1

	if filter.ProjectID != nil {
		conditions = append(conditions, fmt.Sprintf("project_id = $%d", argNum))
		args = append(args, *filter.ProjectID)
		argNum++
	}

	if filter.AuthorID != nil {
		conditions = append(conditions, fmt.Sprintf("author_id = $%d", argNum))
		args = append(args, *filter.AuthorID)
		argNum++
	}

	if filter.AssigneeID != nil {
		conditions = append(conditions, fmt.Sprintf("assignee_id = $%d", argNum))
		args = append(args, *filter.AssigneeID)
		argNum++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argNum))
		args = append(args, *filter.Status)
		argNum++
	}

	if filter.Priority != nil {
		conditions = append(conditions, fmt.Sprintf("priority = $%d", argNum))
		args = append(args, *filter.Priority)
		argNum++
	}

	if filter.IsOverdue != nil && *filter.IsOverdue {
		conditions = append(conditions, "due_date < NOW() AND status NOT IN ('completed', 'canceled')")
	}

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argNum, argNum))
		args = append(args, "%"+*filter.Search+"%")
		argNum++
	}

	if len(filter.Tags) > 0 {
		conditions = append(conditions, fmt.Sprintf("tags && $%d", argNum))
		args = append(args, pq.Array(filter.Tags))
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	if countOnly {
		return "SELECT COUNT(*) FROM tasks" + whereClause, args
	}

	query := `
		SELECT id, project_id, title, description, document_id, author_id,
			assignee_id, status, priority, due_date, start_date, completed_at,
			progress, estimated_hours, actual_hours, tags, metadata,
			created_at, updated_at
		FROM tasks` + whereClause + ` ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	}

	return query, args
}

func (r *TaskRepositoryPG) scanTasks(rows *sql.Rows) ([]*entities.Task, error) {
	var tasks []*entities.Task

	for rows.Next() {
		task := &entities.Task{}
		var tags pq.StringArray

		err := rows.Scan(
			&task.ID, &task.ProjectID, &task.Title, &task.Description,
			&task.DocumentID, &task.AuthorID, &task.AssigneeID, &task.Status,
			&task.Priority, &task.DueDate, &task.StartDate, &task.CompletedAt,
			&task.Progress, &task.EstimatedHours, &task.ActualHours,
			&tags, &task.Metadata, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		task.Tags = tags
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// GetByProject retrieves tasks by project ID.
func (r *TaskRepositoryPG) GetByProject(ctx context.Context, projectID int64, limit, offset int) ([]*entities.Task, error) {
	filter := repositories.TaskFilter{ProjectID: &projectID}
	return r.List(ctx, filter, limit, offset)
}

// GetByAuthor retrieves tasks by author ID.
func (r *TaskRepositoryPG) GetByAuthor(ctx context.Context, authorID int64, limit, offset int) ([]*entities.Task, error) {
	filter := repositories.TaskFilter{AuthorID: &authorID}
	return r.List(ctx, filter, limit, offset)
}

// GetByAssignee retrieves tasks by assignee ID.
func (r *TaskRepositoryPG) GetByAssignee(ctx context.Context, assigneeID int64, limit, offset int) ([]*entities.Task, error) {
	filter := repositories.TaskFilter{AssigneeID: &assigneeID}
	return r.List(ctx, filter, limit, offset)
}

// GetByStatus retrieves tasks by status.
func (r *TaskRepositoryPG) GetByStatus(ctx context.Context, status domain.TaskStatus, limit, offset int) ([]*entities.Task, error) {
	filter := repositories.TaskFilter{Status: &status}
	return r.List(ctx, filter, limit, offset)
}

// GetOverdueTasks retrieves overdue tasks.
func (r *TaskRepositoryPG) GetOverdueTasks(ctx context.Context, limit, offset int) ([]*entities.Task, error) {
	isOverdue := true
	filter := repositories.TaskFilter{IsOverdue: &isOverdue}
	return r.List(ctx, filter, limit, offset)
}

// AddWatcher adds a watcher to a task.
func (r *TaskRepositoryPG) AddWatcher(ctx context.Context, watcher *entities.TaskWatcher) error {
	query := `INSERT INTO task_watchers (task_id, user_id, created_at) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, watcher.TaskID, watcher.UserID, watcher.CreatedAt)
	return err
}

// RemoveWatcher removes a watcher from a task.
func (r *TaskRepositoryPG) RemoveWatcher(ctx context.Context, taskID, userID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM task_watchers WHERE task_id = $1 AND user_id = $2", taskID, userID)
	return err
}

// GetWatchers retrieves watchers of a task.
func (r *TaskRepositoryPG) GetWatchers(ctx context.Context, taskID int64) ([]*entities.TaskWatcher, error) {
	query := `SELECT task_id, user_id, created_at FROM task_watchers WHERE task_id = $1`
	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var watchers []*entities.TaskWatcher
	for rows.Next() {
		w := &entities.TaskWatcher{}
		if err := rows.Scan(&w.TaskID, &w.UserID, &w.CreatedAt); err != nil {
			return nil, err
		}
		watchers = append(watchers, w)
	}
	return watchers, rows.Err()
}

// IsWatching checks if a user is watching a task.
func (r *TaskRepositoryPG) IsWatching(ctx context.Context, taskID, userID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM task_watchers WHERE task_id = $1 AND user_id = $2)`
	err := r.db.QueryRowContext(ctx, query, taskID, userID).Scan(&exists)
	return exists, err
}

// AddAttachment adds an attachment to a task.
func (r *TaskRepositoryPG) AddAttachment(ctx context.Context, attachment *entities.TaskAttachment) error {
	query := `
		INSERT INTO task_attachments (task_id, file_name, file_path, file_size, mime_type, uploaded_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	return r.db.QueryRowContext(ctx, query,
		attachment.TaskID, attachment.FileName, attachment.FilePath,
		attachment.FileSize, attachment.MimeType, attachment.UploadedBy, attachment.CreatedAt,
	).Scan(&attachment.ID)
}

// RemoveAttachment removes an attachment.
func (r *TaskRepositoryPG) RemoveAttachment(ctx context.Context, attachmentID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM task_attachments WHERE id = $1", attachmentID)
	return err
}

// GetAttachments retrieves attachments of a task.
func (r *TaskRepositoryPG) GetAttachments(ctx context.Context, taskID int64) ([]*entities.TaskAttachment, error) {
	query := `SELECT id, task_id, file_name, file_path, file_size, mime_type, uploaded_by, created_at
		FROM task_attachments WHERE task_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var attachments []*entities.TaskAttachment
	for rows.Next() {
		a := &entities.TaskAttachment{}
		if err := rows.Scan(&a.ID, &a.TaskID, &a.FileName, &a.FilePath, &a.FileSize, &a.MimeType, &a.UploadedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		attachments = append(attachments, a)
	}
	return attachments, rows.Err()
}

// GetAttachmentByID retrieves an attachment by ID.
func (r *TaskRepositoryPG) GetAttachmentByID(ctx context.Context, attachmentID int64) (*entities.TaskAttachment, error) {
	query := `SELECT id, task_id, file_name, file_path, file_size, mime_type, uploaded_by, created_at
		FROM task_attachments WHERE id = $1`
	a := &entities.TaskAttachment{}
	err := r.db.QueryRowContext(ctx, query, attachmentID).Scan(
		&a.ID, &a.TaskID, &a.FileName, &a.FilePath, &a.FileSize, &a.MimeType, &a.UploadedBy, &a.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return a, err
}

// AddComment adds a comment to a task.
func (r *TaskRepositoryPG) AddComment(ctx context.Context, comment *entities.TaskComment) error {
	query := `
		INSERT INTO task_comments (task_id, author_id, content, parent_comment_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	return r.db.QueryRowContext(ctx, query,
		comment.TaskID, comment.AuthorID, comment.Content,
		comment.ParentCommentID, comment.CreatedAt, comment.UpdatedAt,
	).Scan(&comment.ID)
}

// UpdateComment updates a comment.
func (r *TaskRepositoryPG) UpdateComment(ctx context.Context, comment *entities.TaskComment) error {
	query := `UPDATE task_comments SET content = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, comment.Content, comment.UpdatedAt, comment.ID)
	return err
}

// DeleteComment deletes a comment.
func (r *TaskRepositoryPG) DeleteComment(ctx context.Context, commentID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM task_comments WHERE id = $1", commentID)
	return err
}

// GetComments retrieves comments of a task.
func (r *TaskRepositoryPG) GetComments(ctx context.Context, taskID int64) ([]*entities.TaskComment, error) {
	query := `SELECT id, task_id, author_id, content, parent_comment_id, created_at, updated_at
		FROM task_comments WHERE task_id = $1 ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var comments []*entities.TaskComment
	for rows.Next() {
		c := &entities.TaskComment{}
		if err := rows.Scan(&c.ID, &c.TaskID, &c.AuthorID, &c.Content, &c.ParentCommentID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

// GetCommentByID retrieves a comment by ID.
func (r *TaskRepositoryPG) GetCommentByID(ctx context.Context, commentID int64) (*entities.TaskComment, error) {
	query := `SELECT id, task_id, author_id, content, parent_comment_id, created_at, updated_at
		FROM task_comments WHERE id = $1`
	c := &entities.TaskComment{}
	err := r.db.QueryRowContext(ctx, query, commentID).Scan(
		&c.ID, &c.TaskID, &c.AuthorID, &c.Content, &c.ParentCommentID, &c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return c, err
}

// AddChecklist adds a checklist to a task.
func (r *TaskRepositoryPG) AddChecklist(ctx context.Context, checklist *entities.TaskChecklist) error {
	query := `INSERT INTO task_checklists (task_id, title, position, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	return r.db.QueryRowContext(ctx, query,
		checklist.TaskID, checklist.Title, checklist.Position, checklist.CreatedAt,
	).Scan(&checklist.ID)
}

// UpdateChecklist updates a checklist.
func (r *TaskRepositoryPG) UpdateChecklist(ctx context.Context, checklist *entities.TaskChecklist) error {
	query := `UPDATE task_checklists SET title = $1, position = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, checklist.Title, checklist.Position, checklist.ID)
	return err
}

// DeleteChecklist deletes a checklist.
func (r *TaskRepositoryPG) DeleteChecklist(ctx context.Context, checklistID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM task_checklists WHERE id = $1", checklistID)
	return err
}

// GetChecklists retrieves checklists of a task.
func (r *TaskRepositoryPG) GetChecklists(ctx context.Context, taskID int64) ([]*entities.TaskChecklist, error) {
	query := `SELECT id, task_id, title, position, created_at FROM task_checklists WHERE task_id = $1 ORDER BY position`
	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var checklists []*entities.TaskChecklist
	for rows.Next() {
		c := &entities.TaskChecklist{}
		if err := rows.Scan(&c.ID, &c.TaskID, &c.Title, &c.Position, &c.CreatedAt); err != nil {
			return nil, err
		}
		checklists = append(checklists, c)
	}
	return checklists, rows.Err()
}

// AddChecklistItem adds an item to a checklist.
func (r *TaskRepositoryPG) AddChecklistItem(ctx context.Context, item *entities.TaskChecklistItem) error {
	query := `INSERT INTO task_checklist_items (checklist_id, title, is_completed, position, created_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`
	return r.db.QueryRowContext(ctx, query,
		item.ChecklistID, item.Title, item.IsCompleted, item.Position, item.CreatedAt,
	).Scan(&item.ID)
}

// UpdateChecklistItem updates a checklist item.
func (r *TaskRepositoryPG) UpdateChecklistItem(ctx context.Context, item *entities.TaskChecklistItem) error {
	query := `UPDATE task_checklist_items SET title = $1, is_completed = $2, position = $3,
		completed_by = $4, completed_at = $5 WHERE id = $6`
	_, err := r.db.ExecContext(ctx, query,
		item.Title, item.IsCompleted, item.Position, item.CompletedBy, item.CompletedAt, item.ID,
	)
	return err
}

// DeleteChecklistItem deletes a checklist item.
func (r *TaskRepositoryPG) DeleteChecklistItem(ctx context.Context, itemID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM task_checklist_items WHERE id = $1", itemID)
	return err
}

// GetChecklistItems retrieves items of a checklist.
func (r *TaskRepositoryPG) GetChecklistItems(ctx context.Context, checklistID int64) ([]*entities.TaskChecklistItem, error) {
	query := `SELECT id, checklist_id, title, is_completed, position, completed_by, completed_at, created_at
		FROM task_checklist_items WHERE checklist_id = $1 ORDER BY position`
	rows, err := r.db.QueryContext(ctx, query, checklistID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []*entities.TaskChecklistItem
	for rows.Next() {
		i := &entities.TaskChecklistItem{}
		if err := rows.Scan(&i.ID, &i.ChecklistID, &i.Title, &i.IsCompleted, &i.Position, &i.CompletedBy, &i.CompletedAt, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

// AddHistory adds a history entry.
func (r *TaskRepositoryPG) AddHistory(ctx context.Context, history *entities.TaskHistory) error {
	query := `INSERT INTO task_history (task_id, user_id, field_name, old_value, new_value, created_at)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	return r.db.QueryRowContext(ctx, query,
		history.TaskID, history.UserID, history.FieldName, history.OldValue, history.NewValue, history.CreatedAt,
	).Scan(&history.ID)
}

// GetHistory retrieves history of a task.
func (r *TaskRepositoryPG) GetHistory(ctx context.Context, taskID int64, limit, offset int) ([]*entities.TaskHistory, error) {
	query := `SELECT id, task_id, user_id, field_name, old_value, new_value, created_at
		FROM task_history WHERE task_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, taskID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var history []*entities.TaskHistory
	for rows.Next() {
		h := &entities.TaskHistory{}
		if err := rows.Scan(&h.ID, &h.TaskID, &h.UserID, &h.FieldName, &h.OldValue, &h.NewValue, &h.CreatedAt); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, rows.Err()
}
