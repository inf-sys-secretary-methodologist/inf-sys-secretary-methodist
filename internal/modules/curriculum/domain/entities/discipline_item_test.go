package entities

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func validDisciplineItemParams(now time.Time) NewDisciplineItemParams {
	return NewDisciplineItemParams{
		SectionID:     7,
		Title:         "Математический анализ",
		HoursLectures: 36,
		HoursPractice: 36,
		HoursLab:      0,
		HoursSelf:     72,
		ControlForm:   ControlFormExam,
		Credits:       4,
		Semester:      1,
		OrderIndex:    0,
		Now:           now,
	}
}

func TestNewDisciplineItem_HappyPath(t *testing.T) {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	d, err := NewDisciplineItem(validDisciplineItemParams(now))
	if err != nil {
		t.Fatalf("NewDisciplineItem returned unexpected error: %v", err)
	}
	if d == nil {
		t.Fatal("NewDisciplineItem returned nil entity with no error")
	}
	if got, want := d.SectionID(), int64(7); got != want {
		t.Errorf("SectionID() = %d, want %d", got, want)
	}
	if got, want := d.Title(), "Математический анализ"; got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
	if got, want := d.HoursLectures(), 36; got != want {
		t.Errorf("HoursLectures() = %d, want %d", got, want)
	}
	if got, want := d.HoursPractice(), 36; got != want {
		t.Errorf("HoursPractice() = %d, want %d", got, want)
	}
	if got, want := d.HoursLab(), 0; got != want {
		t.Errorf("HoursLab() = %d, want %d", got, want)
	}
	if got, want := d.HoursSelf(), 72; got != want {
		t.Errorf("HoursSelf() = %d, want %d", got, want)
	}
	if got, want := d.ControlForm(), ControlFormExam; got != want {
		t.Errorf("ControlForm() = %q, want %q", got, want)
	}
	if got, want := d.Credits(), 4; got != want {
		t.Errorf("Credits() = %d, want %d", got, want)
	}
	if got, want := d.Semester(), 1; got != want {
		t.Errorf("Semester() = %d, want %d", got, want)
	}
	if got, want := d.OrderIndex(), 0; got != want {
		t.Errorf("OrderIndex() = %d, want %d", got, want)
	}
	if got, want := d.Version(), 0; got != want {
		t.Errorf("Version() = %d, want %d for fresh item", got, want)
	}
	if !d.CreatedAt().Equal(now) {
		t.Errorf("CreatedAt() = %v, want %v", d.CreatedAt(), now)
	}
}

func TestNewDisciplineItem_TrimsTitle(t *testing.T) {
	now := time.Now()
	p := validDisciplineItemParams(now)
	p.Title = "  Математический анализ  "
	d, err := NewDisciplineItem(p)
	if err != nil {
		t.Fatalf("NewDisciplineItem returned unexpected error: %v", err)
	}
	if got, want := d.Title(), "Математический анализ"; got != want {
		t.Errorf("Title() = %q, want trimmed %q", got, want)
	}
}

func TestNewDisciplineItem_AcceptsAllZeroHours(t *testing.T) {
	// Some disciplines have no lecture / practice / lab / self
	// breakdown specified — 0 hours per type is legitimate (will be
	// flagged by UI validation, not domain invariant).
	now := time.Now()
	p := validDisciplineItemParams(now)
	p.HoursLectures = 0
	p.HoursPractice = 0
	p.HoursLab = 0
	p.HoursSelf = 0
	d, err := NewDisciplineItem(p)
	if err != nil {
		t.Fatalf("NewDisciplineItem rejected zero hours: %v", err)
	}
	if d.HoursLectures()+d.HoursPractice()+d.HoursLab()+d.HoursSelf() != 0 {
		t.Errorf("total hours = %d, want 0", d.HoursLectures()+d.HoursPractice()+d.HoursLab()+d.HoursSelf())
	}
}

// TestNewDisciplineItem_InvariantViolations is table-driven per
// CLAUDE.md ≥3-variant gate. Each row pins one rejection path,
// wrapping ErrInvalidDisciplineItem so handlers can errors.Is the
// sentinel for the 422 mapping.
func TestNewDisciplineItem_InvariantViolations(t *testing.T) {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name       string
		mutate     func(*NewDisciplineItemParams)
		wantSubstr string
	}{
		{"empty section_id", func(p *NewDisciplineItemParams) { p.SectionID = 0 }, "section_id"},
		{"negative section_id", func(p *NewDisciplineItemParams) { p.SectionID = -3 }, "section_id"},
		{"empty title", func(p *NewDisciplineItemParams) { p.Title = "" }, "title"},
		{"whitespace title", func(p *NewDisciplineItemParams) { p.Title = "   " }, "title"},
		{"title over 255", func(p *NewDisciplineItemParams) { p.Title = strings.Repeat("я", 256) }, "title"},
		{"negative hours_lectures", func(p *NewDisciplineItemParams) { p.HoursLectures = -1 }, "hours_lectures"},
		{"negative hours_practice", func(p *NewDisciplineItemParams) { p.HoursPractice = -1 }, "hours_practice"},
		{"negative hours_lab", func(p *NewDisciplineItemParams) { p.HoursLab = -1 }, "hours_lab"},
		{"negative hours_self", func(p *NewDisciplineItemParams) { p.HoursSelf = -1 }, "hours_self"},
		{"negative credits", func(p *NewDisciplineItemParams) { p.Credits = -1 }, "credits"},
		{"semester below 1", func(p *NewDisciplineItemParams) { p.Semester = 0 }, "semester"},
		{"semester above 12", func(p *NewDisciplineItemParams) { p.Semester = 13 }, "semester"},
		{"empty control_form", func(p *NewDisciplineItemParams) { p.ControlForm = "" }, "control_form"},
		{"unknown control_form", func(p *NewDisciplineItemParams) { p.ControlForm = "vyzov" }, "control_form"},
		{"negative order_index", func(p *NewDisciplineItemParams) { p.OrderIndex = -1 }, "order_index"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := validDisciplineItemParams(now)
			tc.mutate(&p)
			d, err := NewDisciplineItem(p)
			if err == nil {
				t.Fatalf("NewDisciplineItem accepted invalid params %+v", p)
			}
			if d != nil {
				t.Errorf("NewDisciplineItem returned non-nil entity %+v alongside error", d)
			}
			if !errors.Is(err, ErrInvalidDisciplineItem) {
				t.Errorf("error %v does not wrap ErrInvalidDisciplineItem", err)
			}
			if !strings.Contains(err.Error(), tc.wantSubstr) {
				t.Errorf("error %q does not mention %q", err.Error(), tc.wantSubstr)
			}
		})
	}
}

func TestReconstituteDisciplineItem_RoundtripsAllFields(t *testing.T) {
	created := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 5, 9, 18, 30, 0, 0, time.UTC)
	d := ReconstituteDisciplineItem(
		202,                // id
		7,                  // sectionID
		"Программирование", // title
		36, 18, 36, 90,     // hours: lectures, practice, lab, self
		ControlFormDifferentialZachet, // controlForm
		5, 3, 2, 8,                    // credits, semester, orderIndex, version
		created, updated,
	)
	if d == nil {
		t.Fatal("ReconstituteDisciplineItem returned nil")
	}
	if d.ID != 202 {
		t.Errorf("ID = %d, want 202", d.ID)
	}
	if got, want := d.SectionID(), int64(7); got != want {
		t.Errorf("SectionID() = %d, want %d", got, want)
	}
	if got, want := d.Title(), "Программирование"; got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
	if got, want := d.HoursLab(), 36; got != want {
		t.Errorf("HoursLab() = %d, want %d", got, want)
	}
	if got, want := d.ControlForm(), ControlFormDifferentialZachet; got != want {
		t.Errorf("ControlForm() = %q, want %q", got, want)
	}
	if got, want := d.Credits(), 5; got != want {
		t.Errorf("Credits() = %d, want %d", got, want)
	}
	if got, want := d.Semester(), 3; got != want {
		t.Errorf("Semester() = %d, want %d", got, want)
	}
	if got, want := d.Version(), 8; got != want {
		t.Errorf("Version() = %d, want %d", got, want)
	}
	if !d.CreatedAt().Equal(created) {
		t.Errorf("CreatedAt() = %v, want %v", d.CreatedAt(), created)
	}
	if !d.UpdatedAt().Equal(updated) {
		t.Errorf("UpdatedAt() = %v, want %v", d.UpdatedAt(), updated)
	}
}
