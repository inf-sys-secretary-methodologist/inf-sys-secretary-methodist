package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// CreateCurriculumInput is the public request DTO for the use case.
// Title / Code / Specialty / Year / Description map directly to the
// NewCurriculum invariants; the actor (created_by) is supplied as a
// separate argument so handlers wire the JWT subject explicitly
// rather than through the same struct that may be deserialised from
// untrusted JSON.
type CreateCurriculumInput struct {
	Title       string
	Code        string
	Specialty   string
	Year        int
	Description string
}

// createCurriculumRepo is the narrow port the use case requires from
// the persistence layer. Defining it here (rather than importing the
// concrete *CurriculumRepositoryPG) keeps use-case tests free of a
// real DB.
type createCurriculumRepo interface {
	Save(ctx context.Context, c *entities.Curriculum) error
}

// CreateCurriculumUseCase persists a fresh draft curriculum and emits
// the matching audit event.
type CreateCurriculumUseCase struct {
	repo  createCurriculumRepo
	audit AuditSink
	clock func() time.Time
}

// NewCreateCurriculumUseCase wires the use case. The repo is required
// (non-nil): a nil dependency would let requests reach a panic deeper
// in the call stack instead of failing during DI wiring. Nil audit
// sink is tolerated (callers may opt out of audit during tests). A
// nil clock falls back to time.Now so production wiring can omit the
// argument when injection is unnecessary.
func NewCreateCurriculumUseCase(repo createCurriculumRepo, audit AuditSink, clock func() time.Time) *CreateCurriculumUseCase {
	if repo == nil {
		panic("curriculum: NewCreateCurriculumUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &CreateCurriculumUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute runs the use case end-to-end:
//  1. Build the entity through NewCurriculum (invariant gate).
//  2. Persist via repo.Save (which translates unique-constraint
//     violations to ErrCurriculumCodeExists).
//  3. Emit a forensic audit event reflecting success or domain
//     denial. Transport errors propagate without an audit event —
//     the audit log records policy decisions, not infrastructure
//     outages.
//
// Returns the persisted entity (with ID populated) on success. On
// domain failures the entity is nil and the returned error wraps
// either ErrInvalidCurriculum or ErrCurriculumCodeExists so
// errors.Is resolves cleanly in handler error mapping.
func (uc *CreateCurriculumUseCase) Execute(ctx context.Context, actorID int64, in CreateCurriculumInput) (*entities.Curriculum, error) {
	c, err := entities.NewCurriculum(entities.NewCurriculumParams{
		Title:       in.Title,
		Code:        in.Code,
		Specialty:   in.Specialty,
		Year:        in.Year,
		Description: in.Description,
		CreatedBy:   actorID,
		Now:         uc.clock(),
	})
	if err != nil {
		// curriculum_id is 0 because the row was never built — Create
		// is the only path where the entity itself fails before
		// allocation. denialFields tolerates a zero id; operators
		// reading the audit log learn the actor + reason regardless.
		emitAudit(uc.audit, ctx, "curriculum.create_denied", denialFields(actorID, 0, "invalid", in.Code))
		return nil, err
	}

	if err := uc.repo.Save(ctx, c); err != nil {
		if errors.Is(err, repositories.ErrCurriculumCodeExists) {
			emitAudit(uc.audit, ctx, "curriculum.create_denied", denialFields(actorID, 0, "code_conflict", c.Code()))
		}
		return nil, err
	}

	emitAudit(uc.audit, ctx, "curriculum.created", map[string]any{
		"actor_user_id": actorID,
		"curriculum_id": c.ID,
		"code":          c.Code(),
		"year":          c.Year(),
		"specialty":     c.Specialty(),
		"status":        string(c.Status()),
	})
	return c, nil
}
