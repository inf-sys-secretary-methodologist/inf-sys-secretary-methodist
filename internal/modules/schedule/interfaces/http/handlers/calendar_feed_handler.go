package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// CalendarFeedService is the use-case surface the handler depends on.
type CalendarFeedService interface {
	EnsureToken(ctx context.Context, userID int64) (*entities.CalendarFeedToken, error)
	RotateToken(ctx context.Context, userID int64) (*entities.CalendarFeedToken, error)
	GetToken(ctx context.Context, userID int64) (*entities.CalendarFeedToken, error)
	DeleteToken(ctx context.Context, userID int64) error
	RenderFeed(ctx context.Context, token string) (string, error)
}

// CalendarFeedHandler serves the public iCal feed and the authenticated
// subscription-management endpoints.
type CalendarFeedHandler struct {
	svc     CalendarFeedService
	baseURL string
}

// NewCalendarFeedHandler creates a CalendarFeedHandler. baseURL is the public
// origin used to build the subscription URL (e.g. https://host).
func NewCalendarFeedHandler(svc CalendarFeedService, baseURL string) *CalendarFeedHandler {
	return &CalendarFeedHandler{svc: svc, baseURL: baseURL}
}

// feedURL builds the secret subscription address for a token.
func (h *CalendarFeedHandler) feedURL(token string) string {
	return strings.TrimRight(h.baseURL, "/") + "/api/public/calendar/" + token + "/feed.ics"
}

func (h *CalendarFeedHandler) currentUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("user not authenticated"))
		return 0, false
	}
	return userID.(int64), true
}

// GetSubscription returns the user's current subscription (or subscribed:false).
func (h *CalendarFeedHandler) GetSubscription(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}

	tok, err := h.svc.GetToken(c.Request.Context(), userID)
	if errors.Is(err, entities.ErrCalendarFeedTokenNotFound) {
		c.JSON(http.StatusOK, response.Success(dto.CalendarSubscriptionOutput{Subscribed: false}))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.InternalError("failed to load subscription"))
		return
	}
	c.JSON(http.StatusOK, response.Success(dto.CalendarSubscriptionOutput{Subscribed: true, URL: h.feedURL(tok.Token)}))
}

// CreateSubscription returns the user's subscription, creating one if needed.
func (h *CalendarFeedHandler) CreateSubscription(c *gin.Context) {
	h.issue(c, h.svc.EnsureToken)
}

// RotateSubscription issues a fresh token, invalidating the previous URL.
func (h *CalendarFeedHandler) RotateSubscription(c *gin.Context) {
	h.issue(c, h.svc.RotateToken)
}

// issue runs a token-producing action and returns the resulting subscription.
func (h *CalendarFeedHandler) issue(c *gin.Context, action func(context.Context, int64) (*entities.CalendarFeedToken, error)) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	tok, err := action(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.InternalError("failed to update subscription"))
		return
	}
	c.JSON(http.StatusOK, response.Success(dto.CalendarSubscriptionOutput{Subscribed: true, URL: h.feedURL(tok.Token)}))
}

// DeleteSubscription disables the user's feed.
func (h *CalendarFeedHandler) DeleteSubscription(c *gin.Context) {
	userID, ok := h.currentUserID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteToken(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, response.InternalError("failed to delete subscription"))
		return
	}
	c.Status(http.StatusNoContent)
}

// ServeFeed renders the iCalendar document for a secret token. It is a public,
// unauthenticated endpoint (external calendar clients cannot send a JWT).
func (h *CalendarFeedHandler) ServeFeed(c *gin.Context) {
	token := c.Param("token")

	ics, err := h.svc.RenderFeed(c.Request.Context(), token)
	if errors.Is(err, entities.ErrCalendarFeedTokenNotFound) {
		c.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", "text/calendar; charset=utf-8")
	c.Header("Content-Disposition", `inline; filename="schedule.ics"`)
	c.String(http.StatusOK, ics)
}
