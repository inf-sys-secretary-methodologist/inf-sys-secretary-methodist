package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

type fakeLoadRepo struct {
	created    *entities.TeachingLoad
	updated    *entities.TeachingLoad
	deletedID  int64
	listRet    []*entities.TeachingLoad
	listFilter TeachingLoadFilter
	createErr  error
	createCnt  int
	updateCnt  int
	deleteCnt  int
}

func (f *fakeLoadRepo) Create(_ context.Context, l *entities.TeachingLoad) error {
	f.createCnt++
	f.created = l
	if f.createErr != nil {
		return f.createErr
	}
	l.ID = 77
	return nil
}
func (f *fakeLoadRepo) Update(_ context.Context, l *entities.TeachingLoad) error {
	f.updateCnt++
	f.updated = l
	return nil
}
func (f *fakeLoadRepo) Delete(_ context.Context, id int64) error {
	f.deleteCnt++
	f.deletedID = id
	return nil
}
func (f *fakeLoadRepo) GetByID(_ context.Context, _ int64) (*entities.TeachingLoad, error) {
	return nil, entities.ErrTeachingLoadNotFound
}
func (f *fakeLoadRepo) List(_ context.Context, filter TeachingLoadFilter) ([]*entities.TeachingLoad, error) {
	f.listFilter = filter
	return f.listRet, nil
}

func loadClock() func() time.Time {
	now := time.Date(2026, 2, 2, 9, 0, 0, 0, time.UTC)
	return func() time.Time { return now }
}

func validParams() TeachingLoadParams {
	return TeachingLoadParams{SemesterID: 1, GroupID: 2, DisciplineID: 3, TeacherID: 4, LessonTypeID: 5, PairsPerWeek: 2, WeekType: domain.WeekTypeAll}
}

func TestTeachingLoadUseCase_Create_Valid(t *testing.T) {
	repo := &fakeLoadRepo{}
	uc := NewTeachingLoadUseCase(repo, WithLoadClock(loadClock()))

	load, err := uc.Create(context.Background(), validParams())

	require.NoError(t, err)
	require.NotNil(t, load)
	assert.Equal(t, int64(77), load.ID)
	assert.Equal(t, 2, load.PairsPerWeek)
	assert.Equal(t, 1, repo.createCnt)
	assert.True(t, load.CreatedAt.Equal(loadClock()()))
}

func TestTeachingLoadUseCase_Create_InvalidNotPersisted(t *testing.T) {
	repo := &fakeLoadRepo{}
	uc := NewTeachingLoadUseCase(repo, WithLoadClock(loadClock()))

	p := validParams()
	p.PairsPerWeek = 0

	load, err := uc.Create(context.Background(), p)

	assert.Nil(t, load)
	assert.ErrorIs(t, err, entities.ErrInvalidLoadPairs)
	assert.Equal(t, 0, repo.createCnt, "invalid load must not reach the repository")
}

func TestTeachingLoadUseCase_Create_DuplicatePropagates(t *testing.T) {
	repo := &fakeLoadRepo{createErr: entities.ErrTeachingLoadDuplicate}
	uc := NewTeachingLoadUseCase(repo, WithLoadClock(loadClock()))

	load, err := uc.Create(context.Background(), validParams())

	assert.Nil(t, load)
	assert.ErrorIs(t, err, entities.ErrTeachingLoadDuplicate)
}

func TestTeachingLoadUseCase_Update_Valid(t *testing.T) {
	repo := &fakeLoadRepo{}
	uc := NewTeachingLoadUseCase(repo, WithLoadClock(loadClock()))

	load, err := uc.Update(context.Background(), 7, validParams())

	require.NoError(t, err)
	require.NotNil(t, load)
	assert.Equal(t, int64(7), load.ID)
	require.NotNil(t, repo.updated)
	assert.Equal(t, int64(7), repo.updated.ID)
	assert.True(t, repo.updated.UpdatedAt.Equal(loadClock()()))
	assert.True(t, load.CreatedAt.IsZero(), "update must not stamp CreatedAt")
}

func TestTeachingLoadUseCase_Update_InvalidNotPersisted(t *testing.T) {
	repo := &fakeLoadRepo{}
	uc := NewTeachingLoadUseCase(repo, WithLoadClock(loadClock()))

	p := validParams()
	p.WeekType = domain.WeekType("weekly")

	load, err := uc.Update(context.Background(), 7, p)

	assert.Nil(t, load)
	assert.ErrorIs(t, err, entities.ErrInvalidLoadWeekType)
	assert.Equal(t, 0, repo.updateCnt)
}

func TestTeachingLoadUseCase_Delete(t *testing.T) {
	repo := &fakeLoadRepo{}
	uc := NewTeachingLoadUseCase(repo)

	err := uc.Delete(context.Background(), 9)

	require.NoError(t, err)
	assert.Equal(t, int64(9), repo.deletedID)
}

func TestTeachingLoadUseCase_List_PassesFilter(t *testing.T) {
	want := []*entities.TeachingLoad{{ID: 1}}
	repo := &fakeLoadRepo{listRet: want}
	uc := NewTeachingLoadUseCase(repo)
	sem := int64(3)

	got, err := uc.List(context.Background(), TeachingLoadFilter{SemesterID: &sem})

	require.NoError(t, err)
	assert.Equal(t, want, got)
	require.NotNil(t, repo.listFilter.SemesterID)
	assert.Equal(t, int64(3), *repo.listFilter.SemesterID)
}
