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
