package entities

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// freshSection builds a Section through ReconstituteSection so tests
// can mutate at known field values + a known initial timestamp.
func freshSection(t *testing.T, created time.Time) *Section {
	t.Helper()
	return ReconstituteSection(
		1,
		7,
		"Базовая часть",
		"Описание раздела",
		0,
		3,
		created,
		created,
	)
}

func TestSection_UpdateBasics_HappyPath(t *testing.T) {
	created := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	now := time.Date(2026, 5, 9, 14, 0, 0, 0, time.UTC)
	s := freshSection(t, created)

	if err := s.UpdateBasics("Вариативная часть", "Дисциплины по выбору", 2, now); err != nil {
		t.Fatalf("UpdateBasics returned unexpected error: %v", err)
	}
	if got, want := s.Title(), "Вариативная часть"; got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
	if got, want := s.Description(), "Дисциплины по выбору"; got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
	if got, want := s.OrderIndex(), 2; got != want {
		t.Errorf("OrderIndex() = %d, want %d", got, want)
	}
	if !s.UpdatedAt().Equal(now) {
		t.Errorf("UpdatedAt() = %v, want %v", s.UpdatedAt(), now)
	}
	if !s.CreatedAt().Equal(created) {
		t.Errorf("CreatedAt() = %v, must not change after UpdateBasics", s.CreatedAt())
	}
	// Version is repo-managed (ADR-3 — DB increments on UPDATE).
	// UpdateBasics on the entity must NOT touch it.
	if got, want := s.Version(), 3; got != want {
		t.Errorf("Version() = %d, want %d (UpdateBasics must not bump version — repo concern)", got, want)
	}
}

func TestSection_UpdateBasics_TrimsTextFields(t *testing.T) {
	now := time.Now()
	s := freshSection(t, now)
	if err := s.UpdateBasics("  Раздел 2  ", "  Описание  ", 1, now); err != nil {
		t.Fatalf("UpdateBasics returned unexpected error: %v", err)
	}
	if s.Title() != "Раздел 2" {
		t.Errorf("Title() = %q, want trimmed", s.Title())
	}
	if s.Description() != "Описание" {
		t.Errorf("Description() = %q, want trimmed", s.Description())
	}
}

// TestSection_UpdateBasics_AtomicOnFailure pins that a failed
// invariant leaves every field at its prior value — partial mutation
// would corrupt the aggregate (same gate as Curriculum.UpdateBasics).
func TestSection_UpdateBasics_AtomicOnFailure(t *testing.T) {
	created := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	s := freshSection(t, created)
	priorTitle := s.Title()
	priorDesc := s.Description()
	priorOrder := s.OrderIndex()
	priorUpdated := s.UpdatedAt()

	// Negative order_index — fails the last invariant; if the impl
	// applied mutations early, title/description would have changed
	// before the check.
	err := s.UpdateBasics("Новый заголовок", "Новое описание", -5, time.Now())
	if err == nil {
		t.Fatal("UpdateBasics accepted negative order_index")
	}
	if !errors.Is(err, ErrInvalidSection) {
		t.Errorf("error %v does not wrap ErrInvalidSection", err)
	}
	if s.Title() != priorTitle {
		t.Errorf("Title() leaked = %q, want %q", s.Title(), priorTitle)
	}
	if s.Description() != priorDesc {
		t.Errorf("Description() leaked = %q, want %q", s.Description(), priorDesc)
	}
	if s.OrderIndex() != priorOrder {
		t.Errorf("OrderIndex() leaked = %d, want %d", s.OrderIndex(), priorOrder)
	}
	if !s.UpdatedAt().Equal(priorUpdated) {
		t.Errorf("UpdatedAt() leaked = %v, want %v", s.UpdatedAt(), priorUpdated)
	}
}

// TestSection_UpdateBasics_InvariantViolations table-driven per
// CLAUDE.md ≥3-variant gate, mirroring the constructor cases.
func TestSection_UpdateBasics_InvariantViolations(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name        string
		title       string
		description string
		orderIndex  int
		wantSubstr  string
	}{
		{"empty title", "", "ok", 0, "title"},
		{"whitespace title", "   ", "ok", 0, "title"},
		{"title over 255", strings.Repeat("я", 256), "ok", 0, "title"},
		{"description over 4096", "ok", strings.Repeat("a", 4097), 0, "description"},
		{"negative order_index", "ok", "ok", -1, "order_index"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := freshSection(t, now)
			err := s.UpdateBasics(tc.title, tc.description, tc.orderIndex, now)
			if err == nil {
				t.Fatalf("UpdateBasics accepted invalid input")
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

// TestSection_AuthorizeEdit_StatusFrozenBeforeOwnership pins that the
// status gate fires before ownership — pending_approval / approved /
// archived curricula reject everyone, including admins (status
// inheritance per ADR-2). This is symmetric with Curriculum.AuthorizeEdit
// but the curriculum status is supplied as a primitive parameter
// because Section is an independent AR (ADR-1 Beta).
func TestSection_AuthorizeEdit_StatusFrozenBeforeOwnership(t *testing.T) {
	now := time.Now()
	s := freshSection(t, now)

	frozenStatuses := []CurriculumStatus{
		StatusPendingApproval,
		StatusApproved,
		StatusArchived,
	}
	for _, st := range frozenStatuses {
		t.Run(string(st), func(t *testing.T) {
			err := s.AuthorizeEdit(99, true /* isAdmin */, st, 99 /* curCreatedBy */)
			if err == nil {
				t.Fatalf("AuthorizeEdit allowed edit on status %q", st)
			}
			if !errors.Is(err, ErrCannotEditSection) {
				t.Errorf("error %v does not wrap ErrCannotEditSection", err)
			}
		})
	}
}

func TestSection_AuthorizeEdit_DraftAdminOverrideAllowed(t *testing.T) {
	now := time.Now()
	s := freshSection(t, now)
	// Admin acting on someone else's curriculum (curCreatedBy=42, actorID=99).
	if err := s.AuthorizeEdit(99, true, StatusDraft, 42); err != nil {
		t.Errorf("AuthorizeEdit denied admin on draft curriculum: %v", err)
	}
}

func TestSection_AuthorizeEdit_DraftAuthorMethodistAllowed(t *testing.T) {
	now := time.Now()
	s := freshSection(t, now)
	if err := s.AuthorizeEdit(42, false, StatusDraft, 42); err != nil {
		t.Errorf("AuthorizeEdit denied author methodist on draft: %v", err)
	}
}

func TestSection_AuthorizeEdit_NonAuthorMethodistDenied(t *testing.T) {
	now := time.Now()
	s := freshSection(t, now)
	err := s.AuthorizeEdit(99, false, StatusDraft, 42)
	if err == nil {
		t.Fatal("AuthorizeEdit allowed non-author methodist")
	}
	if !errors.Is(err, ErrSectionScopeForbidden) {
		t.Errorf("error %v does not wrap ErrSectionScopeForbidden", err)
	}
}

func TestSection_AuthorizeEdit_ZeroActorIDDenied(t *testing.T) {
	// Defense in depth: a JWT subject of 0 (lost upstream) must never
	// satisfy the actor==author check even when the curriculum was
	// created by user 0 (not possible in production, but the entity
	// must not implicitly trust the zero value). Mirror к Curriculum
	// AuthorizeEdit's `actorID > 0` clause.
	now := time.Now()
	s := freshSection(t, now)
	err := s.AuthorizeEdit(0, false, StatusDraft, 0)
	if err == nil {
		t.Fatal("AuthorizeEdit allowed zero actor id")
	}
	if !errors.Is(err, ErrSectionScopeForbidden) {
		t.Errorf("error %v does not wrap ErrSectionScopeForbidden", err)
	}
}
