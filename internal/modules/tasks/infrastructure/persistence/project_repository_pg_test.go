package persistence

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
)

func newProjectRepoMock(t *testing.T) (*ProjectRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewProjectRepositoryPG(db), mock
}

var projectCols = []string{
	"id", "name", "description", "owner_id", "status",
	"start_date", "end_date", "created_at", "updated_at",
}

func newProjectRows() *sqlmock.Rows { return sqlmock.NewRows(projectCols) }

func addProjectRow(rows *sqlmock.Rows, id int64, name string) *sqlmock.Rows {
	now := time.Now()
	desc := "test desc"
	return rows.AddRow(id, name, &desc, int64(1), domain.ProjectStatusActive, &now, &now, now, now)
}

// --- Create ---

func TestProjectRepositoryPG_Create_Success(t *testing.T) {
	repo, mock := newProjectRepoMock(t)
	project := entities.NewProject("Test Project", 1)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO projects")).
		WithArgs(project.Name, project.Description, project.OwnerID, project.Status,
			project.StartDate, project.EndDate, project.CreatedAt, project.UpdatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.Create(context.Background(), project)
	require.NoError(t, err)
	assert.Equal(t, int64(1), project.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepositoryPG_Create_Error(t *testing.T) {
	repo, mock := newProjectRepoMock(t)
	project := entities.NewProject("Test", 1)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO projects")).
		WillReturnError(sql.ErrConnDone)

	err := repo.Create(context.Background(), project)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Save ---

func TestProjectRepositoryPG_Save_Success(t *testing.T) {
	repo, mock := newProjectRepoMock(t)
	project := entities.NewProject("Updated", 1)
	project.ID = 5

	mock.ExpectExec(regexp.QuoteMeta("UPDATE projects SET")).
		WithArgs(project.Name, project.Description, project.Status,
			project.StartDate, project.EndDate, project.UpdatedAt, project.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Save(context.Background(), project)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepositoryPG_Save_Error(t *testing.T) {
	repo, mock := newProjectRepoMock(t)
	project := entities.NewProject("err", 1)
	project.ID = 5

	mock.ExpectExec(regexp.QuoteMeta("UPDATE projects SET")).
		WillReturnError(sql.ErrConnDone)

	err := repo.Save(context.Background(), project)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByID ---

func TestProjectRepositoryPG_GetByID_Success(t *testing.T) {
	repo, mock := newProjectRepoMock(t)
	rows := addProjectRow(newProjectRows(), 1, "Project1")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, owner_id, status, start_date, end_date, created_at, updated_at")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	project, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, project)
	assert.Equal(t, "Project1", project.Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepositoryPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newProjectRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	project, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, project)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepositoryPG_GetByID_Error(t *testing.T) {
	repo, mock := newProjectRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description")).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	project, err := repo.GetByID(context.Background(), 1)
	require.Error(t, err)
	assert.Nil(t, project)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Delete ---

func TestProjectRepositoryPG_Delete_Success(t *testing.T) {
	repo, mock := newProjectRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM projects WHERE id = $1")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepositoryPG_Delete_Error(t *testing.T) {
	repo, mock := newProjectRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM projects WHERE id = $1")).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	err := repo.Delete(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- List ---

func TestProjectRepositoryPG_List_NoFilter(t *testing.T) {
	repo, mock := newProjectRepoMock(t)
	rows := addProjectRow(newProjectRows(), 1, "P1")
	addProjectRow(rows, 2, "P2")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description, owner_id, status, start_date, end_date, created_at, updated_at")).
		WillReturnRows(rows)

	projects, err := repo.List(context.Background(), repositories.ProjectFilter{}, 10, 0)
	require.NoError(t, err)
	assert.Len(t, projects, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepositoryPG_List_WithFilter(t *testing.T) {
	repo, mock := newProjectRepoMock(t)
	ownerID := int64(5)
	status := domain.ProjectStatusActive
	search := "test"
	filter := repositories.ProjectFilter{
		OwnerID: &ownerID,
		Status:  &status,
		Search:  &search,
	}

	rows := addProjectRow(newProjectRows(), 1, "test project")

	mock.ExpectQuery("SELECT id, name, description").
		WillReturnRows(rows)

	projects, err := repo.List(context.Background(), filter, 10, 0)
	require.NoError(t, err)
	assert.Len(t, projects, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepositoryPG_List_Error(t *testing.T) {
	repo, mock := newProjectRepoMock(t)

	mock.ExpectQuery("SELECT id, name, description").
		WillReturnError(sql.ErrConnDone)

	projects, err := repo.List(context.Background(), repositories.ProjectFilter{}, 10, 0)
	require.Error(t, err)
	assert.Nil(t, projects)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepositoryPG_List_ScanError(t *testing.T) {
	repo, mock := newProjectRepoMock(t)
	rows := sqlmock.NewRows(projectCols).AddRow(1, "P", nil, "not-a-number", "active", nil, nil, time.Now(), time.Now())

	mock.ExpectQuery("SELECT id, name, description").
		WillReturnRows(rows)

	projects, err := repo.List(context.Background(), repositories.ProjectFilter{}, 10, 0)
	require.Error(t, err)
	assert.Nil(t, projects)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Count ---

func TestProjectRepositoryPG_Count_Success(t *testing.T) {
	repo, mock := newProjectRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM projects")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))

	count, err := repo.Count(context.Background(), repositories.ProjectFilter{})
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepositoryPG_Count_WithFilter(t *testing.T) {
	repo, mock := newProjectRepoMock(t)
	ownerID := int64(3)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM projects")).
		WithArgs(ownerID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))

	count, err := repo.Count(context.Background(), repositories.ProjectFilter{OwnerID: &ownerID})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProjectRepositoryPG_Count_Error(t *testing.T) {
	repo, mock := newProjectRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.Count(context.Background(), repositories.ProjectFilter{})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByOwner ---

func TestProjectRepositoryPG_GetByOwner_Success(t *testing.T) {
	repo, mock := newProjectRepoMock(t)
	rows := addProjectRow(newProjectRows(), 1, "P1")

	mock.ExpectQuery("SELECT id, name, description").
		WillReturnRows(rows)

	projects, err := repo.GetByOwner(context.Background(), 5, 10, 0)
	require.NoError(t, err)
	assert.Len(t, projects, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByStatus ---

func TestProjectRepositoryPG_GetByStatus_Success(t *testing.T) {
	repo, mock := newProjectRepoMock(t)
	rows := addProjectRow(newProjectRows(), 1, "P1")

	mock.ExpectQuery("SELECT id, name, description").
		WillReturnRows(rows)

	projects, err := repo.GetByStatus(context.Background(), domain.ProjectStatusActive, 10, 0)
	require.NoError(t, err)
	assert.Len(t, projects, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- buildListQuery coverage ---

func TestProjectRepositoryPG_List_NoLimit(t *testing.T) {
	repo, mock := newProjectRepoMock(t)
	rows := addProjectRow(newProjectRows(), 1, "P1")

	mock.ExpectQuery("SELECT id, name, description").
		WillReturnRows(rows)

	projects, err := repo.List(context.Background(), repositories.ProjectFilter{}, 0, 0)
	require.NoError(t, err)
	assert.Len(t, projects, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}
