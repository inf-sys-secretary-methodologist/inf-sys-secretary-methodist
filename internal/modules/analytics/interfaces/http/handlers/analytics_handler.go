// Package handlers contains HTTP request handlers for the analytics module.
package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/application/dto"
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
	// Scope is wired in Cycle 5 (handler scope assembly); nil = unrestricted.
	result, err := h.usecase.GetStudentRisk(ctx, nil, studentID)
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
	// Scope is wired in Cycle 5 (handler scope assembly); nil = unrestricted.
	result, err := h.usecase.GetGroupSummary(ctx, nil, groupName)
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

// GetStudentRiskHistory returns risk score history for a student.
func (h *AnalyticsHandler) GetStudentRiskHistory(c *gin.Context) {
	studentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("Invalid student ID"))
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "90"))

	// Scope is wired in Cycle 5 (handler scope assembly); nil = unrestricted.
	result, err := h.usecase.GetStudentRiskHistory(c.Request.Context(), nil, studentID, limit)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// GetRiskWeightConfig returns the current risk weight configuration.
func (h *AnalyticsHandler) GetRiskWeightConfig(c *gin.Context) {
	cfg, err := h.usecase.GetRiskWeightConfig(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(cfg))
}

// UpdateRiskWeightConfig updates risk weight configuration (admin only).
func (h *AnalyticsHandler) UpdateRiskWeightConfig(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("authentication required"))
		return
	}

	var req dto.UpdateRiskWeightConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("Invalid request body"))
		return
	}

	if err := h.usecase.UpdateRiskWeightConfig(c.Request.Context(), req, userID.(int64)); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]string{"message": "Risk weight config updated"}))
}

// ExportAtRiskStudents exports at-risk students as CSV or XLSX.
func (h *AnalyticsHandler) ExportAtRiskStudents(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")

	result, err := h.usecase.GetAtRiskStudents(c.Request.Context(), 1, 1000)
	if err != nil {
		h.handleError(c, err)
		return
	}

	headers := []string{"Student ID", "Name", "Group", "Risk Score", "Risk Level", "Attendance Rate", "Grade Average"}

	switch format {
	case "xlsx":
		f := excelize.NewFile()
		sheet := "At-Risk Students"
		f.SetSheetName("Sheet1", sheet)

		for i, hdr := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			_ = f.SetCellValue(sheet, cell, hdr)
		}

		for row, s := range result.Students {
			r := row + 2
			_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", r), s.StudentID)
			_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", r), s.StudentName)
			if s.GroupName != nil {
				_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", r), *s.GroupName)
			}
			_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", r), s.RiskScore)
			_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", r), s.RiskLevel)
			if s.AttendanceRate != nil {
				_ = f.SetCellValue(sheet, fmt.Sprintf("F%d", r), *s.AttendanceRate)
			}
			if s.GradeAverage != nil {
				_ = f.SetCellValue(sheet, fmt.Sprintf("G%d", r), *s.GradeAverage)
			}
		}

		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename=at-risk-students.xlsx")
		_ = f.Write(c.Writer)

	default: // csv
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", "attachment; filename=at-risk-students.csv")

		w := csv.NewWriter(c.Writer)
		_ = w.Write(headers)

		for _, s := range result.Students {
			group := ""
			if s.GroupName != nil {
				group = *s.GroupName
			}
			attendance := ""
			if s.AttendanceRate != nil {
				attendance = fmt.Sprintf("%.1f", *s.AttendanceRate)
			}
			grade := ""
			if s.GradeAverage != nil {
				grade = fmt.Sprintf("%.1f", *s.GradeAverage)
			}
			_ = w.Write([]string{
				strconv.FormatInt(s.StudentID, 10),
				s.StudentName,
				group,
				fmt.Sprintf("%.1f", s.RiskScore),
				string(s.RiskLevel),
				attendance,
				grade,
			})
		}
		w.Flush()
	}
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
