// Package domain provides domain types and entities for the tasks module.
package domain

// TaskStatus represents the status of a task.
type TaskStatus string

const (
	TaskStatusNew        TaskStatus = "new"
	TaskStatusAssigned   TaskStatus = "assigned"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusReview     TaskStatus = "review"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
	TaskStatusDeferred   TaskStatus = "deferred"
)

// IsValid checks if the task status is valid.
func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskStatusNew, TaskStatusAssigned, TaskStatusInProgress,
		TaskStatusReview, TaskStatusCompleted, TaskStatusCancelled, TaskStatusDeferred:
		return true
	}
	return false
}

// TaskPriority represents the priority level of a task.
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityNormal TaskPriority = "normal"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

// IsValid checks if the task priority is valid.
func (p TaskPriority) IsValid() bool {
	switch p {
	case TaskPriorityLow, TaskPriorityNormal, TaskPriorityHigh, TaskPriorityUrgent:
		return true
	}
	return false
}

// ProjectStatus represents the status of a project.
type ProjectStatus string

const (
	ProjectStatusPlanning  ProjectStatus = "planning"
	ProjectStatusActive    ProjectStatus = "active"
	ProjectStatusOnHold    ProjectStatus = "on_hold"
	ProjectStatusCompleted ProjectStatus = "completed"
	ProjectStatusCancelled ProjectStatus = "cancelled"
)

// IsValid checks if the project status is valid.
func (s ProjectStatus) IsValid() bool {
	switch s {
	case ProjectStatusPlanning, ProjectStatusActive, ProjectStatusOnHold,
		ProjectStatusCompleted, ProjectStatusCancelled:
		return true
	}
	return false
}

// DependencyType represents the type of dependency between tasks.
type DependencyType string

const (
	DependencyTypeFinishToStart  DependencyType = "finish_to_start"
	DependencyTypeStartToStart   DependencyType = "start_to_start"
	DependencyTypeFinishToFinish DependencyType = "finish_to_finish"
	DependencyTypeStartToFinish  DependencyType = "start_to_finish"
)

// IsValid checks if the dependency type is valid.
func (d DependencyType) IsValid() bool {
	switch d {
	case DependencyTypeFinishToStart, DependencyTypeStartToStart,
		DependencyTypeFinishToFinish, DependencyTypeStartToFinish:
		return true
	}
	return false
}
