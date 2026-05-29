package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

type fakeRevisionRepo struct {
	programs    map[int64]*entities.WorkProgram
	getErr      map[int64]error
	updateErr   map[int64]error
	updateCalls []int64
}

func (f *fakeRevisionRepo) GetByID(_ context.Context, id int64) (*entities.WorkProgram, error) {
	if err := f.getErr[id]; err != nil {
		return nil, err
	}
	wp, ok := f.programs[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return wp, nil
}

func (f *fakeRevisionRepo) Update(_ context.Context, wp *entities.WorkProgram) error {
	f.updateCalls = append(f.updateCalls, wp.ID())
	return f.updateErr[wp.ID()]
}

type fakeDelegator struct {
	calls []RevisionDelegation
	err   error
}

func (f *fakeDelegator) DelegateRevision(_ context.Context, d RevisionDelegation) error {
	f.calls = append(f.calls, d)
	return f.err
}

func reconWP(id, author int64, status domain.Status) *entities.WorkProgram {
	return entities.ReconstituteWorkProgram(entities.ReconstituteWorkProgramInput{
		ID:                 id,
		DisciplineID:       42,
		SpecialtyCode:      "09.03.01",
		ApplicableFromYear: 2026,
		Title:              "Базы данных",
		Annotation:         "Курс по основам СУБД",
		Status:             status,
		AuthorID:           author,
		Version:            1,
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	})
}

func TestTriggerOrderRevisions_MarksApprovedSkipsRestAndDelegates(t *testing.T) {
	wp100 := reconWP(100, 7, domain.StatusApproved)
	repo := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{
		100: wp100,
		200: reconWP(200, 8, domain.StatusDraft),
		300: reconWP(300, 9, domain.StatusNeedsRevision),
	}}
	del := &fakeDelegator{}
	uc := NewTriggerOrderRevisionsUseCase(repo, del, nil)

	res, err := uc.Execute(context.Background(), 5, 55, "С/123", []int64{100, 200, 300})

	require.NoError(t, err)
	assert.Equal(t, 1, res.Marked, "only the approved program is driven into needs_revision")
	assert.Equal(t, 2, res.Skipped, "draft and already-needs_revision programs are skipped")
	assert.Equal(t, 1, res.Delegated)
	assert.Equal(t, 0, res.Failures)

	assert.Equal(t, []int64{100}, repo.updateCalls, "only the marked program is persisted")
	assert.Equal(t, domain.StatusNeedsRevision, wp100.Status())

	require.Len(t, del.calls, 1)
	assert.Equal(t, RevisionDelegation{
		CreatorID:          5,
		TeacherID:          7,
		WorkProgramID:      100,
		MinobrnaukiOrderID: 55,
		OrderNumber:        "С/123",
	}, del.calls[0], "task is delegated to the program author with order linkage")
}

func TestTriggerOrderRevisions_GetByIDError_CountedAsFailure_DoesNotAbort(t *testing.T) {
	repo := &fakeRevisionRepo{
		programs: map[int64]*entities.WorkProgram{200: reconWP(200, 8, domain.StatusApproved)},
		getErr:   map[int64]error{100: errors.New("db down")},
	}
	del := &fakeDelegator{}
	uc := NewTriggerOrderRevisionsUseCase(repo, del, nil)

	res, err := uc.Execute(context.Background(), 5, 55, "С/123", []int64{100, 200})

	require.NoError(t, err)
	assert.Equal(t, 1, res.Failures, "the load error is counted but does not abort the batch")
	assert.Equal(t, 1, res.Marked, "the second program is still processed")
	assert.Equal(t, 1, res.Delegated)
	assert.Equal(t, []int64{200}, repo.updateCalls)
}

func TestTriggerOrderRevisions_UpdateError_NotMarkedNotDelegated(t *testing.T) {
	repo := &fakeRevisionRepo{
		programs:  map[int64]*entities.WorkProgram{100: reconWP(100, 7, domain.StatusApproved)},
		updateErr: map[int64]error{100: errors.New("version conflict")},
	}
	del := &fakeDelegator{}
	uc := NewTriggerOrderRevisionsUseCase(repo, del, nil)

	res, err := uc.Execute(context.Background(), 5, 55, "С/123", []int64{100})

	require.NoError(t, err)
	assert.Equal(t, 0, res.Marked, "a failed Update is not counted as marked")
	assert.Equal(t, 0, res.Delegated, "no task is delegated when persistence failed")
	assert.Equal(t, 1, res.Failures)
	assert.Equal(t, []int64{100}, repo.updateCalls, "Update was attempted")
	assert.Empty(t, del.calls, "delegation is skipped when Update failed")
}

func TestTriggerOrderRevisions_DelegateError_ProgramStillMarked(t *testing.T) {
	wp100 := reconWP(100, 7, domain.StatusApproved)
	repo := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{100: wp100}}
	del := &fakeDelegator{err: errors.New("tasks unavailable")}
	uc := NewTriggerOrderRevisionsUseCase(repo, del, nil)

	res, err := uc.Execute(context.Background(), 5, 55, "С/123", []int64{100})

	require.NoError(t, err)
	assert.Equal(t, 1, res.Marked, "the program is marked + persisted even if delegation later fails")
	assert.Equal(t, 0, res.Delegated)
	assert.Equal(t, 1, res.Failures)
	assert.Equal(t, domain.StatusNeedsRevision, wp100.Status())
	require.Len(t, del.calls, 1, "delegation was attempted")
}

func TestTriggerOrderRevisions_EmptyAffected_NoOp(t *testing.T) {
	repo := &fakeRevisionRepo{programs: map[int64]*entities.WorkProgram{}}
	del := &fakeDelegator{}
	uc := NewTriggerOrderRevisionsUseCase(repo, del, nil)

	res, err := uc.Execute(context.Background(), 5, 55, "С/123", nil)

	require.NoError(t, err)
	assert.Equal(t, TriggerOrderRevisionsResult{}, res)
	assert.Empty(t, repo.updateCalls)
	assert.Empty(t, del.calls)
}
