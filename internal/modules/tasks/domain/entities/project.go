package entities

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
)

// Project represents a project for grouping tasks.
type Project struct {
	ID          int64                `json:"id"`
	Name        string               `json:"name"`
	Description *string              `json:"description,omitempty"`
	OwnerID     int64                `json:"owner_id"`
	Status      domain.ProjectStatus `json:"status"`
	StartDate   *time.Time           `json:"start_date,omitempty"`
	EndDate     *time.Time           `json:"end_date,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`

	// Associations
	Tasks []Task `json:"tasks,omitempty"`
}

// NewProject creates a new project with default values.
func NewProject(name string, ownerID int64) *Project {
	now := time.Now()
	return &Project{
		Name:      name,
		OwnerID:   ownerID,
		Status:    domain.ProjectStatusPlanning,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Activate sets the project status to active.
func (p *Project) Activate() {
	p.Status = domain.ProjectStatusActive
	p.UpdatedAt = time.Now()
}

// PutOnHold sets the project status to on_hold.
func (p *Project) PutOnHold() {
	p.Status = domain.ProjectStatusOnHold
	p.UpdatedAt = time.Now()
}

// Complete marks the project as completed.
func (p *Project) Complete() {
	p.Status = domain.ProjectStatusCompleted
	p.UpdatedAt = time.Now()
}

// Cancel cancels the project.
func (p *Project) Cancel() {
	p.Status = domain.ProjectStatusCancelled
	p.UpdatedAt = time.Now()
}

// IsActive checks if the project is active.
func (p *Project) IsActive() bool {
	return p.Status == domain.ProjectStatusActive
}
