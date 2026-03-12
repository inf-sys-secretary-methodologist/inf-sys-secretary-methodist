// Package http contains HTTP handlers for the notifications module.
package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
)

// NotificationHandler handles notification HTTP requests
type NotificationHandler struct {
	notificationUseCase *usecases.NotificationUseCase
	validate            *validator.Validate
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationUseCase *usecases.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{
		notificationUseCase: notificationUseCase,
		validate:            validator.New(),
	}
}

// List godoc
// @Summary List notifications
// @Description Get a list of notifications for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Param type query string false "Filter by notification type"
// @Param priority query string false "Filter by priority"
// @Param is_read query bool false "Filter by read status"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} dto.NotificationListOutput
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid, _ := userID.(int64)
	input := &dto.NotificationListInput{
		UserID: uid,
		Limit:  50,
		Offset: 0,
	}

	// Parse query parameters
	if typeStr := c.Query("type"); typeStr != "" {
		input.Type = entities.NotificationType(typeStr)
	}
	if priorityStr := c.Query("priority"); priorityStr != "" {
		input.Priority = entities.NotificationPriority(priorityStr)
	}
	if isReadStr := c.Query("is_read"); isReadStr != "" {
		isRead := isReadStr == "true"
		input.IsRead = &isRead
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			input.Limit = limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			input.Offset = offset
		}
	}

	result, err := h.notificationUseCase.List(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetByID godoc
// @Summary Get notification by ID
// @Description Get a single notification by its ID
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} dto.NotificationOutput
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications/{id} [get]
func (h *NotificationHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	notification, err := h.notificationUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if notification == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		return
	}

	c.JSON(http.StatusOK, notification)
}

// MarkAsRead godoc
// @Summary Mark notification as read
// @Description Mark a single notification as read
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications/{id}/read [put]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	if err := h.notificationUseCase.MarkAsRead(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}

// MarkAllAsRead godoc
// @Summary Mark all notifications as read
// @Description Mark all notifications as read for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications/read-all [put]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid, _ := userID.(int64)
	if err := h.notificationUseCase.MarkAllAsRead(c.Request.Context(), uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "all notifications marked as read"})
}

// Delete godoc
// @Summary Delete notification
// @Description Delete a notification by its ID
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications/{id} [delete]
func (h *NotificationHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	if err := h.notificationUseCase.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification deleted"})
}

// DeleteAll godoc
// @Summary Delete all notifications
// @Description Delete all notifications for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications [delete]
func (h *NotificationHandler) DeleteAll(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid, _ := userID.(int64)
	if err := h.notificationUseCase.DeleteAll(c.Request.Context(), uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "all notifications deleted"})
}

// GetUnreadCount godoc
// @Summary Get unread count
// @Description Get the count of unread notifications for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} dto.UnreadCountOutput
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications/unread-count [get]
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	result, err := h.notificationUseCase.GetUnreadCount(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetStats godoc
// @Summary Get notification statistics
// @Description Get notification statistics for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} dto.NotificationStatsOutput
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications/stats [get]
func (h *NotificationHandler) GetStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	result, err := h.notificationUseCase.GetStats(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Create godoc
// @Summary Create notification (admin only)
// @Description Create a new notification for a user (admin only)
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification body dto.CreateNotificationInput true "Notification data"
// @Success 201 {object} dto.NotificationOutput
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/notifications [post]
func (h *NotificationHandler) Create(c *gin.Context) {
	var input dto.CreateNotificationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.notificationUseCase.Create(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// CreateBulk godoc
// @Summary Create bulk notifications (admin only)
// @Description Create notifications for multiple users (admin only)
// @Tags notifications
// @Accept json
// @Produce json
// @Param notifications body dto.CreateBulkNotificationInput true "Bulk notification data"
// @Success 201 {array} dto.NotificationOutput
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/notifications/bulk [post]
func (h *NotificationHandler) CreateBulk(c *gin.Context) {
	var input dto.CreateBulkNotificationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.notificationUseCase.CreateBulk(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}
