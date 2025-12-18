package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/usecases"
)

// EmployeeHandler handles external employee HTTP requests
type EmployeeHandler struct {
	employeeUseCase *usecases.EmployeeUseCase
}

// NewEmployeeHandler creates a new employee handler
func NewEmployeeHandler(employeeUseCase *usecases.EmployeeUseCase) *EmployeeHandler {
	return &EmployeeHandler{
		employeeUseCase: employeeUseCase,
	}
}

// RegisterRoutes registers employee routes
func (h *EmployeeHandler) RegisterRoutes(router *gin.RouterGroup) {
	employees := router.Group("/employees")
	{
		employees.GET("", h.List)
		employees.GET("/unlinked", h.GetUnlinked)
		employees.GET("/:id", h.GetByID)
		employees.POST("/:id/link", h.Link)
		employees.DELETE("/:id/link", h.Unlink)
		employees.DELETE("/:id", h.Delete)
	}
}

// List lists external employees
func (h *EmployeeHandler) List(c *gin.Context) {
	var req dto.ExternalEmployeeListRequest

	req.Search = c.Query("search")
	req.Department = c.Query("department")
	req.Position = c.Query("position")

	if isActive := c.Query("is_active"); isActive != "" {
		b, _ := strconv.ParseBool(isActive)
		req.IsActive = &b
	}
	if isLinked := c.Query("is_linked"); isLinked != "" {
		b, _ := strconv.ParseBool(isLinked)
		req.IsLinked = &b
	}

	req.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))
	req.Offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))

	result, err := h.employeeUseCase.List(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetByID gets an external employee by ID
func (h *EmployeeHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	result, err := h.employeeUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetUnlinked gets unlinked external employees
func (h *EmployeeHandler) GetUnlinked(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	result, err := h.employeeUseCase.GetUnlinked(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Link links an external employee to a local user
func (h *EmployeeHandler) Link(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	var req dto.LinkEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.employeeUseCase.LinkToLocalUser(c.Request.Context(), id, req.LocalUserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Employee linked successfully"})
}

// Unlink unlinks an external employee from a local user
func (h *EmployeeHandler) Unlink(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	if err := h.employeeUseCase.Unlink(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Employee unlinked successfully"})
}

// Delete deletes an external employee
func (h *EmployeeHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	if err := h.employeeUseCase.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Employee deleted successfully"})
}
