package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// SyncHandler handles sync-related HTTP requests
type SyncHandler struct {
	syncUseCase *usecases.SyncUseCase
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(syncUseCase *usecases.SyncUseCase) *SyncHandler {
	return &SyncHandler{
		syncUseCase: syncUseCase,
	}
}

// RegisterRoutes registers sync routes
func (h *SyncHandler) RegisterRoutes(router *gin.RouterGroup) {
	sync := router.Group("/sync")
	{
		sync.POST("/start", h.StartSync)
		sync.GET("/logs", h.ListSyncLogs)
		sync.GET("/logs/:id", h.GetSyncLog)
		sync.POST("/logs/:id/cancel", h.CancelSync)
		sync.GET("/stats", h.GetSyncStats)
		sync.GET("/ping", h.Ping)
		sync.GET("/status", h.GetStatus)
	}
}

// StartSync starts a new synchronization
func (h *SyncHandler) StartSync(c *gin.Context) {
	var req dto.StartSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate entity type
	switch req.EntityType {
	case entities.SyncEntityEmployee, entities.SyncEntityStudent:
		// Valid
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity type. Supported: employee, student"})
		return
	}

	// Validate direction
	switch req.Direction {
	case entities.SyncDirectionImport, entities.SyncDirectionExport, entities.SyncDirectionBoth:
		// Valid
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid direction. Supported: import, export, both"})
		return
	}

	result, err := h.syncUseCase.StartSync(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListSyncLogs lists sync logs
func (h *SyncHandler) ListSyncLogs(c *gin.Context) {
	var req dto.SyncListRequest

	if entityType := c.Query("entity_type"); entityType != "" {
		et := entities.SyncEntityType(entityType)
		req.EntityType = &et
	}
	if direction := c.Query("direction"); direction != "" {
		d := entities.SyncDirection(direction)
		req.Direction = &d
	}
	if status := c.Query("status"); status != "" {
		s := entities.SyncStatus(status)
		req.Status = &s
	}

	req.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))
	req.Offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))

	result, err := h.syncUseCase.GetSyncLogs(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetSyncLog gets a single sync log
func (h *SyncHandler) GetSyncLog(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sync log ID"})
		return
	}

	result, err := h.syncUseCase.GetSyncLog(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sync log not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CancelSync cancels a running sync
func (h *SyncHandler) CancelSync(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sync log ID"})
		return
	}

	if err := h.syncUseCase.CancelSync(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sync canceled"})
}

// GetSyncStats gets sync statistics
func (h *SyncHandler) GetSyncStats(c *gin.Context) {
	var entityType *entities.SyncEntityType
	if et := c.Query("entity_type"); et != "" {
		t := entities.SyncEntityType(et)
		entityType = &t
	}

	result, err := h.syncUseCase.GetSyncStats(c.Request.Context(), entityType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Ping checks 1C connection
func (h *SyncHandler) Ping(c *gin.Context) {
	if err := h.syncUseCase.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "1C server is not reachable: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "1C server is reachable"})
}

// GetStatus gets current sync status
func (h *SyncHandler) GetStatus(c *gin.Context) {
	status := gin.H{
		"employee_sync_running": h.syncUseCase.IsSyncRunning(entities.SyncEntityEmployee),
		"student_sync_running":  h.syncUseCase.IsSyncRunning(entities.SyncEntityStudent),
	}

	c.JSON(http.StatusOK, status)
}
