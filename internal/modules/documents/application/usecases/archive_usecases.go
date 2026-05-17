package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// --- Archive ---

// ArchiveDocumentInput is the public DTO.
type ArchiveDocumentInput struct {
	ID int64
}

// ArchiveDocumentUseCase flips an executed document к the terminal
// archived state. Admin-only по route gate (academic_secretary,
// system_admin); entity Archive enforces the status invariant.
// Mirror к MarkExecutedUseCase pattern.
//
// Issue: #233
type ArchiveDocumentUseCase struct {
	repo  workflowRepo
	audit AuditSink
	clock func() time.Time
}

// NewArchiveDocumentUseCase wires the use case.
func NewArchiveDocumentUseCase(repo workflowRepo, audit AuditSink, clock func() time.Time) *ArchiveDocumentUseCase {
	if repo == nil {
		panic("documents: NewArchiveDocumentUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &ArchiveDocumentUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute runs the archive flow:
//  1. Load by ID; not-found → 'not_found' denial.
//  2. Apply Archive; ErrCannotArchive → 'not_executed' denial.
//  3. Persist via repo.Update. Transport errors propagate без
//     success audit.
func (uc *ArchiveDocumentUseCase) Execute(ctx context.Context, actorID int64, in ArchiveDocumentInput) (*entities.Document, error) {
	d, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			emitAudit(uc.audit, ctx, "document.archive_denied", denialFields(actorID, in.ID, "not_found"))
		}
		return nil, err
	}
	if err := d.Archive(actorID, uc.clock()); err != nil {
		if errors.Is(err, entities.ErrCannotArchive) {
			emitAudit(uc.audit, ctx, "document.archive_denied", denialFields(actorID, in.ID, "not_executed"))
		}
		return nil, err
	}
	if err := uc.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	emitAudit(uc.audit, ctx, "document.archived", map[string]any{
		"actor_user_id":      actorID,
		auditFieldDocumentID: d.ID,
		"status":             string(d.Status),
	})
	return d, nil
}

// --- Resubmit ---

// ResubmitDocumentInput is the public DTO.
type ResubmitDocumentInput struct {
	ID int64
}

// ResubmitDocumentUseCase returns a rejected document к the draft
// cycle. Authorization: author OR Methodist/Secretary/Admin role
// (mirror к SubmitDocumentUseCase pattern, NOT admin-only). Entity
// Resubmit enforces the status invariant.
//
// Issue: #233
type ResubmitDocumentUseCase struct {
	repo  workflowRepo
	audit AuditSink
	clock func() time.Time
}

// NewResubmitDocumentUseCase wires the use case.
func NewResubmitDocumentUseCase(repo workflowRepo, audit AuditSink, clock func() time.Time) *ResubmitDocumentUseCase {
	if repo == nil {
		panic("documents: NewResubmitDocumentUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &ResubmitDocumentUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute runs the resubmit flow:
//  1. Load by ID; not-found → 'not_found' denial.
//  2. Authorize: edit-roles (Methodist/Secretary/Admin) or author —
//     otherwise 'forbidden' denial + ErrDocumentForbidden.
//  3. Apply Resubmit; ErrCannotResubmit → 'not_rejected' denial.
//  4. Persist; transport errors propagate без success audit.
func (uc *ResubmitDocumentUseCase) Execute(ctx context.Context, actorID int64, role entities.UserRole, in ResubmitDocumentInput) (*entities.Document, error) {
	d, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			emitAudit(uc.audit, ctx, "document.resubmit_denied", denialFields(actorID, in.ID, "not_found"))
		}
		return nil, err
	}
	if !canSubmit(actorID, role, d) {
		emitAudit(uc.audit, ctx, "document.resubmit_denied", denialFields(actorID, in.ID, "forbidden"))
		return nil, ErrDocumentForbidden
	}
	if err := d.Resubmit(actorID, uc.clock()); err != nil {
		if errors.Is(err, entities.ErrCannotResubmit) {
			emitAudit(uc.audit, ctx, "document.resubmit_denied", denialFields(actorID, in.ID, "not_rejected"))
		}
		return nil, err
	}
	if err := uc.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	emitAudit(uc.audit, ctx, "document.resubmitted", map[string]any{
		"actor_user_id":      actorID,
		auditFieldDocumentID: d.ID,
		"status":             string(d.Status),
	})
	return d, nil
}
