// Package dto contains data transfer objects for the users module.
package dto

// CreateDepartmentInput represents input for creating a department.
type CreateDepartmentInput struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Code        string `json:"code" validate:"required,min=2,max=20,alphanum"`
	Description string `json:"description" validate:"omitempty,max=500"`
	ParentID    *int64 `json:"parent_id"`
}

// UpdateDepartmentInput represents input for updating a department.
type UpdateDepartmentInput struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Code        string `json:"code" validate:"required,min=2,max=20,alphanum"`
	Description string `json:"description" validate:"omitempty,max=500"`
	ParentID    *int64 `json:"parent_id"`
	HeadID      *int64 `json:"head_id"`
	IsActive    *bool  `json:"is_active"`
}

// DepartmentListResponse represents paginated department list response.
type DepartmentListResponse struct {
	Departments interface{} `json:"departments"`
	Total       int64       `json:"total"`
	Page        int         `json:"page"`
	Limit       int         `json:"limit"`
	TotalPages  int         `json:"total_pages"`
}
