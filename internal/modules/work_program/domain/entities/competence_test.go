package entities_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

func TestNewCompetence_HappyPath(t *testing.T) {
	c, err := entities.NewCompetence("ПК-3", domain.CompetenceTypePK,
		"Способен разрабатывать модели данных")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Code() != "ПК-3" {
		t.Errorf("Code: %q", c.Code())
	}
	if c.Type() != domain.CompetenceTypePK {
		t.Errorf("Type: %s", c.Type())
	}
	if c.Description() != "Способен разрабатывать модели данных" {
		t.Errorf("Description: %q", c.Description())
	}
}

func TestNewCompetence_TrimsCodeAndDescription(t *testing.T) {
	c, err := entities.NewCompetence("  УК-7  ", domain.CompetenceTypeUK, "  Описание  ")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if c.Code() != "УК-7" {
		t.Errorf("Code not trimmed: %q", c.Code())
	}
	if c.Description() != "Описание" {
		t.Errorf("Description not trimmed: %q", c.Description())
	}
}

func TestNewCompetence_InvariantViolations(t *testing.T) {
	cases := []struct {
		name      string
		code      string
		ctype     domain.CompetenceType
		desc      string
		wantField string
	}{
		{name: "empty code", code: "", ctype: domain.CompetenceTypePK, desc: "x", wantField: "code"},
		{name: "whitespace code", code: "   ", ctype: domain.CompetenceTypeOK, desc: "x", wantField: "code"},
		{name: "invalid type", code: "ПК-1", ctype: domain.CompetenceType("xx"), desc: "x", wantField: "type"},
		{name: "empty description", code: "ОК-1", ctype: domain.CompetenceTypeOK, desc: "", wantField: "description"},
		{name: "description > 2048", code: "ПК-1", ctype: domain.CompetenceTypePK, desc: strings.Repeat("я", 2049), wantField: "description"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := entities.NewCompetence(tc.code, tc.ctype, tc.desc)
			if err == nil {
				t.Fatalf("expected error, got nil; c=%+v", c)
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
