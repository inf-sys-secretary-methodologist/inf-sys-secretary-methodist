package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/repositories"
)

// MockCustomReportRepository is a mock implementation of CustomReportRepository
type MockCustomReportRepository struct {
	mock.Mock
}

func (m *MockCustomReportRepository) Create(ctx context.Context, report *entities.CustomReport) error {
	args := m.Called(ctx, report)
	if args.Error(0) == nil {
		report.ID = uuid.New()
		report.CreatedAt = time.Now()
		report.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockCustomReportRepository) Update(ctx context.Context, report *entities.CustomReport) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

func (m *MockCustomReportRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.CustomReport, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.CustomReport), args.Error(1)
}

func (m *MockCustomReportRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCustomReportRepository) List(ctx context.Context, filter repositories.CustomReportFilter) ([]*entities.CustomReport, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.CustomReport), args.Error(1)
}

func (m *MockCustomReportRepository) Count(ctx context.Context, filter repositories.CustomReportFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCustomReportRepository) GetByCreator(ctx context.Context, creatorID int64, page, pageSize int) ([]*entities.CustomReport, error) {
	args := m.Called(ctx, creatorID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.CustomReport), args.Error(1)
}

func (m *MockCustomReportRepository) GetPublicReports(ctx context.Context, page, pageSize int) ([]*entities.CustomReport, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.CustomReport), args.Error(1)
}

// MockQueryBuilder is a mock implementation of QueryBuilder interface
type MockQueryBuilder struct {
	mock.Mock
}

func (m *MockQueryBuilder) Execute(ctx context.Context, report *entities.CustomReport, page, pageSize int) (*entities.ReportExecutionResult, error) {
	args := m.Called(ctx, report, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ReportExecutionResult), args.Error(1)
}

func (m *MockQueryBuilder) Export(result *entities.ReportExecutionResult, options entities.ExportOptions, reportName string) ([]byte, string, error) {
	args := m.Called(result, options, reportName)
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}

func (m *MockQueryBuilder) GetAvailableFields(dataSource entities.DataSourceType) []entities.ReportField {
	args := m.Called(dataSource)
	return args.Get(0).([]entities.ReportField)
}

// Helpers
func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

func newCustomUC() (*CustomReportUseCase, *MockCustomReportRepository, *MockQueryBuilder) {
	repo := new(MockCustomReportRepository)
	qb := new(MockQueryBuilder)
	uc := NewCustomReportUseCase(repo, qb)
	return uc, repo, qb
}

func validFields() []dto.SelectedFieldDTO {
	return []dto.SelectedFieldDTO{
		{Field: dto.ReportFieldDTO{ID: "id", Name: "id", Label: "ID", Type: "number"}, Order: 1},
	}
}

func newCustomReport(id uuid.UUID, createdBy int64, isPublic bool) *entities.CustomReport {
	return &entities.CustomReport{
		ID: id, Name: "R", DataSource: entities.DataSourceUsers,
		CreatedBy: createdBy, IsPublic: isPublic,
		Fields: []entities.SelectedField{}, Filters: []entities.ReportFilterConfig{},
		Groupings: []entities.ReportGrouping{}, Sortings: []entities.ReportSorting{},
	}
}

// =============================================================================
// Create
// =============================================================================

func TestCustomReportUseCase_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.CustomReport")).Return(nil).Once()
		out, err := uc.Create(context.Background(), dto.CreateCustomReportInput{
			Name: "R", Description: "D", DataSource: "users", Fields: validFields(),
		}, 1)
		require.NoError(t, err)
		assert.Equal(t, "R", out.Name)
	})

	t.Run("invalid data source", func(t *testing.T) {
		uc, _, _ := newCustomUC()
		_, err := uc.Create(context.Background(), dto.CreateCustomReportInput{
			Name: "R", DataSource: "bad", Fields: validFields(),
		}, 1)
		assert.ErrorIs(t, err, ErrInvalidDataSource)
	})

	t.Run("no fields", func(t *testing.T) {
		uc, _, _ := newCustomUC()
		_, err := uc.Create(context.Background(), dto.CreateCustomReportInput{
			Name: "R", DataSource: "users", Fields: []dto.SelectedFieldDTO{},
		}, 1)
		assert.ErrorIs(t, err, ErrInvalidFields)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.CustomReport")).Return(errors.New("db")).Once()
		_, err := uc.Create(context.Background(), dto.CreateCustomReportInput{
			Name: "R", DataSource: "users", Fields: validFields(),
		}, 1)
		assert.Error(t, err)
	})

	t.Run("with filters groupings sortings", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.CustomReport")).Return(nil).Once()
		out, err := uc.Create(context.Background(), dto.CreateCustomReportInput{
			Name: "R", DataSource: "users", Fields: validFields(), IsPublic: true,
			Filters:   []dto.ReportFilterDTO{{ID: "f1", Field: dto.ReportFieldDTO{ID: "id", Name: "id", Label: "ID", Type: "number"}, Operator: "equals", Value: 1}},
			Groupings: []dto.ReportGroupingDTO{{Field: dto.ReportFieldDTO{ID: "id", Name: "id", Label: "ID", Type: "number"}, Order: "asc"}},
			Sortings:  []dto.ReportSortingDTO{{Field: dto.ReportFieldDTO{ID: "id", Name: "id", Label: "ID", Type: "number"}, Order: "desc"}},
		}, 1)
		require.NoError(t, err)
		assert.Equal(t, "R", out.Name)
	})
}

// =============================================================================
// GetByID
// =============================================================================

func TestCustomReportUseCase_GetByID(t *testing.T) {
	id := uuid.New()

	t.Run("creator access", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(newCustomReport(id, 1, false), nil).Once()
		out, err := uc.GetByID(context.Background(), id, 1)
		require.NoError(t, err)
		assert.Equal(t, id, out.ID)
	})

	t.Run("public access by non-creator", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(newCustomReport(id, 1, true), nil).Once()
		out, err := uc.GetByID(context.Background(), id, 2)
		require.NoError(t, err)
		assert.Equal(t, id, out.ID)
	})

	t.Run("private unauthorized", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(newCustomReport(id, 1, false), nil).Once()
		_, err := uc.GetByID(context.Background(), id, 2)
		assert.ErrorIs(t, err, ErrUnauthorizedAccess)
	})

	t.Run("not found", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(nil, nil).Once()
		_, err := uc.GetByID(context.Background(), id, 1)
		assert.ErrorIs(t, err, ErrCustomReportNotFound)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("db")).Once()
		_, err := uc.GetByID(context.Background(), id, 1)
		assert.Error(t, err)
	})
}

// =============================================================================
// Update
// =============================================================================

func TestCustomReportUseCase_Update(t *testing.T) {
	id := uuid.New()

	t.Run("success all fields", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(newCustomReport(id, 1, false), nil).Once()
		repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.CustomReport")).Return(nil).Once()

		out, err := uc.Update(context.Background(), id, dto.UpdateCustomReportInput{
			Name: strPtr("New"), Description: strPtr("Desc"), DataSource: strPtr("users"),
			Fields: validFields(), IsPublic: boolPtr(true),
			Filters:   []dto.ReportFilterDTO{{ID: "f1", Field: dto.ReportFieldDTO{ID: "id", Name: "id", Label: "ID", Type: "number"}, Operator: "equals", Value: 1}},
			Groupings: []dto.ReportGroupingDTO{{Field: dto.ReportFieldDTO{ID: "id", Name: "id", Label: "ID", Type: "number"}, Order: "asc"}},
			Sortings:  []dto.ReportSortingDTO{{Field: dto.ReportFieldDTO{ID: "id", Name: "id", Label: "ID", Type: "number"}, Order: "desc"}},
		}, 1)
		require.NoError(t, err)
		assert.Equal(t, "New", out.Name)
	})

	t.Run("non-creator", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(newCustomReport(id, 1, false), nil).Once()
		_, err := uc.Update(context.Background(), id, dto.UpdateCustomReportInput{Name: strPtr("X")}, 2)
		assert.ErrorIs(t, err, ErrUnauthorizedAccess)
	})

	t.Run("not found", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(nil, nil).Once()
		_, err := uc.Update(context.Background(), id, dto.UpdateCustomReportInput{Name: strPtr("X")}, 1)
		assert.ErrorIs(t, err, ErrCustomReportNotFound)
	})

	t.Run("repo error on get", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("db")).Once()
		_, err := uc.Update(context.Background(), id, dto.UpdateCustomReportInput{Name: strPtr("X")}, 1)
		assert.Error(t, err)
	})

	t.Run("invalid data source", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(newCustomReport(id, 1, false), nil).Once()
		_, err := uc.Update(context.Background(), id, dto.UpdateCustomReportInput{DataSource: strPtr("bad")}, 1)
		assert.ErrorIs(t, err, ErrInvalidDataSource)
	})

	t.Run("empty fields", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(newCustomReport(id, 1, false), nil).Once()
		_, err := uc.Update(context.Background(), id, dto.UpdateCustomReportInput{Fields: []dto.SelectedFieldDTO{}}, 1)
		assert.ErrorIs(t, err, ErrInvalidFields)
	})

	t.Run("update repo error", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(newCustomReport(id, 1, false), nil).Once()
		repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.CustomReport")).Return(errors.New("db")).Once()
		_, err := uc.Update(context.Background(), id, dto.UpdateCustomReportInput{Name: strPtr("X")}, 1)
		assert.Error(t, err)
	})

	t.Run("partial update name only", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(newCustomReport(id, 1, false), nil).Once()
		repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.CustomReport")).Return(nil).Once()

		out, err := uc.Update(context.Background(), id, dto.UpdateCustomReportInput{Name: strPtr("New")}, 1)
		require.NoError(t, err)
		assert.Equal(t, "New", out.Name)
	})
}

// =============================================================================
// Delete
// =============================================================================

func TestCustomReportUseCase_Delete(t *testing.T) {
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(newCustomReport(id, 1, false), nil).Once()
		repo.On("Delete", mock.Anything, id).Return(nil).Once()
		assert.NoError(t, uc.Delete(context.Background(), id, 1))
	})

	t.Run("non-creator", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(newCustomReport(id, 1, false), nil).Once()
		assert.ErrorIs(t, uc.Delete(context.Background(), id, 2), ErrUnauthorizedAccess)
	})

	t.Run("not found", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(nil, nil).Once()
		assert.ErrorIs(t, uc.Delete(context.Background(), id, 1), ErrCustomReportNotFound)
	})

	t.Run("repo error on get", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("db")).Once()
		assert.Error(t, uc.Delete(context.Background(), id, 1))
	})

	t.Run("delete repo error", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(newCustomReport(id, 1, false), nil).Once()
		repo.On("Delete", mock.Anything, id).Return(errors.New("db")).Once()
		assert.Error(t, uc.Delete(context.Background(), id, 1))
	})
}

// =============================================================================
// List
// =============================================================================

func TestCustomReportUseCase_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		reports := []*entities.CustomReport{
			newCustomReport(uuid.New(), 1, true),
			newCustomReport(uuid.New(), 1, false),
		}
		repo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(2), nil).Once()
		repo.On("List", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(reports, nil).Once()

		out, err := uc.List(context.Background(), dto.CustomReportFilterInput{Page: 1, PageSize: 20}, 1)
		require.NoError(t, err)
		assert.Len(t, out.Reports, 2)
	})

	t.Run("filters accessible reports", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		reports := []*entities.CustomReport{
			newCustomReport(uuid.New(), 1, false), // own
			newCustomReport(uuid.New(), 2, true),  // public
			newCustomReport(uuid.New(), 2, false), // other's private - filtered out
		}
		repo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(3), nil).Once()
		repo.On("List", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(reports, nil).Once()

		out, err := uc.List(context.Background(), dto.CustomReportFilterInput{Page: 1, PageSize: 20}, 1)
		require.NoError(t, err)
		assert.Len(t, out.Reports, 2)
	})

	t.Run("default pagination", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(0), nil).Once()
		repo.On("List", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return([]*entities.CustomReport{}, nil).Once()

		out, err := uc.List(context.Background(), dto.CustomReportFilterInput{Page: 0, PageSize: 0}, 1)
		require.NoError(t, err)
		assert.Equal(t, 1, out.Page)
		assert.Equal(t, 10, out.PageSize)
	})

	t.Run("count error", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(0), errors.New("db")).Once()
		_, err := uc.List(context.Background(), dto.CustomReportFilterInput{Page: 1, PageSize: 10}, 1)
		assert.Error(t, err)
	})

	t.Run("list error", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(1), nil).Once()
		repo.On("List", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(nil, errors.New("db")).Once()
		_, err := uc.List(context.Background(), dto.CustomReportFilterInput{Page: 1, PageSize: 10}, 1)
		assert.Error(t, err)
	})

	t.Run("empty", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(0), nil).Once()
		repo.On("List", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return([]*entities.CustomReport{}, nil).Once()

		out, err := uc.List(context.Background(), dto.CustomReportFilterInput{Page: 1, PageSize: 20}, 1)
		require.NoError(t, err)
		assert.Len(t, out.Reports, 0)
	})
}

// =============================================================================
// Execute
// =============================================================================

func TestCustomReportUseCase_Execute(t *testing.T) {
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		uc, repo, qb := newCustomUC()
		r := newCustomReport(id, 1, false)
		result := &entities.ReportExecutionResult{
			Columns:    []entities.ReportColumn{{Key: "id", Label: "ID"}},
			Rows:       []map[string]any{{"id": 1}},
			TotalCount: 1, Page: 1, PageSize: 10, TotalPages: 1,
		}
		repo.On("GetByID", mock.Anything, id).Return(r, nil).Once()
		qb.On("Execute", mock.Anything, r, 1, 10).Return(result, nil).Once()

		out, err := uc.Execute(context.Background(), id, dto.ExecuteReportInput{Page: 1, PageSize: 10}, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), out.TotalCount)
		assert.Len(t, out.Columns, 1)
	})

	t.Run("not found", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(nil, nil).Once()
		_, err := uc.Execute(context.Background(), id, dto.ExecuteReportInput{Page: 1, PageSize: 10}, 1)
		assert.ErrorIs(t, err, ErrCustomReportNotFound)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("db")).Once()
		_, err := uc.Execute(context.Background(), id, dto.ExecuteReportInput{Page: 1, PageSize: 10}, 1)
		assert.Error(t, err)
	})

	t.Run("unauthorized", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(newCustomReport(id, 1, false), nil).Once()
		_, err := uc.Execute(context.Background(), id, dto.ExecuteReportInput{Page: 1, PageSize: 10}, 2)
		assert.ErrorIs(t, err, ErrUnauthorizedAccess)
	})

	t.Run("public report by non-creator", func(t *testing.T) {
		uc, repo, qb := newCustomUC()
		r := newCustomReport(id, 1, true)
		result := &entities.ReportExecutionResult{
			Columns: []entities.ReportColumn{}, Rows: []map[string]any{},
			TotalCount: 0, Page: 1, PageSize: 50, TotalPages: 0,
		}
		repo.On("GetByID", mock.Anything, id).Return(r, nil).Once()
		qb.On("Execute", mock.Anything, r, 1, 50).Return(result, nil).Once()

		out, err := uc.Execute(context.Background(), id, dto.ExecuteReportInput{Page: 0, PageSize: 0}, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(0), out.TotalCount)
	})

	t.Run("default pagination", func(t *testing.T) {
		uc, repo, qb := newCustomUC()
		r := newCustomReport(id, 1, false)
		result := &entities.ReportExecutionResult{
			Columns: []entities.ReportColumn{}, Rows: []map[string]any{},
			TotalCount: 0, Page: 1, PageSize: 50, TotalPages: 0,
		}
		repo.On("GetByID", mock.Anything, id).Return(r, nil).Once()
		qb.On("Execute", mock.Anything, r, 1, 50).Return(result, nil).Once()

		_, err := uc.Execute(context.Background(), id, dto.ExecuteReportInput{Page: 0, PageSize: 0}, 1)
		require.NoError(t, err)
	})

	t.Run("pageSize capped at 1000", func(t *testing.T) {
		uc, repo, qb := newCustomUC()
		r := newCustomReport(id, 1, false)
		result := &entities.ReportExecutionResult{
			Columns: []entities.ReportColumn{}, Rows: []map[string]any{},
			TotalCount: 0, Page: 1, PageSize: 1000, TotalPages: 0,
		}
		repo.On("GetByID", mock.Anything, id).Return(r, nil).Once()
		qb.On("Execute", mock.Anything, r, 1, 1000).Return(result, nil).Once()

		_, err := uc.Execute(context.Background(), id, dto.ExecuteReportInput{Page: 1, PageSize: 5000}, 1)
		require.NoError(t, err)
	})

	t.Run("query error", func(t *testing.T) {
		uc, repo, qb := newCustomUC()
		r := newCustomReport(id, 1, false)
		repo.On("GetByID", mock.Anything, id).Return(r, nil).Once()
		qb.On("Execute", mock.Anything, r, 1, 10).Return(nil, errors.New("query err")).Once()

		_, err := uc.Execute(context.Background(), id, dto.ExecuteReportInput{Page: 1, PageSize: 10}, 1)
		assert.Error(t, err)
	})
}

// =============================================================================
// Export
// =============================================================================

func TestCustomReportUseCase_Export(t *testing.T) {
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		uc, repo, qb := newCustomUC()
		r := newCustomReport(id, 1, false)
		result := &entities.ReportExecutionResult{
			Columns:    []entities.ReportColumn{{Key: "id", Label: "ID"}},
			Rows:       []map[string]any{{"id": 1}},
			TotalCount: 1, Page: 1, PageSize: 10000, TotalPages: 1,
		}
		repo.On("GetByID", mock.Anything, id).Return(r, nil).Once()
		qb.On("Execute", mock.Anything, r, 1, 10000).Return(result, nil).Once()
		qb.On("Export", result, mock.AnythingOfType("entities.ExportOptions"), "R").Return([]byte("data"), "report.csv", nil).Once()

		data, filename, err := uc.Export(context.Background(), id, dto.ExportReportInput{
			Format: "csv", IncludeHeaders: true,
		}, 1)
		require.NoError(t, err)
		assert.Equal(t, []byte("data"), data)
		assert.Equal(t, "report.csv", filename)
	})

	t.Run("not found", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(nil, nil).Once()
		_, _, err := uc.Export(context.Background(), id, dto.ExportReportInput{Format: "csv"}, 1)
		assert.ErrorIs(t, err, ErrCustomReportNotFound)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("db")).Once()
		_, _, err := uc.Export(context.Background(), id, dto.ExportReportInput{Format: "csv"}, 1)
		assert.Error(t, err)
	})

	t.Run("unauthorized", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("GetByID", mock.Anything, id).Return(newCustomReport(id, 1, false), nil).Once()
		_, _, err := uc.Export(context.Background(), id, dto.ExportReportInput{Format: "csv"}, 2)
		assert.ErrorIs(t, err, ErrUnauthorizedAccess)
	})

	t.Run("public report by non-creator", func(t *testing.T) {
		uc, repo, qb := newCustomUC()
		r := newCustomReport(id, 1, true)
		result := &entities.ReportExecutionResult{
			Columns: []entities.ReportColumn{}, Rows: []map[string]any{},
			TotalCount: 0, Page: 1, PageSize: 10000, TotalPages: 0,
		}
		repo.On("GetByID", mock.Anything, id).Return(r, nil).Once()
		qb.On("Execute", mock.Anything, r, 1, 10000).Return(result, nil).Once()
		qb.On("Export", result, mock.AnythingOfType("entities.ExportOptions"), "R").Return([]byte("data"), "report.csv", nil).Once()

		data, _, err := uc.Export(context.Background(), id, dto.ExportReportInput{Format: "csv"}, 2)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("execute error", func(t *testing.T) {
		uc, repo, qb := newCustomUC()
		r := newCustomReport(id, 1, false)
		repo.On("GetByID", mock.Anything, id).Return(r, nil).Once()
		qb.On("Execute", mock.Anything, r, 1, 10000).Return(nil, errors.New("query err")).Once()
		_, _, err := uc.Export(context.Background(), id, dto.ExportReportInput{Format: "csv"}, 1)
		assert.Error(t, err)
	})

	t.Run("export error", func(t *testing.T) {
		uc, repo, qb := newCustomUC()
		r := newCustomReport(id, 1, false)
		result := &entities.ReportExecutionResult{
			Columns: []entities.ReportColumn{}, Rows: []map[string]any{},
			TotalCount: 0, Page: 1, PageSize: 10000, TotalPages: 0,
		}
		repo.On("GetByID", mock.Anything, id).Return(r, nil).Once()
		qb.On("Execute", mock.Anything, r, 1, 10000).Return(result, nil).Once()
		qb.On("Export", result, mock.AnythingOfType("entities.ExportOptions"), "R").Return([]byte{}, "", errors.New("export err")).Once()
		_, _, err := uc.Export(context.Background(), id, dto.ExportReportInput{Format: "csv"}, 1)
		assert.Error(t, err)
	})

	t.Run("with pdf options", func(t *testing.T) {
		uc, repo, qb := newCustomUC()
		r := newCustomReport(id, 1, false)
		result := &entities.ReportExecutionResult{
			Columns: []entities.ReportColumn{}, Rows: []map[string]any{},
			TotalCount: 0, Page: 1, PageSize: 10000, TotalPages: 0,
		}
		repo.On("GetByID", mock.Anything, id).Return(r, nil).Once()
		qb.On("Execute", mock.Anything, r, 1, 10000).Return(result, nil).Once()
		qb.On("Export", result, mock.AnythingOfType("entities.ExportOptions"), "R").Return([]byte("pdf"), "report.pdf", nil).Once()

		data, filename, err := uc.Export(context.Background(), id, dto.ExportReportInput{
			Format: "pdf", IncludeHeaders: true, PageSize: "A4", Orientation: "landscape",
		}, 1)
		require.NoError(t, err)
		assert.Equal(t, []byte("pdf"), data)
		assert.Equal(t, "report.pdf", filename)
	})
}

// =============================================================================
// GetMyReports
// =============================================================================

func TestCustomReportUseCase_GetMyReports(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		reports := []*entities.CustomReport{newCustomReport(uuid.New(), 1, false)}
		repo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(1), nil).Once()
		repo.On("List", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(reports, nil).Once()

		out, err := uc.GetMyReports(context.Background(), 1, 10, 1)
		require.NoError(t, err)
		assert.Len(t, out.Reports, 1)
	})

	t.Run("default pagination", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(0), nil).Once()
		repo.On("List", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return([]*entities.CustomReport{}, nil).Once()

		out, err := uc.GetMyReports(context.Background(), 0, 0, 1)
		require.NoError(t, err)
		assert.Equal(t, 1, out.Page)
		assert.Equal(t, 10, out.PageSize)
	})

	t.Run("count error", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(0), errors.New("db")).Once()
		_, err := uc.GetMyReports(context.Background(), 1, 10, 1)
		assert.Error(t, err)
	})

	t.Run("list error", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(1), nil).Once()
		repo.On("List", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(nil, errors.New("db")).Once()
		_, err := uc.GetMyReports(context.Background(), 1, 10, 1)
		assert.Error(t, err)
	})
}

// =============================================================================
// GetPublicReports
// =============================================================================

func TestCustomReportUseCase_GetPublicReports(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		reports := []*entities.CustomReport{newCustomReport(uuid.New(), 1, true)}
		repo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(1), nil).Once()
		repo.On("List", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(reports, nil).Once()

		out, err := uc.GetPublicReports(context.Background(), 1, 10)
		require.NoError(t, err)
		assert.Len(t, out.Reports, 1)
	})

	t.Run("default pagination", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(0), nil).Once()
		repo.On("List", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return([]*entities.CustomReport{}, nil).Once()

		out, err := uc.GetPublicReports(context.Background(), 0, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, out.Page)
		assert.Equal(t, 10, out.PageSize)
	})

	t.Run("count error", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(0), errors.New("db")).Once()
		_, err := uc.GetPublicReports(context.Background(), 1, 10)
		assert.Error(t, err)
	})

	t.Run("list error", func(t *testing.T) {
		uc, repo, _ := newCustomUC()
		repo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(1), nil).Once()
		repo.On("List", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(nil, errors.New("db")).Once()
		_, err := uc.GetPublicReports(context.Background(), 1, 10)
		assert.Error(t, err)
	})
}
