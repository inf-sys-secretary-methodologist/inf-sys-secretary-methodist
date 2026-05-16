package entities

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// TestDocumentSubmit_TableDriven pins the v0.148.0 Submit transition:
//   - draft → approval (sets submitted_by + submitted_at)
//   - any non-draft → ErrCannotSubmit
//
// Issue: #227
func TestDocumentSubmit_TableDriven(t *testing.T) {
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	authorID := int64(42)

	cases := []struct {
		name       string
		startState DocumentStatus
		wantStatus DocumentStatus
		wantErr    error
	}{
		{name: "draft → approval", startState: DocumentStatusDraft, wantStatus: DocumentStatusApproval},
		{name: "registered → err", startState: DocumentStatusRegistered, wantErr: ErrCannotSubmit},
		{name: "approval → err (already submitted)", startState: DocumentStatusApproval, wantErr: ErrCannotSubmit},
		{name: "approved → err", startState: DocumentStatusApproved, wantErr: ErrCannotSubmit},
		{name: "rejected → err", startState: DocumentStatusRejected, wantErr: ErrCannotSubmit},
		{name: "archived → err", startState: DocumentStatusArchived, wantErr: ErrCannotSubmit},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDocument("План занятий 2026", 1, authorID)
			d.Status = tc.startState
			err := d.Submit(authorID, now)

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
			if d.SubmittedBy == nil || *d.SubmittedBy != authorID {
				t.Errorf("expected SubmittedBy=%d, got %v", authorID, d.SubmittedBy)
			}
			if d.SubmittedAt == nil || !d.SubmittedAt.Equal(now) {
				t.Errorf("expected SubmittedAt=%v, got %v", now, d.SubmittedAt)
			}
		})
	}
}

// TestDocumentApprove_TableDriven pins the Approve transition:
//   - approval → approved (sets approved_by + approved_at)
//   - any non-approval → ErrCannotApprove
func TestDocumentApprove_TableDriven(t *testing.T) {
	now := time.Date(2026, 5, 16, 13, 0, 0, 0, time.UTC)
	adminID := int64(7)

	cases := []struct {
		name       string
		startState DocumentStatus
		wantStatus DocumentStatus
		wantErr    error
	}{
		{name: "approval → approved", startState: DocumentStatusApproval, wantStatus: DocumentStatusApproved},
		{name: "draft → err (skip submit)", startState: DocumentStatusDraft, wantErr: ErrCannotApprove},
		{name: "approved → err (already)", startState: DocumentStatusApproved, wantErr: ErrCannotApprove},
		{name: "rejected → err", startState: DocumentStatusRejected, wantErr: ErrCannotApprove},
		{name: "archived → err", startState: DocumentStatusArchived, wantErr: ErrCannotApprove},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDocument("План занятий", 1, 42)
			d.Status = tc.startState
			err := d.Approve(adminID, now)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.Status != tc.wantStatus {
				t.Errorf("expected status %q, got %q", tc.wantStatus, d.Status)
			}
			if d.ApprovedBy == nil || *d.ApprovedBy != adminID {
				t.Errorf("expected ApprovedBy=%d, got %v", adminID, d.ApprovedBy)
			}
			if d.ApprovedAt == nil || !d.ApprovedAt.Equal(now) {
				t.Errorf("expected ApprovedAt=%v, got %v", now, d.ApprovedAt)
			}
		})
	}
}

// TestDocumentReject_TableDriven pins the Reject transition:
//   - approval → rejected (sets rejected_by + rejected_reason + rejected_at)
//   - any non-approval → ErrCannotReject
//   - reason VO required и validated separately в TestRejectionReason
func TestDocumentReject_TableDriven(t *testing.T) {
	now := time.Date(2026, 5, 16, 14, 0, 0, 0, time.UTC)
	adminID := int64(7)
	reason := MustRejectionReason("Шаблон 2023 устарел — обновите за неделю")

	cases := []struct {
		name       string
		startState DocumentStatus
		wantStatus DocumentStatus
		wantErr    error
	}{
		{name: "approval → rejected", startState: DocumentStatusApproval, wantStatus: DocumentStatusRejected},
		{name: "draft → err", startState: DocumentStatusDraft, wantErr: ErrCannotReject},
		{name: "approved → err", startState: DocumentStatusApproved, wantErr: ErrCannotReject},
		{name: "rejected → err (already)", startState: DocumentStatusRejected, wantErr: ErrCannotReject},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDocument("План занятий", 1, 42)
			d.Status = tc.startState
			err := d.Reject(adminID, reason, now)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.Status != tc.wantStatus {
				t.Errorf("expected status %q, got %q", tc.wantStatus, d.Status)
			}
			if d.RejectedBy == nil || *d.RejectedBy != adminID {
				t.Errorf("expected RejectedBy=%d, got %v", adminID, d.RejectedBy)
			}
			if d.RejectedReason == nil || *d.RejectedReason != reason.String() {
				t.Errorf("expected RejectedReason=%q, got %v", reason.String(), d.RejectedReason)
			}
			if d.RejectedAt == nil || !d.RejectedAt.Equal(now) {
				t.Errorf("expected RejectedAt=%v, got %v", now, d.RejectedAt)
			}
		})
	}
}

// TestRejectionReason_Validation pins the RejectionReason VO invariant:
// length between 10 и 500 characters inclusive. Empty / too short /
// too long → ErrRejectionReasonInvalid; valid → NewRejectionReason
// returns the value.
func TestRejectionReason_Validation(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "valid — typical reject reason", raw: "Шаблон 2023 устарел — обновите за неделю", wantErr: false},
		{name: "valid — exactly 10 chars", raw: "ровно10чрс", wantErr: false},
		{name: "invalid — empty", raw: "", wantErr: true},
		{name: "invalid — 9 chars", raw: "девять зн", wantErr: true},
		{name: "invalid — only whitespace", raw: "          ", wantErr: true},
		{name: "invalid — 501 chars", raw: strings.Repeat("а", 501), wantErr: true},
		{name: "valid — exactly 500 chars", raw: strings.Repeat("а", 500), wantErr: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewRejectionReason(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got reason=%q", tc.raw, r.String())
				}
				if !errors.Is(err, ErrRejectionReasonInvalid) {
					t.Errorf("expected ErrRejectionReasonInvalid, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.raw, err)
			}
			if r.String() != strings.TrimSpace(tc.raw) {
				t.Errorf("expected reason %q (trimmed), got %q", strings.TrimSpace(tc.raw), r.String())
			}
		})
	}
}

// TestDocumentReject_RejectionReasonRequired confirms Reject refuses
// the zero-value RejectionReason. Defense-in-depth — usecase layer
// constructs reason via NewRejectionReason which validates, но domain
// must also reject zero-value passed по ошибке.
func TestDocumentReject_RejectionReasonRequired(t *testing.T) {
	d := NewDocument("План", 1, 42)
	d.Status = DocumentStatusApproval
	err := d.Reject(7, RejectionReason{}, time.Now())
	if !errors.Is(err, ErrRejectionReasonInvalid) {
		t.Fatalf("expected ErrRejectionReasonInvalid for zero-value reason, got %v", err)
	}
}
