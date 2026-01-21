package entities

import "testing"

func TestNewSyncConflict(t *testing.T) {
	syncLogID := int64(42)
	entityType := SyncEntityEmployee
	entityID := "ext-123"

	conflict := NewSyncConflict(syncLogID, entityType, entityID)

	if conflict.SyncLogID != syncLogID {
		t.Errorf("expected sync log ID %d, got %d", syncLogID, conflict.SyncLogID)
	}
	if conflict.EntityType != entityType {
		t.Errorf("expected entity type %q, got %q", entityType, conflict.EntityType)
	}
	if conflict.EntityID != entityID {
		t.Errorf("expected entity ID %q, got %q", entityID, conflict.EntityID)
	}
	if conflict.Resolution != ConflictResolutionPending {
		t.Errorf("expected resolution %q, got %q", ConflictResolutionPending, conflict.Resolution)
	}
	if conflict.ConflictFields == nil {
		t.Error("expected conflict fields to be initialized")
	}
	if conflict.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
	if conflict.UpdatedAt.IsZero() {
		t.Error("expected updated_at to be set")
	}
}

func TestSyncConflict_Resolve(t *testing.T) {
	conflict := NewSyncConflict(1, SyncEntityEmployee, "ext-1")
	userID := int64(99)
	resolvedData := `{"merged": true}`

	conflict.Resolve(ConflictResolutionMerge, userID, resolvedData)

	if conflict.Resolution != ConflictResolutionMerge {
		t.Errorf("expected resolution %q, got %q", ConflictResolutionMerge, conflict.Resolution)
	}
	if conflict.ResolvedBy == nil || *conflict.ResolvedBy != userID {
		t.Errorf("expected resolved by %d, got %v", userID, conflict.ResolvedBy)
	}
	if conflict.ResolvedAt == nil {
		t.Error("expected resolved_at to be set")
	}
	if conflict.ResolvedData != resolvedData {
		t.Errorf("expected resolved data %q, got %q", resolvedData, conflict.ResolvedData)
	}
}

func TestSyncConflict_ResolveWithDifferentResolutions(t *testing.T) {
	tests := []struct {
		name       string
		resolution ConflictResolution
	}{
		{"use local", ConflictResolutionUseLocal},
		{"use external", ConflictResolutionUseExternal},
		{"merge", ConflictResolutionMerge},
		{"skip", ConflictResolutionSkip},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflict := NewSyncConflict(1, SyncEntityStudent, "ext-2")
			conflict.Resolve(tt.resolution, 1, "{}")

			if conflict.Resolution != tt.resolution {
				t.Errorf("expected resolution %q, got %q", tt.resolution, conflict.Resolution)
			}
		})
	}
}

func TestSyncConflict_IsPending(t *testing.T) {
	conflict := NewSyncConflict(1, SyncEntityEmployee, "ext-1")

	if !conflict.IsPending() {
		t.Error("new conflict should be pending")
	}

	conflict.Resolve(ConflictResolutionUseLocal, 1, "{}")

	if conflict.IsPending() {
		t.Error("resolved conflict should not be pending")
	}
}

func TestSyncConflict_SetLocalData(t *testing.T) {
	conflict := NewSyncConflict(1, SyncEntityEmployee, "ext-1")
	originalUpdateTime := conflict.UpdatedAt
	localData := `{"name": "Local Name"}`

	conflict.SetLocalData(localData)

	if conflict.LocalData != localData {
		t.Errorf("expected local data %q, got %q", localData, conflict.LocalData)
	}
	if !conflict.UpdatedAt.After(originalUpdateTime) && !conflict.UpdatedAt.Equal(originalUpdateTime) {
		t.Error("expected updated_at to be updated")
	}
}

func TestSyncConflict_SetExternalData(t *testing.T) {
	conflict := NewSyncConflict(1, SyncEntityEmployee, "ext-1")
	externalData := `{"name": "External Name"}`

	conflict.SetExternalData(externalData)

	if conflict.ExternalData != externalData {
		t.Errorf("expected external data %q, got %q", externalData, conflict.ExternalData)
	}
}

func TestSyncConflict_SetConflictFields(t *testing.T) {
	conflict := NewSyncConflict(1, SyncEntityEmployee, "ext-1")
	fields := []string{"name", "email", "phone"}

	conflict.SetConflictFields(fields)

	if len(conflict.ConflictFields) != len(fields) {
		t.Errorf("expected %d conflict fields, got %d", len(fields), len(conflict.ConflictFields))
	}
	for i, field := range fields {
		if conflict.ConflictFields[i] != field {
			t.Errorf("expected field %q at index %d, got %q", field, i, conflict.ConflictFields[i])
		}
	}
}

func TestConflictTypes(t *testing.T) {
	if ConflictTypeUpdate != "update" {
		t.Errorf("expected ConflictTypeUpdate to be %q, got %q", "update", ConflictTypeUpdate)
	}
	if ConflictTypeDelete != "delete" {
		t.Errorf("expected ConflictTypeDelete to be %q, got %q", "delete", ConflictTypeDelete)
	}
	if ConflictTypeCreate != "create" {
		t.Errorf("expected ConflictTypeCreate to be %q, got %q", "create", ConflictTypeCreate)
	}
}
