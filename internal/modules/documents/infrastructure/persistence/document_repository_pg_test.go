package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
)

const testHelloStr = "hello"

func newDocRepoMock(t *testing.T) (*DocumentRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewDocumentRepositoryPG(db), mock
}

var docSelectCols = []string{
	"id", "document_type_id", "category_id", "registration_number", "registration_date",
	"title", "subject", "content", "author_id", "author_department", "author_position",
	"recipient_id", "recipient_department", "recipient_position", "recipient_external",
	"status", "file_name", "file_path", "file_size", "mime_type", "version",
	"parent_document_id", "deadline", "execution_date", "metadata", "is_public", "importance",
	"created_at", "updated_at", "deleted_at",
	"author_name", "recipient_name",
}

func newDocRows() *sqlmock.Rows { return sqlmock.NewRows(docSelectCols) }

func addDocRow(rows *sqlmock.Rows, id int64, title string, ver int, meta []byte) *sqlmock.Rows {
	now := time.Now()
	return rows.AddRow(
		id, int64(1), nil, nil, nil,
		title, nil, nil, int64(10), nil, nil,
		nil, nil, nil, nil,
		"draft", nil, nil, nil, nil, ver,
		nil, nil, nil, meta, false, "",
		now, now, nil, nil, nil,
	)
}

var versionCols = []string{
	"id", "document_id", "version", "title", "subject", "content", "status",
	"file_name", "file_path", "file_size", "mime_type", "storage_key",
	"metadata", "changed_by", "change_description", "created_at",
	"changed_by_name",
}

func newVerRows() *sqlmock.Rows { return sqlmock.NewRows(versionCols) }

func addVerRow(rows *sqlmock.Rows, id, docID int64, ver int, title *string) *sqlmock.Rows {
	return rows.AddRow(
		id, docID, ver, title, nil, nil, nil,
		nil, nil, nil, nil, nil,
		nil, int64(5), nil, time.Now(), nil,
	)
}

func TestNewDocumentRepositoryPG(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewDocumentRepositoryPG(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestDocumentRepositoryPG_Create_Success(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	doc := &entities.Document{
		DocumentTypeID: 1, Title: "Test", AuthorID: 10,
		Status: entities.DocumentStatusDraft, Version: 1,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO documents")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(42)))
	err := repo.Create(context.Background(), doc)
	require.NoError(t, err)
	assert.Equal(t, int64(42), doc.ID)
}

func TestDocumentRepositoryPG_Create_WithMetadata(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	doc := &entities.Document{
		Title: "Test", Metadata: map[string]interface{}{"k": "v"},
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO documents")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	err := repo.Create(context.Background(), doc)
	require.NoError(t, err)
}

func TestDocumentRepositoryPG_Create_Error(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	doc := &entities.Document{Title: "Test", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO documents")).
		WillReturnError(fmt.Errorf("db error"))
	err := repo.Create(context.Background(), doc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create document")
}

func TestDocumentRepositoryPG_Update_Success(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	doc := &entities.Document{ID: 1, Title: "Updated", Status: entities.DocumentStatusDraft}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE documents SET")).WillReturnResult(sqlmock.NewResult(0, 1))
	err := repo.Update(context.Background(), doc)
	require.NoError(t, err)
}

func TestDocumentRepositoryPG_Update_NotFound(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	doc := &entities.Document{ID: 999, Title: "Updated"}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE documents SET")).WillReturnResult(sqlmock.NewResult(0, 0))
	err := repo.Update(context.Background(), doc)
	assert.Contains(t, err.Error(), "document not found")
}

func TestDocumentRepositoryPG_Update_WithMetadata(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	doc := &entities.Document{ID: 1, Title: "Up", Metadata: map[string]interface{}{"k": "v"}}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE documents SET")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Update(context.Background(), doc))
}

func TestDocumentRepositoryPG_Update_DBError(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	doc := &entities.Document{ID: 1, Title: "Up"}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE documents SET")).WillReturnError(fmt.Errorf("db error"))
	err := repo.Update(context.Background(), doc)
	assert.Contains(t, err.Error(), "failed to update document")
}

func TestDocumentRepositoryPG_GetByID_Success(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	rows := addDocRow(newDocRows(), 1, "Test Doc", 1, nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT d.id")).WithArgs(int64(1)).WillReturnRows(rows)
	doc, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Test Doc", doc.Title)
}

func TestDocumentRepositoryPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT d.id")).WithArgs(int64(999)).WillReturnError(sql.ErrNoRows)
	_, err := repo.GetByID(context.Background(), 999)
	assert.Contains(t, err.Error(), "document not found")
}

func TestDocumentRepositoryPG_GetByID_WithMetadata(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	metaJSON, _ := json.Marshal(map[string]interface{}{"key": "val"})
	rows := addDocRow(newDocRows(), 1, "Test", 1, metaJSON)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT d.id")).WithArgs(int64(1)).WillReturnRows(rows)
	doc, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "val", doc.Metadata["key"])
}

func TestDocumentRepositoryPG_GetByID_DBError(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT d.id")).WithArgs(int64(1)).WillReturnError(fmt.Errorf("conn"))
	_, err := repo.GetByID(context.Background(), 1)
	assert.Contains(t, err.Error(), "failed to get document")
}

func TestDocumentRepositoryPG_Delete(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM documents WHERE id = $1")).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Delete(context.Background(), 1))
}

func TestDocumentRepositoryPG_SoftDelete_Success(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE documents SET deleted_at")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.SoftDelete(context.Background(), 1))
}

func TestDocumentRepositoryPG_SoftDelete_NotFound(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE documents SET deleted_at")).WillReturnResult(sqlmock.NewResult(0, 0))
	assert.Contains(t, repo.SoftDelete(context.Background(), 999).Error(), "document not found")
}

func TestDocumentRepositoryPG_SoftDelete_DBError(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE documents SET deleted_at")).WillReturnError(fmt.Errorf("db"))
	assert.Contains(t, repo.SoftDelete(context.Background(), 1).Error(), "failed to soft delete")
}

func TestDocumentRepositoryPG_List_Empty(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery("SELECT d.id").WillReturnRows(newDocRows())
	docs, total, err := repo.List(context.Background(), repositories.DocumentFilter{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, docs)
}

func TestDocumentRepositoryPG_List_WithAllFilters(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	authorID := int64(5)
	recipientID := int64(6)
	docTypeID := int64(1)
	categoryID := int64(2)
	status := entities.DocumentStatusDraft
	importance := entities.DocumentImportance("high")
	isPublic := true
	search := "test"

	filter := repositories.DocumentFilter{
		AuthorID: &authorID, RecipientID: &recipientID,
		DocumentTypeID: &docTypeID, CategoryID: &categoryID,
		Status: &status, Importance: &importance,
		IsPublic: &isPublic, SearchQuery: &search,
		CurrentUserID: 10, CurrentUserRole: "user",
		OrderBy: "title ASC", IncludeDeleted: true,
		Limit: 10, Offset: 0,
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT d.id").WillReturnRows(addDocRow(newDocRows(), 1, "Found", 1, nil))
	docs, total, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, docs, 1)
}

func TestDocumentRepositoryPG_List_AdminBypass(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery("SELECT d.id").WillReturnRows(newDocRows())
	_, _, err := repo.List(context.Background(), repositories.DocumentFilter{CurrentUserID: 10, CurrentUserRole: "admin", Limit: 10})
	require.NoError(t, err)
}

func TestDocumentRepositoryPG_List_CountError(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnError(fmt.Errorf("count error"))
	_, _, err := repo.List(context.Background(), repositories.DocumentFilter{Limit: 10})
	assert.Contains(t, err.Error(), "failed to count documents")
}

func TestDocumentRepositoryPG_List_QueryError(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT d.id").WillReturnError(fmt.Errorf("query error"))
	_, _, err := repo.List(context.Background(), repositories.DocumentFilter{Limit: 10})
	assert.Contains(t, err.Error(), "failed to list documents")
}

func TestDocumentRepositoryPG_GetByAuthorID(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery("SELECT d.id").WillReturnRows(newDocRows())
	docs, err := repo.GetByAuthorID(context.Background(), 5, 10, 0)
	require.NoError(t, err)
	assert.Empty(t, docs)
}

func TestDocumentRepositoryPG_GetByStatus(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery("SELECT d.id").WillReturnRows(newDocRows())
	docs, err := repo.GetByStatus(context.Background(), entities.DocumentStatusDraft, 10, 0)
	require.NoError(t, err)
	assert.Empty(t, docs)
}

func TestDocumentRepositoryPG_CreateVersion_Success(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	title := "Title"
	v := &entities.DocumentVersion{DocumentID: 1, Version: 1, Title: &title}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO document_versions")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(10)))
	require.NoError(t, repo.CreateVersion(context.Background(), v))
	assert.Equal(t, int64(10), v.ID)
}

func TestDocumentRepositoryPG_CreateVersion_WithMetadata(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	v := &entities.DocumentVersion{DocumentID: 1, Version: 1, Metadata: map[string]interface{}{"k": "v"}}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO document_versions")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(10)))
	require.NoError(t, repo.CreateVersion(context.Background(), v))
}

func TestDocumentRepositoryPG_GetVersions_Success(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	title := "Title v1"
	rows := addVerRow(newVerRows(), 1, 1, 1, &title)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT dv.id")).WithArgs(int64(1)).WillReturnRows(rows)
	versions, err := repo.GetVersions(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, versions, 1)
}

func TestDocumentRepositoryPG_GetVersions_Error(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT dv.id")).WithArgs(int64(1)).WillReturnError(fmt.Errorf("err"))
	_, err := repo.GetVersions(context.Background(), 1)
	assert.Error(t, err)
}

func TestDocumentRepositoryPG_GetVersion_Success(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	title := "Title"
	rows := addVerRow(newVerRows(), 1, 1, 1, &title)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT dv.id")).WithArgs(int64(1), 1).WillReturnRows(rows)
	v, err := repo.GetVersion(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.Equal(t, &title, v.Title)
}

func TestDocumentRepositoryPG_GetVersion_NotFound(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT dv.id")).WithArgs(int64(1), 99).WillReturnError(sql.ErrNoRows)
	_, err := repo.GetVersion(context.Background(), 1, 99)
	assert.Contains(t, err.Error(), "version not found")
}

func TestDocumentRepositoryPG_GetVersion_DBError(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT dv.id")).WithArgs(int64(1), 1).WillReturnError(fmt.Errorf("conn"))
	_, err := repo.GetVersion(context.Background(), 1, 1)
	assert.Contains(t, err.Error(), "failed to get version")
}

func TestDocumentRepositoryPG_GetLatestVersion_Success(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	title := "Latest"
	rows := addVerRow(newVerRows(), 1, 1, 2, &title)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT dv.id")).WithArgs(int64(1)).WillReturnRows(rows)
	v, err := repo.GetLatestVersion(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, &title, v.Title)
}

func TestDocumentRepositoryPG_GetLatestVersion_NoVersions(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT dv.id")).WithArgs(int64(1)).WillReturnError(sql.ErrNoRows)
	_, err := repo.GetLatestVersion(context.Background(), 1)
	assert.Contains(t, err.Error(), "no versions found")
}

func TestDocumentRepositoryPG_AddHistory(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	uid := int64(5)
	h := &entities.DocumentHistory{DocumentID: 1, UserID: &uid, Action: "create"}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO document_history")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	require.NoError(t, repo.AddHistory(context.Background(), h))
	assert.Equal(t, int64(1), h.ID)
}

func TestDocumentRepositoryPG_AddHistory_WithDetails(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	h := &entities.DocumentHistory{DocumentID: 1, Action: "update", Details: map[string]interface{}{"f": "t"}}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO document_history")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	require.NoError(t, repo.AddHistory(context.Background(), h))
}

func TestDocumentRepositoryPG_GetHistory(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	uid := int64(5)
	rows := sqlmock.NewRows([]string{"id", "document_id", "user_id", "action", "details", "ip_address", "user_agent", "created_at"}).
		AddRow(1, 1, &uid, "create", nil, nil, nil, time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, document_id")).WithArgs(int64(1)).WillReturnRows(rows)
	history, err := repo.GetHistory(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, history, 1)
}

func TestDocumentRepositoryPG_GetHistory_WithDetails(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	dj, _ := json.Marshal(map[string]interface{}{"field": "title"})
	rows := sqlmock.NewRows([]string{"id", "document_id", "user_id", "action", "details", "ip_address", "user_agent", "created_at"}).
		AddRow(1, 1, nil, "update", dj, nil, nil, time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, document_id")).WithArgs(int64(1)).WillReturnRows(rows)
	history, err := repo.GetHistory(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "title", history[0].Details["field"])
}

func TestDocumentRepositoryPG_CreateVersionDiff(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	diff := &entities.DocumentVersionDiff{
		DocumentID: 1, FromVersion: 1, ToVersion: 2,
		ChangedFields: []string{"title"},
		DiffData:      map[string]interface{}{"title": "changed"},
	}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO document_version_diffs")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	require.NoError(t, repo.CreateVersionDiff(context.Background(), diff))
	assert.Equal(t, int64(1), diff.ID)
}

func TestDocumentRepositoryPG_GetVersionDiff_Found(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	dj, _ := json.Marshal(map[string]interface{}{"title": "changed"})
	rows := sqlmock.NewRows([]string{"id", "document_id", "from_version", "to_version", "changed_fields", "diff_data", "created_at"}).
		AddRow(1, 1, 1, 2, pq.Array([]string{"title"}), dj, time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, document_id")).WithArgs(int64(1), 1, 2).WillReturnRows(rows)
	diff, err := repo.GetVersionDiff(context.Background(), 1, 1, 2)
	require.NoError(t, err)
	assert.NotNil(t, diff)
	assert.Equal(t, []string{"title"}, diff.ChangedFields)
}

func TestDocumentRepositoryPG_GetVersionDiff_NotFound(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, document_id")).WithArgs(int64(1), 1, 2).WillReturnError(sql.ErrNoRows)
	diff, err := repo.GetVersionDiff(context.Background(), 1, 1, 2)
	require.NoError(t, err)
	assert.Nil(t, diff)
}

func TestDocumentRepositoryPG_GetVersionDiff_DBError(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, document_id")).WithArgs(int64(1), 1, 2).WillReturnError(fmt.Errorf("err"))
	_, err := repo.GetVersionDiff(context.Background(), 1, 1, 2)
	assert.Error(t, err)
}

func TestDocumentRepositoryPG_Search_EmptyQuery(t *testing.T) {
	repo, _ := newDocRepoMock(t)
	_, _, err := repo.Search(context.Background(), repositories.SearchFilter{Query: ""})
	assert.Contains(t, err.Error(), "search query cannot be empty")
}

func TestDocumentRepositoryPG_Search_SpecialCharsOnly(t *testing.T) {
	repo, _ := newDocRepoMock(t)
	results, total, err := repo.Search(context.Background(), repositories.SearchFilter{Query: "&|!()"})
	require.NoError(t, err)
	assert.Empty(t, results)
	assert.Equal(t, int64(0), total)
}

func TestDocumentRepositoryPG_Search_ZeroResults(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	results, total, err := repo.Search(context.Background(), repositories.SearchFilter{Query: "test", Limit: 10})
	require.NoError(t, err)
	assert.Empty(t, results)
	assert.Equal(t, int64(0), total)
}

func TestDocumentRepositoryPG_Search_WithFilters(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	dtID := int64(1)
	cID := int64(2)
	aID := int64(3)
	status := entities.DocumentStatus("draft")
	imp := entities.DocumentImportance("high")
	from := "2024-01-01"
	to := "2024-12-31"
	filter := repositories.SearchFilter{
		Query: "test", DocumentTypeID: &dtID, CategoryID: &cID,
		AuthorID: &aID, Status: &status, Importance: &imp,
		FromDate: &from, ToDate: &to,
		CurrentUserID: 10, CurrentUserRole: "user",
		Limit: 10,
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	results, total, err := repo.Search(context.Background(), filter)
	require.NoError(t, err)
	assert.Empty(t, results)
	assert.Equal(t, int64(0), total)
}

func TestDocumentRepositoryPG_DeleteVersion_CannotDeleteCurrent(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT d.id")).WithArgs(int64(1)).WillReturnRows(addDocRow(newDocRows(), 1, "Test", 3, nil))
	err := repo.DeleteVersion(context.Background(), 1, 3)
	assert.Contains(t, err.Error(), "cannot delete current version")
}

func TestDocumentRepositoryPG_DeleteVersion_Success(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT d.id")).WithArgs(int64(1)).WillReturnRows(addDocRow(newDocRows(), 1, "Test", 3, nil))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_versions")).WithArgs(int64(1), 1).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.DeleteVersion(context.Background(), 1, 1))
}

func TestDocumentRepositoryPG_DeleteVersion_VersionNotFound(t *testing.T) {
	repo, mock := newDocRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT d.id")).WithArgs(int64(1)).WillReturnRows(addDocRow(newDocRows(), 1, "Test", 3, nil))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_versions")).WithArgs(int64(1), 1).WillReturnResult(sqlmock.NewResult(0, 0))
	assert.Contains(t, repo.DeleteVersion(context.Background(), 1, 1).Error(), "version not found")
}

// Test helper functions
func TestStrPtrEqual(t *testing.T) {
	a, b, c := testHelloStr, testHelloStr, "world"
	assert.True(t, strPtrEqual(nil, nil))
	assert.False(t, strPtrEqual(&a, nil))
	assert.False(t, strPtrEqual(nil, &a))
	assert.True(t, strPtrEqual(&a, &b))
	assert.False(t, strPtrEqual(&a, &c))
}

func TestInt64PtrEqual(t *testing.T) {
	a, b, c := int64(1), int64(1), int64(2)
	assert.True(t, int64PtrEqual(nil, nil))
	assert.False(t, int64PtrEqual(&a, nil))
	assert.False(t, int64PtrEqual(nil, &a))
	assert.True(t, int64PtrEqual(&a, &b))
	assert.False(t, int64PtrEqual(&a, &c))
}

func TestPtrToStr(t *testing.T) {
	s := testHelloStr
	assert.Equal(t, "", ptrToStr(nil))
	assert.Equal(t, testHelloStr, ptrToStr(&s))
}

func TestPtrToInt64(t *testing.T) {
	i := int64(42)
	assert.Equal(t, int64(0), ptrToInt64(nil))
	assert.Equal(t, int64(42), ptrToInt64(&i))
}

func TestSanitizeForTsquery(t *testing.T) {
	tests := []struct {
		name, input, expected string
	}{
		{"single word", "со", "со:*"},
		{"single word with spaces", "  собака  ", "собака:*"},
		{"multiple words", "собака документ", "собака:* & документ:*"},
		{"special characters escaped", "test&query|here", "test:* & query:* & here:*"},
		{"parentheses removed", "test(query)", "test:* & query:*"},
		{"colon and asterisk removed", "test:*query", "test:* & query:*"},
		{"empty string", "", ""},
		{"only spaces", "   ", ""},
		{"only special characters", "&|!()", ""},
		{"russian text", "Документ Приказ", "Документ:* & Приказ:*"},
		{"mixed alphanumeric", "документ123 test456", "документ123:* & test456:*"},
		{"backslash removed", "test\\query", "test:* & query:*"},
		{"angle brackets removed", "test<query>", "test:* & query:*"},
		{"single quote removed", "test'query", "test:* & query:*"},
		{"exclamation removed", "!test", "test:*"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, sanitizeForTsquery(tt.input))
		})
	}
}

func TestDocumentRepositoryPG_RestoreVersion_GetVersionError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewDocumentRepositoryPG(db)

	mock.ExpectQuery("FROM document_versions").
		WillReturnError(fmt.Errorf("version error"))
	restoreErr := repo.RestoreVersion(context.Background(), 1, 2, 10)
	assert.Error(t, restoreErr)
	assert.Contains(t, restoreErr.Error(), "get version to restore")
}

func TestDocumentRepositoryPG_RestoreVersion_GetDocError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewDocumentRepositoryPG(db)

	title := "Test"
	verRows := addVerRow(newVerRows(), 1, 1, 2, &title)
	mock.ExpectQuery("FROM document_versions").WillReturnRows(verRows)
	mock.ExpectQuery("FROM documents").WillReturnError(fmt.Errorf("doc error"))

	restoreErr := repo.RestoreVersion(context.Background(), 1, 2, 10)
	assert.Error(t, restoreErr)
	assert.Contains(t, restoreErr.Error(), "get current document")
}

func TestDocumentRepositoryPG_CompareVersions_CachedDiffReturned(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewDocumentRepositoryPG(db)

	now := time.Now()
	diffCols := []string{"id", "document_id", "from_version", "to_version", "changed_fields", "diff_data", "created_at"}
	diffData, _ := json.Marshal(map[string]interface{}{"title": "changed"})
	mock.ExpectQuery("FROM document_version_diffs").
		WillReturnRows(sqlmock.NewRows(diffCols).AddRow(
			int64(1), int64(1), 1, 2, pq.Array([]string{"title"}), diffData, now,
		))

	diff, err := repo.CompareVersions(context.Background(), 1, 1, 2)
	require.NoError(t, err)
	assert.NotNil(t, diff)
	assert.Contains(t, diff.ChangedFields, "title")
}

func TestDocumentRepositoryPG_CompareVersions_NoCacheComputesDiff(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewDocumentRepositoryPG(db)

	// No cached diff
	mock.ExpectQuery("FROM document_version_diffs").WillReturnError(sql.ErrNoRows)

	// GetVersion(from)
	title1 := "Old Title"
	mock.ExpectQuery("FROM document_versions").WillReturnRows(addVerRow(newVerRows(), 1, 1, 1, &title1))

	// GetVersion(to)
	title2 := "New Title"
	mock.ExpectQuery("FROM document_versions").WillReturnRows(addVerRow(newVerRows(), 2, 1, 2, &title2))

	// CreateVersionDiff
	mock.ExpectQuery("INSERT INTO document_version_diffs").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	diff, err := repo.CompareVersions(context.Background(), 1, 1, 2)
	require.NoError(t, err)
	assert.NotNil(t, diff)
	assert.Contains(t, diff.ChangedFields, "title")
}

func TestDocumentRepositoryPG_CompareVersions_CacheLookupError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewDocumentRepositoryPG(db)

	mock.ExpectQuery("FROM document_version_diffs").WillReturnError(fmt.Errorf("cache error"))

	diff, err := repo.CompareVersions(context.Background(), 1, 1, 2)
	assert.Error(t, err)
	assert.Nil(t, diff)
}

func TestDocumentRepositoryPG_CompareVersions_GetFromVersionError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewDocumentRepositoryPG(db)

	mock.ExpectQuery("FROM document_version_diffs").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("FROM document_versions").WillReturnError(fmt.Errorf("version not found"))

	diff, err := repo.CompareVersions(context.Background(), 1, 1, 2)
	assert.Error(t, err)
	assert.Nil(t, diff)
}

func TestDocumentRepositoryPG_CompareVersions_GetToVersionError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewDocumentRepositoryPG(db)

	mock.ExpectQuery("FROM document_version_diffs").WillReturnError(sql.ErrNoRows)
	title := "V1"
	mock.ExpectQuery("FROM document_versions").WillReturnRows(addVerRow(newVerRows(), 1, 1, 1, &title))
	mock.ExpectQuery("FROM document_versions").WillReturnError(fmt.Errorf("to version error"))

	diff, err := repo.CompareVersions(context.Background(), 1, 1, 2)
	assert.Error(t, err)
	assert.Nil(t, diff)
}

func TestDocumentRepositoryPG_CompareVersions_CacheSaveFail(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewDocumentRepositoryPG(db)

	mock.ExpectQuery("FROM document_version_diffs").WillReturnError(sql.ErrNoRows)
	title := "Same"
	mock.ExpectQuery("FROM document_versions").WillReturnRows(addVerRow(newVerRows(), 1, 1, 1, &title))
	mock.ExpectQuery("FROM document_versions").WillReturnRows(addVerRow(newVerRows(), 2, 1, 2, &title))
	// CreateVersionDiff fails (cache save)
	mock.ExpectQuery("INSERT INTO document_version_diffs").WillReturnError(fmt.Errorf("cache error"))

	diff, err := repo.CompareVersions(context.Background(), 1, 1, 2)
	require.NoError(t, err) // cache failure is non-fatal
	assert.NotNil(t, diff)
}

func TestDocumentRepositoryPG_Search_Error(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewDocumentRepositoryPG(db)

	mock.ExpectQuery("ts_rank").WillReturnError(fmt.Errorf("search error"))

	filter := repositories.SearchFilter{Query: "test", Limit: 20}
	_, _, err = repo.Search(context.Background(), filter)
	assert.Error(t, err)
}

func TestDocumentRepositoryPG_Search_CountError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewDocumentRepositoryPG(db)

	searchCols := []string{
		"id", "title", "subject", "content", "author_id", "status",
		"file_name", "file_size", "mime_type", "version", "is_public",
		"importance", "category_id", "document_type_id", "created_at", "updated_at",
		"rank", "headline",
	}
	mock.ExpectQuery("ts_rank").WillReturnRows(sqlmock.NewRows(searchCols))
	mock.ExpectQuery("COUNT").WillReturnError(fmt.Errorf("count error"))

	filter := repositories.SearchFilter{Query: "test", Limit: 20}
	_, _, err = repo.Search(context.Background(), filter)
	assert.Error(t, err)
}

func TestDocumentRepositoryPG_GetVersions_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewDocumentRepositoryPG(db)

	mock.ExpectQuery("FROM document_versions").WillReturnError(fmt.Errorf("db error"))
	_, err = repo.GetVersions(context.Background(), 1)
	assert.Error(t, err)
}

func TestDocumentRepositoryPG_GetVersion_Error(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewDocumentRepositoryPG(db)

	mock.ExpectQuery("FROM document_versions").WillReturnError(fmt.Errorf("db error"))
	_, err = repo.GetVersion(context.Background(), 1, 1)
	assert.Error(t, err)
}

func TestDocumentRepositoryPG_GetHistory_Error(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewDocumentRepositoryPG(db)

	mock.ExpectQuery("FROM document_history").WillReturnError(fmt.Errorf("db error"))
	_, err = repo.GetHistory(context.Background(), 1)
	assert.Error(t, err)
}

func TestDocumentRepositoryPG_DeleteVersion_Error(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewDocumentRepositoryPG(db)

	mock.ExpectQuery("FROM documents").WillReturnError(fmt.Errorf("db error"))
	err = repo.DeleteVersion(context.Background(), 1, 1)
	assert.Error(t, err)
}

func TestDocumentRepositoryPG_GetLatestVersion_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := NewDocumentRepositoryPG(db)

	mock.ExpectQuery("FROM document_versions").WillReturnError(sql.ErrNoRows)

	v, err := repo.GetLatestVersion(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, v)
}
