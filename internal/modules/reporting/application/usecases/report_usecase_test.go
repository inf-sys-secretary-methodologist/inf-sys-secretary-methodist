package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/repositories"
)

const updatedTitle = "Updated Title"

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

// ---- helper ----

func newReportUC() (*ReportUseCase, *MockReportRepository, *MockReportTypeRepository) {
	rr := new(MockReportRepository)
	tr := new(MockReportTypeRepository)
	uc := NewReportUseCase(rr, tr, nil, nil, nil)
	return uc, rr, tr
}

// =============================================================================
// Create
// =============================================================================

func TestReportUseCase_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uc, rr, tr := newReportUC()
		rt := &entities.ReportType{ID: 1, Name: "T", Code: "t", OutputFormat: domain.OutputFormatPDF}
		tr.On("GetByID", mock.Anything, int64(1)).Return(rt, nil).Once()
		rr.On("Create", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(nil).Once()
		rr.On("AddHistory", mock.Anything, mock.AnythingOfType("*entities.ReportHistory")).Return(nil).Once()

		out, err := uc.Create(context.Background(), 1, &dto.CreateReportInput{
			ReportTypeID: 1, Title: "R",
		})
		require.NoError(t, err)
		assert.Equal(t, "R", out.Title)
	})

	t.Run("report type not found", func(t *testing.T) {
		uc, _, tr := newReportUC()
		tr.On("GetByID", mock.Anything, int64(99)).Return(nil, nil).Once()
		_, err := uc.Create(context.Background(), 1, &dto.CreateReportInput{ReportTypeID: 99, Title: "R"})
		assert.ErrorIs(t, err, ErrReportTypeNotFound)
	})

	t.Run("report type repo error", func(t *testing.T) {
		uc, _, tr := newReportUC()
		tr.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db")).Once()
		_, err := uc.Create(context.Background(), 1, &dto.CreateReportInput{ReportTypeID: 1, Title: "R"})
		assert.Error(t, err)
	})

	t.Run("with period and parameters", func(t *testing.T) {
		uc, rr, tr := newReportUC()
		rt := &entities.ReportType{ID: 1, Name: "T", Code: "t", OutputFormat: domain.OutputFormatPDF}
		tr.On("GetByID", mock.Anything, int64(1)).Return(rt, nil).Once()
		rr.On("Create", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(nil).Once()
		rr.On("AddHistory", mock.Anything, mock.AnythingOfType("*entities.ReportHistory")).Return(nil).Once()

		start := time.Now().Add(-24 * time.Hour)
		end := time.Now()
		desc := "desc"
		out, err := uc.Create(context.Background(), 1, &dto.CreateReportInput{
			ReportTypeID: 1, Title: "R", Description: &desc,
			PeriodStart: &start, PeriodEnd: &end,
			Parameters: map[string]interface{}{"key": "val"},
			IsPublic:   true,
		})
		require.NoError(t, err)
		assert.Equal(t, "R", out.Title)
	})

	t.Run("create repo error", func(t *testing.T) {
		uc, rr, tr := newReportUC()
		rt := &entities.ReportType{ID: 1, Name: "T", Code: "t", OutputFormat: domain.OutputFormatPDF}
		tr.On("GetByID", mock.Anything, int64(1)).Return(rt, nil).Once()
		rr.On("Create", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(errors.New("db")).Once()
		_, err := uc.Create(context.Background(), 1, &dto.CreateReportInput{ReportTypeID: 1, Title: "R"})
		assert.Error(t, err)
	})
}

// =============================================================================
// GetByID
// =============================================================================

func TestReportUseCase_GetByID(t *testing.T) {
	t.Run("author access", func(t *testing.T) {
		uc, rr, tr := newReportUC()
		r := &entities.Report{ID: 1, ReportTypeID: 1, Title: "R", AuthorID: 1, Status: domain.ReportStatusDraft}
		rt := &entities.ReportType{ID: 1, Name: "T", Code: "t", OutputFormat: domain.OutputFormatPDF}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(1), domain.ReportPermissionRead).Return(false, nil).Once()
		tr.On("GetByID", mock.Anything, int64(1)).Return(rt, nil).Once()

		out, err := uc.GetByID(context.Background(), 1, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), out.ID)
	})

	t.Run("not found", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(99)).Return(nil, nil).Once()
		_, err := uc.GetByID(context.Background(), 99, 1)
		assert.ErrorIs(t, err, ErrReportNotFound)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db")).Once()
		_, err := uc.GetByID(context.Background(), 1, 1)
		assert.Error(t, err)
	})

	t.Run("hasAccess error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, ReportTypeID: 1, Title: "R", AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionRead).Return(false, errors.New("db")).Once()
		_, err := uc.GetByID(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("non-author with explicit access", func(t *testing.T) {
		uc, rr, tr := newReportUC()
		r := &entities.Report{ID: 1, ReportTypeID: 1, Title: "R", AuthorID: 1, Status: domain.ReportStatusDraft}
		rt := &entities.ReportType{ID: 1, Name: "T", Code: "t", OutputFormat: domain.OutputFormatPDF}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionRead).Return(true, nil).Once()
		tr.On("GetByID", mock.Anything, int64(1)).Return(rt, nil).Once()

		out, err := uc.GetByID(context.Background(), 1, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(1), out.ID)
	})

	t.Run("non-author public report", func(t *testing.T) {
		uc, rr, tr := newReportUC()
		r := &entities.Report{ID: 1, ReportTypeID: 1, Title: "R", AuthorID: 1, Status: domain.ReportStatusDraft, IsPublic: true}
		rt := &entities.ReportType{ID: 1, Name: "T", Code: "t", OutputFormat: domain.OutputFormatPDF}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionRead).Return(false, nil).Once()
		tr.On("GetByID", mock.Anything, int64(1)).Return(rt, nil).Once()

		out, err := uc.GetByID(context.Background(), 1, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(1), out.ID)
	})

	t.Run("unauthorized non-author private no access", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, ReportTypeID: 1, Title: "R", AuthorID: 1, Status: domain.ReportStatusDraft, IsPublic: false}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionRead).Return(false, nil).Once()
		_, err := uc.GetByID(context.Background(), 1, 2)
		assert.ErrorIs(t, err, ErrUnauthorized)
	})
}

// =============================================================================
// Update
// =============================================================================

func TestReportUseCase_Update(t *testing.T) {
	t.Run("success by author", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, ReportTypeID: 1, Title: "Old", AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("Save", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(nil).Once()
		rr.On("AddHistory", mock.Anything, mock.AnythingOfType("*entities.ReportHistory")).Return(nil).Once()

		title := updatedTitle
		out, err := uc.Update(context.Background(), 1, 1, &dto.UpdateReportInput{Title: &title})
		require.NoError(t, err)
		assert.Equal(t, updatedTitle, out.Title)
	})

	t.Run("not found", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(99)).Return(nil, nil).Once()
		title := "x"
		_, err := uc.Update(context.Background(), 99, 1, &dto.UpdateReportInput{Title: &title})
		assert.ErrorIs(t, err, ErrReportNotFound)
	})

	t.Run("repo error on get", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db")).Once()
		title := "x"
		_, err := uc.Update(context.Background(), 1, 1, &dto.UpdateReportInput{Title: &title})
		assert.Error(t, err)
	})

	t.Run("non-author without access", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, Title: "Old", AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionWrite).Return(false, nil).Once()
		title := "x"
		_, err := uc.Update(context.Background(), 1, 2, &dto.UpdateReportInput{Title: &title})
		assert.ErrorIs(t, err, ErrUnauthorized)
	})

	t.Run("non-author hasAccess error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, Title: "Old", AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionWrite).Return(false, errors.New("db")).Once()
		title := "x"
		_, err := uc.Update(context.Background(), 1, 2, &dto.UpdateReportInput{Title: &title})
		assert.Error(t, err)
	})

	t.Run("non-author with write access", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, Title: "Old", AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionWrite).Return(true, nil).Once()
		rr.On("Save", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(nil).Once()
		rr.On("AddHistory", mock.Anything, mock.AnythingOfType("*entities.ReportHistory")).Return(nil).Once()

		title := updatedTitle
		out, err := uc.Update(context.Background(), 1, 2, &dto.UpdateReportInput{Title: &title})
		require.NoError(t, err)
		assert.Equal(t, updatedTitle, out.Title)
	})

	t.Run("cannot edit non-draft report", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, Title: "Old", AuthorID: 1, Status: domain.ReportStatusPublished}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		title := "x"
		_, err := uc.Update(context.Background(), 1, 1, &dto.UpdateReportInput{Title: &title})
		assert.ErrorIs(t, err, ErrCannotModifyReport)
	})

	t.Run("can edit rejected report", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, Title: "Old", AuthorID: 1, Status: domain.ReportStatusRejected}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("Save", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(nil).Once()
		rr.On("AddHistory", mock.Anything, mock.AnythingOfType("*entities.ReportHistory")).Return(nil).Once()

		title := updatedTitle
		out, err := uc.Update(context.Background(), 1, 1, &dto.UpdateReportInput{Title: &title})
		require.NoError(t, err)
		assert.Equal(t, updatedTitle, out.Title)
	})

	t.Run("update all fields", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, Title: "Old", AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("Save", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(nil).Once()
		rr.On("AddHistory", mock.Anything, mock.AnythingOfType("*entities.ReportHistory")).Return(nil).Once()

		title := updatedTitle
		desc := "new desc"
		start := time.Now().Add(-24 * time.Hour)
		end := time.Now()
		isPublic := true
		out, err := uc.Update(context.Background(), 1, 1, &dto.UpdateReportInput{
			Title: &title, Description: &desc,
			PeriodStart: &start, PeriodEnd: &end,
			Parameters: map[string]interface{}{"k": "v"},
			IsPublic:   &isPublic,
		})
		require.NoError(t, err)
		assert.Equal(t, updatedTitle, out.Title)
	})

	t.Run("save error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, Title: "Old", AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("Save", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(errors.New("db")).Once()
		title := "x"
		_, err := uc.Update(context.Background(), 1, 1, &dto.UpdateReportInput{Title: &title})
		assert.Error(t, err)
	})
}

// =============================================================================
// Delete
// =============================================================================

func TestReportUseCase_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("Delete", mock.Anything, int64(1)).Return(nil).Once()
		assert.NoError(t, uc.Delete(context.Background(), 1, 1))
	})

	t.Run("not found", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(99)).Return(nil, nil).Once()
		assert.ErrorIs(t, uc.Delete(context.Background(), 99, 1), ErrReportNotFound)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db")).Once()
		assert.Error(t, uc.Delete(context.Background(), 1, 1))
	})

	t.Run("non-author", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		assert.ErrorIs(t, uc.Delete(context.Background(), 1, 2), ErrUnauthorized)
	})

	t.Run("finalized report", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusPublished}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		assert.ErrorIs(t, uc.Delete(context.Background(), 1, 1), ErrCannotModifyReport)
	})

	t.Run("with file path no s3", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		fp := "reports/file.pdf"
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusDraft, FilePath: &fp}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("Delete", mock.Anything, int64(1)).Return(nil).Once()
		assert.NoError(t, uc.Delete(context.Background(), 1, 1))
	})

	t.Run("delete repo error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("Delete", mock.Anything, int64(1)).Return(errors.New("db")).Once()
		assert.Error(t, uc.Delete(context.Background(), 1, 1))
	})
}

// =============================================================================
// List
// =============================================================================

func TestReportUseCase_List(t *testing.T) {
	t.Run("basic pagination", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		reports := []*entities.Report{
			{ID: 1, Title: "R1", AuthorID: 1, Status: domain.ReportStatusDraft},
			{ID: 2, Title: "R2", AuthorID: 1, Status: domain.ReportStatusReady},
		}
		rr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportFilter")).Return(int64(2), nil).Once()
		rr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportFilter"), 20, 0).Return(reports, nil).Once()

		out, err := uc.List(context.Background(), 1, &dto.ReportFilterInput{Page: 1, PageSize: 20})
		require.NoError(t, err)
		assert.Len(t, out.Reports, 2)
		assert.Equal(t, 1, out.TotalPages)
	})

	t.Run("empty result", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportFilter")).Return(int64(0), nil).Once()
		rr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportFilter"), 20, 0).Return([]*entities.Report{}, nil).Once()

		out, err := uc.List(context.Background(), 1, &dto.ReportFilterInput{Page: 1, PageSize: 20})
		require.NoError(t, err)
		assert.Len(t, out.Reports, 0)
	})

	t.Run("with status filter", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		status := "draft"
		rr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportFilter")).Return(int64(1), nil).Once()
		rr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportFilter"), 20, 0).Return([]*entities.Report{{ID: 1, Status: domain.ReportStatusDraft}}, nil).Once()

		out, err := uc.List(context.Background(), 1, &dto.ReportFilterInput{Page: 1, PageSize: 20, Status: &status})
		require.NoError(t, err)
		assert.Len(t, out.Reports, 1)
	})

	t.Run("with invalid status filter", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		status := "invalid_status"
		rr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportFilter")).Return(int64(0), nil).Once()
		rr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportFilter"), 20, 0).Return([]*entities.Report{}, nil).Once()

		out, err := uc.List(context.Background(), 1, &dto.ReportFilterInput{Page: 1, PageSize: 20, Status: &status})
		require.NoError(t, err)
		assert.Len(t, out.Reports, 0)
	})

	t.Run("with date filters", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		ps := "2024-01-01"
		pe := "2024-12-31"
		rr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportFilter")).Return(int64(0), nil).Once()
		rr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportFilter"), 20, 0).Return([]*entities.Report{}, nil).Once()

		out, err := uc.List(context.Background(), 1, &dto.ReportFilterInput{Page: 1, PageSize: 20, PeriodStart: &ps, PeriodEnd: &pe})
		require.NoError(t, err)
		assert.NotNil(t, out)
	})

	t.Run("with invalid date filters", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		ps := "not-a-date"
		pe := "also-bad"
		rr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportFilter")).Return(int64(0), nil).Once()
		rr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportFilter"), 20, 0).Return([]*entities.Report{}, nil).Once()

		out, err := uc.List(context.Background(), 1, &dto.ReportFilterInput{Page: 1, PageSize: 20, PeriodStart: &ps, PeriodEnd: &pe})
		require.NoError(t, err)
		assert.NotNil(t, out)
	})

	t.Run("default pagination for bad values", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportFilter")).Return(int64(0), nil).Once()
		rr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportFilter"), 20, 0).Return([]*entities.Report{}, nil).Once()

		out, err := uc.List(context.Background(), 1, &dto.ReportFilterInput{Page: 0, PageSize: 0})
		require.NoError(t, err)
		assert.Equal(t, 1, out.Page)
		assert.Equal(t, 20, out.PageSize)
	})

	t.Run("pageSize over 100 defaults to 20", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportFilter")).Return(int64(0), nil).Once()
		rr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportFilter"), 20, 0).Return([]*entities.Report{}, nil).Once()

		out, err := uc.List(context.Background(), 1, &dto.ReportFilterInput{Page: 1, PageSize: 200})
		require.NoError(t, err)
		assert.Equal(t, 20, out.PageSize)
	})

	t.Run("total pages rounding", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportFilter")).Return(int64(21), nil).Once()
		rr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportFilter"), 20, 0).Return([]*entities.Report{}, nil).Once()

		out, err := uc.List(context.Background(), 1, &dto.ReportFilterInput{Page: 1, PageSize: 20})
		require.NoError(t, err)
		assert.Equal(t, 2, out.TotalPages)
	})

	t.Run("count error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportFilter")).Return(int64(0), errors.New("db")).Once()
		_, err := uc.List(context.Background(), 1, &dto.ReportFilterInput{Page: 1, PageSize: 20})
		assert.Error(t, err)
	})

	t.Run("list error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportFilter")).Return(int64(1), nil).Once()
		rr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportFilter"), 20, 0).Return(nil, errors.New("db")).Once()
		_, err := uc.List(context.Background(), 1, &dto.ReportFilterInput{Page: 1, PageSize: 20})
		assert.Error(t, err)
	})

	t.Run("page 2 offset calculation", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportFilter")).Return(int64(30), nil).Once()
		rr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportFilter"), 10, 10).Return([]*entities.Report{}, nil).Once()

		out, err := uc.List(context.Background(), 1, &dto.ReportFilterInput{Page: 2, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, 2, out.Page)
	})
}

// =============================================================================
// Generate
// =============================================================================

func TestReportUseCase_Generate(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(99)).Return(nil, nil).Once()
		_, err := uc.Generate(context.Background(), 99, 1, nil)
		assert.ErrorIs(t, err, ErrReportNotFound)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db")).Once()
		_, err := uc.Generate(context.Background(), 1, 1, nil)
		assert.Error(t, err)
	})

	t.Run("non-author", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		_, err := uc.Generate(context.Background(), 1, 2, nil)
		assert.ErrorIs(t, err, ErrUnauthorized)
	})

	t.Run("invalid status transition", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusReady}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		_, err := uc.Generate(context.Background(), 1, 1, nil)
		assert.Error(t, err)
	})

	t.Run("success without parameters", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("CreateGenerationLog", mock.Anything, mock.AnythingOfType("*entities.ReportGenerationLog")).Return(nil).Once()
		rr.On("Save", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(nil).Maybe()
		rr.On("UpdateGenerationLog", mock.Anything, mock.AnythingOfType("*entities.ReportGenerationLog")).Return(nil).Maybe()
		rr.On("AddHistory", mock.Anything, mock.AnythingOfType("*entities.ReportHistory")).Return(nil).Maybe()

		out, err := uc.Generate(context.Background(), 1, 1, nil)
		require.NoError(t, err)
		assert.NotNil(t, out)
	})

	t.Run("success with parameters", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("CreateGenerationLog", mock.Anything, mock.AnythingOfType("*entities.ReportGenerationLog")).Return(nil).Once()
		rr.On("Save", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(nil).Maybe()
		rr.On("UpdateGenerationLog", mock.Anything, mock.AnythingOfType("*entities.ReportGenerationLog")).Return(nil).Maybe()
		rr.On("AddHistory", mock.Anything, mock.AnythingOfType("*entities.ReportHistory")).Return(nil).Maybe()

		input := &dto.GenerateReportInput{Parameters: map[string]interface{}{"k": "v"}}
		out, err := uc.Generate(context.Background(), 1, 1, input)
		require.NoError(t, err)
		assert.NotNil(t, out)
	})

	t.Run("create generation log error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("CreateGenerationLog", mock.Anything, mock.AnythingOfType("*entities.ReportGenerationLog")).Return(errors.New("db")).Once()
		_, err := uc.Generate(context.Background(), 1, 1, nil)
		assert.Error(t, err)
	})

	t.Run("save error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("CreateGenerationLog", mock.Anything, mock.AnythingOfType("*entities.ReportGenerationLog")).Return(nil).Once()
		rr.On("Save", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(errors.New("db")).Once()
		_, err := uc.Generate(context.Background(), 1, 1, nil)
		assert.Error(t, err)
	})
}

// =============================================================================
// SubmitForReview
// =============================================================================

func TestReportUseCase_SubmitForReview(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusReady}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("Save", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(nil).Once()
		rr.On("AddHistory", mock.Anything, mock.AnythingOfType("*entities.ReportHistory")).Return(nil).Once()

		out, err := uc.SubmitForReview(context.Background(), 1, 1)
		require.NoError(t, err)
		assert.NotNil(t, out)
	})

	t.Run("not found", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(99)).Return(nil, nil).Once()
		_, err := uc.SubmitForReview(context.Background(), 99, 1)
		assert.ErrorIs(t, err, ErrReportNotFound)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db")).Once()
		_, err := uc.SubmitForReview(context.Background(), 1, 1)
		assert.Error(t, err)
	})

	t.Run("non-author", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusReady}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		_, err := uc.SubmitForReview(context.Background(), 1, 2)
		assert.ErrorIs(t, err, ErrUnauthorized)
	})

	t.Run("wrong status", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		_, err := uc.SubmitForReview(context.Background(), 1, 1)
		assert.Error(t, err)
	})

	t.Run("save error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusReady}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("Save", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(errors.New("db")).Once()
		_, err := uc.SubmitForReview(context.Background(), 1, 1)
		assert.Error(t, err)
	})
}

// =============================================================================
// Review
// =============================================================================

func TestReportUseCase_Review(t *testing.T) {
	t.Run("approve success", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusReviewing}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionApprove).Return(true, nil).Once()
		rr.On("Save", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(nil).Once()
		rr.On("AddHistory", mock.Anything, mock.AnythingOfType("*entities.ReportHistory")).Return(nil).Once()

		out, err := uc.Review(context.Background(), 1, 2, &dto.ReviewReportInput{Action: "approve", Comment: "ok"})
		require.NoError(t, err)
		assert.NotNil(t, out)
	})

	t.Run("reject success", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusReviewing}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionApprove).Return(true, nil).Once()
		rr.On("Save", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(nil).Once()
		rr.On("AddHistory", mock.Anything, mock.AnythingOfType("*entities.ReportHistory")).Return(nil).Once()

		out, err := uc.Review(context.Background(), 1, 2, &dto.ReviewReportInput{Action: "reject", Comment: "bad"})
		require.NoError(t, err)
		assert.NotNil(t, out)
	})

	t.Run("invalid action", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusReviewing}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionApprove).Return(true, nil).Once()

		_, err := uc.Review(context.Background(), 1, 2, &dto.ReviewReportInput{Action: "invalid"})
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("not found", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(99)).Return(nil, nil).Once()
		_, err := uc.Review(context.Background(), 99, 1, &dto.ReviewReportInput{Action: "approve"})
		assert.ErrorIs(t, err, ErrReportNotFound)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db")).Once()
		_, err := uc.Review(context.Background(), 1, 1, &dto.ReviewReportInput{Action: "approve"})
		assert.Error(t, err)
	})

	t.Run("hasAccess error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusReviewing}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionApprove).Return(false, errors.New("db")).Once()
		_, err := uc.Review(context.Background(), 1, 2, &dto.ReviewReportInput{Action: "approve"})
		assert.Error(t, err)
	})

	t.Run("no access", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusReviewing}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionApprove).Return(false, nil).Once()
		_, err := uc.Review(context.Background(), 1, 2, &dto.ReviewReportInput{Action: "approve"})
		assert.ErrorIs(t, err, ErrUnauthorized)
	})

	t.Run("approve wrong status", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionApprove).Return(true, nil).Once()
		_, err := uc.Review(context.Background(), 1, 2, &dto.ReviewReportInput{Action: "approve"})
		assert.Error(t, err)
	})

	t.Run("save error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusReviewing}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionApprove).Return(true, nil).Once()
		rr.On("Save", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(errors.New("db")).Once()
		_, err := uc.Review(context.Background(), 1, 2, &dto.ReviewReportInput{Action: "approve"})
		assert.Error(t, err)
	})
}

// =============================================================================
// Publish
// =============================================================================

func TestReportUseCase_Publish(t *testing.T) {
	t.Run("success by author", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusApproved}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(1), domain.ReportPermissionPublish).Return(false, nil).Once()
		rr.On("Save", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(nil).Once()
		rr.On("AddHistory", mock.Anything, mock.AnythingOfType("*entities.ReportHistory")).Return(nil).Once()

		out, err := uc.Publish(context.Background(), 1, 1, &dto.PublishReportInput{IsPublic: true})
		require.NoError(t, err)
		assert.NotNil(t, out)
	})

	t.Run("success by non-author with publish access", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusApproved}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionPublish).Return(true, nil).Once()
		rr.On("Save", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(nil).Once()
		rr.On("AddHistory", mock.Anything, mock.AnythingOfType("*entities.ReportHistory")).Return(nil).Once()

		out, err := uc.Publish(context.Background(), 1, 2, &dto.PublishReportInput{IsPublic: false})
		require.NoError(t, err)
		assert.NotNil(t, out)
	})

	t.Run("not found", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(99)).Return(nil, nil).Once()
		_, err := uc.Publish(context.Background(), 99, 1, &dto.PublishReportInput{})
		assert.ErrorIs(t, err, ErrReportNotFound)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db")).Once()
		_, err := uc.Publish(context.Background(), 1, 1, &dto.PublishReportInput{})
		assert.Error(t, err)
	})

	t.Run("hasAccess error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusApproved}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionPublish).Return(false, errors.New("db")).Once()
		_, err := uc.Publish(context.Background(), 1, 2, &dto.PublishReportInput{})
		assert.Error(t, err)
	})

	t.Run("unauthorized", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusApproved}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionPublish).Return(false, nil).Once()
		_, err := uc.Publish(context.Background(), 1, 2, &dto.PublishReportInput{})
		assert.ErrorIs(t, err, ErrUnauthorized)
	})

	t.Run("wrong status", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusDraft}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(1), domain.ReportPermissionPublish).Return(false, nil).Once()
		_, err := uc.Publish(context.Background(), 1, 1, &dto.PublishReportInput{})
		assert.Error(t, err)
	})

	t.Run("save error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusApproved}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(1), domain.ReportPermissionPublish).Return(false, nil).Once()
		rr.On("Save", mock.Anything, mock.AnythingOfType("*entities.Report")).Return(errors.New("db")).Once()
		_, err := uc.Publish(context.Background(), 1, 1, &dto.PublishReportInput{})
		assert.Error(t, err)
	})
}

// =============================================================================
// AddAccess
// =============================================================================

func TestReportUseCase_AddAccess(t *testing.T) {
	t.Run("add user access", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("AddAccess", mock.Anything, mock.AnythingOfType("*entities.ReportAccess")).Return(nil).Once()

		uid := int64(2)
		out, err := uc.AddAccess(context.Background(), 1, 1, &dto.AddAccessInput{UserID: &uid, Permission: "read"})
		require.NoError(t, err)
		assert.NotNil(t, out)
	})

	t.Run("add role access", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("AddAccess", mock.Anything, mock.AnythingOfType("*entities.ReportAccess")).Return(nil).Once()

		out, err := uc.AddAccess(context.Background(), 1, 1, &dto.AddAccessInput{Role: "admin", Permission: "write"})
		require.NoError(t, err)
		assert.NotNil(t, out)
	})

	t.Run("invalid input no user or role", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		_, err := uc.AddAccess(context.Background(), 1, 1, &dto.AddAccessInput{Permission: "read"})
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("non-author", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		uid := int64(3)
		_, err := uc.AddAccess(context.Background(), 1, 2, &dto.AddAccessInput{UserID: &uid, Permission: "read"})
		assert.ErrorIs(t, err, ErrUnauthorized)
	})

	t.Run("not found", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(99)).Return(nil, nil).Once()
		uid := int64(2)
		_, err := uc.AddAccess(context.Background(), 99, 1, &dto.AddAccessInput{UserID: &uid, Permission: "read"})
		assert.ErrorIs(t, err, ErrReportNotFound)
	})

	t.Run("repo error on get", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db")).Once()
		uid := int64(2)
		_, err := uc.AddAccess(context.Background(), 1, 1, &dto.AddAccessInput{UserID: &uid, Permission: "read"})
		assert.Error(t, err)
	})

	t.Run("add access repo error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("AddAccess", mock.Anything, mock.AnythingOfType("*entities.ReportAccess")).Return(errors.New("db")).Once()
		uid := int64(2)
		_, err := uc.AddAccess(context.Background(), 1, 1, &dto.AddAccessInput{UserID: &uid, Permission: "read"})
		assert.Error(t, err)
	})
}

// =============================================================================
// RemoveAccess
// =============================================================================

func TestReportUseCase_RemoveAccess(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("RemoveAccess", mock.Anything, int64(1), int64(5)).Return(nil).Once()
		assert.NoError(t, uc.RemoveAccess(context.Background(), 1, 5, 1))
	})

	t.Run("non-author", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		assert.ErrorIs(t, uc.RemoveAccess(context.Background(), 1, 5, 2), ErrUnauthorized)
	})

	t.Run("not found", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(99)).Return(nil, nil).Once()
		assert.ErrorIs(t, uc.RemoveAccess(context.Background(), 99, 5, 1), ErrReportNotFound)
	})

	t.Run("repo error on get", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db")).Once()
		assert.Error(t, uc.RemoveAccess(context.Background(), 1, 5, 1))
	})
}

// =============================================================================
// GetAccess
// =============================================================================

func TestReportUseCase_GetAccess(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		uid := int64(2)
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("GetAccessByReport", mock.Anything, int64(1)).Return([]*entities.ReportAccess{
			{ID: 1, ReportID: 1, UserID: &uid, Permission: domain.ReportPermissionRead},
		}, nil).Once()
		out, err := uc.GetAccess(context.Background(), 1, 1)
		require.NoError(t, err)
		assert.Len(t, out, 1)
	})

	t.Run("non-author", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		_, err := uc.GetAccess(context.Background(), 1, 2)
		assert.ErrorIs(t, err, ErrUnauthorized)
	})

	t.Run("not found", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(99)).Return(nil, nil).Once()
		_, err := uc.GetAccess(context.Background(), 99, 1)
		assert.ErrorIs(t, err, ErrReportNotFound)
	})

	t.Run("repo error on get", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db")).Once()
		_, err := uc.GetAccess(context.Background(), 1, 1)
		assert.Error(t, err)
	})

	t.Run("get access repo error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("GetAccessByReport", mock.Anything, int64(1)).Return(nil, errors.New("db")).Once()
		_, err := uc.GetAccess(context.Background(), 1, 1)
		assert.Error(t, err)
	})
}

// =============================================================================
// AddComment
// =============================================================================

func TestReportUseCase_AddComment(t *testing.T) {
	t.Run("by author", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(1), domain.ReportPermissionRead).Return(false, nil).Once()
		rr.On("AddComment", mock.Anything, mock.AnythingOfType("*entities.ReportComment")).Return(nil).Once()

		out, err := uc.AddComment(context.Background(), 1, 1, &dto.AddCommentInput{Content: "test"})
		require.NoError(t, err)
		assert.Equal(t, "test", out.Content)
	})

	t.Run("by user with access", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionRead).Return(true, nil).Once()
		rr.On("AddComment", mock.Anything, mock.AnythingOfType("*entities.ReportComment")).Return(nil).Once()

		out, err := uc.AddComment(context.Background(), 1, 2, &dto.AddCommentInput{Content: "test"})
		require.NoError(t, err)
		assert.NotNil(t, out)
	})

	t.Run("on public report by non-author", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, IsPublic: true}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionRead).Return(false, nil).Once()
		rr.On("AddComment", mock.Anything, mock.AnythingOfType("*entities.ReportComment")).Return(nil).Once()

		out, err := uc.AddComment(context.Background(), 1, 2, &dto.AddCommentInput{Content: "test"})
		require.NoError(t, err)
		assert.NotNil(t, out)
	})

	t.Run("unauthorized", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, IsPublic: false}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionRead).Return(false, nil).Once()
		_, err := uc.AddComment(context.Background(), 1, 2, &dto.AddCommentInput{Content: "test"})
		assert.ErrorIs(t, err, ErrUnauthorized)
	})

	t.Run("not found", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(99)).Return(nil, nil).Once()
		_, err := uc.AddComment(context.Background(), 99, 1, &dto.AddCommentInput{Content: "test"})
		assert.ErrorIs(t, err, ErrReportNotFound)
	})

	t.Run("repo error on get", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db")).Once()
		_, err := uc.AddComment(context.Background(), 1, 1, &dto.AddCommentInput{Content: "test"})
		assert.Error(t, err)
	})

	t.Run("hasAccess error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionRead).Return(false, errors.New("db")).Once()
		_, err := uc.AddComment(context.Background(), 1, 2, &dto.AddCommentInput{Content: "test"})
		assert.Error(t, err)
	})

	t.Run("add comment repo error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(1), domain.ReportPermissionRead).Return(false, nil).Once()
		rr.On("AddComment", mock.Anything, mock.AnythingOfType("*entities.ReportComment")).Return(errors.New("db")).Once()
		_, err := uc.AddComment(context.Background(), 1, 1, &dto.AddCommentInput{Content: "test"})
		assert.Error(t, err)
	})
}

// =============================================================================
// GetComments
// =============================================================================

func TestReportUseCase_GetComments(t *testing.T) {
	t.Run("by author", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(1), domain.ReportPermissionRead).Return(false, nil).Once()
		rr.On("GetCommentsByReport", mock.Anything, int64(1)).Return([]*entities.ReportComment{
			{ID: 1, ReportID: 1, AuthorID: 1, Content: "c1"},
		}, nil).Once()

		out, err := uc.GetComments(context.Background(), 1, 1)
		require.NoError(t, err)
		assert.Len(t, out, 1)
	})

	t.Run("unauthorized", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, IsPublic: false}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionRead).Return(false, nil).Once()
		_, err := uc.GetComments(context.Background(), 1, 2)
		assert.ErrorIs(t, err, ErrUnauthorized)
	})

	t.Run("not found", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(99)).Return(nil, nil).Once()
		_, err := uc.GetComments(context.Background(), 99, 1)
		assert.ErrorIs(t, err, ErrReportNotFound)
	})

	t.Run("repo error on get", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db")).Once()
		_, err := uc.GetComments(context.Background(), 1, 1)
		assert.Error(t, err)
	})

	t.Run("hasAccess error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionRead).Return(false, errors.New("db")).Once()
		_, err := uc.GetComments(context.Background(), 1, 2)
		assert.Error(t, err)
	})

	t.Run("get comments repo error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(1), domain.ReportPermissionRead).Return(false, nil).Once()
		rr.On("GetCommentsByReport", mock.Anything, int64(1)).Return(nil, errors.New("db")).Once()
		_, err := uc.GetComments(context.Background(), 1, 1)
		assert.Error(t, err)
	})

	t.Run("public report by non-author", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1, IsPublic: true}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("HasAccess", mock.Anything, int64(1), int64(2), domain.ReportPermissionRead).Return(false, nil).Once()
		rr.On("GetCommentsByReport", mock.Anything, int64(1)).Return([]*entities.ReportComment{}, nil).Once()

		out, err := uc.GetComments(context.Background(), 1, 2)
		require.NoError(t, err)
		assert.Len(t, out, 0)
	})
}

// =============================================================================
// GetHistory
// =============================================================================

func TestReportUseCase_GetHistory(t *testing.T) {
	t.Run("success with default limit", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		uid := int64(1)
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("GetHistoryByReport", mock.Anything, int64(1), 50, 0).Return([]*entities.ReportHistory{
			{ID: 1, ReportID: 1, UserID: &uid, Action: entities.ReportActionCreated},
		}, nil).Once()

		out, err := uc.GetHistory(context.Background(), 1, 1, 0, 0)
		require.NoError(t, err)
		assert.Len(t, out, 1)
	})

	t.Run("with custom limit", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("GetHistoryByReport", mock.Anything, int64(1), 10, 5).Return([]*entities.ReportHistory{}, nil).Once()

		out, err := uc.GetHistory(context.Background(), 1, 1, 10, 5)
		require.NoError(t, err)
		assert.Len(t, out, 0)
	})

	t.Run("non-author", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		_, err := uc.GetHistory(context.Background(), 1, 2, 0, 0)
		assert.ErrorIs(t, err, ErrUnauthorized)
	})

	t.Run("not found", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(99)).Return(nil, nil).Once()
		_, err := uc.GetHistory(context.Background(), 99, 1, 0, 0)
		assert.ErrorIs(t, err, ErrReportNotFound)
	})

	t.Run("repo error on get", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		rr.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db")).Once()
		_, err := uc.GetHistory(context.Background(), 1, 1, 0, 0)
		assert.Error(t, err)
	})

	t.Run("get history repo error", func(t *testing.T) {
		uc, rr, _ := newReportUC()
		r := &entities.Report{ID: 1, AuthorID: 1}
		rr.On("GetByID", mock.Anything, int64(1)).Return(r, nil).Once()
		rr.On("GetHistoryByReport", mock.Anything, int64(1), 50, 0).Return(nil, errors.New("db")).Once()
		_, err := uc.GetHistory(context.Background(), 1, 1, 0, 0)
		assert.Error(t, err)
	})
}

// =============================================================================
// GetReportTypes
// =============================================================================

func TestReportUseCase_GetReportTypes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uc, _, tr := newReportUC()
		rts := []*entities.ReportType{
			{ID: 1, Name: "T1", Code: "t1", OutputFormat: domain.OutputFormatPDF},
		}
		tr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportTypeFilter")).Return(int64(1), nil).Once()
		tr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportTypeFilter"), 20, 0).Return(rts, nil).Once()
		tr.On("GetParametersByReportType", mock.Anything, int64(1)).Return([]*entities.ReportParameter{
			{ID: 1, ReportTypeID: 1, ParameterName: "p1", ParameterType: domain.ParameterTypeString},
		}, nil).Once()

		out, err := uc.GetReportTypes(context.Background(), &dto.ReportTypeFilterInput{Page: 1, PageSize: 20})
		require.NoError(t, err)
		assert.Len(t, out.ReportTypes, 1)
	})

	t.Run("with category filter", func(t *testing.T) {
		uc, _, tr := newReportUC()
		cat := "academic"
		tr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportTypeFilter")).Return(int64(0), nil).Once()
		tr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportTypeFilter"), 20, 0).Return([]*entities.ReportType{}, nil).Once()

		out, err := uc.GetReportTypes(context.Background(), &dto.ReportTypeFilterInput{Page: 1, PageSize: 20, Category: &cat})
		require.NoError(t, err)
		assert.Len(t, out.ReportTypes, 0)
	})

	t.Run("with invalid category filter", func(t *testing.T) {
		uc, _, tr := newReportUC()
		cat := "invalid_cat"
		tr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportTypeFilter")).Return(int64(0), nil).Once()
		tr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportTypeFilter"), 20, 0).Return([]*entities.ReportType{}, nil).Once()

		out, err := uc.GetReportTypes(context.Background(), &dto.ReportTypeFilterInput{Page: 1, PageSize: 20, Category: &cat})
		require.NoError(t, err)
		assert.NotNil(t, out)
	})

	t.Run("default pagination", func(t *testing.T) {
		uc, _, tr := newReportUC()
		tr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportTypeFilter")).Return(int64(0), nil).Once()
		tr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportTypeFilter"), 20, 0).Return([]*entities.ReportType{}, nil).Once()

		out, err := uc.GetReportTypes(context.Background(), &dto.ReportTypeFilterInput{Page: 0, PageSize: 0})
		require.NoError(t, err)
		assert.Equal(t, 1, out.Page)
		assert.Equal(t, 20, out.PageSize)
	})

	t.Run("total pages rounding", func(t *testing.T) {
		uc, _, tr := newReportUC()
		tr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportTypeFilter")).Return(int64(21), nil).Once()
		tr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportTypeFilter"), 20, 0).Return([]*entities.ReportType{}, nil).Once()

		out, err := uc.GetReportTypes(context.Background(), &dto.ReportTypeFilterInput{Page: 1, PageSize: 20})
		require.NoError(t, err)
		assert.Equal(t, 2, out.TotalPages)
	})

	t.Run("count error", func(t *testing.T) {
		uc, _, tr := newReportUC()
		tr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportTypeFilter")).Return(int64(0), errors.New("db")).Once()
		_, err := uc.GetReportTypes(context.Background(), &dto.ReportTypeFilterInput{Page: 1, PageSize: 20})
		assert.Error(t, err)
	})

	t.Run("list error", func(t *testing.T) {
		uc, _, tr := newReportUC()
		tr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportTypeFilter")).Return(int64(1), nil).Once()
		tr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportTypeFilter"), 20, 0).Return(nil, errors.New("db")).Once()
		_, err := uc.GetReportTypes(context.Background(), &dto.ReportTypeFilterInput{Page: 1, PageSize: 20})
		assert.Error(t, err)
	})

	t.Run("with nil parameters from repo", func(t *testing.T) {
		uc, _, tr := newReportUC()
		rts := []*entities.ReportType{
			{ID: 1, Name: "T1", Code: "t1", OutputFormat: domain.OutputFormatPDF},
		}
		tr.On("Count", mock.Anything, mock.AnythingOfType("repositories.ReportTypeFilter")).Return(int64(1), nil).Once()
		tr.On("List", mock.Anything, mock.AnythingOfType("repositories.ReportTypeFilter"), 20, 0).Return(rts, nil).Once()
		tr.On("GetParametersByReportType", mock.Anything, int64(1)).Return(nil, nil).Once()

		out, err := uc.GetReportTypes(context.Background(), &dto.ReportTypeFilterInput{Page: 1, PageSize: 20})
		require.NoError(t, err)
		assert.Len(t, out.ReportTypes, 1)
	})
}

// =============================================================================
// GetReportTypeByID
// =============================================================================

func TestReportUseCase_GetReportTypeByID(t *testing.T) {
	t.Run("success with params and templates", func(t *testing.T) {
		uc, _, tr := newReportUC()
		rt := &entities.ReportType{ID: 1, Name: "T", Code: "t", OutputFormat: domain.OutputFormatPDF}
		tr.On("GetByID", mock.Anything, int64(1)).Return(rt, nil).Once()
		tr.On("GetParametersByReportType", mock.Anything, int64(1)).Return([]*entities.ReportParameter{
			{ID: 1, ReportTypeID: 1, ParameterName: "p1", ParameterType: domain.ParameterTypeString},
		}, nil).Once()
		tr.On("GetTemplatesByReportType", mock.Anything, int64(1)).Return([]*entities.ReportTemplate{
			{ID: 1, ReportTypeID: 1, Name: "tmpl1", Content: "content"},
		}, nil).Once()

		out, err := uc.GetReportTypeByID(context.Background(), 1)
		require.NoError(t, err)
		assert.Equal(t, "T", out.Name)
	})

	t.Run("not found", func(t *testing.T) {
		uc, _, tr := newReportUC()
		tr.On("GetByID", mock.Anything, int64(99)).Return(nil, nil).Once()
		_, err := uc.GetReportTypeByID(context.Background(), 99)
		assert.ErrorIs(t, err, ErrReportTypeNotFound)
	})

	t.Run("repo error", func(t *testing.T) {
		uc, _, tr := newReportUC()
		tr.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db")).Once()
		_, err := uc.GetReportTypeByID(context.Background(), 1)
		assert.Error(t, err)
	})

	t.Run("nil params and templates", func(t *testing.T) {
		uc, _, tr := newReportUC()
		rt := &entities.ReportType{ID: 1, Name: "T", Code: "t", OutputFormat: domain.OutputFormatPDF}
		tr.On("GetByID", mock.Anything, int64(1)).Return(rt, nil).Once()
		tr.On("GetParametersByReportType", mock.Anything, int64(1)).Return(nil, nil).Once()
		tr.On("GetTemplatesByReportType", mock.Anything, int64(1)).Return(nil, nil).Once()

		out, err := uc.GetReportTypeByID(context.Background(), 1)
		require.NoError(t, err)
		assert.Equal(t, "T", out.Name)
	})
}

// =============================================================================
// logAudit
// =============================================================================

func TestReportUseCase_logAudit(t *testing.T) {
	t.Run("nil audit log", func(t *testing.T) {
		uc, _, _ := newReportUC()
		// Should not panic
		uc.logAudit(context.Background(), "test", 1, 1, nil)
	})

	t.Run("nil audit log with details", func(t *testing.T) {
		uc, _, _ := newReportUC()
		uc.logAudit(context.Background(), "test", 1, 1, map[string]interface{}{"key": "val"})
	})
}
