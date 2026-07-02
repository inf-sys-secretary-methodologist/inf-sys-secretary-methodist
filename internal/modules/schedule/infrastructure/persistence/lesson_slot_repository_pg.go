package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

const pqUniqueViolation = "23505"

// isUniqueViolation reports whether err is a PostgreSQL unique-constraint error.
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == pqUniqueViolation
}

// LessonSlotRepositoryPG implements LessonSlotRepository on PostgreSQL.
type LessonSlotRepositoryPG struct {
	db *sql.DB
}

var _ usecases.LessonSlotRepository = (*LessonSlotRepositoryPG)(nil)

// NewLessonSlotRepositoryPG creates a new LessonSlotRepositoryPG.
func NewLessonSlotRepositoryPG(db *sql.DB) *LessonSlotRepositoryPG {
	return &LessonSlotRepositoryPG{db: db}
}

// Create inserts a new slot, translating a duplicate Number (unique violation)
// into entities.ErrLessonSlotNumberTaken.
func (r *LessonSlotRepositoryPG) Create(ctx context.Context, slot *entities.LessonSlot) error {
	const query = `
		INSERT INTO lesson_slots (number, time_start, time_end, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query, slot.Number, slot.TimeStart, slot.TimeEnd, slot.CreatedAt, slot.UpdatedAt).Scan(&slot.ID)
	if isUniqueViolation(err) {
		return entities.ErrLessonSlotNumberTaken
	}
	if err != nil {
		return fmt.Errorf("failed to create lesson slot: %w", err)
	}
	return nil
}

// Update mutates an existing slot. A missing row returns ErrLessonSlotNotFound;
// a duplicate Number returns ErrLessonSlotNumberTaken.
func (r *LessonSlotRepositoryPG) Update(ctx context.Context, slot *entities.LessonSlot) error {
	const query = `
		UPDATE lesson_slots
		SET number = $1, time_start = $2, time_end = $3, updated_at = $4
		WHERE id = $5`

	res, err := r.db.ExecContext(ctx, query, slot.Number, slot.TimeStart, slot.TimeEnd, slot.UpdatedAt, slot.ID)
	if isUniqueViolation(err) {
		return entities.ErrLessonSlotNumberTaken
	}
	if err != nil {
		return fmt.Errorf("failed to update lesson slot: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read update result: %w", err)
	}
	if affected == 0 {
		return entities.ErrLessonSlotNotFound
	}
	return nil
}

// Delete removes a slot by id, returning ErrLessonSlotNotFound if none matched.
func (r *LessonSlotRepositoryPG) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM lesson_slots WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete lesson slot: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read delete result: %w", err)
	}
	if affected == 0 {
		return entities.ErrLessonSlotNotFound
	}
	return nil
}

// GetByID returns one slot or ErrLessonSlotNotFound.
func (r *LessonSlotRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.LessonSlot, error) {
	const query = `SELECT id, number, time_start, time_end, created_at, updated_at FROM lesson_slots WHERE id = $1`

	slot := &entities.LessonSlot{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&slot.ID, &slot.Number, &slot.TimeStart, &slot.TimeEnd, &slot.CreatedAt, &slot.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, entities.ErrLessonSlotNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lesson slot: %w", err)
	}
	return slot, nil
}

// List returns all slots ordered by Number.
func (r *LessonSlotRepositoryPG) List(ctx context.Context) ([]*entities.LessonSlot, error) {
	const query = `SELECT id, number, time_start, time_end, created_at, updated_at FROM lesson_slots ORDER BY number`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list lesson slots: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var slots []*entities.LessonSlot
	for rows.Next() {
		slot := &entities.LessonSlot{}
		if err := rows.Scan(&slot.ID, &slot.Number, &slot.TimeStart, &slot.TimeEnd, &slot.CreatedAt, &slot.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan lesson slot: %w", err)
		}
		slots = append(slots, slot)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate lesson slots: %w", err)
	}
	return slots, nil
}
