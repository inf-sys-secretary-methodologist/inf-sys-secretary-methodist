package entities

import "time"

// TaskComment represents a comment on a task.
type TaskComment struct {
	ID              int64     `json:"id"`
	TaskID          int64     `json:"task_id"`
	AuthorID        int64     `json:"author_id"`
	Content         string    `json:"content"`
	ParentCommentID *int64    `json:"parent_comment_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Associations
	Author  *CommentAuthor `json:"author,omitempty"`
	Replies []TaskComment  `json:"replies,omitempty"`
}

// CommentAuthor represents basic author info for comment response.
type CommentAuthor struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// NewTaskComment creates a new task comment.
func NewTaskComment(taskID, authorID int64, content string) *TaskComment {
	now := time.Now()
	return &TaskComment{
		TaskID:    taskID,
		AuthorID:  authorID,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// SetParent sets the parent comment for a reply.
func (c *TaskComment) SetParent(parentID int64) {
	c.ParentCommentID = &parentID
}

// Update updates the comment content.
func (c *TaskComment) Update(content string) {
	c.Content = content
	c.UpdatedAt = time.Now()
}
