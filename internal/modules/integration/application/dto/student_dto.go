package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// ExternalStudentDTO represents an external student response
type ExternalStudentDTO struct {
	ID             int64      `json:"id"`
	ExternalID     string     `json:"external_id"`
	Code           string     `json:"code"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	MiddleName     string     `json:"middle_name,omitempty"`
	FullName       string     `json:"full_name"`
	Email          string     `json:"email,omitempty"`
	Phone          string     `json:"phone,omitempty"`
	GroupName      string     `json:"group_name,omitempty"`
	Faculty        string     `json:"faculty,omitempty"`
	Specialty      string     `json:"specialty,omitempty"`
	Course         int        `json:"course,omitempty"`
	StudyForm      string     `json:"study_form,omitempty"`
	EnrollmentDate *time.Time `json:"enrollment_date,omitempty"`
	ExpulsionDate  *time.Time `json:"expulsion_date,omitempty"`
	GraduationDate *time.Time `json:"graduation_date,omitempty"`
	Status         string     `json:"status"`
	IsActive       bool       `json:"is_active"`
	IsLinked       bool       `json:"is_linked"`
	LocalUserID    *int64     `json:"local_user_id,omitempty"`
	LastSyncAt     time.Time  `json:"last_sync_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// FromExternalStudent converts entity to DTO
func FromExternalStudent(student *entities.ExternalStudent) *ExternalStudentDTO {
	return &ExternalStudentDTO{
		ID:             student.ID,
		ExternalID:     student.ExternalID,
		Code:           student.Code,
		FirstName:      student.FirstName,
		LastName:       student.LastName,
		MiddleName:     student.MiddleName,
		FullName:       student.GetFullName(),
		Email:          student.Email,
		Phone:          student.Phone,
		GroupName:      student.GroupName,
		Faculty:        student.Faculty,
		Specialty:      student.Specialty,
		Course:         student.Course,
		StudyForm:      student.StudyForm,
		EnrollmentDate: student.EnrollmentDate,
		ExpulsionDate:  student.ExpulsionDate,
		GraduationDate: student.GraduationDate,
		Status:         student.Status,
		IsActive:       student.IsActive,
		IsLinked:       student.IsLinked(),
		LocalUserID:    student.LocalUserID,
		LastSyncAt:     student.LastSyncAt,
		CreatedAt:      student.CreatedAt,
		UpdatedAt:      student.UpdatedAt,
	}
}

// ExternalStudentListRequest represents a request to list external students
type ExternalStudentListRequest struct {
	Search    string `json:"search,omitempty" form:"search"`
	GroupName string `json:"group_name,omitempty" form:"group_name"`
	Faculty   string `json:"faculty,omitempty" form:"faculty"`
	Course    *int   `json:"course,omitempty" form:"course"`
	Status    string `json:"status,omitempty" form:"status"`
	IsActive  *bool  `json:"is_active,omitempty" form:"is_active"`
	IsLinked  *bool  `json:"is_linked,omitempty" form:"is_linked"`
	Limit     int    `json:"limit,omitempty" form:"limit"`
	Offset    int    `json:"offset,omitempty" form:"offset"`
}

// ExternalStudentListResponse represents a paginated list of external students
type ExternalStudentListResponse struct {
	Items []*ExternalStudentDTO `json:"items"`
	Total int64                 `json:"total"`
}

// LinkStudentRequest represents a request to link an external student to a local user
type LinkStudentRequest struct {
	LocalUserID int64 `json:"local_user_id" validate:"required"`
}

// GroupsResponse represents a list of available groups
type GroupsResponse struct {
	Groups []string `json:"groups"`
}

// FacultiesResponse represents a list of available faculties
type FacultiesResponse struct {
	Faculties []string `json:"faculties"`
}
