package entities

import "testing"

func TestNewSyncLog(t *testing.T) {
	entityType := SyncEntityEmployee
	direction := SyncDirectionImport

	log := NewSyncLog(entityType, direction)

	if log.EntityType != entityType {
		t.Errorf("expected entity type %q, got %q", entityType, log.EntityType)
	}
	if log.Direction != direction {
		t.Errorf("expected direction %q, got %q", direction, log.Direction)
	}
	if log.Status != SyncStatusPending {
		t.Errorf("expected status %q, got %q", SyncStatusPending, log.Status)
	}
	if log.TotalRecords != 0 {
		t.Errorf("expected total records 0, got %d", log.TotalRecords)
	}
	if log.ProcessedCount != 0 {
		t.Errorf("expected processed count 0, got %d", log.ProcessedCount)
	}
}

func TestSyncLog_Start(t *testing.T) {
	log := NewSyncLog(SyncEntityEmployee, SyncDirectionImport)

	log.Start()

	if log.Status != SyncStatusInProgress {
		t.Errorf("expected status %q, got %q", SyncStatusInProgress, log.Status)
	}
}

func TestSyncLog_Complete(t *testing.T) {
	log := NewSyncLog(SyncEntityEmployee, SyncDirectionImport)
	log.Start()

	log.Complete()

	if log.Status != SyncStatusCompleted {
		t.Errorf("expected status %q, got %q", SyncStatusCompleted, log.Status)
	}
	if log.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
}

func TestSyncLog_Fail(t *testing.T) {
	log := NewSyncLog(SyncEntityEmployee, SyncDirectionImport)
	log.Start()
	errMsg := "connection failed"

	log.Fail(errMsg)

	if log.Status != SyncStatusFailed {
		t.Errorf("expected status %q, got %q", SyncStatusFailed, log.Status)
	}
	if log.ErrorMessage != errMsg {
		t.Errorf("expected error message %q, got %q", errMsg, log.ErrorMessage)
	}
	if log.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
}

func TestSyncLog_Cancel(t *testing.T) {
	log := NewSyncLog(SyncEntityEmployee, SyncDirectionImport)
	log.Start()

	log.Cancel()

	if log.Status != SyncStatusCancelled {
		t.Errorf("expected status %q, got %q", SyncStatusCancelled, log.Status)
	}
}

func TestSyncLog_IncrementProcessed(t *testing.T) {
	log := NewSyncLog(SyncEntityEmployee, SyncDirectionImport)

	log.IncrementProcessed()
	log.IncrementProcessed()
	log.IncrementProcessed()

	if log.ProcessedCount != 3 {
		t.Errorf("expected processed count 3, got %d", log.ProcessedCount)
	}
}

func TestSyncLog_IncrementSuccess(t *testing.T) {
	log := NewSyncLog(SyncEntityEmployee, SyncDirectionImport)

	log.IncrementSuccess()
	log.IncrementSuccess()

	if log.SuccessCount != 2 {
		t.Errorf("expected success count 2, got %d", log.SuccessCount)
	}
}

func TestSyncLog_IncrementError(t *testing.T) {
	log := NewSyncLog(SyncEntityEmployee, SyncDirectionImport)

	log.IncrementError()

	if log.ErrorCount != 1 {
		t.Errorf("expected error count 1, got %d", log.ErrorCount)
	}
}

func TestSyncLog_IncrementConflict(t *testing.T) {
	log := NewSyncLog(SyncEntityEmployee, SyncDirectionImport)

	log.IncrementConflict()
	log.IncrementConflict()

	if log.ConflictCount != 2 {
		t.Errorf("expected conflict count 2, got %d", log.ConflictCount)
	}
}

func TestSyncLog_SetTotalRecords(t *testing.T) {
	log := NewSyncLog(SyncEntityEmployee, SyncDirectionImport)

	log.SetTotalRecords(100)

	if log.TotalRecords != 100 {
		t.Errorf("expected total records 100, got %d", log.TotalRecords)
	}
}

func TestSyncLog_GetProgress(t *testing.T) {
	tests := []struct {
		name      string
		total     int
		processed int
		want      float64
	}{
		{
			name:      "50% progress",
			total:     100,
			processed: 50,
			want:      50.0,
		},
		{
			name:      "zero total",
			total:     0,
			processed: 0,
			want:      0.0,
		},
		{
			name:      "100% complete",
			total:     100,
			processed: 100,
			want:      100.0,
		},
		{
			name:      "partial progress",
			total:     200,
			processed: 50,
			want:      25.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := NewSyncLog(SyncEntityEmployee, SyncDirectionImport)
			log.TotalRecords = tt.total
			log.ProcessedCount = tt.processed

			got := log.GetProgress()
			if got != tt.want {
				t.Errorf("GetProgress() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestSyncLog_IsRunning(t *testing.T) {
	log := NewSyncLog(SyncEntityEmployee, SyncDirectionImport)

	if log.IsRunning() {
		t.Error("new log should not be running")
	}

	log.Start()

	if !log.IsRunning() {
		t.Error("started log should be running")
	}

	log.Complete()

	if log.IsRunning() {
		t.Error("completed log should not be running")
	}
}
