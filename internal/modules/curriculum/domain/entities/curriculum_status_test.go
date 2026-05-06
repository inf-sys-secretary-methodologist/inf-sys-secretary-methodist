package entities

import "testing"

func TestCurriculumStatus_IsValid(t *testing.T) {
	cases := []struct {
		name string
		s    CurriculumStatus
		want bool
	}{
		{"draft is valid", StatusDraft, true},
		{"pending_approval is valid", StatusPendingApproval, true},
		{"approved is valid", StatusApproved, true},
		{"archived is valid", StatusArchived, true},
		{"empty is invalid", CurriculumStatus(""), false},
		{"unknown literal is invalid", CurriculumStatus("rejected"), false},
		{"capitalised is invalid", CurriculumStatus("Draft"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.s.IsValid(); got != tc.want {
				t.Fatalf("(%q).IsValid() = %v, want %v", string(tc.s), got, tc.want)
			}
		})
	}
}

func TestCurriculumStatus_CanEdit(t *testing.T) {
	cases := []struct {
		name string
		s    CurriculumStatus
		want bool
	}{
		{"draft can be edited", StatusDraft, true},
		{"pending_approval cannot be edited", StatusPendingApproval, false},
		{"approved cannot be edited", StatusApproved, false},
		{"archived cannot be edited", StatusArchived, false},
		{"unknown cannot be edited", CurriculumStatus("rejected"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.s.CanEdit(); got != tc.want {
				t.Fatalf("(%q).CanEdit() = %v, want %v", string(tc.s), got, tc.want)
			}
		})
	}
}

func TestCurriculumStatus_IsApproved(t *testing.T) {
	cases := []struct {
		name string
		s    CurriculumStatus
		want bool
	}{
		{"draft not approved", StatusDraft, false},
		{"pending_approval not approved", StatusPendingApproval, false},
		{"approved is approved", StatusApproved, true},
		{"archived not approved", StatusArchived, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.s.IsApproved(); got != tc.want {
				t.Fatalf("(%q).IsApproved() = %v, want %v", string(tc.s), got, tc.want)
			}
		})
	}
}

func TestCurriculumStatus_StringMatchesDBLiteral(t *testing.T) {
	// The DB CHECK constraint chk_curricula_status_enum pins these literals.
	// If a domain literal drifts, Reconstitute would silently produce an
	// invalid status (because the DB read still returns the old string,
	// but CanEdit / IsApproved would no longer match).
	cases := map[CurriculumStatus]string{
		StatusDraft:           "draft",
		StatusPendingApproval: "pending_approval",
		StatusApproved:        "approved",
		StatusArchived:        "archived",
	}
	for s, want := range cases {
		if string(s) != want {
			t.Fatalf("status literal drift: %s = %q, want %q", string(s), string(s), want)
		}
	}
}
