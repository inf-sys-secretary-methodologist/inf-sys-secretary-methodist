package persistence

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
)

func newResolverMock(t *testing.T) (*StudentGroupResolverPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewStudentGroupResolverPG(db), mock
}

func TestStudentGroupResolverPG_ResolveGroupID_Found(t *testing.T) {
	repo, mock := newResolverMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("FROM external_students es")).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(5)))

	id, err := repo.ResolveGroupID(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, int64(5), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStudentGroupResolverPG_ResolveGroupID_NotFound(t *testing.T) {
	repo, mock := newResolverMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("FROM external_students es")).
		WithArgs(int64(99)).
		WillReturnError(sql.ErrNoRows)

	id, err := repo.ResolveGroupID(context.Background(), 99)
	assert.Zero(t, id)
	assert.ErrorIs(t, err, usecases.ErrStudentGroupNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStudentGroupResolverPG_ResolveGroupID_DBError(t *testing.T) {
	repo, mock := newResolverMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("FROM external_students es")).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.ResolveGroupID(context.Background(), 1)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, usecases.ErrStudentGroupNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}
