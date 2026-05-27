package entities_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

func TestNewTopic_HappyPath(t *testing.T) {
	week := 3
	tp, err := entities.NewTopic(entities.NewTopicInput{
		Kind:             domain.TopicKindLecture,
		Title:            "Реляционная модель: операции реляционной алгебры",
		Hours:            4,
		WeekNumber:       &week,
		LearningOutcomes: "Студент способен выразить запрос на языке РА",
		OrderIndex:       1,
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if tp.Kind() != domain.TopicKindLecture {
		t.Errorf("Kind: %s", tp.Kind())
	}
	if tp.Hours() != 4 {
		t.Errorf("Hours: %d", tp.Hours())
	}
	if tp.WeekNumber() == nil || *tp.WeekNumber() != 3 {
		t.Errorf("WeekNumber: %v", tp.WeekNumber())
	}
	if tp.LearningOutcomes() != "Студент способен выразить запрос на языке РА" {
		t.Errorf("Outcomes: %q", tp.LearningOutcomes())
	}
}

func TestNewTopic_NilWeekAllowed(t *testing.T) {
	tp, err := entities.NewTopic(entities.NewTopicInput{
		Kind:  domain.TopicKindSelfStudy,
		Title: "Самостоятельная работа",
		Hours: 12,
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if tp.WeekNumber() != nil {
		t.Errorf("WeekNumber should be nil for self-study")
	}
}

func TestNewTopic_InvariantViolations(t *testing.T) {
	base := func() entities.NewTopicInput {
		return entities.NewTopicInput{
			Kind:  domain.TopicKindPractice,
			Title: "Семинар по нормализации",
			Hours: 2,
		}
	}
	cases := []struct {
		name      string
		mutate    func(*entities.NewTopicInput)
		wantField string
	}{
		{name: "invalid kind", mutate: func(in *entities.NewTopicInput) { in.Kind = domain.TopicKind("xx") }, wantField: "kind"},
		{name: "empty title", mutate: func(in *entities.NewTopicInput) { in.Title = "" }, wantField: "title"},
		{name: "whitespace title", mutate: func(in *entities.NewTopicInput) { in.Title = " \t " }, wantField: "title"},
		{name: "non-positive hours", mutate: func(in *entities.NewTopicInput) { in.Hours = 0 }, wantField: "hours"},
		{name: "negative hours", mutate: func(in *entities.NewTopicInput) { in.Hours = -1 }, wantField: "hours"},
		{name: "week < 1", mutate: func(in *entities.NewTopicInput) { w := 0; in.WeekNumber = &w }, wantField: "week_number"},
		{name: "week > 52", mutate: func(in *entities.NewTopicInput) { w := 53; in.WeekNumber = &w }, wantField: "week_number"},
		{name: "negative order_index", mutate: func(in *entities.NewTopicInput) { in.OrderIndex = -1 }, wantField: "order_index"},
		{name: "outcomes > 2048", mutate: func(in *entities.NewTopicInput) { in.LearningOutcomes = strings.Repeat("я", 2049) }, wantField: "learning_outcomes"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := base()
			tc.mutate(&in)
			tp, err := entities.NewTopic(in)
			if err == nil {
				t.Fatalf("expected error, got nil; topic=%+v", tp)
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
