package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// Manual editing of РПД inner collections (slice 12). A methodist or the РПД
// author fills/edits goals / competences / topics / assessments / references
// by hand (not only via LLM generation or import). Author-scoped per ADR-10
// (author OR system_admin) — mirrors CreateRevisionUseCase; the domain layer
// enforces the status gate (draft / needs_revision) and per-item invariants.

// workProgramContentRepo is the narrow load-mutate-persist port.
type workProgramContentRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.WorkProgram, error)
	Update(ctx context.Context, wp *entities.WorkProgram) error
}

// TopicContentInput is the application DTO for adding/updating a Topic. Kind
// is a string mapped to the domain enum inside the use case (a bad value
// fails the domain constructor).
type TopicContentInput struct {
	Kind             string
	Title            string
	Hours            int
	WeekNumber       *int
	LearningOutcomes string
	OrderIndex       int
}

func (in TopicContentInput) toDomain() entities.NewTopicInput {
	return entities.NewTopicInput{
		Kind:             domain.TopicKind(in.Kind),
		Title:            in.Title,
		Hours:            in.Hours,
		WeekNumber:       in.WeekNumber,
		LearningOutcomes: in.LearningOutcomes,
		OrderIndex:       in.OrderIndex,
	}
}

// AssessmentContentInput is the application DTO for adding/updating an
// AssessmentCriterion (ФОС item).
type AssessmentContentInput struct {
	Type             string
	Description      string
	MaxScore         int
	ExampleQuestions []string
}

func (in AssessmentContentInput) toDomain() entities.NewAssessmentCriterionInput {
	return entities.NewAssessmentCriterionInput{
		Type:             domain.AssessmentType(in.Type),
		Description:      in.Description,
		MaxScore:         in.MaxScore,
		ExampleQuestions: in.ExampleQuestions,
	}
}

// ReferenceContentInput is the application DTO for adding/updating a Reference.
type ReferenceContentInput struct {
	Kind       string
	Citation   string
	Year       *int
	ISBN       string
	URL        string
	OrderIndex int
}

func (in ReferenceContentInput) toDomain() entities.NewReferenceInput {
	return entities.NewReferenceInput{
		Kind:       domain.ReferenceKind(in.Kind),
		Citation:   in.Citation,
		Year:       in.Year,
		ISBN:       in.ISBN,
		URL:        in.URL,
		OrderIndex: in.OrderIndex,
	}
}

// WorkProgramContentUseCase orchestrates manual collection edits.
type WorkProgramContentUseCase struct {
	repo  workProgramContentRepo
	audit AuditSink
}

// NewWorkProgramContentUseCase wires the use case. Repo is required.
func NewWorkProgramContentUseCase(repo workProgramContentRepo, audit AuditSink) *WorkProgramContentUseCase {
	if repo == nil {
		panic("work_program: NewWorkProgramContentUseCase requires non-nil repo")
	}
	return &WorkProgramContentUseCase{repo: repo, audit: audit}
}

// mutate is the shared load → authorize → apply → persist skeleton for every
// collection edit. apply runs the domain mutation (which enforces the status
// gate + invariants). Authorization is author-scoped (author OR system_admin).
// Denials emit a forensic audit event; transport errors propagate without one.
func (uc *WorkProgramContentUseCase) mutate(
	ctx context.Context, actorID int64, actorRole string, wpID int64, action string,
	apply func(*entities.WorkProgram) error,
) (*entities.WorkProgram, error) {
	wp, err := uc.repo.GetByID(ctx, wpID)
	if err != nil {
		if errors.Is(err, repositories.ErrWorkProgramNotFound) {
			emitAudit(uc.audit, ctx, "work_program.content_"+action+"_denied",
				denialFields(actorID, wpID, "not_found", ""))
		}
		return nil, err
	}
	if !isAuthorOrSystemAdmin(actorID, actorRole, wp.AuthorID()) {
		emitAudit(uc.audit, ctx, "work_program.content_"+action+"_denied",
			denialFields(actorID, wpID, "forbidden", wp.SpecialtyCode()))
		return nil, fmt.Errorf("%w: actor %d is not the author (%d) and not system_admin",
			domain.ErrWorkProgramScopeForbidden, actorID, wp.AuthorID())
	}
	if err := apply(wp); err != nil {
		emitAudit(uc.audit, ctx, "work_program.content_"+action+"_denied",
			denialFields(actorID, wpID, "invalid", wp.SpecialtyCode()))
		return nil, err
	}
	if err := uc.repo.Update(ctx, wp); err != nil {
		return nil, err
	}
	emitAudit(uc.audit, ctx, "work_program.content_"+action,
		successFields(actorID, wpID, wp.SpecialtyCode(), string(wp.Status())))
	return wp, nil
}

// --- Goal ---

// AddGoal appends a manually-authored Goal.
func (uc *WorkProgramContentUseCase) AddGoal(ctx context.Context, actorID int64, actorRole string, wpID int64, text string, orderIndex int) (*entities.WorkProgram, error) {
	return uc.mutate(ctx, actorID, actorRole, wpID, "goal_add", func(wp *entities.WorkProgram) error {
		g, err := entities.NewGoal(text, orderIndex)
		if err != nil {
			return err
		}
		return wp.AddGoal(g)
	})
}

// UpdateGoal edits a Goal by id.
func (uc *WorkProgramContentUseCase) UpdateGoal(ctx context.Context, actorID int64, actorRole string, wpID, goalID int64, text string, orderIndex int) (*entities.WorkProgram, error) {
	return uc.mutate(ctx, actorID, actorRole, wpID, "goal_update", func(wp *entities.WorkProgram) error {
		return wp.UpdateGoal(goalID, text, orderIndex)
	})
}

// RemoveGoal deletes a Goal by id.
func (uc *WorkProgramContentUseCase) RemoveGoal(ctx context.Context, actorID int64, actorRole string, wpID, goalID int64) (*entities.WorkProgram, error) {
	return uc.mutate(ctx, actorID, actorRole, wpID, "goal_remove", func(wp *entities.WorkProgram) error {
		return wp.RemoveGoal(goalID)
	})
}

// --- Competence ---

// AddCompetence appends a Competence (ctype string → domain enum).
func (uc *WorkProgramContentUseCase) AddCompetence(ctx context.Context, actorID int64, actorRole string, wpID int64, code, ctype, description string) (*entities.WorkProgram, error) {
	return uc.mutate(ctx, actorID, actorRole, wpID, "competence_add", func(wp *entities.WorkProgram) error {
		c, err := entities.NewCompetence(code, domain.CompetenceType(ctype), description)
		if err != nil {
			return err
		}
		return wp.AddCompetence(c)
	})
}

// UpdateCompetence edits a Competence by id.
func (uc *WorkProgramContentUseCase) UpdateCompetence(ctx context.Context, actorID int64, actorRole string, wpID, compID int64, code, ctype, description string) (*entities.WorkProgram, error) {
	return uc.mutate(ctx, actorID, actorRole, wpID, "competence_update", func(wp *entities.WorkProgram) error {
		return wp.UpdateCompetence(compID, code, domain.CompetenceType(ctype), description)
	})
}

// RemoveCompetence deletes a Competence by id.
func (uc *WorkProgramContentUseCase) RemoveCompetence(ctx context.Context, actorID int64, actorRole string, wpID, compID int64) (*entities.WorkProgram, error) {
	return uc.mutate(ctx, actorID, actorRole, wpID, "competence_remove", func(wp *entities.WorkProgram) error {
		return wp.RemoveCompetence(compID)
	})
}

// --- Topic ---

// AddTopic appends a Topic.
func (uc *WorkProgramContentUseCase) AddTopic(ctx context.Context, actorID int64, actorRole string, wpID int64, in TopicContentInput) (*entities.WorkProgram, error) {
	return uc.mutate(ctx, actorID, actorRole, wpID, "topic_add", func(wp *entities.WorkProgram) error {
		t, err := entities.NewTopic(in.toDomain())
		if err != nil {
			return err
		}
		return wp.AddTopic(t)
	})
}

// UpdateTopic edits a Topic by id.
func (uc *WorkProgramContentUseCase) UpdateTopic(ctx context.Context, actorID int64, actorRole string, wpID, topicID int64, in TopicContentInput) (*entities.WorkProgram, error) {
	return uc.mutate(ctx, actorID, actorRole, wpID, "topic_update", func(wp *entities.WorkProgram) error {
		return wp.UpdateTopic(topicID, in.toDomain())
	})
}

// RemoveTopic deletes a Topic by id.
func (uc *WorkProgramContentUseCase) RemoveTopic(ctx context.Context, actorID int64, actorRole string, wpID, topicID int64) (*entities.WorkProgram, error) {
	return uc.mutate(ctx, actorID, actorRole, wpID, "topic_remove", func(wp *entities.WorkProgram) error {
		return wp.RemoveTopic(topicID)
	})
}

// --- AssessmentCriterion ---

// AddAssessment appends an AssessmentCriterion.
func (uc *WorkProgramContentUseCase) AddAssessment(ctx context.Context, actorID int64, actorRole string, wpID int64, in AssessmentContentInput) (*entities.WorkProgram, error) {
	return uc.mutate(ctx, actorID, actorRole, wpID, "assessment_add", func(wp *entities.WorkProgram) error {
		a, err := entities.NewAssessmentCriterion(in.toDomain())
		if err != nil {
			return err
		}
		return wp.AddAssessment(a)
	})
}

// UpdateAssessment edits an AssessmentCriterion by id.
func (uc *WorkProgramContentUseCase) UpdateAssessment(ctx context.Context, actorID int64, actorRole string, wpID, assessmentID int64, in AssessmentContentInput) (*entities.WorkProgram, error) {
	return uc.mutate(ctx, actorID, actorRole, wpID, "assessment_update", func(wp *entities.WorkProgram) error {
		return wp.UpdateAssessment(assessmentID, in.toDomain())
	})
}

// RemoveAssessment deletes an AssessmentCriterion by id.
func (uc *WorkProgramContentUseCase) RemoveAssessment(ctx context.Context, actorID int64, actorRole string, wpID, assessmentID int64) (*entities.WorkProgram, error) {
	return uc.mutate(ctx, actorID, actorRole, wpID, "assessment_remove", func(wp *entities.WorkProgram) error {
		return wp.RemoveAssessment(assessmentID)
	})
}

// --- Reference ---

// AddReference appends a Reference.
func (uc *WorkProgramContentUseCase) AddReference(ctx context.Context, actorID int64, actorRole string, wpID int64, in ReferenceContentInput) (*entities.WorkProgram, error) {
	return uc.mutate(ctx, actorID, actorRole, wpID, "reference_add", func(wp *entities.WorkProgram) error {
		r, err := entities.NewReference(in.toDomain())
		if err != nil {
			return err
		}
		return wp.AddReference(r)
	})
}

// UpdateReference edits a Reference by id.
func (uc *WorkProgramContentUseCase) UpdateReference(ctx context.Context, actorID int64, actorRole string, wpID, referenceID int64, in ReferenceContentInput) (*entities.WorkProgram, error) {
	return uc.mutate(ctx, actorID, actorRole, wpID, "reference_update", func(wp *entities.WorkProgram) error {
		return wp.UpdateReference(referenceID, in.toDomain())
	})
}

// RemoveReference deletes a Reference by id.
func (uc *WorkProgramContentUseCase) RemoveReference(ctx context.Context, actorID int64, actorRole string, wpID, referenceID int64) (*entities.WorkProgram, error) {
	return uc.mutate(ctx, actorID, actorRole, wpID, "reference_remove", func(wp *entities.WorkProgram) error {
		return wp.RemoveReference(referenceID)
	})
}
