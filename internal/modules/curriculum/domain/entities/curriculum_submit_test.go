package entities

import (
	"errors"
	"testing"
	"time"
)

func TestCurriculum_SubmitForApproval(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)

	t.Run("draft transitions to pending_approval and bumps updatedAt", func(t *testing.T) {
		c := buildCurriculum(t, 42, StatusDraft)
		updatedBefore := c.UpdatedAt()
		err := c.SubmitForApproval(now)
		if err != nil {
			t.Fatalf("SubmitForApproval returned error: %v", err)
		}
		if got, want := c.Status(), StatusPendingApproval; got != want {
			t.Errorf("Status() = %q, want %q", got, want)
		}
		if !c.UpdatedAt().Equal(now) {
			t.Errorf("UpdatedAt() = %v, want %v", c.UpdatedAt(), now)
		}
		if c.UpdatedAt().Equal(updatedBefore) {
			t.Errorf("UpdatedAt was not bumped (still %v)", updatedBefore)
		}
		// Approval audit fields stay untouched on Submit — only Approve sets them.
		if c.ApprovedBy() != nil {
			t.Errorf("ApprovedBy must remain nil after Submit, got %v", *c.ApprovedBy())
		}
		if c.ApprovedAt() != nil {
			t.Errorf("ApprovedAt must remain nil after Submit, got %v", *c.ApprovedAt())
		}
	})

	t.Run("non-draft statuses reject with ErrCannotSubmit and leave entity untouched", func(t *testing.T) {
		cases := []struct {
			name   string
			status CurriculumStatus
		}{
			{"pending_approval", StatusPendingApproval},
			{"approved", StatusApproved},
			{"archived", StatusArchived},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				c := buildCurriculum(t, 42, tc.status)
				statusBefore := c.Status()
				updatedBefore := c.UpdatedAt()

				err := c.SubmitForApproval(now)
				if !errors.Is(err, ErrCannotSubmit) {
					t.Fatalf("SubmitForApproval(%s) = %v; want errors.Is(... , ErrCannotSubmit)",
						tc.name, err)
				}
				// Atomicity — failed transition must leave the entity untouched.
				if c.Status() != statusBefore {
					t.Errorf("Status mutated despite error: got %q, want %q",
						c.Status(), statusBefore)
				}
				if !c.UpdatedAt().Equal(updatedBefore) {
					t.Errorf("UpdatedAt mutated despite error: got %v, want %v",
						c.UpdatedAt(), updatedBefore)
				}
			})
		}
	})
}
