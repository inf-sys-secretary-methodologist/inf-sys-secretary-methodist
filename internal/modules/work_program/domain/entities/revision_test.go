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
