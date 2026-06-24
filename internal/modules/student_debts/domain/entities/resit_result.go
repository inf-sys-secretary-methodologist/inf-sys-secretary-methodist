package entities

import (
	"errors"
	"fmt"
)

// ErrInvalidResitResult signals an unrecognized ResitResult value.
var ErrInvalidResitResult = errors.New("resit_result: unknown result")

// ResitResult is the typed outcome of a resit attempt. String literals match
// the SQL CHECK on debt_resit_attempts.result (migration 050).
//
//	pending — пересдача назначена, результат не внесён
//	passed  — сдал
//	failed  — не сдал
//	no_show — не явился
type ResitResult string

// Recognized resit results.
const (
	ResitResultPending ResitResult = "pending"
	ResitResultPassed  ResitResult = "passed"
	ResitResultFailed  ResitResult = "failed"
	ResitResultNoShow  ResitResult = "no_show"
)

// IsValid reports whether r is one of the recognized results.
func (r ResitResult) IsValid() bool { return false } // RED stub

// IsFinal reports whether r is a recorded outcome (anything but pending).
func (r ResitResult) IsFinal() bool { return false } // RED stub

// Validate returns nil for a recognized result, else wraps ErrInvalidResitResult.
func (r ResitResult) Validate() error {
	if !r.IsValid() {
		return fmt.Errorf("%w: %q", ErrInvalidResitResult, string(r))
	}
	return nil
}

// String returns the canonical wire representation.
func (r ResitResult) String() string { return string(r) }
