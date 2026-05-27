package usecases

import (
	"context"
	"errors"
	"fmt"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// CreateWorkProgramInput is the public request DTO. The actor
// (created_by → AuthorID) and actor's role are supplied as separate
// arguments to Execute so handlers wire the JWT subject + role
// explicitly rather than through the same struct that may be
// deserialised from untrusted JSON.
type CreateWorkProgramInput struct {
	DisciplineID       int64
	SpecialtyCode      string
	ApplicableFromYear int
	Title              string
	Annotation         string
}

// createWorkProgramRepo is the narrow port the use case requires from
// the persistence layer. Defining it here (rather than importing the
// wide *WorkProgramRepository) keeps use-case tests free of GetByID /
// Update / Delete / List wiring they do not exercise.
type createWorkProgramRepo interface {
	Save(ctx context.Context, wp *entities.WorkProgram) error
}

// CreateWorkProgramUseCase persists a fresh draft WorkProgram and
// emits the matching audit event. Role gate per ADR-018 ADR-5: only
// teacher / methodist / system_admin may create. academic_secretary is
// view-only on РПД (curriculum is their author surface); student is
// denied. methodist is allowed as a backup author per ADR-018 ADR-5
// ("резервно creates если teacher не успевает").
type CreateWorkProgramUseCase struct {
	repo  createWorkProgramRepo
	audit AuditSink
}

// NewCreateWorkProgramUseCase wires the use case. The repo is required
// (non-nil): a nil dependency would let requests reach a panic deeper
// in the call stack instead of failing during DI wiring. Nil audit
// sink is tolerated (tests may opt out).
func NewCreateWorkProgramUseCase(repo createWorkProgramRepo, audit AuditSink) *CreateWorkProgramUseCase {
	if repo == nil {
		panic("work_program: NewCreateWorkProgramUseCase requires non-nil repo")
	}
	return &CreateWorkProgramUseCase{repo: repo, audit: audit}
}

// Execute runs the use case end-to-end:
//  1. Role gate per ADR-018 ADR-5 (teacher / methodist / system_admin).
//  2. Build the entity through NewWorkProgram (invariant gate).
//  3. Persist via repo.Save (which translates uniqueness violations to
//     ErrWorkProgramIdentityExists).
//  4. Emit a forensic audit event reflecting success or domain denial.
//     Transport errors propagate without an audit event — the audit log
//     records policy decisions, not infrastructure outages.
//
// Returns the persisted entity (with ID populated) on success. On
// domain failures the entity is nil and the returned error wraps
// either ErrWorkProgramScopeForbidden / ErrInvalidWorkProgram /
// ErrWorkProgramIdentityExists so errors.Is resolves cleanly in
// handler error mapping.
func (uc *CreateWorkProgramUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in CreateWorkProgramInput) (*entities.WorkProgram, error) {
	if !isAllowedToCreateWorkProgram(actorRole) {
		emitAudit(uc.audit, ctx, "work_program.create_denied",
			denialFields(actorID, 0, "forbidden_role", in.SpecialtyCode))
		return nil, fmt.Errorf("%w: role %q cannot create work program", domain.ErrWorkProgramScopeForbidden, actorRole)
	}

	wp, err := entities.NewWorkProgram(entities.NewWorkProgramInput{
		DisciplineID:       in.DisciplineID,
		SpecialtyCode:      in.SpecialtyCode,
		ApplicableFromYear: in.ApplicableFromYear,
		Title:              in.Title,
		Annotation:         in.Annotation,
		AuthorID:           actorID,
	})
	if err != nil {
		// work_program_id is 0 — the row was never built. denialFields
		// tolerates a zero id; operators reading the audit log still see
		// actor + reason + specialty_code.
		emitAudit(uc.audit, ctx, "work_program.create_denied",
			denialFields(actorID, 0, "invalid", in.SpecialtyCode))
		return nil, err
	}

	if err := uc.repo.Save(ctx, wp); err != nil {
		if errors.Is(err, repositories.ErrWorkProgramIdentityExists) {
			emitAudit(uc.audit, ctx, "work_program.create_denied",
				denialFields(actorID, 0, "identity_conflict", in.SpecialtyCode))
		}
		return nil, err
	}

	emitAudit(uc.audit, ctx, "work_program.created", map[string]any{
		"actor_user_id":        actorID,
		"work_program_id":      wp.ID(),
		"specialty_code":       wp.SpecialtyCode(),
		"applicable_from_year": wp.ApplicableFromYear(),
		"discipline_id":        wp.DisciplineID(),
		"status":               string(wp.Status()),
	})
	return wp, nil
}

// isAllowedToCreateWorkProgram encodes the ADR-018 ADR-5 role matrix
// for the create operation. Typed against authDomain.RoleType so a
// typo in the role string would fail at compile time on the constant
// reference, not silently at runtime through default-deny.
func isAllowedToCreateWorkProgram(role string) bool {
	r := authDomain.RoleType(role)
	return r == authDomain.RoleTeacher ||
		r == authDomain.RoleMethodist ||
		r == authDomain.RoleSystemAdmin
}
