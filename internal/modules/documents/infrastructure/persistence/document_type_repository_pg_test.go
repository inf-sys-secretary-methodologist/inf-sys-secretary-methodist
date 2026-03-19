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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

func newTypeRepoMock(t *testing.T) (*DocumentTypeRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewDocumentTypeRepositoryPG(db), mock
}

var typeCols = []string{
	"id", "name", "code", "description", "template_path", "template_content", "template_variables",
	"requires_approval", "requires_registration", "numbering_pattern", "retention_period",
	"created_at", "updated_at",
}

func addTypeRow(rows *sqlmock.Rows, id int64, name, code string, tvJSON []byte) *sqlmock.Rows {
	now := time.Now()
	return rows.AddRow(id, name, code, nil, nil, nil, tvJSON, false, false, nil, nil, now, now)
}

func TestNewDocumentTypeRepositoryPG(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	assert.NotNil(t, NewDocumentTypeRepositoryPG(db))
}

func TestTypeRepo_GetAll_Success(t *testing.T) {
	repo, mock := newTypeRepoMock(t)
	rows := sqlmock.NewRows(typeCols)
	addTypeRow(rows, 1, "Type1", "T1", nil)
	addTypeRow(rows, 2, "Type2", "T2", nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).WillReturnRows(rows)
	types, err := repo.GetAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, types, 2)
}

func TestTypeRepo_GetAll_WithTemplateVariables(t *testing.T) {
	repo, mock := newTypeRepoMock(t)
	tvJSON, _ := json.Marshal([]entities.TemplateVariable{{Name: "var1", Label: "Var 1", Type: "string"}})
	rows := sqlmock.NewRows(typeCols)
	addTypeRow(rows, 1, "Type1", "T1", tvJSON)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).WillReturnRows(rows)
	types, err := repo.GetAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, types[0].TemplateVariables, 1)
}

func TestTypeRepo_GetAll_Error(t *testing.T) {
	repo, mock := newTypeRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).WillReturnError(fmt.Errorf("err"))
	_, err := repo.GetAll(context.Background())
	assert.Error(t, err)
}

func TestTypeRepo_GetByID_Success(t *testing.T) {
	repo, mock := newTypeRepoMock(t)
	rows := sqlmock.NewRows(typeCols)
	addTypeRow(rows, 1, "Type1", "T1", nil)
	mock.ExpectQuery(regexp.QuoteMeta("FROM document_types WHERE id")).WithArgs(int64(1)).WillReturnRows(rows)
	dt, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Type1", dt.Name)
}

func TestTypeRepo_GetByID_NotFound(t *testing.T) {
	repo, mock := newTypeRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM document_types WHERE id")).WithArgs(int64(999)).WillReturnError(sql.ErrNoRows)
	_, err := repo.GetByID(context.Background(), 999)
	assert.Contains(t, err.Error(), "document type not found")
}

func TestTypeRepo_GetByCode_Success(t *testing.T) {
	repo, mock := newTypeRepoMock(t)
	rows := sqlmock.NewRows(typeCols)
	addTypeRow(rows, 1, "Type1", "T1", nil)
	mock.ExpectQuery(regexp.QuoteMeta("FROM document_types WHERE code")).WithArgs("T1").WillReturnRows(rows)
	dt, err := repo.GetByCode(context.Background(), "T1")
	require.NoError(t, err)
	assert.Equal(t, "T1", dt.Code)
}

func TestTypeRepo_GetByCode_NotFound(t *testing.T) {
	repo, mock := newTypeRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM document_types WHERE code")).WithArgs("XX").WillReturnError(sql.ErrNoRows)
	_, err := repo.GetByCode(context.Background(), "XX")
	assert.Contains(t, err.Error(), "document type not found")
}

func TestTypeRepo_UpdateTemplate_Success(t *testing.T) {
	repo, mock := newTypeRepoMock(t)
	content := "template content"
	mock.ExpectExec(regexp.QuoteMeta("UPDATE document_types")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UpdateTemplate(context.Background(), 1, &content, nil))
}

func TestTypeRepo_UpdateTemplate_NotFound(t *testing.T) {
	repo, mock := newTypeRepoMock(t)
	content := "content"
	mock.ExpectExec(regexp.QuoteMeta("UPDATE document_types")).WillReturnResult(sqlmock.NewResult(0, 0))
	assert.Contains(t, repo.UpdateTemplate(context.Background(), 999, &content, nil).Error(), "document type not found")
}

func TestTypeRepo_UpdateTemplate_WithVariables(t *testing.T) {
	repo, mock := newTypeRepoMock(t)
	content := "content"
	vars := []entities.TemplateVariable{{Name: "var1", Label: "Var 1", Type: "string"}}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE document_types")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UpdateTemplate(context.Background(), 1, &content, vars))
}

func TestTypeRepo_GetAllWithTemplates(t *testing.T) {
	repo, mock := newTypeRepoMock(t)
	rows := sqlmock.NewRows(typeCols)
	addTypeRow(rows, 1, "Type1", "T1", nil)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE template_content IS NOT NULL")).WillReturnRows(rows)
	types, err := repo.GetAllWithTemplates(context.Background())
	require.NoError(t, err)
	assert.Len(t, types, 1)
}

func TestTypeRepo_GetAllWithTemplates_Error(t *testing.T) {
	repo, mock := newTypeRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE template_content IS NOT NULL")).WillReturnError(fmt.Errorf("err"))
	_, err := repo.GetAllWithTemplates(context.Background())
	assert.Error(t, err)
}

func TestTemplateRepositoryAdapter_GetAll(t *testing.T) {
	repo, mock := newTypeRepoMock(t)
	adapter := NewTemplateRepositoryAdapter(repo)
	rows := sqlmock.NewRows(typeCols)
	addTypeRow(rows, 1, "Type1", "T1", nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).WillReturnRows(rows)
	types, err := adapter.GetAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, types, 1)
	assert.Equal(t, "Type1", types[0].Name)
}

func TestTemplateRepositoryAdapter_GetByID(t *testing.T) {
	repo, mock := newTypeRepoMock(t)
	adapter := NewTemplateRepositoryAdapter(repo)
	rows := sqlmock.NewRows(typeCols)
	addTypeRow(rows, 1, "Type1", "T1", nil)
	mock.ExpectQuery(regexp.QuoteMeta("FROM document_types WHERE id")).WithArgs(int64(1)).WillReturnRows(rows)
	dt, err := adapter.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Type1", dt.Name)
}

func TestTemplateRepositoryAdapter_UpdateTemplate(t *testing.T) {
	repo, mock := newTypeRepoMock(t)
	adapter := NewTemplateRepositoryAdapter(repo)
	content := "content"
	mock.ExpectExec(regexp.QuoteMeta("UPDATE document_types")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, adapter.UpdateTemplate(context.Background(), 1, &content, nil))
}

// DocumentCategoryRepositoryPG tests
func newCatRepoMock(t *testing.T) (*DocumentCategoryRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewDocumentCategoryRepositoryPG(db), mock
}

var catCols = []string{"id", "name", "description", "parent_id", "created_at", "updated_at"}

func addCatRow(rows *sqlmock.Rows, id int64, name string, parentID *int64) *sqlmock.Rows {
	now := time.Now()
	return rows.AddRow(id, name, nil, parentID, now, now)
}

func TestCatRepo_GetAll(t *testing.T) {
	repo, mock := newCatRepoMock(t)
	rows := sqlmock.NewRows(catCols)
	addCatRow(rows, 1, "Cat1", nil)
	addCatRow(rows, 2, "Cat2", nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description")).WillReturnRows(rows)
	cats, err := repo.GetAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, cats, 2)
}

func TestCatRepo_GetByID_Success(t *testing.T) {
	repo, mock := newCatRepoMock(t)
	rows := sqlmock.NewRows(catCols)
	addCatRow(rows, 1, "Cat1", nil)
	mock.ExpectQuery(regexp.QuoteMeta("FROM document_categories WHERE id")).WithArgs(int64(1)).WillReturnRows(rows)
	cat, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Cat1", cat.Name)
}

func TestCatRepo_GetByID_NotFound(t *testing.T) {
	repo, mock := newCatRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM document_categories WHERE id")).WithArgs(int64(999)).WillReturnError(sql.ErrNoRows)
	_, err := repo.GetByID(context.Background(), 999)
	assert.Contains(t, err.Error(), "document category not found")
}

func TestCatRepo_GetByParentID_NilParent(t *testing.T) {
	repo, mock := newCatRepoMock(t)
	rows := sqlmock.NewRows(catCols)
	addCatRow(rows, 1, "Root", nil)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE parent_id IS NULL")).WillReturnRows(rows)
	cats, err := repo.GetByParentID(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, cats, 1)
}

func TestCatRepo_GetByParentID_WithParent(t *testing.T) {
	repo, mock := newCatRepoMock(t)
	parentID := int64(1)
	rows := sqlmock.NewRows(catCols)
	addCatRow(rows, 2, "Child", &parentID)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE parent_id = $1")).WithArgs(parentID).WillReturnRows(rows)
	cats, err := repo.GetByParentID(context.Background(), &parentID)
	require.NoError(t, err)
	assert.Len(t, cats, 1)
}

func TestCatRepo_Create(t *testing.T) {
	repo, mock := newCatRepoMock(t)
	now := time.Now()
	cat := &entities.DocumentCategory{Name: "New Cat"}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO document_categories")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(1), now, now))
	require.NoError(t, repo.Create(context.Background(), cat))
	assert.Equal(t, int64(1), cat.ID)
}

func TestCatRepo_Update_Success(t *testing.T) {
	repo, mock := newCatRepoMock(t)
	cat := &entities.DocumentCategory{ID: 1, Name: "Updated"}
	mock.ExpectQuery(regexp.QuoteMeta("UPDATE document_categories")).
		WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow(time.Now()))
	require.NoError(t, repo.Update(context.Background(), cat))
}

func TestCatRepo_Update_NotFound(t *testing.T) {
	repo, mock := newCatRepoMock(t)
	cat := &entities.DocumentCategory{ID: 999, Name: "Updated"}
	mock.ExpectQuery(regexp.QuoteMeta("UPDATE document_categories")).WillReturnError(sql.ErrNoRows)
	assert.Contains(t, repo.Update(context.Background(), cat).Error(), "document category not found")
}

func TestCatRepo_Delete_Success(t *testing.T) {
	repo, mock := newCatRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE document_categories SET parent_id = NULL")).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE documents SET category_id = NULL")).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_categories WHERE id = $1")).WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Delete(context.Background(), 1))
}

func TestCatRepo_Delete_NotFound(t *testing.T) {
	repo, mock := newCatRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE document_categories SET parent_id = NULL")).WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE documents SET category_id = NULL")).WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_categories WHERE id = $1")).WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))
	assert.Contains(t, repo.Delete(context.Background(), 999).Error(), "document category not found")
}

func TestCatRepo_GetTree(t *testing.T) {
	repo, mock := newCatRepoMock(t)
	now := time.Now()
	parentID := int64(1)
	rows := sqlmock.NewRows([]string{"id", "name", "description", "parent_id", "created_at", "updated_at", "doc_count"}).
		AddRow(1, "Root", nil, nil, now, now, int64(5)).
		AddRow(2, "Child", nil, &parentID, now, now, int64(3))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT c.id")).WillReturnRows(rows)
	tree, err := repo.GetTree(context.Background())
	require.NoError(t, err)
	assert.Len(t, tree, 1)
	assert.Len(t, tree[0].Children, 1)
}

func TestCatRepo_GetChildren(t *testing.T) {
	repo, mock := newCatRepoMock(t)
	parentID := int64(1)
	rows := sqlmock.NewRows(catCols)
	addCatRow(rows, 2, "Child", &parentID)
	mock.ExpectQuery(regexp.QuoteMeta("FROM document_categories WHERE parent_id = $1")).WithArgs(int64(1)).WillReturnRows(rows)
	children, err := repo.GetChildren(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, children, 1)
}

func TestCatRepo_GetAncestors(t *testing.T) {
	repo, mock := newCatRepoMock(t)
	rows := sqlmock.NewRows(catCols)
	addCatRow(rows, 1, "Root", nil)
	mock.ExpectQuery(regexp.QuoteMeta("WITH RECURSIVE ancestors")).WithArgs(int64(2)).WillReturnRows(rows)
	ancestors, err := repo.GetAncestors(context.Background(), 2)
	require.NoError(t, err)
	assert.Len(t, ancestors, 1)
}

func TestCatRepo_HasChildren_True(t *testing.T) {
	repo, mock := newCatRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	has, err := repo.HasChildren(context.Background(), 1)
	require.NoError(t, err)
	assert.True(t, has)
}

func TestCatRepo_HasChildren_False(t *testing.T) {
	repo, mock := newCatRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	has, err := repo.HasChildren(context.Background(), 1)
	require.NoError(t, err)
	assert.False(t, has)
}

func TestCatRepo_GetDocumentCount_WithoutSubcategories(t *testing.T) {
	repo, mock := newCatRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(10)))
	count, err := repo.GetDocumentCount(context.Background(), 1, false)
	require.NoError(t, err)
	assert.Equal(t, int64(10), count)
}

func TestCatRepo_GetDocumentCount_WithSubcategories(t *testing.T) {
	repo, mock := newCatRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WITH RECURSIVE subcategories")).WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(15)))
	count, err := repo.GetDocumentCount(context.Background(), 1, true)
	require.NoError(t, err)
	assert.Equal(t, int64(15), count)
}
