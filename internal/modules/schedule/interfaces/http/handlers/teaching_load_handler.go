package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
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

// List returns load lines. STUB — see GREEN commit.
func (h *TeachingLoadHandler) List(c *gin.Context) {
	c.Status(http.StatusOK)
}

// Create adds a load line. STUB — see GREEN commit.
func (h *TeachingLoadHandler) Create(c *gin.Context) {
	c.Status(http.StatusOK)
}

// Update edits a load line. STUB — see GREEN commit.
func (h *TeachingLoadHandler) Update(c *gin.Context) {
	c.Status(http.StatusOK)
}

// Delete removes a load line. STUB — see GREEN commit.
func (h *TeachingLoadHandler) Delete(c *gin.Context) {
	c.Status(http.StatusOK)
}
