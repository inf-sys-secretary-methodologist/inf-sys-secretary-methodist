package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
)

// errNotImplemented marks the RED-state stubs in this slice. The GREEN
// commit replaces every stubbed Execute with the real orchestration.
var errNotImplemented = errors.New("student_debts: not implemented")

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

// Execute schedules a resit and returns the updated aggregate.
func (uc *ScheduleResitUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in ScheduleResitInput) (*entities.StudentDebt, error) {
	return nil, errNotImplemented
}
