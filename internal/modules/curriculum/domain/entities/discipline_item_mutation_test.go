package entities

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// freshItem builds a DisciplineItem через ReconstituteDisciplineItem
// at known field values + known initial timestamp. Mirror к freshSection
// helper in section_mutation_test.go.
func freshItem(t *testing.T, created time.Time) *DisciplineItem {
	t.Helper()
	return ReconstituteDisciplineItem(
		202, 7, "Математический анализ",
		36, 36, 0, 72,
		ControlFormExam,
		4, 1, 0, 5,
		created, created,
	)
}

func TestDisciplineItem_UpdateBasics_HappyPath(t *testing.T) {
	created := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	now := time.Date(2026, 5, 9, 14, 0, 0, 0, time.UTC)
	d := freshItem(t, created)

	err := d.UpdateBasics("Программирование", 24, 24, 36, 96,
		ControlFormDifferentialZachet, 5, 2, 1, now)
	if err != nil {
		t.Fatalf("UpdateBasics returned unexpected error: %v", err)
	}
	if got, want := d.Title(), "Программирование"; got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
	if got, want := d.HoursLectures(), 24; got != want {
		t.Errorf("HoursLectures() = %d, want %d", got, want)
	}
	if got, want := d.HoursLab(), 36; got != want {
		t.Errorf("HoursLab() = %d, want %d", got, want)
	}
	if got, want := d.HoursSelf(), 96; got != want {
		t.Errorf("HoursSelf() = %d, want %d", got, want)
	}
	if got, want := d.ControlForm(), ControlFormDifferentialZachet; got != want {
		t.Errorf("ControlForm() = %q, want %q", got, want)
	}
	if got, want := d.Credits(), 5; got != want {
		t.Errorf("Credits() = %d, want %d", got, want)
	}
	if got, want := d.Semester(), 2; got != want {
		t.Errorf("Semester() = %d, want %d", got, want)
	}
	if got, want := d.OrderIndex(), 1; got != want {
		t.Errorf("OrderIndex() = %d, want %d", got, want)
	}
	if !d.UpdatedAt().Equal(now) {
		t.Errorf("UpdatedAt() = %v, want %v", d.UpdatedAt(), now)
	}
	if !d.CreatedAt().Equal(created) {
		t.Errorf("CreatedAt() must not change after UpdateBasics, got %v", d.CreatedAt())
	}
	// Version is repo-managed (ADR-3) — entity UpdateBasics must NOT
	// touch it (mirror к Section behavior).
	if got, want := d.Version(), 5; got != want {
		t.Errorf("Version() = %d, want %d (UpdateBasics must not bump — repo concern)", got, want)
	}
}

// TestDisciplineItem_UpdateBasics_AtomicOnFailure pins atomic semantics —
// failed invariant leaves всё untouched (mirror к Section atomicity).
func TestDisciplineItem_UpdateBasics_AtomicOnFailure(t *testing.T) {
	d := freshItem(t, time.Now())
	priorTitle := d.Title()
	priorHours := d.HoursLectures()
	priorCredits := d.Credits()
	priorUpdated := d.UpdatedAt()

	// Negative credits — last invariant. If impl applied mutations
	// early, prior fields would have changed before the check fires.
	err := d.UpdateBasics("Новая дисциплина", 100, 100, 100, 100,
		ControlFormZachet, -5, 3, 2, time.Now())
	if err == nil {
		t.Fatal("UpdateBasics accepted negative credits")
	}
	if !errors.Is(err, ErrInvalidDisciplineItem) {
		t.Errorf("error %v does not wrap ErrInvalidDisciplineItem", err)
	}
	if d.Title() != priorTitle {
		t.Errorf("Title() leaked = %q, want %q", d.Title(), priorTitle)
	}
	if d.HoursLectures() != priorHours {
		t.Errorf("HoursLectures() leaked = %d, want %d", d.HoursLectures(), priorHours)
	}
	if d.Credits() != priorCredits {
		t.Errorf("Credits() leaked = %d, want %d", d.Credits(), priorCredits)
	}
	if !d.UpdatedAt().Equal(priorUpdated) {
		t.Errorf("UpdatedAt() leaked = %v, want %v", d.UpdatedAt(), priorUpdated)
	}
}

// TestDisciplineItem_UpdateBasics_InvariantViolations table-driven per
// CLAUDE.md ≥3-variant gate — mirrors constructor cases for the
// mutation path.
func TestDisciplineItem_UpdateBasics_InvariantViolations(t *testing.T) {
	now := time.Now()
	type testInputs struct {
		title         string
		hoursLectures int
		hoursPractice int
		hoursLab      int
		hoursSelf     int
		controlForm   ControlForm
		credits       int
		semester      int
		orderIndex    int
	}
	valid := testInputs{
		title:       "ok",
		controlForm: ControlFormExam,
		credits:     1,
		semester:    1,
	}
	cases := []struct {
		name       string
		mutate     func(*testInputs)
		wantSubstr string
	}{
		{"empty title", func(in *testInputs) { in.title = "" }, "title"},
		{"title over 255", func(in *testInputs) { in.title = strings.Repeat("я", 256) }, "title"},
		{"negative hours_lectures", func(in *testInputs) { in.hoursLectures = -1 }, "hours_lectures"},
		{"negative hours_practice", func(in *testInputs) { in.hoursPractice = -1 }, "hours_practice"},
		{"negative hours_lab", func(in *testInputs) { in.hoursLab = -1 }, "hours_lab"},
		{"negative hours_self", func(in *testInputs) { in.hoursSelf = -1 }, "hours_self"},
		{"invalid control_form", func(in *testInputs) { in.controlForm = "vyzov" }, "control_form"},
		{"negative credits", func(in *testInputs) { in.credits = -1 }, "credits"},
		{"semester 0", func(in *testInputs) { in.semester = 0 }, "semester"},
		{"semester 13", func(in *testInputs) { in.semester = 13 }, "semester"},
		{"negative order_index", func(in *testInputs) { in.orderIndex = -1 }, "order_index"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := valid
			tc.mutate(&in)
			d := freshItem(t, now)
			err := d.UpdateBasics(in.title, in.hoursLectures, in.hoursPractice,
				in.hoursLab, in.hoursSelf, in.controlForm,
				in.credits, in.semester, in.orderIndex, now)
			if err == nil {
				t.Fatalf("UpdateBasics accepted invalid input")
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

// TestAuthorizeDisciplineItemEdit_StatusFrozenBeforeOwnership pins
// gate ordering: non-editable curriculum status freezes everyone,
// admin override included (lifecycle inheritance per ADR-2). Mirror
// к Section.AuthorizeEdit StatusFrozenBeforeOwnership.
func TestAuthorizeDisciplineItemEdit_StatusFrozenBeforeOwnership(t *testing.T) {
	frozen := []CurriculumStatus{StatusPendingApproval, StatusApproved, StatusArchived}
	for _, st := range frozen {
		t.Run(string(st), func(t *testing.T) {
			err := AuthorizeDisciplineItemEdit(99, true /*isAdmin*/, st, 99 /*curCreatedBy*/)
			if err == nil {
				t.Fatalf("AuthorizeDisciplineItemEdit allowed edit on status %q", st)
			}
			if !errors.Is(err, ErrCannotEditDisciplineItem) {
				t.Errorf("error %v does not wrap ErrCannotEditDisciplineItem", err)
			}
		})
	}
}

func TestAuthorizeDisciplineItemEdit_DraftAdminOverrideAllowed(t *testing.T) {
	if err := AuthorizeDisciplineItemEdit(99, true, StatusDraft, 42); err != nil {
		t.Errorf("AuthorizeDisciplineItemEdit denied admin on draft: %v", err)
	}
}

func TestAuthorizeDisciplineItemEdit_DraftAuthorMethodistAllowed(t *testing.T) {
	if err := AuthorizeDisciplineItemEdit(42, false, StatusDraft, 42); err != nil {
		t.Errorf("AuthorizeDisciplineItemEdit denied author methodist on draft: %v", err)
	}
}

func TestAuthorizeDisciplineItemEdit_NonAuthorMethodistDenied(t *testing.T) {
	err := AuthorizeDisciplineItemEdit(99, false, StatusDraft, 42)
	if err == nil {
		t.Fatal("AuthorizeDisciplineItemEdit allowed non-author methodist")
	}
	if !errors.Is(err, ErrDisciplineItemScopeForbidden) {
		t.Errorf("error %v does not wrap ErrDisciplineItemScopeForbidden", err)
	}
}

func TestAuthorizeDisciplineItemEdit_ZeroActorIDDenied(t *testing.T) {
	// Defense in depth — lost JWT subject must not satisfy zero-zero
	// match (mirror к Section AuthorizeEdit zero-actor guard).
	err := AuthorizeDisciplineItemEdit(0, false, StatusDraft, 0)
	if err == nil {
		t.Fatal("AuthorizeDisciplineItemEdit allowed zero actor id")
	}
	if !errors.Is(err, ErrDisciplineItemScopeForbidden) {
		t.Errorf("error %v does not wrap ErrDisciplineItemScopeForbidden", err)
	}
}

// TestDisciplineItem_AuthorizeEdit_MethodDelegatesToFreeFunction pins
// that the method form (ergonomic call site for Update / Delete
// usecases) and the free function (used by Create with no instance
// yet) share identical logic — eliminating drift risk.
func TestDisciplineItem_AuthorizeEdit_MethodDelegatesToFreeFunction(t *testing.T) {
	d := freshItem(t, time.Now())
	cases := []struct {
		name         string
		actorID      int64
		isAdmin      bool
		curStatus    CurriculumStatus
		curCreatedBy int64
	}{
		{"admin override on draft", 99, true, StatusDraft, 42},
		{"author methodist on draft", 42, false, StatusDraft, 42},
		{"non-author methodist on draft", 99, false, StatusDraft, 42},
		{"frozen status with admin", 99, true, StatusApproved, 99},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			methodErr := d.AuthorizeEdit(tc.actorID, tc.isAdmin, tc.curStatus, tc.curCreatedBy)
			freeErr := AuthorizeDisciplineItemEdit(tc.actorID, tc.isAdmin, tc.curStatus, tc.curCreatedBy)
			if (methodErr == nil) != (freeErr == nil) {
				t.Fatalf("method/free function divergence: method=%v free=%v", methodErr, freeErr)
			}
			if methodErr != nil && freeErr != nil &&
				methodErr.Error() != freeErr.Error() {
				t.Errorf("method/free function error message divergence:\n  method: %v\n  free:   %v",
					methodErr, freeErr)
			}
		})
	}
}
