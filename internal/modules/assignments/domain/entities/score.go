// Package entities provides domain entities and value objects for the
// assignments module — the academic Tasks Context (separate from the
// project-management tasks module).
package entities

import "errors"

// ErrInvalidScore signals that a score value violates one of the Score
// invariants: max must be positive, value must be in [0, max]. Handlers
// map this sentinel to HTTP 422.
var ErrInvalidScore = errors.New("assignments: invalid score")

// Score is a value object representing a graded score together with the
// assignment's max possible score. Both are required so that downstream
// consumers (UI, analytics) can render the score in context (e.g. "85/100")
// without re-fetching the assignment.
type Score struct {
	value int
	max   int
}

// NewScore returns a Score after enforcing invariants. Returns
// ErrInvalidScore on violation.
func NewScore(value, max int) (Score, error) {
	// stub for RED — invariants are not enforced yet
	return Score{value: value, max: max}, nil
}

// Value returns the awarded score.
func (s Score) Value() int { return s.value }

// Max returns the assignment's maximum possible score.
func (s Score) Max() int { return s.max }
