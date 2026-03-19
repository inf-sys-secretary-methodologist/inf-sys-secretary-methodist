package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/interfaces/http/handlers"
)

// --- Mock Report Repository ---

type mockReportRepo struct{ mock.Mock }

func (m *mockReportRepo) Create(ctx context.Context, r *entities.Report) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}
func (m *mockReportRepo) Save(ctx context.Context, r *entities.Report) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}
func (m *mockReportRepo) GetByID(ctx context.Context, id int64) (*entities.Report, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Report), args.Error(1)
}
func (m *mockReportRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockReportRepo) List(ctx context.Context, f repositories.ReportFilter, l, o int) ([]*entities.Report, error) {
	args := m.Called(ctx, f, l, o)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Report), args.Error(1)
}
func (m *mockReportRepo) Count(ctx context.Context, f repositories.ReportFilter) (int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockReportRepo) GetByAuthor(ctx context.Context, authorID int64, l, o int) ([]*entities.Report, error) {
	return nil, nil
}
func (m *mockReportRepo) GetByStatus(ctx context.Context, s domain.ReportStatus, l, o int) ([]*entities.Report, error) {
	return nil, nil
}
func (m *mockReportRepo) GetByReportType(ctx context.Context, id int64, l, o int) ([]*entities.Report, error) {
	return nil, nil
}
func (m *mockReportRepo) GetPublicReports(ctx context.Context, l, o int) ([]*entities.Report, error) {
	return nil, nil
}
func (m *mockReportRepo) AddAccess(ctx context.Context, a *entities.ReportAccess) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}
func (m *mockReportRepo) RemoveAccess(ctx context.Context, reportID, accessID int64) error {
	args := m.Called(ctx, reportID, accessID)
	return args.Error(0)
}
func (m *mockReportRepo) GetAccessByReport(ctx context.Context, reportID int64) ([]*entities.ReportAccess, error) {
	args := m.Called(ctx, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ReportAccess), args.Error(1)
}
func (m *mockReportRepo) HasAccess(ctx context.Context, reportID, userID int64, p domain.ReportPermission) (bool, error) {
	args := m.Called(ctx, reportID, userID, p)
	return args.Bool(0), args.Error(1)
}
func (m *mockReportRepo) AddComment(ctx context.Context, c *entities.ReportComment) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}
func (m *mockReportRepo) UpdateComment(ctx context.Context, c *entities.ReportComment) error { return nil }
func (m *mockReportRepo) DeleteComment(ctx context.Context, id int64) error                  { return nil }
func (m *mockReportRepo) GetCommentsByReport(ctx context.Context, reportID int64) ([]*entities.ReportComment, error) {
	args := m.Called(ctx, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ReportComment), args.Error(1)
}
func (m *mockReportRepo) AddHistory(ctx context.Context, h *entities.ReportHistory) error {
	args := m.Called(ctx, h)
	return args.Error(0)
}
func (m *mockReportRepo) GetHistoryByReport(ctx context.Context, reportID int64, l, o int) ([]*entities.ReportHistory, error) {
	args := m.Called(ctx, reportID, l, o)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ReportHistory), args.Error(1)
}
func (m *mockReportRepo) CreateGenerationLog(ctx context.Context, l *entities.ReportGenerationLog) error {
	args := m.Called(ctx, l)
	return args.Error(0)
}
func (m *mockReportRepo) UpdateGenerationLog(ctx context.Context, l *entities.ReportGenerationLog) error {
	return nil
}
func (m *mockReportRepo) GetGenerationLogsByReport(ctx context.Context, reportID int64) ([]*entities.ReportGenerationLog, error) {
	return nil, nil
}

// --- Mock Report Type Repository ---

type mockReportTypeRepo struct{ mock.Mock }

func (m *mockReportTypeRepo) Create(ctx context.Context, rt *entities.ReportType) error { return nil }
func (m *mockReportTypeRepo) Save(ctx context.Context, rt *entities.ReportType) error   { return nil }
func (m *mockReportTypeRepo) GetByID(ctx context.Context, id int64) (*entities.ReportType, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ReportType), args.Error(1)
}
func (m *mockReportTypeRepo) GetByCode(ctx context.Context, code string) (*entities.ReportType, error) {
	return nil, nil
}
func (m *mockReportTypeRepo) Delete(ctx context.Context, id int64) error { return nil }
func (m *mockReportTypeRepo) List(ctx context.Context, f repositories.ReportTypeFilter, l, o int) ([]*entities.ReportType, error) {
	args := m.Called(ctx, f, l, o)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ReportType), args.Error(1)
}
func (m *mockReportTypeRepo) Count(ctx context.Context, f repositories.ReportTypeFilter) (int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockReportTypeRepo) GetByCategory(ctx context.Context, c domain.ReportCategory) ([]*entities.ReportType, error) {
	return nil, nil
}
func (m *mockReportTypeRepo) GetPeriodic(ctx context.Context) ([]*entities.ReportType, error) {
	return nil, nil
}
func (m *mockReportTypeRepo) AddParameter(ctx context.Context, p *entities.ReportParameter) error {
	return nil
}
func (m *mockReportTypeRepo) UpdateParameter(ctx context.Context, p *entities.ReportParameter) error {
	return nil
}
func (m *mockReportTypeRepo) DeleteParameter(ctx context.Context, id int64) error { return nil }
func (m *mockReportTypeRepo) GetParametersByReportType(ctx context.Context, id int64) ([]*entities.ReportParameter, error) {
	return nil, nil
}
func (m *mockReportTypeRepo) AddTemplate(ctx context.Context, t *entities.ReportTemplate) error {
	return nil
}
func (m *mockReportTypeRepo) UpdateTemplate(ctx context.Context, t *entities.ReportTemplate) error {
	return nil
}
func (m *mockReportTypeRepo) DeleteTemplate(ctx context.Context, id int64) error { return nil }
func (m *mockReportTypeRepo) GetTemplatesByReportType(ctx context.Context, id int64) ([]*entities.ReportTemplate, error) {
	return nil, nil
}
func (m *mockReportTypeRepo) GetDefaultTemplate(ctx context.Context, id int64) (*entities.ReportTemplate, error) {
	return nil, nil
}
func (m *mockReportTypeRepo) SetDefaultTemplate(ctx context.Context, rtID, tID int64) error {
	return nil
}
func (m *mockReportTypeRepo) Subscribe(ctx context.Context, s *entities.ReportSubscription) error {
	return nil
}
func (m *mockReportTypeRepo) Unsubscribe(ctx context.Context, rtID, uID int64) error { return nil }
func (m *mockReportTypeRepo) GetSubscription(ctx context.Context, rtID, uID int64) (*entities.ReportSubscription, error) {
	return nil, nil
}
func (m *mockReportTypeRepo) GetSubscribersByReportType(ctx context.Context, id int64) ([]*entities.ReportSubscription, error) {
	return nil, nil
}
func (m *mockReportTypeRepo) GetSubscriptionsByUser(ctx context.Context, id int64) ([]*entities.ReportSubscription, error) {
	return nil, nil
}
func (m *mockReportTypeRepo) UpdateSubscription(ctx context.Context, s *entities.ReportSubscription) error {
	return nil
}

// --- Mock Custom Report Repository ---

type mockCustomReportRepo struct{ mock.Mock }

func (m *mockCustomReportRepo) Create(ctx context.Context, r *entities.CustomReport) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}
func (m *mockCustomReportRepo) Update(ctx context.Context, r *entities.CustomReport) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}
func (m *mockCustomReportRepo) GetByID(ctx context.Context, id uuid.UUID) (*entities.CustomReport, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.CustomReport), args.Error(1)
}
func (m *mockCustomReportRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockCustomReportRepo) List(ctx context.Context, f repositories.CustomReportFilter) ([]*entities.CustomReport, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.CustomReport), args.Error(1)
}
func (m *mockCustomReportRepo) Count(ctx context.Context, f repositories.CustomReportFilter) (int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockCustomReportRepo) GetByCreator(ctx context.Context, creatorID int64, page, pageSize int) ([]*entities.CustomReport, error) {
	args := m.Called(ctx, creatorID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.CustomReport), args.Error(1)
}
func (m *mockCustomReportRepo) GetPublicReports(ctx context.Context, page, pageSize int) ([]*entities.CustomReport, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.CustomReport), args.Error(1)
}

// --- Mock QueryBuilder ---

type mockQueryBuilder struct{ mock.Mock }

func (m *mockQueryBuilder) Execute(ctx context.Context, report *entities.CustomReport, page, pageSize int) (*entities.ReportExecutionResult, error) {
	args := m.Called(ctx, report, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ReportExecutionResult), args.Error(1)
}
func (m *mockQueryBuilder) Export(result *entities.ReportExecutionResult, options entities.ExportOptions, reportName string) ([]byte, string, error) {
	args := m.Called(result, options, reportName)
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}
func (m *mockQueryBuilder) GetAvailableFields(dataSource entities.DataSourceType) []entities.ReportField {
	args := m.Called(dataSource)
	return args.Get(0).([]entities.ReportField)
}

// --- Handler Tests with Mocks ---

func newReportHandlerWithMocks(reportRepo *mockReportRepo, typeRepo *mockReportTypeRepo) *handlers.ReportHandler {
	uc := usecases.NewReportUseCase(reportRepo, typeRepo, nil, nil, nil)
	return handlers.NewReportHandler(uc)
}

func newCustomReportHandlerWithMocks(repo *mockCustomReportRepo, qb *mockQueryBuilder) *handlers.CustomReportHandler {
	uc := usecases.NewCustomReportUseCase(repo, qb)
	return handlers.NewCustomReportHandler(uc)
}

// --- ReportHandler Tests ---

func TestReportHandler_Create_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	rt := &entities.ReportType{ID: 1, Name: "Test Type", CreatedAt: now, UpdatedAt: now}
	typeRepo.On("GetByID", mock.Anything, int64(1)).Return(rt, nil)
	reportRepo.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		r := args.Get(1).(*entities.Report)
		r.ID = 1
	}).Return(nil)
	reportRepo.On("AddHistory", mock.Anything, mock.Anything).Return(nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.POST("/reports", withAuthUser(1, "methodist"), h.Create)

	w := doRequest(router, http.MethodPost, "/reports", map[string]interface{}{
		"title":          "Test Report",
		"report_type_id": 1,
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestReportHandler_Create_WithDescriptionSanitized(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	rt := &entities.ReportType{ID: 1, Name: "Test", CreatedAt: now, UpdatedAt: now}
	typeRepo.On("GetByID", mock.Anything, int64(1)).Return(rt, nil)
	reportRepo.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		r := args.Get(1).(*entities.Report)
		r.ID = 1
	}).Return(nil)
	reportRepo.On("AddHistory", mock.Anything, mock.Anything).Return(nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.POST("/reports", withAuthUser(1, "methodist"), h.Create)

	w := doRequest(router, http.MethodPost, "/reports", map[string]interface{}{
		"title":          "Report",
		"report_type_id": 1,
		"description":    "<script>alert('xss')</script>Desc",
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestReportHandler_Create_UsecaseError(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	typeRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, usecases.ErrReportTypeNotFound)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.POST("/reports", withAuthUser(1, "methodist"), h.Create)

	w := doRequest(router, http.MethodPost, "/reports", map[string]interface{}{
		"title":          "Report",
		"report_type_id": 1,
	})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestReportHandler_GetByID_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	report := &entities.Report{ID: 1, Title: "Test", AuthorID: 1, Status: domain.ReportStatusDraft, ReportTypeID: 1, CreatedAt: now, UpdatedAt: now}
	reportRepo.On("GetByID", mock.Anything, int64(1)).Return(report, nil)
	reportRepo.On("HasAccess", mock.Anything, int64(1), int64(1), mock.Anything).Return(true, nil)
	typeRepo.On("GetByID", mock.Anything, int64(1)).Return(&entities.ReportType{ID: 1, Name: "T", CreatedAt: now, UpdatedAt: now}, nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.GET("/reports/:id", withAuthUser(1, "methodist"), h.GetByID)

	w := doRequest(router, http.MethodGet, "/reports/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_GetByID_NotFound(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	reportRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, usecases.ErrReportNotFound)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.GET("/reports/:id", withAuthUser(1, "methodist"), h.GetByID)

	w := doRequest(router, http.MethodGet, "/reports/1", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestReportHandler_Update_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	report := &entities.Report{ID: 1, Title: "Old", AuthorID: 1, Status: domain.ReportStatusDraft, CreatedAt: now, UpdatedAt: now}
	reportRepo.On("GetByID", mock.Anything, int64(1)).Return(report, nil)
	reportRepo.On("HasAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	reportRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
	reportRepo.On("AddHistory", mock.Anything, mock.Anything).Return(nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.PUT("/reports/:id", withAuthUser(1, "methodist"), h.Update)

	w := doRequest(router, http.MethodPut, "/reports/1", map[string]interface{}{
		"title":       "New Title",
		"description": "<b>Updated</b>",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_Delete_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	report := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusDraft, CreatedAt: now, UpdatedAt: now}
	reportRepo.On("GetByID", mock.Anything, int64(1)).Return(report, nil)
	reportRepo.On("HasAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	reportRepo.On("Delete", mock.Anything, int64(1)).Return(nil)
	reportRepo.On("AddHistory", mock.Anything, mock.Anything).Return(nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.DELETE("/reports/:id", withAuthUser(1, "methodist"), h.Delete)

	w := doRequest(router, http.MethodDelete, "/reports/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_List_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	reportRepo.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.Report{}, nil)
	reportRepo.On("Count", mock.Anything, mock.Anything).Return(int64(0), nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.GET("/reports", withAuthUser(1, "methodist"), h.List)

	w := doRequest(router, http.MethodGet, "/reports", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_List_Error(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	reportRepo.On("Count", mock.Anything, mock.Anything).Return(int64(0), assert.AnError)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.GET("/reports", withAuthUser(1, "methodist"), h.List)

	w := doRequest(router, http.MethodGet, "/reports", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestReportHandler_Generate_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	report := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusDraft, CreatedAt: now, UpdatedAt: now}
	reportRepo.On("GetByID", mock.Anything, int64(1)).Return(report, nil)
	reportRepo.On("HasAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	reportRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
	reportRepo.On("CreateGenerationLog", mock.Anything, mock.Anything).Return(nil)
	reportRepo.On("AddHistory", mock.Anything, mock.Anything).Return(nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.POST("/reports/:id/generate", withAuthUser(1, "methodist"), h.Generate)

	w := doRequest(router, http.MethodPost, "/reports/1/generate", nil)
	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestReportHandler_SubmitForReview_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	report := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusReady, CreatedAt: now, UpdatedAt: now}
	reportRepo.On("GetByID", mock.Anything, int64(1)).Return(report, nil)
	reportRepo.On("HasAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	reportRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
	reportRepo.On("AddHistory", mock.Anything, mock.Anything).Return(nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.POST("/reports/:id/submit", withAuthUser(1, "methodist"), h.SubmitForReview)

	w := doRequest(router, http.MethodPost, "/reports/1/submit", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_Review_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	report := &entities.Report{ID: 1, AuthorID: 2, Status: domain.ReportStatusReviewing, CreatedAt: now, UpdatedAt: now}
	reportRepo.On("GetByID", mock.Anything, int64(1)).Return(report, nil)
	reportRepo.On("HasAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	reportRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
	reportRepo.On("AddHistory", mock.Anything, mock.Anything).Return(nil)
	reportRepo.On("AddComment", mock.Anything, mock.Anything).Return(nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.POST("/reports/:id/review", withAuthUser(1, "methodist"), h.Review)

	w := doRequest(router, http.MethodPost, "/reports/1/review", map[string]interface{}{
		"action":  "approve",
		"comment": "Looks good",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_Publish_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	report := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusApproved, CreatedAt: now, UpdatedAt: now}
	reportRepo.On("GetByID", mock.Anything, int64(1)).Return(report, nil)
	reportRepo.On("HasAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	reportRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
	reportRepo.On("AddHistory", mock.Anything, mock.Anything).Return(nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.POST("/reports/:id/publish", withAuthUser(1, "methodist"), h.Publish)

	w := doRequest(router, http.MethodPost, "/reports/1/publish", map[string]interface{}{
		"is_public": true,
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_AddAccess_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	report := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusDraft, CreatedAt: now, UpdatedAt: now}
	reportRepo.On("GetByID", mock.Anything, int64(1)).Return(report, nil)
	reportRepo.On("HasAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	reportRepo.On("AddAccess", mock.Anything, mock.Anything).Return(nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.POST("/reports/:id/access", withAuthUser(1, "methodist"), h.AddAccess)

	userID := int64(2)
	w := doRequest(router, http.MethodPost, "/reports/1/access", map[string]interface{}{
		"user_id":    userID,
		"permission": "read",
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestReportHandler_RemoveAccess_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	report := &entities.Report{ID: 1, AuthorID: 1, Status: domain.ReportStatusDraft, CreatedAt: now, UpdatedAt: now}
	reportRepo.On("GetByID", mock.Anything, int64(1)).Return(report, nil)
	reportRepo.On("HasAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	reportRepo.On("RemoveAccess", mock.Anything, int64(1), int64(5)).Return(nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.DELETE("/reports/:id/access/:access_id", withAuthUser(1, "methodist"), h.RemoveAccess)

	w := doRequest(router, http.MethodDelete, "/reports/1/access/5", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_GetAccess_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	report := &entities.Report{ID: 1, AuthorID: 1, CreatedAt: now, UpdatedAt: now}
	reportRepo.On("GetByID", mock.Anything, int64(1)).Return(report, nil)
	reportRepo.On("HasAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	reportRepo.On("GetAccessByReport", mock.Anything, int64(1)).Return([]*entities.ReportAccess{}, nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.GET("/reports/:id/access", withAuthUser(1, "methodist"), h.GetAccess)

	w := doRequest(router, http.MethodGet, "/reports/1/access", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_AddComment_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	report := &entities.Report{ID: 1, AuthorID: 1, CreatedAt: now, UpdatedAt: now}
	reportRepo.On("GetByID", mock.Anything, int64(1)).Return(report, nil)
	reportRepo.On("HasAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	reportRepo.On("AddComment", mock.Anything, mock.Anything).Return(nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.POST("/reports/:id/comments", withAuthUser(1, "methodist"), h.AddComment)

	w := doRequest(router, http.MethodPost, "/reports/1/comments", map[string]interface{}{
		"content": "Nice report",
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestReportHandler_GetComments_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	report := &entities.Report{ID: 1, AuthorID: 1, CreatedAt: now, UpdatedAt: now}
	reportRepo.On("GetByID", mock.Anything, int64(1)).Return(report, nil)
	reportRepo.On("HasAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	reportRepo.On("GetCommentsByReport", mock.Anything, int64(1)).Return([]*entities.ReportComment{}, nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.GET("/reports/:id/comments", withAuthUser(1, "methodist"), h.GetComments)

	w := doRequest(router, http.MethodGet, "/reports/1/comments", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_GetHistory_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	report := &entities.Report{ID: 1, AuthorID: 1, CreatedAt: now, UpdatedAt: now}
	reportRepo.On("GetByID", mock.Anything, int64(1)).Return(report, nil)
	reportRepo.On("HasAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	reportRepo.On("GetHistoryByReport", mock.Anything, int64(1), 50, 0).Return([]*entities.ReportHistory{}, nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.GET("/reports/:id/history", withAuthUser(1, "methodist"), h.GetHistory)

	w := doRequest(router, http.MethodGet, "/reports/1/history", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_GetHistory_WithParams(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	report := &entities.Report{ID: 1, AuthorID: 1, CreatedAt: now, UpdatedAt: now}
	reportRepo.On("GetByID", mock.Anything, int64(1)).Return(report, nil)
	reportRepo.On("HasAccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	reportRepo.On("GetHistoryByReport", mock.Anything, int64(1), 10, 5).Return([]*entities.ReportHistory{}, nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.GET("/reports/:id/history", withAuthUser(1, "methodist"), h.GetHistory)

	w := doRequest(router, http.MethodGet, "/reports/1/history?limit=10&offset=5", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_GetReportTypes_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	types := []*entities.ReportType{{ID: 1, Name: "Test", CreatedAt: now, UpdatedAt: now}}
	typeRepo.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(types, nil)
	typeRepo.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.GET("/report-types", h.GetReportTypes)

	w := doRequest(router, http.MethodGet, "/report-types", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_GetReportTypeByID_Success(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	now := time.Now()
	rt := &entities.ReportType{ID: 1, Name: "Test", CreatedAt: now, UpdatedAt: now}
	typeRepo.On("GetByID", mock.Anything, int64(1)).Return(rt, nil)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.GET("/report-types/:id", h.GetReportTypeByID)

	w := doRequest(router, http.MethodGet, "/report-types/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportHandler_GetReportTypeByID_NotFound(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	typeRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, usecases.ErrReportTypeNotFound)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.GET("/report-types/:id", h.GetReportTypeByID)

	w := doRequest(router, http.MethodGet, "/report-types/999", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestReportHandler_handleError_AllBranches(t *testing.T) {
	typeRepo := new(mockReportTypeRepo)

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"not found", usecases.ErrReportNotFound, http.StatusNotFound},
		{"type not found", usecases.ErrReportTypeNotFound, http.StatusNotFound},
		{"unauthorized", usecases.ErrUnauthorized, http.StatusForbidden},
		{"cannot modify", usecases.ErrCannotModifyReport, http.StatusBadRequest},
		{"invalid input", usecases.ErrInvalidInput, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockReportRepo)
			repo.On("GetByID", mock.Anything, int64(1)).Return(nil, tt.err)

			h := newReportHandlerWithMocks(repo, typeRepo)
			router := newRouter()
			router.GET("/reports/:id", withAuthUser(1, "methodist"), h.GetByID)

			w := doRequest(router, http.MethodGet, "/reports/1", nil)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	// default case - generic error via List's Count
	t.Run("default error", func(t *testing.T) {
		repo := new(mockReportRepo)
		repo.On("Count", mock.Anything, mock.Anything).Return(int64(0), assert.AnError)

		h := newReportHandlerWithMocks(repo, typeRepo)
		router := newRouter()
		router.GET("/reports", withAuthUser(1, "methodist"), h.List)

		w := doRequest(router, http.MethodGet, "/reports", nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// --- CustomReportHandler Tests with Mocks ---

func TestCustomReportHandler_GetByID_NotFoundError(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	repo.On("GetByID", mock.Anything, mock.Anything).Return(nil, nil)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.GET("/reports/custom/:id", withAuthUser(1, "methodist"), h.GetByID)

	w := doRequest(router, http.MethodGet, "/reports/custom/550e8400-e29b-41d4-a716-446655440000", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCustomReportHandler_GetByID_Forbidden(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	report := entities.NewCustomReport("Test", "", entities.DataSourceDocuments, 999)
	report.SetFields([]entities.SelectedField{{Field: entities.ReportField{ID: "id"}, Order: 1}})
	report.SetPublic(false)
	repo.On("GetByID", mock.Anything, mock.Anything).Return(report, nil)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.GET("/reports/custom/:id", withAuthUser(1, "methodist"), h.GetByID)

	w := doRequest(router, http.MethodGet, "/reports/custom/"+report.ID.String(), nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCustomReportHandler_Update_Success(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	report := entities.NewCustomReport("Test", "", entities.DataSourceDocuments, 1)
	report.SetFields([]entities.SelectedField{{Field: entities.ReportField{ID: "id"}, Order: 1}})
	repo.On("GetByID", mock.Anything, mock.Anything).Return(report, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.PUT("/reports/custom/:id", withAuthUser(1, "methodist"), h.Update)

	w := doRequest(router, http.MethodPut, "/reports/custom/"+report.ID.String(), map[string]interface{}{
		"name": "Updated",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCustomReportHandler_Delete_Success(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	report := entities.NewCustomReport("Test", "", entities.DataSourceDocuments, 1)
	report.SetFields([]entities.SelectedField{{Field: entities.ReportField{ID: "id"}, Order: 1}})
	repo.On("GetByID", mock.Anything, mock.Anything).Return(report, nil)
	repo.On("Delete", mock.Anything, mock.Anything).Return(nil)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.DELETE("/reports/custom/:id", withAuthUser(1, "methodist"), h.Delete)

	w := doRequest(router, http.MethodDelete, "/reports/custom/"+report.ID.String(), nil)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCustomReportHandler_List_Success(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	repo.On("List", mock.Anything, mock.Anything).Return([]*entities.CustomReport{}, nil)
	repo.On("Count", mock.Anything, mock.Anything).Return(int64(0), nil)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.GET("/reports/custom", withAuthUser(1, "methodist"), h.List)

	w := doRequest(router, http.MethodGet, "/reports/custom?page=1&pageSize=10", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCustomReportHandler_Execute_Success(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	report := entities.NewCustomReport("Test", "", entities.DataSourceDocuments, 1)
	report.SetFields([]entities.SelectedField{{Field: entities.ReportField{ID: "id"}, Order: 1}})
	repo.On("GetByID", mock.Anything, mock.Anything).Return(report, nil)

	result := &entities.ReportExecutionResult{
		Columns:    []entities.ReportColumn{{Key: "id", Label: "ID"}},
		Rows:       []map[string]interface{}{{"id": 1}},
		TotalCount: 1, Page: 1, PageSize: 50, TotalPages: 1,
	}
	qb.On("Execute", mock.Anything, mock.Anything, 1, 50).Return(result, nil)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.POST("/reports/custom/:id/execute", withAuthUser(1, "methodist"), h.Execute)

	w := doRequest(router, http.MethodPost, "/reports/custom/"+report.ID.String()+"/execute", map[string]interface{}{
		"page": 1, "pageSize": 50,
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCustomReportHandler_Execute_WithQueryParamsMocked(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	report := entities.NewCustomReport("Test", "", entities.DataSourceDocuments, 1)
	report.SetFields([]entities.SelectedField{{Field: entities.ReportField{ID: "id"}, Order: 1}})
	repo.On("GetByID", mock.Anything, mock.Anything).Return(report, nil)

	result := &entities.ReportExecutionResult{
		Columns: []entities.ReportColumn{{Key: "id", Label: "ID"}}, Rows: []map[string]interface{}{},
		TotalCount: 0, Page: 2, PageSize: 10, TotalPages: 0,
	}
	qb.On("Execute", mock.Anything, mock.Anything, 2, 10).Return(result, nil)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.POST("/reports/custom/:id/execute", withAuthUser(1, "methodist"), h.Execute)

	w := doRequest(router, http.MethodPost, "/reports/custom/"+report.ID.String()+"/execute?page=2&pageSize=10", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCustomReportHandler_Export_Success(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	report := entities.NewCustomReport("Test Report", "", entities.DataSourceDocuments, 1)
	report.SetFields([]entities.SelectedField{{Field: entities.ReportField{ID: "id"}, Order: 1}})
	repo.On("GetByID", mock.Anything, mock.Anything).Return(report, nil)

	result := &entities.ReportExecutionResult{
		Columns: []entities.ReportColumn{{Key: "id", Label: "ID"}}, Rows: []map[string]interface{}{},
	}
	qb.On("Execute", mock.Anything, mock.Anything, 1, 10000).Return(result, nil)
	qb.On("Export", mock.Anything, mock.Anything, mock.Anything).Return([]byte("data"), "report.csv", nil)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.POST("/reports/custom/:id/export", withAuthUser(1, "methodist"), h.Export)

	w := doRequest(router, http.MethodPost, "/reports/custom/"+report.ID.String()+"/export", map[string]interface{}{
		"format": "csv",
	})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/csv")
}

func TestCustomReportHandler_Export_PDF(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	report := entities.NewCustomReport("Test", "", entities.DataSourceDocuments, 1)
	report.SetFields([]entities.SelectedField{{Field: entities.ReportField{ID: "id"}, Order: 1}})
	repo.On("GetByID", mock.Anything, mock.Anything).Return(report, nil)
	result := &entities.ReportExecutionResult{}
	qb.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(result, nil)
	qb.On("Export", mock.Anything, mock.Anything, mock.Anything).Return([]byte("pdf"), "report.pdf", nil)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.POST("/reports/custom/:id/export", withAuthUser(1, "methodist"), h.Export)

	w := doRequest(router, http.MethodPost, "/reports/custom/"+report.ID.String()+"/export", map[string]interface{}{
		"format": "pdf",
	})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/pdf")
}

func TestCustomReportHandler_Export_XLSX(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	report := entities.NewCustomReport("Test", "", entities.DataSourceDocuments, 1)
	report.SetFields([]entities.SelectedField{{Field: entities.ReportField{ID: "id"}, Order: 1}})
	repo.On("GetByID", mock.Anything, mock.Anything).Return(report, nil)
	result := &entities.ReportExecutionResult{}
	qb.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(result, nil)
	qb.On("Export", mock.Anything, mock.Anything, mock.Anything).Return([]byte("xlsx"), "report.xlsx", nil)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.POST("/reports/custom/:id/export", withAuthUser(1, "methodist"), h.Export)

	w := doRequest(router, http.MethodPost, "/reports/custom/"+report.ID.String()+"/export", map[string]interface{}{
		"format": "xlsx",
	})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "spreadsheet")
}

func TestCustomReportHandler_GetMyReports_Success(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	repo.On("Count", mock.Anything, mock.Anything).Return(int64(0), nil)
	repo.On("List", mock.Anything, mock.Anything).Return([]*entities.CustomReport{}, nil)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.GET("/reports/custom/my", withAuthUser(1, "methodist"), h.GetMyReports)

	w := doRequest(router, http.MethodGet, "/reports/custom/my", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCustomReportHandler_GetPublicReports_Success(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	repo.On("Count", mock.Anything, mock.Anything).Return(int64(0), nil)
	repo.On("List", mock.Anything, mock.Anything).Return([]*entities.CustomReport{}, nil)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.GET("/reports/custom/public", h.GetPublicReports)

	w := doRequest(router, http.MethodGet, "/reports/custom/public", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCustomReportHandler_handleError_AllBranches(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"not found", usecases.ErrCustomReportNotFound, http.StatusNotFound},
		{"unauthorized", usecases.ErrUnauthorizedAccess, http.StatusForbidden},
		{"invalid data source", usecases.ErrInvalidDataSource, http.StatusBadRequest},
		{"invalid fields", usecases.ErrInvalidFields, http.StatusBadRequest},
		{"default", assert.AnError, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockCustomReportRepo)
			qb := new(mockQueryBuilder)

			repo.On("GetByID", mock.Anything, mock.Anything).Return(nil, tt.err)

			h := newCustomReportHandlerWithMocks(repo, qb)
			router := newRouter()
			router.GET("/reports/custom/:id", withAuthUser(1, "methodist"), h.GetByID)

			w := doRequest(router, http.MethodGet, "/reports/custom/550e8400-e29b-41d4-a716-446655440000", nil)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestCustomReportHandler_List_WithDefaults(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	repo.On("Count", mock.Anything, mock.Anything).Return(int64(0), nil)
	repo.On("List", mock.Anything, mock.Anything).Return([]*entities.CustomReport{}, nil)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.GET("/reports/custom", withAuthUser(1, "methodist"), h.List)

	// Without page params -> uses defaults
	w := doRequest(router, http.MethodGet, "/reports/custom", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCustomReportHandler_List_Error(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	repo.On("Count", mock.Anything, mock.Anything).Return(int64(0), assert.AnError)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.GET("/reports/custom", withAuthUser(1, "methodist"), h.List)

	w := doRequest(router, http.MethodGet, "/reports/custom", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCustomReportHandler_GetPublicReports_Error(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	repo.On("Count", mock.Anything, mock.Anything).Return(int64(0), assert.AnError)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.GET("/reports/custom/public", h.GetPublicReports)

	w := doRequest(router, http.MethodGet, "/reports/custom/public", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCustomReportHandler_GetMyReports_Error(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	repo.On("Count", mock.Anything, mock.Anything).Return(int64(0), assert.AnError)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.GET("/reports/custom/my", withAuthUser(1, "methodist"), h.GetMyReports)

	w := doRequest(router, http.MethodGet, "/reports/custom/my", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCustomReportHandler_Create_Success(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	repo.On("Create", mock.Anything, mock.Anything).Return(nil)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.POST("/reports/custom", withAuthUser(1, "methodist"), h.Create)

	w := doRequest(router, http.MethodPost, "/reports/custom", map[string]interface{}{
		"name":       "Report",
		"dataSource": "documents",
		"fields":     []map[string]interface{}{{"fieldKey": "id", "displayName": "ID", "order": 1}},
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCustomReportHandler_Export_DefaultFormat(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	report := entities.NewCustomReport("Test", "", entities.DataSourceDocuments, 1)
	report.SetFields([]entities.SelectedField{{Field: entities.ReportField{ID: "id"}, Order: 1}})
	repo.On("GetByID", mock.Anything, mock.Anything).Return(report, nil)
	result := &entities.ReportExecutionResult{}
	qb.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(result, nil)
	qb.On("Export", mock.Anything, mock.Anything, mock.Anything).Return([]byte("data"), "report.xlsx", nil)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.POST("/reports/custom/:id/export", withAuthUser(1, "methodist"), h.Export)

	// No format specified -> defaults to xlsx
	w := doRequest(router, http.MethodPost, "/reports/custom/"+report.ID.String()+"/export", map[string]interface{}{
		"includeHeaders": true,
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCustomReportHandler_Export_Error(t *testing.T) {
	repo := new(mockCustomReportRepo)
	qb := new(mockQueryBuilder)

	repo.On("GetByID", mock.Anything, mock.Anything).Return(nil, nil)

	h := newCustomReportHandlerWithMocks(repo, qb)
	router := newRouter()
	router.POST("/reports/custom/:id/export", withAuthUser(1, "methodist"), h.Export)

	w := doRequest(router, http.MethodPost, "/reports/custom/550e8400-e29b-41d4-a716-446655440000/export", map[string]interface{}{
		"format": "csv",
	})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestReportHandler_GetReportTypes_Error(t *testing.T) {
	reportRepo := new(mockReportRepo)
	typeRepo := new(mockReportTypeRepo)

	typeRepo.On("Count", mock.Anything, mock.Anything).Return(int64(0), assert.AnError)

	h := newReportHandlerWithMocks(reportRepo, typeRepo)
	router := newRouter()
	router.GET("/report-types", h.GetReportTypes)

	w := doRequest(router, http.MethodGet, "/report-types", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCustomReportHandler_GetAvailableFields_HasAllSources(t *testing.T) {
	h := newCustomReportHandlerUnit()
	router := newRouter()
	router.GET("/reports/custom/fields", h.GetAvailableFields)

	w := doRequest(router, http.MethodGet, "/reports/custom/fields", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	for _, source := range []string{"documents", "users", "events", "tasks", "students"} {
		_, ok := resp[source]
		assert.True(t, ok, "should have fields for "+source)
	}
}
