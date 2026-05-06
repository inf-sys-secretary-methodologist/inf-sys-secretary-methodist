package entities

import (
	"errors"
	"testing"
	"time"
)

// buildCurriculum constructs a minimal Curriculum via Reconstitute so tests
// can exercise the AuthorizeEdit predicate against arbitrary statuses
// without going through SubmitForApproval / Approve transitions
// (which land in v0.117.0).
func buildCurriculum(t *testing.T, createdBy int64, status CurriculumStatus) *Curriculum {
	t.Helper()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var approvedBy *int64
	var approvedAt *time.Time
	if status == StatusApproved {
		ab := int64(99)
		at := now.Add(48 * time.Hour)
		approvedBy = &ab
		approvedAt = &at
	}
	return ReconstituteCurriculum(
		1, "title", "code", "specialty", 2026, "desc",
		status, createdBy, approvedBy, approvedAt, now, now,
	)
}

func TestCurriculum_AuthorizeEdit(t *testing.T) {
	const author = int64(42)
	const stranger = int64(7)
	cases := []struct {
		name      string
		status    CurriculumStatus
		actorID   int64
		isAdmin   bool
		wantSent  error // nil = OK
	}{
		// Author edits own draft — happy path.
		{"author edits own draft", StatusDraft, author, false, nil},
		// Stranger methodist may not edit a foreign draft.
		{"stranger methodist on foreign draft", StatusDraft, stranger, false, ErrCurriculumScopeForbidden},
		// Admin overrides the ownership check.
		{"admin on foreign draft", StatusDraft, stranger, true, nil},
		// Status gate: only draft is editable. The status check fires
		// BEFORE ownership / admin overrides — approved curricula are
		// frozen for everyone (Approve → UpdateBasics is not allowed;
		// the proper path in v0.117.0 will be Reject → Edit → Resubmit).
		{"author on own pending_approval", StatusPendingApproval, author, false, ErrCannotEditApproved},
		{"admin on pending_approval", StatusPendingApproval, stranger, true, ErrCannotEditApproved},
		{"author on own approved", StatusApproved, author, false, ErrCannotEditApproved},
		{"admin on approved", StatusApproved, stranger, true, ErrCannotEditApproved},
		{"author on own archived", StatusArchived, author, false, ErrCannotEditApproved},
		{"admin on archived", StatusArchived, stranger, true, ErrCannotEditApproved},
		// Defensive: zero / negative actor with no admin flag is treated
		// as "no actor" — must not equal author by accident.
		{"zero actor not admin", StatusDraft, 0, false, ErrCurriculumScopeForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := buildCurriculum(t, author, tc.status)
			err := c.AuthorizeEdit(tc.actorID, tc.isAdmin)
			if tc.wantSent == nil {
				if err != nil {
					t.Fatalf("AuthorizeEdit(%d, %v) = %v; want nil", tc.actorID, tc.isAdmin, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("AuthorizeEdit(%d, %v) = nil; want %v", tc.actorID, tc.isAdmin, tc.wantSent)
			}
			if !errors.Is(err, tc.wantSent) {
				t.Fatalf("AuthorizeEdit(%d, %v) = %v; want errors.Is(... , %v)",
					tc.actorID, tc.isAdmin, err, tc.wantSent)
			}
		})
	}
}
