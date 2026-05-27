package entities_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

func TestNewGoal_HappyPath(t *testing.T) {
	g, err := entities.NewGoal("Сформировать у студента системное мышление в области реляционных СУБД", 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if g.Text() != "Сформировать у студента системное мышление в области реляционных СУБД" {
		t.Errorf("Text: %q", g.Text())
	}
	if g.OrderIndex() != 1 {
		t.Errorf("OrderIndex: got %d, want 1", g.OrderIndex())
	}
}

func TestNewGoal_TrimsText(t *testing.T) {
	g, err := entities.NewGoal("  Освоить SQL  ", 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if g.Text() != "Освоить SQL" {
		t.Errorf("Text not trimmed: %q", g.Text())
	}
}

func TestNewGoal_InvariantViolations(t *testing.T) {
	cases := []struct {
		name       string
		text       string
		orderIndex int
		wantField  string
	}{
		{name: "empty text", text: "", orderIndex: 0, wantField: "text"},
		{name: "whitespace text", text: "  \t  ", orderIndex: 0, wantField: "text"},
		{name: "text > 2048 chars", text: strings.Repeat("я", 2049), orderIndex: 0, wantField: "text"},
		{name: "negative order_index", text: "Освоить SQL", orderIndex: -1, wantField: "order_index"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := entities.NewGoal(tc.text, tc.orderIndex)
			if err == nil {
				t.Fatalf("expected error, got nil; goal=%+v", g)
			}
			if !errors.Is(err, domain.ErrInvalidWorkProgram) {
				t.Errorf("error should wrap ErrInvalidWorkProgram, got %v", err)
			}
			if !strings.Contains(err.Error(), tc.wantField) {
				t.Errorf("error %q should mention %q", err.Error(), tc.wantField)
			}
			if g != nil {
				t.Errorf("expected nil goal on invariant violation, got %+v", g)
			}
		})
	}
}
