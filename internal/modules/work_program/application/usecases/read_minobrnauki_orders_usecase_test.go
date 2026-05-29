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
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// fakeReadOrderRepo doubles getMinobrnaukiOrderRepo + listMinobrnaukiOrdersRepo.
type fakeReadOrderRepo struct {
	order       *entities.MinobrnaukiOrder
	getErr      error
	affected    []int64
	affectedErr error
	listResult  repositories.MinobrnaukiOrderListResult
	listErr     error
	lastFilter  repositories.MinobrnaukiOrderListFilter

	getCalls, findCalls, listCalls int
}

func (f *fakeReadOrderRepo) GetByID(_ context.Context, _ int64) (*entities.MinobrnaukiOrder, error) {
	f.getCalls++
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.order, nil
}

func (f *fakeReadOrderRepo) FindAffected(_ context.Context, _ int64) ([]int64, error) {
	f.findCalls++
	if f.affectedErr != nil {
		return nil, f.affectedErr
	}
	return f.affected, nil
}

func (f *fakeReadOrderRepo) List(_ context.Context, filter repositories.MinobrnaukiOrderListFilter) (repositories.MinobrnaukiOrderListResult, error) {
	f.listCalls++
	f.lastFilter = filter
	if f.listErr != nil {
		return repositories.MinobrnaukiOrderListResult{}, f.listErr
	}
	return f.listResult, nil
}

func sampleOrder(id int64) *entities.MinobrnaukiOrder {
	return entities.ReconstituteMinobrnaukiOrder(entities.ReconstituteMinobrnaukiOrderInput{
		ID:          id,
		OrderNumber: "№ 1",
		Title:       "T",
		PublishedAt: time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC),
		ChangeScope: domain.MinobrnaukiOrderChangeScopeMajor,
		UploadedBy:  42,
		CreatedAt:   time.Date(2026, 5, 12, 8, 0, 0, 0, time.UTC),
	})
}

// --- Get ---

func TestGetMinobrnaukiOrderUseCase_HappyPath_ReturnsOrderAndAffected(t *testing.T) {
	repo := &fakeReadOrderRepo{order: sampleOrder(100), affected: []int64{11, 22}}
	uc := NewGetMinobrnaukiOrderUseCase(repo)

	order, affected, err := uc.Execute(context.Background(), "methodist", 100)
	require.NoError(t, err)
	require.NotNil(t, order)
	assert.Equal(t, int64(100), order.ID())
	assert.Equal(t, []int64{11, 22}, affected)
	assert.Equal(t, 1, repo.getCalls)
	assert.Equal(t, 1, repo.findCalls)
}

func TestGetMinobrnaukiOrderUseCase_AllowedRoles(t *testing.T) {
	for _, role := range []string{"methodist", "academic_secretary", "teacher", "system_admin"} {
		t.Run(role, func(t *testing.T) {
			repo := &fakeReadOrderRepo{order: sampleOrder(1)}
			uc := NewGetMinobrnaukiOrderUseCase(repo)
			_, _, err := uc.Execute(context.Background(), role, 1)
			require.NoError(t, err, "role %q must be allowed to view orders", role)
		})
	}
}

func TestGetMinobrnaukiOrderUseCase_StudentDenied(t *testing.T) {
	repo := &fakeReadOrderRepo{order: sampleOrder(1)}
	uc := NewGetMinobrnaukiOrderUseCase(repo)

	_, _, err := uc.Execute(context.Background(), "student", 1)
	assert.True(t, errors.Is(err, domain.ErrMinobrnaukiOrderScopeForbidden),
		"student must be denied, got %v", err)
	assert.Zero(t, repo.getCalls, "repo must not be hit on denied role")
}

func TestGetMinobrnaukiOrderUseCase_NotFoundPropagates(t *testing.T) {
	repo := &fakeReadOrderRepo{getErr: repositories.ErrMinobrnaukiOrderNotFound}
	uc := NewGetMinobrnaukiOrderUseCase(repo)

	_, _, err := uc.Execute(context.Background(), "methodist", 999)
	assert.True(t, errors.Is(err, repositories.ErrMinobrnaukiOrderNotFound), "got %v", err)
	assert.Zero(t, repo.findCalls, "FindAffected must not run when GetByID fails")
}

func TestGetMinobrnaukiOrderUseCase_FindAffectedErrorPropagates(t *testing.T) {
	boom := errors.New("affected query failed")
	repo := &fakeReadOrderRepo{order: sampleOrder(100), affectedErr: boom}
	uc := NewGetMinobrnaukiOrderUseCase(repo)

	_, _, err := uc.Execute(context.Background(), "methodist", 100)
	assert.ErrorIs(t, err, boom)
}

// --- List ---

func TestListMinobrnaukiOrdersUseCase_HappyPath(t *testing.T) {
	repo := &fakeReadOrderRepo{listResult: repositories.MinobrnaukiOrderListResult{
		Items: []repositories.MinobrnaukiOrderListItem{{ID: 1}, {ID: 2}},
		Total: 2,
	}}
	uc := NewListMinobrnaukiOrdersUseCase(repo)

	res, err := uc.Execute(context.Background(), "methodist", repositories.MinobrnaukiOrderListFilter{Limit: 20})
	require.NoError(t, err)
	assert.Equal(t, 2, res.Total)
	assert.Len(t, res.Items, 2)
	assert.Equal(t, 1, repo.listCalls)
}

func TestListMinobrnaukiOrdersUseCase_ClampsPagination(t *testing.T) {
	cases := []struct {
		name                  string
		inLimit, inOffset     int
		wantLimit, wantOffset int
	}{
		{"zero limit defaults to 50", 0, 0, 50, 0},
		{"negative limit defaults to 50", -3, 0, 50, 0},
		{"over-max limit clamps to 200", 500, 0, 200, 0},
		{"negative offset floors to 0", 20, -7, 20, 0},
		{"in-range passes through", 30, 10, 30, 10},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeReadOrderRepo{}
			uc := NewListMinobrnaukiOrdersUseCase(repo)
			_, err := uc.Execute(context.Background(), "methodist", repositories.MinobrnaukiOrderListFilter{
				Limit:  tc.inLimit,
				Offset: tc.inOffset,
			})
			require.NoError(t, err)
			assert.Equal(t, tc.wantLimit, repo.lastFilter.Limit, "limit clamp")
			assert.Equal(t, tc.wantOffset, repo.lastFilter.Offset, "offset clamp")
		})
	}
}

func TestListMinobrnaukiOrdersUseCase_StudentDenied(t *testing.T) {
	repo := &fakeReadOrderRepo{}
	uc := NewListMinobrnaukiOrdersUseCase(repo)

	_, err := uc.Execute(context.Background(), "student", repositories.MinobrnaukiOrderListFilter{Limit: 20})
	assert.True(t, errors.Is(err, domain.ErrMinobrnaukiOrderScopeForbidden), "got %v", err)
	assert.Zero(t, repo.listCalls)
}
