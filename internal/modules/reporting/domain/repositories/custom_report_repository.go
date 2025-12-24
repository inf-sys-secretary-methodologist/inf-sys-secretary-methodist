package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
)

// CustomReportFilter represents filter options for listing custom reports
type CustomReportFilter struct {
	CreatedBy  *int64
	DataSource *entities.DataSourceType
	IsPublic   *bool
	Search     string
	Page       int
	PageSize   int
}

// CustomReportRepository defines the interface for custom report persistence
type CustomReportRepository interface {
	// Create creates a new custom report
	Create(ctx context.Context, report *entities.CustomReport) error

	// Update updates an existing custom report
	Update(ctx context.Context, report *entities.CustomReport) error

	// GetByID retrieves a custom report by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.CustomReport, error)

	// Delete deletes a custom report by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// List lists custom reports with filtering and pagination
	List(ctx context.Context, filter CustomReportFilter) ([]*entities.CustomReport, error)

	// Count counts custom reports matching the filter
	Count(ctx context.Context, filter CustomReportFilter) (int64, error)

	// GetByCreator retrieves all custom reports created by a user
	GetByCreator(ctx context.Context, creatorID int64, page, pageSize int) ([]*entities.CustomReport, error)

	// GetPublicReports retrieves all public custom reports
	GetPublicReports(ctx context.Context, page, pageSize int) ([]*entities.CustomReport, error)
}
