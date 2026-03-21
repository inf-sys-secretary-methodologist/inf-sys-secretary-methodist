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

func newFileMetaRepoMock(t *testing.T) (*FileMetadataRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewFileMetadataRepositoryPG(db)
	return repo.(*FileMetadataRepositoryPG), mock
}

var fileCols = []string{
	"id", "original_name", "storage_key", "size", "mime_type", "checksum", "uploaded_by",
	"document_id", "task_id", "announcement_id", "is_temporary", "expires_at",
	"created_at", "updated_at", "deleted_at",
}

func addFileRow(rows *sqlmock.Rows, id int64, name string) *sqlmock.Rows {
	now := time.Now()
	return rows.AddRow(id, name, "key-"+name, int64(1024), "application/pdf", "abc123", int64(1),
		nil, nil, nil, false, nil, now, now, nil)
}

// ---- Create ----

func TestFileMetadataCreate_Success(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)
	now := time.Now()
	file := &entities.FileMetadata{
		OriginalName: "test.pdf", StorageKey: "key", Size: 1024, MimeType: "application/pdf",
		Checksum: "abc", UploadedBy: 1, IsTemporary: true, CreatedAt: now, UpdatedAt: now,
	}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO file_metadata")).
		WithArgs(file.OriginalName, file.StorageKey, file.Size, file.MimeType, file.Checksum,
			file.UploadedBy, file.DocumentID, file.TaskID, file.AnnouncementID, file.IsTemporary,
			file.ExpiresAt, file.CreatedAt, file.UpdatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.Create(context.Background(), file)
	require.NoError(t, err)
	assert.Equal(t, int64(1), file.ID)
}

func TestFileMetadataCreate_Error(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)
	file := &entities.FileMetadata{OriginalName: "test.pdf"}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO file_metadata")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert error"))

	err := repo.Create(context.Background(), file)
	assert.Error(t, err)
}

// ---- GetByID ----

func TestFileMetadataGetByID_Success(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	rows := sqlmock.NewRows(fileCols)
	addFileRow(rows, 1, "test.pdf")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, original_name")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	file, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "test.pdf", file.OriginalName)
}

func TestFileMetadataGetByID_NotFound(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, original_name")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByID(context.Background(), 999)
	assert.Error(t, err)
}

// ---- GetByStorageKey ----

func TestFileMetadataGetByStorageKey_Success(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	rows := sqlmock.NewRows(fileCols)
	addFileRow(rows, 1, "test.pdf")
	mock.ExpectQuery(regexp.QuoteMeta("WHERE storage_key = $1")).
		WithArgs("key-test.pdf").
		WillReturnRows(rows)

	file, err := repo.GetByStorageKey(context.Background(), "key-test.pdf")
	require.NoError(t, err)
	assert.Equal(t, "test.pdf", file.OriginalName)
}

func TestFileMetadataGetByStorageKey_NotFound(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE storage_key = $1")).
		WithArgs("nope").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByStorageKey(context.Background(), "nope")
	assert.Error(t, err)
}

// ---- Update ----

func TestFileMetadataUpdate_Success(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)
	now := time.Now()
	file := &entities.FileMetadata{ID: 1, OriginalName: "updated.pdf", StorageKey: "key", Size: 2048, MimeType: "application/pdf", Checksum: "xyz", UpdatedAt: now}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE file_metadata")).
		WithArgs(file.OriginalName, file.StorageKey, file.Size, file.MimeType, file.Checksum,
			file.DocumentID, file.TaskID, file.AnnouncementID, file.IsTemporary, file.ExpiresAt, file.UpdatedAt, file.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), file)
	require.NoError(t, err)
}

func TestFileMetadataUpdate_NotFound(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)
	file := &entities.FileMetadata{ID: 999}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE file_metadata")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(context.Background(), file)
	assert.Error(t, err)
}

func TestFileMetadataUpdate_ExecError(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)
	file := &entities.FileMetadata{ID: 1}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE file_metadata")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("exec error"))

	err := repo.Update(context.Background(), file)
	assert.Error(t, err)
}

func TestFileMetadataUpdate_RowsAffectedError(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)
	file := &entities.FileMetadata{ID: 1}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE file_metadata")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows error")))

	err := repo.Update(context.Background(), file)
	assert.Error(t, err)
}

// ---- Delete (soft) ----

func TestFileMetadataDelete_Success(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE file_metadata SET deleted_at")).
		WithArgs(sqlmock.AnyArg(), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestFileMetadataDelete_NotFound(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE file_metadata SET deleted_at")).
		WithArgs(sqlmock.AnyArg(), int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 999)
	assert.Error(t, err)
}

func TestFileMetadataDelete_ExecError(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE file_metadata SET deleted_at")).
		WithArgs(sqlmock.AnyArg(), int64(1)).
		WillReturnError(fmt.Errorf("delete error"))

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestFileMetadataDelete_RowsAffectedError(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE file_metadata SET deleted_at")).
		WithArgs(sqlmock.AnyArg(), int64(1)).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows error")))

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

// ---- HardDelete ----

func TestFileMetadataHardDelete_Success(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM file_metadata")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.HardDelete(context.Background(), 1)
	require.NoError(t, err)
}

func TestFileMetadataHardDelete_NotFound(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM file_metadata")).
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.HardDelete(context.Background(), 999)
	assert.Error(t, err)
}

func TestFileMetadataHardDelete_ExecError(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM file_metadata")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("delete error"))

	err := repo.HardDelete(context.Background(), 1)
	assert.Error(t, err)
}

func TestFileMetadataHardDelete_RowsAffectedError(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM file_metadata")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows error")))

	err := repo.HardDelete(context.Background(), 1)
	assert.Error(t, err)
}

// ---- List ----

func TestFileMetadataList_Success(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	rows := sqlmock.NewRows(fileCols)
	addFileRow(rows, 1, "file1.pdf")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, original_name")).
		WithArgs(10, 0).
		WillReturnRows(rows)

	files, err := repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, files, 1)
}

func TestFileMetadataList_DefaultLimits(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, original_name")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows(fileCols))

	_, err := repo.List(context.Background(), -1, -5)
	require.NoError(t, err)
}

func TestFileMetadataList_MaxLimit(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, original_name")).
		WithArgs(100, 0).
		WillReturnRows(sqlmock.NewRows(fileCols))

	_, err := repo.List(context.Background(), 500, 0)
	require.NoError(t, err)
}

func TestFileMetadataList_QueryError(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, original_name")).
		WithArgs(10, 0).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.List(context.Background(), 10, 0)
	assert.Error(t, err)
}

func TestFileMetadataList_ScanError(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, original_name")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.List(context.Background(), 10, 0)
	assert.Error(t, err)
}

// ---- Count ----

func TestFileMetadataCount_Success(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))

	count, err := repo.Count(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestFileMetadataCount_Error(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnError(fmt.Errorf("count error"))

	_, err := repo.Count(context.Background())
	assert.Error(t, err)
}

// ---- GetByDocumentID ----

func TestFileMetadataGetByDocumentID_Success(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	rows := sqlmock.NewRows(fileCols)
	addFileRow(rows, 1, "doc.pdf")
	mock.ExpectQuery(regexp.QuoteMeta("WHERE document_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	files, err := repo.GetByDocumentID(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, files, 1)
}

// ---- GetByTaskID ----

func TestFileMetadataGetByTaskID_Success(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	rows := sqlmock.NewRows(fileCols)
	addFileRow(rows, 1, "task.pdf")
	mock.ExpectQuery(regexp.QuoteMeta("WHERE task_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	files, err := repo.GetByTaskID(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, files, 1)
}

// ---- GetByAnnouncementID ----

func TestFileMetadataGetByAnnouncementID_Success(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	rows := sqlmock.NewRows(fileCols)
	addFileRow(rows, 1, "announce.pdf")
	mock.ExpectQuery(regexp.QuoteMeta("WHERE announcement_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	files, err := repo.GetByAnnouncementID(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, files, 1)
}

// ---- GetByUploadedBy ----

func TestFileMetadataGetByUploadedBy_Success(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	rows := sqlmock.NewRows(fileCols)
	addFileRow(rows, 1, "user.pdf")
	mock.ExpectQuery(regexp.QuoteMeta("WHERE uploaded_by = $1")).
		WithArgs(int64(1), 10, 0).
		WillReturnRows(rows)

	files, err := repo.GetByUploadedBy(context.Background(), 1, 10, 0)
	require.NoError(t, err)
	assert.Len(t, files, 1)
}

func TestFileMetadataGetByUploadedBy_DefaultLimits(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE uploaded_by = $1")).
		WithArgs(int64(1), 10, 0).
		WillReturnRows(sqlmock.NewRows(fileCols))

	_, err := repo.GetByUploadedBy(context.Background(), 1, -1, -5)
	require.NoError(t, err)
}

func TestFileMetadataGetByUploadedBy_MaxLimit(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE uploaded_by = $1")).
		WithArgs(int64(1), 100, 0).
		WillReturnRows(sqlmock.NewRows(fileCols))

	_, err := repo.GetByUploadedBy(context.Background(), 1, 500, 0)
	require.NoError(t, err)
}

// ---- GetExpiredTemporaryFiles ----

func TestFileMetadataGetExpiredTemporaryFiles_Success(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	rows := sqlmock.NewRows(fileCols)
	addFileRow(rows, 1, "temp.pdf")
	mock.ExpectQuery(regexp.QuoteMeta("WHERE is_temporary = true")).
		WithArgs(sqlmock.AnyArg(), 100).
		WillReturnRows(rows)

	files, err := repo.GetExpiredTemporaryFiles(context.Background(), 100)
	require.NoError(t, err)
	assert.Len(t, files, 1)
}

func TestFileMetadataGetExpiredTemporaryFiles_DefaultLimit(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE is_temporary = true")).
		WithArgs(sqlmock.AnyArg(), 100).
		WillReturnRows(sqlmock.NewRows(fileCols))

	_, err := repo.GetExpiredTemporaryFiles(context.Background(), -1)
	require.NoError(t, err)
}

// ---- CleanupExpired ----

func TestFileMetadataCleanupExpired_Success(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE file_metadata")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 3))

	count, err := repo.CleanupExpired(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestFileMetadataCleanupExpired_Error(t *testing.T) {
	repo, mock := newFileMetaRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE file_metadata")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("exec error"))

	_, err := repo.CleanupExpired(context.Background())
	assert.Error(t, err)
}
