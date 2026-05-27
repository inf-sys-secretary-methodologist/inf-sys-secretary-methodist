package entities_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

func TestNewReference_HappyPath(t *testing.T) {
	year := 2024
	r, err := entities.NewReference(entities.NewReferenceInput{
		Kind:       domain.ReferenceKindMain,
		Citation:   "Дейт К. Дж. Введение в системы баз данных. — М.: Вильямс, 2024.",
		Year:       &year,
		ISBN:       "978-5-8459-0788-2",
		URL:        "",
		OrderIndex: 0,
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if r.Kind() != domain.ReferenceKindMain {
		t.Errorf("Kind: %s", r.Kind())
	}
	if r.Year() == nil || *r.Year() != 2024 {
		t.Errorf("Year: %v", r.Year())
	}
	if r.ISBN() != "978-5-8459-0788-2" {
		t.Errorf("ISBN: %q", r.ISBN())
	}
}

func TestNewReference_ElectronicWithURL(t *testing.T) {
	r, err := entities.NewReference(entities.NewReferenceInput{
		Kind:     domain.ReferenceKindElectronic,
		Citation: "PostgreSQL Documentation",
		URL:      "https://www.postgresql.org/docs/17/",
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if r.URL() != "https://www.postgresql.org/docs/17/" {
		t.Errorf("URL: %q", r.URL())
	}
	if r.Year() != nil {
		t.Errorf("Year should be nil")
	}
}

func TestNewReference_InvariantViolations(t *testing.T) {
	base := func() entities.NewReferenceInput {
		return entities.NewReferenceInput{
			Kind:     domain.ReferenceKindAdditional,
			Citation: "ГОСТ Р 7.0.5-2008",
		}
	}
	cases := []struct {
		name      string
		mutate    func(*entities.NewReferenceInput)
		wantField string
	}{
		{name: "invalid kind", mutate: func(in *entities.NewReferenceInput) { in.Kind = domain.ReferenceKind("xx") }, wantField: "kind"},
		{name: "empty citation", mutate: func(in *entities.NewReferenceInput) { in.Citation = "" }, wantField: "citation"},
		{name: "whitespace citation", mutate: func(in *entities.NewReferenceInput) { in.Citation = " \t " }, wantField: "citation"},
		{name: "year < 1900", mutate: func(in *entities.NewReferenceInput) { y := 1899; in.Year = &y }, wantField: "year"},
		{name: "year > 2100", mutate: func(in *entities.NewReferenceInput) { y := 2101; in.Year = &y }, wantField: "year"},
		{name: "URL > 500 chars", mutate: func(in *entities.NewReferenceInput) { in.URL = "https://" + strings.Repeat("x", 500) }, wantField: "url"},
		{name: "negative order_index", mutate: func(in *entities.NewReferenceInput) { in.OrderIndex = -1 }, wantField: "order_index"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := base()
			tc.mutate(&in)
			r, err := entities.NewReference(in)
			if err == nil {
				t.Fatalf("expected error, got nil; ref=%+v", r)
			}
			if !errors.Is(err, domain.ErrInvalidWorkProgram) {
				t.Errorf("error should wrap ErrInvalidWorkProgram, got %v", err)
			}
			if !strings.Contains(err.Error(), tc.wantField) {
				t.Errorf("error %q should mention %q", err.Error(), tc.wantField)
			}
		})
	}
}
