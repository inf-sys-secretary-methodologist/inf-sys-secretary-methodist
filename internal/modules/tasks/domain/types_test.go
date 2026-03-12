package domain

import "testing"

func TestTaskStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status TaskStatus
		want   bool
	}{
		{"new", TaskStatusNew, true},
		{"assigned", TaskStatusAssigned, true},
		{"in_progress", TaskStatusInProgress, true},
		{"review", TaskStatusReview, true},
		{"completed", TaskStatusCompleted, true},
		{"canceled", TaskStatusCancelled, true},
		{"deferred", TaskStatusDeferred, true},
		{"invalid", TaskStatus("invalid"), false},
		{"empty", TaskStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskPriority_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		priority TaskPriority
		want     bool
	}{
		{"low", TaskPriorityLow, true},
		{"normal", TaskPriorityNormal, true},
		{"high", TaskPriorityHigh, true},
		{"urgent", TaskPriorityUrgent, true},
		{"invalid", TaskPriority("invalid"), false},
		{"empty", TaskPriority(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.priority.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   TaskStatus
		expected string
	}{
		{"new", TaskStatusNew, "new"},
		{"assigned", TaskStatusAssigned, "assigned"},
		{"in_progress", TaskStatusInProgress, "in_progress"},
		{"review", TaskStatusReview, "review"},
		{"completed", TaskStatusCompleted, "completed"},
		{"canceled", TaskStatusCancelled, "canceled"},
		{"deferred", TaskStatusDeferred, "deferred"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.status)
			}
		})
	}
}

func TestTaskPriorityConstants(t *testing.T) {
	tests := []struct {
		name     string
		priority TaskPriority
		expected string
	}{
		{"low", TaskPriorityLow, "low"},
		{"normal", TaskPriorityNormal, "normal"},
		{"high", TaskPriorityHigh, "high"},
		{"urgent", TaskPriorityUrgent, "urgent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.priority) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.priority)
			}
		})
	}
}
