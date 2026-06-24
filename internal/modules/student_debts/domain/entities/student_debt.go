package entities

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Domain sentinels for the StudentDebt aggregate. Handlers map them to HTTP:
// ErrInvalidStudentDebt → 422, ErrDebtClosed/ErrInvalidTransition/
// ErrNoScheduledResit → 409.
var (
	ErrInvalidStudentDebt = errors.New("student_debts: invalid student debt")
	ErrDebtClosed         = errors.New("student_debts: debt is closed")
	ErrInvalidTransition  = errors.New("student_debts: invalid status transition")
	ErrNoScheduledResit   = errors.New("student_debts: no scheduled resit to record")
)

// StudentDebt is the aggregate root: one student owes one discipline in one
// semester. Source-denormalized identity fields are public; best-effort links
// to internal entities are nullable; lifecycle state is private and mutated
// only through the FSM methods (ScheduleResit, RecordResitResult).
type StudentDebt struct {
	ID int64

	// Denormalized from the import source (Excel/1С).
	StudentFullName string
	GroupName       string
	DisciplineName  string
	Semester        int
	ControlForm     ControlForm

	// Best-effort links, resolved when the student/discipline exist locally.
	StudentUserID *int64
	DisciplineID  *int64

	// Import provenance.
	SourceRef  string
	SourceHash string

	Version int

	status    DebtStatus
	attempts  []*ResitAttempt
	createdAt time.Time
	updatedAt time.Time
}

// NewStudentDebt creates an open debt. Required: student name, group,
// discipline name; semester ∈ [1,12]; a valid control form.
func NewStudentDebt(studentName, group, discipline string, semester int, form ControlForm) (*StudentDebt, error) {
	studentName = strings.TrimSpace(studentName)
	group = strings.TrimSpace(group)
	discipline = strings.TrimSpace(discipline)
	switch {
	case studentName == "":
		return nil, fmt.Errorf("%w: student name is required", ErrInvalidStudentDebt)
	case group == "":
		return nil, fmt.Errorf("%w: group is required", ErrInvalidStudentDebt)
	case discipline == "":
		return nil, fmt.Errorf("%w: discipline is required", ErrInvalidStudentDebt)
	case semester < 1 || semester > 12:
		return nil, fmt.Errorf("%w: semester must be in [1,12], got %d", ErrInvalidStudentDebt, semester)
	}
	if err := form.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidStudentDebt, err)
	}
	return &StudentDebt{
		StudentFullName: studentName,
		GroupName:       group,
		DisciplineName:  discipline,
		Semester:        semester,
		ControlForm:     form,
		Version:         1,
		status:          DebtStatusOpen,
	}, nil
}

// ScheduleResit appends a new resit attempt and moves the debt into
// resit_scheduled. Allowed only from open or commission state. A resit
// scheduled from commission state is itself a commission attempt.
func (d *StudentDebt) ScheduleResit(scheduledDate time.Time, examiner string, now time.Time) error {
	if d.status.IsClosed() {
		return ErrDebtClosed
	}
	if d.status != DebtStatusOpen && d.status != DebtStatusCommission {
		return fmt.Errorf("%w: cannot schedule a resit from %q", ErrInvalidTransition, d.status)
	}
	isCommission := d.status == DebtStatusCommission
	attempt, err := NewResitAttempt(len(d.attempts)+1, scheduledDate, examiner, isCommission)
	if err != nil {
		return err
	}
	d.attempts = append(d.attempts, attempt)
	d.status = DebtStatusResitScheduled
	d.updatedAt = now
	return nil
}

// RecordResitResult records the outcome of the currently scheduled resit and
// transitions the debt: passed → closed_passed; failed/no_show → closed_failed
// when it was a commission attempt, else → commission once failed regular
// attempts reach attemptsBeforeCommission, otherwise back to open.
func (d *StudentDebt) RecordResitResult(result ResitResult, grade *int, recordedBy int64, recordedAt time.Time, attemptsBeforeCommission int) error {
	if d.status != DebtStatusResitScheduled || len(d.attempts) == 0 {
		return ErrNoScheduledResit
	}
	latest := d.attempts[len(d.attempts)-1]
	wasCommission := latest.IsCommission
	if err := latest.Record(result, grade, recordedBy, recordedAt); err != nil {
		return err
	}

	if attemptsBeforeCommission < 1 {
		attemptsBeforeCommission = 1
	}
	switch {
	case result == ResitResultPassed:
		d.status = DebtStatusClosedPassed
	case wasCommission:
		d.status = DebtStatusClosedFailed
	case d.failedRegularAttempts() >= attemptsBeforeCommission:
		d.status = DebtStatusCommission
	default:
		d.status = DebtStatusOpen
	}
	d.updatedAt = recordedAt
	return nil
}

// failedRegularAttempts counts recorded failed/no_show non-commission attempts.
func (d *StudentDebt) failedRegularAttempts() int {
	n := 0
	for _, a := range d.attempts {
		if a.IsCommission {
			continue
		}
		if r := a.Result(); r == ResitResultFailed || r == ResitResultNoShow {
			n++
		}
	}
	return n
}

// Status returns the current FSM state.
func (d *StudentDebt) Status() DebtStatus { return d.status }

// Attempts returns the resit attempts in order.
func (d *StudentDebt) Attempts() []*ResitAttempt { return d.attempts }

// CreatedAt returns the creation timestamp.
func (d *StudentDebt) CreatedAt() time.Time { return d.createdAt }

// UpdatedAt returns the last-update timestamp.
func (d *StudentDebt) UpdatedAt() time.Time { return d.updatedAt }

// ReconstituteStudentDebt rebuilds an aggregate from persisted rows without
// re-validating invariants.
func ReconstituteStudentDebt(id int64, studentName, group, discipline string, semester int, form ControlForm,
	studentUserID, disciplineID *int64, sourceRef, sourceHash string, version int, status DebtStatus,
	attempts []*ResitAttempt, createdAt, updatedAt time.Time) *StudentDebt {
	return &StudentDebt{
		ID: id, StudentFullName: studentName, GroupName: group, DisciplineName: discipline,
		Semester: semester, ControlForm: form, StudentUserID: studentUserID, DisciplineID: disciplineID,
		SourceRef: sourceRef, SourceHash: sourceHash, Version: version, status: status,
		attempts: attempts, createdAt: createdAt, updatedAt: updatedAt,
	}
}
