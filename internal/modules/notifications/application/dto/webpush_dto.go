// Package dto contains data transfer objects for the notifications module.
package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
)

// WebPushSubscribeInput represents input for subscribing to web push notifications
type WebPushSubscribeInput struct {
	Endpoint   string `json:"endpoint" validate:"required,url"`
	P256dhKey  string `json:"p256dh" validate:"required"`
	AuthKey    string `json:"auth" validate:"required"`
	UserAgent  string `json:"user_agent,omitempty"`
	DeviceName string `json:"device_name,omitempty" validate:"max=255"`
}

// ToEntity converts the input to a WebPushSubscription entity
func (i *WebPushSubscribeInput) ToEntity(userID int64) *entities.WebPushSubscription {
	sub := entities.NewWebPushSubscription(userID, i.Endpoint, i.P256dhKey, i.AuthKey)
	sub.UserAgent = i.UserAgent
	sub.DeviceName = i.DeviceName
	return sub
}

// WebPushUnsubscribeInput represents input for unsubscribing from web push notifications
type WebPushUnsubscribeInput struct {
	Endpoint string `json:"endpoint" validate:"required,url"`
}

// WebPushSubscriptionOutput represents a web push subscription in API response
type WebPushSubscriptionOutput struct {
	ID         int64      `json:"id"`
	DeviceName string     `json:"device_name,omitempty"`
	UserAgent  string     `json:"user_agent,omitempty"`
	IsActive   bool       `json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ToSubscriptionOutput converts a WebPushSubscription entity to output DTO
func ToSubscriptionOutput(sub *entities.WebPushSubscription) *WebPushSubscriptionOutput {
	return &WebPushSubscriptionOutput{
		ID:         sub.ID,
		DeviceName: sub.DeviceName,
		UserAgent:  sub.UserAgent,
		IsActive:   sub.IsActive,
		LastUsedAt: sub.LastUsedAt,
		CreatedAt:  sub.CreatedAt,
	}
}

// ToSubscriptionOutputList converts a list of WebPushSubscription entities to output DTOs
func ToSubscriptionOutputList(subs []*entities.WebPushSubscription) []*WebPushSubscriptionOutput {
	outputs := make([]*WebPushSubscriptionOutput, len(subs))
	for i, sub := range subs {
		outputs[i] = ToSubscriptionOutput(sub)
	}
	return outputs
}

// WebPushStatusOutput represents the web push subscription status for a user
type WebPushStatusOutput struct {
	IsEnabled     bool                         `json:"is_enabled"`
	Subscriptions []*WebPushSubscriptionOutput `json:"subscriptions"`
	TotalDevices  int                          `json:"total_devices"`
}

// WebPushVAPIDKeyOutput represents the VAPID public key response
type WebPushVAPIDKeyOutput struct {
	PublicKey string `json:"public_key"`
}

// WebPushTestInput represents input for testing a push notification
type WebPushTestInput struct {
	Title   string `json:"title" validate:"required,max=100"`
	Message string `json:"message" validate:"required,max=500"`
}
