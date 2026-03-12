// Package handlers contains HTTP request handlers for the analytics module.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

// AttendanceHandler handles HTTP requests for attendance endpoints.
type AttendanceHandler struct {
	usecase   *usecases.AnalyticsUseCase
	validator *validation.Validator
}

// NewAttendanceHandler creates a new attendance handler.
func NewAttendanceHandler(usecase *usecases.AnalyticsUseCase) *AttendanceHandler {
	return &AttendanceHandler{
		usecase:   usecase,
		validator: validation.NewValidator(),
	}
}

// getUserID extracts user ID from context
func (h *AttendanceHandler) getUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	uid, ok := userID.(int64)
	if !ok {
		return 0, false
	}
	return uid, true
}

// MarkAttendance marks attendance for a single student
// @Summary Mark attendance
// @Description Marks attendance for a single student in a lesson
// @Tags Attendance
// @Accept json
// @Produce json
// @Param request body dto.MarkAttendanceRequest true "Attendance data"
// @Success 201 {object} response.Response{data=dto.AttendanceRecordResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/attendance/mark [post]
func (h *AttendanceHandler) MarkAttendance(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	var req dto.MarkAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp := response.BadRequest("Invalid request format")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.usecase.MarkAttendance(ctx, &req, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(result)
	c.JSON(http.StatusCreated, resp)
}

// BulkMarkAttendance marks attendance for multiple students
// @Summary Bulk mark attendance
// @Description Marks attendance for multiple students in a lesson
// @Tags Attendance
// @Accept json
// @Produce json
// @Param request body dto.BulkMarkAttendanceRequest true "Bulk attendance data"
// @Success 201 {object} response.Response{data=[]dto.AttendanceRecordResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/attendance/bulk [post]
func (h *AttendanceHandler) BulkMarkAttendance(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	var req dto.BulkMarkAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp := response.BadRequest("Invalid request format")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.usecase.BulkMarkAttendance(ctx, &req, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(result)
	c.JSON(http.StatusCreated, resp)
}

// GetLessonAttendance returns attendance records for a specific lesson on a date
// @Summary Get lesson attendance
// @Description Returns attendance records for a specific lesson on a date
// @Tags Attendance
// @Accept json
// @Produce json
// @Param id path int true "Lesson ID"
// @Param date path string true "Date (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=dto.LessonAttendanceResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/attendance/lesson/{id}/date/{date} [get]
func (h *AttendanceHandler) GetLessonAttendance(c *gin.Context) {
	lessonID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Invalid lesson ID")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	date := c.Param("date")
	if date == "" {
		resp := response.BadRequest("Date is required")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.usecase.GetLessonAttendance(ctx, lessonID, date)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(result)
	c.JSON(http.StatusOK, resp)
}

// CreateLesson creates a new lesson
// @Summary Create lesson
// @Description Creates a new lesson for attendance tracking
// @Tags Attendance
// @Accept json
// @Produce json
// @Param request body dto.CreateLessonRequest true "Lesson data"
// @Success 201 {object} response.Response{data=entities.Lesson}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/attendance/lessons [post]
func (h *AttendanceHandler) CreateLesson(c *gin.Context) {
	_, ok := h.getUserID(c)
	if !ok {
		resp := response.Unauthorized("Authorization required")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	var req dto.CreateLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp := response.BadRequest("Invalid request format")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.validator.Validate(req); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.usecase.CreateLesson(ctx, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(result)
	c.JSON(http.StatusCreated, resp)
}

// handleError maps errors to HTTP responses
func (h *AttendanceHandler) handleError(c *gin.Context, err error) {
	errMsg := err.Error()
	if errMsg == "lesson not found" || errMsg == "student not found" {
		resp := response.NotFound(errMsg)
		c.JSON(http.StatusNotFound, resp)
		return
	}

	if errMsg == "invalid lesson date format" {
		resp := response.BadRequest(errMsg)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	httpErr := response.MapDomainError(err)
	c.JSON(httpErr.Status, httpErr.Response)
}
