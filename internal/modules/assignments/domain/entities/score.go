// Package entities provides domain entities and value objects for the
// assignments module — the academic Tasks Context (separate from the
// project-management tasks module).
package entities

import (
	"errors"
	"fmt"
)

// ErrInvalidScore signals that a score value violates one of the Score
// invariants: max must be positive, value must be in [0, max]. Handlers
// map this sentinel to HTTP 422.
var ErrInvalidScore = errors.New("assignments: invalid score")

// Score is a value object representing a graded score. The non-negative
// invariant is enforced here; the upper bound (value ≤ assignment max)
// is owned by Assignment.NewSubmissionScore because it is a
// cross-aggregate rule that the Score VO alone cannot evaluate. Storing
// only value (and not max) follows the v0.109.0 reviewer's note that
// the previous max field was dead data — it was never read by callers.
type Score struct {
	value int
}

// NewScore returns a Score after enforcing the value invariant. The
// assignment-relative upper bound is enforced by
// Assignment.NewSubmissionScore.
//
// Returns ErrInvalidScore (wrapped) on violation.
func NewScore(value int) (Score, error) {
	if value < 0 {
		return Score{}, fmt.Errorf("%w: value cannot be negative, got %d", ErrInvalidScore, value)
	}
	return Score{value: value}, nil
}

// Value returns the awarded score.
func (s Score) Value() int { return s.value }
