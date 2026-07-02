package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// LessonSlotService is the use-case surface the slot handler depends on (DIP,
// enables a fake in tests).
type LessonSlotService interface {
	List(ctx context.Context) ([]*entities.LessonSlot, error)
	Create(ctx context.Context, number int, timeStart, timeEnd string) (*entities.LessonSlot, error)
	Update(ctx context.Context, id int64, number int, timeStart, timeEnd string) (*entities.LessonSlot, error)
	Delete(ctx context.Context, id int64) error
}

// LessonSlotHandler serves the bell-schedule catalog endpoints.
type LessonSlotHandler struct {
	svc LessonSlotService
}

// NewLessonSlotHandler creates a new LessonSlotHandler.
func NewLessonSlotHandler(svc LessonSlotService) *LessonSlotHandler {
	return &LessonSlotHandler{svc: svc}
}

// requireSlotWrite allows only system_admin and academic_secretary to mutate
// the shared bell-schedule catalog; others get 403.
func (h *LessonSlotHandler) requireSlotWrite(c *gin.Context) bool {
	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	if roleStr == "system_admin" || roleStr == "academic_secretary" {
		return true
	}
	c.JSON(http.StatusForbidden, gin.H{errorKey: "forbidden: insufficient permissions for schedule modification"})
	return false
}

// mapSlotError maps domain errors to HTTP status codes.
func (h *LessonSlotHandler) mapSlotError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, entities.ErrLessonSlotNotFound):
		c.JSON(http.StatusNotFound, gin.H{errorKey: "lesson slot not found"})
	case errors.Is(err, entities.ErrLessonSlotNumberTaken):
		c.JSON(http.StatusConflict, gin.H{errorKey: "lesson slot number already exists"})
	case errors.Is(err, entities.ErrInvalidSlotNumber),
		errors.Is(err, entities.ErrInvalidSlotTimeFormat),
		errors.Is(err, entities.ErrInvalidSlotTimeRange):
		c.JSON(http.StatusBadRequest, gin.H{errorKey: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{errorKey: "internal server error"})
	}
}

// List returns all slots ordered by number. Available to any authenticated user.
func (h *LessonSlotHandler) List(c *gin.Context) {
	slots, err := h.svc.List(c.Request.Context())
	if err != nil {
		h.mapSlotError(c, err)
		return
	}
	output := make([]dto.LessonSlotOutput, 0, len(slots))
	for _, s := range slots {
		output = append(output, dto.ToLessonSlotOutput(s))
	}
	c.JSON(http.StatusOK, gin.H{"lesson_slots": output})
}

// Create adds a new slot.
func (h *LessonSlotHandler) Create(c *gin.Context) {
	if !h.requireSlotWrite(c) {
		return
	}
	var input dto.LessonSlotInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: err.Error()})
		return
	}
	slot, err := h.svc.Create(c.Request.Context(), input.Number, input.TimeStart, input.TimeEnd)
	if err != nil {
		h.mapSlotError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.ToLessonSlotOutput(slot))
}

// Update edits an existing slot.
func (h *LessonSlotHandler) Update(c *gin.Context) {
	if !h.requireSlotWrite(c) {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid id"})
		return
	}
	var input dto.LessonSlotInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: err.Error()})
		return
	}
	slot, err := h.svc.Update(c.Request.Context(), id, input.Number, input.TimeStart, input.TimeEnd)
	if err != nil {
		h.mapSlotError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToLessonSlotOutput(slot))
}

// Delete removes a slot.
func (h *LessonSlotHandler) Delete(c *gin.Context) {
	if !h.requireSlotWrite(c) {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid id"})
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		h.mapSlotError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
