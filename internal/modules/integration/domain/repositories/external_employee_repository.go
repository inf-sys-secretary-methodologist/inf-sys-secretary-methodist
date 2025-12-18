package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// ExternalEmployeeRepository defines the interface for external employee persistence
type ExternalEmployeeRepository interface {
	// Create creates a new external employee record
	Create(ctx context.Context, employee *entities.ExternalEmployee) error

	// Update updates an existing external employee record
	Update(ctx context.Context, employee *entities.ExternalEmployee) error

	// Upsert creates or updates an external employee by external ID
	Upsert(ctx context.Context, employee *entities.ExternalEmployee) error

	// GetByID retrieves an external employee by ID
	GetByID(ctx context.Context, id int64) (*entities.ExternalEmployee, error)

	// GetByExternalID retrieves an external employee by 1C external ID
	GetByExternalID(ctx context.Context, externalID string) (*entities.ExternalEmployee, error)

	// GetByCode retrieves an external employee by 1C code
	GetByCode(ctx context.Context, code string) (*entities.ExternalEmployee, error)

	// GetByLocalUserID retrieves an external employee by linked local user ID
	GetByLocalUserID(ctx context.Context, localUserID int64) (*entities.ExternalEmployee, error)

	// List retrieves external employees with optional filtering
	List(ctx context.Context, filter entities.ExternalEmployeeFilter) ([]*entities.ExternalEmployee, int64, error)

	// GetUnlinked retrieves external employees not linked to local users
	GetUnlinked(ctx context.Context, limit, offset int) ([]*entities.ExternalEmployee, int64, error)

	// LinkToLocalUser links an external employee to a local user
	LinkToLocalUser(ctx context.Context, id int64, localUserID int64) error

	// Unlink removes the link between external employee and local user
	Unlink(ctx context.Context, id int64) error

	// Delete deletes an external employee record
	Delete(ctx context.Context, id int64) error

	// GetAllExternalIDs retrieves all external IDs for change detection
	GetAllExternalIDs(ctx context.Context) ([]string, error)

	// BulkUpsert creates or updates multiple external employees
	BulkUpsert(ctx context.Context, employees []*entities.ExternalEmployee) error

	// MarkInactiveExcept marks all employees as inactive except those with given external IDs
	MarkInactiveExcept(ctx context.Context, activeExternalIDs []string) error
}
