package events

import (
	"sync"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/ddd"
	"github.com/stretchr/testify/assert"
)

// testEvent implements ddd.DomainEvent for testing
type testEvent struct {
	ddd.BaseDomainEvent
}

func newTestEvent(eventType, aggregateID string) *testEvent {
	return &testEvent{
		BaseDomainEvent: ddd.BaseDomainEvent{
			EventType:   eventType,
			AggregateID: aggregateID,
			OccurredAt:  time.Now(),
		},
	}
}

// testHandler implements EventHandler for testing
type testHandler struct {
	name       string
	mu         sync.Mutex
	events     []ddd.DomainEvent
	handleFunc func(event ddd.DomainEvent) error
}

func newTestHandler(name string) *testHandler {
	return &testHandler{
		name: name,
	}
}

func (h *testHandler) Handle(event ddd.DomainEvent) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, event)
	if h.handleFunc != nil {
		return h.handleFunc(event)
	}
	return nil
}

func (h *testHandler) GetHandlerName() string {
	return h.name
}

func (h *testHandler) receivedCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.events)
}

func TestNewInMemoryEventBus(t *testing.T) {
	bus := NewInMemoryEventBus()

	assert.NotNil(t, bus)
	assert.NotNil(t, bus.handlers)
	assert.Empty(t, bus.handlers)
}

func TestInMemoryEventBus_Subscribe(t *testing.T) {
	bus := NewInMemoryEventBus()
	handler := newTestHandler("test-handler")

	err := bus.Subscribe("user.created", handler)

	assert.NoError(t, err)
	assert.Len(t, bus.handlers["user.created"], 1)
}

func TestInMemoryEventBus_Subscribe_Multiple(t *testing.T) {
	bus := NewInMemoryEventBus()
	handler1 := newTestHandler("handler-1")
	handler2 := newTestHandler("handler-2")

	err := bus.Subscribe("user.created", handler1)
	assert.NoError(t, err)

	err = bus.Subscribe("user.created", handler2)
	assert.NoError(t, err)

	assert.Len(t, bus.handlers["user.created"], 2)
}

func TestInMemoryEventBus_Subscribe_DifferentEvents(t *testing.T) {
	bus := NewInMemoryEventBus()
	handler1 := newTestHandler("handler-1")
	handler2 := newTestHandler("handler-2")

	_ = bus.Subscribe("user.created", handler1)
	_ = bus.Subscribe("user.deleted", handler2)

	assert.Len(t, bus.handlers["user.created"], 1)
	assert.Len(t, bus.handlers["user.deleted"], 1)
}

func TestInMemoryEventBus_Unsubscribe(t *testing.T) {
	bus := NewInMemoryEventBus()
	handler := newTestHandler("test-handler")

	_ = bus.Subscribe("user.created", handler)
	assert.Len(t, bus.handlers["user.created"], 1)

	err := bus.Unsubscribe("user.created", handler)
	assert.NoError(t, err)
	assert.Empty(t, bus.handlers["user.created"])
}

func TestInMemoryEventBus_Unsubscribe_NonExistent(t *testing.T) {
	bus := NewInMemoryEventBus()
	handler := newTestHandler("test-handler")

	err := bus.Unsubscribe("user.created", handler)
	assert.NoError(t, err)
}

func TestInMemoryEventBus_Unsubscribe_ByName(t *testing.T) {
	bus := NewInMemoryEventBus()
	handler1 := newTestHandler("handler-1")
	handler2 := newTestHandler("handler-2")

	_ = bus.Subscribe("user.created", handler1)
	_ = bus.Subscribe("user.created", handler2)
	assert.Len(t, bus.handlers["user.created"], 2)

	err := bus.Unsubscribe("user.created", handler1)
	assert.NoError(t, err)
	assert.Len(t, bus.handlers["user.created"], 1)
	assert.Equal(t, "handler-2", bus.handlers["user.created"][0].GetHandlerName())
}

func TestInMemoryEventBus_Publish(t *testing.T) {
	bus := NewInMemoryEventBus()
	handler := newTestHandler("test-handler")

	_ = bus.Subscribe("user.created", handler)

	event := newTestEvent("user.created", "user-123")
	err := bus.Publish(event)
	assert.NoError(t, err)

	// Wait for async handler execution
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 1, handler.receivedCount())
}

func TestInMemoryEventBus_Publish_MultipleHandlers(t *testing.T) {
	bus := NewInMemoryEventBus()
	handler1 := newTestHandler("handler-1")
	handler2 := newTestHandler("handler-2")

	_ = bus.Subscribe("user.created", handler1)
	_ = bus.Subscribe("user.created", handler2)

	event := newTestEvent("user.created", "user-123")
	err := bus.Publish(event)
	assert.NoError(t, err)

	// Wait for async handler execution
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 1, handler1.receivedCount())
	assert.Equal(t, 1, handler2.receivedCount())
}

func TestInMemoryEventBus_Publish_NoHandlers(t *testing.T) {
	bus := NewInMemoryEventBus()

	event := newTestEvent("user.created", "user-123")
	err := bus.Publish(event)
	assert.NoError(t, err)
}

func TestInMemoryEventBus_Publish_WrongEventType(t *testing.T) {
	bus := NewInMemoryEventBus()
	handler := newTestHandler("test-handler")

	_ = bus.Subscribe("user.created", handler)

	event := newTestEvent("user.deleted", "user-123")
	err := bus.Publish(event)
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 0, handler.receivedCount())
}

func TestInMemoryEventBus_ImplementsInterface(t *testing.T) {
	var _ EventBus = &InMemoryEventBus{}
}

func TestInMemoryEventBus_ConcurrentPublish(t *testing.T) {
	bus := NewInMemoryEventBus()
	handler := newTestHandler("test-handler")

	_ = bus.Subscribe("user.created", handler)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			event := newTestEvent("user.created", "user-123")
			_ = bus.Publish(event)
		}(i)
	}
	wg.Wait()

	// Wait for all async handlers to complete
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 10, handler.receivedCount())
}
