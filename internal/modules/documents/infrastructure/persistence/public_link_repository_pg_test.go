package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
)

func newPLRepoMock(t *testing.T) (*PublicLinkRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewPublicLinkRepositoryPG(db), mock
}

var plJoinCols = []string{
	"id", "document_id", "token", "permission", "created_by",
	"expires_at", "max_uses", "use_count", "password_hash",
	"is_active", "created_at", "updated_at",
	"document_title", "created_by_name",
}

func addPLRow(rows *sqlmock.Rows, id int64, token string) *sqlmock.Rows {
	now := time.Now()
	return rows.AddRow(
		id, int64(1), token, "read", int64(5),
		nil, nil, 0, nil,
		true, now, now,
		ptrStr("Doc Title"), ptrStr("User Name"),
	)
}

func newPLRows() *sqlmock.Rows { return sqlmock.NewRows(plJoinCols) }

func TestNewPublicLinkRepositoryPG(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	assert.NotNil(t, NewPublicLinkRepositoryPG(db))
}

func TestPLRepo_Create_Success(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	link := &entities.PublicLink{DocumentID: 1, Token: "abc", Permission: "read", CreatedBy: 5, IsActive: true}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO document_public_links")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.Create(context.Background(), link)
	require.NoError(t, err)
	assert.Equal(t, int64(1), link.ID)
}

func TestPLRepo_Create_Error(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	link := &entities.PublicLink{DocumentID: 1, Token: "abc"}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO document_public_links")).
		WillReturnError(fmt.Errorf("db err"))

	err := repo.Create(context.Background(), link)
	assert.Error(t, err)
}

func TestPLRepo_Update_Success(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	link := &entities.PublicLink{ID: 1, Permission: "write", IsActive: true}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE document_public_links")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), link)
	require.NoError(t, err)
}

func TestPLRepo_Update_NotFound(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	link := &entities.PublicLink{ID: 999}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE document_public_links")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(context.Background(), link)
	assert.ErrorIs(t, err, domainErrors.ErrNotFound)
}

func TestPLRepo_Delete_Success(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_public_links WHERE id = $1")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestPLRepo_Delete_NotFound(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_public_links WHERE id = $1")).
		WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 999)
	assert.ErrorIs(t, err, domainErrors.ErrNotFound)
}

func TestPLRepo_GetByID_Success(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	rows := addPLRow(newPLRows(),1, "abc")
	mock.ExpectQuery(regexp.QuoteMeta("WHERE pl.id = $1")).
		WithArgs(int64(1)).WillReturnRows(rows)

	link, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "abc", link.Token)
}

func TestPLRepo_GetByID_NotFound(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE pl.id = $1")).
		WithArgs(int64(999)).WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByID(context.Background(), 999)
	assert.ErrorIs(t, err, domainErrors.ErrNotFound)
}

func TestPLRepo_GetByToken_Success(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	rows := addPLRow(newPLRows(),1, "token123")
	mock.ExpectQuery(regexp.QuoteMeta("WHERE pl.token = $1")).
		WithArgs("token123").WillReturnRows(rows)

	link, err := repo.GetByToken(context.Background(), "token123")
	require.NoError(t, err)
	assert.Equal(t, "token123", link.Token)
}

func TestPLRepo_GetByToken_NotFound(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE pl.token = $1")).
		WithArgs("nonexistent").WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByToken(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, domainErrors.ErrNotFound)
}

func TestPLRepo_GetByDocumentID(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	rows := addPLRow(newPLRows(),1, "abc")
	mock.ExpectQuery(regexp.QuoteMeta("WHERE pl.document_id = $1")).
		WithArgs(int64(1)).WillReturnRows(rows)

	links, err := repo.GetByDocumentID(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, links, 1)
}

func TestPLRepo_GetByCreatedBy(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	rows := sqlmock.NewRows(plJoinCols)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE pl.created_by = $1")).
		WithArgs(int64(5)).WillReturnRows(rows)

	links, err := repo.GetByCreatedBy(context.Background(), 5)
	require.NoError(t, err)
	assert.Empty(t, links)
}

func TestPLRepo_GetActiveByDocumentID(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	rows := addPLRow(newPLRows(),1, "active")
	mock.ExpectQuery(regexp.QuoteMeta("AND pl.is_active = true")).
		WithArgs(int64(1)).WillReturnRows(rows)

	links, err := repo.GetActiveByDocumentID(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, links, 1)
}

func TestPLRepo_IncrementUseCount_Success(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("SET use_count = use_count + 1")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.IncrementUseCount(context.Background(), 1)
	require.NoError(t, err)
}

func TestPLRepo_IncrementUseCount_NotFound(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("SET use_count = use_count + 1")).
		WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.IncrementUseCount(context.Background(), 999)
	assert.ErrorIs(t, err, domainErrors.ErrNotFound)
}

func TestPLRepo_Deactivate_Success(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("SET is_active = false")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Deactivate(context.Background(), 1)
	require.NoError(t, err)
}

func TestPLRepo_Deactivate_NotFound(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("SET is_active = false")).
		WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Deactivate(context.Background(), 999)
	assert.ErrorIs(t, err, domainErrors.ErrNotFound)
}

func TestPLRepo_Activate_Success(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("SET is_active = true")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Activate(context.Background(), 1)
	require.NoError(t, err)
}

func TestPLRepo_Activate_NotFound(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("SET is_active = true")).
		WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Activate(context.Background(), 999)
	assert.ErrorIs(t, err, domainErrors.ErrNotFound)
}

func TestPLRepo_DeleteByDocumentID(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_public_links WHERE document_id = $1")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 3))

	err := repo.DeleteByDocumentID(context.Background(), 1)
	require.NoError(t, err)
}

func TestPLRepo_DeactivateExpired(t *testing.T) {
	repo, mock := newPLRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("SET is_active = false")).
		WillReturnResult(sqlmock.NewResult(0, 4))

	count, err := repo.DeactivateExpired(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(4), count)
}
