package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// fakeSectionUpdateRepo combines the GetByID + Update narrow port the
// update use case needs.
type fakeSectionUpdateRepo struct {
	got         *entities.Section
	getErr      error
	updateErr   error
	updateCalls int
	updated     *entities.Section
}

func (f *fakeSectionUpdateRepo) GetByID(_ context.Context, id int64) (*entities.Section, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.got != nil && f.got.ID != id {
		// Defense — caller mismatched fake.
		return nil, errors.New("fake: GetByID id mismatch")
	}
	return f.got, nil
}

func (f *fakeSectionUpdateRepo) Update(_ context.Context, s *entities.Section) error {
	f.updateCalls++
	f.updated = s
	return f.updateErr
}

func TestNewUpdateSectionUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewUpdateSectionUseCase(nil, ..., ...) did not panic")
		}
	}()
	NewUpdateSectionUseCase(nil, &fakeCurriculumLookup{}, &recordingAuditSink{}, time.Now)
}

func TestNewUpdateSectionUseCase_PanicsOnNilCurriculumLookup(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewUpdateSectionUseCase(repo, nil, ...) did not panic")
		}
	}()
	NewUpdateSectionUseCase(&fakeSectionUpdateRepo{}, nil, &recordingAuditSink{}, time.Now)
}

func TestUpdateSectionUseCase_HappyPath_AuthorMethodist(t *testing.T) {
	now := time.Now()
	s := entities.ReconstituteSection(101, 7, "Старый", "old", 0, 3, now, now)
	cur := draftCurriculum(t, 42)
	repo := &fakeSectionUpdateRepo{got: s}
	lookup := &fakeCurriculumLookup{got: cur}
	audit := &recordingAuditSink{}
	frozenNow := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)

	uc := NewUpdateSectionUseCase(repo, lookup, audit, func() time.Time { return frozenNow })
	got, err := uc.Execute(context.Background(), 42, false, UpdateSectionInput{
		ID:          101,
		Title:       "Новый",
		Description: "new desc",
		OrderIndex:  2,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Новый", got.Title())
	assert.Equal(t, "new desc", got.Description())
	assert.Equal(t, 2, got.OrderIndex())
	assert.Equal(t, frozenNow, got.UpdatedAt())
	assert.Equal(t, 1, repo.updateCalls)

	require.Len(t, audit.events, 1)
	ev := audit.events[0]
	assert.Equal(t, "section.updated", ev.Action)
	assert.Equal(t, int64(101), ev.Fields["section_id"])
	assert.Equal(t, int64(7), ev.Fields["curriculum_id"])
}

func TestUpdateSectionUseCase_AdminOverride(t *testing.T) {
	now := time.Now()
	s := entities.ReconstituteSection(101, 7, "T", "d", 0, 0, now, now)
	cur := draftCurriculum(t, 42)
	repo := &fakeSectionUpdateRepo{got: s}
	lookup := &fakeCurriculumLookup{got: cur}
	audit := &recordingAuditSink{}

	uc := NewUpdateSectionUseCase(repo, lookup, audit, time.Now)
	_, err := uc.Execute(context.Background(), 99, true, UpdateSectionInput{
		ID:    101,
		Title: "Админ правит",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.updateCalls)
}

func TestUpdateSectionUseCase_NonAuthorMethodistDenied(t *testing.T) {
	now := time.Now()
	s := entities.ReconstituteSection(101, 7, "T", "d", 0, 0, now, now)
	cur := draftCurriculum(t, 42)
	repo := &fakeSectionUpdateRepo{got: s}
	lookup := &fakeCurriculumLookup{got: cur}
	audit := &recordingAuditSink{}

	uc := NewUpdateSectionUseCase(repo, lookup, audit, time.Now)
	_, err := uc.Execute(context.Background(), 99, false, UpdateSectionInput{
		ID:    101,
		Title: "Чужой правит",
	})
	assert.True(t, errors.Is(err, entities.ErrSectionScopeForbidden))
	assert.Equal(t, 0, repo.updateCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "section.update_denied", audit.events[0].Action)
	assert.Equal(t, "forbidden", audit.events[0].Fields["reason"])
}

func TestUpdateSectionUseCase_FrozenStatusDenied(t *testing.T) {
	now := time.Now()
	s := entities.ReconstituteSection(101, 7, "T", "d", 0, 0, now, now)
	cur := frozenCurriculum(t, entities.StatusApproved, 42)
	repo := &fakeSectionUpdateRepo{got: s}
	lookup := &fakeCurriculumLookup{got: cur}
	audit := &recordingAuditSink{}

	uc := NewUpdateSectionUseCase(repo, lookup, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, UpdateSectionInput{
		ID:    101,
		Title: "После approve",
	})
	assert.True(t, errors.Is(err, entities.ErrCannotEditSection))
	assert.Equal(t, 0, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "not_editable", audit.events[0].Fields["reason"])
}

func TestUpdateSectionUseCase_SectionNotFound(t *testing.T) {
	repo := &fakeSectionUpdateRepo{getErr: repositories.ErrSectionNotFound}
	lookup := &fakeCurriculumLookup{}
	audit := &recordingAuditSink{}

	uc := NewUpdateSectionUseCase(repo, lookup, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, UpdateSectionInput{
		ID:    999,
		Title: "Призрак",
	})
	assert.True(t, errors.Is(err, repositories.ErrSectionNotFound))
	require.Len(t, audit.events, 1)
	assert.Equal(t, "section.update_denied", audit.events[0].Action)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}

func TestUpdateSectionUseCase_CurriculumNotFound(t *testing.T) {
	// Section exists, but its parent curriculum disappeared (FK CASCADE
	// would normally prevent this, but defense-in-depth).
	now := time.Now()
	s := entities.ReconstituteSection(101, 7, "T", "d", 0, 0, now, now)
	repo := &fakeSectionUpdateRepo{got: s}
	lookup := &fakeCurriculumLookup{getErr: repositories.ErrCurriculumNotFound}
	audit := &recordingAuditSink{}

	uc := NewUpdateSectionUseCase(repo, lookup, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, UpdateSectionInput{
		ID:    101,
		Title: "Орфан",
	})
	assert.True(t, errors.Is(err, repositories.ErrCurriculumNotFound))
	assert.Equal(t, 0, repo.updateCalls)
}

func TestUpdateSectionUseCase_InvalidInput(t *testing.T) {
	now := time.Now()
	s := entities.ReconstituteSection(101, 7, "T", "d", 0, 0, now, now)
	cur := draftCurriculum(t, 42)
	repo := &fakeSectionUpdateRepo{got: s}
	lookup := &fakeCurriculumLookup{got: cur}
	audit := &recordingAuditSink{}

	uc := NewUpdateSectionUseCase(repo, lookup, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, UpdateSectionInput{
		ID:         101,
		Title:      "", // invariant: title non-empty
		OrderIndex: 0,
	})
	assert.True(t, errors.Is(err, entities.ErrInvalidSection))
	assert.Equal(t, 0, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "invalid", audit.events[0].Fields["reason"])
}

func TestUpdateSectionUseCase_VersionConflict(t *testing.T) {
	// Repository optimistic-lock fired between load and write —
	// surface ErrSectionVersionConflict, no audit on transport-class
	// errors per established convention.
	now := time.Now()
	s := entities.ReconstituteSection(101, 7, "T", "d", 0, 3, now, now)
	cur := draftCurriculum(t, 42)
	repo := &fakeSectionUpdateRepo{got: s, updateErr: repositories.ErrSectionVersionConflict}
	lookup := &fakeCurriculumLookup{got: cur}
	audit := &recordingAuditSink{}

	uc := NewUpdateSectionUseCase(repo, lookup, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, UpdateSectionInput{
		ID:    101,
		Title: "Новый",
	})
	assert.True(t, errors.Is(err, repositories.ErrSectionVersionConflict))
	require.Len(t, audit.events, 1, "version conflict is a domain-policy event, audited")
	assert.Equal(t, "section.update_denied", audit.events[0].Action)
	assert.Equal(t, "version_conflict", audit.events[0].Fields["reason"])
}
