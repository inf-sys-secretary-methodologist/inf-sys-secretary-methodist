package entities

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
)

const maxGoalTextLen = 2048

// Goal — цель/задача освоения дисциплины. Inner aggregate of
// WorkProgram (ADR-1); fields stay unexported so collection
// invariants enforced via WorkProgram.AddGoal/RemoveGoal land in PR
// 1c. This PR ships only the value-object construction surface.
type Goal struct {
	id            int64
	workProgramID int64
	text          string
	orderIndex    int
	createdAt     time.Time
}

// NewGoal constructs a fresh Goal. Text is trimmed; empty/whitespace
// after trim and overlong text fail with ErrInvalidWorkProgram naming
// the offending field. order_index ≥ 0 enforced.
func NewGoal(text string, orderIndex int) (*Goal, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, fmt.Errorf("%w: text is required", domain.ErrInvalidWorkProgram)
	}
	if utf8.RuneCountInString(trimmed) > maxGoalTextLen {
		return nil, fmt.Errorf("%w: text must be <= %d runes", domain.ErrInvalidWorkProgram, maxGoalTextLen)
	}
	if orderIndex < 0 {
		return nil, fmt.Errorf("%w: order_index must be non-negative", domain.ErrInvalidWorkProgram)
	}
	return &Goal{
		text:       trimmed,
		orderIndex: orderIndex,
		createdAt:  time.Now().UTC(),
	}, nil
}

// ID returns the persistent identifier (0 for unsaved goals).
func (g *Goal) ID() int64 { return g.id }

// WorkProgramID returns the parent aggregate identifier (0 until
// attached via WorkProgram.AddGoal).
func (g *Goal) WorkProgramID() int64 { return g.workProgramID }

// Text returns the goal/task description (trimmed, ≤ 2048 runes).
func (g *Goal) Text() string { return g.text }

// OrderIndex returns the display ordering hint (≥ 0).
func (g *Goal) OrderIndex() int { return g.orderIndex }

// CreatedAt returns the creation timestamp.
func (g *Goal) CreatedAt() time.Time { return g.createdAt }
