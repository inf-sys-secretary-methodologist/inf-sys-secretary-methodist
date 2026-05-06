package entities

import (
	"errors"
	"testing"
	"time"
)

func TestCurriculum_Approve(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)
	const adminID = int64(99)

	t.Run("pending_approval transitions to approved + records admin/timestamp", func(t *testing.T) {
		c := buildCurriculum(t, 42, StatusPendingApproval)
		err := c.Approve(adminID, now)
		if err != nil {
			t.Fatalf("Approve returned error: %v", err)
		}
		if got, want := c.Status(), StatusApproved; got != want {
			t.Errorf("Status() = %q, want %q", got, want)
		}
		if c.ApprovedBy() == nil || *c.ApprovedBy() != adminID {
			t.Errorf("ApprovedBy() = %v, want pointer to %d", c.ApprovedBy(), adminID)
		}
		if c.ApprovedAt() == nil || !c.ApprovedAt().Equal(now) {
			t.Errorf("ApprovedAt() = %v, want %v", c.ApprovedAt(), now)
		}
		if !c.UpdatedAt().Equal(now) {
			t.Errorf("UpdatedAt() = %v, want %v", c.UpdatedAt(), now)
		}
	})

	t.Run("non-pending statuses reject with ErrCannotApprove and leave entity untouched", func(t *testing.T) {
		cases := []struct {
			name   string
			status CurriculumStatus
		}{
			{"draft", StatusDraft},
			{"approved (already)", StatusApproved},
			{"archived", StatusArchived},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				c := buildCurriculum(t, 42, tc.status)
				statusBefore := c.Status()
				updatedBefore := c.UpdatedAt()

				err := c.Approve(adminID, now)
				if !errors.Is(err, ErrCannotApprove) {
					t.Fatalf("Approve(%s) = %v; want errors.Is(... , ErrCannotApprove)",
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
				// approvedBy / approvedAt nil for non-approved sources;
				// for the "already approved" case they retain their
				// pre-existing values.
				if tc.status != StatusApproved {
					if c.ApprovedBy() != nil {
						t.Errorf("ApprovedBy mutated despite error: got %v",
							*c.ApprovedBy())
					}
					if c.ApprovedAt() != nil {
						t.Errorf("ApprovedAt mutated despite error: got %v",
							*c.ApprovedAt())
					}
				}
			})
		}
	})

	t.Run("non-positive admin id rejected", func(t *testing.T) {
		c := buildCurriculum(t, 42, StatusPendingApproval)
		err := c.Approve(0, now)
		if !errors.Is(err, ErrCannotApprove) {
			t.Fatalf("Approve(0, now) = %v; want errors.Is(... , ErrCannotApprove)", err)
		}
		// Defense in depth: a missing admin id (silent admin scenario)
		// must not transition the curriculum.
		if c.Status() != StatusPendingApproval {
			t.Errorf("Status mutated despite invalid admin id: got %q",
				c.Status())
		}
	})
}
