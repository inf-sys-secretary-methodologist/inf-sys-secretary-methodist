package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
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

// GetSubscription returns the user's current subscription.
func (h *CalendarFeedHandler) GetSubscription(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}

// CreateSubscription returns the user's subscription, creating one if needed.
func (h *CalendarFeedHandler) CreateSubscription(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}

// RotateSubscription issues a fresh token, invalidating the previous URL.
func (h *CalendarFeedHandler) RotateSubscription(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}

// DeleteSubscription disables the user's feed.
func (h *CalendarFeedHandler) DeleteSubscription(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}

// ServeFeed renders the iCalendar document for a secret token.
func (h *CalendarFeedHandler) ServeFeed(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}
