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

func TestUpdateDisciplineItem_HappyPath(t *testing.T) {
	now := time.Now()
	d := entities.ReconstituteDisciplineItem(202, 11, "Старый", 18, 18, 0, 36,
		entities.ControlFormZachet, 2, 1, 0, 5, now, now)
	repo := &fakeDisciplineItemUpdateRepo{got: d}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: draftCurriculumForItem(t, 42)}
	audit := &recordingAuditSink{}
	frozenNow := time.Date(2026, 5, 9, 14, 0, 0, 0, time.UTC)

	uc := NewUpdateDisciplineItemUseCase(repo, section, curriculum, audit, func() time.Time { return frozenNow })
	got, err := uc.Execute(context.Background(), 42, false, UpdateDisciplineItemInput{
		ID: 202, Title: "Новый",
		HoursLectures: 36, HoursPractice: 36, HoursLab: 0, HoursSelf: 72,
		ControlForm: entities.ControlFormExam, Credits: 4, Semester: 2, OrderIndex: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Новый", got.Title())
	assert.Equal(t, 1, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.updated", audit.events[0].Action)
}

func TestUpdateDisciplineItem_VersionConflict(t *testing.T) {
	now := time.Now()
	d := entities.ReconstituteDisciplineItem(202, 11, "T", 18, 18, 0, 36,
		entities.ControlFormZachet, 2, 1, 0, 3, now, now)
	repo := &fakeDisciplineItemUpdateRepo{got: d, updateErr: repositories.ErrDisciplineItemVersionConflict}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: draftCurriculumForItem(t, 42)}
	audit := &recordingAuditSink{}

	uc := NewUpdateDisciplineItemUseCase(repo, section, curriculum, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, UpdateDisciplineItemInput{
		ID: 202, Title: "T",
		HoursLectures: 18, HoursPractice: 18, HoursSelf: 36,
		ControlForm: entities.ControlFormZachet, Credits: 2, Semester: 1,
	})
	assert.True(t, errors.Is(err, repositories.ErrDisciplineItemVersionConflict))
	require.Len(t, audit.events, 1)
	assert.Equal(t, "version_conflict", audit.events[0].Fields["reason"])
}

func TestUpdateDisciplineItem_ItemNotFound(t *testing.T) {
	repo := &fakeDisciplineItemUpdateRepo{getErr: repositories.ErrDisciplineItemNotFound}
	section := &fakeSectionLookup{}
	curriculum := &fakeCurriculumLookup{}
	audit := &recordingAuditSink{}

	uc := NewUpdateDisciplineItemUseCase(repo, section, curriculum, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, UpdateDisciplineItemInput{ID: 999, Title: "T"})
	assert.True(t, errors.Is(err, repositories.ErrDisciplineItemNotFound))
	assert.Equal(t, 0, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.update_denied", audit.events[0].Action)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}

// TestUpdateDisciplineItem_FrozenStatusDenied pins the not_editable
// audit reason on the Update denial path.
func TestUpdateDisciplineItem_FrozenStatusDenied(t *testing.T) {
	now := time.Now()
	d := entities.ReconstituteDisciplineItem(202, 11, "T", 18, 18, 0, 36,
		entities.ControlFormZachet, 2, 1, 0, 0, now, now)
	repo := &fakeDisciplineItemUpdateRepo{got: d}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: frozenCurriculumForItem(t, entities.StatusPendingApproval, 42)}
	audit := &recordingAuditSink{}

	uc := NewUpdateDisciplineItemUseCase(repo, section, curriculum, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, UpdateDisciplineItemInput{
		ID: 202, Title: "T",
		HoursLectures: 18, HoursPractice: 18, HoursSelf: 36,
		ControlForm: entities.ControlFormZachet, Credits: 2, Semester: 1,
	})
	assert.True(t, errors.Is(err, entities.ErrCannotEditDisciplineItem))
	assert.Equal(t, 0, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.update_denied", audit.events[0].Action)
	assert.Equal(t, "not_editable", audit.events[0].Fields["reason"])
}

// TestUpdateDisciplineItem_NonAuthorMethodistDenied pins the forbidden
// audit reason on the Update denial path (non-author methodist trying
// to edit someone else's curriculum's items).
func TestUpdateDisciplineItem_NonAuthorMethodistDenied(t *testing.T) {
	now := time.Now()
	d := entities.ReconstituteDisciplineItem(202, 11, "T", 18, 18, 0, 36,
		entities.ControlFormZachet, 2, 1, 0, 0, now, now)
	repo := &fakeDisciplineItemUpdateRepo{got: d}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: draftCurriculumForItem(t, 42)} // owned by 42
	audit := &recordingAuditSink{}

	uc := NewUpdateDisciplineItemUseCase(repo, section, curriculum, audit, time.Now)
	_, err := uc.Execute(context.Background(), 99, false, UpdateDisciplineItemInput{ // actor 99 ≠ owner 42
		ID: 202, Title: "T",
		HoursLectures: 18, HoursPractice: 18, HoursSelf: 36,
		ControlForm: entities.ControlFormZachet, Credits: 2, Semester: 1,
	})
	assert.True(t, errors.Is(err, entities.ErrDisciplineItemScopeForbidden))
	assert.Equal(t, 0, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.update_denied", audit.events[0].Action)
	assert.Equal(t, "forbidden", audit.events[0].Fields["reason"])
}

// TestUpdateDisciplineItem_InvalidInput pins the invalid audit reason
// on the Update denial path (UpdateBasics invariant fail).
func TestUpdateDisciplineItem_InvalidInput(t *testing.T) {
	now := time.Now()
	d := entities.ReconstituteDisciplineItem(202, 11, "T", 18, 18, 0, 36,
		entities.ControlFormZachet, 2, 1, 0, 0, now, now)
	repo := &fakeDisciplineItemUpdateRepo{got: d}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: draftCurriculumForItem(t, 42)}
	audit := &recordingAuditSink{}

	uc := NewUpdateDisciplineItemUseCase(repo, section, curriculum, audit, time.Now)
	_, err := uc.Execute(context.Background(), 42, false, UpdateDisciplineItemInput{
		ID: 202, Title: "", // invariant fail
		HoursLectures: 18, HoursPractice: 18, HoursSelf: 36,
		ControlForm: entities.ControlFormZachet, Credits: 2, Semester: 1,
	})
	assert.True(t, errors.Is(err, entities.ErrInvalidDisciplineItem))
	assert.Equal(t, 0, repo.updateCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.update_denied", audit.events[0].Action)
	assert.Equal(t, "invalid", audit.events[0].Fields["reason"])
}
