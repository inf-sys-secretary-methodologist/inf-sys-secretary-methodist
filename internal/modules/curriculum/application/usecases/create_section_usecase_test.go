package usecases

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// ===== Fakes =====

type fakeSectionSaveRepo struct {
	saved      *entities.Section
	idAssigned int64
	saveErr    error
	saveCalls  int
}

func (f *fakeSectionSaveRepo) Save(_ context.Context, s *entities.Section) error {
	f.saveCalls++
	f.saved = s
	if f.saveErr != nil {
		return f.saveErr
	}
	if f.idAssigned > 0 {
		s.ID = f.idAssigned
	}
	return nil
}

type fakeCurriculumLookup struct {
	got    *entities.Curriculum
	getErr error
	calls  int
	gotID  int64
}

func (f *fakeCurriculumLookup) GetByID(_ context.Context, id int64) (*entities.Curriculum, error) {
	f.calls++
	f.gotID = id
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.got, nil
}

// draftCurriculum builds a curriculum в draft status for "editable"
// authorization tests. ID exposed to prove the lookup wired correctly.
func draftCurriculum(t *testing.T, createdBy int64) *entities.Curriculum {
	t.Helper()
	c, err := entities.NewCurriculum(entities.NewCurriculumParams{
		Title:       "ИВТ-2026",
		Code:        fmt.Sprintf("CODE-%d", createdBy),
		Specialty:   "ИВТ",
		Year:        2026,
		Description: "",
		CreatedBy:   createdBy,
		Now:         time.Now(),
	})
	require.NoError(t, err)
	c.ID = 7
	return c
}

// frozenCurriculum loads a curriculum-shaped entity at a non-editable
// status (pending_approval / approved / archived). We use
// ReconstituteCurriculum because the constructor only ever returns
// drafts.
func frozenCurriculum(t *testing.T, status entities.CurriculumStatus, createdBy int64) *entities.Curriculum {
	t.Helper()
	now := time.Now()
	return entities.ReconstituteCurriculum(
		7, "ИВТ-2026", "C-1", "ИВТ", 2026, "",
		status, createdBy, nil, nil, now, now, 0,
	)
}

// ===== Tests =====

func TestNewCreateSectionUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewCreateSectionUseCase(nil, ..., ...) did not panic")
		}
	}()
	NewCreateSectionUseCase(nil, &fakeCurriculumLookup{}, &recordingAuditSink{}, time.Now)
}

func TestNewCreateSectionUseCase_PanicsOnNilCurriculumLookup(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewCreateSectionUseCase(repo, nil, ...) did not panic")
		}
	}()
	NewCreateSectionUseCase(&fakeSectionSaveRepo{}, nil, &recordingAuditSink{}, time.Now)
}

func TestCreateSectionUseCase_HappyPath_Author(t *testing.T) {
	cur := draftCurriculum(t, 42)
	repo := &fakeSectionSaveRepo{idAssigned: 101}
	lookup := &fakeCurriculumLookup{got: cur}
	audit := &recordingAuditSink{}
	frozenNow := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)

	uc := NewCreateSectionUseCase(repo, lookup, audit, func() time.Time { return frozenNow })
	s, err := uc.Execute(context.Background(), 42, false, CreateSectionInput{
		CurriculumID: 7,
		Title:        "Базовая часть",
		Description:  "",
		OrderIndex:   0,
	})
	require.NoError(t, err)
	require.NotNil(t, s)
	assert.Equal(t, int64(101), s.ID)
	assert.Equal(t, int64(7), s.CurriculumID())
	assert.Equal(t, "Базовая часть", s.Title())

	require.Equal(t, 1, repo.saveCalls)
	require.Equal(t, 1, lookup.calls)
	assert.Equal(t, int64(7), lookup.gotID, "lookup must be called with input.CurriculumID")

	// Audit trail captures success event.
	require.Len(t, audit.events, 1)
	ev := audit.events[0]
	assert.Equal(t, "section.created", ev.Action)
	assert.Equal(t, "curriculum_section", ev.Resource)
	assert.Equal(t, int64(42), ev.Fields["actor_user_id"])
	assert.Equal(t, int64(101), ev.Fields["section_id"])
	assert.Equal(t, int64(7), ev.Fields["curriculum_id"])
}

func TestCreateSectionUseCase_AdminOverride_NonAuthor(t *testing.T) {
	// Curriculum authored by user 42; admin (user 99) creates a section.
	cur := draftCurriculum(t, 42)
	repo := &fakeSectionSaveRepo{idAssigned: 101}
	lookup := &fakeCurriculumLookup{got: cur}
	audit := &recordingAuditSink{}

	uc := NewCreateSectionUseCase(repo, lookup, audit, time.Now)
	s, err := uc.Execute(context.Background(), 99, true, CreateSectionInput{
		CurriculumID: 7,
		Title:        "Раздел админа",
		OrderIndex:   1,
	})
	require.NoError(t, err)
	require.NotNil(t, s)
	assert.Equal(t, int64(101), s.ID)
}

func TestCreateSectionUseCase_NonAuthorDenied(t *testing.T) {
	cur := draftCurriculum(t, 42)
	repo := &fakeSectionSaveRepo{}
	lookup := &fakeCurriculumLookup{got: cur}
	audit := &recordingAuditSink{}

	uc := NewCreateSectionUseCase(repo, lookup, audit, time.Now)
	_, err := uc.Execute(context.Background(), 99, false, CreateSectionInput{
		CurriculumID: 7,
		Title:        "Раздел чужого",
		OrderIndex:   0,
	})
	assert.True(t, errors.Is(err, entities.ErrSectionScopeForbidden),
		"err must wrap ErrSectionScopeForbidden, got %v", err)
	assert.Equal(t, 0, repo.saveCalls, "Save must not be called on auth denial")

	require.Len(t, audit.events, 1)
	assert.Equal(t, "section.create_denied", audit.events[0].Action)
	assert.Equal(t, "forbidden", audit.events[0].Fields["reason"])
}

func TestCreateSectionUseCase_FrozenStatusDenied(t *testing.T) {
	cur := frozenCurriculum(t, entities.StatusPendingApproval, 42)
	repo := &fakeSectionSaveRepo{}
	lookup := &fakeCurriculumLookup{got: cur}
	audit := &recordingAuditSink{}

	uc := NewCreateSectionUseCase(repo, lookup, audit, time.Now)
	// Even the author hits the status freeze — gate ordering
	// (status before ownership) per ADR-2.
	_, err := uc.Execute(context.Background(), 42, false, CreateSectionInput{
		CurriculumID: 7,
		Title:        "Раздел в pending",
		OrderIndex:   0,
	})
	assert.True(t, errors.Is(err, entities.ErrCannotEditSection),
		"err must wrap ErrCannotEditSection, got %v", err)
	assert.Equal(t, 0, repo.saveCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "section.create_denied", audit.events[0].Action)
	assert.Equal(t, "not_editable", audit.events[0].Fields["reason"])
}

func TestCreateSectionUseCase_CurriculumNotFound(t *testing.T) {
	repo := &fakeSectionSaveRepo{}
	lookup := &fakeCurriculumLookup{getErr: repositories.ErrCurriculumNotFound}
	audit := &recordingAuditSink{}

	uc := NewCreateSectionUseCase(repo, lookup, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, CreateSectionInput{
		CurriculumID: 999,
		Title:        "Раздел в призраке",
	})
	assert.True(t, errors.Is(err, repositories.ErrCurriculumNotFound),
		"err must wrap ErrCurriculumNotFound, got %v", err)
	assert.Equal(t, 0, repo.saveCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "section.create_denied", audit.events[0].Action)
	assert.Equal(t, "curriculum_not_found", audit.events[0].Fields["reason"])
}

func TestCreateSectionUseCase_InvalidInput(t *testing.T) {
	cur := draftCurriculum(t, 42)
	repo := &fakeSectionSaveRepo{}
	lookup := &fakeCurriculumLookup{got: cur}
	audit := &recordingAuditSink{}

	uc := NewCreateSectionUseCase(repo, lookup, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, CreateSectionInput{
		CurriculumID: 7,
		Title:        "", // invariant: title non-empty
		OrderIndex:   0,
	})
	assert.True(t, errors.Is(err, entities.ErrInvalidSection),
		"err must wrap ErrInvalidSection, got %v", err)
	assert.Equal(t, 0, repo.saveCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "section.create_denied", audit.events[0].Action)
	assert.Equal(t, "invalid", audit.events[0].Fields["reason"])
}

func TestCreateSectionUseCase_CurriculumLookupTransportErrorPropagates(t *testing.T) {
	repo := &fakeSectionSaveRepo{}
	lookup := &fakeCurriculumLookup{getErr: errors.New("conn refused")}
	audit := &recordingAuditSink{}

	uc := NewCreateSectionUseCase(repo, lookup, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, CreateSectionInput{
		CurriculumID: 7,
		Title:        "Раздел",
	})
	require.Error(t, err)
	// Transport errors must NOT produce audit events (operators read
	// them from logger stack traces, not the audit log) — same
	// contract as curriculum use cases.
	assert.Len(t, audit.events, 0,
		"transport errors must not emit audit events")
	assert.Equal(t, 0, repo.saveCalls)
}

func TestCreateSectionUseCase_RepoSaveTransportErrorPropagates(t *testing.T) {
	cur := draftCurriculum(t, 42)
	repo := &fakeSectionSaveRepo{saveErr: errors.New("db down")}
	lookup := &fakeCurriculumLookup{got: cur}
	audit := &recordingAuditSink{}

	uc := NewCreateSectionUseCase(repo, lookup, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, CreateSectionInput{
		CurriculumID: 7,
		Title:        "Раздел",
	})
	require.Error(t, err)
	assert.Len(t, audit.events, 0)
}
