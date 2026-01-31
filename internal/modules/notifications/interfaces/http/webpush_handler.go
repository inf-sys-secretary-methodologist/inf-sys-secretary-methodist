// Package http contains HTTP handlers for the notifications module.
package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
)

// WebPushHandler handles Web Push notification HTTP requests
type WebPushHandler struct {
	webpushRepo    repositories.WebPushRepository
	webpushService services.WebPushService
}

// NewWebPushHandler creates a new Web Push handler
func NewWebPushHandler(
	webpushRepo repositories.WebPushRepository,
	webpushService services.WebPushService,
) *WebPushHandler {
	return &WebPushHandler{
		webpushRepo:    webpushRepo,
		webpushService: webpushService,
	}
}

// GetVAPIDKey godoc
// @Summary Get VAPID public key
// @Description Returns the VAPID public key for Web Push subscription
// @Tags push
// @Accept json
// @Produce json
// @Success 200 {object} dto.WebPushVAPIDKeyOutput
// @Failure 503 {object} map[string]string
// @Router /api/notifications/push/vapid-key [get]
func (h *WebPushHandler) GetVAPIDKey(c *gin.Context) {
	if !h.webpushService.IsConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "web push is not configured"})
		return
	}

	c.JSON(http.StatusOK, dto.WebPushVAPIDKeyOutput{
		PublicKey: h.webpushService.GetVAPIDPublicKey(),
	})
}

// Subscribe godoc
// @Summary Subscribe to push notifications
// @Description Subscribes the user's browser to Web Push notifications
// @Tags push
// @Accept json
// @Produce json
// @Param input body dto.WebPushSubscribeInput true "Subscription details"
// @Success 200 {object} dto.WebPushSubscriptionOutput
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /api/notifications/push/subscribe [post]
func (h *WebPushHandler) Subscribe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if !h.webpushService.IsConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "web push is not configured"})
		return
	}

	var input dto.WebPushSubscribeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate required fields
	if input.Endpoint == "" || input.P256dhKey == "" || input.AuthKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endpoint, p256dh, and auth are required"})
		return
	}

	// Convert to entity
	sub := input.ToEntity(userID.(int64))

	// Create or update subscription
	if err := h.webpushRepo.Create(c.Request.Context(), sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save subscription"})
		return
	}

	c.JSON(http.StatusOK, dto.ToSubscriptionOutput(sub))
}

// Unsubscribe godoc
// @Summary Unsubscribe from push notifications
// @Description Removes the Web Push subscription for the given endpoint
// @Tags push
// @Accept json
// @Produce json
// @Param input body dto.WebPushUnsubscribeInput true "Unsubscribe details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/notifications/push/unsubscribe [post]
func (h *WebPushHandler) Unsubscribe(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input dto.WebPushUnsubscribeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if input.Endpoint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endpoint is required"})
		return
	}

	// Delete subscription by endpoint
	if err := h.webpushRepo.DeleteByEndpoint(c.Request.Context(), input.Endpoint); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "unsubscribed successfully"})
}

// GetStatus godoc
// @Summary Get push notification status
// @Description Returns the user's push notification subscription status and devices
// @Tags push
// @Accept json
// @Produce json
// @Success 200 {object} dto.WebPushStatusOutput
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications/push/status [get]
func (h *WebPushHandler) GetStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get all subscriptions for user
	subscriptions, err := h.webpushRepo.GetByUserID(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscriptions"})
		return
	}

	// Count active subscriptions
	activeCount := 0
	for _, sub := range subscriptions {
		if sub.IsActive {
			activeCount++
		}
	}

	c.JSON(http.StatusOK, dto.WebPushStatusOutput{
		IsEnabled:     activeCount > 0,
		Subscriptions: dto.ToSubscriptionOutputList(subscriptions),
		TotalDevices:  len(subscriptions),
	})
}

// DeleteSubscription godoc
// @Summary Delete a push subscription
// @Description Deletes a specific push subscription by ID
// @Tags push
// @Accept json
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/notifications/push/subscriptions/{id} [delete]
func (h *WebPushHandler) DeleteSubscription(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	subscriptionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	// Get subscription to verify ownership
	sub, err := h.webpushRepo.GetByID(c.Request.Context(), subscriptionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription"})
		return
	}

	if sub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	// Verify ownership
	if sub.UserID != userID.(int64) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Delete subscription
	if err := h.webpushRepo.Delete(c.Request.Context(), subscriptionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription deleted successfully"})
}

// TestPush godoc
// @Summary Test push notification
// @Description Sends a test push notification to the user's devices
// @Tags push
// @Accept json
// @Produce json
// @Param input body dto.WebPushTestInput true "Test notification content"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /api/notifications/push/test [post]
func (h *WebPushHandler) TestPush(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if !h.webpushService.IsConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "web push is not configured"})
		return
	}

	var input dto.WebPushTestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if input.Title == "" {
		input.Title = "Test Notification"
	}
	if input.Message == "" {
		input.Message = "This is a test push notification."
	}

	// Create payload
	payload := entities.NewWebPushPayload(input.Title, input.Message)
	payload.WithTag("test")

	// Send to user
	if err := h.webpushService.SendToUser(c.Request.Context(), userID.(int64), payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send test notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "test notification sent"})
}
