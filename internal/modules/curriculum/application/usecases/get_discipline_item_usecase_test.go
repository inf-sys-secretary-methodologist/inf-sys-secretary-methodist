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

// ===== Constructor nil-panic =====

func TestNewGetDisciplineItemUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("did not panic on nil repo")
		}
	}()
	NewGetDisciplineItemUseCase(nil)
}

// ===== Path tests =====

func TestGetDisciplineItem_HappyPath(t *testing.T) {
	want := entities.ReconstituteDisciplineItem(202, 11, "T", 36, 36, 0, 72,
		entities.ControlFormExam, 4, 1, 0, 0, time.Now(), time.Now())
	repo := &fakeDisciplineItemGetRepo{got: want}

	uc := NewGetDisciplineItemUseCase(repo)
	got, err := uc.Execute(context.Background(), 202)
	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestGetDisciplineItem_NotFound(t *testing.T) {
	repo := &fakeDisciplineItemGetRepo{getErr: repositories.ErrDisciplineItemNotFound}
	uc := NewGetDisciplineItemUseCase(repo)
	got, err := uc.Execute(context.Background(), 999)
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, repositories.ErrDisciplineItemNotFound))
}
