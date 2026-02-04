// Package handlers contains HTTP request handlers for the analytics module.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// AnalyticsHandler handles HTTP requests for analytics endpoints.
type AnalyticsHandler struct {
	usecase *usecases.AnalyticsUseCase
}

// NewAnalyticsHandler creates a new analytics handler.
func NewAnalyticsHandler(usecase *usecases.AnalyticsUseCase) *AnalyticsHandler {
	return &AnalyticsHandler{
		usecase: usecase,
	}
}

// GetAtRiskStudents returns students at risk based on attendance and grades
// @Summary Get at-risk students
// @Description Returns a paginated list of students who are at risk based on their attendance and grades
// @Tags Analytics
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=dto.AtRiskStudentsResponse}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/analytics/at-risk-students [get]
func (h *AnalyticsHandler) GetAtRiskStudents(c *gin.Context) {
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	ctx := c.Request.Context()
	result, err := h.usecase.GetAtRiskStudents(ctx, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(result)
	c.JSON(http.StatusOK, resp)
}

// GetStudentRisk returns risk assessment for a specific student
// @Summary Get student risk
// @Description Returns detailed risk assessment for a specific student
// @Tags Analytics
// @Accept json
// @Produce json
// @Param id path int true "Student ID"
// @Success 200 {object} response.Response{data=dto.StudentRiskResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/analytics/students/{id}/risk [get]
func (h *AnalyticsHandler) GetStudentRisk(c *gin.Context) {
	studentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Invalid student ID")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.usecase.GetStudentRisk(ctx, studentID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(result)
	c.JSON(http.StatusOK, resp)
}

// GetGroupSummary returns analytics summary for a specific group
// @Summary Get group analytics summary
// @Description Returns analytics summary for a specific student group
// @Tags Analytics
// @Accept json
// @Produce json
// @Param name path string true "Group name"
// @Success 200 {object} response.Response{data=dto.GroupSummaryResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/analytics/groups/{name}/summary [get]
func (h *AnalyticsHandler) GetGroupSummary(c *gin.Context) {
	groupName := c.Param("name")
	if groupName == "" {
		resp := response.BadRequest("Group name is required")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.usecase.GetGroupSummary(ctx, groupName)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(result)
	c.JSON(http.StatusOK, resp)
}

// GetAllGroupsSummary returns analytics summary for all groups
// @Summary Get all groups analytics summary
// @Description Returns analytics summary for all student groups
// @Tags Analytics
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=dto.AllGroupsSummaryResponse}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/analytics/groups/summary [get]
func (h *AnalyticsHandler) GetAllGroupsSummary(c *gin.Context) {
	ctx := c.Request.Context()
	result, err := h.usecase.GetAllGroupsSummary(ctx)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(result)
	c.JSON(http.StatusOK, resp)
}

// GetStudentsByRiskLevel returns students filtered by risk level
// @Summary Get students by risk level
// @Description Returns a paginated list of students filtered by risk level
// @Tags Analytics
// @Accept json
// @Produce json
// @Param level path string true "Risk level (low, medium, high, critical)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=dto.AtRiskStudentsResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/analytics/risk-level/{level} [get]
func (h *AnalyticsHandler) GetStudentsByRiskLevel(c *gin.Context) {
	riskLevel := c.Param("level")
	if riskLevel == "" {
		resp := response.BadRequest("Risk level is required")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Validate risk level
	validLevels := map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
	if !validLevels[riskLevel] {
		resp := response.BadRequest("Invalid risk level. Must be one of: low, medium, high, critical")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	ctx := c.Request.Context()
	result, err := h.usecase.GetStudentsByRiskLevel(ctx, riskLevel, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(result)
	c.JSON(http.StatusOK, resp)
}

// GetAttendanceTrend returns monthly attendance trend data
// @Summary Get attendance trend
// @Description Returns monthly attendance trend data for the specified number of months
// @Tags Analytics
// @Accept json
// @Produce json
// @Param months query int false "Number of months" default(6)
// @Success 200 {object} response.Response{data=dto.AttendanceTrendResponse}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/analytics/attendance-trend [get]
func (h *AnalyticsHandler) GetAttendanceTrend(c *gin.Context) {
	months := 6

	if m := c.Query("months"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed > 0 {
			months = parsed
		}
	}

	ctx := c.Request.Context()
	result, err := h.usecase.GetAttendanceTrend(ctx, months)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := response.Success(result)
	c.JSON(http.StatusOK, resp)
}

// handleError maps errors to HTTP responses
func (h *AnalyticsHandler) handleError(c *gin.Context, err error) {
	if err.Error() == "student not found" || err.Error() == "group not found" {
		resp := response.NotFound(err.Error())
		c.JSON(http.StatusNotFound, resp)
		return
	}

	httpErr := response.MapDomainError(err)
	c.JSON(httpErr.Status, httpErr.Response)
}
