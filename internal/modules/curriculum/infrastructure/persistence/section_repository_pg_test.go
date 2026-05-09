package persistence

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

func newSectionRepoMock(t *testing.T) (*SectionRepositoryPG, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewSectionRepositoryPG(db), mock
}

func freshSectionForRepo(t *testing.T) *entities.Section {
	t.Helper()
	s, err := entities.NewSection(entities.NewSectionParams{
		CurriculumID: 7,
		Title:        "Базовая часть",
		Description:  "desc",
		OrderIndex:   0,
		Now:          time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	return s
}

// ===== Save =====

func TestSectionRepoPG_Save_HappyPath(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	s := freshSectionForRepo(t)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO curriculum_sections")).
		WithArgs(int64(7), "Базовая часть", sql.NullString{String: "desc", Valid: true},
			0, s.CreatedAt(), s.UpdatedAt()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(101)))

	err := repo.Save(context.Background(), s)
	require.NoError(t, err)
	assert.Equal(t, int64(101), s.ID, "Save must populate generated id back onto entity")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSectionRepoPG_Save_NullDescription(t *testing.T) {
	// Empty description maps to SQL NULL (mirror к nullableDescription
	// in curriculum repo) — keeps the column nullable distinction
	// honest, no needless empty-string rows.
	repo, mock := newSectionRepoMock(t)
	s, err := entities.NewSection(entities.NewSectionParams{
		CurriculumID: 7,
		Title:        "Без описания",
		Description:  "",
		OrderIndex:   1,
		Now:          time.Now(),
	})
	require.NoError(t, err)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO curriculum_sections")).
		WithArgs(int64(7), "Без описания", sql.NullString{}, 1, s.CreatedAt(), s.UpdatedAt()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(102)))

	err = repo.Save(context.Background(), s)
	require.NoError(t, err)
	assert.Equal(t, int64(102), s.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSectionRepoPG_Save_TransportError(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	s := freshSectionForRepo(t)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO curriculum_sections")).
		WillReturnError(fmt.Errorf("conn refused"))

	err := repo.Save(context.Background(), s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save")
}

// ===== GetByID =====

func TestSectionRepoPG_GetByID_RowFound(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"id", "curriculum_id", "title", "description",
		"order_index", "version", "created_at", "updated_at",
	}).AddRow(int64(101), int64(7), "Базовая часть",
		sql.NullString{String: "desc", Valid: true}, 2, 5, now, now)

	mock.ExpectQuery(regexp.QuoteMeta("FROM curriculum_sections WHERE id = $1")).
		WithArgs(int64(101)).
		WillReturnRows(rows)

	got, err := repo.GetByID(context.Background(), 101)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, int64(101), got.ID)
	assert.Equal(t, int64(7), got.CurriculumID())
	assert.Equal(t, "Базовая часть", got.Title())
	assert.Equal(t, "desc", got.Description())
	assert.Equal(t, 2, got.OrderIndex())
	assert.Equal(t, 5, got.Version())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSectionRepoPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM curriculum_sections WHERE id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	got, err := repo.GetByID(context.Background(), 999)
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, repositories.ErrSectionNotFound),
		"err must wrap ErrSectionNotFound, got %v", err)
}

func TestSectionRepoPG_GetByID_TransportError(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM curriculum_sections WHERE id = $1")).
		WithArgs(int64(10)).
		WillReturnError(fmt.Errorf("conn refused"))

	_, err := repo.GetByID(context.Background(), 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get by id")
}

func TestSectionRepoPG_GetByID_NullDescriptionScansAsEmpty(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "curriculum_id", "title", "description",
		"order_index", "version", "created_at", "updated_at",
	}).AddRow(int64(50), int64(7), "Без описания", sql.NullString{},
		0, 0, now, now)

	mock.ExpectQuery(regexp.QuoteMeta("FROM curriculum_sections WHERE id = $1")).
		WithArgs(int64(50)).
		WillReturnRows(rows)

	got, err := repo.GetByID(context.Background(), 50)
	require.NoError(t, err)
	assert.Equal(t, "", got.Description())
}

// ===== ListByCurriculumID =====

func TestSectionRepoPG_ListByCurriculumID_OrderedResults(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	now := time.Now()
	earlier := now.Add(-1 * time.Hour)

	rows := sqlmock.NewRows([]string{
		"id", "curriculum_id", "title", "description",
		"order_index", "version", "created_at", "updated_at",
	}).
		AddRow(int64(101), int64(7), "Базовая", sql.NullString{String: "a", Valid: true}, 0, 0, earlier, earlier).
		AddRow(int64(102), int64(7), "Вариативная", sql.NullString{}, 1, 0, now, now)

	mock.ExpectQuery(regexp.QuoteMeta("FROM curriculum_sections WHERE curriculum_id = $1")).
		WithArgs(int64(7)).
		WillReturnRows(rows)

	got, err := repo.ListByCurriculumID(context.Background(), 7)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, int64(101), got[0].ID)
	assert.Equal(t, int64(102), got[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSectionRepoPG_ListByCurriculumID_EmptyResult(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	rows := sqlmock.NewRows([]string{
		"id", "curriculum_id", "title", "description",
		"order_index", "version", "created_at", "updated_at",
	})
	mock.ExpectQuery(regexp.QuoteMeta("FROM curriculum_sections WHERE curriculum_id = $1")).
		WithArgs(int64(99)).
		WillReturnRows(rows)

	got, err := repo.ListByCurriculumID(context.Background(), 99)
	require.NoError(t, err)
	assert.Len(t, got, 0)
}

// ===== Update =====
//
// CLAUDE.md ≥3-variant gate: covers the 3 distinct outcomes of
// optimistic-locking Update — RowsAffected==1 (success + version bump),
// RowsAffected==0 + row exists (version conflict), RowsAffected==0 +
// row missing (deleted between load and write).

func TestSectionRepoPG_Update_HappyPath(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	s := entities.ReconstituteSection(101, 7, "Старый", "old desc", 0, 3,
		time.Now().Add(-time.Hour), time.Now().Add(-time.Hour))
	require.NoError(t, s.UpdateBasics("Новый", "new desc", 1, time.Now()))
	priorVersion := s.Version()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE curriculum_sections SET")).
		WithArgs("Новый", sql.NullString{String: "new desc", Valid: true},
			1, s.UpdatedAt(), int64(101), priorVersion).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Update(context.Background(), s))
	assert.Equal(t, priorVersion+1, s.Version(),
		"Update must bump entity version on RowsAffected==1")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSectionRepoPG_Update_VersionConflict(t *testing.T) {
	// RowsAffected==0 + follow-up SELECT finds the row → version stale.
	// Distinct from NotFound (row deleted entirely).
	repo, mock := newSectionRepoMock(t)
	s := entities.ReconstituteSection(101, 7, "T", "d", 0, 3, time.Now(), time.Now())

	mock.ExpectExec(regexp.QuoteMeta("UPDATE curriculum_sections SET")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM curriculum_sections WHERE id = $1")).
		WithArgs(int64(101)).
		WillReturnRows(sqlmock.NewRows([]string{"?column?"}).AddRow(1))

	err := repo.Update(context.Background(), s)
	assert.True(t, errors.Is(err, repositories.ErrSectionVersionConflict),
		"err must wrap ErrSectionVersionConflict, got %v", err)
	assert.Equal(t, 3, s.Version(), "Update must NOT bump version on conflict")
}

func TestSectionRepoPG_Update_NotFound(t *testing.T) {
	// RowsAffected==0 + follow-up SELECT finds no row → entity gone.
	repo, mock := newSectionRepoMock(t)
	s := entities.ReconstituteSection(999, 7, "T", "d", 0, 0, time.Now(), time.Now())

	mock.ExpectExec(regexp.QuoteMeta("UPDATE curriculum_sections SET")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM curriculum_sections WHERE id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	err := repo.Update(context.Background(), s)
	assert.True(t, errors.Is(err, repositories.ErrSectionNotFound),
		"err must wrap ErrSectionNotFound, got %v", err)
}

func TestSectionRepoPG_Update_TransportError(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	s := entities.ReconstituteSection(101, 7, "T", "d", 0, 0, time.Now(), time.Now())
	mock.ExpectExec(regexp.QuoteMeta("UPDATE curriculum_sections SET")).
		WillReturnError(fmt.Errorf("conn refused"))

	err := repo.Update(context.Background(), s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update")
}

// ===== Delete =====

func TestSectionRepoPG_Delete_HappyPath(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM curriculum_sections WHERE id = $1")).
		WithArgs(int64(101)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Delete(context.Background(), 101))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSectionRepoPG_Delete_NotFound(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM curriculum_sections WHERE id = $1")).
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 999)
	assert.True(t, errors.Is(err, repositories.ErrSectionNotFound),
		"err must wrap ErrSectionNotFound, got %v", err)
}

func TestSectionRepoPG_Delete_TransportError(t *testing.T) {
	repo, mock := newSectionRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM curriculum_sections WHERE id = $1")).
		WillReturnError(fmt.Errorf("conn refused"))

	err := repo.Delete(context.Background(), 101)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete")
}
