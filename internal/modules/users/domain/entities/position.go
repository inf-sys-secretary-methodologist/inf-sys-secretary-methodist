// Package entities contains domain entities for the users module.
package entities

import "time"

// Position represents a job position/title within the organization.
type Position struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Code        string    `db:"code" json:"code"`           // Краткий код должности
	Description string    `db:"description" json:"description,omitempty"`
	Level       int       `db:"level" json:"level"`         // Уровень должности (для иерархии)
	IsActive    bool      `db:"is_active" json:"is_active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
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
