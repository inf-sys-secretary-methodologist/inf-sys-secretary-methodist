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

type fakeSectionGetByIDRepo struct {
	got    *entities.Section
	getErr error
	calls  int
	gotID  int64
}

func (f *fakeSectionGetByIDRepo) GetByID(_ context.Context, id int64) (*entities.Section, error) {
	f.calls++
	f.gotID = id
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.got, nil
}

func TestNewGetSectionUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewGetSectionUseCase(nil) did not panic")
		}
	}()
	NewGetSectionUseCase(nil)
}

func TestGetSectionUseCase_HappyPath(t *testing.T) {
	want := entities.ReconstituteSection(101, 7, "T", "d", 0, 3, time.Now(), time.Now())
	repo := &fakeSectionGetByIDRepo{got: want}

	uc := NewGetSectionUseCase(repo)
	got, err := uc.Execute(context.Background(), 101)
	require.NoError(t, err)
	assert.Same(t, want, got)
	assert.Equal(t, int64(101), repo.gotID)
}

func TestGetSectionUseCase_NotFound(t *testing.T) {
	repo := &fakeSectionGetByIDRepo{getErr: repositories.ErrSectionNotFound}

	uc := NewGetSectionUseCase(repo)
	got, err := uc.Execute(context.Background(), 999)
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, repositories.ErrSectionNotFound))
}

func TestGetSectionUseCase_TransportErrorPropagates(t *testing.T) {
	repo := &fakeSectionGetByIDRepo{getErr: errors.New("conn refused")}

	uc := NewGetSectionUseCase(repo)
	_, err := uc.Execute(context.Background(), 101)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conn refused")
}
