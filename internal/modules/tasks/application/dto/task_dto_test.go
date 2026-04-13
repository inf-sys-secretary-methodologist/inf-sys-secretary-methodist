package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToTaskOutput_Basic(t *testing.T) {
	now := time.Now()
	desc := "A task"
	dueDate := now.Add(24 * time.Hour)
	assigneeID := int64(10)

	task := &entities.Task{
		ID:          1,
		Title:       "Test Task",
		Description: &desc,
		AuthorID:    42,
		AssigneeID:  &assigneeID,
		Status:      domain.TaskStatusNew,
		Priority:    domain.TaskPriorityHigh,
		DueDate:     &dueDate,
		Progress:    50,
		Tags:        []string{"urgent"},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	output := ToTaskOutput(task)

	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, "Test Task", output.Title)
	assert.Equal(t, &desc, output.Description)
	assert.Equal(t, int64(42), output.AuthorID)
	assert.Equal(t, &assigneeID, output.AssigneeID)
	assert.Equal(t, "new", output.Status)
	assert.Equal(t, "high", output.Priority)
	assert.Equal(t, &dueDate, output.DueDate)
	assert.Equal(t, 50, output.Progress)
	assert.Equal(t, []string{"urgent"}, output.Tags)
}

func TestToTaskOutput_WithProject(t *testing.T) {
	now := time.Now()
	task := &entities.Task{
		ID:        1,
		Title:     "Task",
		AuthorID:  1,
		Status:    domain.TaskStatusNew,
		Priority:  domain.TaskPriorityNormal,
		CreatedAt: now,
		UpdatedAt: now,
		Project: &entities.Project{
			ID:      10,
			Name:    "Project A",
			OwnerID: 1,
			Status:  domain.ProjectStatusActive,
		},
	}

	output := ToTaskOutput(task)

	require.NotNil(t, output.Project)
	assert.Equal(t, int64(10), output.Project.ID)
	assert.Equal(t, "Project A", output.Project.Name)
}

func TestToTaskOutput_WithAssignee(t *testing.T) {
	now := time.Now()
	assigneeID := int64(10)
	task := &entities.Task{
		ID:         1,
		Title:      "Task",
		AuthorID:   1,
		AssigneeID: &assigneeID,
		Status:     domain.TaskStatusNew,
		Priority:   domain.TaskPriorityLow,
		Assignee:   &entities.TaskAssignee{ID: 10, Name: "Bob", Email: "bob@example.com"},
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	output := ToTaskOutput(task)

	require.NotNil(t, output.Assignee)
	assert.Equal(t, int64(10), output.Assignee.ID)
	assert.Equal(t, "Bob", output.Assignee.Name)
}

func TestToTaskOutput_WithWatchers(t *testing.T) {
	now := time.Now()
	task := &entities.Task{
		ID: 1, Title: "Task", AuthorID: 1,
		Status: domain.TaskStatusNew, Priority: domain.TaskPriorityNormal,
		Watchers: []entities.TaskWatcher{
			{UserID: 10, User: &entities.WatcherUser{ID: 10, Name: "Alice", Email: "alice@example.com"}},
		},
		CreatedAt: now, UpdatedAt: now,
	}

	output := ToTaskOutput(task)

	require.Len(t, output.Watchers, 1)
	assert.Equal(t, "Alice", output.Watchers[0].Name)
}

func TestToTaskCommentOutput(t *testing.T) {
	now := time.Now()
	comment := &entities.TaskComment{
		ID:        1,
		TaskID:    10,
		AuthorID:  42,
		Content:   "Great work!",
		Author:    &entities.CommentAuthor{ID: 42, Name: "Admin", Email: "admin@example.com"},
		CreatedAt: now,
		UpdatedAt: now,
	}

	output := ToTaskCommentOutput(comment)

	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, int64(10), output.TaskID)
	assert.Equal(t, "Great work!", output.Content)
	require.NotNil(t, output.Author)
	assert.Equal(t, "Admin", output.Author.Name)
}

func TestToTaskCommentOutput_WithReplies(t *testing.T) {
	now := time.Now()
	parentID := int64(1)
	comment := &entities.TaskComment{
		ID: 1, TaskID: 10, AuthorID: 42, Content: "Top",
		Replies: []entities.TaskComment{
			{ID: 2, TaskID: 10, AuthorID: 43, Content: "Reply", ParentCommentID: &parentID, CreatedAt: now, UpdatedAt: now},
		},
		CreatedAt: now, UpdatedAt: now,
	}

	output := ToTaskCommentOutput(comment)

	require.Len(t, output.Replies, 1)
	assert.Equal(t, "Reply", output.Replies[0].Content)
}

func TestToTaskAttachmentOutput(t *testing.T) {
	now := time.Now()
	mimeType := "application/pdf"
	att := &entities.TaskAttachment{
		ID:         1,
		TaskID:     10,
		FileName:   "doc.pdf",
		FilePath:   "/files/doc.pdf",
		FileSize:   1024,
		MimeType:   &mimeType,
		UploadedBy: 42,
		CreatedAt:  now,
	}

	output := ToTaskAttachmentOutput(att)

	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, int64(10), output.TaskID)
	assert.Equal(t, "doc.pdf", output.FileName)
	assert.Equal(t, int64(1024), output.FileSize)
	assert.Equal(t, &mimeType, output.MimeType)
}

func TestToTaskChecklistOutput(t *testing.T) {
	now := time.Now()
	completedBy := int64(42)
	completedAt := now

	checklist := &entities.TaskChecklist{
		ID:        1,
		TaskID:    10,
		Title:     "Steps",
		Position:  1,
		Items: []entities.TaskChecklistItem{
			{ID: 1, ChecklistID: 1, Title: "Step 1", IsCompleted: true, Position: 1,
				CompletedBy: &completedBy, CompletedAt: &completedAt, CreatedAt: now},
			{ID: 2, ChecklistID: 1, Title: "Step 2", IsCompleted: false, Position: 2, CreatedAt: now},
		},
		CreatedAt: now,
	}

	output := ToTaskChecklistOutput(checklist)

	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, "Steps", output.Title)
	assert.Equal(t, 50, output.CompletionPercentage) // 1 out of 2
	require.Len(t, output.Items, 2)
	assert.True(t, output.Items[0].IsCompleted)
	assert.False(t, output.Items[1].IsCompleted)
}

func TestToTaskHistoryOutput(t *testing.T) {
	now := time.Now()
	oldVal := "open"
	newVal := "in_progress"
	history := &entities.TaskHistory{
		ID:        1,
		TaskID:    10,
		UserID:    ptrInt64(42),
		FieldName: "status",
		OldValue:  &oldVal,
		NewValue:  &newVal,
		User:      &entities.HistoryUser{ID: 42, Name: "Admin", Email: "admin@example.com"},
		CreatedAt: now,
	}

	output := ToTaskHistoryOutput(history)

	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, "status", output.FieldName)
	assert.Equal(t, &oldVal, output.OldValue)
	assert.Equal(t, &newVal, output.NewValue)
	require.NotNil(t, output.User)
	assert.Equal(t, "Admin", output.User.Name)
}

func TestTaskFilterInput_ToTaskFilter(t *testing.T) {
	status := "open"
	priority := "high"
	projectID := int64(5)
	isOverdue := true
	search := "test"
	f := &TaskFilterInput{
		ProjectID: &projectID,
		Status:    &status,
		Priority:  &priority,
		IsOverdue: &isOverdue,
		Search:    &search,
		Tags:      []string{"urgent"},
	}

	filter := f.ToTaskFilter()

	assert.Equal(t, &projectID, filter.ProjectID)
	require.NotNil(t, filter.Status)
	assert.Equal(t, domain.TaskStatus("open"), *filter.Status)
	require.NotNil(t, filter.Priority)
	assert.Equal(t, domain.TaskPriority("high"), *filter.Priority)
	assert.Equal(t, &isOverdue, filter.IsOverdue)
	assert.Equal(t, &search, filter.Search)
	assert.Equal(t, []string{"urgent"}, filter.Tags)
}

func ptrInt64(v int64) *int64 {
	return &v
}
