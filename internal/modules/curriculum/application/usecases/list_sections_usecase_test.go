package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
)

type fakeSectionListRepo struct {
	got     []*entities.Section
	listErr error
	calls   int
	gotID   int64
}

func (f *fakeSectionListRepo) ListByCurriculumID(_ context.Context, curriculumID int64) ([]*entities.Section, error) {
	f.calls++
	f.gotID = curriculumID
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.got, nil
}

func TestNewListSectionsByCurriculumUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewListSectionsByCurriculumUseCase(nil) did not panic")
		}
	}()
	NewListSectionsByCurriculumUseCase(nil)
}

func TestListSectionsByCurriculumUseCase_HappyPath(t *testing.T) {
	now := time.Now()
	items := []*entities.Section{
		entities.ReconstituteSection(101, 7, "Базовая", "", 0, 0, now, now),
		entities.ReconstituteSection(102, 7, "Вариативная", "", 1, 0, now, now),
	}
	repo := &fakeSectionListRepo{got: items}

	uc := NewListSectionsByCurriculumUseCase(repo)
	got, err := uc.Execute(context.Background(), 7)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, int64(7), repo.gotID)
	assert.Equal(t, int64(101), got[0].ID)
	assert.Equal(t, int64(102), got[1].ID)
}

func TestListSectionsByCurriculumUseCase_EmptyResult(t *testing.T) {
	repo := &fakeSectionListRepo{got: nil}

	uc := NewListSectionsByCurriculumUseCase(repo)
	got, err := uc.Execute(context.Background(), 99)
	require.NoError(t, err, "empty list is not an error condition")
	assert.Len(t, got, 0)
}

func TestListSectionsByCurriculumUseCase_TransportErrorPropagates(t *testing.T) {
	repo := &fakeSectionListRepo{listErr: errors.New("conn refused")}

	uc := NewListSectionsByCurriculumUseCase(repo)
	_, err := uc.Execute(context.Background(), 7)
	require.Error(t, err)
}
