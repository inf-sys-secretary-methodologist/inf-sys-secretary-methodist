package entities

import "testing"

func TestSyncStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   SyncStatus
		expected string
	}{
		{"pending", SyncStatusPending, "pending"},
		{"in_progress", SyncStatusInProgress, "in_progress"},
		{"completed", SyncStatusCompleted, "completed"},
		{"failed", SyncStatusFailed, "failed"},
		{"canceled", SyncStatusCancelled, "canceled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.status)
			}
		})
	}
}

func TestSyncDirectionConstants(t *testing.T) {
	tests := []struct {
		name      string
		direction SyncDirection
		expected  string
	}{
		{"import", SyncDirectionImport, "import"},
		{"export", SyncDirectionExport, "export"},
		{"both", SyncDirectionBoth, "both"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.direction) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.direction)
			}
		})
	}
}

func TestSyncEntityTypeConstants(t *testing.T) {
	tests := []struct {
		name       string
		entityType SyncEntityType
		expected   string
	}{
		{"employee", SyncEntityEmployee, "employee"},
		{"student", SyncEntityStudent, "student"},
		{"finance", SyncEntityFinance, "finance"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.entityType) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.entityType)
			}
		})
	}
}

func TestConflictResolutionConstants(t *testing.T) {
	tests := []struct {
		name       string
		resolution ConflictResolution
		expected   string
	}{
		{"pending", ConflictResolutionPending, "pending"},
		{"use_local", ConflictResolutionUseLocal, "use_local"},
		{"use_external", ConflictResolutionUseExternal, "use_external"},
		{"merge", ConflictResolutionMerge, "merge"},
		{"skip", ConflictResolutionSkip, "skip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.resolution) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.resolution)
			}
		})
	}
}
