// Package handlers contains HTTP request handlers for the reporting module.
package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/sanitization"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

// ReportHandler handles HTTP requests for report endpoints.
type ReportHandler struct {
	usecase   *usecases.ReportUseCase
	validator *validation.Validator
	sanitizer *sanitization.Sanitizer
}

// NewReportHandler creates a new report handler.
func NewReportHandler(usecase *usecases.ReportUseCase) *ReportHandler {
	return &ReportHandler{
		usecase:   usecase,
		validator: validation.NewValidator(),
		sanitizer: sanitization.NewSanitizer(),
	}
}

// getUserID extracts user ID from context
func (h *ReportHandler) getUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(int64), true
}

// getIDParam extracts ID parameter from URL
func (h *ReportHandler) getIDParam(c *gin.Context, name string) (int64, error) {
	idStr := c.Param(name)
	if idStr == "" {
		return 0, errors.New("missing id parameter")
	}
	return strconv.ParseInt(idStr, 10, 64)
}

// Create handles report creation
func (h *ReportHandler) Create(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	var input dto.CreateReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Invalid request format")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize inputs
	input.Title = h.sanitizer.SanitizeString(input.Title)
	if input.Description != nil {
		sanitized := h.sanitizer.SanitizeString(*input.Description)
		input.Description = &sanitized
	}

	// Validate
	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	report, err := h.usecase.Create(ctx, userID, &input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(report)
	c.JSON(http.StatusCreated, resp)
}

// GetByID handles getting a report by ID
func (h *ReportHandler) GetByID(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := h.getIDParam(c, "id")
	if err != nil {
		resp := response.BadRequest("Invalid report ID")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	report, err := h.usecase.GetByID(ctx, id, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(report)
	c.JSON(http.StatusOK, resp)
}

// Update handles report update
func (h *ReportHandler) Update(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := h.getIDParam(c, "id")
	if err != nil {
		resp := response.BadRequest("Invalid report ID")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.UpdateReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Invalid request format")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize inputs
	if input.Title != nil {
		sanitized := h.sanitizer.SanitizeString(*input.Title)
		input.Title = &sanitized
	}
	if input.Description != nil {
		sanitized := h.sanitizer.SanitizeString(*input.Description)
		input.Description = &sanitized
	}

	ctx := c.Request.Context()
	report, err := h.usecase.Update(ctx, id, userID, &input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(report)
	c.JSON(http.StatusOK, resp)
}

// Delete handles report deletion
func (h *ReportHandler) Delete(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := h.getIDParam(c, "id")
	if err != nil {
		resp := response.BadRequest("Invalid report ID")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.Delete(ctx, id, userID); err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(nil)
	c.JSON(http.StatusOK, resp)
}

// List handles listing reports with filters
func (h *ReportHandler) List(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	var input dto.ReportFilterInput
	if err := c.ShouldBindQuery(&input); err != nil {
		resp := response.BadRequest("Invalid query parameters")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	reports, err := h.usecase.List(ctx, userID, &input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(reports)
	c.JSON(http.StatusOK, resp)
}

// Generate handles report generation
func (h *ReportHandler) Generate(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := h.getIDParam(c, "id")
	if err != nil {
		resp := response.BadRequest("Invalid report ID")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.GenerateReportInput
	// Input is optional
	_ = c.ShouldBindJSON(&input)

	ctx := c.Request.Context()
	report, err := h.usecase.Generate(ctx, id, userID, &input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(report)
	c.JSON(http.StatusAccepted, resp)
}

// SubmitForReview handles submitting a report for review
func (h *ReportHandler) SubmitForReview(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := h.getIDParam(c, "id")
	if err != nil {
		resp := response.BadRequest("Invalid report ID")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	report, err := h.usecase.SubmitForReview(ctx, id, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(report)
	c.JSON(http.StatusOK, resp)
}

// Review handles report review (approve/reject)
func (h *ReportHandler) Review(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := h.getIDParam(c, "id")
	if err != nil {
		resp := response.BadRequest("Invalid report ID")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.ReviewReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Invalid request format")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize comment
	input.Comment = h.sanitizer.SanitizeString(input.Comment)

	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	report, err := h.usecase.Review(ctx, id, userID, &input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(report)
	c.JSON(http.StatusOK, resp)
}

// Publish handles report publication
func (h *ReportHandler) Publish(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := h.getIDParam(c, "id")
	if err != nil {
		resp := response.BadRequest("Invalid report ID")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.PublishReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Invalid request format")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	report, err := h.usecase.Publish(ctx, id, userID, &input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(report)
	c.JSON(http.StatusOK, resp)
}

// AddAccess handles adding access to a report
func (h *ReportHandler) AddAccess(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := h.getIDParam(c, "id")
	if err != nil {
		resp := response.BadRequest("Invalid report ID")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.AddAccessInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Invalid request format")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	access, err := h.usecase.AddAccess(ctx, id, userID, &input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(access)
	c.JSON(http.StatusCreated, resp)
}

// RemoveAccess handles removing access from a report
func (h *ReportHandler) RemoveAccess(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	reportID, err := h.getIDParam(c, "id")
	if err != nil {
		resp := response.BadRequest("Invalid report ID")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	accessID, err := h.getIDParam(c, "access_id")
	if err != nil {
		resp := response.BadRequest("Invalid access ID")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.RemoveAccess(ctx, reportID, accessID, userID); err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(nil)
	c.JSON(http.StatusOK, resp)
}

// GetAccess handles getting access permissions for a report
func (h *ReportHandler) GetAccess(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := h.getIDParam(c, "id")
	if err != nil {
		resp := response.BadRequest("Invalid report ID")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	accesses, err := h.usecase.GetAccess(ctx, id, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(accesses)
	c.JSON(http.StatusOK, resp)
}

// AddComment handles adding a comment to a report
func (h *ReportHandler) AddComment(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := h.getIDParam(c, "id")
	if err != nil {
		resp := response.BadRequest("Invalid report ID")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.AddCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Invalid request format")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize
	input.Content = h.sanitizer.SanitizeString(input.Content)

	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	comment, err := h.usecase.AddComment(ctx, id, userID, &input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(comment)
	c.JSON(http.StatusCreated, resp)
}

// GetComments handles getting comments for a report
func (h *ReportHandler) GetComments(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := h.getIDParam(c, "id")
	if err != nil {
		resp := response.BadRequest("Invalid report ID")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	comments, err := h.usecase.GetComments(ctx, id, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(comments)
	c.JSON(http.StatusOK, resp)
}

// GetHistory handles getting history for a report
func (h *ReportHandler) GetHistory(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	id, err := h.getIDParam(c, "id")
	if err != nil {
		resp := response.BadRequest("Invalid report ID")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	ctx := c.Request.Context()
	history, err := h.usecase.GetHistory(ctx, id, userID, limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(history)
	c.JSON(http.StatusOK, resp)
}

// GetReportTypes handles getting all report types
func (h *ReportHandler) GetReportTypes(c *gin.Context) {
	var input dto.ReportTypeFilterInput
	_ = c.ShouldBindQuery(&input)

	ctx := c.Request.Context()
	reportTypes, err := h.usecase.GetReportTypes(ctx, &input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(reportTypes)
	c.JSON(http.StatusOK, resp)
}

// GetReportTypeByID handles getting a specific report type
func (h *ReportHandler) GetReportTypeByID(c *gin.Context) {
	id, err := h.getIDParam(c, "id")
	if err != nil {
		resp := response.BadRequest("Invalid report type ID")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	reportType, err := h.usecase.GetReportTypeByID(ctx, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(reportType)
	c.JSON(http.StatusOK, resp)
}

// handleError maps domain errors to HTTP responses
func (h *ReportHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecases.ErrReportNotFound):
		resp := response.NotFound("Report not found")
		c.JSON(http.StatusNotFound, resp)
	case errors.Is(err, usecases.ErrReportTypeNotFound):
		resp := response.NotFound("Report type not found")
		c.JSON(http.StatusNotFound, resp)
	case errors.Is(err, usecases.ErrUnauthorized):
		resp := response.Forbidden("Access denied")
		c.JSON(http.StatusForbidden, resp)
	case errors.Is(err, usecases.ErrCannotModifyReport):
		resp := response.BadRequest("Cannot modify report in current status")
		c.JSON(http.StatusBadRequest, resp)
	case errors.Is(err, usecases.ErrInvalidInput):
		resp := response.BadRequest("Invalid input")
		c.JSON(http.StatusBadRequest, resp)
	default:
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
	}
}
