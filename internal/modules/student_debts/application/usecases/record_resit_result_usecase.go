package usecases

import (
	"context"
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

// Execute records the resit outcome and returns the updated aggregate.
func (uc *RecordResitResultUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in RecordResitResultInput) (*entities.StudentDebt, error) {
	return nil, errNotImplemented
}
