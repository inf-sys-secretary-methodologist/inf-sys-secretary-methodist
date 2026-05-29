package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// RecordMinobrnaukiOrderInput is the public request DTO for recording a
// приказ Минобрнауки (ADR-11). The actor (→ UploadedBy) and actor role
// are supplied as separate Execute arguments, not in this struct, so
// handlers wire the JWT subject + role explicitly rather than through a
// struct deserialised from untrusted JSON.
type RecordMinobrnaukiOrderInput struct {
	OrderNumber            string
	Title                  string
	PublishedAt            time.Time
	DocumentID             *int64
	ChangeScope            string // "minor" / "major"; mapped to the domain enum
	Summary                string
	AffectedWorkProgramIDs []int64
}

// recordMinobrnaukiOrderRepo is the narrow persistence port this use
// case needs — Save only. Defining it here (not the wide
// MinobrnaukiOrderRepository) keeps use-case tests free of GetByID /
// List / FindAffected wiring they do not exercise.
type recordMinobrnaukiOrderRepo interface {
	Save(ctx context.Context, order *entities.MinobrnaukiOrder, affectedWorkProgramIDs []int64) error
}

// orderRevisionTrigger is the optional collaborator fired after an order
// is recorded: it drives the affected РПД into needs_revision and
// delegates teacher tasks (ADR-11 pipeline step 2). Defined here on the
// consumer side so Record depends on the narrow capability, not the
// concrete TriggerOrderRevisionsUseCase.
type orderRevisionTrigger interface {
	Execute(ctx context.Context, actorID, orderID int64, orderNumber string, affectedWorkProgramIDs []int64) (TriggerOrderRevisionsResult, error)
}

// RecordMinobrnaukiOrderUseCase records a Минобрнауки order artifact and
// emits the matching audit event. Role gate per ADR-11: only methodist /
// academic_secretary / system_admin may record an order; teacher and
// student are denied.
type RecordMinobrnaukiOrderUseCase struct {
	repo            recordMinobrnaukiOrderRepo
	audit           AuditSink
	revisionTrigger orderRevisionTrigger
}

// NewRecordMinobrnaukiOrderUseCase wires the use case. repo is required
// (non-nil) so a missing dependency fails at DI wiring, not deep in the
// call stack. Nil audit sink is tolerated (tests may opt out).
func NewRecordMinobrnaukiOrderUseCase(repo recordMinobrnaukiOrderRepo, audit AuditSink) *RecordMinobrnaukiOrderUseCase {
	if repo == nil {
		panic("work_program: NewRecordMinobrnaukiOrderUseCase requires non-nil repo")
	}
	return &RecordMinobrnaukiOrderUseCase{repo: repo, audit: audit}
}

// WithRevisionTrigger attaches the collaborator fired after a successful
// record to drive affected РПД into needs_revision + delegate teacher
// tasks. Optional and chainable — nil leaves order recording standalone
// (a missing trigger never blocks recording an order).
func (uc *RecordMinobrnaukiOrderUseCase) WithRevisionTrigger(t orderRevisionTrigger) *RecordMinobrnaukiOrderUseCase {
	uc.revisionTrigger = t
	return uc
}

// Execute runs the use case end-to-end:
//  1. Role gate per ADR-11 (methodist / academic_secretary / system_admin).
//  2. Build the entity through NewMinobrnaukiOrder (invariant gate);
//     UploadedBy is the actor, and the wire change_scope string is mapped
//     to the domain enum (an unknown value fails the domain invariant).
//  3. Persist via repo.Save together with the affected-work-program ids.
//  4. Emit a forensic audit event reflecting success or domain denial.
//     Transport errors propagate without an audit event — the audit log
//     records policy decisions, not infrastructure outages.
//
// On domain failures the order is nil and the error wraps either
// ErrMinobrnaukiOrderScopeForbidden or ErrInvalidMinobrnaukiOrder so
// errors.Is resolves cleanly in handler error mapping.
func (uc *RecordMinobrnaukiOrderUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in RecordMinobrnaukiOrderInput) (*entities.MinobrnaukiOrder, error) {
	if !isAllowedToRecordMinobrnaukiOrder(actorRole) {
		emitOrderAudit(uc.audit, ctx, "minobrnauki_order.record_denied",
			orderDenialFields(actorID, 0, "forbidden_role", in.OrderNumber))
		return nil, fmt.Errorf("%w: role %q cannot record minobrnauki order", domain.ErrMinobrnaukiOrderScopeForbidden, actorRole)
	}

	order, err := entities.NewMinobrnaukiOrder(entities.NewMinobrnaukiOrderInput{
		OrderNumber: in.OrderNumber,
		Title:       in.Title,
		PublishedAt: in.PublishedAt,
		DocumentID:  in.DocumentID,
		ChangeScope: domain.MinobrnaukiOrderChangeScope(in.ChangeScope),
		Summary:     in.Summary,
		UploadedBy:  actorID,
	})
	if err != nil {
		emitOrderAudit(uc.audit, ctx, "minobrnauki_order.record_denied",
			orderDenialFields(actorID, 0, "invalid", in.OrderNumber))
		return nil, err
	}

	if err := uc.repo.Save(ctx, order, in.AffectedWorkProgramIDs); err != nil {
		return nil, err
	}

	emitOrderAudit(uc.audit, ctx, "minobrnauki_order.recorded",
		orderSuccessFields(actorID, order.ID(), order.OrderNumber(), string(order.ChangeScope())))

	_ = uc.revisionTrigger // RED placeholder — trigger fire lands in GREEN
	return order, nil
}
