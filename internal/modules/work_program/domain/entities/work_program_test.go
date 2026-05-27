package entities_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// --- Constructor: NewWorkProgram ---

func TestNewWorkProgram_HappyPath(t *testing.T) {
	wp, err := entities.NewWorkProgram(entities.NewWorkProgramInput{
		DisciplineID:       42,
		SpecialtyCode:      "09.03.01",
		ApplicableFromYear: 2026,
		Title:              "Базы данных",
		Annotation:         "Курс по основам реляционных и NoSQL СУБД",
		AuthorID:           7,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if wp == nil {
		t.Fatal("expected non-nil WorkProgram")
	}
	if wp.Title() != "Базы данных" {
		t.Errorf("Title: got %q, want %q", wp.Title(), "Базы данных")
	}
	if wp.SpecialtyCode() != "09.03.01" {
		t.Errorf("SpecialtyCode: got %q, want %q", wp.SpecialtyCode(), "09.03.01")
	}
	if wp.ApplicableFromYear() != 2026 {
		t.Errorf("ApplicableFromYear: got %d, want 2026", wp.ApplicableFromYear())
	}
	if wp.DisciplineID() != 42 {
		t.Errorf("DisciplineID: got %d, want 42", wp.DisciplineID())
	}
	if wp.AuthorID() != 7 {
		t.Errorf("AuthorID: got %d, want 7", wp.AuthorID())
	}
	if wp.Status() != domain.StatusDraft {
		t.Errorf("Status: got %s, want %s (fresh aggregates start as draft)", wp.Status(), domain.StatusDraft)
	}
	if wp.Version() != 0 {
		t.Errorf("Version: got %d, want 0 (optimistic-lock starts at 0)", wp.Version())
	}
	if wp.ApprovedAt() != nil {
		t.Errorf("ApprovedAt: should be nil for draft, got %v", wp.ApprovedAt())
	}
}

func TestNewWorkProgram_InvariantViolations(t *testing.T) {
	base := entities.NewWorkProgramInput{
		DisciplineID:       42,
		SpecialtyCode:      "09.03.01",
		ApplicableFromYear: 2026,
		Title:              "Базы данных",
		Annotation:         "Описание",
		AuthorID:           7,
	}

	tests := []struct {
		name    string
		mutate  func(*entities.NewWorkProgramInput)
		wantMsg string
	}{
		{
			name:    "empty title rejected",
			mutate:  func(in *entities.NewWorkProgramInput) { in.Title = "" },
			wantMsg: "title",
		},
		{
			name:    "whitespace-only title rejected",
			mutate:  func(in *entities.NewWorkProgramInput) { in.Title = "   \t  " },
			wantMsg: "title",
		},
		{
			name:    "empty specialty_code rejected",
			mutate:  func(in *entities.NewWorkProgramInput) { in.SpecialtyCode = "" },
			wantMsg: "specialty_code",
		},
		{
			name:    "non-positive discipline_id rejected",
			mutate:  func(in *entities.NewWorkProgramInput) { in.DisciplineID = 0 },
			wantMsg: "discipline_id",
		},
		{
			name:    "non-positive author_id rejected",
			mutate:  func(in *entities.NewWorkProgramInput) { in.AuthorID = -1 },
			wantMsg: "author_id",
		},
		{
			name:    "year below 2000 rejected",
			mutate:  func(in *entities.NewWorkProgramInput) { in.ApplicableFromYear = 1999 },
			wantMsg: "applicable_from_year",
		},
		{
			name:    "year above 2100 rejected",
			mutate:  func(in *entities.NewWorkProgramInput) { in.ApplicableFromYear = 2101 },
			wantMsg: "applicable_from_year",
		},
		{
			name:    "annotation > 8192 chars rejected",
			mutate:  func(in *entities.NewWorkProgramInput) { in.Annotation = strings.Repeat("я", 8193) },
			wantMsg: "annotation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := base
			tt.mutate(&in)
			wp, err := entities.NewWorkProgram(in)
			if err == nil {
				t.Fatalf("expected error, got nil; wp=%+v", wp)
			}
			if !errors.Is(err, domain.ErrInvalidWorkProgram) {
				t.Errorf("error should wrap ErrInvalidWorkProgram, got %v", err)
			}
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("error %q should mention %q field", err.Error(), tt.wantMsg)
			}
			if wp != nil {
				t.Errorf("expected nil wp on invariant violation, got %+v", wp)
			}
		})
	}
}

// --- Submit transition ---

func TestSubmit_FromDraft_TransitionsToPendingApproval(t *testing.T) {
	wp := newDraft(t)
	if err := wp.Submit(); err != nil {
		t.Fatalf("Submit on draft: unexpected error %v", err)
	}
	if wp.Status() != domain.StatusPendingApproval {
		t.Errorf("Status: got %s, want %s", wp.Status(), domain.StatusPendingApproval)
	}
}

func TestSubmit_FromNeedsRevision_TransitionsToPendingApproval(t *testing.T) {
	wp := newDraft(t)
	if err := wp.Submit(); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if err := wp.Approve(99); err != nil {
		t.Fatalf("Approve: %v", err)
	}
	if err := wp.MarkNeedsRevision(); err != nil {
		t.Fatalf("MarkNeedsRevision: %v", err)
	}
	if err := wp.Submit(); err != nil {
		t.Fatalf("Submit from needs_revision: %v", err)
	}
	if wp.Status() != domain.StatusPendingApproval {
		t.Errorf("Status: got %s, want %s", wp.Status(), domain.StatusPendingApproval)
	}
}

func TestSubmit_FromForbiddenStatus_ReturnsTransitionError(t *testing.T) {
	cases := []struct {
		name      string
		setStatus func(t *testing.T, wp *entities.WorkProgram)
	}{
		{
			name: "from pending_approval",
			setStatus: func(t *testing.T, wp *entities.WorkProgram) {
				t.Helper()
				if err := wp.Submit(); err != nil {
					t.Fatalf("setup Submit: %v", err)
				}
			},
		},
		{
			name: "from approved",
			setStatus: func(t *testing.T, wp *entities.WorkProgram) {
				t.Helper()
				if err := wp.Submit(); err != nil {
					t.Fatalf("setup Submit: %v", err)
				}
				if err := wp.Approve(99); err != nil {
					t.Fatalf("setup Approve: %v", err)
				}
			},
		},
		{
			name: "from archived",
			setStatus: func(t *testing.T, wp *entities.WorkProgram) {
				t.Helper()
				if err := wp.Archive(); err != nil {
					t.Fatalf("setup Archive: %v", err)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wp := newDraft(t)
			tc.setStatus(t, wp)
			err := wp.Submit()
			if !errors.Is(err, domain.ErrInvalidStatusTransition) {
				t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
			}
		})
	}
}

// --- Reject transition ---

func TestReject_FromPendingApproval_TransitionsToDraftWithReason(t *testing.T) {
	wp := newDraft(t)
	if err := wp.Submit(); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if err := wp.Reject("Не соответствует ФГОС: отсутствуют ПК-3, ПК-7"); err != nil {
		t.Fatalf("Reject: unexpected error %v", err)
	}
	if wp.Status() != domain.StatusDraft {
		t.Errorf("Status: got %s, want %s", wp.Status(), domain.StatusDraft)
	}
	if wp.RejectReason() != "Не соответствует ФГОС: отсутствуют ПК-3, ПК-7" {
		t.Errorf("RejectReason: got %q", wp.RejectReason())
	}
}

func TestReject_EmptyReason_ReturnsRejectReasonRequired(t *testing.T) {
	cases := []struct {
		name   string
		reason string
	}{
		{name: "empty string", reason: ""},
		{name: "whitespace only", reason: "   \t\n  "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wp := newDraft(t)
			if err := wp.Submit(); err != nil {
				t.Fatalf("setup Submit: %v", err)
			}
			err := wp.Reject(tc.reason)
			if !errors.Is(err, domain.ErrRejectReasonRequired) {
				t.Errorf("expected ErrRejectReasonRequired, got %v", err)
			}
			if wp.Status() != domain.StatusPendingApproval {
				t.Errorf("Status should be unchanged: got %s", wp.Status())
			}
		})
	}
}

func TestReject_TrimsReason(t *testing.T) {
	wp := newDraft(t)
	if err := wp.Submit(); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if err := wp.Reject("  Доработать раздел ФОС  "); err != nil {
		t.Fatalf("Reject: %v", err)
	}
	if wp.RejectReason() != "Доработать раздел ФОС" {
		t.Errorf("RejectReason not trimmed: %q", wp.RejectReason())
	}
}

func TestReject_FromForbiddenStatus_ReturnsTransitionError(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, wp *entities.WorkProgram)
	}{
		{
			name:  "from draft",
			setup: func(_ *testing.T, _ *entities.WorkProgram) { /* stays in draft */ },
		},
		{
			name: "from approved",
			setup: func(t *testing.T, wp *entities.WorkProgram) {
				t.Helper()
				if err := wp.Submit(); err != nil {
					t.Fatalf("setup Submit: %v", err)
				}
				if err := wp.Approve(99); err != nil {
					t.Fatalf("setup Approve: %v", err)
				}
			},
		},
		{
			name: "from archived",
			setup: func(t *testing.T, wp *entities.WorkProgram) {
				t.Helper()
				if err := wp.Archive(); err != nil {
					t.Fatalf("setup Archive: %v", err)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wp := newDraft(t)
			tc.setup(t, wp)
			err := wp.Reject("valid reason")
			if !errors.Is(err, domain.ErrInvalidStatusTransition) {
				t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
			}
		})
	}
}

// --- DiscardDraft transition ---

func TestDiscardDraft_FromDraft_TransitionsToArchived(t *testing.T) {
	wp := newDraft(t)
	if err := wp.DiscardDraft(); err != nil {
		t.Fatalf("DiscardDraft: %v", err)
	}
	if wp.Status() != domain.StatusArchived {
		t.Errorf("Status: got %s, want %s", wp.Status(), domain.StatusArchived)
	}
}

func TestDiscardDraft_FromNonDraft_ReturnsTransitionError(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, wp *entities.WorkProgram)
	}{
		{
			name: "from pending_approval",
			setup: func(t *testing.T, wp *entities.WorkProgram) {
				t.Helper()
				if err := wp.Submit(); err != nil {
					t.Fatalf("setup Submit: %v", err)
				}
			},
		},
		{
			name: "from approved",
			setup: func(t *testing.T, wp *entities.WorkProgram) {
				t.Helper()
				if err := wp.Submit(); err != nil {
					t.Fatalf("setup Submit: %v", err)
				}
				if err := wp.Approve(99); err != nil {
					t.Fatalf("setup Approve: %v", err)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wp := newDraft(t)
			tc.setup(t, wp)
			err := wp.DiscardDraft()
			if !errors.Is(err, domain.ErrInvalidStatusTransition) {
				t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
			}
		})
	}
}

// newDraft is a test helper that builds a valid draft WorkProgram.
func newDraft(t *testing.T) *entities.WorkProgram {
	t.Helper()
	wp, err := entities.NewWorkProgram(entities.NewWorkProgramInput{
		DisciplineID:       42,
		SpecialtyCode:      "09.03.01",
		ApplicableFromYear: 2026,
		Title:              "Базы данных",
		Annotation:         "Курс по основам СУБД",
		AuthorID:           7,
	})
	if err != nil {
		t.Fatalf("newDraft: NewWorkProgram failed: %v", err)
	}
	return wp
}

func TestNewWorkProgram_TrimsWhitespace(t *testing.T) {
	wp, err := entities.NewWorkProgram(entities.NewWorkProgramInput{
		DisciplineID:       42,
		SpecialtyCode:      "  09.03.01  ",
		ApplicableFromYear: 2026,
		Title:              "  Базы данных  ",
		Annotation:         "  Описание  ",
		AuthorID:           7,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if wp.Title() != "Базы данных" {
		t.Errorf("Title not trimmed: %q", wp.Title())
	}
	if wp.SpecialtyCode() != "09.03.01" {
		t.Errorf("SpecialtyCode not trimmed: %q", wp.SpecialtyCode())
	}
	if wp.Annotation() != "Описание" {
		t.Errorf("Annotation not trimmed: %q", wp.Annotation())
	}
}
