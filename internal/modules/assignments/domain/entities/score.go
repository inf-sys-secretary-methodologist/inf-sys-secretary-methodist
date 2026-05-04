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

// Score is a value object representing a graded score together with the
// assignment's max possible score. Both are required so that downstream
// consumers (UI, analytics) can render the score in context (e.g. "85/100")
// without re-fetching the assignment.
type Score struct {
	value int
	max   int
}

// NewScore returns a Score after enforcing invariants. Returns
// ErrInvalidScore (wrapped with context) on violation.
//
// Invariants:
//   - max > 0  (an assignment with non-positive max would be meaningless)
//   - value >= 0
//   - value <= max
func NewScore(value, max int) (Score, error) {
	if max <= 0 {
		return Score{}, fmt.Errorf("%w: max must be positive, got %d", ErrInvalidScore, max)
	}
	if value < 0 {
		return Score{}, fmt.Errorf("%w: value cannot be negative, got %d", ErrInvalidScore, value)
	}
	if value > max {
		return Score{}, fmt.Errorf("%w: value %d exceeds max %d", ErrInvalidScore, value, max)
	}
	return Score{value: value, max: max}, nil
}

// Value returns the awarded score.
func (s Score) Value() int { return s.value }

// Max returns the assignment's maximum possible score.
func (s Score) Max() int { return s.max }
