package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// ErrDocumentForbidden is surfaced by the workflow use cases when the actor is
// not authorized for a transition. Handlers errors.Is it to map к a stable 4xx
// без string parsing. The companion sentinel ErrDocumentNotFound is declared
// alongside the repository interfaces in this package (document_repository.go).
//
// Issue: #227
var ErrDocumentForbidden = errors.New("documents: actor not authorized for transition")

// workflowRepo is the narrow port the three v0.148.0 workflow use
// cases share: load + write back. Cannot live в a smaller interface
// because все three use the same shape — DRY beats hyper-narrow ports.
type workflowRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.Document, error)
	Update(ctx context.Context, d *entities.Document) error
}

// --- Submit ---

// SubmitDocumentInput is the public DTO.
type SubmitDocumentInput struct {
	ID int64
}

// SubmitDocumentUseCase moves a draft document into the approval
// queue. Authorization: author OR Methodist/Secretary/Admin role.
//
// Issue: #227
type SubmitDocumentUseCase struct {
	repo  workflowRepo
	audit AuditSink
	clock func() time.Time
}

// NewSubmitDocumentUseCase wires the use case. RED stub for v0.148.0 —
// Execute returns ErrDocumentNotFound unconditionally so the paired
// RED tests fail с the right sentinels; GREEN replaces the body с the
// real flow.
func NewSubmitDocumentUseCase(repo workflowRepo, audit AuditSink, clock func() time.Time) *SubmitDocumentUseCase {
	if repo == nil {
		panic("documents: NewSubmitDocumentUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &SubmitDocumentUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute runs the submit flow:
//  1. Load by ID; not-found → 'not_found' denial.
//  2. Authorize: edit-roles (Methodist/Secretary/Admin) or author —
//     otherwise 'forbidden' denial + ErrDocumentForbidden.
//  3. Apply Submit; ErrCannotSubmit → 'not_draft' denial.
//  4. Persist; transport errors propagate без success audit.
func (uc *SubmitDocumentUseCase) Execute(ctx context.Context, actorID int64, role entities.UserRole, in SubmitDocumentInput) (*entities.Document, error) {
	d, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			emitAudit(uc.audit, ctx, "document.submit_denied", denialFields(actorID, in.ID, "not_found"))
		}
		return nil, err
	}
	if !canSubmit(actorID, role, d) {
		emitAudit(uc.audit, ctx, "document.submit_denied", denialFields(actorID, in.ID, "forbidden"))
		return nil, ErrDocumentForbidden
	}
	if err := d.Submit(actorID, uc.clock()); err != nil {
		if errors.Is(err, entities.ErrCannotSubmit) {
			emitAudit(uc.audit, ctx, "document.submit_denied", denialFields(actorID, in.ID, "not_draft"))
		}
		return nil, err
	}
	if err := uc.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	emitAudit(uc.audit, ctx, "document.submitted", map[string]any{
		"actor_user_id":      actorID,
		auditFieldDocumentID: d.ID,
		"status":             string(d.Status),
	})
	return d, nil
}

// canSubmit encodes the authorization rule: Methodist/Secretary/Admin
// can submit any document, Teacher only owns. Mirror к
// Document.CanBeEditedBy but explicit at the usecase boundary so the
// authorization audit reason is unambiguous.
func canSubmit(actorID int64, role entities.UserRole, d *entities.Document) bool {
	switch role {
	case entities.RoleMethodist, entities.RoleAcademicSecretary, entities.RoleSystemAdmin:
		return true
	case entities.RoleTeacher:
		return actorID == d.AuthorID
	default:
		return false
	}
}

// --- Approve ---

// ApproveDocumentInput is the public DTO.
type ApproveDocumentInput struct {
	ID int64
}

// ApproveDocumentUseCase advances an approval-queue document к the
// approved state. Admin-only by construction — route-level
// RequireRole gates; entity Approve enforces the status invariant.
//
// Issue: #227
type ApproveDocumentUseCase struct {
	repo  workflowRepo
	audit AuditSink
	clock func() time.Time
}

// NewApproveDocumentUseCase wires the use case (RED stub).
func NewApproveDocumentUseCase(repo workflowRepo, audit AuditSink, clock func() time.Time) *ApproveDocumentUseCase {
	if repo == nil {
		panic("documents: NewApproveDocumentUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &ApproveDocumentUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute runs the approve flow:
//  1. Load by ID; not-found → 'not_found' denial.
//  2. Apply Approve; ErrCannotApprove → 'not_approval' denial.
//  3. Persist; transport errors propagate без success audit.
func (uc *ApproveDocumentUseCase) Execute(ctx context.Context, adminID int64, in ApproveDocumentInput) (*entities.Document, error) {
	d, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			emitAudit(uc.audit, ctx, "document.approve_denied", denialFields(adminID, in.ID, "not_found"))
		}
		return nil, err
	}
	if err := d.Approve(adminID, uc.clock()); err != nil {
		if errors.Is(err, entities.ErrCannotApprove) {
			emitAudit(uc.audit, ctx, "document.approve_denied", denialFields(adminID, in.ID, "not_approval"))
		}
		return nil, err
	}
	if err := uc.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	emitAudit(uc.audit, ctx, "document.approved", map[string]any{
		"actor_user_id":      adminID,
		auditFieldDocumentID: d.ID,
		"status":             string(d.Status),
	})
	return d, nil
}

// --- Reject ---

// RejectDocumentInput is the public DTO. Reason is the raw string
// before VO validation — the use case validates via
// entities.NewRejectionReason.
type RejectDocumentInput struct {
	ID     int64
	Reason string
}

// RejectDocumentUseCase marks an approval-queue document as rejected
// с обоснованием. Admin-only by route gate.
//
// Issue: #227
type RejectDocumentUseCase struct {
	repo  workflowRepo
	audit AuditSink
	clock func() time.Time
}

// NewRejectDocumentUseCase wires the use case (RED stub).
func NewRejectDocumentUseCase(repo workflowRepo, audit AuditSink, clock func() time.Time) *RejectDocumentUseCase {
	if repo == nil {
		panic("documents: NewRejectDocumentUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &RejectDocumentUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute runs the reject flow:
//  1. Load by ID; not-found → 'not_found' denial.
//  2. Validate RejectionReason VO; invalid → 'invalid_reason' denial.
//  3. Apply Reject; ErrCannotReject → 'not_approval' denial.
//  4. Persist; transport errors propagate без success audit.
func (uc *RejectDocumentUseCase) Execute(ctx context.Context, adminID int64, in RejectDocumentInput) (*entities.Document, error) {
	d, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			emitAudit(uc.audit, ctx, "document.reject_denied", denialFields(adminID, in.ID, "not_found"))
		}
		return nil, err
	}
	reason, err := entities.NewRejectionReason(in.Reason)
	if err != nil {
		emitAudit(uc.audit, ctx, "document.reject_denied", denialFields(adminID, in.ID, "invalid_reason"))
		return nil, err
	}
	if err := d.Reject(adminID, reason, uc.clock()); err != nil {
		if errors.Is(err, entities.ErrCannotReject) {
			emitAudit(uc.audit, ctx, "document.reject_denied", denialFields(adminID, in.ID, "not_approval"))
		}
		return nil, err
	}
	if err := uc.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	emitAudit(uc.audit, ctx, "document.rejected", map[string]any{
		"actor_user_id":      adminID,
		auditFieldDocumentID: d.ID,
		"status":             string(d.Status),
		"rejected_reason":    reason.String(),
	})
	return d, nil
}
