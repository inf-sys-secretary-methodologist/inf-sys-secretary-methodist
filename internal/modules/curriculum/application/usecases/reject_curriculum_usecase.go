package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// RejectCurriculumInput is the public DTO. Reason is the admin's
// free-form rejection note — flows verbatim into the
// 'curriculum.rejected' audit event but is NOT persisted on the
// entity (ADR-3: audit-only). Handlers enforce non-empty reason
// at the boundary; the use case accepts any string so future
// callers (CLI, batch job) can submit unconventional reasons.
type RejectCurriculumInput struct {
	ID     int64
	Reason string
}

type rejectCurriculumRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.Curriculum, error)
	Update(ctx context.Context, c *entities.Curriculum) error
}

// RejectCurriculumUseCase moves a pending_approval curriculum back
// to draft so the methodist may revise and re-submit. The cycle
// Reject → Edit → Submit is the rework loop; the audit log captures
// the admin's reason for forensic and compliance purposes.
//
// Admin-only by construction (mirror ApproveCurriculumUseCase).
type RejectCurriculumUseCase struct {
	repo  rejectCurriculumRepo
	audit AuditSink
	clock func() time.Time
}

// NewRejectCurriculumUseCase wires the use case. Repo is required.
func NewRejectCurriculumUseCase(repo rejectCurriculumRepo, audit AuditSink, clock func() time.Time) *RejectCurriculumUseCase {
	if repo == nil {
		panic("curriculum: NewRejectCurriculumUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &RejectCurriculumUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute runs the reject flow:
//  1. Load by id; ErrCurriculumNotFound → 'not_found' denial.
//  2. Apply Reject(now); ErrCannotReject → 'not_pending' denial.
//  3. Persist via repo.Update.
//  4. Success → 'curriculum.rejected' audit with the admin's
//     free-form reason in the dedicated field.
func (uc *RejectCurriculumUseCase) Execute(
	ctx context.Context, adminID int64, in RejectCurriculumInput,
) (*entities.Curriculum, error) {
	c, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrCurriculumNotFound) {
			emitAudit(uc.audit, ctx, "curriculum.reject_denied",
				denialFields(adminID, in.ID, "not_found", ""))
		}
		return nil, err
	}

	if err := c.Reject(uc.clock()); err != nil {
		if errors.Is(err, entities.ErrCannotReject) {
			emitAudit(uc.audit, ctx, "curriculum.reject_denied",
				denialFields(adminID, in.ID, "not_pending", ""))
		}
		return nil, err
	}

	if err := uc.repo.Update(ctx, c); err != nil {
		return nil, err
	}

	emitAudit(uc.audit, ctx, "curriculum.rejected", map[string]any{
		"actor_user_id": adminID,
		"curriculum_id": c.ID,
		"code":          c.Code(),
		"status":        string(c.Status()),
		// reason field is distinct from the *_denied 'reason' (which
		// names the canonical denial cause). Here it carries the
		// admin's free-form rejection note.
		"reason": in.Reason,
	})
	return c, nil
}
