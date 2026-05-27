package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// DiscardDraftWorkProgramInput is the public DTO.
type DiscardDraftWorkProgramInput struct {
	ID int64
}

// discardDraftWorkProgramRepo is the narrow port: load by id (so we
// can authorize against AuthorID) + write back the archived aggregate.
type discardDraftWorkProgramRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.WorkProgram, error)
	Update(ctx context.Context, wp *entities.WorkProgram) error
}

// DiscardDraftWorkProgramUseCase archives a draft WorkProgram without
// going through approval — author abandons their own work. Authorized
// for the author or for a system_admin override (mirrors Submit's
// authorship predicate — methodist is intentionally excluded from
// non-author destructive actions even on draft state).
type DiscardDraftWorkProgramUseCase struct {
	repo  discardDraftWorkProgramRepo
	audit AuditSink
}

// NewDiscardDraftWorkProgramUseCase wires the use case. Repo is required.
func NewDiscardDraftWorkProgramUseCase(repo discardDraftWorkProgramRepo, audit AuditSink) *DiscardDraftWorkProgramUseCase {
	if repo == nil {
		panic("work_program: NewDiscardDraftWorkProgramUseCase requires non-nil repo")
	}
	return &DiscardDraftWorkProgramUseCase{repo: repo, audit: audit}
}

// Execute runs the discard flow:
//  1. Load by id; ErrWorkProgramNotFound → 'not_found' denial.
//  2. Authorize via isAuthorOrSystemAdmin; deny → 'forbidden' denial.
//  3. Apply wp.DiscardDraft(); ErrInvalidStatusTransition → 'not_draft'
//     denial (DiscardDraft is permitted only from draft per ADR-2 FSM —
//     archived approved WPs need the Archive path that preserves the
//     approver trail; pending_approval must Reject first so reason is
//     captured).
//  4. Persist via repo.Update. Transport errors propagate without audit.
func (uc *DiscardDraftWorkProgramUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in DiscardDraftWorkProgramInput) (*entities.WorkProgram, error) {
	wp, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrWorkProgramNotFound) {
			emitAudit(uc.audit, ctx, "work_program.discard_denied",
				denialFields(actorID, in.ID, "not_found", ""))
		}
		return nil, err
	}

	if !isAuthorOrSystemAdmin(actorID, actorRole, wp.AuthorID()) {
		emitAudit(uc.audit, ctx, "work_program.discard_denied",
			denialFields(actorID, in.ID, "forbidden", wp.SpecialtyCode()))
		return nil, fmt.Errorf("%w: actor %d is not the author (%d) and not system_admin",
			domain.ErrWorkProgramScopeForbidden, actorID, wp.AuthorID())
	}

	if err := wp.DiscardDraft(); err != nil {
		if errors.Is(err, domain.ErrInvalidStatusTransition) {
			emitAudit(uc.audit, ctx, "work_program.discard_denied",
				denialFields(actorID, in.ID, "not_draft", wp.SpecialtyCode()))
		}
		return nil, err
	}

	if err := uc.repo.Update(ctx, wp); err != nil {
		return nil, err
	}

	emitAudit(uc.audit, ctx, "work_program.discarded", map[string]any{
		"actor_user_id":   actorID,
		"work_program_id": wp.ID(),
		"specialty_code":  wp.SpecialtyCode(),
		"status":          string(wp.Status()),
	})
	return wp, nil
}
