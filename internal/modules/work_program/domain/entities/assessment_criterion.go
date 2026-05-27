package entities

import (
	"fmt"
	"strings"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
)

const (
	minAssessmentScore       = 1
	maxAssessmentScore       = 100
	maxExampleQuestionsCount = 10
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

// NewAssessmentCriterion constructs a fresh AssessmentCriterion. Type
// is validated against canonical values, description trimmed and
// non-empty, max_score in [1, 100], example_questions limited to 10
// entries with each non-empty after trim (empty entries cannot carry
// meaning, so reject rather than silently skip).
func NewAssessmentCriterion(in NewAssessmentCriterionInput) (*AssessmentCriterion, error) {
	trimmedDesc := strings.TrimSpace(in.Description)

	if !in.Type.IsValid() {
		return nil, fmt.Errorf("%w: type %q must be one of current/intermediate/final",
			domain.ErrInvalidWorkProgram, in.Type)
	}
	if trimmedDesc == "" {
		return nil, fmt.Errorf("%w: description is required", domain.ErrInvalidWorkProgram)
	}
	if in.MaxScore < minAssessmentScore || in.MaxScore > maxAssessmentScore {
		return nil, fmt.Errorf("%w: max_score must be in [%d, %d]",
			domain.ErrInvalidWorkProgram, minAssessmentScore, maxAssessmentScore)
	}
	if len(in.ExampleQuestions) > maxExampleQuestionsCount {
		return nil, fmt.Errorf("%w: example_questions must contain at most %d entries",
			domain.ErrInvalidWorkProgram, maxExampleQuestionsCount)
	}

	var trimmedQuestions []string
	if len(in.ExampleQuestions) > 0 {
		trimmedQuestions = make([]string, 0, len(in.ExampleQuestions))
		for i, q := range in.ExampleQuestions {
			tq := strings.TrimSpace(q)
			if tq == "" {
				return nil, fmt.Errorf("%w: example_questions[%d] must not be empty",
					domain.ErrInvalidWorkProgram, i)
			}
			trimmedQuestions = append(trimmedQuestions, tq)
		}
	}

	return &AssessmentCriterion{
		atype:            in.Type,
		description:      trimmedDesc,
		maxScore:         in.MaxScore,
		exampleQuestions: trimmedQuestions,
	}, nil
}

// ReconstituteAssessmentCriterionInput collects fields for repository hydration.
type ReconstituteAssessmentCriterionInput struct {
	ID               int64
	WorkProgramID    int64
	Type             domain.AssessmentType
	Description      string
	MaxScore         int
	ExampleQuestions []string
}

// ReconstituteAssessmentCriterion builds an AssessmentCriterion from
// persisted state. Skips invariant checks — DB CHECK constraints
// already validated.
func ReconstituteAssessmentCriterion(in ReconstituteAssessmentCriterionInput) *AssessmentCriterion {
	return &AssessmentCriterion{
		id:               in.ID,
		workProgramID:    in.WorkProgramID,
		atype:            in.Type,
		description:      in.Description,
		maxScore:         in.MaxScore,
		exampleQuestions: in.ExampleQuestions,
	}
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
