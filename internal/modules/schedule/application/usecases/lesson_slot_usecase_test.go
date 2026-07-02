package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// fakeSlotRepo is an in-memory LessonSlotRepository for use-case tests.
type fakeSlotRepo struct {
	created   *entities.LessonSlot
	updated   *entities.LessonSlot
	deletedID int64
	listRet   []*entities.LessonSlot
	createErr error
	updateErr error
	deleteErr error
	listErr   error
	createCnt int
	updateCnt int
	deleteCnt int
}

func (f *fakeSlotRepo) Create(_ context.Context, slot *entities.LessonSlot) error {
	f.createCnt++
	f.created = slot
	if f.createErr != nil {
		return f.createErr
	}
	slot.ID = 42
	return nil
}

func (f *fakeSlotRepo) Update(_ context.Context, slot *entities.LessonSlot) error {
	f.updateCnt++
	f.updated = slot
	return f.updateErr
}

func (f *fakeSlotRepo) Delete(_ context.Context, id int64) error {
	f.deleteCnt++
	f.deletedID = id
	return f.deleteErr
}

func (f *fakeSlotRepo) GetByID(_ context.Context, _ int64) (*entities.LessonSlot, error) {
	return nil, entities.ErrLessonSlotNotFound
}

func (f *fakeSlotRepo) List(_ context.Context) ([]*entities.LessonSlot, error) {
	return f.listRet, f.listErr
}

func slotFixedClock() func() time.Time {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	return func() time.Time { return now }
}

func TestLessonSlotUseCase_Create_Valid(t *testing.T) {
	repo := &fakeSlotRepo{}
	uc := NewLessonSlotUseCase(repo, WithSlotClock(slotFixedClock()))

	slot, err := uc.Create(context.Background(), 1, "08:30", "10:00")

	require.NoError(t, err)
	require.NotNil(t, slot)
	assert.Equal(t, int64(42), slot.ID)
	assert.Equal(t, 1, slot.Number)
	assert.Equal(t, "08:30", slot.TimeStart)
	assert.Equal(t, 1, repo.createCnt)
	assert.True(t, slot.CreatedAt.Equal(slotFixedClock()()), "clock must be applied")
}

func TestLessonSlotUseCase_Create_InvalidNotPersisted(t *testing.T) {
	repo := &fakeSlotRepo{}
	uc := NewLessonSlotUseCase(repo, WithSlotClock(slotFixedClock()))

	slot, err := uc.Create(context.Background(), 1, "10:00", "08:30") // end before start

	assert.Nil(t, slot)
	assert.ErrorIs(t, err, entities.ErrInvalidSlotTimeRange)
	assert.Equal(t, 0, repo.createCnt, "invalid slot must not reach the repository")
}

func TestLessonSlotUseCase_Create_RepoErrorPropagates(t *testing.T) {
	repo := &fakeSlotRepo{createErr: entities.ErrLessonSlotNumberTaken}
	uc := NewLessonSlotUseCase(repo, WithSlotClock(slotFixedClock()))

	slot, err := uc.Create(context.Background(), 1, "08:30", "10:00")

	assert.Nil(t, slot)
	assert.ErrorIs(t, err, entities.ErrLessonSlotNumberTaken)
}

func TestLessonSlotUseCase_Update_Valid(t *testing.T) {
	repo := &fakeSlotRepo{}
	uc := NewLessonSlotUseCase(repo, WithSlotClock(slotFixedClock()))

	slot, err := uc.Update(context.Background(), 7, 2, "10:10", "11:40")

	require.NoError(t, err)
	require.NotNil(t, slot)
	assert.Equal(t, int64(7), slot.ID, "update must carry the target id")
	assert.Equal(t, 2, slot.Number)
	require.NotNil(t, repo.updated)
	assert.Equal(t, int64(7), repo.updated.ID)
	assert.True(t, repo.updated.UpdatedAt.Equal(slotFixedClock()()))
	// Update does not own created_at (the UPDATE never writes it); the returned
	// entity must not carry a fabricated CreatedAt that could mislead callers.
	assert.True(t, slot.CreatedAt.IsZero(), "update must not stamp CreatedAt")
}

func TestLessonSlotUseCase_Update_InvalidNotPersisted(t *testing.T) {
	repo := &fakeSlotRepo{}
	uc := NewLessonSlotUseCase(repo, WithSlotClock(slotFixedClock()))

	slot, err := uc.Update(context.Background(), 7, 0, "10:10", "11:40") // zero number

	assert.Nil(t, slot)
	assert.ErrorIs(t, err, entities.ErrInvalidSlotNumber)
	assert.Equal(t, 0, repo.updateCnt)
}

func TestLessonSlotUseCase_Delete(t *testing.T) {
	repo := &fakeSlotRepo{}
	uc := NewLessonSlotUseCase(repo, WithSlotClock(slotFixedClock()))

	err := uc.Delete(context.Background(), 9)

	require.NoError(t, err)
	assert.Equal(t, int64(9), repo.deletedID)
	assert.Equal(t, 1, repo.deleteCnt)
}

func TestLessonSlotUseCase_Delete_ErrorPropagates(t *testing.T) {
	repo := &fakeSlotRepo{deleteErr: entities.ErrLessonSlotNotFound}
	uc := NewLessonSlotUseCase(repo, WithSlotClock(slotFixedClock()))

	err := uc.Delete(context.Background(), 9)

	assert.ErrorIs(t, err, entities.ErrLessonSlotNotFound)
}

func TestLessonSlotUseCase_List(t *testing.T) {
	want := []*entities.LessonSlot{{ID: 1, Number: 1}, {ID: 2, Number: 2}}
	repo := &fakeSlotRepo{listRet: want}
	uc := NewLessonSlotUseCase(repo)

	got, err := uc.List(context.Background())

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestLessonSlotUseCase_List_ErrorPropagates(t *testing.T) {
	repo := &fakeSlotRepo{listErr: errors.New("db down")}
	uc := NewLessonSlotUseCase(repo)

	got, err := uc.List(context.Background())

	assert.Nil(t, got)
	require.Error(t, err)
}
