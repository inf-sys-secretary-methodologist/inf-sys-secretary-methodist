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

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/entities"
)

func newFileVerRepoMock(t *testing.T) (*FileVersionRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewFileVersionRepositoryPG(db)
	return repo.(*FileVersionRepositoryPG), mock
}

var versionCols = []string{"id", "file_metadata_id", "version_number", "storage_key", "size", "checksum", "comment", "created_by", "created_at"}

func addVersionRow(rows *sqlmock.Rows, id int64, ver int) *sqlmock.Rows {
	now := time.Now()
	return rows.AddRow(id, int64(1), ver, fmt.Sprintf("key-v%d", ver), int64(1024), "abc", "comment", int64(1), now)
}

// ---- Create ----

func TestFileVersionCreate_Success(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)
	now := time.Now()
	ver := &entities.FileVersion{
		FileMetadataID: 1, VersionNumber: 1, StorageKey: "key", Size: 1024,
		Checksum: "abc", Comment: "initial", CreatedBy: 1, CreatedAt: now,
	}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO file_versions")).
		WithArgs(ver.FileMetadataID, ver.VersionNumber, ver.StorageKey, ver.Size, ver.Checksum, ver.Comment, ver.CreatedBy, ver.CreatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.Create(context.Background(), ver)
	require.NoError(t, err)
	assert.Equal(t, int64(1), ver.ID)
}

func TestFileVersionCreate_Error(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)
	ver := &entities.FileVersion{FileMetadataID: 1}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO file_versions")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert error"))

	err := repo.Create(context.Background(), ver)
	assert.Error(t, err)
}

// ---- GetByID ----

func TestFileVersionGetByID_Success(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	rows := sqlmock.NewRows(versionCols)
	addVersionRow(rows, 1, 1)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, file_metadata_id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	ver, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), ver.ID)
}

func TestFileVersionGetByID_NotFound(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, file_metadata_id")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByID(context.Background(), 999)
	assert.Error(t, err)
}

// ---- GetByFileMetadataID ----

func TestFileVersionGetByFileMetadataID_Success(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	rows := sqlmock.NewRows(versionCols)
	addVersionRow(rows, 1, 2)
	addVersionRow(rows, 2, 1)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE file_metadata_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	versions, err := repo.GetByFileMetadataID(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, versions, 2)
}

func TestFileVersionGetByFileMetadataID_QueryError(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE file_metadata_id = $1")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetByFileMetadataID(context.Background(), 1)
	assert.Error(t, err)
}

func TestFileVersionGetByFileMetadataID_ScanError(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE file_metadata_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.GetByFileMetadataID(context.Background(), 1)
	assert.Error(t, err)
}

// ---- GetLatestVersion ----

func TestFileVersionGetLatestVersion_Success(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	rows := sqlmock.NewRows(versionCols)
	addVersionRow(rows, 1, 3)
	mock.ExpectQuery(regexp.QuoteMeta("ORDER BY version_number DESC")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	ver, err := repo.GetLatestVersion(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 3, ver.VersionNumber)
}

func TestFileVersionGetLatestVersion_NotFound(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("ORDER BY version_number DESC")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetLatestVersion(context.Background(), 999)
	assert.Error(t, err)
}

// ---- GetByVersionNumber ----

func TestFileVersionGetByVersionNumber_Success(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	rows := sqlmock.NewRows(versionCols)
	addVersionRow(rows, 1, 2)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE file_metadata_id = $1 AND version_number = $2")).
		WithArgs(int64(1), 2).
		WillReturnRows(rows)

	ver, err := repo.GetByVersionNumber(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 2, ver.VersionNumber)
}

func TestFileVersionGetByVersionNumber_NotFound(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE file_metadata_id = $1 AND version_number = $2")).
		WithArgs(int64(1), 99).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByVersionNumber(context.Background(), 1, 99)
	assert.Error(t, err)
}

// ---- Delete ----

func TestFileVersionDelete_Success(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM file_versions WHERE id")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestFileVersionDelete_NotFound(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM file_versions WHERE id")).
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 999)
	assert.Error(t, err)
}

func TestFileVersionDelete_ExecError(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM file_versions WHERE id")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("delete error"))

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestFileVersionDelete_RowsAffectedError(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM file_versions WHERE id")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows error")))

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

// ---- DeleteByFileMetadataID ----

func TestFileVersionDeleteByFileMetadataID_Success(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM file_versions WHERE file_metadata_id")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 3))

	err := repo.DeleteByFileMetadataID(context.Background(), 1)
	require.NoError(t, err)
}

func TestFileVersionDeleteByFileMetadataID_Error(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM file_versions WHERE file_metadata_id")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("delete error"))

	err := repo.DeleteByFileMetadataID(context.Background(), 1)
	assert.Error(t, err)
}

// ---- CountByFileMetadataID ----

func TestFileVersionCountByFileMetadataID_Success(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

	count, err := repo.CountByFileMetadataID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestFileVersionCountByFileMetadataID_Error(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("count error"))

	_, err := repo.CountByFileMetadataID(context.Background(), 1)
	assert.Error(t, err)
}

// ---- GetNextVersionNumber ----

func TestFileVersionGetNextVersionNumber_Success(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT MAX(version_number)")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(int64(3)))

	next, err := repo.GetNextVersionNumber(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 4, next)
}

func TestFileVersionGetNextVersionNumber_NoVersions(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT MAX(version_number)")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(nil))

	next, err := repo.GetNextVersionNumber(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 1, next)
}

func TestFileVersionGetNextVersionNumber_Error(t *testing.T) {
	repo, mock := newFileVerRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT MAX(version_number)")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetNextVersionNumber(context.Background(), 1)
	assert.Error(t, err)
}
