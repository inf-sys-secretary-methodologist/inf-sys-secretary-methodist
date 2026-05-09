package entities

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// validSectionParams returns a syntactically valid NewSectionParams
// instance for use as a baseline in invariant-violation tests. Tests
// mutate one field at a time so each assertion isolates a single
// invariant (mirror к validParams helper в curriculum_test.go).
func validSectionParams(now time.Time) NewSectionParams {
	return NewSectionParams{
		CurriculumID: 7,
		Title:        "Базовая часть",
		Description:  "Дисциплины обязательной части программы",
		OrderIndex:   0,
		Now:          now,
	}
}

func TestNewSection_HappyPath(t *testing.T) {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	s, err := NewSection(validSectionParams(now))
	if err != nil {
		t.Fatalf("NewSection returned unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("NewSection returned nil entity with no error")
	}
	if got, want := s.CurriculumID(), int64(7); got != want {
		t.Errorf("CurriculumID() = %d, want %d", got, want)
	}
	if got, want := s.Title(), "Базовая часть"; got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
	if got, want := s.Description(), "Дисциплины обязательной части программы"; got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
	if got, want := s.OrderIndex(), 0; got != want {
		t.Errorf("OrderIndex() = %d, want %d", got, want)
	}
	// Fresh sections start at version 0; optimistic-lock baseline (ADR-3).
	if got, want := s.Version(), 0; got != want {
		t.Errorf("Version() = %d, want %d for fresh section", got, want)
	}
	if !s.CreatedAt().Equal(now) {
		t.Errorf("CreatedAt() = %v, want %v", s.CreatedAt(), now)
	}
	if !s.UpdatedAt().Equal(now) {
		t.Errorf("UpdatedAt() = %v, want %v", s.UpdatedAt(), now)
	}
}

func TestNewSection_TrimsTextFields(t *testing.T) {
	now := time.Now()
	p := validSectionParams(now)
	p.Title = "  Базовая часть  "
	p.Description = "  Дисциплины обязательной части программы  "
	s, err := NewSection(p)
	if err != nil {
		t.Fatalf("NewSection returned unexpected error: %v", err)
	}
	if got, want := s.Title(), "Базовая часть"; got != want {
		t.Errorf("Title() = %q, want trimmed canonical form %q", got, want)
	}
	if got, want := s.Description(), "Дисциплины обязательной части программы"; got != want {
		t.Errorf("Description() = %q, want trimmed canonical form %q", got, want)
	}
}

func TestNewSection_AcceptsBlankDescription(t *testing.T) {
	// Description is optional — blank input must not raise an
	// invariant error (mirror к curriculum.description which is
	// also optional). Title remains the only required text field.
	now := time.Now()
	p := validSectionParams(now)
	p.Description = ""
	s, err := NewSection(p)
	if err != nil {
		t.Fatalf("NewSection returned unexpected error for blank description: %v", err)
	}
	if s.Description() != "" {
		t.Errorf("Description() = %q, want empty", s.Description())
	}
}

// TestNewSection_InvariantViolations is table-driven per CLAUDE.md
// ≥3-variant gate — each row pins one rejection path. Each rejected
// input wraps ErrInvalidSection so handlers can errors.Is the sentinel
// for the 422 mapping.
func TestNewSection_InvariantViolations(t *testing.T) {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name       string
		mutate     func(*NewSectionParams)
		wantSubstr string
	}{
		{
			name:       "empty curriculum id",
			mutate:     func(p *NewSectionParams) { p.CurriculumID = 0 },
			wantSubstr: "curriculum_id",
		},
		{
			name:       "negative curriculum id",
			mutate:     func(p *NewSectionParams) { p.CurriculumID = -3 },
			wantSubstr: "curriculum_id",
		},
		{
			name:       "empty title",
			mutate:     func(p *NewSectionParams) { p.Title = "" },
			wantSubstr: "title",
		},
		{
			name:       "whitespace-only title",
			mutate:     func(p *NewSectionParams) { p.Title = "   \t\n  " },
			wantSubstr: "title",
		},
		{
			name:       "title exceeds 255 chars",
			mutate:     func(p *NewSectionParams) { p.Title = strings.Repeat("я", 256) },
			wantSubstr: "title",
		},
		{
			name: "description exceeds 4096 chars",
			mutate: func(p *NewSectionParams) {
				p.Description = strings.Repeat("a", 4097)
			},
			wantSubstr: "description",
		},
		{
			name:       "negative order_index",
			mutate:     func(p *NewSectionParams) { p.OrderIndex = -1 },
			wantSubstr: "order_index",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := validSectionParams(now)
			tc.mutate(&p)
			s, err := NewSection(p)
			if err == nil {
				t.Fatalf("NewSection accepted invalid params %+v", p)
			}
			if s != nil {
				t.Errorf("NewSection returned non-nil entity %+v alongside error", s)
			}
			if !errors.Is(err, ErrInvalidSection) {
				t.Errorf("error %v does not wrap ErrInvalidSection", err)
			}
			if !strings.Contains(err.Error(), tc.wantSubstr) {
				t.Errorf("error %q does not mention %q", err.Error(), tc.wantSubstr)
			}
		})
	}
}

func TestReconstituteSection_RoundtripsAllFields(t *testing.T) {
	created := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 5, 9, 18, 30, 0, 0, time.UTC)
	s := ReconstituteSection(
		101,        // id
		7,          // curriculumID
		"Раздел 1", // title
		"Описание", // description
		3,          // orderIndex
		5,          // version
		created,
		updated,
	)
	if s == nil {
		t.Fatal("ReconstituteSection returned nil")
	}
	if s.ID != 101 {
		t.Errorf("ID = %d, want 101", s.ID)
	}
	if got, want := s.CurriculumID(), int64(7); got != want {
		t.Errorf("CurriculumID() = %d, want %d", got, want)
	}
	if got, want := s.Title(), "Раздел 1"; got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
	if got, want := s.Description(), "Описание"; got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
	if got, want := s.OrderIndex(), 3; got != want {
		t.Errorf("OrderIndex() = %d, want %d", got, want)
	}
	if got, want := s.Version(), 5; got != want {
		t.Errorf("Version() = %d, want %d", got, want)
	}
	if !s.CreatedAt().Equal(created) {
		t.Errorf("CreatedAt() = %v, want %v", s.CreatedAt(), created)
	}
	if !s.UpdatedAt().Equal(updated) {
		t.Errorf("UpdatedAt() = %v, want %v", s.UpdatedAt(), updated)
	}
}
