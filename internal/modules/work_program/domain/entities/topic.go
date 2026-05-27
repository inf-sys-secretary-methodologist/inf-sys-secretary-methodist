package entities

import (
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
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

// NewTopic — stub for RED commit.
func NewTopic(_ NewTopicInput) (*Topic, error) {
	return nil, domain.ErrInvalidWorkProgram
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
