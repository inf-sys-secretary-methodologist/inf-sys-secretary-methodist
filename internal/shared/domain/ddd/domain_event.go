package ddd

import "time"

// DomainEvent represents a domain event in the system
type DomainEvent interface {
	GetEventType() string
	GetOccurredAt() time.Time
	GetAggregateID() string
}

// BaseDomainEvent provides common fields for all domain events
type BaseDomainEvent struct {
	EventType   string    `json:"event_type"`
	AggregateID string    `json:"aggregate_id"`
	OccurredAt  time.Time `json:"occurred_at"`
}

// GetEventType returns the event type
func (e BaseDomainEvent) GetEventType() string {
	return e.EventType
}

// GetOccurredAt returns when the event occurred
func (e BaseDomainEvent) GetOccurredAt() time.Time {
	return e.OccurredAt
}

// GetAggregateID returns the aggregate ID
func (e BaseDomainEvent) GetAggregateID() string {
	return e.AggregateID
}
