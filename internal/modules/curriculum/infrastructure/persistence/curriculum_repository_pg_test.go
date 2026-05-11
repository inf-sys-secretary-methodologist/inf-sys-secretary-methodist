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
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

func newCurriculumRepoMock(t *testing.T) (*CurriculumRepositoryPG, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewCurriculumRepositoryPG(db), mock
}

func freshDraft(t *testing.T, code string) *entities.Curriculum {
	t.Helper()
	c, err := entities.NewCurriculum(entities.NewCurriculumParams{
		Title:       "ИВТ-2026",
		Code:        code,
		Specialty:   "Информатика и вычислительная техника",
		Year:        2026,
		Description: "desc",
		CreatedBy:   42,
		Now:         time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	return c
}

func TestCurriculumRepositoryPG_GetByID(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)

	t.Run("row found populates entity fields", func(t *testing.T) {
		repo, mock := newCurriculumRepoMock(t)
		rows := sqlmock.NewRows([]string{
			"id", "title", "code", "specialty", "year", "description",
			"status", "created_by", "approved_by", "approved_at",
			"created_at", "updated_at",
		}).AddRow(int64(10), "ИВТ-2026", "09.03.04-2026",
			"Информатика и вычислительная техника", 2026, "desc",
			"draft", int64(42), sql.NullInt64{}, sql.NullTime{}, now, now)

		mock.ExpectQuery(regexp.QuoteMeta("FROM curricula WHERE id = $1")).
			WithArgs(int64(10)).
			WillReturnRows(rows)

		got, err := repo.GetByID(context.Background(), 10)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(10), got.ID)
		assert.Equal(t, "ИВТ-2026", got.Title())
		assert.Equal(t, "09.03.04-2026", got.Code())
		assert.Equal(t, 2026, got.Year())
		assert.Equal(t, entities.StatusDraft, got.Status())
		assert.Equal(t, int64(42), got.CreatedBy())
		assert.Nil(t, got.ApprovedBy())
		assert.Nil(t, got.ApprovedAt())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("approved row populates approved_by/at pointers", func(t *testing.T) {
		repo, mock := newCurriculumRepoMock(t)
		approvedAt := now.Add(48 * time.Hour)
		rows := sqlmock.NewRows([]string{
			"id", "title", "code", "specialty", "year", "description",
			"status", "created_by", "approved_by", "approved_at",
			"created_at", "updated_at",
		}).AddRow(int64(7), "T", "C", "S", 2026, "",
			"approved", int64(42),
			sql.NullInt64{Int64: 99, Valid: true},
			sql.NullTime{Time: approvedAt, Valid: true},
			now, now)

		mock.ExpectQuery(regexp.QuoteMeta("FROM curricula WHERE id = $1")).
			WithArgs(int64(7)).
			WillReturnRows(rows)

		got, err := repo.GetByID(context.Background(), 7)
		require.NoError(t, err)
		require.NotNil(t, got.ApprovedBy())
		assert.Equal(t, int64(99), *got.ApprovedBy())
		require.NotNil(t, got.ApprovedAt())
		assert.True(t, got.ApprovedAt().Equal(approvedAt))
	})

	t.Run("no rows returns ErrCurriculumNotFound sentinel", func(t *testing.T) {
		repo, mock := newCurriculumRepoMock(t)
		mock.ExpectQuery(regexp.QuoteMeta("FROM curricula WHERE id = $1")).
			WithArgs(int64(999)).
			WillReturnError(sql.ErrNoRows)

		got, err := repo.GetByID(context.Background(), 999)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, repositories.ErrCurriculumNotFound))
	})

	t.Run("transport error wraps original", func(t *testing.T) {
		repo, mock := newCurriculumRepoMock(t)
		mock.ExpectQuery(regexp.QuoteMeta("FROM curricula WHERE id = $1")).
			WithArgs(int64(10)).
			WillReturnError(fmt.Errorf("conn refused"))

		_, err := repo.GetByID(context.Background(), 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get by id")
	})
}

func TestCurriculumRepositoryPG_List(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)

	t.Run("no filters returns rows and total", func(t *testing.T) {
		repo, mock := newCurriculumRepoMock(t)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM curricula")).
			WithArgs("", sql.NullInt64{}, "", sql.NullInt64{}).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		rows := sqlmock.NewRows([]string{
			"id", "title", "code", "specialty", "year", "description",
			"status", "created_by", "approved_by", "approved_at",
			"created_at", "updated_at",
		}).
			AddRow(int64(1), "T1", "C1", "S", 2026, "",
				"draft", int64(42), sql.NullInt64{}, sql.NullTime{}, now, now).
			AddRow(int64(2), "T2", "C2", "S", 2025, "",
				"draft", int64(42), sql.NullInt64{}, sql.NullTime{}, now, now)

		mock.ExpectQuery(regexp.QuoteMeta("ORDER BY year DESC, created_at DESC")).
			WithArgs("", sql.NullInt64{}, "", sql.NullInt64{}, 50, 0).
			WillReturnRows(rows)

		got, err := repo.List(context.Background(),
			repositories.CurriculumListFilter{Limit: 50})
		require.NoError(t, err)
		assert.Equal(t, 2, got.Total)
		require.Len(t, got.Items, 2)
		assert.Equal(t, "T1", got.Items[0].Title())
		assert.Equal(t, "T2", got.Items[1].Title())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("status + year + specialty + created_by filters propagate", func(t *testing.T) {
		repo, mock := newCurriculumRepoMock(t)

		status := entities.StatusPendingApproval
		year := 2026
		creator := int64(42)

		expectStatus := "pending_approval"
		expectYear := sql.NullInt64{Int64: 2026, Valid: true}
		expectCreator := sql.NullInt64{Int64: creator, Valid: true}

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM curricula")).
			WithArgs(expectStatus, expectYear, "Информатика", expectCreator).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		mock.ExpectQuery(regexp.QuoteMeta("ORDER BY year DESC, created_at DESC")).
			WithArgs(expectStatus, expectYear, "Информатика", expectCreator, 25, 100).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "title", "code", "specialty", "year", "description",
				"status", "created_by", "approved_by", "approved_at",
				"created_at", "updated_at",
			}))

		got, err := repo.List(context.Background(), repositories.CurriculumListFilter{
			Status:    &status,
			Year:      &year,
			Specialty: "Информатика",
			CreatedBy: &creator,
			Limit:     25,
			Offset:    100,
		})
		require.NoError(t, err)
		assert.Equal(t, 0, got.Total)
		assert.Empty(t, got.Items)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("count error wrapped", func(t *testing.T) {
		repo, mock := newCurriculumRepoMock(t)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM curricula")).
			WithArgs("", sql.NullInt64{}, "", sql.NullInt64{}).
			WillReturnError(fmt.Errorf("conn refused"))

		_, err := repo.List(context.Background(),
			repositories.CurriculumListFilter{Limit: 50})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "count")
	})
}

func TestCurriculumRepositoryPG_Save(t *testing.T) {
	t.Run("happy path returns generated id and writes it back", func(t *testing.T) {
		repo, mock := newCurriculumRepoMock(t)

		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO curricula")).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(123)))

		c := freshDraft(t, "09.03.04-2026")
		err := repo.Save(context.Background(), c)
		require.NoError(t, err)
		assert.Equal(t, int64(123), c.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("unique violation maps to ErrCurriculumCodeExists", func(t *testing.T) {
		repo, mock := newCurriculumRepoMock(t)

		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO curricula")).
			WillReturnError(&pq.Error{Code: "23505"})

		c := freshDraft(t, "DUP-2026")
		err := repo.Save(context.Background(), c)
		assert.True(t, errors.Is(err, repositories.ErrCurriculumCodeExists),
			"expected ErrCurriculumCodeExists, got %v", err)
	})

	t.Run("transport error wraps original", func(t *testing.T) {
		repo, mock := newCurriculumRepoMock(t)

		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO curricula")).
			WillReturnError(fmt.Errorf("conn refused"))

		c := freshDraft(t, "X-2026")
		err := repo.Save(context.Background(), c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "save")
	})
}

func TestCurriculumRepositoryPG_Update(t *testing.T) {
	t.Run("happy path updates one row", func(t *testing.T) {
		repo, mock := newCurriculumRepoMock(t)

		mock.ExpectExec(regexp.QuoteMeta("UPDATE curricula SET")).
			WillReturnResult(sqlmock.NewResult(0, 1))

		c := freshDraft(t, "09.03.04-2026")
		c.ID = 7
		err := repo.Update(context.Background(), c)
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("zero rows updated returns ErrCurriculumNotFound", func(t *testing.T) {
		repo, mock := newCurriculumRepoMock(t)

		mock.ExpectExec(regexp.QuoteMeta("UPDATE curricula SET")).
			WillReturnResult(sqlmock.NewResult(0, 0))

		c := freshDraft(t, "09.03.04-2026")
		c.ID = 999
		err := repo.Update(context.Background(), c)
		assert.True(t, errors.Is(err, repositories.ErrCurriculumNotFound))
	})

	t.Run("unique violation maps to ErrCurriculumCodeExists", func(t *testing.T) {
		repo, mock := newCurriculumRepoMock(t)

		mock.ExpectExec(regexp.QuoteMeta("UPDATE curricula SET")).
			WillReturnError(&pq.Error{Code: "23505"})

		c := freshDraft(t, "DUP-2026")
		c.ID = 7
		err := repo.Update(context.Background(), c)
		assert.True(t, errors.Is(err, repositories.ErrCurriculumCodeExists))
	})

	t.Run("transport error wraps original", func(t *testing.T) {
		repo, mock := newCurriculumRepoMock(t)

		mock.ExpectExec(regexp.QuoteMeta("UPDATE curricula SET")).
			WillReturnError(fmt.Errorf("conn refused"))

		c := freshDraft(t, "X-2026")
		c.ID = 7
		err := repo.Update(context.Background(), c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "update")
	})
}

func TestCurriculumRepositoryPG_AggregateByYearSpecialty(t *testing.T) {
	cases := []struct {
		name string
		year int
		rows *sqlmock.Rows
		want []repositories.CurriculumYearSpecialtyAgg
	}{
		{
			name: "no matching rows returns empty slice",
			year: 2026,
			rows: sqlmock.NewRows([]string{"specialty", "status", "count"}),
			want: nil,
		},
		{
			name: "rows grouped by specialty and status",
			year: 2026,
			rows: sqlmock.NewRows([]string{"specialty", "status", "count"}).
				AddRow("Информатика и вычислительная техника", "approved", 3).
				AddRow("Информатика и вычислительная техника", "pending_approval", 1).
				AddRow("Прикладная информатика", "approved", 2),
			want: []repositories.CurriculumYearSpecialtyAgg{
				{Specialty: "Информатика и вычислительная техника", Status: entities.StatusApproved, Count: 3},
				{Specialty: "Информатика и вычислительная техника", Status: entities.StatusPendingApproval, Count: 1},
				{Specialty: "Прикладная информатика", Status: entities.StatusApproved, Count: 2},
			},
		},
		{
			name: "single specialty single status",
			year: 2025,
			rows: sqlmock.NewRows([]string{"specialty", "status", "count"}).
				AddRow("Информационные системы", "draft", 1),
			want: []repositories.CurriculumYearSpecialtyAgg{
				{Specialty: "Информационные системы", Status: entities.StatusDraft, Count: 1},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := newCurriculumRepoMock(t)

			mock.ExpectQuery(`SELECT specialty, status, COUNT\(\*\) FROM curricula\s+WHERE year = \$1\s+GROUP BY specialty, status`).
				WithArgs(tc.year).
				WillReturnRows(tc.rows)

			got, err := repo.AggregateByYearSpecialty(context.Background(), tc.year)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}

	t.Run("query error propagates wrapped", func(t *testing.T) {
		repo, mock := newCurriculumRepoMock(t)

		mock.ExpectQuery(`SELECT specialty, status, COUNT\(\*\) FROM curricula`).
			WithArgs(2026).
			WillReturnError(fmt.Errorf("conn refused"))

		got, err := repo.AggregateByYearSpecialty(context.Background(), 2026)
		require.Error(t, err)
		require.Nil(t, got)
	})
}
