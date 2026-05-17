package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// --- Start routing ---

// StartRoutingInput is the public DTO.
type StartRoutingInput struct {
	ID int64
}

// StartRoutingUseCase advances a registered document к the routing
// queue. Admin-only по route gate (academic_secretary, system_admin);
// entity SendToRouting enforces the status invariant.
//
// Issue: #231
type StartRoutingUseCase struct {
	repo  workflowRepo
	audit AuditSink
	clock func() time.Time
}

// NewStartRoutingUseCase wires the use case.
func NewStartRoutingUseCase(repo workflowRepo, audit AuditSink, clock func() time.Time) *StartRoutingUseCase {
	if repo == nil {
		panic("documents: NewStartRoutingUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &StartRoutingUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute runs the start-routing flow:
//  1. Load by ID; not-found → 'not_found' denial.
//  2. Apply SendToRouting; ErrCannotRoute → 'not_registered' denial.
//  3. Persist via repo.Update. Transport errors propagate без
//     success audit.
func (uc *StartRoutingUseCase) Execute(ctx context.Context, routerID int64, in StartRoutingInput) (*entities.Document, error) {
	d, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			emitAudit(uc.audit, ctx, "document.route_denied", denialFields(routerID, in.ID, "not_found"))
		}
		return nil, err
	}
	if err := d.SendToRouting(routerID, uc.clock()); err != nil {
		if errors.Is(err, entities.ErrCannotRoute) {
			emitAudit(uc.audit, ctx, "document.route_denied", denialFields(routerID, in.ID, "not_registered"))
		}
		return nil, err
	}
	if err := uc.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	emitAudit(uc.audit, ctx, "document.routed", map[string]any{
		"actor_user_id":      routerID,
		auditFieldDocumentID: d.ID,
		"status":             string(d.Status),
	})
	return d, nil
}

// --- Sign visa ---

// SignVisaInput is the public DTO.
type SignVisaInput struct {
	ID int64
}

// SignVisaUseCase advances a routing-queue document к the execution
// state. Single-step visa per ADR-1 — one approver per document.
// Admin-only по route gate; entity SignVisa enforces the status
// invariant.
//
// Issue: #231
type SignVisaUseCase struct {
	repo  workflowRepo
	audit AuditSink
	clock func() time.Time
}

// NewSignVisaUseCase wires the use case.
func NewSignVisaUseCase(repo workflowRepo, audit AuditSink, clock func() time.Time) *SignVisaUseCase {
	if repo == nil {
		panic("documents: NewSignVisaUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &SignVisaUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute runs the sign-visa flow:
//  1. Load by ID; not-found → 'not_found' denial.
//  2. Apply SignVisa; ErrCannotSignVisa → 'not_routing' denial.
//  3. Persist via repo.Update. Transport errors propagate без
//     success audit.
func (uc *SignVisaUseCase) Execute(ctx context.Context, visaID int64, in SignVisaInput) (*entities.Document, error) {
	d, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			emitAudit(uc.audit, ctx, "document.sign_visa_denied", denialFields(visaID, in.ID, "not_found"))
		}
		return nil, err
	}
	if err := d.SignVisa(visaID, uc.clock()); err != nil {
		if errors.Is(err, entities.ErrCannotSignVisa) {
			emitAudit(uc.audit, ctx, "document.sign_visa_denied", denialFields(visaID, in.ID, "not_routing"))
		}
		return nil, err
	}
	if err := uc.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	emitAudit(uc.audit, ctx, "document.visa_signed", map[string]any{
		"actor_user_id":      visaID,
		auditFieldDocumentID: d.ID,
		"status":             string(d.Status),
	})
	return d, nil
}
