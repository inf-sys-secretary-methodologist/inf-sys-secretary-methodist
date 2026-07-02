package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// TeachingLoadService is the use-case surface the load handler depends on (DIP).
type TeachingLoadService interface {
	List(ctx context.Context, filter usecases.TeachingLoadFilter) ([]*entities.TeachingLoad, error)
	Create(ctx context.Context, p usecases.TeachingLoadParams) (*entities.TeachingLoad, error)
	Update(ctx context.Context, id int64, p usecases.TeachingLoadParams) (*entities.TeachingLoad, error)
	Delete(ctx context.Context, id int64) error
}

// TeachingLoadHandler serves the planned-teaching-load endpoints.
type TeachingLoadHandler struct {
	svc TeachingLoadService
}

// NewTeachingLoadHandler creates a new TeachingLoadHandler.
func NewTeachingLoadHandler(svc TeachingLoadService) *TeachingLoadHandler {
	return &TeachingLoadHandler{svc: svc}
}

// requireLoadWrite allows methodist, academic_secretary and system_admin to
// manage the teaching load (academic planning); others get 403.
func (h *TeachingLoadHandler) requireLoadWrite(c *gin.Context) bool {
	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	switch roleStr {
	case "system_admin", "academic_secretary", "methodist":
		return true
	}
	c.JSON(http.StatusForbidden, gin.H{errorKey: "forbidden: insufficient permissions for teaching load"})
	return false
}

// mapLoadError maps domain errors to HTTP status codes.
func (h *TeachingLoadHandler) mapLoadError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, entities.ErrTeachingLoadNotFound):
		c.JSON(http.StatusNotFound, gin.H{errorKey: "teaching load not found"})
	case errors.Is(err, entities.ErrTeachingLoadDuplicate):
		c.JSON(http.StatusConflict, gin.H{errorKey: "teaching load already exists"})
	case errors.Is(err, entities.ErrInvalidLoadReference),
		errors.Is(err, entities.ErrInvalidLoadPairs),
		errors.Is(err, entities.ErrInvalidLoadWeekType):
		c.JSON(http.StatusBadRequest, gin.H{errorKey: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{errorKey: "internal server error"})
	}
}

// paramsFromInput builds a use-case params struct from a request body.
func paramsFromInput(in dto.TeachingLoadInput) usecases.TeachingLoadParams {
	return usecases.TeachingLoadParams{
		SemesterID:   in.SemesterID,
		GroupID:      in.GroupID,
		DisciplineID: in.DisciplineID,
		TeacherID:    in.TeacherID,
		LessonTypeID: in.LessonTypeID,
		PairsPerWeek: in.PairsPerWeek,
		WeekType:     domain.WeekType(in.WeekType),
	}
}

// List returns load lines filtered by optional semester/group/teacher. Open to any authenticated user.
func (h *TeachingLoadHandler) List(c *gin.Context) {
	var filter usecases.TeachingLoadFilter
	if v := c.Query("semester_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.SemesterID = &id
		}
	}
	if v := c.Query("group_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.GroupID = &id
		}
	}
	if v := c.Query("teacher_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.TeacherID = &id
		}
	}

	loads, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		h.mapLoadError(c, err)
		return
	}
	output := make([]dto.TeachingLoadOutput, 0, len(loads))
	for _, l := range loads {
		output = append(output, dto.ToTeachingLoadOutput(l))
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"teaching_loads": output}))
}

// Create adds a new load line.
func (h *TeachingLoadHandler) Create(c *gin.Context) {
	if !h.requireLoadWrite(c) {
		return
	}
	var input dto.TeachingLoadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: err.Error()})
		return
	}
	load, err := h.svc.Create(c.Request.Context(), paramsFromInput(input))
	if err != nil {
		h.mapLoadError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success(dto.ToTeachingLoadOutput(load)))
}

// Update edits an existing load line.
func (h *TeachingLoadHandler) Update(c *gin.Context) {
	if !h.requireLoadWrite(c) {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid id"})
		return
	}
	var input dto.TeachingLoadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: err.Error()})
		return
	}
	load, err := h.svc.Update(c.Request.Context(), id, paramsFromInput(input))
	if err != nil {
		h.mapLoadError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(dto.ToTeachingLoadOutput(load)))
}

// Delete removes a load line.
func (h *TeachingLoadHandler) Delete(c *gin.Context) {
	if !h.requireLoadWrite(c) {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid id"})
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		h.mapLoadError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
