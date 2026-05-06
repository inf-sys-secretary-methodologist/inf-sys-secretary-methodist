package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

type fakeListRepo struct {
	got   repositories.CurriculumListFilter
	out   repositories.CurriculumListResult
	err   error
	calls int
}

func (f *fakeListRepo) List(_ context.Context, filter repositories.CurriculumListFilter) (repositories.CurriculumListResult, error) {
	f.calls++
	f.got = filter
	return f.out, f.err
}

func TestNewListCurriculaUseCase_PanicsOnNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewListCurriculaUseCase(nil) did not panic")
		}
	}()
	NewListCurriculaUseCase(nil)
}

func TestListCurriculaUseCase_DefaultLimitWhenZero(t *testing.T) {
	repo := &fakeListRepo{out: repositories.CurriculumListResult{Total: 0}}
	uc := NewListCurriculaUseCase(repo)

	_, err := uc.Execute(context.Background(), ListCurriculaInput{Limit: 0})
	require.NoError(t, err)
	assert.Equal(t, 50, repo.got.Limit, "zero Limit should default to 50")
	assert.Equal(t, 0, repo.got.Offset)
}

func TestListCurriculaUseCase_NegativeLimitDefaultsToFifty(t *testing.T) {
	repo := &fakeListRepo{}
	uc := NewListCurriculaUseCase(repo)

	_, err := uc.Execute(context.Background(), ListCurriculaInput{Limit: -10})
	require.NoError(t, err)
	assert.Equal(t, 50, repo.got.Limit)
}

func TestListCurriculaUseCase_ClampsLimitToTwoHundred(t *testing.T) {
	repo := &fakeListRepo{}
	uc := NewListCurriculaUseCase(repo)

	_, err := uc.Execute(context.Background(), ListCurriculaInput{Limit: 500})
	require.NoError(t, err)
	assert.Equal(t, 200, repo.got.Limit, "limit must clamp to maximum 200")
}

func TestListCurriculaUseCase_NegativeOffsetCoercedToZero(t *testing.T) {
	repo := &fakeListRepo{}
	uc := NewListCurriculaUseCase(repo)

	_, err := uc.Execute(context.Background(), ListCurriculaInput{
		Limit:  20,
		Offset: -1,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, repo.got.Offset)
}

func TestListCurriculaUseCase_ForwardsAllFilters(t *testing.T) {
	repo := &fakeListRepo{}
	uc := NewListCurriculaUseCase(repo)

	status := entities.StatusPendingApproval
	year := 2026
	creator := int64(42)

	_, err := uc.Execute(context.Background(), ListCurriculaInput{
		Status:    &status,
		Year:      &year,
		Specialty: "Информатика",
		CreatedBy: &creator,
		Limit:     25,
		Offset:    50,
	})
	require.NoError(t, err)
	require.NotNil(t, repo.got.Status)
	assert.Equal(t, entities.StatusPendingApproval, *repo.got.Status)
	require.NotNil(t, repo.got.Year)
	assert.Equal(t, 2026, *repo.got.Year)
	assert.Equal(t, "Информатика", repo.got.Specialty)
	require.NotNil(t, repo.got.CreatedBy)
	assert.Equal(t, int64(42), *repo.got.CreatedBy)
	assert.Equal(t, 25, repo.got.Limit)
	assert.Equal(t, 50, repo.got.Offset)
}

func TestListCurriculaUseCase_ReturnsItemsAndTotal(t *testing.T) {
	c := newDraftEntity(t)
	c.ID = 7
	repo := &fakeListRepo{out: repositories.CurriculumListResult{
		Items: []*entities.Curriculum{c},
		Total: 1,
	}}
	uc := NewListCurriculaUseCase(repo)

	page, err := uc.Execute(context.Background(), ListCurriculaInput{Limit: 50})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	assert.Equal(t, int64(7), page.Items[0].ID)
	assert.Equal(t, 1, page.Total)
}

func TestListCurriculaUseCase_RepoErrorPropagates(t *testing.T) {
	transport := errors.New("conn refused")
	repo := &fakeListRepo{err: transport}
	uc := NewListCurriculaUseCase(repo)

	_, err := uc.Execute(context.Background(), ListCurriculaInput{Limit: 50})
	assert.ErrorIs(t, err, transport)
}
