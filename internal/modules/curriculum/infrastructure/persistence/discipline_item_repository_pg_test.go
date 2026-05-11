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

func newDisciplineItemRepoMock(t *testing.T) (*DisciplineItemRepositoryPG, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewDisciplineItemRepositoryPG(db), mock
}

func freshItemForRepo(t *testing.T) *entities.DisciplineItem {
	t.Helper()
	d, err := entities.NewDisciplineItem(entities.NewDisciplineItemParams{
		SectionID:     7,
		Title:         "Математический анализ",
		HoursLectures: 36,
		HoursPractice: 36,
		HoursLab:      0,
		HoursSelf:     72,
		ControlForm:   entities.ControlFormExam,
		Credits:       4,
		Semester:      1,
		OrderIndex:    0,
		Now:           time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	return d
}

// ===== Save =====

func TestDisciplineItemRepoPG_Save_HappyPath(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	d := freshItemForRepo(t)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO curriculum_section_items")).
		WithArgs(int64(7), "Математический анализ",
			36, 36, 0, 72,
			"exam", 4, 1, 0,
			d.CreatedAt(), d.UpdatedAt()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(202)))

	err := repo.Save(context.Background(), d)
	require.NoError(t, err)
	assert.Equal(t, int64(202), d.ID, "Save must populate generated id back onto entity")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDisciplineItemRepoPG_Save_TransportError(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	d := freshItemForRepo(t)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO curriculum_section_items")).
		WillReturnError(fmt.Errorf("conn refused"))

	err := repo.Save(context.Background(), d)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save")
}

// ===== GetByID =====

func TestDisciplineItemRepoPG_GetByID_RowFound(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"id", "section_id", "title",
		"hours_lectures", "hours_practice", "hours_lab", "hours_self",
		"control_form", "credits", "semester",
		"order_index", "version", "created_at", "updated_at",
	}).AddRow(int64(202), int64(7), "Программирование",
		36, 18, 36, 90,
		"differential_zachet", 5, 3, 2, 8, now, now)

	mock.ExpectQuery(regexp.QuoteMeta("FROM curriculum_section_items WHERE id = $1")).
		WithArgs(int64(202)).
		WillReturnRows(rows)

	got, err := repo.GetByID(context.Background(), 202)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, int64(202), got.ID)
	assert.Equal(t, int64(7), got.SectionID())
	assert.Equal(t, "Программирование", got.Title())
	assert.Equal(t, entities.ControlFormDifferentialZachet, got.ControlForm())
	assert.Equal(t, 5, got.Credits())
	assert.Equal(t, 3, got.Semester())
	assert.Equal(t, 8, got.Version())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDisciplineItemRepoPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM curriculum_section_items WHERE id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	got, err := repo.GetByID(context.Background(), 999)
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, repositories.ErrDisciplineItemNotFound))
}

func TestDisciplineItemRepoPG_GetByID_TransportError(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM curriculum_section_items WHERE id = $1")).
		WithArgs(int64(10)).
		WillReturnError(fmt.Errorf("conn refused"))

	_, err := repo.GetByID(context.Background(), 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get by id")
}

// ===== ListBySectionID =====

func TestDisciplineItemRepoPG_ListBySectionID_OrderedResults(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	now := time.Now()
	earlier := now.Add(-1 * time.Hour)

	rows := sqlmock.NewRows([]string{
		"id", "section_id", "title",
		"hours_lectures", "hours_practice", "hours_lab", "hours_self",
		"control_form", "credits", "semester",
		"order_index", "version", "created_at", "updated_at",
	}).
		AddRow(int64(202), int64(7), "Дисциплина 1",
			18, 18, 0, 36,
			"zachet", 2, 1, 0, 0, earlier, earlier).
		AddRow(int64(203), int64(7), "Дисциплина 2",
			36, 36, 0, 72,
			"exam", 4, 1, 1, 0, now, now)

	mock.ExpectQuery(regexp.QuoteMeta("FROM curriculum_section_items WHERE section_id = $1")).
		WithArgs(int64(7)).
		WillReturnRows(rows)

	got, err := repo.ListBySectionID(context.Background(), 7)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, int64(202), got[0].ID)
	assert.Equal(t, int64(203), got[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDisciplineItemRepoPG_ListBySectionID_EmptyResult(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	rows := sqlmock.NewRows([]string{
		"id", "section_id", "title",
		"hours_lectures", "hours_practice", "hours_lab", "hours_self",
		"control_form", "credits", "semester",
		"order_index", "version", "created_at", "updated_at",
	})
	mock.ExpectQuery(regexp.QuoteMeta("FROM curriculum_section_items WHERE section_id = $1")).
		WithArgs(int64(99)).
		WillReturnRows(rows)

	got, err := repo.ListBySectionID(context.Background(), 99)
	require.NoError(t, err)
	assert.Len(t, got, 0)
}

// ===== Update =====
//
// CLAUDE.md ≥3-variant gate с WithArgs от первого draft per chronicles
// lesson — covers все 3 distinct outcomes of optimistic-locking Update
// (HappyPath / VersionConflict / NotFound) с per-branch arg pinning.
// No mutation-resistance gap.

func TestDisciplineItemRepoPG_Update_HappyPath(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	now := time.Date(2026, 5, 9, 14, 0, 0, 0, time.UTC)
	d := entities.ReconstituteDisciplineItem(202, 7, "Старый",
		36, 36, 0, 72, entities.ControlFormExam, 4, 1, 0, 5, now, now)
	require.NoError(t, d.UpdateBasics("Новый", 24, 24, 36, 96,
		entities.ControlFormDifferentialZachet, 5, 2, 1, now))
	priorVersion := d.Version()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE curriculum_section_items SET")).
		WithArgs("Новый", 24, 24, 36, 96,
			"differential_zachet", 5, 2, 1,
			d.UpdatedAt(), int64(202), priorVersion).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Update(context.Background(), d))
	assert.Equal(t, priorVersion+1, d.Version(),
		"Update must bump entity version on RowsAffected==1")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDisciplineItemRepoPG_Update_VersionConflict(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	now := time.Date(2026, 5, 9, 14, 0, 0, 0, time.UTC)
	d := entities.ReconstituteDisciplineItem(202, 7, "T",
		18, 18, 0, 36, entities.ControlFormZachet, 2, 1, 0, 3, now, now)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE curriculum_section_items SET")).
		WithArgs("T", 18, 18, 0, 36,
			"zachet", 2, 1, 0,
			d.UpdatedAt(), int64(202), 3).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM curriculum_section_items WHERE id = $1")).
		WithArgs(int64(202)).
		WillReturnRows(sqlmock.NewRows([]string{"?column?"}).AddRow(1))

	err := repo.Update(context.Background(), d)
	assert.True(t, errors.Is(err, repositories.ErrDisciplineItemVersionConflict),
		"err must wrap ErrDisciplineItemVersionConflict, got %v", err)
	assert.Equal(t, 3, d.Version(), "Update must NOT bump version on conflict")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDisciplineItemRepoPG_Update_NotFound(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	now := time.Date(2026, 5, 9, 14, 0, 0, 0, time.UTC)
	d := entities.ReconstituteDisciplineItem(999, 7, "T",
		18, 18, 0, 36, entities.ControlFormZachet, 2, 1, 0, 5, now, now)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE curriculum_section_items SET")).
		WithArgs("T", 18, 18, 0, 36,
			"zachet", 2, 1, 0,
			d.UpdatedAt(), int64(999), 5).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM curriculum_section_items WHERE id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	err := repo.Update(context.Background(), d)
	assert.True(t, errors.Is(err, repositories.ErrDisciplineItemNotFound),
		"err must wrap ErrDisciplineItemNotFound, got %v", err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDisciplineItemRepoPG_Update_TransportError(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	d := entities.ReconstituteDisciplineItem(202, 7, "T",
		18, 18, 0, 36, entities.ControlFormZachet, 2, 1, 0, 0,
		time.Now(), time.Now())
	mock.ExpectExec(regexp.QuoteMeta("UPDATE curriculum_section_items SET")).
		WillReturnError(fmt.Errorf("conn refused"))

	err := repo.Update(context.Background(), d)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update")
}

// ===== Delete =====

func TestDisciplineItemRepoPG_Delete_HappyPath(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM curriculum_section_items WHERE id = $1")).
		WithArgs(int64(202)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Delete(context.Background(), 202))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDisciplineItemRepoPG_Delete_NotFound(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM curriculum_section_items WHERE id = $1")).
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 999)
	assert.True(t, errors.Is(err, repositories.ErrDisciplineItemNotFound))
}

func TestDisciplineItemRepoPG_Delete_TransportError(t *testing.T) {
	repo, mock := newDisciplineItemRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM curriculum_section_items WHERE id = $1")).
		WillReturnError(fmt.Errorf("conn refused"))

	err := repo.Delete(context.Background(), 202)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete")
}

func TestDisciplineItemRepositoryPG_AggregateHoursByYear(t *testing.T) {
	cases := []struct {
		name string
		year int
		rows *sqlmock.Rows
		want []repositories.DisciplineItemHoursAgg
	}{
		{
			name: "no matching rows returns empty slice",
			year: 2026,
			rows: sqlmock.NewRows([]string{"id", "title", "lectures", "practice", "lab", "self_study"}),
			want: nil,
		},
		{
			name: "rows aggregated per curriculum",
			year: 2026,
			rows: sqlmock.NewRows([]string{"id", "title", "lectures", "practice", "lab", "self_study"}).
				AddRow(int64(11), "ИВТ-2026", 64, 32, 16, 88).
				AddRow(int64(12), "ПИ-2026", 48, 48, 0, 64),
			want: []repositories.DisciplineItemHoursAgg{
				{CurriculumID: 11, CurriculumTitle: "ИВТ-2026", Lectures: 64, Practice: 32, Lab: 16, SelfStudy: 88},
				{CurriculumID: 12, CurriculumTitle: "ПИ-2026", Lectures: 48, Practice: 48, Lab: 0, SelfStudy: 64},
			},
		},
		{
			name: "single curriculum zero-hours row still present",
			year: 2025,
			rows: sqlmock.NewRows([]string{"id", "title", "lectures", "practice", "lab", "self_study"}).
				AddRow(int64(7), "ИС-2025", 0, 0, 0, 0),
			want: []repositories.DisciplineItemHoursAgg{
				{CurriculumID: 7, CurriculumTitle: "ИС-2025", Lectures: 0, Practice: 0, Lab: 0, SelfStudy: 0},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := newDisciplineItemRepoMock(t)

			mock.ExpectQuery(`SELECT c.id, c.title,\s+COALESCE\(SUM\(ci.hours_lectures\), 0\).*FROM curricula c\s+LEFT JOIN curriculum_sections s ON s.curriculum_id = c.id\s+LEFT JOIN curriculum_section_items ci ON ci.section_id = s.id\s+WHERE c.year = \$1\s+GROUP BY c.id, c.title`).
				WithArgs(tc.year).
				WillReturnRows(tc.rows)

			got, err := repo.AggregateHoursByYear(context.Background(), tc.year)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}

	t.Run("query error propagates wrapped", func(t *testing.T) {
		repo, mock := newDisciplineItemRepoMock(t)

		mock.ExpectQuery(`SELECT c.id, c.title.*FROM curricula c`).
			WithArgs(2026).
			WillReturnError(fmt.Errorf("conn refused"))

		got, err := repo.AggregateHoursByYear(context.Background(), 2026)
		require.Error(t, err)
		require.Nil(t, got)
	})
}
