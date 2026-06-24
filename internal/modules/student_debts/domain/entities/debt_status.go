package entities

import (
	"errors"
	"fmt"
)

// ErrInvalidDebtStatus signals an unrecognized DebtStatus value.
var ErrInvalidDebtStatus = errors.New("debt_status: unknown status")

// DebtStatus is the typed FSM state of an academic debt. String literals
// match the SQL CHECK on student_debts.status (migration 050) byte-for-byte
// so domain values round-trip without translation.
//
//	open            — долг зафиксирован, пересдача не назначена
//	resit_scheduled — назначена пересдача (обычная или комиссионная)
//	commission      — исчерпаны обычные попытки, требуется комиссия
//	closed_passed   — ликвидирован (сдал)            [терминальный]
//	closed_failed   — не ликвидирован (провал комиссии) [терминальный]
type DebtStatus string

// Recognized debt statuses.
const (
	DebtStatusOpen           DebtStatus = "open"
	DebtStatusResitScheduled DebtStatus = "resit_scheduled"
	DebtStatusCommission     DebtStatus = "commission"
	DebtStatusClosedPassed   DebtStatus = "closed_passed"
	DebtStatusClosedFailed   DebtStatus = "closed_failed"
)

// IsValid reports whether s is one of the recognized statuses.
func (s DebtStatus) IsValid() bool {
	switch s {
	case DebtStatusOpen, DebtStatusResitScheduled, DebtStatusCommission,
		DebtStatusClosedPassed, DebtStatusClosedFailed:
		return true
	default:
		return false
	}
}

// IsClosed reports whether s is a terminal status (no further transitions).
func (s DebtStatus) IsClosed() bool {
	return s == DebtStatusClosedPassed || s == DebtStatusClosedFailed
}

// Validate returns nil for a recognized status, else wraps ErrInvalidDebtStatus.
func (s DebtStatus) Validate() error {
	if !s.IsValid() {
		return fmt.Errorf("%w: %q", ErrInvalidDebtStatus, string(s))
	}
	return nil
}

// String returns the canonical wire representation.
func (s DebtStatus) String() string { return string(s) }
