// Package dto contains data transfer objects for the users module.
package dto

// CreatePositionInput represents input for creating a position.
type CreatePositionInput struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Code        string `json:"code" validate:"required,min=2,max=20,alphanum"`
	Description string `json:"description" validate:"omitempty,max=500"`
	Level       int    `json:"level" validate:"gte=0,lte=100"`
}

// UpdatePositionInput represents input for updating a position.
type UpdatePositionInput struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Code        string `json:"code" validate:"required,min=2,max=20,alphanum"`
	Description string `json:"description" validate:"omitempty,max=500"`
	Level       int    `json:"level" validate:"gte=0,lte=100"`
	IsActive    *bool  `json:"is_active"`
}

// PositionListResponse represents paginated position list response.
type PositionListResponse struct {
	Positions  interface{} `json:"positions"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}
