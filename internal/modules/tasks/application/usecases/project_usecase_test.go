package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
)

const updatedProjectName = "Updated Project"

// --- Error-returning mock project repository ---

type ErrorMockProjectRepository struct {
	MockProjectRepository
	createErr  error
	saveErr    error
	getByIDErr error
	deleteErr  error
	listErr    error
	countErr   error
}

func (m *ErrorMockProjectRepository) Create(ctx context.Context, project *entities.Project) error {
	if m.createErr != nil {
		return m.createErr
	}
	return m.MockProjectRepository.Create(ctx, project)
}

func (m *ErrorMockProjectRepository) Save(ctx context.Context, project *entities.Project) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	return m.MockProjectRepository.Save(ctx, project)
}

func (m *ErrorMockProjectRepository) GetByID(ctx context.Context, id int64) (*entities.Project, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.MockProjectRepository.GetByID(ctx, id)
}

func (m *ErrorMockProjectRepository) Delete(ctx context.Context, id int64) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return m.MockProjectRepository.Delete(ctx, id)
}

func (m *ErrorMockProjectRepository) List(ctx context.Context, f repositories.ProjectFilter, limit, offset int) ([]*entities.Project, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.MockProjectRepository.List(ctx, f, limit, offset)
}

func (m *ErrorMockProjectRepository) Count(ctx context.Context, f repositories.ProjectFilter) (int64, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	return m.MockProjectRepository.Count(ctx, f)
}

func newErrorMockProjectRepo() *ErrorMockProjectRepository {
	return &ErrorMockProjectRepository{
		MockProjectRepository: *NewMockProjectRepository(),
	}
}

func setupProjectUseCase() *ProjectUseCase {
	return NewProjectUseCase(NewMockProjectRepository(), nil)
}

func setupProjectUseCaseWithAudit() *ProjectUseCase {
	return NewProjectUseCase(NewMockProjectRepository(), createTestAuditLogger())
}

func createProject(t *testing.T, uc *ProjectUseCase, name string, ownerID int64) *entities.Project {
	t.Helper()
	project, err := uc.Create(context.Background(), ownerID, dto.CreateProjectInput{Name: name})
	require.NoError(t, err)
	return project
}

// ===================== Create =====================

func TestProjectUseCase_Create(t *testing.T) {
	uc := setupProjectUseCase()

	project, err := uc.Create(context.Background(), 1, dto.CreateProjectInput{
		Name:        "Test Project",
		Description: strPtr("Test Description"),
	})

	require.NoError(t, err)
	assert.NotZero(t, project.ID)
	assert.Equal(t, "Test Project", project.Name)
	assert.Equal(t, domain.ProjectStatusPlanning, project.Status)
	assert.Equal(t, int64(1), project.OwnerID)
	require.NotNil(t, project.Description)
	assert.Equal(t, "Test Description", *project.Description)
}

func TestProjectUseCase_Create_WithDates(t *testing.T) {
	uc := setupProjectUseCase()
	start := time.Now()
	end := time.Now().Add(30 * 24 * time.Hour)

	project, err := uc.Create(context.Background(), 1, dto.CreateProjectInput{
		Name:      "Project with dates",
		StartDate: &start,
		EndDate:   &end,
	})

	require.NoError(t, err)
	assert.NotNil(t, project.StartDate)
	assert.NotNil(t, project.EndDate)
}

func TestProjectUseCase_Create_RepoError(t *testing.T) {
	repo := newErrorMockProjectRepo()
	repo.createErr = errors.New("db error")
	uc := NewProjectUseCase(repo, nil)

	_, err := uc.Create(context.Background(), 1, dto.CreateProjectInput{Name: "Test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create project")
}

func TestProjectUseCase_Create_WithAuditLogger(t *testing.T) {
	uc := setupProjectUseCaseWithAudit()

	project, err := uc.Create(context.Background(), 1, dto.CreateProjectInput{Name: "Test"})
	require.NoError(t, err)
	assert.NotZero(t, project.ID)
}

// ===================== GetByID =====================

func TestProjectUseCase_GetByID(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test Project", 1)

	project, err := uc.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, project.ID)
}

func TestProjectUseCase_GetByID_NotFound(t *testing.T) {
	uc := setupProjectUseCase()

	_, err := uc.GetByID(context.Background(), 999)
	assert.ErrorIs(t, err, ErrProjectNotFound)
}

func TestProjectUseCase_GetByID_RepoError(t *testing.T) {
	repo := newErrorMockProjectRepo()
	repo.getByIDErr = errors.New("db error")
	uc := NewProjectUseCase(repo, nil)

	_, err := uc.GetByID(context.Background(), 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get project")
}

// ===================== Update =====================

func TestProjectUseCase_Update(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test Project", 1)

	newName := updatedProjectName
	updated, err := uc.Update(context.Background(), 1, created.ID, dto.UpdateProjectInput{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, updatedProjectName, updated.Name)
}

func TestProjectUseCase_Update_Unauthorized(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test Project", 1)

	newName := updatedProjectName
	_, err := uc.Update(context.Background(), 2, created.ID, dto.UpdateProjectInput{Name: &newName})
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestProjectUseCase_Update_NotFound(t *testing.T) {
	uc := setupProjectUseCase()

	newName := "New"
	_, err := uc.Update(context.Background(), 1, 999, dto.UpdateProjectInput{Name: &newName})
	assert.ErrorIs(t, err, ErrProjectNotFound)
}

func TestProjectUseCase_Update_Description(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test", 1)

	desc := "new description"
	updated, err := uc.Update(context.Background(), 1, created.ID, dto.UpdateProjectInput{Description: &desc})
	require.NoError(t, err)
	require.NotNil(t, updated.Description)
	assert.Equal(t, "new description", *updated.Description)
}

func TestProjectUseCase_Update_Status(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test", 1)

	status := "active"
	updated, err := uc.Update(context.Background(), 1, created.ID, dto.UpdateProjectInput{Status: &status})
	require.NoError(t, err)
	assert.Equal(t, domain.ProjectStatusActive, updated.Status)
}

func TestProjectUseCase_Update_InvalidStatus(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test", 1)

	status := "invalid_status"
	_, err := uc.Update(context.Background(), 1, created.ID, dto.UpdateProjectInput{Status: &status})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestProjectUseCase_Update_StartDate(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test", 1)

	start := time.Now()
	updated, err := uc.Update(context.Background(), 1, created.ID, dto.UpdateProjectInput{StartDate: &start})
	require.NoError(t, err)
	assert.NotNil(t, updated.StartDate)
}

func TestProjectUseCase_Update_EndDate(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test", 1)

	end := time.Now().Add(30 * 24 * time.Hour)
	updated, err := uc.Update(context.Background(), 1, created.ID, dto.UpdateProjectInput{EndDate: &end})
	require.NoError(t, err)
	assert.NotNil(t, updated.EndDate)
}

func TestProjectUseCase_Update_SaveError(t *testing.T) {
	repo := newErrorMockProjectRepo()
	uc := NewProjectUseCase(repo, nil)
	ctx := context.Background()

	project, err := uc.Create(ctx, 1, dto.CreateProjectInput{Name: "Test"})
	require.NoError(t, err)

	repo.saveErr = errors.New("save error")
	newName := "Updated"
	_, err = uc.Update(ctx, 1, project.ID, dto.UpdateProjectInput{Name: &newName})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update project")
}

func TestProjectUseCase_Update_WithAuditLogger(t *testing.T) {
	uc := setupProjectUseCaseWithAudit()
	created := createProject(t, uc, "Test", 1)

	newName := "Updated"
	updated, err := uc.Update(context.Background(), 1, created.ID, dto.UpdateProjectInput{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)
}

func TestProjectUseCase_Update_AllFields(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test", 1)

	newName := "Updated Name"
	desc := "Updated Desc"
	status := "active"
	start := time.Now()
	end := time.Now().Add(72 * time.Hour)

	updated, err := uc.Update(context.Background(), 1, created.ID, dto.UpdateProjectInput{
		Name:        &newName,
		Description: &desc,
		Status:      &status,
		StartDate:   &start,
		EndDate:     &end,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated Desc", *updated.Description)
	assert.Equal(t, domain.ProjectStatusActive, updated.Status)
	assert.NotNil(t, updated.StartDate)
	assert.NotNil(t, updated.EndDate)
}

// ===================== Delete =====================

func TestProjectUseCase_Delete(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test", 1)

	err := uc.Delete(context.Background(), 1, created.ID)
	require.NoError(t, err)

	_, err = uc.GetByID(context.Background(), created.ID)
	assert.ErrorIs(t, err, ErrProjectNotFound)
}

func TestProjectUseCase_Delete_Unauthorized(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test", 1)

	err := uc.Delete(context.Background(), 2, created.ID)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestProjectUseCase_Delete_NotFound(t *testing.T) {
	uc := setupProjectUseCase()

	err := uc.Delete(context.Background(), 1, 999)
	assert.ErrorIs(t, err, ErrProjectNotFound)
}

func TestProjectUseCase_Delete_RepoError(t *testing.T) {
	repo := newErrorMockProjectRepo()
	uc := NewProjectUseCase(repo, nil)
	ctx := context.Background()

	project, err := uc.Create(ctx, 1, dto.CreateProjectInput{Name: "Test"})
	require.NoError(t, err)

	repo.deleteErr = errors.New("delete error")
	err = uc.Delete(ctx, 1, project.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete project")
}

func TestProjectUseCase_Delete_WithAuditLogger(t *testing.T) {
	uc := setupProjectUseCaseWithAudit()
	created := createProject(t, uc, "Test", 1)

	err := uc.Delete(context.Background(), 1, created.ID)
	require.NoError(t, err)
}

// ===================== List =====================

func TestProjectUseCase_List(t *testing.T) {
	uc := setupProjectUseCase()
	ctx := context.Background()

	createProject(t, uc, "Project 1", 1)
	createProject(t, uc, "Project 2", 1)

	output, err := uc.List(ctx, dto.ProjectFilterInput{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(2), output.Total)
	assert.Len(t, output.Projects, 2)
}

func TestProjectUseCase_List_Empty(t *testing.T) {
	uc := setupProjectUseCase()

	output, err := uc.List(context.Background(), dto.ProjectFilterInput{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(0), output.Total)
	assert.Empty(t, output.Projects)
}

func TestProjectUseCase_List_ListError(t *testing.T) {
	repo := newErrorMockProjectRepo()
	repo.listErr = errors.New("list error")
	uc := NewProjectUseCase(repo, nil)

	_, err := uc.List(context.Background(), dto.ProjectFilterInput{Limit: 10})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list projects")
}

func TestProjectUseCase_List_CountError(t *testing.T) {
	repo := newErrorMockProjectRepo()
	repo.countErr = errors.New("count error")
	uc := NewProjectUseCase(repo, nil)

	_, err := uc.List(context.Background(), dto.ProjectFilterInput{Limit: 10})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to count projects")
}

// ===================== Activate =====================

func TestProjectUseCase_Activate(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test", 1)

	project, err := uc.Activate(context.Background(), 1, created.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ProjectStatusActive, project.Status)
}

func TestProjectUseCase_Activate_Unauthorized(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test", 1)

	_, err := uc.Activate(context.Background(), 2, created.ID)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestProjectUseCase_Activate_NotFound(t *testing.T) {
	uc := setupProjectUseCase()

	_, err := uc.Activate(context.Background(), 1, 999)
	assert.ErrorIs(t, err, ErrProjectNotFound)
}

func TestProjectUseCase_Activate_SaveError(t *testing.T) {
	repo := newErrorMockProjectRepo()
	uc := NewProjectUseCase(repo, nil)
	ctx := context.Background()

	project, err := uc.Create(ctx, 1, dto.CreateProjectInput{Name: "Test"})
	require.NoError(t, err)

	repo.saveErr = errors.New("save error")
	_, err = uc.Activate(ctx, 1, project.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to activate project")
}

func TestProjectUseCase_Activate_WithAuditLogger(t *testing.T) {
	uc := setupProjectUseCaseWithAudit()
	created := createProject(t, uc, "Test", 1)

	project, err := uc.Activate(context.Background(), 1, created.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ProjectStatusActive, project.Status)
}

// ===================== PutOnHold =====================

func TestProjectUseCase_PutOnHold(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test", 1)

	project, err := uc.PutOnHold(context.Background(), 1, created.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ProjectStatusOnHold, project.Status)
}

func TestProjectUseCase_PutOnHold_Unauthorized(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test", 1)

	_, err := uc.PutOnHold(context.Background(), 2, created.ID)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestProjectUseCase_PutOnHold_NotFound(t *testing.T) {
	uc := setupProjectUseCase()

	_, err := uc.PutOnHold(context.Background(), 1, 999)
	assert.ErrorIs(t, err, ErrProjectNotFound)
}

func TestProjectUseCase_PutOnHold_SaveError(t *testing.T) {
	repo := newErrorMockProjectRepo()
	uc := NewProjectUseCase(repo, nil)
	ctx := context.Background()

	project, err := uc.Create(ctx, 1, dto.CreateProjectInput{Name: "Test"})
	require.NoError(t, err)

	repo.saveErr = errors.New("save error")
	_, err = uc.PutOnHold(ctx, 1, project.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to put project on hold")
}

func TestProjectUseCase_PutOnHold_WithAuditLogger(t *testing.T) {
	uc := setupProjectUseCaseWithAudit()
	created := createProject(t, uc, "Test", 1)

	project, err := uc.PutOnHold(context.Background(), 1, created.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ProjectStatusOnHold, project.Status)
}

// ===================== Complete =====================

func TestProjectUseCase_Complete(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test", 1)

	project, err := uc.Complete(context.Background(), 1, created.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ProjectStatusCompleted, project.Status)
}

func TestProjectUseCase_Complete_Unauthorized(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test", 1)

	_, err := uc.Complete(context.Background(), 2, created.ID)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestProjectUseCase_Complete_NotFound(t *testing.T) {
	uc := setupProjectUseCase()

	_, err := uc.Complete(context.Background(), 1, 999)
	assert.ErrorIs(t, err, ErrProjectNotFound)
}

func TestProjectUseCase_Complete_SaveError(t *testing.T) {
	repo := newErrorMockProjectRepo()
	uc := NewProjectUseCase(repo, nil)
	ctx := context.Background()

	project, err := uc.Create(ctx, 1, dto.CreateProjectInput{Name: "Test"})
	require.NoError(t, err)

	repo.saveErr = errors.New("save error")
	_, err = uc.Complete(ctx, 1, project.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to complete project")
}

func TestProjectUseCase_Complete_WithAuditLogger(t *testing.T) {
	uc := setupProjectUseCaseWithAudit()
	created := createProject(t, uc, "Test", 1)

	project, err := uc.Complete(context.Background(), 1, created.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ProjectStatusCompleted, project.Status)
}

// ===================== Cancel =====================

func TestProjectUseCase_Cancel(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test", 1)

	project, err := uc.Cancel(context.Background(), 1, created.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ProjectStatusCancelled, project.Status)
}

func TestProjectUseCase_Cancel_Unauthorized(t *testing.T) {
	uc := setupProjectUseCase()
	created := createProject(t, uc, "Test", 1)

	_, err := uc.Cancel(context.Background(), 2, created.ID)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestProjectUseCase_Cancel_NotFound(t *testing.T) {
	uc := setupProjectUseCase()

	_, err := uc.Cancel(context.Background(), 1, 999)
	assert.ErrorIs(t, err, ErrProjectNotFound)
}

func TestProjectUseCase_Cancel_SaveError(t *testing.T) {
	repo := newErrorMockProjectRepo()
	uc := NewProjectUseCase(repo, nil)
	ctx := context.Background()

	project, err := uc.Create(ctx, 1, dto.CreateProjectInput{Name: "Test"})
	require.NoError(t, err)

	repo.saveErr = errors.New("save error")
	_, err = uc.Cancel(ctx, 1, project.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to cancel project")
}

func TestProjectUseCase_Cancel_WithAuditLogger(t *testing.T) {
	uc := setupProjectUseCaseWithAudit()
	created := createProject(t, uc, "Test", 1)

	project, err := uc.Cancel(context.Background(), 1, created.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ProjectStatusCancelled, project.Status)
}

// ===================== Status Workflow =====================

func TestProjectUseCase_StatusWorkflow(t *testing.T) {
	uc := setupProjectUseCaseWithAudit()
	ctx := context.Background()

	project := createProject(t, uc, "Workflow Test", 1)
	assert.Equal(t, domain.ProjectStatusPlanning, project.Status)

	// Activate
	project, err := uc.Activate(ctx, 1, project.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ProjectStatusActive, project.Status)

	// Put on hold
	project, err = uc.PutOnHold(ctx, 1, project.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ProjectStatusOnHold, project.Status)

	// Complete
	project, err = uc.Complete(ctx, 1, project.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ProjectStatusCompleted, project.Status)
}

func TestProjectUseCase_StatusChange_Unauthorized(t *testing.T) {
	uc := setupProjectUseCase()
	project := createProject(t, uc, "Test", 1)

	_, err := uc.Activate(context.Background(), 2, project.ID)
	assert.ErrorIs(t, err, ErrUnauthorized)

	_, err = uc.PutOnHold(context.Background(), 2, project.ID)
	assert.ErrorIs(t, err, ErrUnauthorized)

	_, err = uc.Complete(context.Background(), 2, project.ID)
	assert.ErrorIs(t, err, ErrUnauthorized)

	_, err = uc.Cancel(context.Background(), 2, project.ID)
	assert.ErrorIs(t, err, ErrUnauthorized)
}
