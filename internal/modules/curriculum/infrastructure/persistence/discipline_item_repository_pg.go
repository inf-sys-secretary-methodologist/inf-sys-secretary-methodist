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

// DisciplineItemRepositoryPG is the SQL implementation of
// DisciplineItemRepository (curriculum_section_items table, migration 035).
// Optimistic locking per ADR-3 — Update uses WHERE id = ? AND version = ?
// + atomic version increment + disambiguates RowsAffected == 0 via
// follow-up existence SELECT.
//
// Accepts DBTX (not *sql.DB) so the same struct can run against single-
// connection mode или a `*sql.Tx` inside BulkDisciplineItemsUnitOfWork
// (v0.128.3 ADR-10).
type DisciplineItemRepositoryPG struct {
	db DBTX
}

// NewDisciplineItemRepositoryPG constructs the repository. db can be
// `*sql.DB` (default DI) или `*sql.Tx` (bulk-edit transactional path).
func NewDisciplineItemRepositoryPG(db DBTX) *DisciplineItemRepositoryPG {
	return &DisciplineItemRepositoryPG{db: db}
}

const disciplineItemSelectColumns = `id, section_id, title, hours_lectures, hours_practice, hours_lab, hours_self, control_form, credits, semester, order_index, version, created_at, updated_at`

// Save inserts a new DisciplineItem row and writes the generated id
// back onto the entity.
func (r *DisciplineItemRepositoryPG) Save(ctx context.Context, d *entities.DisciplineItem) error {
	const query = `
		INSERT INTO curriculum_section_items (
			section_id, title,
			hours_lectures, hours_practice, hours_lab, hours_self,
			control_form, credits, semester, order_index,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`

	var newID int64
	err := r.db.QueryRowContext(ctx, query,
		d.SectionID(), d.Title(),
		d.HoursLectures(), d.HoursPractice(), d.HoursLab(), d.HoursSelf(),
		string(d.ControlForm()), d.Credits(), d.Semester(), d.OrderIndex(),
		d.CreatedAt(), d.UpdatedAt(),
	).Scan(&newID)
	if err != nil {
		return fmt.Errorf("discipline_item: save: %w", err)
	}
	d.ID = newID
	return nil
}

// GetByID returns the DisciplineItem with the given id or
// ErrDisciplineItemNotFound when the row is missing.
func (r *DisciplineItemRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.DisciplineItem, error) {
	const query = `SELECT ` + disciplineItemSelectColumns + ` FROM curriculum_section_items WHERE id = $1`
	var (
		idv, sectionID                                    int64
		title, controlForm                                string
		hoursLectures, hoursPractice, hoursLab, hoursSelf int
		credits, semester, orderIndex, version            int
		createdAt, updatedAt                              time.Time
	)
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&idv, &sectionID, &title,
		&hoursLectures, &hoursPractice, &hoursLab, &hoursSelf,
		&controlForm, &credits, &semester, &orderIndex, &version,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrDisciplineItemNotFound
		}
		return nil, fmt.Errorf("discipline_item: get by id: %w", err)
	}
	return entities.ReconstituteDisciplineItem(
		idv, sectionID, title,
		hoursLectures, hoursPractice, hoursLab, hoursSelf,
		entities.ControlForm(controlForm), credits, semester, orderIndex, version,
		createdAt, updatedAt,
	), nil
}

// ListBySectionID returns every DisciplineItem attached to the given
// section, ordered by (order_index ASC, created_at ASC, id ASC).
func (r *DisciplineItemRepositoryPG) ListBySectionID(ctx context.Context, sectionID int64) ([]*entities.DisciplineItem, error) {
	const query = `SELECT ` + disciplineItemSelectColumns + `
		FROM curriculum_section_items
		WHERE section_id = $1
		ORDER BY order_index ASC, created_at ASC, id ASC`

	rows, err := r.db.QueryContext(ctx, query, sectionID)
	if err != nil {
		return nil, fmt.Errorf("discipline_item: list by section id: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []*entities.DisciplineItem
	for rows.Next() {
		var (
			idv, secID                                        int64
			title, controlForm                                string
			hoursLectures, hoursPractice, hoursLab, hoursSelf int
			credits, semester, orderIndex, version            int
			createdAt, updatedAt                              time.Time
		)
		if err := rows.Scan(&idv, &secID, &title,
			&hoursLectures, &hoursPractice, &hoursLab, &hoursSelf,
			&controlForm, &credits, &semester, &orderIndex, &version,
			&createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("discipline_item: list scan: %w", err)
		}
		items = append(items, entities.ReconstituteDisciplineItem(
			idv, secID, title,
			hoursLectures, hoursPractice, hoursLab, hoursSelf,
			entities.ControlForm(controlForm), credits, semester, orderIndex, version,
			createdAt, updatedAt,
		))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("discipline_item: list iter: %w", err)
	}
	return items, nil
}

// Update writes the (already-mutated) entity back, enforcing optimistic
// locking. Same semantics as Section.Update — see disambiguateAbsentDisciplineItemUpdate.
func (r *DisciplineItemRepositoryPG) Update(ctx context.Context, d *entities.DisciplineItem) error {
	const query = `
		UPDATE curriculum_section_items SET
			title          = $1,
			hours_lectures = $2,
			hours_practice = $3,
			hours_lab      = $4,
			hours_self     = $5,
			control_form   = $6,
			credits        = $7,
			semester       = $8,
			order_index    = $9,
			version        = version + 1,
			updated_at     = $10
		WHERE id = $11 AND version = $12`

	res, err := r.db.ExecContext(ctx, query,
		d.Title(),
		d.HoursLectures(), d.HoursPractice(), d.HoursLab(), d.HoursSelf(),
		string(d.ControlForm()), d.Credits(), d.Semester(), d.OrderIndex(),
		d.UpdatedAt(),
		d.ID, d.Version(),
	)
	if err != nil {
		return fmt.Errorf("discipline_item: update: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("discipline_item: update: rows affected: %w", err)
	}
	if rows == 0 {
		return r.disambiguateAbsentDisciplineItemUpdate(ctx, d.ID)
	}
	bumpDisciplineItemVersion(d)
	return nil
}

// disambiguateAbsentDisciplineItemUpdate runs follow-up existence query
// when Update's RowsAffected was 0 — distinguishes stale-version race
// from row-deleted-entirely.
//
// TOCTOU note: window между failed UPDATE и this SELECT during which
// parallel transaction could DELETE — version conflict race may be
// reported here as ErrDisciplineItemNotFound. Same caveat as Section
// disambiguateAbsentUpdate; B1b bulk-edit (v0.128.2) must serialize
// reads inside transaction to close window.
func (r *DisciplineItemRepositoryPG) disambiguateAbsentDisciplineItemUpdate(ctx context.Context, id int64) error {
	const probe = `SELECT 1 FROM curriculum_section_items WHERE id = $1`
	var found int
	err := r.db.QueryRowContext(ctx, probe, id).Scan(&found)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repositories.ErrDisciplineItemNotFound
		}
		return fmt.Errorf("discipline_item: update: disambiguate: %w", err)
	}
	return repositories.ErrDisciplineItemVersionConflict
}

// Delete removes the row by id. Returns ErrDisciplineItemNotFound if
// no row was deleted.
func (r *DisciplineItemRepositoryPG) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM curriculum_section_items WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("discipline_item: delete: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("discipline_item: delete: rows affected: %w", err)
	}
	if rows == 0 {
		return repositories.ErrDisciplineItemNotFound
	}
	return nil
}

// AggregateHoursByYear sums the four hours columns per curriculum
// for curricula with curricula.year = year. LEFT JOIN keeps curricula
// that have no sections or items (they contribute 0 to each total).
func (r *DisciplineItemRepositoryPG) AggregateHoursByYear(ctx context.Context, year int) ([]repositories.DisciplineItemHoursAgg, error) {
	const query = `SELECT c.id, c.title,
		COALESCE(SUM(ci.hours_lectures), 0),
		COALESCE(SUM(ci.hours_practice), 0),
		COALESCE(SUM(ci.hours_lab), 0),
		COALESCE(SUM(ci.hours_self), 0)
		FROM curricula c
		LEFT JOIN curriculum_sections s ON s.curriculum_id = c.id
		LEFT JOIN curriculum_section_items ci ON ci.section_id = s.id
		WHERE c.year = $1
		GROUP BY c.id, c.title
		ORDER BY c.title, c.id`

	rows, err := r.db.QueryContext(ctx, query, year)
	if err != nil {
		return nil, fmt.Errorf("discipline_item: aggregate hours by year: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []repositories.DisciplineItemHoursAgg
	for rows.Next() {
		var (
			id                               int64
			title                            string
			lectures, practice, lab, selfHrs int
		)
		if err := rows.Scan(&id, &title, &lectures, &practice, &lab, &selfHrs); err != nil {
			return nil, fmt.Errorf("discipline_item: aggregate hours scan: %w", err)
		}
		out = append(out, repositories.DisciplineItemHoursAgg{
			CurriculumID:    id,
			CurriculumTitle: title,
			Lectures:        lectures,
			Practice:        practice,
			Lab:             lab,
			SelfStudy:       selfHrs,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("discipline_item: aggregate hours rows: %w", err)
	}
	return out, nil
}

// bumpDisciplineItemVersion increments the version on the entity to
// reflect the SQL UPDATE's `version = version + 1`. Same encapsulation
// pattern as bumpSectionVersion (Reconstitute-based re-build).
func bumpDisciplineItemVersion(d *entities.DisciplineItem) {
	*d = *entities.ReconstituteDisciplineItem(
		d.ID, d.SectionID(), d.Title(),
		d.HoursLectures(), d.HoursPractice(), d.HoursLab(), d.HoursSelf(),
		d.ControlForm(), d.Credits(), d.Semester(), d.OrderIndex(),
		d.Version()+1,
		d.CreatedAt(), d.UpdatedAt(),
	)
}
