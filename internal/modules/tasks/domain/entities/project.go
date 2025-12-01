package entities

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
)

// Project represents a project for grouping tasks.
type Project struct {
	ID          int64                `db:"id" json:"id"`
	Name        string               `db:"name" json:"name"`
	Description *string              `db:"description" json:"description,omitempty"`
	OwnerID     int64                `db:"owner_id" json:"owner_id"`
	Status      domain.ProjectStatus `db:"status" json:"status"`
	StartDate   *time.Time           `db:"start_date" json:"start_date,omitempty"`
	EndDate     *time.Time           `db:"end_date" json:"end_date,omitempty"`
	CreatedAt   time.Time            `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time            `db:"updated_at" json:"updated_at"`

	// Associations
	Tasks []Task `db:"-" json:"tasks,omitempty"`
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
