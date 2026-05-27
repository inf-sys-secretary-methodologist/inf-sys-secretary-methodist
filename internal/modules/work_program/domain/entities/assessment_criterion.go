package entities

import (
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
)

// NewAssessmentCriterionInput collects constructor parameters for an
// AssessmentCriterion (ФОС item).
type NewAssessmentCriterionInput struct {
	Type             domain.AssessmentType
	Description      string
	MaxScore         int
	ExampleQuestions []string // optional, ≤ 10 entries, each non-empty after trim
}

// AssessmentCriterion — ФОС (фонд оценочных средств) item. Inner
// aggregate of WorkProgram per ADR-1.
type AssessmentCriterion struct {
	id               int64
	workProgramID    int64
	atype            domain.AssessmentType
	description      string
	maxScore         int
	exampleQuestions []string
}

// NewAssessmentCriterion stub for PR 1c RED phase — real invariants
// land in the GREEN commit.
func NewAssessmentCriterion(_ NewAssessmentCriterionInput) (*AssessmentCriterion, error) {
	return nil, domain.ErrInvalidWorkProgram
}

// ID returns the persistent identifier.
func (a *AssessmentCriterion) ID() int64 { return a.id }

// WorkProgramID returns the parent aggregate identifier.
func (a *AssessmentCriterion) WorkProgramID() int64 { return a.workProgramID }

// Type returns the typed kind (current/intermediate/final).
func (a *AssessmentCriterion) Type() domain.AssessmentType { return a.atype }

// Description returns the assessment description (trimmed).
func (a *AssessmentCriterion) Description() string { return a.description }

// MaxScore returns the maximum achievable score for this assessment (1..100).
func (a *AssessmentCriterion) MaxScore() int { return a.maxScore }

// ExampleQuestions returns the (optional) list of sample questions, ≤ 10 entries.
func (a *AssessmentCriterion) ExampleQuestions() []string { return a.exampleQuestions }
