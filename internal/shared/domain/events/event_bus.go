package events

import (
	"sync"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/common"
)

// EventHandler handles domain events
type EventHandler interface {
	Handle(event common.DomainEvent) error
	GetHandlerName() string
}

// EventBus manages event publishing and subscription
type EventBus interface {
	Subscribe(eventType string, handler EventHandler) error
	Unsubscribe(eventType string, handler EventHandler) error
	Publish(event common.DomainEvent) error
}

// InMemoryEventBus is an in-memory implementation of EventBus
type InMemoryEventBus struct {
	handlers map[string][]EventHandler
	mutex    sync.RWMutex
}

// NewInMemoryEventBus creates a new in-memory event bus
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Subscribe registers an event handler for a specific event type
func (bus *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	bus.handlers[eventType] = append(bus.handlers[eventType], handler)
	return nil
}

// Unsubscribe removes an event handler
func (bus *InMemoryEventBus) Unsubscribe(eventType string, handler EventHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	handlers := bus.handlers[eventType]
	for i, h := range handlers {
		if h.GetHandlerName() == handler.GetHandlerName() {
			bus.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	return nil
}

// Publish publishes an event to all registered handlers
func (bus *InMemoryEventBus) Publish(event common.DomainEvent) error {
	bus.mutex.RLock()
	handlers := bus.handlers[event.GetEventType()]
	bus.mutex.RUnlock()

	for _, handler := range handlers {
		// Execute handlers asynchronously
		go func(h EventHandler, e common.DomainEvent) {
			_ = h.Handle(e) // TODO: Add proper error logging
		}(handler, event)
	}

	return nil
}
