package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
)

// RecordResitResultInput is the public DTO for recording a resit outcome.
// AttemptNo addresses the attempt being recorded (from the URL path) and
// guards against recording a stale/non-current attempt.
type RecordResitResultInput struct {
	DebtID    int64
	AttemptNo int
	Result    entities.ResitResult
	Grade     *int
}

// RecordResitResultUseCase records the outcome of the currently scheduled
// resit (EDIT_ROLES only) and lets the domain FSM advance the debt
// (passed → closed_passed; failed/no_show → open / commission /
// closed_failed depending on attempts and commission status).
//
// attemptsBeforeCommission is the institution policy N (regular failures
// before a commission resit is required) — a configuration value, not a
// domain constant (design §2), injected at construction.
type RecordResitResultUseCase struct {
	repo                     writeDebtRepo
	audit                    AuditSink
	now                      func() time.Time
	attemptsBeforeCommission int
}

// NewRecordResitResultUseCase wires the use case. repo is required; audit
// may be nil. now defaults to time.Now when nil. attemptsBeforeCommission
// is clamped to a minimum of 1 (mirrors the domain's own clamp).
func NewRecordResitResultUseCase(repo writeDebtRepo, audit AuditSink, now func() time.Time, attemptsBeforeCommission int) *RecordResitResultUseCase {
	if repo == nil {
		panic("student_debts: NewRecordResitResultUseCase requires non-nil repo")
	}
	if now == nil {
		now = time.Now
	}
	if attemptsBeforeCommission < 1 {
		attemptsBeforeCommission = 1
	}
	return &RecordResitResultUseCase{repo: repo, audit: audit, now: now, attemptsBeforeCommission: attemptsBeforeCommission}
}

// Execute records the resit outcome and returns the updated aggregate:
// EDIT_ROLES gate → load → attempt-no guard → domain FSM
// (RecordResitResult, which advances the status) → persist → audit. The
// attempt-no guard rejects recording anything but the current (latest)
// attempt, so a stale URL (…/attempts/1/result after attempt 2 was
// scheduled) cannot overwrite history; it maps to ErrNoScheduledResit.
func (uc *RecordResitResultUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in RecordResitResultInput) (*entities.StudentDebt, error) {
	if !isDebtManager(actorRole) {
		emitAudit(uc.audit, ctx, "student_debts.record_denied", denialFields(actorID, in.DebtID, "forbidden"))
		return nil, fmt.Errorf("%w: actor %d (role %q) cannot record resit results",
			entities.ErrDebtAccessForbidden, actorID, actorRole)
	}

	debt, err := uc.repo.GetByID(ctx, in.DebtID)
	if err != nil {
		return nil, err
	}

	attempts := debt.Attempts()
	if len(attempts) == 0 || attempts[len(attempts)-1].AttemptNo != in.AttemptNo {
		return nil, fmt.Errorf("%w: attempt %d is not the current scheduled attempt",
			entities.ErrNoScheduledResit, in.AttemptNo)
	}

	if err := debt.RecordResitResult(in.Result, in.Grade, actorID, uc.now(), uc.attemptsBeforeCommission); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(ctx, debt); err != nil {
		return nil, err
	}

	emitAudit(uc.audit, ctx, "student_debts.resit_recorded",
		successFields(actorID, debt.ID, debt.Status().String()))
	return debt, nil
}
