package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// RegisterDocumentInput is the public DTO. Number trim-validated в the
// entity layer via Register; usecase passes raw к the domain.
type RegisterDocumentInput struct {
	ID     int64
	Number string
}

// RegisterDocumentUseCase advances an approved document к the
// registered state с the supplied registration number + admin audit.
// Admin-only по route gate (academic_secretary, system_admin);
// entity Register method enforces the status invariant + number
// validation.
//
// Issue: #230
type RegisterDocumentUseCase struct {
	repo  workflowRepo
	audit AuditSink
	clock func() time.Time
}

// NewRegisterDocumentUseCase wires the use case.
func NewRegisterDocumentUseCase(repo workflowRepo, audit AuditSink, clock func() time.Time) *RegisterDocumentUseCase {
	if repo == nil {
		panic("documents: NewRegisterDocumentUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &RegisterDocumentUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute runs the register flow:
//  1. Load by ID; not-found → 'not_found' denial.
//  2. Apply Register; ErrCannotRegister → 'not_approved' denial;
//     ErrInvalidRegistrationNumber → 'invalid_number' denial.
//  3. Persist via repo.Update. Transport errors propagate без
//     success audit.
func (uc *RegisterDocumentUseCase) Execute(ctx context.Context, registrarID int64, in RegisterDocumentInput) (*entities.Document, error) {
	d, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			emitAudit(uc.audit, ctx, "document.register_denied", denialFields(registrarID, in.ID, "not_found"))
		}
		return nil, err
	}
	if err := d.Register(in.Number, registrarID, uc.clock()); err != nil {
		switch {
		case errors.Is(err, entities.ErrInvalidRegistrationNumber):
			emitAudit(uc.audit, ctx, "document.register_denied", denialFields(registrarID, in.ID, "invalid_number"))
		case errors.Is(err, entities.ErrCannotRegister):
			emitAudit(uc.audit, ctx, "document.register_denied", denialFields(registrarID, in.ID, "not_approved"))
		}
		return nil, err
	}
	if err := uc.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	emitAudit(uc.audit, ctx, "document.registered", map[string]any{
		"actor_user_id":       registrarID,
		auditFieldDocumentID:  d.ID,
		"status":              string(d.Status),
		"registration_number": *d.RegistrationNumber,
	})
	return d, nil
}
