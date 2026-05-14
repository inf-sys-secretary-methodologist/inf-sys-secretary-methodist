package persistence

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDashboardRepoMock(t *testing.T) (*DashboardRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewDashboardRepositoryPG(db), mock
}

// countMethodSpec table-drives the 5 "count current period + count
// previous period" methods (Documents/Reports/Tasks/Events/Students)
// since they share an identical shape — only the FROM table differs.
type countMethodSpec struct {
	name  string
	table string
	call  func(repo *DashboardRepositoryPG) (int64, int64, error)
}

func TestDashboardRepositoryPG_CountMethods(t *testing.T) {
	cases := []countMethodSpec{
		{
			name:  "GetDocumentsCount",
			table: "documents",
			call: func(repo *DashboardRepositoryPG) (int64, int64, error) {
				r, err := repo.GetDocumentsCount(context.Background(), 30)
				if err != nil {
					return 0, 0, err
				}
				return r.Total, r.PreviousTotal, nil
			},
		},
		{
			name:  "GetReportsCount",
			table: "reports",
			call: func(repo *DashboardRepositoryPG) (int64, int64, error) {
				r, err := repo.GetReportsCount(context.Background(), 30)
				if err != nil {
					return 0, 0, err
				}
				return r.Total, r.PreviousTotal, nil
			},
		},
		{
			name:  "GetTasksCount",
			table: "tasks",
			call: func(repo *DashboardRepositoryPG) (int64, int64, error) {
				r, err := repo.GetTasksCount(context.Background(), 30)
				if err != nil {
					return 0, 0, err
				}
				return r.Total, r.PreviousTotal, nil
			},
		},
		{
			name:  "GetEventsCount",
			table: "events",
			call: func(repo *DashboardRepositoryPG) (int64, int64, error) {
				r, err := repo.GetEventsCount(context.Background(), 30)
				if err != nil {
					return 0, 0, err
				}
				return r.Total, r.PreviousTotal, nil
			},
		},
		{
			name:  "GetStudentsCount",
			table: "users",
			call: func(repo *DashboardRepositoryPG) (int64, int64, error) {
				r, err := repo.GetStudentsCount(context.Background(), 30)
				if err != nil {
					return 0, 0, err
				}
				return r.Total, r.PreviousTotal, nil
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/happy_path", func(t *testing.T) {
			repo, mock := newDashboardRepoMock(t)
			mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM " + tc.table)).
				WithArgs(sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(7)))
			mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM "+tc.table)).
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

			total, prev, err := tc.call(repo)
			require.NoError(t, err)
			assert.Equal(t, int64(7), total)
			assert.Equal(t, int64(3), prev)
			require.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run(tc.name+"/current_period_db_error", func(t *testing.T) {
			repo, mock := newDashboardRepoMock(t)
			mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM " + tc.table)).
				WithArgs(sqlmock.AnyArg()).
				WillReturnError(errors.New("db down"))

			_, _, err := tc.call(repo)
			assert.Error(t, err)
		})

		t.Run(tc.name+"/previous_period_db_error", func(t *testing.T) {
			repo, mock := newDashboardRepoMock(t)
			mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM " + tc.table)).
				WithArgs(sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(7)))
			mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM "+tc.table)).
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnError(errors.New("db down on previous"))

			_, _, err := tc.call(repo)
			assert.Error(t, err)
		})
	}
}

// trendMethodSpec table-drives the 4 trend methods (Documents/Reports/
// Tasks/Events). Identical shape, only the FROM table differs.
type trendMethodSpec struct {
	name  string
	table string
	call  func(repo *DashboardRepositoryPG, start, end time.Time) (int, error)
}

func TestDashboardRepositoryPG_TrendMethods(t *testing.T) {
	cases := []trendMethodSpec{
		{
			name:  "GetDocumentsTrend",
			table: "documents",
			call: func(repo *DashboardRepositoryPG, start, end time.Time) (int, error) {
				data, err := repo.GetDocumentsTrend(context.Background(), start, end)
				return len(data), err
			},
		},
		{
			name:  "GetReportsTrend",
			table: "reports",
			call: func(repo *DashboardRepositoryPG, start, end time.Time) (int, error) {
				data, err := repo.GetReportsTrend(context.Background(), start, end)
				return len(data), err
			},
		},
		{
			name:  "GetTasksTrend",
			table: "tasks",
			call: func(repo *DashboardRepositoryPG, start, end time.Time) (int, error) {
				data, err := repo.GetTasksTrend(context.Background(), start, end)
				return len(data), err
			},
		},
		{
			name:  "GetEventsTrend",
			table: "events",
			call: func(repo *DashboardRepositoryPG, start, end time.Time) (int, error) {
				data, err := repo.GetEventsTrend(context.Background(), start, end)
				return len(data), err
			},
		},
	}

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	for _, tc := range cases {
		t.Run(tc.name+"/happy_path", func(t *testing.T) {
			repo, mock := newDashboardRepoMock(t)
			rows := sqlmock.NewRows([]string{"date", "count"}).
				AddRow(time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC), int64(5)).
				AddRow(time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC), int64(7))
			mock.ExpectQuery(regexp.QuoteMeta("FROM "+tc.table)).
				WithArgs(start, end).
				WillReturnRows(rows)

			count, err := tc.call(repo, start, end)
			require.NoError(t, err)
			assert.Equal(t, 2, count)
			require.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run(tc.name+"/db_error", func(t *testing.T) {
			repo, mock := newDashboardRepoMock(t)
			mock.ExpectQuery(regexp.QuoteMeta("FROM "+tc.table)).
				WithArgs(start, end).
				WillReturnError(errors.New("db down"))

			_, err := tc.call(repo, start, end)
			assert.Error(t, err)
		})
	}
}

func TestDashboardRepositoryPG_GetRecentActivity(t *testing.T) {
	t.Run("returns activity entries + total count", func(t *testing.T) {
		repo, mock := newDashboardRepoMock(t)
		now := time.Now()
		// Scan reads 8 columns: id, type, action, title, description,
		// user_id, user_name, created_at.
		rows := sqlmock.NewRows([]string{"id", "type", "action", "title", "description", "user_id", "user_name", "created_at"}).
			AddRow(int64(1), "document", "created", "Test Doc", "desc1", int64(10), "Alice", now).
			AddRow(int64(2), "task", "created", "Task 1", "desc2", int64(11), "Bob", now)
		mock.ExpectQuery("UNION ALL").
			WithArgs(10).
			WillReturnRows(rows)
		// Separate count query
		mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
			WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(int64(42)))

		activities, total, err := repo.GetRecentActivity(context.Background(), 10)
		require.NoError(t, err)
		assert.Len(t, activities, 2)
		assert.Equal(t, int64(42), total)
	})

	t.Run("propagates UNION query error", func(t *testing.T) {
		repo, mock := newDashboardRepoMock(t)
		mock.ExpectQuery("UNION ALL").
			WithArgs(10).
			WillReturnError(errors.New("db down"))

		_, _, err := repo.GetRecentActivity(context.Background(), 10)
		assert.Error(t, err)
	})

	t.Run("propagates count query error", func(t *testing.T) {
		repo, mock := newDashboardRepoMock(t)
		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "type", "action", "title", "description", "user_id", "user_name", "created_at"}).
			AddRow(int64(1), "document", "created", "Test Doc", "desc1", int64(10), "Alice", now)
		mock.ExpectQuery("UNION ALL").
			WithArgs(10).
			WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
			WillReturnError(errors.New("count failed"))

		_, _, err := repo.GetRecentActivity(context.Background(), 10)
		assert.Error(t, err)
	})
}
