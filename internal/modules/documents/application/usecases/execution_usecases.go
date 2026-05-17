package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// --- Assign executor ---

// AssignExecutorInput is the public DTO. DueDate optional per ADR-2 (#232) —
// nil means «no hard deadline».
type AssignExecutorInput struct {
	ID         int64
	ExecutorID int64
	DueDate    *time.Time
}

// AssignExecutorUseCase shapes the executor assignment on an execution-
// status document. Status stays Execution — assign is a shape-only
// operation per ADR-1 (#232). Reassign overwrites prior assignment.
// Admin-only по route gate (academic_secretary, system_admin).
//
// Issue: #232
type AssignExecutorUseCase struct {
	repo  workflowRepo
	audit AuditSink
	clock func() time.Time
}

// NewAssignExecutorUseCase wires the use case.
func NewAssignExecutorUseCase(repo workflowRepo, audit AuditSink, clock func() time.Time) *AssignExecutorUseCase {
	if repo == nil {
		panic("documents: NewAssignExecutorUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &AssignExecutorUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute runs the assign-executor flow:
//  1. Validate executorID > 0 ('invalid_executor' denial — 422 client error).
//  2. Load by ID; not-found → 'not_found' denial.
//  3. Apply AssignExecutor; ErrCannotAssignExecutor → 'not_execution' denial.
//  4. Persist via repo.Update. Transport errors propagate без
//     success audit.
func (uc *AssignExecutorUseCase) Execute(ctx context.Context, actorID int64, in AssignExecutorInput) (*entities.Document, error) {
	if in.ExecutorID <= 0 {
		emitAudit(uc.audit, ctx, "document.assign_executor_denied", denialFields(actorID, in.ID, "invalid_executor"))
		return nil, ErrInvalidExecutor
	}
	d, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			emitAudit(uc.audit, ctx, "document.assign_executor_denied", denialFields(actorID, in.ID, "not_found"))
		}
		return nil, err
	}
	if err := d.AssignExecutor(in.ExecutorID, in.DueDate, actorID, uc.clock()); err != nil {
		if errors.Is(err, entities.ErrCannotAssignExecutor) {
			emitAudit(uc.audit, ctx, "document.assign_executor_denied", denialFields(actorID, in.ID, "not_execution"))
		}
		return nil, err
	}
	if err := uc.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	emitAudit(uc.audit, ctx, "document.executor_assigned", map[string]any{
		"actor_user_id":      actorID,
		auditFieldDocumentID: d.ID,
		"executor_user_id":   in.ExecutorID,
		"status":             string(d.Status),
	})
	return d, nil
}

// ErrInvalidExecutor signals invalid executor id (must be > 0). Wrapped
// at the usecase boundary so handler maps it к 422 без depending on
// domain (executor identity is a usecase concern, not domain invariant).
//
// Issue: #232
var ErrInvalidExecutor = errors.New("documents: executor id must be > 0")

// --- Mark executed ---

// MarkExecutedInput is the public DTO.
type MarkExecutedInput struct {
	ID int64
}

// MarkExecutedUseCase advances an execution-status document к executed.
// Admin-only по route gate (academic_secretary, system_admin); entity
// MarkExecuted enforces the status invariant.
//
// Issue: #232
type MarkExecutedUseCase struct {
	repo  workflowRepo
	audit AuditSink
	clock func() time.Time
}

// NewMarkExecutedUseCase wires the use case.
func NewMarkExecutedUseCase(repo workflowRepo, audit AuditSink, clock func() time.Time) *MarkExecutedUseCase {
	if repo == nil {
		panic("documents: NewMarkExecutedUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &MarkExecutedUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute runs the mark-executed flow:
//  1. Load by ID; not-found → 'not_found' denial.
//  2. Apply MarkExecuted; ErrCannotMarkExecuted → 'not_execution' denial.
//  3. Persist via repo.Update. Transport errors propagate без
//     success audit.
func (uc *MarkExecutedUseCase) Execute(ctx context.Context, actorID int64, in MarkExecutedInput) (*entities.Document, error) {
	d, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			emitAudit(uc.audit, ctx, "document.mark_executed_denied", denialFields(actorID, in.ID, "not_found"))
		}
		return nil, err
	}
	if err := d.MarkExecuted(actorID, uc.clock()); err != nil {
		if errors.Is(err, entities.ErrCannotMarkExecuted) {
			emitAudit(uc.audit, ctx, "document.mark_executed_denied", denialFields(actorID, in.ID, "not_execution"))
		}
		return nil, err
	}
	if err := uc.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	emitAudit(uc.audit, ctx, "document.executed", map[string]any{
		"actor_user_id":      actorID,
		auditFieldDocumentID: d.ID,
		"status":             string(d.Status),
	})
	return d, nil
}
