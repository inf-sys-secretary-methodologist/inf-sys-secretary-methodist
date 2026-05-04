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
