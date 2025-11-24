package ddd

// ValueObject is a marker interface for value objects
// Value objects are immutable and defined by their attributes
type ValueObject interface {
	Equals(other ValueObject) bool
}
