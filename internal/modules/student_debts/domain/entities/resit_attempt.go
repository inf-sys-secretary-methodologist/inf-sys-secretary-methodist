package entities

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrInvalidResitAttempt signals violation of a ResitAttempt construction
// invariant (non-positive attempt number, empty examiner, zero date).
// Handlers map this sentinel to HTTP 422.
var ErrInvalidResitAttempt = errors.New("student_debts: invalid resit attempt")

// ErrAttemptAlreadyRecorded indicates Record was called on an attempt whose
// result is already final. Re-recording must not silently overwrite history.
// Handlers map this sentinel to HTTP 409.
var ErrAttemptAlreadyRecorded = errors.New("student_debts: resit attempt already recorded")

// ErrInvalidResitRecord signals an invalid Record call — a non-final or
// unknown result, or a non-positive recorder. Handlers map this to HTTP 422.
var ErrInvalidResitRecord = errors.New("student_debts: invalid resit record")

// ResitAttempt is a child entity of StudentDebt: one scheduled (and later
// recorded) attempt to liquidate the debt. Identity fields are public;
// mutable outcome state is private and guarded by Record.
type ResitAttempt struct {
	ID           int64
	DebtID       int64
	AttemptNo    int
	IsCommission bool

	scheduledDate time.Time
	examiner      string
	result        ResitResult
	grade         *int
	recordedBy    *int64
	recordedAt    *time.Time
}

// NewResitAttempt creates a pending attempt. attemptNo must be positive,
// examiner non-empty, scheduledDate non-zero.
func NewResitAttempt(attemptNo int, scheduledDate time.Time, examiner string, isCommission bool) (*ResitAttempt, error) {
	if attemptNo <= 0 {
		return nil, fmt.Errorf("%w: attempt number must be positive, got %d", ErrInvalidResitAttempt, attemptNo)
	}
	examiner = strings.TrimSpace(examiner)
	if examiner == "" {
		return nil, fmt.Errorf("%w: examiner is required", ErrInvalidResitAttempt)
	}
	if scheduledDate.IsZero() {
		return nil, fmt.Errorf("%w: scheduled date is required", ErrInvalidResitAttempt)
	}
	return &ResitAttempt{
		AttemptNo:     attemptNo,
		IsCommission:  isCommission,
		scheduledDate: scheduledDate,
		examiner:      examiner,
		result:        ResitResultPending,
	}, nil
}

// Record sets the final outcome of the attempt. The result must be a valid,
// final outcome (not pending); recordedBy must be positive; an attempt whose
// outcome is already final cannot be re-recorded.
func (a *ResitAttempt) Record(result ResitResult, grade *int, recordedBy int64, recordedAt time.Time) error {
	if a.result.IsFinal() {
		return ErrAttemptAlreadyRecorded
	}
	if !result.IsValid() || !result.IsFinal() {
		return fmt.Errorf("%w: result %q is not a final outcome", ErrInvalidResitRecord, result)
	}
	if recordedBy <= 0 {
		return fmt.Errorf("%w: recordedBy must be positive", ErrInvalidResitRecord)
	}
	a.result = result
	a.grade = grade
	a.recordedBy = &recordedBy
	a.recordedAt = &recordedAt
	return nil
}

// ScheduledDate returns the planned resit date.
func (a *ResitAttempt) ScheduledDate() time.Time { return a.scheduledDate }

// Examiner returns the examiner (or commission label) for the attempt.
func (a *ResitAttempt) Examiner() string { return a.examiner }

// Result returns the recorded outcome (pending until Record is called).
func (a *ResitAttempt) Result() ResitResult { return a.result }

// Grade returns the recorded grade, or nil if absent.
func (a *ResitAttempt) Grade() *int { return a.grade }

// RecordedBy returns the user id that recorded the outcome, or nil.
func (a *ResitAttempt) RecordedBy() *int64 { return a.recordedBy }

// RecordedAt returns when the outcome was recorded, or nil.
func (a *ResitAttempt) RecordedAt() *time.Time { return a.recordedAt }

// ReconstituteResitAttempt rebuilds an attempt from a persisted row without
// re-validating invariants (the DB CHECK constraints are the source of truth).
func ReconstituteResitAttempt(id, debtID int64, attemptNo int, isCommission bool, scheduledDate time.Time,
	examiner string, result ResitResult, grade *int, recordedBy *int64, recordedAt *time.Time) *ResitAttempt {
	return &ResitAttempt{
		ID: id, DebtID: debtID, AttemptNo: attemptNo, IsCommission: isCommission,
		scheduledDate: scheduledDate, examiner: examiner, result: result,
		grade: grade, recordedBy: recordedBy, recordedAt: recordedAt,
	}
}
