package ddd

import "time"

// Entity is the base for all entities
type Entity struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetID returns the entity ID
func (e *Entity) GetID() string {
	return e.ID
}

// SetID sets the entity ID
func (e *Entity) SetID(id string) {
	e.ID = id
}

// Touch updates the UpdatedAt timestamp
func (e *Entity) Touch() {
	e.UpdatedAt = time.Now()
}
