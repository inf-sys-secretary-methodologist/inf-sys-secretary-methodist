// Package entities contains domain entities for the users module.
package entities

import "time"

// Department represents an organizational department/division.
type Department struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"` // Краткий код подразделения (например, "IT", "HR")
	Description string    `json:"description,omitempty"`
	ParentID    *int64    `json:"parent_id,omitempty"` // Для иерархии подразделений
	HeadID      *int64    `json:"head_id,omitempty"`   // ID руководителя подразделения
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewDepartment creates a new department instance.
func NewDepartment(name, code, description string, parentID *int64) *Department {
	now := time.Now()
	return &Department{
		Name:        name,
		Code:        code,
		Description: description,
		ParentID:    parentID,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
