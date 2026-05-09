package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// SectionRepositoryPG is the SQL implementation of SectionRepository
// (curriculum_sections table, migration 034). Optimistic locking per
// ADR-3 — Update uses WHERE id = ? AND version = ? and disambiguates
// RowsAffected == 0 via a follow-up existence SELECT.
type SectionRepositoryPG struct {
	db *sql.DB
}

// NewSectionRepositoryPG constructs the repository.
func NewSectionRepositoryPG(db *sql.DB) *SectionRepositoryPG {
	return &SectionRepositoryPG{db: db}
}

const sectionSelectColumns = `id, curriculum_id, title, description, order_index, version, created_at, updated_at`

// Save inserts a new Section row and writes the generated id back onto
// the entity. Sections have no unique-natural-key constraint (mirror
// of methodist working state where the same title may legitimately
// repeat across curricula), so transport errors propagate as wrapped
// errors without a sentinel mapping.
func (r *SectionRepositoryPG) Save(ctx context.Context, s *entities.Section) error {
	const query = `
		INSERT INTO curriculum_sections (
			curriculum_id, title, description, order_index,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	var newID int64
	err := r.db.QueryRowContext(ctx, query,
		s.CurriculumID(), s.Title(), nullableDescription(s.Description()),
		s.OrderIndex(), s.CreatedAt(), s.UpdatedAt(),
	).Scan(&newID)
	if err != nil {
		return fmt.Errorf("section: save: %w", err)
	}
	s.ID = newID
	return nil
}

// GetByID returns the section with the given id or
// ErrSectionNotFound when the row is missing.
func (r *SectionRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.Section, error) {
	const query = `SELECT ` + sectionSelectColumns + ` FROM curriculum_sections WHERE id = $1`
	var (
		idv          int64
		curriculumID int64
		title        string
		description  sql.NullString
		orderIndex   int
		version      int
		createdAt    time.Time
		updatedAt    time.Time
	)
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&idv, &curriculumID, &title, &description,
		&orderIndex, &version, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrSectionNotFound
		}
		return nil, fmt.Errorf("section: get by id: %w", err)
	}
	return entities.ReconstituteSection(
		idv, curriculumID, title, description.String,
		orderIndex, version, createdAt, updatedAt,
	), nil
}

// ListByCurriculumID returns every Section attached to the given
// curriculum, ordered by (order_index ASC, created_at ASC, id ASC).
// An empty slice (not nil error) is returned for curricula with no
// sections.
func (r *SectionRepositoryPG) ListByCurriculumID(ctx context.Context, curriculumID int64) ([]*entities.Section, error) {
	const query = `SELECT ` + sectionSelectColumns + `
		FROM curriculum_sections
		WHERE curriculum_id = $1
		ORDER BY order_index ASC, created_at ASC, id ASC`

	rows, err := r.db.QueryContext(ctx, query, curriculumID)
	if err != nil {
		return nil, fmt.Errorf("section: list by curriculum id: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []*entities.Section
	for rows.Next() {
		var (
			idv         int64
			curID       int64
			title       string
			description sql.NullString
			orderIndex  int
			version     int
			createdAt   time.Time
			updatedAt   time.Time
		)
		if err := rows.Scan(&idv, &curID, &title, &description,
			&orderIndex, &version, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("section: list scan: %w", err)
		}
		items = append(items, entities.ReconstituteSection(
			idv, curID, title, description.String,
			orderIndex, version, createdAt, updatedAt,
		))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("section: list iter: %w", err)
	}
	return items, nil
}

// Update writes the (already-mutated) entity back, enforcing
// optimistic locking via WHERE id = ? AND version = ?. On
// RowsAffected == 0 the impl runs a follow-up SELECT 1 to
// distinguish:
//   - row exists (different version)  → ErrSectionVersionConflict
//   - row missing entirely             → ErrSectionNotFound
//
// On success the entity's version field is bumped by 1 to reflect
// the new row state, so callers see a consistent post-update view
// without a separate reload.
func (r *SectionRepositoryPG) Update(ctx context.Context, s *entities.Section) error {
	const query = `
		UPDATE curriculum_sections SET
			title       = $1,
			description = $2,
			order_index = $3,
			version     = version + 1,
			updated_at  = $4
		WHERE id = $5 AND version = $6`

	res, err := r.db.ExecContext(ctx, query,
		s.Title(), nullableDescription(s.Description()),
		s.OrderIndex(), s.UpdatedAt(),
		s.ID, s.Version(),
	)
	if err != nil {
		return fmt.Errorf("section: update: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("section: update: rows affected: %w", err)
	}
	if rows == 0 {
		return r.disambiguateAbsentUpdate(ctx, s.ID)
	}
	bumpSectionVersion(s)
	return nil
}

// disambiguateAbsentUpdate runs a follow-up existence query when
// Update's RowsAffected was 0 — distinguishes a stale-version race
// from a section that vanished entirely. Sets a clear caller-facing
// sentinel either way.
//
// TOCTOU note: there is a window between the failed UPDATE and this
// SELECT during which a parallel transaction could DELETE the row.
// In that case a "version conflict" race is reported here as
// ErrSectionNotFound. Acceptable for admin-internal CRUD (UI shows
// "section was deleted, refresh"); future bulk-edit (B1b, v0.128.2)
// must serialize reads inside the transaction to close the window.
func (r *SectionRepositoryPG) disambiguateAbsentUpdate(ctx context.Context, id int64) error {
	const probe = `SELECT 1 FROM curriculum_sections WHERE id = $1`
	var found int
	err := r.db.QueryRowContext(ctx, probe, id).Scan(&found)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repositories.ErrSectionNotFound
		}
		return fmt.Errorf("section: update: disambiguate: %w", err)
	}
	return repositories.ErrSectionVersionConflict
}

// Delete removes the row by id. Returns ErrSectionNotFound if no row
// was deleted (idempotency check). CASCADE on curriculum_id (migration
// 034) handles the curriculum-deleted case automatically; this method
// is the explicit per-section delete path used by the DeleteSection
// use case (Pair 4).
func (r *SectionRepositoryPG) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM curriculum_sections WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("section: delete: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("section: delete: rows affected: %w", err)
	}
	if rows == 0 {
		return repositories.ErrSectionNotFound
	}
	return nil
}

// bumpSectionVersion increments the version on the entity to reflect
// the SQL UPDATE's `version = version + 1`. Wrapped in a helper so
// the entity's private version field stays encapsulated — this file
// is in the same module so it can mutate the unexported field.
func bumpSectionVersion(s *entities.Section) {
	// Re-build through Reconstitute to keep the entity opaque from
	// the persistence layer's perspective. Cheaper alternative would
	// be a SetVersion(int) method on the entity, but Reconstitute
	// already exists and serves the "trust the storage" purpose
	// expected here.
	*s = *entities.ReconstituteSection(
		s.ID, s.CurriculumID(), s.Title(), s.Description(),
		s.OrderIndex(), s.Version()+1,
		s.CreatedAt(), s.UpdatedAt(),
	)
}
