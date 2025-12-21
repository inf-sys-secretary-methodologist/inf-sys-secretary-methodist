package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/repositories"
)

// MockReportRepository is a mock implementation of ReportRepository
type MockReportRepository struct {
	mock.Mock
}

func (m *MockReportRepository) Create(ctx context.Context, report *entities.Report) error {
	args := m.Called(ctx, report)
	if args.Error(0) == nil {
		report.ID = 1
		report.CreatedAt = time.Now()
		report.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockReportRepository) Save(ctx context.Context, report *entities.Report) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

func (m *MockReportRepository) GetByID(ctx context.Context, id int64) (*entities.Report, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Report), args.Error(1)
}

func (m *MockReportRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockReportRepository) List(ctx context.Context, filter repositories.ReportFilter, limit, offset int) ([]*entities.Report, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Report), args.Error(1)
}

func (m *MockReportRepository) Count(ctx context.Context, filter repositories.ReportFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockReportRepository) GetByAuthor(ctx context.Context, authorID int64, limit, offset int) ([]*entities.Report, error) {
	args := m.Called(ctx, authorID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Report), args.Error(1)
}

func (m *MockReportRepository) GetByStatus(ctx context.Context, status domain.ReportStatus, limit, offset int) ([]*entities.Report, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Report), args.Error(1)
}

func (m *MockReportRepository) GetByReportType(ctx context.Context, reportTypeID int64, limit, offset int) ([]*entities.Report, error) {
	args := m.Called(ctx, reportTypeID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Report), args.Error(1)
}

func (m *MockReportRepository) GetPublicReports(ctx context.Context, limit, offset int) ([]*entities.Report, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Report), args.Error(1)
}

func (m *MockReportRepository) AddAccess(ctx context.Context, access *entities.ReportAccess) error {
	args := m.Called(ctx, access)
	if args.Error(0) == nil {
		access.ID = 1
	}
	return args.Error(0)
}

func (m *MockReportRepository) RemoveAccess(ctx context.Context, reportID, accessID int64) error {
	args := m.Called(ctx, reportID, accessID)
	return args.Error(0)
}

func (m *MockReportRepository) GetAccessByReport(ctx context.Context, reportID int64) ([]*entities.ReportAccess, error) {
	args := m.Called(ctx, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ReportAccess), args.Error(1)
}

func (m *MockReportRepository) HasAccess(ctx context.Context, reportID, userID int64, permission domain.ReportPermission) (bool, error) {
	args := m.Called(ctx, reportID, userID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *MockReportRepository) AddComment(ctx context.Context, comment *entities.ReportComment) error {
	args := m.Called(ctx, comment)
	if args.Error(0) == nil {
		comment.ID = 1
	}
	return args.Error(0)
}

func (m *MockReportRepository) UpdateComment(ctx context.Context, comment *entities.ReportComment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockReportRepository) DeleteComment(ctx context.Context, commentID int64) error {
	args := m.Called(ctx, commentID)
	return args.Error(0)
}

func (m *MockReportRepository) GetCommentsByReport(ctx context.Context, reportID int64) ([]*entities.ReportComment, error) {
	args := m.Called(ctx, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ReportComment), args.Error(1)
}

func (m *MockReportRepository) AddHistory(ctx context.Context, history *entities.ReportHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *MockReportRepository) GetHistoryByReport(ctx context.Context, reportID int64, limit, offset int) ([]*entities.ReportHistory, error) {
	args := m.Called(ctx, reportID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ReportHistory), args.Error(1)
}

func (m *MockReportRepository) CreateGenerationLog(ctx context.Context, log *entities.ReportGenerationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockReportRepository) UpdateGenerationLog(ctx context.Context, log *entities.ReportGenerationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockReportRepository) GetGenerationLogsByReport(ctx context.Context, reportID int64) ([]*entities.ReportGenerationLog, error) {
	args := m.Called(ctx, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ReportGenerationLog), args.Error(1)
}

// MockReportTypeRepository is a mock implementation of ReportTypeRepository
type MockReportTypeRepository struct {
	mock.Mock
}

func (m *MockReportTypeRepository) Create(ctx context.Context, reportType *entities.ReportType) error {
	args := m.Called(ctx, reportType)
	return args.Error(0)
}

func (m *MockReportTypeRepository) Save(ctx context.Context, reportType *entities.ReportType) error {
	args := m.Called(ctx, reportType)
	return args.Error(0)
}

func (m *MockReportTypeRepository) GetByID(ctx context.Context, id int64) (*entities.ReportType, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ReportType), args.Error(1)
}

func (m *MockReportTypeRepository) GetByCode(ctx context.Context, code string) (*entities.ReportType, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ReportType), args.Error(1)
}

func (m *MockReportTypeRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockReportTypeRepository) List(ctx context.Context, filter repositories.ReportTypeFilter, limit, offset int) ([]*entities.ReportType, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ReportType), args.Error(1)
}

func (m *MockReportTypeRepository) Count(ctx context.Context, filter repositories.ReportTypeFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockReportTypeRepository) GetByCategory(ctx context.Context, category domain.ReportCategory) ([]*entities.ReportType, error) {
	args := m.Called(ctx, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ReportType), args.Error(1)
}

func (m *MockReportTypeRepository) GetPeriodic(ctx context.Context) ([]*entities.ReportType, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ReportType), args.Error(1)
}

func (m *MockReportTypeRepository) AddParameter(ctx context.Context, param *entities.ReportParameter) error {
	args := m.Called(ctx, param)
	return args.Error(0)
}

func (m *MockReportTypeRepository) UpdateParameter(ctx context.Context, param *entities.ReportParameter) error {
	args := m.Called(ctx, param)
	return args.Error(0)
}

func (m *MockReportTypeRepository) DeleteParameter(ctx context.Context, paramID int64) error {
	args := m.Called(ctx, paramID)
	return args.Error(0)
}

func (m *MockReportTypeRepository) GetParametersByReportType(ctx context.Context, reportTypeID int64) ([]*entities.ReportParameter, error) {
	args := m.Called(ctx, reportTypeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ReportParameter), args.Error(1)
}

func (m *MockReportTypeRepository) AddTemplate(ctx context.Context, template *entities.ReportTemplate) error {
	args := m.Called(ctx, template)
	return args.Error(0)
}

func (m *MockReportTypeRepository) UpdateTemplate(ctx context.Context, template *entities.ReportTemplate) error {
	args := m.Called(ctx, template)
	return args.Error(0)
}

func (m *MockReportTypeRepository) DeleteTemplate(ctx context.Context, templateID int64) error {
	args := m.Called(ctx, templateID)
	return args.Error(0)
}

func (m *MockReportTypeRepository) GetTemplatesByReportType(ctx context.Context, reportTypeID int64) ([]*entities.ReportTemplate, error) {
	args := m.Called(ctx, reportTypeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ReportTemplate), args.Error(1)
}

func (m *MockReportTypeRepository) GetDefaultTemplate(ctx context.Context, reportTypeID int64) (*entities.ReportTemplate, error) {
	args := m.Called(ctx, reportTypeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ReportTemplate), args.Error(1)
}

func (m *MockReportTypeRepository) SetDefaultTemplate(ctx context.Context, reportTypeID, templateID int64) error {
	args := m.Called(ctx, reportTypeID, templateID)
	return args.Error(0)
}

func (m *MockReportTypeRepository) Subscribe(ctx context.Context, subscription *entities.ReportSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockReportTypeRepository) Unsubscribe(ctx context.Context, reportTypeID, userID int64) error {
	args := m.Called(ctx, reportTypeID, userID)
	return args.Error(0)
}

func (m *MockReportTypeRepository) GetSubscription(ctx context.Context, reportTypeID, userID int64) (*entities.ReportSubscription, error) {
	args := m.Called(ctx, reportTypeID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ReportSubscription), args.Error(1)
}

func (m *MockReportTypeRepository) GetSubscribersByReportType(ctx context.Context, reportTypeID int64) ([]*entities.ReportSubscription, error) {
	args := m.Called(ctx, reportTypeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ReportSubscription), args.Error(1)
}

func (m *MockReportTypeRepository) GetSubscriptionsByUser(ctx context.Context, userID int64) ([]*entities.ReportSubscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ReportSubscription), args.Error(1)
}

func (m *MockReportTypeRepository) UpdateSubscription(ctx context.Context, subscription *entities.ReportSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

// Tests

func TestReportUseCase_Create(t *testing.T) {
	mockReportRepo := new(MockReportRepository)
	mockTypeRepo := new(MockReportTypeRepository)
	usecase := NewReportUseCase(mockReportRepo, mockTypeRepo, nil, nil, nil)

	t.Run("create report successfully", func(t *testing.T) {
		reportType := &entities.ReportType{
			ID:           1,
			Name:         "Test Report",
			Code:         "test-report",
			OutputFormat: domain.OutputFormatPDF,
		}
		mockTypeRepo.On("GetByID", mock.Anything, int64(1)).Return(reportType, nil).Once()
		mockReportRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(nil).Once()
		mockReportRepo.On("AddHistory", mock.Anything, mock.AnythingOfType("*entities.ReportHistory")).Return(nil).Once()

		input := &dto.CreateReportInput{
			ReportTypeID: 1,
			Title:        "Test Report Instance",
		}

		output, err := usecase.Create(context.Background(), 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "Test Report Instance", output.Title)
	})

	t.Run("create report with invalid report type", func(t *testing.T) {
		mockTypeRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, nil).Once()

		input := &dto.CreateReportInput{
			ReportTypeID: 999,
			Title:        "Test Report",
		}

		output, err := usecase.Create(context.Background(), 1, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Equal(t, ErrReportTypeNotFound, err)
	})
}

func TestReportUseCase_GetByID(t *testing.T) {
	mockReportRepo := new(MockReportRepository)
	mockTypeRepo := new(MockReportTypeRepository)
	usecase := NewReportUseCase(mockReportRepo, mockTypeRepo, nil, nil, nil)

	t.Run("get existing report", func(t *testing.T) {
		report := &entities.Report{
			ID:           1,
			ReportTypeID: 1,
			Title:        "Test Report",
			AuthorID:     1,
			Status:       domain.ReportStatusDraft,
		}
		reportType := &entities.ReportType{
			ID:           1,
			Name:         "Test Type",
			Code:         "test-type",
			OutputFormat: domain.OutputFormatPDF,
		}
		mockReportRepo.On("GetByID", mock.Anything, int64(1)).Return(report, nil).Once()
		mockReportRepo.On("HasAccess", mock.Anything, int64(1), int64(1), domain.ReportPermissionRead).Return(false, nil).Once()
		mockTypeRepo.On("GetByID", mock.Anything, int64(1)).Return(reportType, nil).Once()

		output, err := usecase.GetByID(context.Background(), 1, 1)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, int64(1), output.ID)
		assert.Equal(t, "Test Report", output.Title)
	})

	t.Run("get non-existing report", func(t *testing.T) {
		mockReportRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, nil).Once()

		output, err := usecase.GetByID(context.Background(), 999, 1)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Equal(t, ErrReportNotFound, err)
	})
}

func TestReportUseCase_List(t *testing.T) {
	mockReportRepo := new(MockReportRepository)
	mockTypeRepo := new(MockReportTypeRepository)
	usecase := NewReportUseCase(mockReportRepo, mockTypeRepo, nil, nil, nil)

	t.Run("list reports with pagination", func(t *testing.T) {
		reports := []*entities.Report{
			{ID: 1, Title: "Report 1", AuthorID: 1, Status: domain.ReportStatusDraft},
			{ID: 2, Title: "Report 2", AuthorID: 1, Status: domain.ReportStatusReady},
		}
		mockReportRepo.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportFilter")).Return(int64(2), nil).Once()
		mockReportRepo.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportFilter"), 20, 0).Return(reports, nil).Once()

		input := &dto.ReportFilterInput{
			Page:     1,
			PageSize: 20,
		}

		output, err := usecase.List(context.Background(), 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Len(t, output.Reports, 2)
		assert.Equal(t, int64(2), output.Total)
	})

	t.Run("empty result", func(t *testing.T) {
		mockReportRepo.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportFilter")).Return(int64(0), nil).Once()
		mockReportRepo.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportFilter"), 20, 0).Return([]*entities.Report{}, nil).Once()

		input := &dto.ReportFilterInput{
			Page:     1,
			PageSize: 20,
		}

		output, err := usecase.List(context.Background(), 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Len(t, output.Reports, 0)
		assert.Equal(t, int64(0), output.Total)
	})
}

func TestReportUseCase_Delete(t *testing.T) {
	mockReportRepo := new(MockReportRepository)
	mockTypeRepo := new(MockReportTypeRepository)
	usecase := NewReportUseCase(mockReportRepo, mockTypeRepo, nil, nil, nil)

	t.Run("delete existing report", func(t *testing.T) {
		report := &entities.Report{
			ID:       1,
			Title:    "Test Report",
			AuthorID: 1,
			Status:   domain.ReportStatusDraft,
		}
		mockReportRepo.On("GetByID", mock.Anything, int64(1)).Return(report, nil).Once()
		mockReportRepo.On("Delete", mock.Anything, int64(1)).Return(nil).Once()
		mockReportRepo.On("AddHistory", mock.Anything, mock.AnythingOfType("*entities.ReportHistory")).Return(nil).Once()

		err := usecase.Delete(context.Background(), 1, 1)

		assert.NoError(t, err)
	})

	t.Run("delete non-existing report", func(t *testing.T) {
		mockReportRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, nil).Once()

		err := usecase.Delete(context.Background(), 999, 1)

		assert.Error(t, err)
		assert.Equal(t, ErrReportNotFound, err)
	})

	t.Run("delete report by non-author", func(t *testing.T) {
		report := &entities.Report{
			ID:       1,
			Title:    "Test Report",
			AuthorID: 1,
			Status:   domain.ReportStatusDraft,
		}
		mockReportRepo.On("GetByID", mock.Anything, int64(1)).Return(report, nil).Once()

		err := usecase.Delete(context.Background(), 1, 2) // userID=2 is not the author

		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorized, err)
	})
}

func TestReportUseCase_GetReportTypes(t *testing.T) {
	mockReportRepo := new(MockReportRepository)
	mockTypeRepo := new(MockReportTypeRepository)
	usecase := NewReportUseCase(mockReportRepo, mockTypeRepo, nil, nil, nil)

	t.Run("get report types", func(t *testing.T) {
		reportTypes := []*entities.ReportType{
			{ID: 1, Name: "Type 1", Code: "type-1", OutputFormat: domain.OutputFormatPDF},
			{ID: 2, Name: "Type 2", Code: "type-2", OutputFormat: domain.OutputFormatXLSX},
		}
		mockTypeRepo.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportTypeFilter")).Return(int64(2), nil).Once()
		mockTypeRepo.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportTypeFilter"), 20, 0).Return(reportTypes, nil).Once()
		// Mock GetParametersByReportType for each report type
		mockTypeRepo.On("GetParametersByReportType", mock.Anything, int64(1)).Return([]*entities.ReportParameter{}, nil).Once()
		mockTypeRepo.On("GetParametersByReportType", mock.Anything, int64(2)).Return([]*entities.ReportParameter{}, nil).Once()
		// Mock GetTemplatesByReportType for each report type
		mockTypeRepo.On("GetTemplatesByReportType", mock.Anything, int64(1)).Return([]*entities.ReportTemplate{}, nil).Once()
		mockTypeRepo.On("GetTemplatesByReportType", mock.Anything, int64(2)).Return([]*entities.ReportTemplate{}, nil).Once()

		input := &dto.ReportTypeFilterInput{
			Page:     1,
			PageSize: 20,
		}

		output, err := usecase.GetReportTypes(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Len(t, output.ReportTypes, 2)
	})
}
