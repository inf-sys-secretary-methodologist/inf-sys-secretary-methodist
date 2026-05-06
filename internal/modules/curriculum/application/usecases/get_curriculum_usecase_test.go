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

type fakeGetRepo struct {
	id    int64
	out   *entities.Curriculum
	err   error
	calls int
}

func (f *fakeGetRepo) GetByID(_ context.Context, id int64) (*entities.Curriculum, error) {
	f.calls++
	f.id = id
	return f.out, f.err
}

func newDraftEntity(t *testing.T) *entities.Curriculum {
	t.Helper()
	c, err := entities.NewCurriculum(entities.NewCurriculumParams{
		Title:     "ИВТ-2026",
		Code:      "09.03.04-2026",
		Specialty: "Информатика",
		Year:      2026,
		CreatedBy: 42,
		Now:       time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	return c
}

func TestNewGetCurriculumUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewGetCurriculumUseCase(nil) did not panic")
		}
	}()
	NewGetCurriculumUseCase(nil)
}

func TestGetCurriculumUseCase_HappyPathReturnsEntity(t *testing.T) {
	want := newDraftEntity(t)
	want.ID = 7
	repo := &fakeGetRepo{out: want}
	uc := NewGetCurriculumUseCase(repo)

	got, err := uc.Execute(context.Background(), 7)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, int64(7), got.ID)
	assert.Equal(t, 1, repo.calls)
	assert.Equal(t, int64(7), repo.id)
}

func TestGetCurriculumUseCase_NotFoundPropagatesSentinel(t *testing.T) {
	repo := &fakeGetRepo{err: repositories.ErrCurriculumNotFound}
	uc := NewGetCurriculumUseCase(repo)

	got, err := uc.Execute(context.Background(), 999)
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, repositories.ErrCurriculumNotFound))
}

func TestGetCurriculumUseCase_TransportErrorPropagates(t *testing.T) {
	transport := errors.New("conn refused")
	repo := &fakeGetRepo{err: transport}
	uc := NewGetCurriculumUseCase(repo)

	_, err := uc.Execute(context.Background(), 7)
	assert.ErrorIs(t, err, transport)
}
