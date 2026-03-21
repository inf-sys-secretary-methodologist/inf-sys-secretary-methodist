package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
)

func newTaskRepoMock(t *testing.T) (*TaskRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewTaskRepositoryPG(db), mock
}

var taskCols = []string{
	"id", "project_id", "title", "description", "document_id", "author_id",
	"assignee_id", "status", "priority", "due_date", "start_date", "completed_at",
	"progress", "estimated_hours", "actual_hours", "tags", "metadata",
	"created_at", "updated_at",
}

func newTaskRows() *sqlmock.Rows { return sqlmock.NewRows(taskCols) }

func addTaskRow(rows *sqlmock.Rows, id int64, title string) *sqlmock.Rows {
	now := time.Now()
	meta, _ := json.Marshal(map[string]string{"key": "val"})
	return rows.AddRow(
		id, nil, title, nil, nil, int64(1),
		nil, domain.TaskStatusNew, domain.TaskPriorityNormal, nil, nil, nil,
		0, nil, nil, pq.StringArray{"tag1"}, meta,
		now, now,
	)
}

// --- Create ---

func TestTaskRepositoryPG_Create_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	task := entities.NewTask("Test Task", 1)
	task.Tags = []string{"tag1"}
	task.Metadata = json.RawMessage(`{}`)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO tasks")).
		WithArgs(task.ProjectID, task.Title, task.Description, task.DocumentID,
			task.AuthorID, task.AssigneeID, task.Status, task.Priority,
			task.DueDate, task.StartDate, task.Progress, task.EstimatedHours,
			pq.Array(task.Tags), task.Metadata, task.CreatedAt, task.UpdatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.Create(context.Background(), task)
	require.NoError(t, err)
	assert.Equal(t, int64(1), task.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_Create_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	task := entities.NewTask("Test", 1)
	task.Metadata = json.RawMessage(`{}`)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO tasks")).
		WillReturnError(sql.ErrConnDone)

	err := repo.Create(context.Background(), task)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Save ---

func TestTaskRepositoryPG_Save_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	task := entities.NewTask("Updated", 1)
	task.ID = 5
	task.Tags = []string{"tag1"}
	task.Metadata = json.RawMessage(`{}`)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE tasks SET")).
		WithArgs(task.ProjectID, task.Title, task.Description, task.DocumentID,
			task.AssigneeID, task.Status, task.Priority, task.DueDate,
			task.StartDate, task.CompletedAt, task.Progress,
			task.EstimatedHours, task.ActualHours, pq.Array(task.Tags),
			task.Metadata, task.UpdatedAt, task.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Save(context.Background(), task)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_Save_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	task := entities.NewTask("err", 1)
	task.ID = 5
	task.Metadata = json.RawMessage(`{}`)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE tasks SET")).
		WillReturnError(sql.ErrConnDone)

	err := repo.Save(context.Background(), task)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByID ---

func TestTaskRepositoryPG_GetByID_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	rows := addTaskRow(newTaskRows(), 1, "Task1")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, project_id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	task, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "Task1", task.Title)
	assert.Equal(t, []string{"tag1"}, task.Tags)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, project_id")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	task, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, task)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetByID_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, project_id")).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	task, err := repo.GetByID(context.Background(), 1)
	require.Error(t, err)
	assert.Nil(t, task)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Delete ---

func TestTaskRepositoryPG_Delete_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM tasks WHERE id = $1")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_Delete_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM tasks WHERE id = $1")).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	err := repo.Delete(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- List ---

func TestTaskRepositoryPG_List_NoFilter(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	rows := addTaskRow(newTaskRows(), 1, "T1")

	mock.ExpectQuery("SELECT id, project_id").
		WillReturnRows(rows)

	tasks, err := repo.List(context.Background(), repositories.TaskFilter{}, 10, 0)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_List_AllFilters(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	projectID := int64(1)
	authorID := int64(2)
	assigneeID := int64(3)
	status := domain.TaskStatusInProgress
	priority := domain.TaskPriorityHigh
	isOverdue := true
	search := "test"
	filter := repositories.TaskFilter{
		ProjectID:  &projectID,
		AuthorID:   &authorID,
		AssigneeID: &assigneeID,
		Status:     &status,
		Priority:   &priority,
		IsOverdue:  &isOverdue,
		Search:     &search,
		Tags:       []string{"urgent"},
	}

	rows := addTaskRow(newTaskRows(), 1, "test task")

	mock.ExpectQuery("SELECT id, project_id").
		WillReturnRows(rows)

	tasks, err := repo.List(context.Background(), filter, 10, 0)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_List_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectQuery("SELECT id, project_id").
		WillReturnError(sql.ErrConnDone)

	tasks, err := repo.List(context.Background(), repositories.TaskFilter{}, 10, 0)
	require.Error(t, err)
	assert.Nil(t, tasks)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_List_ScanError(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	rows := sqlmock.NewRows(taskCols).AddRow(
		1, nil, "T", nil, nil, "not-number",
		nil, "new", "normal", nil, nil, nil,
		0, nil, nil, pq.StringArray{}, nil,
		time.Now(), time.Now(),
	)

	mock.ExpectQuery("SELECT id, project_id").
		WillReturnRows(rows)

	tasks, err := repo.List(context.Background(), repositories.TaskFilter{}, 10, 0)
	require.Error(t, err)
	assert.Nil(t, tasks)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_List_NoLimit(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	rows := addTaskRow(newTaskRows(), 1, "T1")

	mock.ExpectQuery("SELECT id, project_id").
		WillReturnRows(rows)

	tasks, err := repo.List(context.Background(), repositories.TaskFilter{}, 0, 0)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Count ---

func TestTaskRepositoryPG_Count_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM tasks")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

	count, err := repo.Count(context.Background(), repositories.TaskFilter{})
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_Count_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.Count(context.Background(), repositories.TaskFilter{})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByProject, GetByAuthor, GetByAssignee, GetByStatus, GetOverdueTasks ---

func TestTaskRepositoryPG_GetByProject(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	rows := addTaskRow(newTaskRows(), 1, "T1")
	mock.ExpectQuery("SELECT id, project_id").WillReturnRows(rows)

	tasks, err := repo.GetByProject(context.Background(), 1, 10, 0)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetByAuthor(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	rows := addTaskRow(newTaskRows(), 1, "T1")
	mock.ExpectQuery("SELECT id, project_id").WillReturnRows(rows)

	tasks, err := repo.GetByAuthor(context.Background(), 1, 10, 0)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetByAssignee(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	rows := addTaskRow(newTaskRows(), 1, "T1")
	mock.ExpectQuery("SELECT id, project_id").WillReturnRows(rows)

	tasks, err := repo.GetByAssignee(context.Background(), 1, 10, 0)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetByStatus(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	rows := addTaskRow(newTaskRows(), 1, "T1")
	mock.ExpectQuery("SELECT id, project_id").WillReturnRows(rows)

	tasks, err := repo.GetByStatus(context.Background(), domain.TaskStatusNew, 10, 0)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetOverdueTasks(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	rows := addTaskRow(newTaskRows(), 1, "T1")
	mock.ExpectQuery("SELECT id, project_id").WillReturnRows(rows)

	tasks, err := repo.GetOverdueTasks(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- AddWatcher ---

func TestTaskRepositoryPG_AddWatcher_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	watcher := entities.NewTaskWatcher(1, 2)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO task_watchers")).
		WithArgs(watcher.TaskID, watcher.UserID, watcher.CreatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.AddWatcher(context.Background(), watcher)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_AddWatcher_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	watcher := entities.NewTaskWatcher(1, 2)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO task_watchers")).
		WillReturnError(sql.ErrConnDone)

	err := repo.AddWatcher(context.Background(), watcher)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- RemoveWatcher ---

func TestTaskRepositoryPG_RemoveWatcher_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM task_watchers WHERE task_id = $1 AND user_id = $2")).
		WithArgs(int64(1), int64(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.RemoveWatcher(context.Background(), 1, 2)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetWatchers ---

func TestTaskRepositoryPG_GetWatchers_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"task_id", "user_id", "created_at"}).
		AddRow(int64(1), int64(2), now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, user_id, created_at FROM task_watchers")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	watchers, err := repo.GetWatchers(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, watchers, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetWatchers_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, user_id, created_at")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetWatchers(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetWatchers_ScanError(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	rows := sqlmock.NewRows([]string{"task_id", "user_id", "created_at"}).
		AddRow("bad", int64(2), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, user_id, created_at")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	_, err := repo.GetWatchers(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- IsWatching ---

func TestTaskRepositoryPG_IsWatching_True(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	watching, err := repo.IsWatching(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.True(t, watching)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_IsWatching_False(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	watching, err := repo.IsWatching(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.False(t, watching)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_IsWatching_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.IsWatching(context.Background(), 1, 2)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- AddAttachment ---

func TestTaskRepositoryPG_AddAttachment_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	att := entities.NewTaskAttachment(1, "file.txt", "/path/file.txt", 1024, 10)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO task_attachments")).
		WithArgs(att.TaskID, att.FileName, att.FilePath, att.FileSize, att.MimeType, att.UploadedBy, att.CreatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.AddAttachment(context.Background(), att)
	require.NoError(t, err)
	assert.Equal(t, int64(1), att.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_AddAttachment_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	att := entities.NewTaskAttachment(1, "file.txt", "/path", 1024, 10)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO task_attachments")).
		WillReturnError(sql.ErrConnDone)

	err := repo.AddAttachment(context.Background(), att)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- RemoveAttachment ---

func TestTaskRepositoryPG_RemoveAttachment_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM task_attachments WHERE id = $1")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.RemoveAttachment(context.Background(), 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetAttachments ---

func TestTaskRepositoryPG_GetAttachments_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	now := time.Now()
	mime := "text/plain"
	rows := sqlmock.NewRows([]string{"id", "task_id", "file_name", "file_path", "file_size", "mime_type", "uploaded_by", "created_at"}).
		AddRow(int64(1), int64(1), "f.txt", "/path", int64(100), &mime, int64(10), now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, task_id, file_name")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	attachments, err := repo.GetAttachments(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, attachments, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetAttachments_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, task_id")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetAttachments(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetAttachments_ScanError(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	rows := sqlmock.NewRows([]string{"id", "task_id", "file_name", "file_path", "file_size", "mime_type", "uploaded_by", "created_at"}).
		AddRow("bad", int64(1), "f.txt", "/path", int64(100), nil, int64(10), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, task_id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	_, err := repo.GetAttachments(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetAttachmentByID ---

func TestTaskRepositoryPG_GetAttachmentByID_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "task_id", "file_name", "file_path", "file_size", "mime_type", "uploaded_by", "created_at"}).
		AddRow(int64(1), int64(1), "f.txt", "/path", int64(100), nil, int64(10), now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, task_id, file_name")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	att, err := repo.GetAttachmentByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, att)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetAttachmentByID_NotFound(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, task_id")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	att, err := repo.GetAttachmentByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, att)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetAttachmentByID_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, task_id")).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetAttachmentByID(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- AddComment ---

func TestTaskRepositoryPG_AddComment_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	comment := entities.NewTaskComment(1, 2, "content")

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO task_comments")).
		WithArgs(comment.TaskID, comment.AuthorID, comment.Content,
			comment.ParentCommentID, comment.CreatedAt, comment.UpdatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.AddComment(context.Background(), comment)
	require.NoError(t, err)
	assert.Equal(t, int64(1), comment.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_AddComment_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	comment := entities.NewTaskComment(1, 2, "content")

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO task_comments")).
		WillReturnError(sql.ErrConnDone)

	err := repo.AddComment(context.Background(), comment)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- UpdateComment ---

func TestTaskRepositoryPG_UpdateComment_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	comment := entities.NewTaskComment(1, 2, "updated")
	comment.ID = 5

	mock.ExpectExec(regexp.QuoteMeta("UPDATE task_comments SET content")).
		WithArgs(comment.Content, comment.UpdatedAt, comment.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateComment(context.Background(), comment)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- DeleteComment ---

func TestTaskRepositoryPG_DeleteComment_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM task_comments WHERE id = $1")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteComment(context.Background(), 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetComments ---

func TestTaskRepositoryPG_GetComments_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "task_id", "author_id", "content", "parent_comment_id", "created_at", "updated_at"}).
		AddRow(int64(1), int64(1), int64(2), "hello", nil, now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, task_id, author_id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	comments, err := repo.GetComments(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, comments, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetComments_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, task_id")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetComments(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetComments_ScanError(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	rows := sqlmock.NewRows([]string{"id", "task_id", "author_id", "content", "parent_comment_id", "created_at", "updated_at"}).
		AddRow("bad", int64(1), int64(2), "hello", nil, time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, task_id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	_, err := repo.GetComments(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetCommentByID ---

func TestTaskRepositoryPG_GetCommentByID_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "task_id", "author_id", "content", "parent_comment_id", "created_at", "updated_at"}).
		AddRow(int64(1), int64(1), int64(2), "hello", nil, now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, task_id, author_id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	c, err := repo.GetCommentByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetCommentByID_NotFound(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, task_id")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	c, err := repo.GetCommentByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, c)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- AddChecklist ---

func TestTaskRepositoryPG_AddChecklist_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	cl := entities.NewTaskChecklist(1, "Checklist", 0)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO task_checklists")).
		WithArgs(cl.TaskID, cl.Title, cl.Position, cl.CreatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.AddChecklist(context.Background(), cl)
	require.NoError(t, err)
	assert.Equal(t, int64(1), cl.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- UpdateChecklist ---

func TestTaskRepositoryPG_UpdateChecklist_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	cl := &entities.TaskChecklist{ID: 1, Title: "Updated", Position: 1}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE task_checklists SET title")).
		WithArgs(cl.Title, cl.Position, cl.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateChecklist(context.Background(), cl)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- DeleteChecklist ---

func TestTaskRepositoryPG_DeleteChecklist_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM task_checklists WHERE id = $1")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteChecklist(context.Background(), 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetChecklists ---

func TestTaskRepositoryPG_GetChecklists_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "task_id", "title", "position", "created_at"}).
		AddRow(int64(1), int64(1), "CL", 0, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, task_id, title, position, created_at")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	cls, err := repo.GetChecklists(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, cls, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetChecklists_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, task_id")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetChecklists(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetChecklists_ScanError(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	rows := sqlmock.NewRows([]string{"id", "task_id", "title", "position", "created_at"}).
		AddRow("bad", int64(1), "CL", 0, time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, task_id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	_, err := repo.GetChecklists(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- AddChecklistItem ---

func TestTaskRepositoryPG_AddChecklistItem_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	item := entities.NewTaskChecklistItem(1, "Item", 0)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO task_checklist_items")).
		WithArgs(item.ChecklistID, item.Title, item.IsCompleted, item.Position, item.CreatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.AddChecklistItem(context.Background(), item)
	require.NoError(t, err)
	assert.Equal(t, int64(1), item.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- UpdateChecklistItem ---

func TestTaskRepositoryPG_UpdateChecklistItem_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	item := &entities.TaskChecklistItem{ID: 1, Title: "Updated", IsCompleted: true, Position: 1}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE task_checklist_items SET")).
		WithArgs(item.Title, item.IsCompleted, item.Position, item.CompletedBy, item.CompletedAt, item.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateChecklistItem(context.Background(), item)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- DeleteChecklistItem ---

func TestTaskRepositoryPG_DeleteChecklistItem_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM task_checklist_items WHERE id = $1")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteChecklistItem(context.Background(), 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetChecklistItems ---

func TestTaskRepositoryPG_GetChecklistItems_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "checklist_id", "title", "is_completed", "position", "completed_by", "completed_at", "created_at"}).
		AddRow(int64(1), int64(1), "Item", false, 0, nil, nil, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, checklist_id, title")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	items, err := repo.GetChecklistItems(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetChecklistItems_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, checklist_id")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetChecklistItems(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetChecklistItems_ScanError(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	rows := sqlmock.NewRows([]string{"id", "checklist_id", "title", "is_completed", "position", "completed_by", "completed_at", "created_at"}).
		AddRow("bad", int64(1), "Item", false, 0, nil, nil, time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, checklist_id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	_, err := repo.GetChecklistItems(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- AddHistory ---

func TestTaskRepositoryPG_AddHistory_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	userID := int64(2)
	oldVal := "old"
	newVal := "new"
	h := entities.NewTaskHistory(1, &userID, "status", &oldVal, &newVal)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO task_history")).
		WithArgs(h.TaskID, h.UserID, h.FieldName, h.OldValue, h.NewValue, h.CreatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.AddHistory(context.Background(), h)
	require.NoError(t, err)
	assert.Equal(t, int64(1), h.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_AddHistory_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	h := entities.NewTaskHistory(1, nil, "status", nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO task_history")).
		WillReturnError(sql.ErrConnDone)

	err := repo.AddHistory(context.Background(), h)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetHistory ---

func TestTaskRepositoryPG_GetHistory_Success(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	now := time.Now()
	userID := int64(2)
	rows := sqlmock.NewRows([]string{"id", "task_id", "user_id", "field_name", "old_value", "new_value", "created_at"}).
		AddRow(int64(1), int64(1), &userID, "status", nil, nil, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, task_id, user_id")).
		WithArgs(int64(1), 10, 0).
		WillReturnRows(rows)

	history, err := repo.GetHistory(context.Background(), 1, 10, 0)
	require.NoError(t, err)
	assert.Len(t, history, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetHistory_Error(t *testing.T) {
	repo, mock := newTaskRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, task_id")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetHistory(context.Background(), 1, 10, 0)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskRepositoryPG_GetHistory_ScanError(t *testing.T) {
	repo, mock := newTaskRepoMock(t)
	rows := sqlmock.NewRows([]string{"id", "task_id", "user_id", "field_name", "old_value", "new_value", "created_at"}).
		AddRow("bad", int64(1), nil, "status", nil, nil, time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, task_id")).
		WithArgs(int64(1), 10, 0).
		WillReturnRows(rows)

	_, err := repo.GetHistory(context.Background(), 1, 10, 0)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
