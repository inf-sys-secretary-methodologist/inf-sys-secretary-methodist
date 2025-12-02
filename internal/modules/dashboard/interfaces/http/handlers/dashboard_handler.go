// Package handlers contains HTTP handlers for the dashboard module.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// DashboardHandler handles HTTP requests for dashboard operations
type DashboardHandler struct {
	usecase *usecases.DashboardUseCase
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(usecase *usecases.DashboardUseCase) *DashboardHandler {
	return &DashboardHandler{usecase: usecase}
}

// GetStats returns KPI statistics for the dashboard
// @Summary Get dashboard statistics
// @Description Returns KPI statistics including documents, reports, tasks, events, and students counts
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param period query string false "Period for comparison (week, month, quarter, year)" default(month)
// @Success 200 {object} dto.DashboardStatsOutput
// @Failure 500 {object} map[string]string
// @Router /api/dashboard/stats [get]
func (h *DashboardHandler) GetStats(c *gin.Context) {
	period := c.DefaultQuery("period", "month")

	// Validate period
	validPeriods := map[string]bool{"week": true, "month": true, "quarter": true, "year": true}
	if !validPeriods[period] {
		period = "month"
	}

	stats, err := h.usecase.GetStats(c.Request.Context(), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.InternalError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(stats))
}

// GetTrends returns trend data for charts
// @Summary Get dashboard trends
// @Description Returns trend data for documents, reports, tasks, and events for charting
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param period query string false "Period (week, month, quarter, year)" default(month)
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} dto.DashboardTrendsOutput
// @Failure 500 {object} map[string]string
// @Router /api/dashboard/trends [get]
func (h *DashboardHandler) GetTrends(c *gin.Context) {
	input := &dto.DashboardTrendsInput{
		Period:    c.DefaultQuery("period", "month"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
	}

	trends, err := h.usecase.GetTrends(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.InternalError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(trends))
}

// GetActivity returns recent activity
// @Summary Get recent activity
// @Description Returns recent activity across all modules (documents, reports, tasks, events, announcements)
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param limit query int false "Number of items to return (max 50)" default(10)
// @Success 200 {object} dto.DashboardActivityOutput
// @Failure 500 {object} map[string]string
// @Router /api/dashboard/activity [get]
func (h *DashboardHandler) GetActivity(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	activity, err := h.usecase.GetActivity(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.InternalError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(activity))
}

// Export exports dashboard data (placeholder for future implementation)
// @Summary Export dashboard data
// @Description Exports dashboard statistics and trends to PDF or Excel format
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param input body dto.ExportDashboardInput true "Export parameters"
// @Success 200 {object} dto.ExportDashboardOutput
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/dashboard/export [post]
func (h *DashboardHandler) Export(c *gin.Context) {
	var input dto.ExportDashboardInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body"))
		return
	}

	// Validate format
	if input.Format != "pdf" && input.Format != "xlsx" {
		c.JSON(http.StatusBadRequest, response.BadRequest("format must be 'pdf' or 'xlsx'"))
		return
	}

	// TODO: Implement actual export functionality
	// For now, return a placeholder response
	c.JSON(http.StatusOK, response.Success(dto.ExportDashboardOutput{
		FileURL:   "/api/files/dashboard-export." + input.Format,
		FileName:  "dashboard-export." + input.Format,
		FileSize:  0,
		ExpiresAt: "",
	}))
}
