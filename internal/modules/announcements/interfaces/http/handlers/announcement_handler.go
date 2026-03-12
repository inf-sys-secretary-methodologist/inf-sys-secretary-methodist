// Package handlers provides HTTP handlers for the announcements module.
package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain/entities"
)

// AnnouncementHandler handles HTTP requests for announcements.
type AnnouncementHandler struct {
	useCase *usecases.AnnouncementUseCase
}

// NewAnnouncementHandler creates a new AnnouncementHandler.
func NewAnnouncementHandler(useCase *usecases.AnnouncementUseCase) *AnnouncementHandler {
	return &AnnouncementHandler{useCase: useCase}
}

// getUserID extracts user ID from context.
func (h *AnnouncementHandler) getUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return 0, false
	}
	id, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID type"})
		return 0, false
	}
	return id, true
}

// isAdmin checks if the user is an admin.
func (h *AnnouncementHandler) isAdmin(c *gin.Context) bool {
	role, exists := c.Get("user_role")
	if !exists {
		return false
	}
	roleStr, ok := role.(string)
	if !ok {
		return false
	}
	return roleStr == "admin"
}

// getIDParam extracts ID parameter from URL.
func (h *AnnouncementHandler) getIDParam(c *gin.Context, param string) (int64, bool) {
	idStr := c.Param(param)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
		return 0, false
	}
	return id, true
}

// handleError handles use case errors.
func (h *AnnouncementHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecases.ErrAnnouncementNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "announcement not found"})
	case errors.Is(err, usecases.ErrUnauthorized):
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
	case errors.Is(err, usecases.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
	case errors.Is(err, entities.ErrAnnouncementAlreadyPublished):
		c.JSON(http.StatusConflict, gin.H{"error": "announcement is already published"})
	case errors.Is(err, entities.ErrAnnouncementArchived):
		c.JSON(http.StatusConflict, gin.H{"error": "announcement is archived"})
	case errors.Is(err, entities.ErrAnnouncementNotPublished):
		c.JSON(http.StatusConflict, gin.H{"error": "announcement is not published"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

// Create creates a new announcement.
func (h *AnnouncementHandler) Create(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var req dto.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	announcement, err := h.useCase.Create(c.Request.Context(), userID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToResponse(announcement))
}

// GetByID retrieves an announcement by ID.
func (h *AnnouncementHandler) GetByID(c *gin.Context) {
	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	// Increment view count for public views
	incrementView := true
	announcement, err := h.useCase.GetByID(c.Request.Context(), id, incrementView)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToResponse(announcement))
}

// Update updates an announcement.
func (h *AnnouncementHandler) Update(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	announcement, err := h.useCase.Update(c.Request.Context(), userID, id, h.isAdmin(c), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToResponse(announcement))
}

// Delete deletes an announcement.
func (h *AnnouncementHandler) Delete(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.useCase.Delete(c.Request.Context(), userID, id, h.isAdmin(c)); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "announcement deleted"})
}

// List lists announcements with filters.
func (h *AnnouncementHandler) List(c *gin.Context) {
	var req dto.ListAnnouncementsRequest

	// Parse query parameters
	if authorID := c.Query("author_id"); authorID != "" {
		if id, err := strconv.ParseInt(authorID, 10, 64); err == nil {
			req.AuthorID = &id
		}
	}

	if status := c.Query("status"); status != "" {
		s := domain.AnnouncementStatus(status)
		if s.IsValid() {
			req.Status = &s
		}
	}

	if priority := c.Query("priority"); priority != "" {
		p := domain.AnnouncementPriority(priority)
		if p.IsValid() {
			req.Priority = &p
		}
	}

	if audience := c.Query("target_audience"); audience != "" {
		a := domain.TargetAudience(audience)
		if a.IsValid() {
			req.TargetAudience = &a
		}
	}

	if pinned := c.Query("is_pinned"); pinned != "" {
		p := pinned == "true"
		req.IsPinned = &p
	}

	if search := c.Query("search"); search != "" {
		req.Search = &search
	}

	if tags := c.QueryArray("tags"); len(tags) > 0 {
		req.Tags = tags
	}

	req.Limit = 20
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			req.Limit = l
		}
	}

	req.Offset = 0
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			req.Offset = o
		}
	}

	result, err := h.useCase.List(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPublished retrieves published announcements.
func (h *AnnouncementHandler) GetPublished(c *gin.Context) {
	audience := domain.TargetAudienceAll
	if a := c.Query("audience"); a != "" {
		audience = domain.TargetAudience(a)
		if !audience.IsValid() {
			audience = domain.TargetAudienceAll
		}
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	announcements, err := h.useCase.GetPublished(c.Request.Context(), audience, limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"announcements": dto.ToResponseList(announcements),
	})
}

// GetPinned retrieves pinned announcements.
func (h *AnnouncementHandler) GetPinned(c *gin.Context) {
	limit := 5
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	announcements, err := h.useCase.GetPinned(c.Request.Context(), limit)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"announcements": dto.ToResponseList(announcements),
	})
}

// GetRecent retrieves recent announcements.
func (h *AnnouncementHandler) GetRecent(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	announcements, err := h.useCase.GetRecent(c.Request.Context(), limit)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"announcements": dto.ToResponseList(announcements),
	})
}

// Publish publishes an announcement.
func (h *AnnouncementHandler) Publish(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	announcement, err := h.useCase.Publish(c.Request.Context(), userID, id, h.isAdmin(c))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToResponse(announcement))
}

// Unpublish moves an announcement back to draft.
func (h *AnnouncementHandler) Unpublish(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	announcement, err := h.useCase.Unpublish(c.Request.Context(), userID, id, h.isAdmin(c))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToResponse(announcement))
}

// Archive archives an announcement.
func (h *AnnouncementHandler) Archive(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	id, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	announcement, err := h.useCase.Archive(c.Request.Context(), userID, id, h.isAdmin(c))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToResponse(announcement))
}
