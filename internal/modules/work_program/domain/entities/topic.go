package entities

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
)

const (
	maxTopicTitleLen       = 500
	maxLearningOutcomesLen = 2048
	minWeekNumber          = 1
	maxWeekNumber          = 52
)

// NewTopicInput collects constructor parameters for a Topic.
type NewTopicInput struct {
	Kind             domain.TopicKind
	Title            string
	Hours            int
	WeekNumber       *int // optional: 1..52
	LearningOutcomes string
	OrderIndex       int
}

// Topic — тема лекции/практики/лабораторной/самостоятельной работы.
// Inner aggregate of WorkProgram per ADR-1. Hours invariant (sum per
// kind matches WorkProgram.HoursTotal[kind]) is enforced at the
// aggregate level — ships with PR 1c (AddTopic method on WorkProgram).
type Topic struct {
	id               int64
	workProgramID    int64
	kind             domain.TopicKind
	title            string
	hours            int
	weekNumber       *int
	learningOutcomes string
	orderIndex       int
}

// NewTopic constructs a fresh Topic. Kind validated against canonical
// values, title trimmed and non-empty, hours strictly positive, week
// (if set) in [1,52], outcomes ≤ 2048 runes, order_index ≥ 0.
func NewTopic(in NewTopicInput) (*Topic, error) {
	trimmedTitle := strings.TrimSpace(in.Title)
	trimmedOutcomes := strings.TrimSpace(in.LearningOutcomes)

	if !in.Kind.IsValid() {
		return nil, fmt.Errorf("%w: kind %q must be one of lecture/practice/lab/self_study",
			domain.ErrInvalidWorkProgram, in.Kind)
	}
	if trimmedTitle == "" {
		return nil, fmt.Errorf("%w: title is required", domain.ErrInvalidWorkProgram)
	}
	if utf8.RuneCountInString(trimmedTitle) > maxTopicTitleLen {
		return nil, fmt.Errorf("%w: title must be <= %d runes", domain.ErrInvalidWorkProgram, maxTopicTitleLen)
	}
	if in.Hours <= 0 {
		return nil, fmt.Errorf("%w: hours must be positive", domain.ErrInvalidWorkProgram)
	}
	if in.WeekNumber != nil {
		if *in.WeekNumber < minWeekNumber || *in.WeekNumber > maxWeekNumber {
			return nil, fmt.Errorf("%w: week_number must be in [%d, %d]",
				domain.ErrInvalidWorkProgram, minWeekNumber, maxWeekNumber)
		}
	}
	if utf8.RuneCountInString(trimmedOutcomes) > maxLearningOutcomesLen {
		return nil, fmt.Errorf("%w: learning_outcomes must be <= %d runes",
			domain.ErrInvalidWorkProgram, maxLearningOutcomesLen)
	}
	if in.OrderIndex < 0 {
		return nil, fmt.Errorf("%w: order_index must be non-negative", domain.ErrInvalidWorkProgram)
	}
	return &Topic{
		kind:             in.Kind,
		title:            trimmedTitle,
		hours:            in.Hours,
		weekNumber:       in.WeekNumber,
		learningOutcomes: trimmedOutcomes,
		orderIndex:       in.OrderIndex,
	}, nil
}

// ID returns the persistent identifier.
func (t *Topic) ID() int64 { return t.id }

// WorkProgramID returns the parent aggregate identifier.
func (t *Topic) WorkProgramID() int64 { return t.workProgramID }

// Kind returns the typed kind (lecture/practice/lab/self_study).
func (t *Topic) Kind() domain.TopicKind { return t.kind }

// Title returns the topic title (trimmed).
func (t *Topic) Title() string { return t.title }

// Hours returns the academic hours allocated (positive).
func (t *Topic) Hours() int { return t.hours }

// WeekNumber returns the semester week (1..52), nil if not scheduled.
func (t *Topic) WeekNumber() *int { return t.weekNumber }

// LearningOutcomes returns the formative outcomes text.
func (t *Topic) LearningOutcomes() string { return t.learningOutcomes }

// OrderIndex returns the display ordering hint (≥ 0).
func (t *Topic) OrderIndex() int { return t.orderIndex }
