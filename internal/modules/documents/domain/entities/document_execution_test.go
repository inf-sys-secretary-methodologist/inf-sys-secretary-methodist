package entities

import (
	"errors"
	"testing"
	"time"
)

// TestDocumentAssignExecutor_TableDriven pins the v0.151.0 AssignExecutor
// shape gate:
//   - execution + dueDate=nil → assigned (no status change; due nil)
//   - execution + dueDate=set → assigned (due captured)
//   - execution + reassign overwrites prior executor
//   - any non-execution → ErrCannotAssignExecutor (no field mutation)
//
// Issue: #232
func TestDocumentAssignExecutor_TableDriven(t *testing.T) {
	now := time.Date(2026, 5, 17, 11, 0, 0, 0, time.UTC)
	due := time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC)
	executorID := int64(13)
	actorID := int64(7)

	cases := []struct {
		name       string
		startState DocumentStatus
		dueDate    *time.Time
		wantErr    error
		wantDue    *time.Time
	}{
		{name: "execution + nil due → assigned", startState: DocumentStatusExecution, dueDate: nil, wantDue: nil},
		{name: "execution + due → assigned", startState: DocumentStatusExecution, dueDate: &due, wantDue: &due},
		{name: "draft → err", startState: DocumentStatusDraft, dueDate: nil, wantErr: ErrCannotAssignExecutor},
		{name: "approval → err", startState: DocumentStatusApproval, dueDate: nil, wantErr: ErrCannotAssignExecutor},
		{name: "approved → err", startState: DocumentStatusApproved, dueDate: nil, wantErr: ErrCannotAssignExecutor},
		{name: "registered → err", startState: DocumentStatusRegistered, dueDate: nil, wantErr: ErrCannotAssignExecutor},
		{name: "routing → err", startState: DocumentStatusRouting, dueDate: nil, wantErr: ErrCannotAssignExecutor},
		{name: "executed → err (already past)", startState: DocumentStatusExecuted, dueDate: nil, wantErr: ErrCannotAssignExecutor},
		{name: "rejected → err", startState: DocumentStatusRejected, dueDate: nil, wantErr: ErrCannotAssignExecutor},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDocument("План", 1, 42)
			d.Status = tc.startState
			err := d.AssignExecutor(executorID, tc.dueDate, actorID, now)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				if d.Status != tc.startState {
					t.Errorf("status mutated on rejected call: %q → %q", tc.startState, d.Status)
				}
				if d.ExecutorAssignedTo != nil {
					t.Errorf("ExecutorAssignedTo mutated on rejected call: %v", d.ExecutorAssignedTo)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.Status != DocumentStatusExecution {
				t.Errorf("status changed unexpectedly: %q (should stay execution)", d.Status)
			}
			if d.ExecutorAssignedTo == nil || *d.ExecutorAssignedTo != executorID {
				t.Errorf("expected ExecutorAssignedTo=%d, got %v", executorID, d.ExecutorAssignedTo)
			}
			if d.ExecutorAssignedAt == nil || !d.ExecutorAssignedAt.Equal(now) {
				t.Errorf("expected ExecutorAssignedAt=%v, got %v", now, d.ExecutorAssignedAt)
			}
			if tc.wantDue == nil && d.ExecutorDueDate != nil {
				t.Errorf("expected ExecutorDueDate=nil, got %v", d.ExecutorDueDate)
			}
			if tc.wantDue != nil && (d.ExecutorDueDate == nil || !d.ExecutorDueDate.Equal(*tc.wantDue)) {
				t.Errorf("expected ExecutorDueDate=%v, got %v", tc.wantDue, d.ExecutorDueDate)
			}
		})
	}
}

// TestDocumentAssignExecutor_Reassign confirms admin can overwrite a
// prior assignment (single-executor model — no audit chain).
//
// Issue: #232 ADR-1
func TestDocumentAssignExecutor_Reassign(t *testing.T) {
	now := time.Date(2026, 5, 17, 11, 0, 0, 0, time.UTC)
	later := now.Add(2 * time.Hour)
	d := NewDocument("План", 1, 42)
	d.Status = DocumentStatusExecution

	if err := d.AssignExecutor(13, nil, 7, now); err != nil {
		t.Fatalf("first assign failed: %v", err)
	}
	if err := d.AssignExecutor(99, nil, 7, later); err != nil {
		t.Fatalf("reassign failed: %v", err)
	}
	if d.ExecutorAssignedTo == nil || *d.ExecutorAssignedTo != 99 {
		t.Errorf("expected reassign к 99, got %v", d.ExecutorAssignedTo)
	}
	if d.ExecutorAssignedAt == nil || !d.ExecutorAssignedAt.Equal(later) {
		t.Errorf("expected ExecutorAssignedAt=%v after reassign, got %v", later, d.ExecutorAssignedAt)
	}
}

// TestDocumentMarkExecuted_TableDriven pins the v0.151.0 MarkExecuted
// transition: execution → executed + audit; any other → ErrCannotMarkExecuted.
//
// Issue: #232
func TestDocumentMarkExecuted_TableDriven(t *testing.T) {
	now := time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC)
	actorID := int64(7)

	cases := []struct {
		name       string
		startState DocumentStatus
		wantStatus DocumentStatus
		wantErr    error
	}{
		{name: "execution → executed", startState: DocumentStatusExecution, wantStatus: DocumentStatusExecuted},
		{name: "draft → err", startState: DocumentStatusDraft, wantErr: ErrCannotMarkExecuted},
		{name: "approval → err", startState: DocumentStatusApproval, wantErr: ErrCannotMarkExecuted},
		{name: "approved → err", startState: DocumentStatusApproved, wantErr: ErrCannotMarkExecuted},
		{name: "registered → err", startState: DocumentStatusRegistered, wantErr: ErrCannotMarkExecuted},
		{name: "routing → err (must sign visa first)", startState: DocumentStatusRouting, wantErr: ErrCannotMarkExecuted},
		{name: "executed → err (already)", startState: DocumentStatusExecuted, wantErr: ErrCannotMarkExecuted},
		{name: "rejected → err", startState: DocumentStatusRejected, wantErr: ErrCannotMarkExecuted},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDocument("План", 1, 42)
			d.Status = tc.startState
			err := d.MarkExecuted(actorID, now)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				if d.Status != tc.startState {
					t.Errorf("status mutated on rejected transition: %q → %q", tc.startState, d.Status)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.Status != tc.wantStatus {
				t.Errorf("expected status %q, got %q", tc.wantStatus, d.Status)
			}
			if d.ExecutedBy == nil || *d.ExecutedBy != actorID {
				t.Errorf("expected ExecutedBy=%d, got %v", actorID, d.ExecutedBy)
			}
			if d.ExecutedAt == nil || !d.ExecutedAt.Equal(now) {
				t.Errorf("expected ExecutedAt=%v, got %v", now, d.ExecutedAt)
			}
		})
	}
}
