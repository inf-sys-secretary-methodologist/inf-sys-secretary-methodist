package entities

import (
	"errors"
	"testing"
	"time"
)

// TestDocumentArchive_TableDriven pins the v0.152.0 Archive terminal
// transition: executed → archived + audit; any other → ErrCannotArchive.
// Mirror к MarkExecuted pattern.
//
// Issue: #233
func TestDocumentArchive_TableDriven(t *testing.T) {
	now := time.Date(2026, 5, 17, 14, 0, 0, 0, time.UTC)
	actorID := int64(7)

	cases := []struct {
		name       string
		startState DocumentStatus
		wantStatus DocumentStatus
		wantErr    error
	}{
		{name: "executed → archived", startState: DocumentStatusExecuted, wantStatus: DocumentStatusArchived},
		{name: "draft → err", startState: DocumentStatusDraft, wantErr: ErrCannotArchive},
		{name: "approval → err", startState: DocumentStatusApproval, wantErr: ErrCannotArchive},
		{name: "approved → err", startState: DocumentStatusApproved, wantErr: ErrCannotArchive},
		{name: "registered → err", startState: DocumentStatusRegistered, wantErr: ErrCannotArchive},
		{name: "routing → err", startState: DocumentStatusRouting, wantErr: ErrCannotArchive},
		{name: "execution → err (must mark executed first)", startState: DocumentStatusExecution, wantErr: ErrCannotArchive},
		{name: "rejected → err", startState: DocumentStatusRejected, wantErr: ErrCannotArchive},
		{name: "archived → err (already)", startState: DocumentStatusArchived, wantErr: ErrCannotArchive},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDocument("План", 1, 42)
			d.Status = tc.startState
			err := d.Archive(actorID, now)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				if d.Status != tc.startState {
					t.Errorf("status mutated on rejected transition: %q → %q", tc.startState, d.Status)
				}
				if d.ArchivedBy != nil {
					t.Errorf("ArchivedBy mutated on rejected transition: %v", d.ArchivedBy)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.Status != tc.wantStatus {
				t.Errorf("expected status %q, got %q", tc.wantStatus, d.Status)
			}
			if d.ArchivedBy == nil || *d.ArchivedBy != actorID {
				t.Errorf("expected ArchivedBy=%d, got %v", actorID, d.ArchivedBy)
			}
			if d.ArchivedAt == nil || !d.ArchivedAt.Equal(now) {
				t.Errorf("expected ArchivedAt=%v, got %v", now, d.ArchivedAt)
			}
		})
	}
}

// TestDocumentResubmit_TableDriven pins the v0.152.0 Resubmit cycle:
// rejected → draft + clears RejectedBy/At/Reason audit fields; any
// other → ErrCannotResubmit (no mutation). Original SubmittedBy/At
// preserved (forensic history).
//
// Issue: #233
func TestDocumentResubmit_TableDriven(t *testing.T) {
	now := time.Date(2026, 5, 17, 15, 0, 0, 0, time.UTC)
	earlier := now.Add(-24 * time.Hour)
	actorID := int64(7)
	originalAuthor := int64(42)
	rejectorID := int64(99)
	reason := "Шаблон 2023 устарел — обновите за неделю"

	cases := []struct {
		name       string
		startState DocumentStatus
		wantStatus DocumentStatus
		wantErr    error
	}{
		{name: "rejected → draft + cleared audit", startState: DocumentStatusRejected, wantStatus: DocumentStatusDraft},
		{name: "draft → err (not rejected)", startState: DocumentStatusDraft, wantErr: ErrCannotResubmit},
		{name: "approval → err", startState: DocumentStatusApproval, wantErr: ErrCannotResubmit},
		{name: "approved → err", startState: DocumentStatusApproved, wantErr: ErrCannotResubmit},
		{name: "registered → err", startState: DocumentStatusRegistered, wantErr: ErrCannotResubmit},
		{name: "routing → err", startState: DocumentStatusRouting, wantErr: ErrCannotResubmit},
		{name: "execution → err", startState: DocumentStatusExecution, wantErr: ErrCannotResubmit},
		{name: "executed → err", startState: DocumentStatusExecuted, wantErr: ErrCannotResubmit},
		{name: "archived → err", startState: DocumentStatusArchived, wantErr: ErrCannotResubmit},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDocument("План", 1, originalAuthor)
			d.Status = tc.startState
			// Pre-populate rejection audit + original submission audit
			// to verify Resubmit clears the right fields.
			d.SubmittedBy = &originalAuthor
			d.SubmittedAt = &earlier
			d.RejectedBy = &rejectorID
			d.RejectedAt = &earlier
			d.RejectedReason = &reason

			err := d.Resubmit(actorID, now)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				if d.Status != tc.startState {
					t.Errorf("status mutated on rejected transition: %q → %q", tc.startState, d.Status)
				}
				// rejection audit must remain intact on rejected branch.
				if d.RejectedBy == nil || d.RejectedReason == nil {
					t.Errorf("rejection audit cleared on rejected transition (should remain): RejectedBy=%v, RejectedReason=%v", d.RejectedBy, d.RejectedReason)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.Status != tc.wantStatus {
				t.Errorf("expected status %q, got %q", tc.wantStatus, d.Status)
			}
			if d.RejectedBy != nil {
				t.Errorf("expected RejectedBy cleared, got %v", d.RejectedBy)
			}
			if d.RejectedAt != nil {
				t.Errorf("expected RejectedAt cleared, got %v", d.RejectedAt)
			}
			if d.RejectedReason != nil {
				t.Errorf("expected RejectedReason cleared, got %v", d.RejectedReason)
			}
			// SubmittedBy/At preserved — historical record.
			if d.SubmittedBy == nil || *d.SubmittedBy != originalAuthor {
				t.Errorf("expected SubmittedBy preserved=%d, got %v", originalAuthor, d.SubmittedBy)
			}
			if d.UpdatedAt != now {
				t.Errorf("expected UpdatedAt=%v, got %v", now, d.UpdatedAt)
			}
		})
	}
}
