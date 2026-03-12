package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/usecases"
)

// ProjectHandler handles HTTP requests for projects.
type ProjectHandler struct {
	projectUseCase *usecases.ProjectUseCase
}

// NewProjectHandler creates a new ProjectHandler.
func NewProjectHandler(projectUseCase *usecases.ProjectUseCase) *ProjectHandler {
	return &ProjectHandler{projectUseCase: projectUseCase}
}

// getUserID extracts user ID from context.
func (h *ProjectHandler) getUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return 0, false
	}
	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID type"})
		return 0, false
	}
	return uid, true
}

// getIDParam extracts ID parameter from URL.
func (h *ProjectHandler) getIDParam(c *gin.Context, param string) (int64, bool) {
	idStr := c.Param(param)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
		return 0, false
	}
	return id, true
}

// handleError handles use case errors.
func (h *ProjectHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecases.ErrProjectNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
	case errors.Is(err, usecases.ErrUnauthorized):
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
	case errors.Is(err, usecases.ErrCannotModifyProject):
		c.JSON(http.StatusConflict, gin.H{"error": "cannot modify project"})
	case errors.Is(err, usecases.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

// Create creates a new project.
func (h *ProjectHandler) Create(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var input dto.CreateProjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectUseCase.Create(c.Request.Context(), userID, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToProjectOutput(project))
}

// GetByID retrieves a project by ID.
func (h *ProjectHandler) GetByID(c *gin.Context) {
	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	project, err := h.projectUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToProjectOutput(project))
}

// Update updates a project.
func (h *ProjectHandler) Update(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	var input dto.UpdateProjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectUseCase.Update(c.Request.Context(), userID, id, input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToProjectOutput(project))
}

// Delete deletes a project.
func (h *ProjectHandler) Delete(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.projectUseCase.Delete(c.Request.Context(), userID, id); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// List lists projects with filters.
func (h *ProjectHandler) List(c *gin.Context) {
	var input dto.ProjectFilterInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Limit == 0 {
		input.Limit = 20
	}

	output, err := h.projectUseCase.List(c.Request.Context(), input)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, output)
}

// Activate activates a project.
func (h *ProjectHandler) Activate(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	project, err := h.projectUseCase.Activate(c.Request.Context(), userID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToProjectOutput(project))
}

// PutOnHold puts a project on hold.
func (h *ProjectHandler) PutOnHold(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	project, err := h.projectUseCase.PutOnHold(c.Request.Context(), userID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToProjectOutput(project))
}

// Complete completes a project.
func (h *ProjectHandler) Complete(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	project, err := h.projectUseCase.Complete(c.Request.Context(), userID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToProjectOutput(project))
}

// Cancel cancels a project.
func (h *ProjectHandler) Cancel(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	project, err := h.projectUseCase.Cancel(c.Request.Context(), userID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToProjectOutput(project))
}
