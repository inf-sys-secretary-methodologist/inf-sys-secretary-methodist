package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

// CustomReportHandler handles HTTP requests for custom reports
type CustomReportHandler struct {
	usecase   *usecases.CustomReportUseCase
	validator *validation.Validator
}

// NewCustomReportHandler creates a new CustomReportHandler
func NewCustomReportHandler(usecase *usecases.CustomReportUseCase) *CustomReportHandler {
	return &CustomReportHandler{
		usecase:   usecase,
		validator: validation.NewValidator(),
	}
}

// getUserID extracts user ID from context
func (h *CustomReportHandler) getUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := userID.(int64)
	if !ok {
		// Try other integer types
		if intID, ok := userID.(int); ok {
			return int64(intID), true
		}
		if uint64ID, ok := userID.(uint64); ok {
			return int64(uint64ID), true
		}
	}
	return id, ok
}

// getIDParam extracts UUID from URL parameter
func (h *CustomReportHandler) getIDParam(c *gin.Context) (uuid.UUID, error) {
	idStr := c.Param("id")
	return uuid.Parse(idStr)
}

// handleError handles errors and sends appropriate responses
func (h *CustomReportHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecases.ErrCustomReportNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Custom report not found"})
	case errors.Is(err, usecases.ErrUnauthorizedAccess):
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
	case errors.Is(err, usecases.ErrInvalidDataSource):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data source"})
	case errors.Is(err, usecases.ErrInvalidFields):
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field is required"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}

// Create creates a new custom report
// @Summary Create custom report
// @Description Create a new custom report template
// @Tags custom-reports
// @Accept json
// @Produce json
// @Param input body dto.CreateCustomReportInput true "Report configuration"
// @Success 201 {object} dto.CustomReportOutput
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/reports/custom [post]
func (h *CustomReportHandler) Create(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input dto.CreateCustomReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validator.Validate(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.usecase.Create(c.Request.Context(), input, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetByID retrieves a custom report by ID
// @Summary Get custom report
// @Description Get a custom report by ID
// @Tags custom-reports
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} dto.CustomReportOutput
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /api/reports/custom/{id} [get]
func (h *CustomReportHandler) GetByID(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id, err := h.getIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// Update updates a custom report
// @Summary Update custom report
// @Description Update an existing custom report
// @Tags custom-reports
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Param input body dto.UpdateCustomReportInput true "Update data"
// @Success 200 {object} dto.CustomReportOutput
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /api/reports/custom/{id} [put]
func (h *CustomReportHandler) Update(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id, err := h.getIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	var input dto.UpdateCustomReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validator.Validate(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.usecase.Update(c.Request.Context(), id, input, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// Delete deletes a custom report
// @Summary Delete custom report
// @Description Delete a custom report
// @Tags custom-reports
// @Param id path string true "Report ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /api/reports/custom/{id} [delete]
func (h *CustomReportHandler) Delete(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id, err := h.getIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), id, userID); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// List lists custom reports
// @Summary List custom reports
// @Description List custom reports with filtering and pagination
// @Tags custom-reports
// @Produce json
// @Param dataSource query string false "Filter by data source"
// @Param isPublic query bool false "Filter by public status"
// @Param search query string false "Search in name and description"
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Success 200 {object} dto.CustomReportListOutput
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /api/reports/custom [get]
func (h *CustomReportHandler) List(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var filter dto.CustomReportFilterInput
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 10
	}

	result, err := h.usecase.List(c.Request.Context(), filter, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// Execute executes a custom report
// @Summary Execute custom report
// @Description Execute a custom report and get data
// @Tags custom-reports
// @Accept json
// @Produce json
// @Param id path string true "Report ID"
// @Param input body dto.ExecuteReportInput true "Execution options"
// @Success 200 {object} dto.ExecuteReportOutput
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /api/reports/custom/{id}/execute [post]
func (h *CustomReportHandler) Execute(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id, err := h.getIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	var input dto.ExecuteReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		// Default values if no body
		input = dto.ExecuteReportInput{
			Page:     1,
			PageSize: 50,
		}
	}

	// Apply query params if provided
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			input.Page = page
		}
	}
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			input.PageSize = pageSize
		}
	}

	result, err := h.usecase.Execute(c.Request.Context(), id, input, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// Export exports a custom report
// @Summary Export custom report
// @Description Export a custom report to PDF, Excel, or CSV
// @Tags custom-reports
// @Accept json
// @Produce application/octet-stream
// @Param id path string true "Report ID"
// @Param input body dto.ExportReportInput true "Export options"
// @Success 200 {file} binary
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /api/reports/custom/{id}/export [post]
func (h *CustomReportHandler) Export(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id, err := h.getIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	var input dto.ExportReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Defaults
	if input.Format == "" {
		input.Format = "xlsx"
	}

	if err := h.validator.Validate(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, filename, err := h.usecase.Export(c.Request.Context(), id, input, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Set content type based on format
	var contentType string
	switch input.Format {
	case "pdf":
		contentType = "application/pdf"
	case "xlsx":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "csv":
		contentType = "text/csv"
	default:
		contentType = "application/octet-stream"
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", strconv.Itoa(len(data)))
	c.Data(http.StatusOK, contentType, data)
}

// GetMyReports gets reports created by the current user
// @Summary Get my reports
// @Description Get all custom reports created by the current user
// @Tags custom-reports
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Success 200 {object} dto.CustomReportListOutput
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /api/reports/custom/my [get]
func (h *CustomReportHandler) GetMyReports(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	result, err := h.usecase.GetMyReports(c.Request.Context(), page, pageSize, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPublicReports gets all public reports
// @Summary Get public reports
// @Description Get all public custom reports
// @Tags custom-reports
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Success 200 {object} dto.CustomReportListOutput
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /api/reports/custom/public [get]
func (h *CustomReportHandler) GetPublicReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	result, err := h.usecase.GetPublicReports(c.Request.Context(), page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetAvailableFields returns available fields for each data source
// @Summary Get available fields
// @Description Get available fields for custom report builder
// @Tags custom-reports
// @Produce json
// @Success 200 {object} map[string][]dto.ReportFieldDTO
// @Security BearerAuth
// @Router /api/reports/custom/fields [get]
func (h *CustomReportHandler) GetAvailableFields(c *gin.Context) {
	// Return the available fields for each data source
	// This matches the frontend AVAILABLE_FIELDS constant
	fields := map[string][]dto.ReportFieldDTO{
		"documents": {
			{ID: "doc_id", Name: "id", Label: "ID", Type: "string", Source: "documents"},
			{ID: "doc_name", Name: "name", Label: "Name", Type: "string", Source: "documents"},
			{ID: "doc_category", Name: "category", Label: "Category", Type: "enum", Source: "documents", EnumValues: []string{"educational", "hr", "administrative", "methodical", "financial", "archive"}},
			{ID: "doc_status", Name: "status", Label: "Status", Type: "enum", Source: "documents", EnumValues: []string{"uploading", "processing", "ready", "error"}},
			{ID: "doc_size", Name: "size", Label: "Size", Type: "number", Source: "documents"},
			{ID: "doc_created", Name: "created_at", Label: "Created At", Type: "date", Source: "documents"},
			{ID: "doc_updated", Name: "updated_at", Label: "Updated At", Type: "date", Source: "documents"},
			{ID: "doc_author", Name: "author_name", Label: "Author", Type: "string", Source: "documents"},
			{ID: "doc_tags", Name: "tags", Label: "Tags", Type: "string", Source: "documents"},
		},
		"users": {
			{ID: "user_id", Name: "id", Label: "ID", Type: "number", Source: "users"},
			{ID: "user_name", Name: "name", Label: "Name", Type: "string", Source: "users"},
			{ID: "user_email", Name: "email", Label: "Email", Type: "string", Source: "users"},
			{ID: "user_role", Name: "role", Label: "Role", Type: "enum", Source: "users", EnumValues: []string{"admin", "methodist", "secretary", "teacher", "student"}},
			{ID: "user_department", Name: "department", Label: "Department", Type: "string", Source: "users"},
			{ID: "user_created", Name: "created_at", Label: "Created At", Type: "date", Source: "users"},
			{ID: "user_active", Name: "is_active", Label: "Is Active", Type: "boolean", Source: "users"},
		},
		"events": {
			{ID: "event_id", Name: "id", Label: "ID", Type: "number", Source: "events"},
			{ID: "event_title", Name: "title", Label: "Title", Type: "string", Source: "events"},
			{ID: "event_type", Name: "type", Label: "Type", Type: "enum", Source: "events", EnumValues: []string{"lecture", "seminar", "exam", "meeting", "other"}},
			{ID: "event_start", Name: "start_time", Label: "Start Time", Type: "date", Source: "events"},
			{ID: "event_end", Name: "end_time", Label: "End Time", Type: "date", Source: "events"},
			{ID: "event_location", Name: "location", Label: "Location", Type: "string", Source: "events"},
			{ID: "event_organizer", Name: "organizer", Label: "Organizer", Type: "string", Source: "events"},
		},
		"tasks": {
			{ID: "task_id", Name: "id", Label: "ID", Type: "number", Source: "tasks"},
			{ID: "task_title", Name: "title", Label: "Title", Type: "string", Source: "tasks"},
			{ID: "task_status", Name: "status", Label: "Status", Type: "enum", Source: "tasks", EnumValues: []string{"pending", "in_progress", "completed", "cancelled"}},
			{ID: "task_priority", Name: "priority", Label: "Priority", Type: "enum", Source: "tasks", EnumValues: []string{"low", "medium", "high", "urgent"}},
			{ID: "task_due", Name: "due_date", Label: "Due Date", Type: "date", Source: "tasks"},
			{ID: "task_assignee", Name: "assignee", Label: "Assignee", Type: "string", Source: "tasks"},
			{ID: "task_created", Name: "created_at", Label: "Created At", Type: "date", Source: "tasks"},
		},
		"students": {
			{ID: "student_id", Name: "id", Label: "ID", Type: "number", Source: "students"},
			{ID: "student_name", Name: "name", Label: "Name", Type: "string", Source: "students"},
			{ID: "student_group", Name: "group", Label: "Group", Type: "string", Source: "students"},
			{ID: "student_course", Name: "course", Label: "Course", Type: "number", Source: "students"},
			{ID: "student_faculty", Name: "faculty", Label: "Faculty", Type: "string", Source: "students"},
			{ID: "student_status", Name: "status", Label: "Status", Type: "enum", Source: "students", EnumValues: []string{"active", "academic_leave", "expelled", "graduated"}},
			{ID: "student_enrolled", Name: "enrolled_at", Label: "Enrolled At", Type: "date", Source: "students"},
		},
	}

	c.JSON(http.StatusOK, fields)
}
