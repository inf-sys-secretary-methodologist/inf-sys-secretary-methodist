// Package dto contains data transfer objects for the users module.
package dto

// UpdateUserProfileInput represents input for updating user profile.
type UpdateUserProfileInput struct {
	DepartmentID *int64 `json:"department_id"`
	PositionID   *int64 `json:"position_id"`
	Phone        string `json:"phone" validate:"omitempty,e164"`
	Avatar       string `json:"avatar" validate:"omitempty,url"`
	Bio          string `json:"bio" validate:"omitempty,max=500"`
}

// UpdateUserRoleInput represents input for changing user role.
type UpdateUserRoleInput struct {
	Role string `json:"role" validate:"required,oneof=system_admin methodist academic_secretary teacher student"`
}

// UpdateUserStatusInput represents input for changing user status.
type UpdateUserStatusInput struct {
	Status string `json:"status" validate:"required,oneof=active inactive blocked"`
}

// BulkUpdateDepartmentInput represents input for bulk department assignment.
type BulkUpdateDepartmentInput struct {
	UserIDs      []int64 `json:"user_ids" validate:"required,min=1"`
	DepartmentID *int64  `json:"department_id"`
}

// BulkUpdatePositionInput represents input for bulk position assignment.
type BulkUpdatePositionInput struct {
	UserIDs    []int64 `json:"user_ids" validate:"required,min=1"`
	PositionID *int64  `json:"position_id"`
}

// UserListFilter represents filter options for listing users.
type UserListFilter struct {
	DepartmentID *int64 `form:"department_id"`
	PositionID   *int64 `form:"position_id"`
	Role         string `form:"role"`
	Status       string `form:"status"`
	Search       string `form:"search"`
	Page         int    `form:"page"`
	Limit        int    `form:"limit"`
}

// UserListResponse represents paginated user list response.
type UserListResponse struct {
	Users      interface{} `json:"users"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}
