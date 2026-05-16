package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// Sentinels surfaced by the workflow use cases. Wrapped so handlers
// can errors.Is them and map к stable 4xx responses без string parsing.
//
// Issue: #227
var (
	ErrDocumentNotFound  = errors.New("documents: not found")
	ErrDocumentForbidden = errors.New("documents: actor not authorized for transition")
)

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

// Execute runs the submit flow (RED stub — references emitAudit/
// denialFields so golangci `unused` stays satisfied; GREEN replaces
// the body с the real flow).
func (uc *SubmitDocumentUseCase) Execute(ctx context.Context, actorID int64, role entities.UserRole, in SubmitDocumentInput) (*entities.Document, error) {
	emitAudit(uc.audit, ctx, "document.submit_denied", denialFields(actorID, in.ID, "stub"))
	return nil, ErrDocumentNotFound
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

// Execute runs the approve flow (RED stub).
func (uc *ApproveDocumentUseCase) Execute(ctx context.Context, adminID int64, in ApproveDocumentInput) (*entities.Document, error) {
	emitAudit(uc.audit, ctx, "document.approve_denied", denialFields(adminID, in.ID, "stub"))
	return nil, ErrDocumentNotFound
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

// Execute runs the reject flow (RED stub).
func (uc *RejectDocumentUseCase) Execute(ctx context.Context, adminID int64, in RejectDocumentInput) (*entities.Document, error) {
	emitAudit(uc.audit, ctx, "document.reject_denied", denialFields(adminID, in.ID, "stub"))
	return nil, ErrDocumentNotFound
}
