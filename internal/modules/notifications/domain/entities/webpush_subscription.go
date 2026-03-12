// Package entities contains domain entities for the notifications module.
package entities

import (
	"time"
)

// WebPushSubscription represents a browser push notification subscription
type WebPushSubscription struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	Endpoint   string     `json:"endpoint"`
	P256dhKey  string     `json:"p256dh_key"`
	AuthKey    string     `json:"auth_key"`
	UserAgent  string     `json:"user_agent,omitempty"`
	DeviceName string     `json:"device_name,omitempty"`
	IsActive   bool       `json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// NewWebPushSubscription creates a new web push subscription with default values
func NewWebPushSubscription(userID int64, endpoint, p256dhKey, authKey string) *WebPushSubscription {
	now := time.Now()
	return &WebPushSubscription{
		UserID:    userID,
		Endpoint:  endpoint,
		P256dhKey: p256dhKey,
		AuthKey:   authKey,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Deactivate marks the subscription as inactive
func (s *WebPushSubscription) Deactivate() {
	s.IsActive = false
	s.UpdatedAt = time.Now()
}

// Activate marks the subscription as active
func (s *WebPushSubscription) Activate() {
	s.IsActive = true
	s.UpdatedAt = time.Now()
}

// UpdateLastUsed updates the last used timestamp
func (s *WebPushSubscription) UpdateLastUsed() {
	now := time.Now()
	s.LastUsedAt = &now
	s.UpdatedAt = now
}

// WebPushPayload represents the payload for a web push notification
type WebPushPayload struct {
	Title              string          `json:"title"`
	Body               string          `json:"body,omitempty"`
	Icon               string          `json:"icon,omitempty"`
	Badge              string          `json:"badge,omitempty"`
	Tag                string          `json:"tag,omitempty"`
	URL                string          `json:"url,omitempty"`
	RequireInteraction bool            `json:"requireInteraction,omitempty"`
	Data               map[string]any  `json:"data,omitempty"`
	Actions            []WebPushAction `json:"actions,omitempty"`
}

// WebPushAction represents an action button in a push notification
type WebPushAction struct {
	Action string `json:"action"`
	Title  string `json:"title"`
	Icon   string `json:"icon,omitempty"`
}

// NewWebPushPayload creates a new web push payload with default values
func NewWebPushPayload(title, body string) *WebPushPayload {
	return &WebPushPayload{
		Title: title,
		Body:  body,
		Icon:  "/icons/icon-192x192.png",
		Badge: "/icons/icon-72x72.png",
	}
}

// WithURL sets the URL to open when the notification is clicked
func (p *WebPushPayload) WithURL(url string) *WebPushPayload {
	p.URL = url
	return p
}

// WithTag sets a tag for notification replacement
func (p *WebPushPayload) WithTag(tag string) *WebPushPayload {
	p.Tag = tag
	return p
}

// WithRequireInteraction sets whether user interaction is required
func (p *WebPushPayload) WithRequireInteraction(require bool) *WebPushPayload {
	p.RequireInteraction = require
	return p
}

// WithData adds custom data to the notification
func (p *WebPushPayload) WithData(data map[string]any) *WebPushPayload {
	p.Data = data
	return p
}

// AddAction adds an action button to the notification
func (p *WebPushPayload) AddAction(action, title string) *WebPushPayload {
	p.Actions = append(p.Actions, WebPushAction{
		Action: action,
		Title:  title,
	})
	return p
}

// WebPushPayloadFromNotification creates a WebPushPayload from a Notification entity.
func WebPushPayloadFromNotification(notification *Notification) *WebPushPayload {
	payload := NewWebPushPayload(notification.Title, notification.Message)

	if notification.Link != "" {
		payload.WithURL(notification.Link)
	}

	if notification.ImageURL != "" {
		payload.Icon = notification.ImageURL
	}

	// Set tag based on notification type
	payload.WithTag(string(notification.Type))

	// High priority notifications require interaction
	if notification.Priority == PriorityHigh || notification.Priority == PriorityUrgent {
		payload.WithRequireInteraction(true)
	}

	// Initialize data with notification metadata
	data := make(map[string]any)
	if notification.Metadata != nil {
		for k, v := range notification.Metadata {
			data[k] = v
		}
	}

	// Add notification tracking info
	data["notification_id"] = notification.ID
	data["type"] = string(notification.Type)
	data["priority"] = string(notification.Priority)

	payload.WithData(data)

	return payload
}
