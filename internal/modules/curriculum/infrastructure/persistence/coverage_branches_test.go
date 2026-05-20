package persistence

// v0.153.7 Phase 6 backfill — closes uncovered branches across
// curriculum / section / discipline_item repository implementations:
// transport-error paths after QueryContext, mid-iteration scan errors,
// rows.Err propagation, RowsAffected inspection failures, the two
// disambiguate branches that fall through to non-NoRows errors, plus
// the previously-uncovered approvedBy/approvedAt populated branches
// in CurriculumRepositoryPG.List and the nil→Valid maps for
// nullableInt64Ptr / nullableTimePtr.
//
// All tests are sqlmock-driven и mirror the existing per-file
// conventions (regexp.QuoteMeta, WithArgs pinning per CLAUDE.md
// feedback_sqlmock_withargs_for_mutation_resistance.md). No
// production change.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// ===== CurriculumRepositoryPG.List branch coverage =====

func TestCurriculumRepositoryPG_List_QueryError(t *testing.T) {
	repo, mock := newCurriculumRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM curricula")).
		WithArgs("", sql.NullInt64{}, "", sql.NullInt64{}).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("ORDER BY year DESC, created_at DESC")).
		WithArgs("", sql.NullInt64{}, "", sql.NullInt64{}, 50, 0).
		WillReturnError(fmt.Errorf("conn refused"))

	_, err := repo.List(context.Background(), repositories.CurriculumListFilter{Limit: 50})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list")
}

func TestCurriculumRepositoryPG_List_ScanError(t *testing.T) {
	repo, mock := newCurriculumRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM curricula")).
		WithArgs("", sql.NullInt64{}, "", sql.NullInt64{}).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// Wrong column count triggers scan error inside loop.
	mock.ExpectQuery(regexp.QuoteMeta("ORDER BY year DESC, created_at DESC")).
		WithArgs("", sql.NullInt64{}, "", sql.NullInt64{}, 50, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	_, err := repo.List(context.Background(), repositories.CurriculumListFilter{Limit: 50})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list scan")
}

func TestCurriculumRepositoryPG_List_RowsErrPropagates(t *testing.T) {
	repo, mock := newCurriculumRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM curricula")).
		WithArgs("", sql.NullInt64{}, "", sql.NullInt64{}).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	rows := sqlmock.NewRows([]string{
		"id", "title", "code", "specialty", "year", "description",
		"status", "created_by", "approved_by", "approved_at",
		"created_at", "updated_at",
	}).
		AddRow(int64(1), "T", "C", "S", 2026, "", "draft", int64(42),
			sql.NullInt64{}, sql.NullTime{}, now, now).
		RowError(0, fmt.Errorf("connection reset during iteration"))
	mock.ExpectQuery(regexp.QuoteMeta("ORDER BY year DESC, created_at DESC")).
		WithArgs("", sql.NullInt64{}, "", sql.NullInt64{}, 50, 0).
		WillReturnRows(rows)

	_, err := repo.List(context.Background(), repositories.CurriculumListFilter{Limit: 50})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list iter")
}

func TestCurriculumRepositoryPG_List_ApprovedByAndAtPopulated(t *testing.T) {
	// Covers both `if approvedBy.Valid` and `if approvedAt.Valid` branches
	// inside the List-loop scan — previously only GetByID exercised them.
	repo, mock := newCurriculumRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	approvedAt := now.Add(48 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM curricula")).
		WithArgs("", sql.NullInt64{}, "", sql.NullInt64{}).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	rows := sqlmock.NewRows([]string{
		"id", "title", "code", "specialty", "year", "description",
		"status", "created_by", "approved_by", "approved_at",
		"created_at", "updated_at", "version",
	}).AddRow(int64(7), "T", "C", "S", 2026, "approved-desc",
		"approved", int64(42),
		sql.NullInt64{Int64: 99, Valid: true},
		sql.NullTime{Time: approvedAt, Valid: true},
		now, now, 0)
	mock.ExpectQuery(regexp.QuoteMeta("ORDER BY year DESC, created_at DESC")).
		WithArgs("", sql.NullInt64{}, "", sql.NullInt64{}, 50, 0).
		WillReturnRows(rows)

	got, err := repo.List(context.Background(), repositories.CurriculumListFilter{Limit: 50})
	require.NoError(t, err)
	require.Len(t, got.Items, 1)
	require.NotNil(t, got.Items[0].ApprovedBy())
	assert.Equal(t, int64(99), *got.Items[0].ApprovedBy())
	require.NotNil(t, got.Items[0].ApprovedAt())
	assert.True(t, got.Items[0].ApprovedAt().Equal(approvedAt))
}

// ===== CurriculumRepositoryPG.Update RowsAffected error =====

func TestCurriculumRepositoryPG_Update_RowsAffectedError(t *testing.T) {
	repo, mock := newCurriculumRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE curricula SET")).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

	c := freshDraft(t, "09.03.04-2026")
	c.ID = 7
	err := repo.Update(context.Background(), c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rows affected")
}

// ===== CurriculumRepositoryPG.AggregateByYearSpecialty branches =====

func TestCurriculumRepositoryPG_Aggregate_ScanError(t *testing.T) {
	repo, mock := newCurriculumRepoMock(t)

	mock.ExpectQuery(`SELECT specialty, status, COUNT\(\*\) FROM curricula`).
		WithArgs(2026).
		WillReturnRows(sqlmock.NewRows([]string{"specialty"}).AddRow("X"))

	got, err := repo.AggregateByYearSpecialty(context.Background(), 2026)
	require.Error(t, err)
	require.Nil(t, got)
	assert.Contains(t, err.Error(), "aggregate scan")
}

func TestCurriculumRepositoryPG_Aggregate_RowsErrPropagates(t *testing.T) {
	repo, mock := newCurriculumRepoMock(t)
	rows := sqlmock.NewRows([]string{"specialty", "status", "count"}).
		AddRow("Информатика", "approved", 1).
		RowError(0, fmt.Errorf("iter failure"))

	mock.ExpectQuery(`SELECT specialty, status, COUNT\(\*\) FROM curricula`).
		WithArgs(2026).
		WillReturnRows(rows)

	got, err := repo.AggregateByYearSpecialty(context.Background(), 2026)
	require.Error(t, err)
	require.Nil(t, got)
	assert.Contains(t, err.Error(), "aggregate rows")
}

// ===== Approved-curriculum Save exercises nullableInt64Ptr/TimePtr non-nil branch =====

func TestCurriculumRepositoryPG_Save_ApprovedExercisesNullableNonNilBranches(t *testing.T) {
	// freshDraft + SubmitForApproval + Approve populates approvedBy +
	// approvedAt — Save then routes those through nullableInt64Ptr /
	// nullableTimePtr's non-nil branches (line 307 / 314 — previously
	// 66.7% covered, only nil path tested).
	repo, mock := newCurriculumRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	c := freshDraft(t, "09.03.04-2026")
	require.NoError(t, c.SubmitForApproval(now))
	require.NoError(t, c.Approve(int64(77), now.Add(time.Hour)))
	require.NotNil(t, c.ApprovedBy())
	require.NotNil(t, c.ApprovedAt())

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO curricula")).
		WithArgs(
			"ИВТ-2026", "09.03.04-2026", "Информатика и вычислительная техника",
			2026, sql.NullString{String: "desc", Valid: true},
			"approved", int64(42),
			sql.NullInt64{Int64: 77, Valid: true},
			sql.NullTime{Time: now.Add(time.Hour), Valid: true},
			c.CreatedAt(), c.UpdatedAt(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(7)))

	require.NoError(t, repo.Save(context.Background(), c))
	assert.Equal(t, int64(7), c.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ===== SectionRepositoryPG branch coverage =====

func TestSectionRepoPG_ListByCurriculumID_QueryError(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM curriculum_sections WHERE curriculum_id = $1")).
		WithArgs(int64(7)).
		WillReturnError(fmt.Errorf("conn refused"))

	_, err := repo.ListByCurriculumID(context.Background(), 7)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list by curriculum id")
}

func TestSectionRepoPG_ListByCurriculumID_ScanError(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM curriculum_sections WHERE curriculum_id = $1")).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	_, err := repo.ListByCurriculumID(context.Background(), 7)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list scan")
}

func TestSectionRepoPG_ListByCurriculumID_RowsErrPropagates(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"id", "curriculum_id", "title", "description",
		"order_index", "version", "created_at", "updated_at",
	}).
		AddRow(int64(101), int64(7), "T", sql.NullString{}, 0, 0, now, now).
		RowError(0, fmt.Errorf("connection reset"))
	mock.ExpectQuery(regexp.QuoteMeta("FROM curriculum_sections WHERE curriculum_id = $1")).
		WithArgs(int64(7)).
		WillReturnRows(rows)

	_, err := repo.ListByCurriculumID(context.Background(), 7)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list iter")
}

func TestSectionRepoPG_Update_RowsAffectedError(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	s := entities.ReconstituteSection(101, 7, "T", "d", 0, 3, now, now)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE curriculum_sections SET")).
		WithArgs("T", sql.NullString{String: "d", Valid: true},
			0, s.UpdatedAt(), int64(101), 3).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

	err := repo.Update(context.Background(), s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rows affected")
}

func TestSectionRepoPG_DisambiguateAbsentUpdate_NonNoRowsError(t *testing.T) {
	// Covers fallthrough branch in disambiguateAbsentUpdate when the
	// SELECT 1 probe fails with a non-ErrNoRows error (e.g. transport)
	// — neither VersionConflict nor NotFound; wraps with "disambiguate".
	repo, mock := newSectionRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	s := entities.ReconstituteSection(101, 7, "T", "d", 0, 3, now, now)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE curriculum_sections SET")).
		WithArgs("T", sql.NullString{String: "d", Valid: true},
			0, s.UpdatedAt(), int64(101), 3).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM curriculum_sections WHERE id = $1")).
		WithArgs(int64(101)).
		WillReturnError(fmt.Errorf("conn refused"))

	err := repo.Update(context.Background(), s)
	require.Error(t, err)
	assert.False(t, errors.Is(err, repositories.ErrSectionNotFound))
	assert.False(t, errors.Is(err, repositories.ErrSectionVersionConflict))
	assert.Contains(t, err.Error(), "disambiguate")
}

func TestSectionRepoPG_Delete_RowsAffectedError(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM curriculum_sections WHERE id = $1")).
		WithArgs(int64(101)).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

	err := repo.Delete(context.Background(), 101)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rows affected")
}

// ===== DisciplineItemRepositoryPG branch coverage =====

func TestDisciplineItemRepoPG_ListBySectionID_QueryError(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM curriculum_section_items WHERE section_id = $1")).
		WithArgs(int64(7)).
		WillReturnError(fmt.Errorf("conn refused"))

	_, err := repo.ListBySectionID(context.Background(), 7)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list by section id")
}

func TestDisciplineItemRepoPG_ListBySectionID_ScanError(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM curriculum_section_items WHERE section_id = $1")).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	_, err := repo.ListBySectionID(context.Background(), 7)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list scan")
}

func TestDisciplineItemRepoPG_ListBySectionID_RowsErrPropagates(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"id", "section_id", "title",
		"hours_lectures", "hours_practice", "hours_lab", "hours_self",
		"control_form", "credits", "semester",
		"order_index", "version", "created_at", "updated_at",
	}).
		AddRow(int64(202), int64(7), "T", 18, 18, 0, 36,
			"zachet", 2, 1, 0, 0, now, now).
		RowError(0, fmt.Errorf("connection reset"))
	mock.ExpectQuery(regexp.QuoteMeta("FROM curriculum_section_items WHERE section_id = $1")).
		WithArgs(int64(7)).
		WillReturnRows(rows)

	_, err := repo.ListBySectionID(context.Background(), 7)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list iter")
}

func TestDisciplineItemRepoPG_Update_RowsAffectedError(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	d := entities.ReconstituteDisciplineItem(202, 7, "T",
		18, 18, 0, 36, entities.ControlFormZachet, 2, 1, 0, 3, now, now)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE curriculum_section_items SET")).
		WithArgs("T", 18, 18, 0, 36,
			"zachet", 2, 1, 0,
			d.UpdatedAt(), int64(202), 3).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

	err := repo.Update(context.Background(), d)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rows affected")
}

func TestDisciplineItemRepoPG_DisambiguateAbsentUpdate_NonNoRowsError(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	d := entities.ReconstituteDisciplineItem(202, 7, "T",
		18, 18, 0, 36, entities.ControlFormZachet, 2, 1, 0, 3, now, now)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE curriculum_section_items SET")).
		WithArgs("T", 18, 18, 0, 36,
			"zachet", 2, 1, 0,
			d.UpdatedAt(), int64(202), 3).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM curriculum_section_items WHERE id = $1")).
		WithArgs(int64(202)).
		WillReturnError(fmt.Errorf("conn refused"))

	err := repo.Update(context.Background(), d)
	require.Error(t, err)
	assert.False(t, errors.Is(err, repositories.ErrDisciplineItemNotFound))
	assert.False(t, errors.Is(err, repositories.ErrDisciplineItemVersionConflict))
	assert.Contains(t, err.Error(), "disambiguate")
}

func TestDisciplineItemRepoPG_Delete_RowsAffectedError(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM curriculum_section_items WHERE id = $1")).
		WithArgs(int64(202)).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

	err := repo.Delete(context.Background(), 202)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rows affected")
}

func TestDisciplineItemRepoPG_AggregateHoursByYear_ScanError(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	mock.ExpectQuery(`SELECT c.id, c.title.*FROM curricula c`).
		WithArgs(2026).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	got, err := repo.AggregateHoursByYear(context.Background(), 2026)
	require.Error(t, err)
	require.Nil(t, got)
	assert.Contains(t, err.Error(), "aggregate hours scan")
}

func TestDisciplineItemRepoPG_AggregateHoursByYear_RowsErrPropagates(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	rows := sqlmock.NewRows([]string{"id", "title", "lectures", "practice", "lab", "self_study"}).
		AddRow(int64(11), "T", 1, 1, 1, 1).
		RowError(0, fmt.Errorf("iter failure"))
	mock.ExpectQuery(`SELECT c.id, c.title.*FROM curricula c`).
		WithArgs(2026).
		WillReturnRows(rows)

	got, err := repo.AggregateHoursByYear(context.Background(), 2026)
	require.Error(t, err)
	require.Nil(t, got)
	assert.Contains(t, err.Error(), "aggregate hours rows")
}
