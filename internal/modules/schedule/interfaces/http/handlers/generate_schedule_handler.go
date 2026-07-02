package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// GenerateScheduleService is the use-case surface the handler depends on (DIP).
type GenerateScheduleService interface {
	Preview(ctx context.Context, params usecases.GenerateParams) (*usecases.SchedulePreview, error)
	Apply(ctx context.Context, params usecases.GenerateParams) (*usecases.ApplyResult, error)
}

// GenerateScheduleHandler serves the automatic schedule-generation endpoints.
type GenerateScheduleHandler struct {
	svc GenerateScheduleService
}

// NewGenerateScheduleHandler creates a new GenerateScheduleHandler.
func NewGenerateScheduleHandler(svc GenerateScheduleService) *GenerateScheduleHandler {
	return &GenerateScheduleHandler{svc: svc}
}

// requireGenerateWrite allows methodist, academic_secretary and system_admin to
// generate schedules (academic planning); others get 403.
func (h *GenerateScheduleHandler) requireGenerateWrite(c *gin.Context) bool {
	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	switch roleStr {
	case "system_admin", "academic_secretary", "methodist":
		return true
	}
	c.JSON(http.StatusForbidden, gin.H{errorKey: "forbidden: insufficient permissions for schedule generation"})
	return false
}

// mapGenerateError maps generation errors to HTTP status codes.
func (h *GenerateScheduleHandler) mapGenerateError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecases.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid input"})
	case errors.Is(err, usecases.ErrScheduleAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{errorKey: "schedule already exists for this semester"})
	case errors.Is(err, usecases.ErrSemesterNotFound):
		c.JSON(http.StatusNotFound, gin.H{errorKey: "semester not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{errorKey: "internal server error"})
	}
}

// paramsFromRequest converts a request body into use-case params.
func paramsFromRequest(req dto.GenerateRequest) usecases.GenerateParams {
	params := usecases.GenerateParams{SemesterID: req.SemesterID}
	for _, d := range req.Days {
		params.Days = append(params.Days, domain.DayOfWeek(d))
	}
	return params
}

// Preview computes a draft schedule without persisting it.
func (h *GenerateScheduleHandler) Preview(c *gin.Context) {
	if !h.requireGenerateWrite(c) {
		return
	}
	var req dto.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: err.Error()})
		return
	}

	preview, err := h.svc.Preview(c.Request.Context(), paramsFromRequest(req))
	if err != nil {
		h.mapGenerateError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(toPreviewOutput(preview)))
}

// Apply generates and persists the schedule for a semester.
func (h *GenerateScheduleHandler) Apply(c *gin.Context) {
	if !h.requireGenerateWrite(c) {
		return
	}
	var req dto.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: err.Error()})
		return
	}

	result, err := h.svc.Apply(c.Request.Context(), paramsFromRequest(req))
	if err != nil {
		h.mapGenerateError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(dto.ApplyResultOutput{
		Created:  result.Created,
		Unplaced: result.Unplaced,
	}))
}

// toPreviewOutput maps the use-case preview read model to its DTO.
func toPreviewOutput(p *usecases.SchedulePreview) dto.SchedulePreviewOutput {
	out := dto.SchedulePreviewOutput{
		Lessons:        make([]dto.GeneratedLessonOutput, 0, len(p.Lessons)),
		Unplaced:       make([]dto.UnplacedLessonOutput, 0, len(p.Unplaced)),
		TotalRequested: p.TotalRequested,
		PlacedCount:    p.PlacedCount,
		UnplacedCount:  p.UnplacedCount,
	}
	for _, l := range p.Lessons {
		out.Lessons = append(out.Lessons, dto.GeneratedLessonOutput{
			LoadID:         l.LoadID,
			GroupID:        l.GroupID,
			GroupName:      l.GroupName,
			TeacherID:      l.TeacherID,
			TeacherName:    l.TeacherName,
			DisciplineID:   l.DisciplineID,
			DisciplineName: l.DisciplineName,
			LessonTypeID:   l.LessonTypeID,
			LessonTypeName: l.LessonTypeName,
			WeekType:       l.WeekType,
			DayOfWeek:      l.DayOfWeek,
			SlotNumber:     l.SlotNumber,
			TimeStart:      l.TimeStart,
			TimeEnd:        l.TimeEnd,
			ClassroomID:    l.ClassroomID,
			ClassroomName:  l.ClassroomName,
		})
	}
	for _, u := range p.Unplaced {
		out.Unplaced = append(out.Unplaced, dto.UnplacedLessonOutput{
			LoadID:         u.LoadID,
			GroupName:      u.GroupName,
			DisciplineName: u.DisciplineName,
			LessonTypeName: u.LessonTypeName,
			WeekType:       u.WeekType,
		})
	}
	return out
}
