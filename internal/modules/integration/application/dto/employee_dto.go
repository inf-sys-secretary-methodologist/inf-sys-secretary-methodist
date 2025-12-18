package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// ExternalEmployeeDTO represents an external employee response
type ExternalEmployeeDTO struct {
	ID             int64      `json:"id"`
	ExternalID     string     `json:"external_id"`
	Code           string     `json:"code"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	MiddleName     string     `json:"middle_name,omitempty"`
	FullName       string     `json:"full_name"`
	Email          string     `json:"email,omitempty"`
	Phone          string     `json:"phone,omitempty"`
	Position       string     `json:"position,omitempty"`
	Department     string     `json:"department,omitempty"`
	EmploymentDate *time.Time `json:"employment_date,omitempty"`
	DismissalDate  *time.Time `json:"dismissal_date,omitempty"`
	IsActive       bool       `json:"is_active"`
	IsLinked       bool       `json:"is_linked"`
	LocalUserID    *int64     `json:"local_user_id,omitempty"`
	LastSyncAt     time.Time  `json:"last_sync_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// FromExternalEmployee converts entity to DTO
func FromExternalEmployee(emp *entities.ExternalEmployee) *ExternalEmployeeDTO {
	return &ExternalEmployeeDTO{
		ID:             emp.ID,
		ExternalID:     emp.ExternalID,
		Code:           emp.Code,
		FirstName:      emp.FirstName,
		LastName:       emp.LastName,
		MiddleName:     emp.MiddleName,
		FullName:       emp.GetFullName(),
		Email:          emp.Email,
		Phone:          emp.Phone,
		Position:       emp.Position,
		Department:     emp.Department,
		EmploymentDate: emp.EmploymentDate,
		DismissalDate:  emp.DismissalDate,
		IsActive:       emp.IsActive,
		IsLinked:       emp.IsLinked(),
		LocalUserID:    emp.LocalUserID,
		LastSyncAt:     emp.LastSyncAt,
		CreatedAt:      emp.CreatedAt,
		UpdatedAt:      emp.UpdatedAt,
	}
}

// ExternalEmployeeListRequest represents a request to list external employees
type ExternalEmployeeListRequest struct {
	Search     string `json:"search,omitempty" form:"search"`
	Department string `json:"department,omitempty" form:"department"`
	Position   string `json:"position,omitempty" form:"position"`
	IsActive   *bool  `json:"is_active,omitempty" form:"is_active"`
	IsLinked   *bool  `json:"is_linked,omitempty" form:"is_linked"`
	Limit      int    `json:"limit,omitempty" form:"limit"`
	Offset     int    `json:"offset,omitempty" form:"offset"`
}

// ExternalEmployeeListResponse represents a paginated list of external employees
type ExternalEmployeeListResponse struct {
	Items []*ExternalEmployeeDTO `json:"items"`
	Total int64                  `json:"total"`
}

// LinkEmployeeRequest represents a request to link an external employee to a local user
type LinkEmployeeRequest struct {
	LocalUserID int64 `json:"local_user_id" validate:"required"`
}
