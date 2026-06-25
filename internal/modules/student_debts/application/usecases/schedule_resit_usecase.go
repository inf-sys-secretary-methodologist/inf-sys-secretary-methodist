package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
)

// writeDebtRepo is the narrow port the write use cases need: load the
// aggregate, then persist the mutated state atomically (optimistic lock).
type writeDebtRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.StudentDebt, error)
	Update(ctx context.Context, debt *entities.StudentDebt) error
}

// ScheduleResitInput is the public DTO for scheduling a resit attempt.
type ScheduleResitInput struct {
	DebtID        int64
	ScheduledDate time.Time
	Examiner      string
}

// ScheduleResitUseCase schedules a resit attempt on a debt (EDIT_ROLES
// only) and, on success, best-effort notifies the student. The FSM guards
// (ErrDebtClosed / ErrInvalidTransition) live in the domain; this use case
// orchestrates load → mutate → persist → notify → audit.
type ScheduleResitUseCase struct {
	repo     writeDebtRepo
	notifier DebtNotifier
	audit    AuditSink
	now      func() time.Time
}

// NewScheduleResitUseCase wires the use case. repo is required; notifier
// and audit may be nil (no-op). now defaults to time.Now when nil.
func NewScheduleResitUseCase(repo writeDebtRepo, notifier DebtNotifier, audit AuditSink, now func() time.Time) *ScheduleResitUseCase {
	if repo == nil {
		panic("student_debts: NewScheduleResitUseCase requires non-nil repo")
	}
	if now == nil {
		now = time.Now
	}
	return &ScheduleResitUseCase{repo: repo, notifier: notifier, audit: audit, now: now}
}

// Execute schedules a resit and returns the updated aggregate:
// EDIT_ROLES gate → load → domain FSM (ScheduleResit) → persist → notify
// the student (best-effort) → audit. FSM errors (ErrDebtClosed /
// ErrInvalidTransition) and repository errors (version conflict, not
// found) propagate unchanged; the notification only fires after a
// successful persist so a rolled-back schedule never notifies.
func (uc *ScheduleResitUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in ScheduleResitInput) (*entities.StudentDebt, error) {
	if !isDebtManager(actorRole) {
		emitAudit(uc.audit, ctx, "student_debts.schedule_denied", denialFields(actorID, in.DebtID, "forbidden"))
		return nil, fmt.Errorf("%w: actor %d (role %q) cannot schedule resits",
			entities.ErrDebtAccessForbidden, actorID, actorRole)
	}

	debt, err := uc.repo.GetByID(ctx, in.DebtID)
	if err != nil {
		return nil, err
	}
	if err := debt.ScheduleResit(in.ScheduledDate, in.Examiner, uc.now()); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(ctx, debt); err != nil {
		return nil, err
	}

	notifyResitScheduled(uc.notifier, ctx, debt.StudentUserID, debt.ID, debt.DisciplineName, in.ScheduledDate)
	emitAudit(uc.audit, ctx, "student_debts.resit_scheduled",
		successFields(actorID, debt.ID, debt.Status().String()))
	return debt, nil
}
