package api

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
)

// Task represents a task resource.
type Task struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	AuthorID    int64  `json:"author_id"`
	AssigneeID  int64  `json:"assignee_id"`
	DueDate     string `json:"due_date"`
	Progress    int    `json:"progress"`
}

// TaskList represents a paginated list of tasks.
type TaskList struct {
	Tasks []Task `json:"tasks"`
	Total int    `json:"total"`
}

// TaskComment represents a comment on a task.
type TaskComment struct {
	ID       int64  `json:"id"`
	TaskID   int64  `json:"task_id"`
	AuthorID int64  `json:"author_id"`
	Content  string `json:"content"`
}

// CreateTaskRequest represents a request to create a new task.
type CreateTaskRequest struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	AssigneeID     int64  `json:"assignee_id,omitempty"`
	Priority       string `json:"priority,omitempty"`
	DueDate        string `json:"due_date,omitempty"`
	EstimatedHours int    `json:"estimated_hours,omitempty"`
}

// CreateTask creates a new task.
func (c *Client) CreateTask(ctx context.Context, a *agent.Agent, req CreateTaskRequest) (*Task, error) {
	resp, err := c.Post(ctx, "/api/tasks", a, req)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	var task Task
	if err := ParseData(resp, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// ListTasks retrieves tasks.
func (c *Client) ListTasks(ctx context.Context, a *agent.Agent, queryParams string) (*TaskList, error) {
	path := "/api/tasks"
	if queryParams != "" {
		path += "?" + queryParams
	}
	resp, err := c.Get(ctx, path, a)
	if err != nil {
		return nil, err
	}
	var list TaskList
	if err := ParseData(resp, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// GetTask retrieves a task by ID.
func (c *Client) GetTask(ctx context.Context, a *agent.Agent, id int64) (*Task, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/tasks/%d", id), a)
	if err != nil {
		return nil, err
	}
	var task Task
	if err := ParseData(resp, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// AssignTask assigns a task to a user.
func (c *Client) AssignTask(ctx context.Context, a *agent.Agent, taskID int64, assigneeID int64) error {
	body := map[string]any{
		"assignee_id": assigneeID,
	}
	_, err := c.Post(ctx, fmt.Sprintf("/api/tasks/%d/assign", taskID), a, body)
	return err
}

// StartTask starts work on a task.
func (c *Client) StartTask(ctx context.Context, a *agent.Agent, taskID int64) error {
	_, err := c.Post(ctx, fmt.Sprintf("/api/tasks/%d/start", taskID), a, nil)
	return err
}

// CompleteTask marks a task as complete.
func (c *Client) CompleteTask(ctx context.Context, a *agent.Agent, taskID int64) error {
	_, err := c.Post(ctx, fmt.Sprintf("/api/tasks/%d/complete", taskID), a, nil)
	return err
}

// AddTaskComment adds a comment to a task.
func (c *Client) AddTaskComment(ctx context.Context, a *agent.Agent, taskID int64, content string) (*TaskComment, error) {
	body := map[string]any{
		"content": content,
	}
	resp, err := c.Post(ctx, fmt.Sprintf("/api/tasks/%d/comments", taskID), a, body)
	if err != nil {
		return nil, err
	}
	var comment TaskComment
	if err := ParseData(resp, &comment); err != nil {
		return nil, err
	}
	return &comment, nil
}

// AddChecklist adds a checklist to a task.
func (c *Client) AddChecklist(ctx context.Context, a *agent.Agent, taskID int64, title string) error {
	body := map[string]any{
		"title": title,
	}
	_, err := c.Post(ctx, fmt.Sprintf("/api/tasks/%d/checklists", taskID), a, body)
	return err
}
