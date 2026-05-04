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

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
)

func newAssignmentRepoMock(t *testing.T) (*AssignmentRepositoryPG, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewAssignmentRepositoryPG(db), mock
}

func TestAssignmentRepositoryPG_GetByID(t *testing.T) {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	due := now.Add(7 * 24 * time.Hour)

	tests := []struct {
		name      string
		setupMock func(sqlmock.Sqlmock)
		wantErr   error
		assertFn  func(t *testing.T, a any)
	}{
		{
			name: "row found populates entity fields",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "title", "description", "teacher_id", "group_name",
					"subject", "max_score", "due_date", "created_at", "updated_at",
				}).AddRow(int64(10), "L1", "desc", int64(42), "ИС-21",
					"Algo", 100, due, now, now)

				mock.ExpectQuery(regexp.QuoteMeta("FROM assignments WHERE id = $1")).
					WithArgs(int64(10)).
					WillReturnRows(rows)
			},
		},
		{
			name: "no rows returns ErrAssignmentNotFound sentinel",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("FROM assignments WHERE id = $1")).
					WithArgs(int64(999)).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: repositories.ErrAssignmentNotFound,
		},
		{
			name: "transport error wraps original",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("FROM assignments WHERE id = $1")).
					WithArgs(int64(10)).
					WillReturnError(fmt.Errorf("conn refused"))
			},
			wantErr: errors.New("get by id"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := newAssignmentRepoMock(t)
			tc.setupMock(mock)

			id := int64(10)
			if tc.name == "no rows returns ErrAssignmentNotFound sentinel" {
				id = 999
			}
			got, err := repo.GetByID(context.Background(), id)

			if tc.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tc.wantErr, repositories.ErrAssignmentNotFound) {
					assert.True(t, errors.Is(err, repositories.ErrAssignmentNotFound))
				} else {
					assert.Contains(t, err.Error(), "get by id")
				}
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, int64(10), got.ID)
			assert.Equal(t, "L1", got.Title())
			assert.Equal(t, int64(42), got.TeacherID())
			assert.Equal(t, "ИС-21", got.GroupName())
			assert.Equal(t, 100, got.MaxScore())
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAssignmentRepositoryPG_List(t *testing.T) {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	teacherID := int64(42)

	t.Run("no filters returns rows and total", func(t *testing.T) {
		repo, mock := newAssignmentRepoMock(t)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM assignments")).
			WithArgs(sql.NullInt64{}, "", "").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		rows := sqlmock.NewRows([]string{
			"id", "title", "description", "teacher_id", "group_name",
			"subject", "max_score", "due_date", "created_at", "updated_at",
		}).
			AddRow(int64(1), "L1", "d1", teacherID, "ИС-21", "Algo", 100, sql.NullTime{}, now, now).
			AddRow(int64(2), "L2", "d2", teacherID, "ИС-22", "Algo", 50, sql.NullTime{}, now, now)

		mock.ExpectQuery(regexp.QuoteMeta("FROM assignments")).
			WithArgs(sql.NullInt64{}, "", "", 50, 0).
			WillReturnRows(rows)

		got, err := repo.List(context.Background(), repositories.AssignmentListFilter{Limit: 50})
		require.NoError(t, err)
		assert.Equal(t, 2, got.Total)
		require.Len(t, got.Items, 2)
		assert.Equal(t, "L1", got.Items[0].Title())
		assert.Equal(t, "L2", got.Items[1].Title())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("teacher filter passes valid sql.NullInt64 to count + select", func(t *testing.T) {
		repo, mock := newAssignmentRepoMock(t)

		expectTID := sql.NullInt64{Int64: teacherID, Valid: true}

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM assignments")).
			WithArgs(expectTID, "", "").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{
			"id", "title", "description", "teacher_id", "group_name",
			"subject", "max_score", "due_date", "created_at", "updated_at",
		}).AddRow(int64(1), "L1", "d", teacherID, "ИС-21", "Algo", 100, sql.NullTime{}, now, now)

		mock.ExpectQuery(regexp.QuoteMeta("FROM assignments")).
			WithArgs(expectTID, "", "", 50, 0).
			WillReturnRows(rows)

		tid := teacherID
		got, err := repo.List(context.Background(), repositories.AssignmentListFilter{
			TeacherID: &tid, Limit: 50,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, got.Total)
		require.Len(t, got.Items, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("subject and group filters propagate", func(t *testing.T) {
		repo, mock := newAssignmentRepoMock(t)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM assignments")).
			WithArgs(sql.NullInt64{}, "Algo", "ИС-21").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		mock.ExpectQuery(regexp.QuoteMeta("FROM assignments")).
			WithArgs(sql.NullInt64{}, "Algo", "ИС-21", 25, 100).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "title", "description", "teacher_id", "group_name",
				"subject", "max_score", "due_date", "created_at", "updated_at",
			}))

		got, err := repo.List(context.Background(), repositories.AssignmentListFilter{
			Subject: "Algo", GroupName: "ИС-21", Limit: 25, Offset: 100,
		})
		require.NoError(t, err)
		assert.Equal(t, 0, got.Total)
		assert.Empty(t, got.Items)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("count query error wrapped", func(t *testing.T) {
		repo, mock := newAssignmentRepoMock(t)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM assignments")).
			WithArgs(sql.NullInt64{}, "", "").
			WillReturnError(fmt.Errorf("conn refused"))

		_, err := repo.List(context.Background(), repositories.AssignmentListFilter{Limit: 50})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "count")
	})

	t.Run("select query error wrapped", func(t *testing.T) {
		repo, mock := newAssignmentRepoMock(t)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM assignments")).
			WithArgs(sql.NullInt64{}, "", "").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		mock.ExpectQuery(regexp.QuoteMeta("FROM assignments")).
			WithArgs(sql.NullInt64{}, "", "", 50, 0).
			WillReturnError(fmt.Errorf("query failed"))

		_, err := repo.List(context.Background(), repositories.AssignmentListFilter{Limit: 50})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "list")
	})
}
