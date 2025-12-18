package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// ExternalStudentRepository defines the interface for external student persistence
type ExternalStudentRepository interface {
	// Create creates a new external student record
	Create(ctx context.Context, student *entities.ExternalStudent) error

	// Update updates an existing external student record
	Update(ctx context.Context, student *entities.ExternalStudent) error

	// Upsert creates or updates an external student by external ID
	Upsert(ctx context.Context, student *entities.ExternalStudent) error

	// GetByID retrieves an external student by ID
	GetByID(ctx context.Context, id int64) (*entities.ExternalStudent, error)

	// GetByExternalID retrieves an external student by 1C external ID
	GetByExternalID(ctx context.Context, externalID string) (*entities.ExternalStudent, error)

	// GetByCode retrieves an external student by 1C code (зачетка)
	GetByCode(ctx context.Context, code string) (*entities.ExternalStudent, error)

	// GetByLocalUserID retrieves an external student by linked local user ID
	GetByLocalUserID(ctx context.Context, localUserID int64) (*entities.ExternalStudent, error)

	// List retrieves external students with optional filtering
	List(ctx context.Context, filter entities.ExternalStudentFilter) ([]*entities.ExternalStudent, int64, error)

	// GetUnlinked retrieves external students not linked to local users
	GetUnlinked(ctx context.Context, limit, offset int) ([]*entities.ExternalStudent, int64, error)

	// GetByGroup retrieves external students by group name
	GetByGroup(ctx context.Context, groupName string) ([]*entities.ExternalStudent, error)

	// GetByFaculty retrieves external students by faculty
	GetByFaculty(ctx context.Context, faculty string) ([]*entities.ExternalStudent, error)

	// LinkToLocalUser links an external student to a local user
	LinkToLocalUser(ctx context.Context, id int64, localUserID int64) error

	// Unlink removes the link between external student and local user
	Unlink(ctx context.Context, id int64) error

	// Delete deletes an external student record
	Delete(ctx context.Context, id int64) error

	// GetAllExternalIDs retrieves all external IDs for change detection
	GetAllExternalIDs(ctx context.Context) ([]string, error)

	// BulkUpsert creates or updates multiple external students
	BulkUpsert(ctx context.Context, students []*entities.ExternalStudent) error

	// MarkInactiveExcept marks all students as inactive except those with given external IDs
	MarkInactiveExcept(ctx context.Context, activeExternalIDs []string) error

	// GetGroups retrieves distinct group names
	GetGroups(ctx context.Context) ([]string, error)

	// GetFaculties retrieves distinct faculty names
	GetFaculties(ctx context.Context) ([]string, error)
}
