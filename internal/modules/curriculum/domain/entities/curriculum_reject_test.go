package entities

import (
	"errors"
	"testing"
	"time"
)

func TestCurriculum_Reject(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)

	t.Run("pending_approval transitions back to draft, updatedAt bumped", func(t *testing.T) {
		c := buildCurriculum(t, 42, StatusPendingApproval)
		err := c.Reject(now)
		if err != nil {
			t.Fatalf("Reject returned error: %v", err)
		}
		if got, want := c.Status(), StatusDraft; got != want {
			t.Errorf("Status() = %q, want %q", got, want)
		}
		if !c.UpdatedAt().Equal(now) {
			t.Errorf("UpdatedAt() = %v, want %v", c.UpdatedAt(), now)
		}
		// Reject does not mutate approval audit fields — they were
		// nil in pending_approval and stay nil after rejection.
		if c.ApprovedBy() != nil {
			t.Errorf("ApprovedBy must remain nil after Reject, got %v", *c.ApprovedBy())
		}
		if c.ApprovedAt() != nil {
			t.Errorf("ApprovedAt must remain nil after Reject, got %v", *c.ApprovedAt())
		}
	})

	t.Run("non-pending statuses reject with ErrCannotReject and leave entity untouched", func(t *testing.T) {
		cases := []struct {
			name   string
			status CurriculumStatus
		}{
			{"draft", StatusDraft},
			{"approved", StatusApproved},
			{"archived", StatusArchived},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				c := buildCurriculum(t, 42, tc.status)
				statusBefore := c.Status()
				updatedBefore := c.UpdatedAt()

				err := c.Reject(now)
				if !errors.Is(err, ErrCannotReject) {
					t.Fatalf("Reject(%s) = %v; want errors.Is(... , ErrCannotReject)",
						tc.name, err)
				}
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
