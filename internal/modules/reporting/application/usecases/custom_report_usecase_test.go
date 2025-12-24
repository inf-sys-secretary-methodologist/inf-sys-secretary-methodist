package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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

// Helper function to create a pointer to string
func strPtr(s string) *string {
	return &s
}

// Helper function to create a pointer to bool
func boolPtr(b bool) *bool {
	return &b
}

// Tests

func TestCustomReportUseCase_Create(t *testing.T) {
	mockRepo := new(MockCustomReportRepository)
	mockQueryBuilder := new(MockQueryBuilder)
	usecase := NewCustomReportUseCase(mockRepo, mockQueryBuilder)

	t.Run("create custom report successfully", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.CustomReport")).Return(nil).Once()

		input := dto.CreateCustomReportInput{
			Name:        "Test Report",
			Description: "Test Description",
			DataSource:  "users",
			Fields: []dto.SelectedFieldDTO{
				{
					Field: dto.ReportFieldDTO{
						ID:    "id",
						Name:  "id",
						Label: "ID",
						Type:  "number",
					},
					Order: 1,
				},
			},
			IsPublic: false,
		}

		output, err := usecase.Create(context.Background(), input, 1)

		assert.NoError(t, err)
		assert.Equal(t, "Test Report", output.Name)
		assert.Equal(t, "Test Description", output.Description)
		mockRepo.AssertExpectations(t)
	})

	t.Run("create custom report with invalid data source", func(t *testing.T) {
		input := dto.CreateCustomReportInput{
			Name:       "Test Report",
			DataSource: "invalid_source",
			Fields: []dto.SelectedFieldDTO{
				{
					Field: dto.ReportFieldDTO{ID: "id", Name: "id", Label: "ID", Type: "number"},
					Order: 1,
				},
			},
		}

		output, err := usecase.Create(context.Background(), input, 1)

		assert.Error(t, err)
		assert.Empty(t, output.ID)
		assert.Equal(t, ErrInvalidDataSource, err)
	})

	t.Run("create custom report without fields", func(t *testing.T) {
		input := dto.CreateCustomReportInput{
			Name:       "Test Report",
			DataSource: "users",
			Fields:     []dto.SelectedFieldDTO{},
		}

		output, err := usecase.Create(context.Background(), input, 1)

		assert.Error(t, err)
		assert.Empty(t, output.ID)
		assert.Equal(t, ErrInvalidFields, err)
	})
}

func TestCustomReportUseCase_GetByID(t *testing.T) {
	mockRepo := new(MockCustomReportRepository)
	mockQueryBuilder := new(MockQueryBuilder)
	usecase := NewCustomReportUseCase(mockRepo, mockQueryBuilder)

	reportID := uuid.New()

	t.Run("get existing report by creator", func(t *testing.T) {
		report := &entities.CustomReport{
			ID:         reportID,
			Name:       "Test Report",
			DataSource: entities.DataSourceUsers,
			CreatedBy:  1,
			IsPublic:   false,
			Fields:     []entities.SelectedField{},
			Filters:    []entities.ReportFilterConfig{},
			Groupings:  []entities.ReportGrouping{},
			Sortings:   []entities.ReportSorting{},
		}
		mockRepo.On("GetByID", mock.Anything, reportID).Return(report, nil).Once()

		output, err := usecase.GetByID(context.Background(), reportID, 1)

		assert.NoError(t, err)
		assert.Equal(t, reportID, output.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get public report by non-creator", func(t *testing.T) {
		report := &entities.CustomReport{
			ID:         reportID,
			Name:       "Public Report",
			DataSource: entities.DataSourceUsers,
			CreatedBy:  1,
			IsPublic:   true,
			Fields:     []entities.SelectedField{},
			Filters:    []entities.ReportFilterConfig{},
			Groupings:  []entities.ReportGrouping{},
			Sortings:   []entities.ReportSorting{},
		}
		mockRepo.On("GetByID", mock.Anything, reportID).Return(report, nil).Once()

		output, err := usecase.GetByID(context.Background(), reportID, 2)

		assert.NoError(t, err)
		assert.NotEmpty(t, output.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get private report by non-creator", func(t *testing.T) {
		report := &entities.CustomReport{
			ID:         reportID,
			Name:       "Private Report",
			DataSource: entities.DataSourceUsers,
			CreatedBy:  1,
			IsPublic:   false,
		}
		mockRepo.On("GetByID", mock.Anything, reportID).Return(report, nil).Once()

		output, err := usecase.GetByID(context.Background(), reportID, 2)

		assert.Error(t, err)
		assert.Empty(t, output.ID)
		assert.Equal(t, ErrUnauthorizedAccess, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get non-existing report", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, reportID).Return(nil, nil).Once()

		output, err := usecase.GetByID(context.Background(), reportID, 1)

		assert.Error(t, err)
		assert.Empty(t, output.ID)
		assert.Equal(t, ErrCustomReportNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomReportUseCase_Update(t *testing.T) {
	mockRepo := new(MockCustomReportRepository)
	mockQueryBuilder := new(MockQueryBuilder)
	usecase := NewCustomReportUseCase(mockRepo, mockQueryBuilder)

	reportID := uuid.New()

	t.Run("update report successfully", func(t *testing.T) {
		existingReport := &entities.CustomReport{
			ID:         reportID,
			Name:       "Original Name",
			DataSource: entities.DataSourceUsers,
			CreatedBy:  1,
			IsPublic:   false,
			Fields:     []entities.SelectedField{},
			Filters:    []entities.ReportFilterConfig{},
			Groupings:  []entities.ReportGrouping{},
			Sortings:   []entities.ReportSorting{},
		}
		mockRepo.On("GetByID", mock.Anything, reportID).Return(existingReport, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.CustomReport")).Return(nil).Once()

		input := dto.UpdateCustomReportInput{
			Name:        strPtr("Updated Name"),
			Description: strPtr("Updated Description"),
			DataSource:  strPtr("users"),
			Fields: []dto.SelectedFieldDTO{
				{
					Field: dto.ReportFieldDTO{ID: "id", Name: "id", Label: "ID", Type: "number"},
					Order: 1,
				},
			},
			IsPublic: boolPtr(true),
		}

		output, err := usecase.Update(context.Background(), reportID, input, 1)

		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", output.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("update report by non-creator", func(t *testing.T) {
		existingReport := &entities.CustomReport{
			ID:         reportID,
			Name:       "Original Name",
			DataSource: entities.DataSourceUsers,
			CreatedBy:  1,
			IsPublic:   false,
		}
		mockRepo.On("GetByID", mock.Anything, reportID).Return(existingReport, nil).Once()

		input := dto.UpdateCustomReportInput{
			Name:       strPtr("Updated Name"),
			DataSource: strPtr("users"),
			Fields: []dto.SelectedFieldDTO{
				{
					Field: dto.ReportFieldDTO{ID: "id", Name: "id", Label: "ID", Type: "number"},
					Order: 1,
				},
			},
		}

		output, err := usecase.Update(context.Background(), reportID, input, 2)

		assert.Error(t, err)
		assert.Empty(t, output.ID)
		assert.Equal(t, ErrUnauthorizedAccess, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomReportUseCase_Delete(t *testing.T) {
	mockRepo := new(MockCustomReportRepository)
	mockQueryBuilder := new(MockQueryBuilder)
	usecase := NewCustomReportUseCase(mockRepo, mockQueryBuilder)

	reportID := uuid.New()

	t.Run("delete report successfully", func(t *testing.T) {
		report := &entities.CustomReport{
			ID:        reportID,
			Name:      "Test Report",
			CreatedBy: 1,
		}
		mockRepo.On("GetByID", mock.Anything, reportID).Return(report, nil).Once()
		mockRepo.On("Delete", mock.Anything, reportID).Return(nil).Once()

		err := usecase.Delete(context.Background(), reportID, 1)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("delete report by non-creator", func(t *testing.T) {
		report := &entities.CustomReport{
			ID:        reportID,
			Name:      "Test Report",
			CreatedBy: 1,
		}
		mockRepo.On("GetByID", mock.Anything, reportID).Return(report, nil).Once()

		err := usecase.Delete(context.Background(), reportID, 2)

		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorizedAccess, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("delete non-existing report", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, reportID).Return(nil, nil).Once()

		err := usecase.Delete(context.Background(), reportID, 1)

		assert.Error(t, err)
		assert.Equal(t, ErrCustomReportNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomReportUseCase_List(t *testing.T) {
	mockRepo := new(MockCustomReportRepository)
	mockQueryBuilder := new(MockQueryBuilder)
	usecase := NewCustomReportUseCase(mockRepo, mockQueryBuilder)

	t.Run("list reports with pagination", func(t *testing.T) {
		reports := []*entities.CustomReport{
			{ID: uuid.New(), Name: "Report 1", DataSource: entities.DataSourceUsers, CreatedBy: 1, IsPublic: true, Fields: []entities.SelectedField{}, Filters: []entities.ReportFilterConfig{}, Groupings: []entities.ReportGrouping{}, Sortings: []entities.ReportSorting{}},
			{ID: uuid.New(), Name: "Report 2", DataSource: entities.DataSourceDocuments, CreatedBy: 1, IsPublic: false, Fields: []entities.SelectedField{}, Filters: []entities.ReportFilterConfig{}, Groupings: []entities.ReportGrouping{}, Sortings: []entities.ReportSorting{}},
		}
		mockRepo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(2), nil).Once()
		mockRepo.On("List", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(reports, nil).Once()

		input := dto.CustomReportFilterInput{
			Page:     1,
			PageSize: 20,
		}

		output, err := usecase.List(context.Background(), input, 1)

		assert.NoError(t, err)
		assert.Len(t, output.Reports, 2)
		assert.Equal(t, int64(2), output.Total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("list with empty result", func(t *testing.T) {
		mockRepo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(0), nil).Once()
		mockRepo.On("List", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return([]*entities.CustomReport{}, nil).Once()

		input := dto.CustomReportFilterInput{
			Page:     1,
			PageSize: 20,
		}

		output, err := usecase.List(context.Background(), input, 1)

		assert.NoError(t, err)
		assert.Len(t, output.Reports, 0)
		assert.Equal(t, int64(0), output.Total)
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomReportUseCase_Execute(t *testing.T) {
	mockRepo := new(MockCustomReportRepository)
	mockQueryBuilder := new(MockQueryBuilder)
	usecase := NewCustomReportUseCase(mockRepo, mockQueryBuilder)

	reportID := uuid.New()

	t.Run("execute report successfully", func(t *testing.T) {
		report := &entities.CustomReport{
			ID:         reportID,
			Name:       "Test Report",
			DataSource: entities.DataSourceUsers,
			CreatedBy:  1,
			IsPublic:   false,
			Fields:     []entities.SelectedField{},
			Filters:    []entities.ReportFilterConfig{},
			Groupings:  []entities.ReportGrouping{},
			Sortings:   []entities.ReportSorting{},
		}
		execResult := &entities.ReportExecutionResult{
			Columns:    []entities.ReportColumn{{Key: "id", Label: "ID"}},
			Rows:       []map[string]any{{"id": 1}},
			TotalCount: 1,
			Page:       1,
			PageSize:   10,
			TotalPages: 1,
		}
		mockRepo.On("GetByID", mock.Anything, reportID).Return(report, nil).Once()
		mockQueryBuilder.On("Execute", mock.Anything, report, 1, 10).Return(execResult, nil).Once()

		input := dto.ExecuteReportInput{
			Page:     1,
			PageSize: 10,
		}

		output, err := usecase.Execute(context.Background(), reportID, input, 1)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), output.TotalCount)
		mockRepo.AssertExpectations(t)
		mockQueryBuilder.AssertExpectations(t)
	})

	t.Run("execute private report by non-creator", func(t *testing.T) {
		report := &entities.CustomReport{
			ID:         reportID,
			Name:       "Private Report",
			DataSource: entities.DataSourceUsers,
			CreatedBy:  1,
			IsPublic:   false,
		}
		mockRepo.On("GetByID", mock.Anything, reportID).Return(report, nil).Once()

		input := dto.ExecuteReportInput{
			Page:     1,
			PageSize: 10,
		}

		output, err := usecase.Execute(context.Background(), reportID, input, 2)

		assert.Error(t, err)
		assert.Empty(t, output.Columns)
		assert.Equal(t, ErrUnauthorizedAccess, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomReportUseCase_GetMyReports(t *testing.T) {
	mockRepo := new(MockCustomReportRepository)
	mockQueryBuilder := new(MockQueryBuilder)
	usecase := NewCustomReportUseCase(mockRepo, mockQueryBuilder)

	t.Run("get my reports", func(t *testing.T) {
		reports := []*entities.CustomReport{
			{ID: uuid.New(), Name: "My Report", DataSource: entities.DataSourceUsers, CreatedBy: 1, Fields: []entities.SelectedField{}, Filters: []entities.ReportFilterConfig{}, Groupings: []entities.ReportGrouping{}, Sortings: []entities.ReportSorting{}},
		}
		mockRepo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(1), nil).Once()
		mockRepo.On("List", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(reports, nil).Once()

		output, err := usecase.GetMyReports(context.Background(), 1, 10, 1)

		assert.NoError(t, err)
		assert.Len(t, output.Reports, 1)
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomReportUseCase_GetPublicReports(t *testing.T) {
	mockRepo := new(MockCustomReportRepository)
	mockQueryBuilder := new(MockQueryBuilder)
	usecase := NewCustomReportUseCase(mockRepo, mockQueryBuilder)

	t.Run("get public reports", func(t *testing.T) {
		reports := []*entities.CustomReport{
			{ID: uuid.New(), Name: "Public Report", DataSource: entities.DataSourceUsers, CreatedBy: 1, IsPublic: true, Fields: []entities.SelectedField{}, Filters: []entities.ReportFilterConfig{}, Groupings: []entities.ReportGrouping{}, Sortings: []entities.ReportSorting{}},
		}
		mockRepo.On("Count", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(int64(1), nil).Once()
		mockRepo.On("List", mock.Anything, mock.AnythingOfType("repositories.CustomReportFilter")).Return(reports, nil).Once()

		output, err := usecase.GetPublicReports(context.Background(), 1, 10)

		assert.NoError(t, err)
		assert.Len(t, output.Reports, 1)
		mockRepo.AssertExpectations(t)
	})
}
