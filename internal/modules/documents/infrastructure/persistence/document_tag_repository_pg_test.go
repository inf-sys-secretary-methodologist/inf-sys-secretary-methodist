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
)

func newTagRepoMock(t *testing.T) (*DocumentTagRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewDocumentTagRepositoryPG(db), mock
}

var tagCols = []string{"id", "name", "color", "created_at"}

func TestNewDocumentTagRepositoryPG(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	repo := NewDocumentTagRepositoryPG(db)
	assert.NotNil(t, repo)
}

func TestTagRepo_Create_Success(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	tag := &entities.DocumentTag{Name: "test"}
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO document_tags")).
		WithArgs("test", tag.Color).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(1), now))

	err := repo.Create(context.Background(), tag)
	require.NoError(t, err)
	assert.Equal(t, int64(1), tag.ID)
}

func TestTagRepo_Create_Duplicate(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	tag := &entities.DocumentTag{Name: "dup"}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO document_tags")).
		WillReturnError(fmt.Errorf("duplicate key value violates unique constraint"))

	err := repo.Create(context.Background(), tag)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "тег с таким именем уже существует")
}

func TestTagRepo_Create_OtherError(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	tag := &entities.DocumentTag{Name: "test"}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO document_tags")).
		WillReturnError(fmt.Errorf("connection error"))

	err := repo.Create(context.Background(), tag)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create tag")
}

func TestTagRepo_Update_Success(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	tag := &entities.DocumentTag{ID: 1, Name: "updated"}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE document_tags")).
		WithArgs("updated", tag.Color, int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), tag)
	require.NoError(t, err)
}

func TestTagRepo_Update_NotFound(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	tag := &entities.DocumentTag{ID: 999, Name: "updated"}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE document_tags")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(context.Background(), tag)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "тег не найден")
}

func TestTagRepo_Update_Duplicate(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	tag := &entities.DocumentTag{ID: 1, Name: "dup"}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE document_tags")).
		WillReturnError(fmt.Errorf("duplicate key"))

	err := repo.Update(context.Background(), tag)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "тег с таким именем уже существует")
}

func TestTagRepo_Delete_Success(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_tags WHERE id = $1")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestTagRepo_Delete_NotFound(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_tags WHERE id = $1")).
		WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "тег не найден")
}

func TestTagRepo_GetByID_Success(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(tagCols).AddRow(1, "test", nil, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, color, created_at FROM document_tags WHERE id")).
		WithArgs(int64(1)).WillReturnRows(rows)

	tag, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "test", tag.Name)
}

func TestTagRepo_GetByID_NotFound(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, color, created_at FROM document_tags WHERE id")).
		WithArgs(int64(999)).WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByID(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "тег не найден")
}

func TestTagRepo_GetByName_Success(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(tagCols).AddRow(1, "test", nil, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, color, created_at FROM document_tags WHERE name")).
		WithArgs("test").WillReturnRows(rows)

	tag, err := repo.GetByName(context.Background(), "test")
	require.NoError(t, err)
	assert.Equal(t, int64(1), tag.ID)
}

func TestTagRepo_GetByName_NotFound(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, color, created_at FROM document_tags WHERE name")).
		WithArgs("nonexistent").WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByName(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "тег не найден")
}

func TestTagRepo_GetAll_Success(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(tagCols).
		AddRow(1, "alpha", nil, now).
		AddRow(2, "beta", nil, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, color, created_at FROM document_tags ORDER BY name")).
		WillReturnRows(rows)

	tags, err := repo.GetAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, tags, 2)
}

func TestTagRepo_GetAll_Error(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, color, created_at FROM document_tags ORDER BY name")).
		WillReturnError(fmt.Errorf("err"))

	_, err := repo.GetAll(context.Background())
	assert.Error(t, err)
}

func TestTagRepo_Search_Success(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(tagCols).AddRow(1, "test", nil, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, color, created_at")).
		WithArgs("tes%", 10).WillReturnRows(rows)

	tags, err := repo.Search(context.Background(), "tes", 10)
	require.NoError(t, err)
	assert.Len(t, tags, 1)
}

func TestTagRepo_Search_DefaultLimit(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, color, created_at")).
		WithArgs("t%", 10).WillReturnRows(sqlmock.NewRows(tagCols))

	_, err := repo.Search(context.Background(), "t", 0)
	require.NoError(t, err)
}

func TestTagRepo_Search_MaxLimit(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, color, created_at")).
		WithArgs("t%", 100).WillReturnRows(sqlmock.NewRows(tagCols))

	_, err := repo.Search(context.Background(), "t", 200)
	require.NoError(t, err)
}

func TestTagRepo_AddTagToDocument(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO document_tag_relations")).
		WithArgs(int64(1), int64(2)).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.AddTagToDocument(context.Background(), 1, 2)
	require.NoError(t, err)
}

func TestTagRepo_RemoveTagFromDocument_Success(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_tag_relations")).
		WithArgs(int64(1), int64(2)).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.RemoveTagFromDocument(context.Background(), 1, 2)
	require.NoError(t, err)
}

func TestTagRepo_RemoveTagFromDocument_NotFound(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_tag_relations")).
		WithArgs(int64(1), int64(2)).WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.RemoveTagFromDocument(context.Background(), 1, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "связь тега с документом не найдена")
}

func TestTagRepo_GetTagsByDocumentID(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(tagCols).AddRow(1, "tag1", nil, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT t.id")).
		WithArgs(int64(1)).WillReturnRows(rows)

	tags, err := repo.GetTagsByDocumentID(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, tags, 1)
}

func TestTagRepo_GetDocumentsByTagID(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT r.document_id")).
		WithArgs(int64(1), 20, 0).WillReturnRows(sqlmock.NewRows([]string{"document_id"}).AddRow(int64(10)).AddRow(int64(20)))

	ids, total, err := repo.GetDocumentsByTagID(context.Background(), 1, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, ids, 2)
}

func TestTagRepo_GetDocumentsByTagID_ClampLimit(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT r.document_id")).
		WithArgs(int64(1), 100, 0).WillReturnRows(sqlmock.NewRows([]string{"document_id"}))

	_, _, err := repo.GetDocumentsByTagID(context.Background(), 1, 200, 0)
	require.NoError(t, err)
}

func TestTagRepo_SetDocumentTags(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_tag_relations WHERE document_id = $1")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO document_tag_relations"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO document_tag_relations")).
		WithArgs(int64(1), int64(10)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO document_tag_relations")).
		WithArgs(int64(1), int64(20)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.SetDocumentTags(context.Background(), 1, []int64{10, 20})
	require.NoError(t, err)
}

func TestTagRepo_SetDocumentTags_EmptyList(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_tag_relations WHERE document_id = $1")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.SetDocumentTags(context.Background(), 1, []int64{})
	require.NoError(t, err)
}

func TestTagRepo_GetTagUsageCount(t *testing.T) {
	repo, mock := newTagRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))

	count, err := repo.GetTagUsageCount(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}
