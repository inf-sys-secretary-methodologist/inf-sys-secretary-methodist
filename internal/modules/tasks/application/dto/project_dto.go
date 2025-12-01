package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
)

// CreateProjectInput represents input for creating a project.
type CreateProjectInput struct {
	Name        string     `json:"name" validate:"required,min=1,max=255"`
	Description *string    `json:"description,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

// UpdateProjectInput represents input for updating a project.
type UpdateProjectInput struct {
	Name        *string    `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string    `json:"description,omitempty"`
	Status      *string    `json:"status,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

// ProjectFilterInput represents input for filtering projects.
type ProjectFilterInput struct {
	OwnerID *int64  `form:"owner_id"`
	Status  *string `form:"status"`
	Search  *string `form:"search"`
	Limit   int     `form:"limit,default=20"`
	Offset  int     `form:"offset,default=0"`
}

// ToProjectFilter converts ProjectFilterInput to domain ProjectFilter.
func (f *ProjectFilterInput) ToProjectFilter() repositories.ProjectFilter {
	filter := repositories.ProjectFilter{
		OwnerID: f.OwnerID,
		Search:  f.Search,
	}
	if f.Status != nil {
		status := domain.ProjectStatus(*f.Status)
		filter.Status = &status
	}
	return filter
}

// ProjectOutput represents the output for a project.
type ProjectOutput struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	OwnerID     int64      `json:"owner_id"`
	Status      string     `json:"status"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	TaskCount   int        `json:"task_count,omitempty"`
}

// ProjectListOutput represents the output for a list of projects.
type ProjectListOutput struct {
	Projects []ProjectOutput `json:"projects"`
	Total    int64           `json:"total"`
	Limit    int             `json:"limit"`
	Offset   int             `json:"offset"`
}

// ToProjectOutput converts a Project entity to ProjectOutput.
func ToProjectOutput(project *entities.Project) ProjectOutput {
	return ProjectOutput{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		OwnerID:     project.OwnerID,
		Status:      string(project.Status),
		StartDate:   project.StartDate,
		EndDate:     project.EndDate,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
		TaskCount:   len(project.Tasks),
	}
}
