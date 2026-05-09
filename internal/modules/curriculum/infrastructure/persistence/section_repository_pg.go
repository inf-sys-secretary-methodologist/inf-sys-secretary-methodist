package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
)

// SectionRepositoryPG is the SQL implementation of SectionRepository
// (curriculum_sections table, migration 034).
type SectionRepositoryPG struct {
	db *sql.DB
}

// NewSectionRepositoryPG constructs the repository.
func NewSectionRepositoryPG(db *sql.DB) *SectionRepositoryPG {
	return &SectionRepositoryPG{db: db}
}

// Save — implementation lands в GREEN commit (Pair 3).
func (r *SectionRepositoryPG) Save(ctx context.Context, s *entities.Section) error {
	_, _ = ctx, s
	return errors.New("section: Save not implemented yet")
}

// GetByID — implementation lands в GREEN commit (Pair 3).
func (r *SectionRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.Section, error) {
	_, _ = ctx, id
	return nil, errors.New("section: GetByID not implemented yet")
}

// ListByCurriculumID — implementation lands в GREEN commit (Pair 3).
func (r *SectionRepositoryPG) ListByCurriculumID(ctx context.Context, curriculumID int64) ([]*entities.Section, error) {
	_, _ = ctx, curriculumID
	return nil, errors.New("section: ListByCurriculumID not implemented yet")
}

// Update — implementation lands в GREEN commit (Pair 3).
func (r *SectionRepositoryPG) Update(ctx context.Context, s *entities.Section) error {
	_, _ = ctx, s
	return errors.New("section: Update not implemented yet")
}

// Delete — implementation lands в GREEN commit (Pair 3).
func (r *SectionRepositoryPG) Delete(ctx context.Context, id int64) error {
	_, _ = ctx, id
	return errors.New("section: Delete not implemented yet")
}
