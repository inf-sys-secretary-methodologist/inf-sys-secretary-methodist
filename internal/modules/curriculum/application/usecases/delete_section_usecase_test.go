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

type fakeSectionDeleteRepo struct {
	got         *entities.Section
	getErr      error
	deleteErr   error
	deleteCalls int
	deletedID   int64
}

func (f *fakeSectionDeleteRepo) GetByID(_ context.Context, id int64) (*entities.Section, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.got, nil
}

func (f *fakeSectionDeleteRepo) Delete(_ context.Context, id int64) error {
	f.deleteCalls++
	f.deletedID = id
	return f.deleteErr
}

func TestNewDeleteSectionUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewDeleteSectionUseCase(nil, ..., ...) did not panic")
		}
	}()
	NewDeleteSectionUseCase(nil, &fakeCurriculumLookup{}, &recordingAuditSink{})
}

func TestNewDeleteSectionUseCase_PanicsOnNilCurriculumLookup(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewDeleteSectionUseCase(repo, nil, ...) did not panic")
		}
	}()
	NewDeleteSectionUseCase(&fakeSectionDeleteRepo{}, nil, &recordingAuditSink{})
}

func TestDeleteSectionUseCase_HappyPath_AuthorMethodist(t *testing.T) {
	now := time.Now()
	s := entities.ReconstituteSection(101, 7, "T", "d", 0, 0, now, now)
	cur := draftCurriculum(t, 42)
	repo := &fakeSectionDeleteRepo{got: s}
	lookup := &fakeCurriculumLookup{got: cur}
	audit := &recordingAuditSink{}

	uc := NewDeleteSectionUseCase(repo, lookup, audit)
	err := uc.Execute(context.Background(), 42, false, 101)
	require.NoError(t, err)
	assert.Equal(t, 1, repo.deleteCalls)
	assert.Equal(t, int64(101), repo.deletedID)

	require.Len(t, audit.events, 1)
	ev := audit.events[0]
	assert.Equal(t, "section.deleted", ev.Action)
	assert.Equal(t, int64(101), ev.Fields["section_id"])
	assert.Equal(t, int64(7), ev.Fields["curriculum_id"])
}

func TestDeleteSectionUseCase_AdminOverride(t *testing.T) {
	now := time.Now()
	s := entities.ReconstituteSection(101, 7, "T", "d", 0, 0, now, now)
	cur := draftCurriculum(t, 42)
	repo := &fakeSectionDeleteRepo{got: s}
	lookup := &fakeCurriculumLookup{got: cur}
	audit := &recordingAuditSink{}

	uc := NewDeleteSectionUseCase(repo, lookup, audit)
	err := uc.Execute(context.Background(), 99, true, 101)
	require.NoError(t, err)
	assert.Equal(t, 1, repo.deleteCalls)
}

func TestDeleteSectionUseCase_NonAuthorMethodistDenied(t *testing.T) {
	now := time.Now()
	s := entities.ReconstituteSection(101, 7, "T", "d", 0, 0, now, now)
	cur := draftCurriculum(t, 42)
	repo := &fakeSectionDeleteRepo{got: s}
	lookup := &fakeCurriculumLookup{got: cur}
	audit := &recordingAuditSink{}

	uc := NewDeleteSectionUseCase(repo, lookup, audit)
	err := uc.Execute(context.Background(), 99, false, 101)
	assert.True(t, errors.Is(err, entities.ErrSectionScopeForbidden))
	assert.Equal(t, 0, repo.deleteCalls)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "section.delete_denied", audit.events[0].Action)
	assert.Equal(t, "forbidden", audit.events[0].Fields["reason"])
}

func TestDeleteSectionUseCase_FrozenStatusDenied(t *testing.T) {
	now := time.Now()
	s := entities.ReconstituteSection(101, 7, "T", "d", 0, 0, now, now)
	cur := frozenCurriculum(t, entities.StatusArchived, 42)
	repo := &fakeSectionDeleteRepo{got: s}
	lookup := &fakeCurriculumLookup{got: cur}
	audit := &recordingAuditSink{}

	uc := NewDeleteSectionUseCase(repo, lookup, audit)
	err := uc.Execute(context.Background(), 42, false, 101)
	assert.True(t, errors.Is(err, entities.ErrCannotEditSection))
	assert.Equal(t, 0, repo.deleteCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "not_editable", audit.events[0].Fields["reason"])
}

func TestDeleteSectionUseCase_SectionNotFound(t *testing.T) {
	repo := &fakeSectionDeleteRepo{getErr: repositories.ErrSectionNotFound}
	lookup := &fakeCurriculumLookup{}
	audit := &recordingAuditSink{}

	uc := NewDeleteSectionUseCase(repo, lookup, audit)
	err := uc.Execute(context.Background(), 42, false, 999)
	assert.True(t, errors.Is(err, repositories.ErrSectionNotFound))
	assert.Equal(t, 0, repo.deleteCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "section.delete_denied", audit.events[0].Action)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}
