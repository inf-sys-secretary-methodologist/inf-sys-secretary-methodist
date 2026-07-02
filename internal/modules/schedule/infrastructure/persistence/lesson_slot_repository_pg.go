package persistence

import (
	"context"
	"database/sql"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// LessonSlotRepositoryPG implements LessonSlotRepository on PostgreSQL.
type LessonSlotRepositoryPG struct {
	db *sql.DB
}

var _ usecases.LessonSlotRepository = (*LessonSlotRepositoryPG)(nil)

// NewLessonSlotRepositoryPG creates a new LessonSlotRepositoryPG.
func NewLessonSlotRepositoryPG(db *sql.DB) *LessonSlotRepositoryPG {
	return &LessonSlotRepositoryPG{db: db}
}

// Create inserts a new slot. STUB — see GREEN commit.
func (r *LessonSlotRepositoryPG) Create(ctx context.Context, slot *entities.LessonSlot) error {
	return nil
}

// Update mutates an existing slot. STUB — see GREEN commit.
func (r *LessonSlotRepositoryPG) Update(ctx context.Context, slot *entities.LessonSlot) error {
	return nil
}

// Delete removes a slot by id. STUB — see GREEN commit.
func (r *LessonSlotRepositoryPG) Delete(ctx context.Context, id int64) error {
	return nil
}

// GetByID returns one slot. STUB — see GREEN commit.
func (r *LessonSlotRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.LessonSlot, error) {
	return nil, nil
}

// List returns all slots ordered by Number. STUB — see GREEN commit.
func (r *LessonSlotRepositoryPG) List(ctx context.Context) ([]*entities.LessonSlot, error) {
	return nil, nil
}
