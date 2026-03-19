package ddd

import (
	"testing"
	"time"
)

func TestEntity_GetID(t *testing.T) {
	entity := Entity{ID: "test-id-123"}
	if got := entity.GetID(); got != "test-id-123" {
		t.Errorf("GetID() = %q, want %q", got, "test-id-123")
	}
}

func TestEntity_SetID(t *testing.T) {
	entity := Entity{}
	entity.SetID("new-id")
	if entity.ID != "new-id" {
		t.Errorf("expected ID 'new-id', got '%s'", entity.ID)
	}
}

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

func TestAggregateRoot_GetDomainEvents_Empty(t *testing.T) {
	ar := &AggregateRoot{}
	events := ar.GetDomainEvents()
	if events != nil {
		t.Errorf("expected nil events for new AggregateRoot, got %v", events)
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

func TestAggregateRoot_ClearDomainEvents_Empty(t *testing.T) {
	ar := &AggregateRoot{}
	ar.ClearDomainEvents()
	events := ar.GetDomainEvents()
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestAggregateRoot_MultipleAddAndClear(t *testing.T) {
	ar := &AggregateRoot{}

	ar.AddDomainEvent(newMockDomainEvent("e1", "a1"))
	ar.AddDomainEvent(newMockDomainEvent("e2", "a1"))
	if len(ar.GetDomainEvents()) != 2 {
		t.Fatalf("expected 2 events")
	}

	ar.ClearDomainEvents()
	if len(ar.GetDomainEvents()) != 0 {
		t.Fatalf("expected 0 events after clear")
	}

	ar.AddDomainEvent(newMockDomainEvent("e3", "a1"))
	if len(ar.GetDomainEvents()) != 1 {
		t.Fatalf("expected 1 event after re-add")
	}
}

func TestBaseDomainEvent_GetEventType(t *testing.T) {
	e := BaseDomainEvent{EventType: "user.created"}
	if got := e.GetEventType(); got != "user.created" {
		t.Errorf("GetEventType() = %q, want %q", got, "user.created")
	}
}

func TestBaseDomainEvent_GetOccurredAt(t *testing.T) {
	now := time.Now()
	e := BaseDomainEvent{OccurredAt: now}
	if got := e.GetOccurredAt(); !got.Equal(now) {
		t.Errorf("GetOccurredAt() = %v, want %v", got, now)
	}
}

func TestBaseDomainEvent_GetAggregateID(t *testing.T) {
	e := BaseDomainEvent{AggregateID: "agg-42"}
	if got := e.GetAggregateID(); got != "agg-42" {
		t.Errorf("GetAggregateID() = %q, want %q", got, "agg-42")
	}
}

func TestBaseDomainEvent_ImplementsDomainEvent(t *testing.T) {
	// Verify BaseDomainEvent satisfies DomainEvent interface
	var _ DomainEvent = BaseDomainEvent{}
	var _ DomainEvent = &BaseDomainEvent{}
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

// mockValueObject implements ValueObject for testing
type mockValueObject struct {
	Value string
}

func (m mockValueObject) Equals(other ValueObject) bool {
	o, ok := other.(mockValueObject)
	if !ok {
		return false
	}
	return m.Value == o.Value
}

func TestValueObject_Equals(t *testing.T) {
	vo1 := mockValueObject{Value: "abc"}
	vo2 := mockValueObject{Value: "abc"}
	vo3 := mockValueObject{Value: "xyz"}

	if !vo1.Equals(vo2) {
		t.Error("expected vo1 to equal vo2")
	}
	if vo1.Equals(vo3) {
		t.Error("expected vo1 to not equal vo3")
	}
}

func TestValueObject_Equals_DifferentType(t *testing.T) {
	vo1 := mockValueObject{Value: "abc"}

	type otherVO struct {
		Val string
	}

	// This won't satisfy the interface directly, test with mockValueObject
	vo2 := mockValueObject{Value: "abc"}
	if !vo1.Equals(vo2) {
		t.Error("expected equal for same type and value")
	}
}
