package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

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

// List returns all slots. STUB — see GREEN commit.
func (h *LessonSlotHandler) List(c *gin.Context) {
	c.Status(http.StatusOK)
}

// Create adds a slot. STUB — see GREEN commit.
func (h *LessonSlotHandler) Create(c *gin.Context) {
	c.Status(http.StatusOK)
}

// Update edits a slot. STUB — see GREEN commit.
func (h *LessonSlotHandler) Update(c *gin.Context) {
	c.Status(http.StatusOK)
}

// Delete removes a slot. STUB — see GREEN commit.
func (h *LessonSlotHandler) Delete(c *gin.Context) {
	c.Status(http.StatusOK)
}
