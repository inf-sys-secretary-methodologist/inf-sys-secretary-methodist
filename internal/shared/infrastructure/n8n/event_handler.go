package n8n

import (
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/ddd"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// WebhookEventHandler forwards domain events to n8n webhooks.
// It maps event types to webhook paths.
type WebhookEventHandler struct {
	client *Client
	logger *logging.Logger
	// pathMap maps event type (e.g. "document.created") to webhook path (e.g. "document-created")
	pathMap map[string]string
}

// NewWebhookEventHandler creates a handler that forwards events to n8n.
func NewWebhookEventHandler(client *Client, logger *logging.Logger) *WebhookEventHandler {
	return &WebhookEventHandler{
		client: client,
		logger: logger,
		pathMap: map[string]string{
			"document.created": "document-created",
			"document.updated": "document-updated",
		},
	}
}

// Handle forwards the domain event to n8n via webhook asynchronously.
func (h *WebhookEventHandler) Handle(event ddd.DomainEvent) error {
	path, ok := h.pathMap[event.GetEventType()]
	if !ok {
		return nil
	}

	payload := map[string]any{
		"event_type":   event.GetEventType(),
		"aggregate_id": event.GetAggregateID(),
		"occurred_at":  event.GetOccurredAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	h.client.TriggerAsync(path, payload)
	return nil
}

// GetHandlerName returns the handler name for EventBus registration.
func (h *WebhookEventHandler) GetHandlerName() string {
	return "n8n-webhook-handler"
}
