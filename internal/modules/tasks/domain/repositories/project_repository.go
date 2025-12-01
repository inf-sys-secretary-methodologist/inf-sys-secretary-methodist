package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
)

// ProjectFilter defines filtering options for project queries.
type ProjectFilter struct {
	OwnerID *int64
	Status  *domain.ProjectStatus
	Search  *string
}

// ProjectRepository defines the interface for project data access.
type ProjectRepository interface {
	// CRUD operations
	Create(ctx context.Context, project *entities.Project) error
	Save(ctx context.Context, project *entities.Project) error
	GetByID(ctx context.Context, id int64) (*entities.Project, error)
	Delete(ctx context.Context, id int64) error

	// Query operations
	List(ctx context.Context, filter ProjectFilter, limit, offset int) ([]*entities.Project, error)
	Count(ctx context.Context, filter ProjectFilter) (int64, error)
	GetByOwner(ctx context.Context, ownerID int64, limit, offset int) ([]*entities.Project, error)
	GetByStatus(ctx context.Context, status domain.ProjectStatus, limit, offset int) ([]*entities.Project, error)
}
