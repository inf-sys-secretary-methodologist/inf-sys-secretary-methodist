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

func TestDeleteDisciplineItem_HappyPath(t *testing.T) {
	now := time.Now()
	d := entities.ReconstituteDisciplineItem(202, 11, "T", 18, 18, 0, 36,
		entities.ControlFormZachet, 2, 1, 0, 0, now, now)
	repo := &fakeDisciplineItemDeleteRepo{got: d}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: draftCurriculumForItem(t, 42)}
	audit := &recordingAuditSink{}

	uc := NewDeleteDisciplineItemUseCase(repo, section, curriculum, audit)
	err := uc.Execute(context.Background(), 42, false, 202)
	require.NoError(t, err)
	assert.Equal(t, 1, repo.deleteCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.deleted", audit.events[0].Action)
}

func TestDeleteDisciplineItem_NonAuthorMethodistDenied(t *testing.T) {
	now := time.Now()
	d := entities.ReconstituteDisciplineItem(202, 11, "T", 18, 18, 0, 36,
		entities.ControlFormZachet, 2, 1, 0, 0, now, now)
	repo := &fakeDisciplineItemDeleteRepo{got: d}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: draftCurriculumForItem(t, 42)}
	audit := &recordingAuditSink{}

	uc := NewDeleteDisciplineItemUseCase(repo, section, curriculum, audit)
	err := uc.Execute(context.Background(), 99, false, 202)
	assert.True(t, errors.Is(err, entities.ErrDisciplineItemScopeForbidden))
	assert.Equal(t, 0, repo.deleteCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.delete_denied", audit.events[0].Action)
	assert.Equal(t, "forbidden", audit.events[0].Fields["reason"])
}

// TestDeleteDisciplineItem_FrozenStatusDenied pins the not_editable
// audit reason on the Delete denial path.
func TestDeleteDisciplineItem_FrozenStatusDenied(t *testing.T) {
	now := time.Now()
	d := entities.ReconstituteDisciplineItem(202, 11, "T", 18, 18, 0, 36,
		entities.ControlFormZachet, 2, 1, 0, 0, now, now)
	repo := &fakeDisciplineItemDeleteRepo{got: d}
	section := &fakeSectionLookup{got: builtSectionForItemTests(t)}
	curriculum := &fakeCurriculumLookup{got: frozenCurriculumForItem(t, entities.StatusPendingApproval, 42)}
	audit := &recordingAuditSink{}

	uc := NewDeleteDisciplineItemUseCase(repo, section, curriculum, audit)
	err := uc.Execute(context.Background(), 42, false, 202)
	assert.True(t, errors.Is(err, entities.ErrCannotEditDisciplineItem))
	assert.Equal(t, 0, repo.deleteCalls)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.delete_denied", audit.events[0].Action)
	assert.Equal(t, "not_editable", audit.events[0].Fields["reason"])
}

func TestDeleteDisciplineItem_ItemNotFound(t *testing.T) {
	repo := &fakeDisciplineItemDeleteRepo{getErr: repositories.ErrDisciplineItemNotFound}
	section := &fakeSectionLookup{}
	curriculum := &fakeCurriculumLookup{}
	audit := &recordingAuditSink{}

	uc := NewDeleteDisciplineItemUseCase(repo, section, curriculum, audit)
	err := uc.Execute(context.Background(), 42, false, 999)
	assert.True(t, errors.Is(err, repositories.ErrDisciplineItemNotFound))
	require.Len(t, audit.events, 1)
	assert.Equal(t, "discipline_item.delete_denied", audit.events[0].Action)
	assert.Equal(t, "not_found", audit.events[0].Fields["reason"])
}
