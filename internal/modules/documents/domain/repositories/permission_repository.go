package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// PermissionRepository defines the interface for document permission persistence
type PermissionRepository interface {
	// Permission CRUD
	Create(ctx context.Context, permission *entities.DocumentPermission) error
	Update(ctx context.Context, permission *entities.DocumentPermission) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*entities.DocumentPermission, error)

	// Query operations
	GetByDocumentID(ctx context.Context, documentID int64) ([]*entities.DocumentPermission, error)
	GetByUserID(ctx context.Context, userID int64) ([]*entities.DocumentPermission, error)
	GetByUserIDOrRole(ctx context.Context, userID int64, role string) ([]*entities.DocumentPermission, error)
	GetByGrantedBy(ctx context.Context, userID int64) ([]*entities.DocumentPermission, error)
	GetByDocumentAndUser(ctx context.Context, documentID, userID int64) (*entities.DocumentPermission, error)
	GetByDocumentAndRole(ctx context.Context, documentID int64, role entities.UserRole) (*entities.DocumentPermission, error)

	// Permission checks
	HasPermission(ctx context.Context, documentID, userID int64, permission entities.PermissionLevel) (bool, error)
	HasAnyPermission(ctx context.Context, documentID, userID int64) (bool, error)
	GetUserPermissionLevel(ctx context.Context, documentID, userID int64, userRole entities.UserRole) (*entities.PermissionLevel, error)

	// Bulk operations
	DeleteByDocumentID(ctx context.Context, documentID int64) error
	DeleteByUserID(ctx context.Context, userID int64) error
	DeleteExpired(ctx context.Context) (int64, error)
}

// PublicLinkRepository defines the interface for public link persistence
type PublicLinkRepository interface {
	// CRUD operations
	Create(ctx context.Context, link *entities.PublicLink) error
	Update(ctx context.Context, link *entities.PublicLink) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*entities.PublicLink, error)
	GetByToken(ctx context.Context, token string) (*entities.PublicLink, error)

	// Query operations
	GetByDocumentID(ctx context.Context, documentID int64) ([]*entities.PublicLink, error)
	GetByCreatedBy(ctx context.Context, userID int64) ([]*entities.PublicLink, error)
	GetActiveByDocumentID(ctx context.Context, documentID int64) ([]*entities.PublicLink, error)

	// Link management
	IncrementUseCount(ctx context.Context, id int64) error
	Deactivate(ctx context.Context, id int64) error
	Activate(ctx context.Context, id int64) error

	// Bulk operations
	DeleteByDocumentID(ctx context.Context, documentID int64) error
	DeactivateExpired(ctx context.Context) (int64, error)
}
