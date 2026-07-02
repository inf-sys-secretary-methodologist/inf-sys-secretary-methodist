package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
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

// Preview computes a draft schedule without persisting it.
func (h *GenerateScheduleHandler) Preview(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{errorKey: "not implemented"})
}

// Apply generates and persists the schedule for a semester.
func (h *GenerateScheduleHandler) Apply(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{errorKey: "not implemented"})
}
