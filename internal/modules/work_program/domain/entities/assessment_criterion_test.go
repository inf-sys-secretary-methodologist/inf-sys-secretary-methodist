package entities_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

func TestNewAssessmentCriterion_HappyPath(t *testing.T) {
	a, err := entities.NewAssessmentCriterion(entities.NewAssessmentCriterionInput{
		Type:        domain.AssessmentTypeCurrent,
		Description: "Контрольная работа №1: реляционная алгебра",
		MaxScore:    20,
		ExampleQuestions: []string{
			"Дайте определение третьей нормальной формы",
			"Приведите пример декомпозиции отношения",
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if a == nil {
		t.Fatal("expected non-nil AssessmentCriterion")
	}
	if a.Type() != domain.AssessmentTypeCurrent {
		t.Errorf("Type: got %s, want %s", a.Type(), domain.AssessmentTypeCurrent)
	}
	if a.Description() != "Контрольная работа №1: реляционная алгебра" {
		t.Errorf("Description: %q", a.Description())
	}
	if a.MaxScore() != 20 {
		t.Errorf("MaxScore: got %d, want 20", a.MaxScore())
	}
	if len(a.ExampleQuestions()) != 2 {
		t.Errorf("ExampleQuestions len: got %d, want 2", len(a.ExampleQuestions()))
	}
}

func TestNewAssessmentCriterion_TrimsDescription(t *testing.T) {
	a, err := entities.NewAssessmentCriterion(entities.NewAssessmentCriterionInput{
		Type:        domain.AssessmentTypeFinal,
		Description: "  Итоговый экзамен  ",
		MaxScore:    100,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if a.Description() != "Итоговый экзамен" {
		t.Errorf("Description not trimmed: %q", a.Description())
	}
}

func TestNewAssessmentCriterion_TrimsExampleQuestions(t *testing.T) {
	a, err := entities.NewAssessmentCriterion(entities.NewAssessmentCriterionInput{
		Type:             domain.AssessmentTypeIntermediate,
		Description:      "Промежуточная аттестация",
		MaxScore:         50,
		ExampleQuestions: []string{"  Вопрос 1  ", " Вопрос 2 "},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(a.ExampleQuestions()) != 2 {
		t.Fatalf("ExampleQuestions len: %d", len(a.ExampleQuestions()))
	}
	if a.ExampleQuestions()[0] != "Вопрос 1" || a.ExampleQuestions()[1] != "Вопрос 2" {
		t.Errorf("ExampleQuestions not trimmed: %#v", a.ExampleQuestions())
	}
}

func TestNewAssessmentCriterion_NilExampleQuestionsOK(t *testing.T) {
	a, err := entities.NewAssessmentCriterion(entities.NewAssessmentCriterionInput{
		Type:        domain.AssessmentTypeCurrent,
		Description: "Опрос",
		MaxScore:    5,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(a.ExampleQuestions()) != 0 {
		t.Errorf("ExampleQuestions should be empty/nil, got %#v", a.ExampleQuestions())
	}
}

func TestNewAssessmentCriterion_InvariantViolations(t *testing.T) {
	base := entities.NewAssessmentCriterionInput{
		Type:        domain.AssessmentTypeCurrent,
		Description: "Тест",
		MaxScore:    20,
	}

	tests := []struct {
		name      string
		mutate    func(*entities.NewAssessmentCriterionInput)
		wantField string
	}{
		{
			name:      "invalid type rejected",
			mutate:    func(in *entities.NewAssessmentCriterionInput) { in.Type = domain.AssessmentType("midterm") },
			wantField: "type",
		},
		{
			name:      "empty type rejected",
			mutate:    func(in *entities.NewAssessmentCriterionInput) { in.Type = "" },
			wantField: "type",
		},
		{
			name:      "empty description rejected",
			mutate:    func(in *entities.NewAssessmentCriterionInput) { in.Description = "" },
			wantField: "description",
		},
		{
			name:      "whitespace description rejected",
			mutate:    func(in *entities.NewAssessmentCriterionInput) { in.Description = "  \t  " },
			wantField: "description",
		},
		{
			name:      "zero max_score rejected",
			mutate:    func(in *entities.NewAssessmentCriterionInput) { in.MaxScore = 0 },
			wantField: "max_score",
		},
		{
			name:      "negative max_score rejected",
			mutate:    func(in *entities.NewAssessmentCriterionInput) { in.MaxScore = -1 },
			wantField: "max_score",
		},
		{
			name:      "max_score above 100 rejected",
			mutate:    func(in *entities.NewAssessmentCriterionInput) { in.MaxScore = 101 },
			wantField: "max_score",
		},
		{
			name: "example_questions > 10 rejected",
			mutate: func(in *entities.NewAssessmentCriterionInput) {
				in.ExampleQuestions = make([]string, 11)
				for i := range in.ExampleQuestions {
					in.ExampleQuestions[i] = "Q"
				}
			},
			wantField: "example_questions",
		},
		{
			name: "empty example_question entry rejected",
			mutate: func(in *entities.NewAssessmentCriterionInput) {
				in.ExampleQuestions = []string{"Вопрос 1", "  "}
			},
			wantField: "example_questions",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := base
			tt.mutate(&in)
			a, err := entities.NewAssessmentCriterion(in)
			if err == nil {
				t.Fatalf("expected error, got nil; criterion=%+v", a)
			}
			if !errors.Is(err, domain.ErrInvalidWorkProgram) {
				t.Errorf("error should wrap ErrInvalidWorkProgram, got %v", err)
			}
			if !strings.Contains(err.Error(), tt.wantField) {
				t.Errorf("error %q should mention %q field", err.Error(), tt.wantField)
			}
			if a != nil {
				t.Errorf("expected nil criterion on invariant violation, got %+v", a)
			}
		})
	}
}
