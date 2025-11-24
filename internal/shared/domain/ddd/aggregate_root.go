// Package ddd provides Domain-Driven Design base types and patterns.
package ddd

// AggregateRoot is the base for all aggregate roots in DDD
type AggregateRoot struct {
	domainEvents []DomainEvent
}

// AddDomainEvent adds a domain event to the aggregate
func (a *AggregateRoot) AddDomainEvent(event DomainEvent) {
	a.domainEvents = append(a.domainEvents, event)
}

// GetDomainEvents returns all domain events
func (a *AggregateRoot) GetDomainEvents() []DomainEvent {
	return a.domainEvents
}

// ClearDomainEvents clears all domain events
func (a *AggregateRoot) ClearDomainEvents() {
	a.domainEvents = []DomainEvent{}
}
