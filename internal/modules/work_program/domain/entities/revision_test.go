package entities_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

func TestNewRevision_HappyPath(t *testing.T) {
	r, err := entities.NewRevision(entities.NewRevisionInput{
		WorkProgramID:  42,
		RevisionNumber: 1,
		ChangeType:     domain.RevisionChangeTypeLiterature,
		ChangeSummary:  "Обновлены ссылки на ГОСТ Р 7.0.5–2008",
		AuthorID:       7,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil Revision")
	}
	if r.WorkProgramID() != 42 {
		t.Errorf("WorkProgramID: got %d, want 42", r.WorkProgramID())
	}
	if r.RevisionNumber() != 1 {
		t.Errorf("RevisionNumber: got %d, want 1", r.RevisionNumber())
	}
	if r.ChangeType() != domain.RevisionChangeTypeLiterature {
		t.Errorf("ChangeType: got %s", r.ChangeType())
	}
	if r.AuthorID() != 7 {
		t.Errorf("AuthorID: got %d, want 7", r.AuthorID())
	}
	if r.Status() != domain.RevisionStatusDraft {
		t.Errorf("Status: got %s, want %s (fresh revisions start as draft)", r.Status(), domain.RevisionStatusDraft)
	}
	if r.ApproverID() != nil {
		t.Errorf("ApproverID: should be nil for draft, got %v", r.ApproverID())
	}
	if r.ApprovedAt() != nil {
		t.Errorf("ApprovedAt: should be nil for draft, got %v", r.ApprovedAt())
	}
	if r.RejectReason() != "" {
		t.Errorf("RejectReason: should be empty for draft, got %q", r.RejectReason())
	}
}

func TestNewRevision_TrimsSummary(t *testing.T) {
	r, err := entities.NewRevision(entities.NewRevisionInput{
		WorkProgramID:  42,
		RevisionNumber: 1,
		ChangeType:     domain.RevisionChangeTypeHours,
		ChangeSummary:  "  Перераспределены часы лекций  ",
		AuthorID:       7,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if r.ChangeSummary() != "Перераспределены часы лекций" {
		t.Errorf("ChangeSummary not trimmed: %q", r.ChangeSummary())
	}
}

func TestNewRevision_StoresDiffPayload(t *testing.T) {
	payload := []byte(`{"before":{"hours":36},"after":{"hours":40}}`)
	r, err := entities.NewRevision(entities.NewRevisionInput{
		WorkProgramID:  42,
		RevisionNumber: 1,
		ChangeType:     domain.RevisionChangeTypeHours,
		ChangeSummary:  "Изменение часов",
		AuthorID:       7,
		DiffPayload:    payload,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(r.DiffPayload()) != string(payload) {
		t.Errorf("DiffPayload: got %q, want %q", r.DiffPayload(), payload)
	}
}

// --- Sub-FSM transitions per ADR-10 ---

func TestRevision_Submit_FromDraft_TransitionsToPendingApproval(t *testing.T) {
	r := newDraftRevision(t)
	if err := r.Submit(); err != nil {
		t.Fatalf("Submit on draft: unexpected error %v", err)
	}
	if r.Status() != domain.RevisionStatusPendingApproval {
		t.Errorf("Status: got %s, want %s", r.Status(), domain.RevisionStatusPendingApproval)
	}
}

func TestRevision_Submit_FromForbiddenStatus_ReturnsTransitionError(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, r *entities.Revision)
	}{
		{
			name: "from pending_approval",
			setup: func(t *testing.T, r *entities.Revision) {
				t.Helper()
				if err := r.Submit(); err != nil {
					t.Fatalf("setup Submit: %v", err)
				}
			},
		},
		{
			name: "from approved",
			setup: func(t *testing.T, r *entities.Revision) {
				t.Helper()
				if err := r.Submit(); err != nil {
					t.Fatalf("setup Submit: %v", err)
				}
				if err := r.Approve(99); err != nil {
					t.Fatalf("setup Approve: %v", err)
				}
			},
		},
		{
			name: "from rejected",
			setup: func(t *testing.T, r *entities.Revision) {
				t.Helper()
				if err := r.Submit(); err != nil {
					t.Fatalf("setup Submit: %v", err)
				}
				if err := r.Reject("Не соответствует ФГОС"); err != nil {
					t.Fatalf("setup Reject: %v", err)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newDraftRevision(t)
			tc.setup(t, r)
			err := r.Submit()
			if !errors.Is(err, domain.ErrInvalidStatusTransition) {
				t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
			}
		})
	}
}

func TestRevision_Approve_FromPendingApproval_SetsApproverAndTimestamp(t *testing.T) {
	r := newDraftRevision(t)
	if err := r.Submit(); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if err := r.Approve(99); err != nil {
		t.Fatalf("Approve: unexpected error %v", err)
	}
	if r.Status() != domain.RevisionStatusApproved {
		t.Errorf("Status: got %s, want %s", r.Status(), domain.RevisionStatusApproved)
	}
	if r.ApproverID() == nil || *r.ApproverID() != 99 {
		t.Errorf("ApproverID: got %v, want *99", r.ApproverID())
	}
	if r.ApprovedAt() == nil {
		t.Error("ApprovedAt: should be set after Approve, got nil")
	}
}

func TestRevision_Approve_NonPositiveApproverID_RejectedAsInvariant(t *testing.T) {
	r := newDraftRevision(t)
	if err := r.Submit(); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	err := r.Approve(0)
	if !errors.Is(err, domain.ErrInvalidWorkProgram) {
		t.Errorf("expected ErrInvalidWorkProgram, got %v", err)
	}
	if r.Status() != domain.RevisionStatusPendingApproval {
		t.Errorf("Status should be unchanged on invariant failure: got %s", r.Status())
	}
}

func TestRevision_Approve_FromForbiddenStatus_ReturnsTransitionError(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, r *entities.Revision)
	}{
		{name: "from draft", setup: func(_ *testing.T, _ *entities.Revision) {}},
		{
			name: "from approved",
			setup: func(t *testing.T, r *entities.Revision) {
				t.Helper()
				if err := r.Submit(); err != nil {
					t.Fatalf("setup Submit: %v", err)
				}
				if err := r.Approve(99); err != nil {
					t.Fatalf("setup Approve: %v", err)
				}
			},
		},
		{
			name: "from rejected",
			setup: func(t *testing.T, r *entities.Revision) {
				t.Helper()
				if err := r.Submit(); err != nil {
					t.Fatalf("setup Submit: %v", err)
				}
				if err := r.Reject("причина"); err != nil {
					t.Fatalf("setup Reject: %v", err)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newDraftRevision(t)
			tc.setup(t, r)
			err := r.Approve(99)
			if !errors.Is(err, domain.ErrInvalidStatusTransition) {
				t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
			}
		})
	}
}

func TestRevision_Reject_FromPendingApproval_TransitionsToRejectedWithReason(t *testing.T) {
	r := newDraftRevision(t)
	if err := r.Submit(); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if err := r.Reject("Не соответствует ФГОС: формулировка ПК-3"); err != nil {
		t.Fatalf("Reject: unexpected error %v", err)
	}
	if r.Status() != domain.RevisionStatusRejected {
		t.Errorf("Status: got %s, want %s", r.Status(), domain.RevisionStatusRejected)
	}
	if r.RejectReason() != "Не соответствует ФГОС: формулировка ПК-3" {
		t.Errorf("RejectReason: %q", r.RejectReason())
	}
}

func TestRevision_Reject_TrimsReason(t *testing.T) {
	r := newDraftRevision(t)
	if err := r.Submit(); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if err := r.Reject("  Доработать раздел ФОС  "); err != nil {
		t.Fatalf("Reject: %v", err)
	}
	if r.RejectReason() != "Доработать раздел ФОС" {
		t.Errorf("RejectReason not trimmed: %q", r.RejectReason())
	}
}

func TestRevision_Reject_EmptyReason_ReturnsRejectReasonRequired(t *testing.T) {
	cases := []struct {
		name   string
		reason string
	}{
		{name: "empty string", reason: ""},
		{name: "whitespace only", reason: "  \t\n  "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newDraftRevision(t)
			if err := r.Submit(); err != nil {
				t.Fatalf("setup Submit: %v", err)
			}
			err := r.Reject(tc.reason)
			if !errors.Is(err, domain.ErrRejectReasonRequired) {
				t.Errorf("expected ErrRejectReasonRequired, got %v", err)
			}
			if r.Status() != domain.RevisionStatusPendingApproval {
				t.Errorf("Status should be unchanged: got %s", r.Status())
			}
		})
	}
}

func TestRevision_Reject_FromForbiddenStatus_ReturnsTransitionError(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, r *entities.Revision)
	}{
		{name: "from draft", setup: func(_ *testing.T, _ *entities.Revision) {}},
		{
			name: "from approved",
			setup: func(t *testing.T, r *entities.Revision) {
				t.Helper()
				if err := r.Submit(); err != nil {
					t.Fatalf("setup Submit: %v", err)
				}
				if err := r.Approve(99); err != nil {
					t.Fatalf("setup Approve: %v", err)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newDraftRevision(t)
			tc.setup(t, r)
			err := r.Reject("valid reason")
			if !errors.Is(err, domain.ErrInvalidStatusTransition) {
				t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
			}
		})
	}
}

// newDraftRevision builds a valid draft Revision for sub-FSM tests.
func newDraftRevision(t *testing.T) *entities.Revision {
	t.Helper()
	r, err := entities.NewRevision(entities.NewRevisionInput{
		WorkProgramID:  42,
		RevisionNumber: 1,
		ChangeType:     domain.RevisionChangeTypeOther,
		ChangeSummary:  "Прочие правки",
		AuthorID:       7,
	})
	if err != nil {
		t.Fatalf("newDraftRevision: NewRevision failed: %v", err)
	}
	return r
}

func TestNewRevision_InvariantViolations(t *testing.T) {
	base := entities.NewRevisionInput{
		WorkProgramID:  42,
		RevisionNumber: 1,
		ChangeType:     domain.RevisionChangeTypeOther,
		ChangeSummary:  "Прочие правки",
		AuthorID:       7,
	}

	tests := []struct {
		name      string
		mutate    func(*entities.NewRevisionInput)
		wantField string
	}{
		{
			name:      "non-positive work_program_id rejected",
			mutate:    func(in *entities.NewRevisionInput) { in.WorkProgramID = 0 },
			wantField: "work_program_id",
		},
		{
			name:      "zero revision_number rejected",
			mutate:    func(in *entities.NewRevisionInput) { in.RevisionNumber = 0 },
			wantField: "revision_number",
		},
		{
			name:      "negative revision_number rejected",
			mutate:    func(in *entities.NewRevisionInput) { in.RevisionNumber = -1 },
			wantField: "revision_number",
		},
		{
			name:      "invalid change_type rejected",
			mutate:    func(in *entities.NewRevisionInput) { in.ChangeType = domain.RevisionChangeType("structure") },
			wantField: "change_type",
		},
		{
			name:      "empty change_type rejected",
			mutate:    func(in *entities.NewRevisionInput) { in.ChangeType = "" },
			wantField: "change_type",
		},
		{
			name:      "empty change_summary rejected",
			mutate:    func(in *entities.NewRevisionInput) { in.ChangeSummary = "" },
			wantField: "change_summary",
		},
		{
			name:      "whitespace change_summary rejected",
			mutate:    func(in *entities.NewRevisionInput) { in.ChangeSummary = "  \t  " },
			wantField: "change_summary",
		},
		{
			name:      "change_summary > 4096 chars rejected",
			mutate:    func(in *entities.NewRevisionInput) { in.ChangeSummary = strings.Repeat("я", 4097) },
			wantField: "change_summary",
		},
		{
			name:      "non-positive author_id rejected",
			mutate:    func(in *entities.NewRevisionInput) { in.AuthorID = 0 },
			wantField: "author_id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := base
			tt.mutate(&in)
			r, err := entities.NewRevision(in)
			if err == nil {
				t.Fatalf("expected error, got nil; revision=%+v", r)
			}
			if !errors.Is(err, domain.ErrInvalidWorkProgram) {
				t.Errorf("error should wrap ErrInvalidWorkProgram, got %v", err)
			}
			if !strings.Contains(err.Error(), tt.wantField) {
				t.Errorf("error %q should mention %q field", err.Error(), tt.wantField)
			}
			if r != nil {
				t.Errorf("expected nil revision on invariant violation, got %+v", r)
			}
		})
	}
}
