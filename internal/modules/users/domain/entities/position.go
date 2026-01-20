// Package entities contains domain entities for the users module.
package entities

import "time"

// Position represents a job position/title within the organization.
type Position struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"` // Краткий код должности
	Description string    `json:"description,omitempty"`
	Level       int       `json:"level"` // Уровень должности (для иерархии)
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewPosition creates a new position instance.
func NewPosition(name, code, description string, level int) *Position {
	now := time.Now()
	return &Position{
		Name:        name,
		Code:        code,
		Description: description,
		Level:       level,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
