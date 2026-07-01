package domain

// TaskFilter defines filtering options for task queries.
type TaskFilter struct {
	ProjectID  *int64
	AuthorID   *int64
	AssigneeID *int64
	Status     *TaskStatus
	Priority   *TaskPriority
	IsOverdue  *bool
	Search     *string
	Tags       []string
}
