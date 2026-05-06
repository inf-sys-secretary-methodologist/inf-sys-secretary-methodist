package entities

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func validParams(now time.Time) NewCurriculumParams {
	return NewCurriculumParams{
		Title:       "ИВТ-2026 / 4 года",
		Code:        "09.03.04-2026",
		Specialty:   "Информатика и вычислительная техника",
		Year:        2026,
		Description: "Учебный план направления подготовки 09.03.04",
		CreatedBy:   42,
		Now:         now,
	}
}

func TestNewCurriculum_HappyPath(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)
	c, err := NewCurriculum(validParams(now))
	if err != nil {
		t.Fatalf("NewCurriculum returned unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("NewCurriculum returned nil entity with no error")
	}
	if got, want := c.Title(), "ИВТ-2026 / 4 года"; got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
	if got, want := c.Code(), "09.03.04-2026"; got != want {
		t.Errorf("Code() = %q, want %q", got, want)
	}
	if got, want := c.Specialty(), "Информатика и вычислительная техника"; got != want {
		t.Errorf("Specialty() = %q, want %q", got, want)
	}
	if got, want := c.Year(), 2026; got != want {
		t.Errorf("Year() = %d, want %d", got, want)
	}
	if got, want := c.Description(), "Учебный план направления подготовки 09.03.04"; got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
	if got, want := c.CreatedBy(), int64(42); got != want {
		t.Errorf("CreatedBy() = %d, want %d", got, want)
	}
	// Default state — newly created curricula start as draft.
	if got, want := c.Status(), StatusDraft; got != want {
		t.Errorf("Status() = %q, want %q", got, want)
	}
	if c.ApprovedBy() != nil {
		t.Errorf("ApprovedBy() = %v, want nil for fresh curriculum", *c.ApprovedBy())
	}
	if c.ApprovedAt() != nil {
		t.Errorf("ApprovedAt() = %v, want nil for fresh curriculum", *c.ApprovedAt())
	}
	if !c.CreatedAt().Equal(now) {
		t.Errorf("CreatedAt() = %v, want %v", c.CreatedAt(), now)
	}
	if !c.UpdatedAt().Equal(now) {
		t.Errorf("UpdatedAt() = %v, want %v", c.UpdatedAt(), now)
	}
}

func TestNewCurriculum_TrimsAndCanonicalisesText(t *testing.T) {
	now := time.Now()
	p := validParams(now)
	p.Title = "  ИВТ-2026  "
	p.Code = "  09.03.04-2026  "
	p.Specialty = "  Информатика и вычислительная техника  "
	p.Description = "  some text  "
	c, err := NewCurriculum(p)
	if err != nil {
		t.Fatalf("NewCurriculum returned unexpected error: %v", err)
	}
	if c.Title() != "ИВТ-2026" {
		t.Errorf("Title() = %q, want trimmed canonical form", c.Title())
	}
	if c.Code() != "09.03.04-2026" {
		t.Errorf("Code() = %q, want trimmed canonical form", c.Code())
	}
	if c.Specialty() != "Информатика и вычислительная техника" {
		t.Errorf("Specialty() = %q, want trimmed canonical form", c.Specialty())
	}
	if c.Description() != "some text" {
		t.Errorf("Description() = %q, want trimmed canonical form", c.Description())
	}
}

func TestNewCurriculum_AcceptsEmptyDescription(t *testing.T) {
	now := time.Now()
	p := validParams(now)
	p.Description = ""
	c, err := NewCurriculum(p)
	if err != nil {
		t.Fatalf("NewCurriculum returned unexpected error: %v", err)
	}
	if c.Description() != "" {
		t.Errorf("Description() = %q, want empty", c.Description())
	}
}

func TestNewCurriculum_InvariantViolations(t *testing.T) {
	mutate := func(fn func(*NewCurriculumParams)) NewCurriculumParams {
		p := validParams(time.Now())
		fn(&p)
		return p
	}
	cases := []struct {
		name string
		p    NewCurriculumParams
	}{
		{"empty title", mutate(func(p *NewCurriculumParams) { p.Title = "" })},
		{"whitespace title", mutate(func(p *NewCurriculumParams) { p.Title = "   " })},
		{"empty code", mutate(func(p *NewCurriculumParams) { p.Code = "" })},
		{"whitespace code", mutate(func(p *NewCurriculumParams) { p.Code = "  \t " })},
		{"empty specialty", mutate(func(p *NewCurriculumParams) { p.Specialty = "" })},
		{"whitespace specialty", mutate(func(p *NewCurriculumParams) { p.Specialty = "   " })},
		{"year too low", mutate(func(p *NewCurriculumParams) { p.Year = 1999 })},
		{"year far in past", mutate(func(p *NewCurriculumParams) { p.Year = 0 })},
		{"year too high", mutate(func(p *NewCurriculumParams) { p.Year = 2101 })},
		{"description too long", mutate(func(p *NewCurriculumParams) {
			p.Description = strings.Repeat("ы", 4097)
		})},
		{"non-positive created_by", mutate(func(p *NewCurriculumParams) { p.CreatedBy = 0 })},
		{"negative created_by", mutate(func(p *NewCurriculumParams) { p.CreatedBy = -1 })},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewCurriculum(tc.p)
			if err == nil {
				t.Fatalf("NewCurriculum(%s) succeeded; want %v", tc.name, ErrInvalidCurriculum)
			}
			if !errors.Is(err, ErrInvalidCurriculum) {
				t.Fatalf("NewCurriculum(%s) returned %v; want errors.Is(... , ErrInvalidCurriculum)", tc.name, err)
			}
			if c != nil {
				t.Errorf("NewCurriculum(%s) returned non-nil entity on error", tc.name)
			}
		})
	}
}

func TestNewCurriculum_BoundaryValuesAreAccepted(t *testing.T) {
	mutate := func(fn func(*NewCurriculumParams)) NewCurriculumParams {
		p := validParams(time.Now())
		fn(&p)
		return p
	}
	cases := []struct {
		name string
		p    NewCurriculumParams
	}{
		{"year = 2000 lower bound", mutate(func(p *NewCurriculumParams) { p.Year = 2000 })},
		{"year = 2100 upper bound", mutate(func(p *NewCurriculumParams) { p.Year = 2100 })},
		{"description = 4096 chars exactly", mutate(func(p *NewCurriculumParams) {
			p.Description = strings.Repeat("a", 4096)
		})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewCurriculum(tc.p); err != nil {
				t.Fatalf("NewCurriculum(%s) returned error %v; boundary should be valid", tc.name, err)
			}
		})
	}
}

func TestReconstituteCurriculum_BypassesInvariants(t *testing.T) {
	// Repository implementations call ReconstituteCurriculum on rows that
	// are already canonical (they came from a row that passed the SQL
	// CHECK at write time). The factory MUST NOT re-run NewCurriculum's
	// validation — otherwise legacy rows that violate a tightened
	// invariant in a future release would fail to load.
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	approvedAt := now.Add(48 * time.Hour)
	approvedBy := int64(99)
	c := ReconstituteCurriculum(
		7, "title", "code", "specialty", 2026, "desc",
		StatusApproved, 42, &approvedBy, &approvedAt, now, now,
	)
	if c == nil {
		t.Fatal("ReconstituteCurriculum returned nil")
	}
	if c.ID != 7 {
		t.Errorf("ID = %d, want 7", c.ID)
	}
	if c.Status() != StatusApproved {
		t.Errorf("Status() = %q, want %q", c.Status(), StatusApproved)
	}
	if c.ApprovedBy() == nil || *c.ApprovedBy() != 99 {
		t.Errorf("ApprovedBy() = %v, want pointer to 99", c.ApprovedBy())
	}
	if c.ApprovedAt() == nil || !c.ApprovedAt().Equal(approvedAt) {
		t.Errorf("ApprovedAt() = %v, want %v", c.ApprovedAt(), approvedAt)
	}
}
