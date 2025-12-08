// Package entities contains domain entities for the users module.
package entities

import "time"

// Department represents an organizational department/division.
type Department struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Code        string    `db:"code" json:"code"`           // Краткий код подразделения (например, "IT", "HR")
	Description string    `db:"description" json:"description,omitempty"`
	ParentID    *int64    `db:"parent_id" json:"parent_id,omitempty"` // Для иерархии подразделений
	HeadID      *int64    `db:"head_id" json:"head_id,omitempty"`     // ID руководителя подразделения
	IsActive    bool      `db:"is_active" json:"is_active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
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
