package entities

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func freshDraftCurriculum(t *testing.T) *Curriculum {
	t.Helper()
	c, err := NewCurriculum(NewCurriculumParams{
		Title:       "Original",
		Code:        "ORIG-2026",
		Specialty:   "Original Specialty",
		Year:        2026,
		Description: "original description",
		CreatedBy:   42,
		Now:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewCurriculum failed: %v", err)
	}
	return c
}

func TestUpdateBasics_HappyPath(t *testing.T) {
	c := freshDraftCurriculum(t)
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)

	err := c.UpdateBasics("New Title", "NEW-2026", "New Specialty", 2027,
		"new description", now)
	if err != nil {
		t.Fatalf("UpdateBasics returned unexpected error: %v", err)
	}
	if got, want := c.Title(), "New Title"; got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
	if got, want := c.Code(), "NEW-2026"; got != want {
		t.Errorf("Code() = %q, want %q", got, want)
	}
	if got, want := c.Specialty(), "New Specialty"; got != want {
		t.Errorf("Specialty() = %q, want %q", got, want)
	}
	if got, want := c.Year(), 2027; got != want {
		t.Errorf("Year() = %d, want %d", got, want)
	}
	if got, want := c.Description(), "new description"; got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
	if !c.UpdatedAt().Equal(now) {
		t.Errorf("UpdatedAt() = %v, want %v", c.UpdatedAt(), now)
	}
}

func TestUpdateBasics_TrimsCanonicalForm(t *testing.T) {
	c := freshDraftCurriculum(t)
	now := time.Now()
	if err := c.UpdateBasics("  Padded Title  ", " PAD-2026 ",
		"  Padded Specialty ", 2026, "  padded desc  ", now); err != nil {
		t.Fatalf("UpdateBasics returned unexpected error: %v", err)
	}
	if c.Title() != "Padded Title" {
		t.Errorf("Title() = %q, want trimmed", c.Title())
	}
	if c.Code() != "PAD-2026" {
		t.Errorf("Code() = %q, want trimmed", c.Code())
	}
	if c.Specialty() != "Padded Specialty" {
		t.Errorf("Specialty() = %q, want trimmed", c.Specialty())
	}
	if c.Description() != "padded desc" {
		t.Errorf("Description() = %q, want trimmed", c.Description())
	}
}

func TestUpdateBasics_RejectsNonDraftStatus(t *testing.T) {
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
			titleBefore := c.Title()
			updatedAtBefore := c.UpdatedAt()

			err := c.UpdateBasics("New Title", "NEW", "New Specialty", 2027,
				"new desc", time.Now())
			if !errors.Is(err, ErrCannotEditApproved) {
				t.Fatalf("UpdateBasics on %s = %v; want errors.Is(... , ErrCannotEditApproved)",
					tc.name, err)
			}
			// No mutations applied — defense in depth: a future caller that
			// ignores the returned error must still see the original state.
			if c.Title() != titleBefore {
				t.Errorf("Title mutated despite error: got %q, want %q",
					c.Title(), titleBefore)
			}
			if !c.UpdatedAt().Equal(updatedAtBefore) {
				t.Errorf("UpdatedAt mutated despite error: got %v, want %v",
					c.UpdatedAt(), updatedAtBefore)
			}
		})
	}
}

func TestUpdateBasics_RejectsInvariantViolations(t *testing.T) {
	cases := []struct {
		name string
		// Each case mutates exactly one field of a valid arg-set.
		title       string
		code        string
		specialty   string
		year        int
		description string
	}{
		{"empty title", "", "OK-2026", "Spec", 2026, "desc"},
		{"whitespace title", "   ", "OK-2026", "Spec", 2026, "desc"},
		{"empty code", "Title", "", "Spec", 2026, "desc"},
		{"empty specialty", "Title", "OK-2026", "", 2026, "desc"},
		{"year too low", "Title", "OK-2026", "Spec", 1999, "desc"},
		{"year too high", "Title", "OK-2026", "Spec", 2101, "desc"},
		{"description too long", "Title", "OK-2026", "Spec", 2026,
			strings.Repeat("ы", 4097)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := freshDraftCurriculum(t)
			titleBefore := c.Title()

			err := c.UpdateBasics(tc.title, tc.code, tc.specialty, tc.year,
				tc.description, time.Now())
			if !errors.Is(err, ErrInvalidCurriculum) {
				t.Fatalf("UpdateBasics(%s) = %v; want errors.Is(... , ErrInvalidCurriculum)",
					tc.name, err)
			}
			// Atomicity: failed validation must leave the entity untouched.
			if c.Title() != titleBefore {
				t.Errorf("Title mutated despite invariant failure: %q", c.Title())
			}
		})
	}
}
