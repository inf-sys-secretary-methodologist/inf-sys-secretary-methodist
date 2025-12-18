package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/usecases"
)

// StudentHandler handles external student HTTP requests
type StudentHandler struct {
	studentUseCase *usecases.StudentUseCase
}

// NewStudentHandler creates a new student handler
func NewStudentHandler(studentUseCase *usecases.StudentUseCase) *StudentHandler {
	return &StudentHandler{
		studentUseCase: studentUseCase,
	}
}

// RegisterRoutes registers student routes
func (h *StudentHandler) RegisterRoutes(router *gin.RouterGroup) {
	students := router.Group("/students")
	{
		students.GET("", h.List)
		students.GET("/unlinked", h.GetUnlinked)
		students.GET("/groups", h.GetGroups)
		students.GET("/faculties", h.GetFaculties)
		students.GET("/:id", h.GetByID)
		students.POST("/:id/link", h.Link)
		students.DELETE("/:id/link", h.Unlink)
		students.DELETE("/:id", h.Delete)
	}
}

// List lists external students
func (h *StudentHandler) List(c *gin.Context) {
	var req dto.ExternalStudentListRequest

	req.Search = c.Query("search")
	req.GroupName = c.Query("group_name")
	req.Faculty = c.Query("faculty")
	req.Status = c.Query("status")

	if course := c.Query("course"); course != "" {
		courseInt, _ := strconv.Atoi(course)
		req.Course = &courseInt
	}
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

	result, err := h.studentUseCase.List(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetByID gets an external student by ID
func (h *StudentHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	result, err := h.studentUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetUnlinked gets unlinked external students
func (h *StudentHandler) GetUnlinked(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	result, err := h.studentUseCase.GetUnlinked(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetGroups gets all student groups
func (h *StudentHandler) GetGroups(c *gin.Context) {
	result, err := h.studentUseCase.GetGroups(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetFaculties gets all faculties
func (h *StudentHandler) GetFaculties(c *gin.Context) {
	result, err := h.studentUseCase.GetFaculties(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Link links an external student to a local user
func (h *StudentHandler) Link(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	var req dto.LinkStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.studentUseCase.LinkToLocalUser(c.Request.Context(), id, req.LocalUserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Student linked successfully"})
}

// Unlink unlinks an external student from a local user
func (h *StudentHandler) Unlink(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	if err := h.studentUseCase.Unlink(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Student unlinked successfully"})
}

// Delete deletes an external student
func (h *StudentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	if err := h.studentUseCase.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Student deleted successfully"})
}
