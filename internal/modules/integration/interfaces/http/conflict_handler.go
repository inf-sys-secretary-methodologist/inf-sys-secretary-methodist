package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// ConflictHandler handles sync conflict HTTP requests
type ConflictHandler struct {
	conflictUseCase *usecases.ConflictUseCase
}

// NewConflictHandler creates a new conflict handler
func NewConflictHandler(conflictUseCase *usecases.ConflictUseCase) *ConflictHandler {
	return &ConflictHandler{
		conflictUseCase: conflictUseCase,
	}
}

// RegisterRoutes registers conflict routes
func (h *ConflictHandler) RegisterRoutes(router *gin.RouterGroup) {
	conflicts := router.Group("/conflicts")
	{
		conflicts.GET("", h.List)
		conflicts.GET("/pending", h.GetPending)
		conflicts.GET("/stats", h.GetStats)
		conflicts.GET("/:id", h.GetByID)
		conflicts.POST("/:id/resolve", h.Resolve)
		conflicts.POST("/bulk-resolve", h.BulkResolve)
		conflicts.DELETE("/:id", h.Delete)
	}
}

// List lists sync conflicts
func (h *ConflictHandler) List(c *gin.Context) {
	var req dto.ConflictListRequest

	if syncLogID := c.Query("sync_log_id"); syncLogID != "" {
		id, _ := strconv.ParseInt(syncLogID, 10, 64)
		req.SyncLogID = &id
	}
	if entityType := c.Query("entity_type"); entityType != "" {
		et := entities.SyncEntityType(entityType)
		req.EntityType = &et
	}
	if resolution := c.Query("resolution"); resolution != "" {
		r := entities.ConflictResolution(resolution)
		req.Resolution = &r
	}

	req.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))
	req.Offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))

	result, err := h.conflictUseCase.List(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetByID gets a conflict by ID
func (h *ConflictHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conflict ID"})
		return
	}

	result, err := h.conflictUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conflict not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPending gets pending conflicts
func (h *ConflictHandler) GetPending(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	result, err := h.conflictUseCase.GetPending(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetStats gets conflict statistics
func (h *ConflictHandler) GetStats(c *gin.Context) {
	result, err := h.conflictUseCase.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Resolve resolves a conflict
func (h *ConflictHandler) Resolve(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conflict ID"})
		return
	}

	var req dto.ResolveConflictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate resolution
	switch req.Resolution {
	case entities.ConflictResolutionUseLocal, entities.ConflictResolutionUseExternal,
		entities.ConflictResolutionMerge, entities.ConflictResolutionSkip:
		// Valid
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resolution. Supported: use_local, use_external, merge, skip"})
		return
	}

	// Get user ID from context (assuming auth middleware sets it)
	userID, _ := c.Get("user_id")
	var resolvedByUserID int64
	if uid, ok := userID.(int64); ok {
		resolvedByUserID = uid
	}

	if err := h.conflictUseCase.Resolve(c.Request.Context(), id, resolvedByUserID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conflict resolved successfully"})
}

// BulkResolve resolves multiple conflicts
func (h *ConflictHandler) BulkResolve(c *gin.Context) {
	var req dto.BulkResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No conflict IDs provided"})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")
	var resolvedByUserID int64
	if uid, ok := userID.(int64); ok {
		resolvedByUserID = uid
	}

	if err := h.conflictUseCase.BulkResolve(c.Request.Context(), resolvedByUserID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Conflicts resolved successfully",
		"count":   len(req.IDs),
	})
}

// Delete deletes a conflict
func (h *ConflictHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conflict ID"})
		return
	}

	if err := h.conflictUseCase.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conflict deleted successfully"})
}
