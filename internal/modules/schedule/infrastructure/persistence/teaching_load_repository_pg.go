package persistence

import (
	"context"
	"database/sql"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// TeachingLoadRepositoryPG implements TeachingLoadRepository on PostgreSQL.
type TeachingLoadRepositoryPG struct {
	db *sql.DB
}

var _ usecases.TeachingLoadRepository = (*TeachingLoadRepositoryPG)(nil)

// NewTeachingLoadRepositoryPG creates a new TeachingLoadRepositoryPG.
func NewTeachingLoadRepositoryPG(db *sql.DB) *TeachingLoadRepositoryPG {
	return &TeachingLoadRepositoryPG{db: db}
}

// Create inserts a new load line. STUB — see GREEN commit.
func (r *TeachingLoadRepositoryPG) Create(ctx context.Context, load *entities.TeachingLoad) error {
	return nil
}

// Update mutates an existing line. STUB — see GREEN commit.
func (r *TeachingLoadRepositoryPG) Update(ctx context.Context, load *entities.TeachingLoad) error {
	return nil
}

// Delete removes a line by id. STUB — see GREEN commit.
func (r *TeachingLoadRepositoryPG) Delete(ctx context.Context, id int64) error {
	return nil
}

// GetByID returns one hydrated line. STUB — see GREEN commit.
func (r *TeachingLoadRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.TeachingLoad, error) {
	return nil, nil
}

// List returns hydrated load lines. STUB — see GREEN commit.
func (r *TeachingLoadRepositoryPG) List(ctx context.Context, filter usecases.TeachingLoadFilter) ([]*entities.TeachingLoad, error) {
	return nil, nil
}
