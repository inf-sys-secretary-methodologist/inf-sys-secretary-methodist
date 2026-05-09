package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
)

// DisciplineItemRepositoryPG is the SQL implementation of
// DisciplineItemRepository (curriculum_section_items table, migration 035).
// Optimistic locking per ADR-3 — Update uses WHERE id = ? AND version = ?
// + atomic version increment + disambiguates RowsAffected == 0 via
// follow-up existence SELECT.
type DisciplineItemRepositoryPG struct {
	db *sql.DB
}

// NewDisciplineItemRepositoryPG constructs the repository.
func NewDisciplineItemRepositoryPG(db *sql.DB) *DisciplineItemRepositoryPG {
	return &DisciplineItemRepositoryPG{db: db}
}

// Save — implementation lands в GREEN commit (Pair 3).
func (r *DisciplineItemRepositoryPG) Save(ctx context.Context, d *entities.DisciplineItem) error {
	_, _ = ctx, d
	return errors.New("discipline_item: Save not implemented yet")
}

// GetByID — implementation lands в GREEN commit (Pair 3).
func (r *DisciplineItemRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.DisciplineItem, error) {
	_, _ = ctx, id
	return nil, errors.New("discipline_item: GetByID not implemented yet")
}

// ListBySectionID — implementation lands в GREEN commit (Pair 3).
func (r *DisciplineItemRepositoryPG) ListBySectionID(ctx context.Context, sectionID int64) ([]*entities.DisciplineItem, error) {
	_, _ = ctx, sectionID
	return nil, errors.New("discipline_item: ListBySectionID not implemented yet")
}

// Update — implementation lands в GREEN commit (Pair 3).
func (r *DisciplineItemRepositoryPG) Update(ctx context.Context, d *entities.DisciplineItem) error {
	_, _ = ctx, d
	return errors.New("discipline_item: Update not implemented yet")
}

// Delete — implementation lands в GREEN commit (Pair 3).
func (r *DisciplineItemRepositoryPG) Delete(ctx context.Context, id int64) error {
	_, _ = ctx, id
	return errors.New("discipline_item: Delete not implemented yet")
}
