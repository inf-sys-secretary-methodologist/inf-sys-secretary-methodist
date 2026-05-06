package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// ApproveCurriculumInput is the public DTO. Admin id flows through
// Execute as a positional argument (separate from the request body)
// so handlers wire the JWT subject directly.
type ApproveCurriculumInput struct {
	ID int64
}

// approveCurriculumRepo is the narrow port: load + write back.
type approveCurriculumRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.Curriculum, error)
	Update(ctx context.Context, c *entities.Curriculum) error
}

// ApproveCurriculumUseCase moves a pending_approval curriculum into
// the approved state, recording the admin's identity + timestamp.
//
// Admin-only by construction: the use case takes only adminID (no
// isAdmin flag). Route-level RequireRole(SystemAdmin) middleware and
// the handler whitelist are the access gates; the entity Approve
// method enforces the status invariant plus a non-zero adminID
// guard as defense in depth against a silent admin scenario.
type ApproveCurriculumUseCase struct {
	repo  approveCurriculumRepo
	audit AuditSink
	clock func() time.Time
}

// NewApproveCurriculumUseCase wires the use case. Repo is required.
func NewApproveCurriculumUseCase(repo approveCurriculumRepo, audit AuditSink, clock func() time.Time) *ApproveCurriculumUseCase {
	if repo == nil {
		panic("curriculum: NewApproveCurriculumUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &ApproveCurriculumUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute runs the approve flow:
//  1. Load by id; ErrCurriculumNotFound → 'not_found' denial.
//  2. Apply Approve(adminID, now); ErrCannotApprove →
//     'not_pending' denial.
//  3. Persist via repo.Update. Transport errors propagate without
//     audit (audit log = policy decisions, not infra outages).
func (uc *ApproveCurriculumUseCase) Execute(
	ctx context.Context, adminID int64, in ApproveCurriculumInput,
) (*entities.Curriculum, error) {
	c, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrCurriculumNotFound) {
			emitAudit(uc.audit, ctx, "curriculum.approve_denied",
				denialFields(adminID, in.ID, "not_found", ""))
		}
		return nil, err
	}

	if err := c.Approve(adminID, uc.clock()); err != nil {
		if errors.Is(err, entities.ErrCannotApprove) {
			emitAudit(uc.audit, ctx, "curriculum.approve_denied",
				denialFields(adminID, in.ID, "not_pending", ""))
		}
		return nil, err
	}

	if err := uc.repo.Update(ctx, c); err != nil {
		return nil, err
	}

	emitAudit(uc.audit, ctx, "curriculum.approved", map[string]any{
		"actor_user_id": adminID,
		"curriculum_id": c.ID,
		"code":          c.Code(),
		"status":        string(c.Status()),
	})
	return c, nil
}
