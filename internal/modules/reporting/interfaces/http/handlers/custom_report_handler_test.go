package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/infrastructure/persistence"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/interfaces/http/handlers"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/testing/helpers"
	testSuite "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/testing/suite"
)

// MockQueryBuilder implements QueryBuilder interface for testing
type MockQueryBuilder struct{}

func (m *MockQueryBuilder) Execute(ctx context.Context, report *entities.CustomReport, page, pageSize int) (*entities.ReportExecutionResult, error) {
	return &entities.ReportExecutionResult{
		Columns: []entities.ReportColumn{
			{Key: "id", Label: "ID"},
			{Key: "name", Label: "Name"},
		},
		Rows: []map[string]interface{}{
			{"id": 1, "name": "Test"},
		},
		TotalCount: 1,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 1,
	}, nil
}

func (m *MockQueryBuilder) Export(result *entities.ReportExecutionResult, options entities.ExportOptions, reportName string) ([]byte, string, error) {
	return []byte("test,data"), "report.csv", nil
}

func (m *MockQueryBuilder) GetAvailableFields(dataSource entities.DataSourceType) []entities.ReportField {
	return []entities.ReportField{
		{ID: "id", Name: "id", Label: "ID", Type: entities.FieldTypeNumber, Source: dataSource},
		{ID: "name", Name: "name", Label: "Name", Type: entities.FieldTypeString, Source: dataSource},
	}
}

// mockAuthMiddleware creates a middleware that sets user_id for testing
func mockAuthMiddleware(userID int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

// testSelectedField creates a SelectedField for testing
func testSelectedField(id, name, label string, order int) entities.SelectedField {
	return entities.SelectedField{
		Field: entities.ReportField{
			ID:    id,
			Name:  name,
			Label: label,
			Type:  entities.FieldTypeString,
		},
		Order: order,
	}
}

// CustomReportHandlerTestSuite tests custom report HTTP handlers
type CustomReportHandlerTestSuite struct {
	testSuite.IntegrationSuite
	handler *handlers.CustomReportHandler
	usecase *usecases.CustomReportUseCase
	router  *gin.Engine
	userID  int64
}

// SetupSuite runs once before all tests
func (s *CustomReportHandlerTestSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()

	s.userID = 1

	// Setup dependencies
	repo := persistence.NewCustomReportRepositoryPG(s.DB)
	queryBuilder := &MockQueryBuilder{}
	s.usecase = usecases.NewCustomReportUseCase(repo, queryBuilder)
	s.handler = handlers.NewCustomReportHandler(s.usecase)

	// Setup router
	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	// Apply mock auth middleware
	authGroup := s.router.Group("/api/custom-reports")
	authGroup.Use(mockAuthMiddleware(s.userID, "methodist"))
	{
		authGroup.GET("", s.handler.List)
		authGroup.POST("", s.handler.Create)
		authGroup.GET("/:id", s.handler.GetByID)
		authGroup.PUT("/:id", s.handler.Update)
		authGroup.DELETE("/:id", s.handler.Delete)
		authGroup.POST("/:id/execute", s.handler.Execute)
		authGroup.POST("/:id/export", s.handler.Export)
		authGroup.GET("/my", s.handler.GetMyReports)
		authGroup.GET("/public", s.handler.GetPublicReports)
		authGroup.GET("/available-fields/:dataSource", s.handler.GetAvailableFields)
	}
}

// TearDownTest runs after each test
func (s *CustomReportHandlerTestSuite) TearDownTest() {
	s.TruncateTables("custom_reports")
}

// TestCreateReport tests report creation endpoint
func (s *CustomReportHandlerTestSuite) TestCreateReport() {
	payload := map[string]interface{}{
		"name":        "Test Report",
		"description": "Test Description",
		"data_source": "documents",
		"fields": []map[string]interface{}{
			{"field_key": "id", "display_name": "ID", "order": 1},
			{"field_key": "name", "display_name": "Name", "order": 2},
		},
		"is_public": false,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/custom-reports", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("success", response["status"])

	data, ok := response["data"].(map[string]interface{})
	s.True(ok)
	s.Equal("Test Report", data["name"])
	s.NotEmpty(data["id"])
}

// TestCreateReportInvalidDataSource tests creation with invalid data source
func (s *CustomReportHandlerTestSuite) TestCreateReportInvalidDataSource() {
	payload := map[string]interface{}{
		"name":        "Test Report",
		"data_source": "invalid_source",
		"fields": []map[string]interface{}{
			{"field_key": "id", "display_name": "ID", "order": 1},
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/custom-reports", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

// TestCreateReportWithoutFields tests creation without fields
func (s *CustomReportHandlerTestSuite) TestCreateReportWithoutFields() {
	payload := map[string]interface{}{
		"name":        "Test Report",
		"data_source": "documents",
		"fields":      []map[string]interface{}{},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/custom-reports", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

// TestGetReportByID tests getting a report by ID
func (s *CustomReportHandlerTestSuite) TestGetReportByID() {
	ctx := helpers.TestContext(s.T())

	// Create a report first
	report := entities.NewCustomReport("Test Report", "Description", entities.DataSourceDocuments, s.userID)
	report.SetFields([]entities.SelectedField{testSelectedField("id", "id", "ID", 1)})
	repo := persistence.NewCustomReportRepositoryPG(s.DB)
	err := repo.Create(ctx, report)
	s.NoError(err)

	// Get the report
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/custom-reports/%s", report.ID.String()), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("success", response["status"])

	data, ok := response["data"].(map[string]interface{})
	s.True(ok)
	s.Equal("Test Report", data["name"])
}

// TestGetReportByIDNotFound tests getting a non-existent report
func (s *CustomReportHandlerTestSuite) TestGetReportByIDNotFound() {
	randomID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/custom-reports/%s", randomID.String()), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

// TestUpdateReport tests updating a report
func (s *CustomReportHandlerTestSuite) TestUpdateReport() {
	ctx := helpers.TestContext(s.T())

	// Create a report first
	report := entities.NewCustomReport("Original Name", "Description", entities.DataSourceDocuments, s.userID)
	report.SetFields([]entities.SelectedField{testSelectedField("id", "id", "ID", 1)})
	repo := persistence.NewCustomReportRepositoryPG(s.DB)
	err := repo.Create(ctx, report)
	s.NoError(err)

	// Update the report
	payload := map[string]interface{}{
		"name":        "Updated Name",
		"description": "Updated Description",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/custom-reports/%s", report.ID.String()), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("success", response["status"])

	data, ok := response["data"].(map[string]interface{})
	s.True(ok)
	s.Equal("Updated Name", data["name"])
	s.Equal("Updated Description", data["description"])
}

// TestDeleteReport tests deleting a report
func (s *CustomReportHandlerTestSuite) TestDeleteReport() {
	ctx := helpers.TestContext(s.T())

	// Create a report first
	report := entities.NewCustomReport("To Delete", "Description", entities.DataSourceDocuments, s.userID)
	report.SetFields([]entities.SelectedField{testSelectedField("id", "id", "ID", 1)})
	repo := persistence.NewCustomReportRepositoryPG(s.DB)
	err := repo.Create(ctx, report)
	s.NoError(err)

	// Delete the report
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/custom-reports/%s", report.ID.String()), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Verify deletion
	deleted, err := repo.GetByID(ctx, report.ID)
	s.NoError(err)
	s.Nil(deleted)
}

// TestListReports tests listing reports
func (s *CustomReportHandlerTestSuite) TestListReports() {
	ctx := helpers.TestContext(s.T())

	// Create a few reports
	repo := persistence.NewCustomReportRepositoryPG(s.DB)
	for i := 0; i < 3; i++ {
		report := entities.NewCustomReport(fmt.Sprintf("Report %d", i), "", entities.DataSourceDocuments, s.userID)
		report.SetFields([]entities.SelectedField{testSelectedField("id", "id", "ID", 1)})
		err := repo.Create(ctx, report)
		s.NoError(err)
	}

	// List reports
	req := httptest.NewRequest(http.MethodGet, "/api/custom-reports?page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("success", response["status"])

	data, ok := response["data"].(map[string]interface{})
	s.True(ok)
	reports, ok := data["reports"].([]interface{})
	s.True(ok)
	s.Equal(3, len(reports))
}

// TestExecuteReport tests executing a report
func (s *CustomReportHandlerTestSuite) TestExecuteReport() {
	ctx := helpers.TestContext(s.T())

	// Create a report first
	report := entities.NewCustomReport("Test Report", "Description", entities.DataSourceDocuments, s.userID)
	report.SetFields([]entities.SelectedField{testSelectedField("id", "id", "ID", 1)})
	repo := persistence.NewCustomReportRepositoryPG(s.DB)
	err := repo.Create(ctx, report)
	s.NoError(err)

	// Execute the report
	payload := map[string]interface{}{
		"page":      1,
		"page_size": 50,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/custom-reports/%s/execute", report.ID.String()), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("success", response["status"])

	data, ok := response["data"].(map[string]interface{})
	s.True(ok)
	s.NotNil(data["columns"])
	s.NotNil(data["rows"])
}

// TestExportReport tests exporting a report
func (s *CustomReportHandlerTestSuite) TestExportReport() {
	ctx := helpers.TestContext(s.T())

	// Create a report first
	report := entities.NewCustomReport("Test Report", "Description", entities.DataSourceDocuments, s.userID)
	report.SetFields([]entities.SelectedField{testSelectedField("id", "id", "ID", 1)})
	repo := persistence.NewCustomReportRepositoryPG(s.DB)
	err := repo.Create(ctx, report)
	s.NoError(err)

	// Export the report
	payload := map[string]interface{}{
		"format":          "csv",
		"include_headers": true,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/custom-reports/%s/export", report.ID.String()), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Header().Get("Content-Disposition"), "attachment")
}

// TestGetMyReports tests getting user's own reports
func (s *CustomReportHandlerTestSuite) TestGetMyReports() {
	ctx := helpers.TestContext(s.T())

	// Create reports for the user
	repo := persistence.NewCustomReportRepositoryPG(s.DB)
	report := entities.NewCustomReport("My Report", "", entities.DataSourceDocuments, s.userID)
	report.SetFields([]entities.SelectedField{testSelectedField("id", "id", "ID", 1)})
	err := repo.Create(ctx, report)
	s.NoError(err)

	// Get my reports
	req := httptest.NewRequest(http.MethodGet, "/api/custom-reports/my?page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("success", response["status"])
}

// TestGetPublicReports tests getting public reports
func (s *CustomReportHandlerTestSuite) TestGetPublicReports() {
	ctx := helpers.TestContext(s.T())

	// Create a public report
	repo := persistence.NewCustomReportRepositoryPG(s.DB)
	report := entities.NewCustomReport("Public Report", "", entities.DataSourceDocuments, s.userID)
	report.SetFields([]entities.SelectedField{testSelectedField("id", "id", "ID", 1)})
	report.SetPublic(true)
	err := repo.Create(ctx, report)
	s.NoError(err)

	// Get public reports
	req := httptest.NewRequest(http.MethodGet, "/api/custom-reports/public?page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("success", response["status"])
}

// TestGetAvailableFields tests getting available fields for a data source
func (s *CustomReportHandlerTestSuite) TestGetAvailableFields() {
	req := httptest.NewRequest(http.MethodGet, "/api/custom-reports/available-fields/documents", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("success", response["status"])

	data, ok := response["data"].(map[string]interface{})
	s.True(ok)
	fields, ok := data["fields"].([]interface{})
	s.True(ok)
	s.Greater(len(fields), 0)
}

// TestGetAvailableFieldsInvalidDataSource tests getting fields for invalid data source
func (s *CustomReportHandlerTestSuite) TestGetAvailableFieldsInvalidDataSource() {
	req := httptest.NewRequest(http.MethodGet, "/api/custom-reports/available-fields/invalid_source", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

// TestAccessPrivateReportByOtherUser tests that users cannot access private reports of others
func (s *CustomReportHandlerTestSuite) TestAccessPrivateReportByOtherUser() {
	ctx := helpers.TestContext(s.T())

	// Create a private report by another user
	otherUserID := int64(999)
	report := entities.NewCustomReport("Private Report", "", entities.DataSourceDocuments, otherUserID)
	report.SetFields([]entities.SelectedField{testSelectedField("id", "id", "ID", 1)})
	report.SetPublic(false)
	repo := persistence.NewCustomReportRepositoryPG(s.DB)
	err := repo.Create(ctx, report)
	s.NoError(err)

	// Try to access the private report
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/custom-reports/%s", report.ID.String()), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusForbidden, w.Code)
}

// TestAccessPublicReportByOtherUser tests that users can access public reports of others
func (s *CustomReportHandlerTestSuite) TestAccessPublicReportByOtherUser() {
	ctx := helpers.TestContext(s.T())

	// Create a public report by another user
	otherUserID := int64(999)
	report := entities.NewCustomReport("Public Report", "", entities.DataSourceDocuments, otherUserID)
	report.SetFields([]entities.SelectedField{testSelectedField("id", "id", "ID", 1)})
	report.SetPublic(true)
	repo := persistence.NewCustomReportRepositoryPG(s.DB)
	err := repo.Create(ctx, report)
	s.NoError(err)

	// Access the public report
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/custom-reports/%s", report.ID.String()), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("success", response["status"])
}

// TestSuite runs the test suite
func TestCustomReportHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(CustomReportHandlerTestSuite))
}
