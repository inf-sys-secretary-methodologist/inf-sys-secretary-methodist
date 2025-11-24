package ddd

import (
	"testing"
	"time"
)

func TestEntity_Touch(t *testing.T) {
	entity := Entity{
		ID:        "test-id",
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	oldUpdatedAt := entity.UpdatedAt
	time.Sleep(10 * time.Millisecond)
	entity.Touch()

	if !entity.UpdatedAt.After(oldUpdatedAt) {
		t.Errorf("expected UpdatedAt to be updated after Touch()")
	}
}

func TestAggregateRoot_AddDomainEvent(t *testing.T) {
	ar := &AggregateRoot{}
	event := newMockDomainEvent("test-event", "agg-1")

	ar.AddDomainEvent(event)

	events := ar.GetDomainEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
}

func TestAggregateRoot_GetDomainEvents(t *testing.T) {
	ar := &AggregateRoot{}
	event1 := newMockDomainEvent("event1", "agg-1")
	event2 := newMockDomainEvent("event2", "agg-1")

	ar.AddDomainEvent(event1)
	ar.AddDomainEvent(event2)

	events := ar.GetDomainEvents()
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestAggregateRoot_ClearDomainEvents(t *testing.T) {
	ar := &AggregateRoot{}
	event := newMockDomainEvent("test-event", "agg-1")

	ar.AddDomainEvent(event)
	ar.ClearDomainEvents()

	events := ar.GetDomainEvents()
	if len(events) != 0 {
		t.Errorf("expected 0 events after clear, got %d", len(events))
	}
}

type mockDomainEvent struct {
	BaseDomainEvent
}

func newMockDomainEvent(eventType, aggregateID string) *mockDomainEvent {
	return &mockDomainEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventType:   eventType,
			AggregateID: aggregateID,
			OccurredAt:  time.Now(),
		},
	}
}
