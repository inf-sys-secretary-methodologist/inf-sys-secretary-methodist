package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
)

func sampleProject() *entities.Project {
	now := time.Now()
	return &entities.Project{
		ID:        1,
		Name:      "Test Project",
		OwnerID:   1,
		Status:    domain.ProjectStatusPlanning,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestNewProjectHandler(t *testing.T) {
	h := NewProjectHandler(nil)
	assert.NotNil(t, h)
}

func TestProjectHandler_getUserID_NoAuth(t *testing.T) {
	h := NewProjectHandler(nil)
	router := setupTaskRouter()
	router.POST("/projects", h.Create)

	w := performTaskRequest(router, http.MethodPost, "/projects", map[string]interface{}{"name": "P"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProjectHandler_getUserID_InvalidType(t *testing.T) {
	h := NewProjectHandler(nil)
	router := setupTaskRouter()
	router.POST("/projects", func(c *gin.Context) {
		c.Set("user_id", "not_an_int64") // wrong type
		c.Next()
	}, h.Create)

	w := performTaskRequest(router, http.MethodPost, "/projects", map[string]interface{}{"name": "P"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProjectHandler_getIDParam_Invalid(t *testing.T) {
	h := NewProjectHandler(nil)
	router := setupTaskRouter()
	router.GET("/projects/:id", withTaskAuth(1), h.GetByID)

	w := performTaskRequest(router, http.MethodGet, "/projects/abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProjectHandler_Create_Success(t *testing.T) {
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewProjectUseCase(projectRepo, nil)
	h := NewProjectHandler(uc)

	projectRepo.On("Create", anyCtx, anyProject).Run(func(args mockArgs) {
		p := args.Get(1).(*entities.Project)
		p.ID = 1
	}).Return(nil)

	router := setupTaskRouter()
	router.POST("/projects", withTaskAuth(1), h.Create)

	w := performTaskRequest(router, http.MethodPost, "/projects", map[string]interface{}{
		"name": "Test Project",
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestProjectHandler_Create_InvalidJSON(t *testing.T) {
	h := NewProjectHandler(nil)
	router := setupTaskRouter()
	router.POST("/projects", withTaskAuth(1), h.Create)

	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProjectHandler_Create_Error(t *testing.T) {
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewProjectUseCase(projectRepo, nil)
	h := NewProjectHandler(uc)

	projectRepo.On("Create", anyCtx, anyProject).Return(assert.AnError)

	router := setupTaskRouter()
	router.POST("/projects", withTaskAuth(1), h.Create)

	w := performTaskRequest(router, http.MethodPost, "/projects", map[string]interface{}{
		"name": "Test Project",
	})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestProjectHandler_GetByID_Success(t *testing.T) {
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewProjectUseCase(projectRepo, nil)
	h := NewProjectHandler(uc)

	project := sampleProject()
	projectRepo.On("GetByID", anyCtx, int64(1)).Return(project, nil)

	router := setupTaskRouter()
	router.GET("/projects/:id", h.GetByID)

	w := performTaskRequest(router, http.MethodGet, "/projects/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProjectHandler_GetByID_NotFound(t *testing.T) {
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewProjectUseCase(projectRepo, nil)
	h := NewProjectHandler(uc)

	projectRepo.On("GetByID", anyCtx, int64(1)).Return(nil, nil)

	router := setupTaskRouter()
	router.GET("/projects/:id", h.GetByID)

	w := performTaskRequest(router, http.MethodGet, "/projects/1", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestProjectHandler_Update_Success(t *testing.T) {
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewProjectUseCase(projectRepo, nil)
	h := NewProjectHandler(uc)

	project := sampleProject()
	project.OwnerID = 1
	projectRepo.On("GetByID", anyCtx, int64(1)).Return(project, nil)
	projectRepo.On("Save", anyCtx, anyProject).Return(nil)

	router := setupTaskRouter()
	router.PUT("/projects/:id", withTaskAuth(1), h.Update)

	w := performTaskRequest(router, http.MethodPut, "/projects/1", map[string]interface{}{
		"name": "Updated",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProjectHandler_Update_NoAuth(t *testing.T) {
	h := NewProjectHandler(nil)
	router := setupTaskRouter()
	router.PUT("/projects/:id", h.Update)

	w := performTaskRequest(router, http.MethodPut, "/projects/1", map[string]interface{}{"name": "x"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProjectHandler_Update_InvalidID(t *testing.T) {
	h := NewProjectHandler(nil)
	router := setupTaskRouter()
	router.PUT("/projects/:id", withTaskAuth(1), h.Update)

	w := performTaskRequest(router, http.MethodPut, "/projects/abc", map[string]interface{}{"name": "x"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProjectHandler_Update_InvalidJSON(t *testing.T) {
	h := NewProjectHandler(nil)
	router := setupTaskRouter()
	router.PUT("/projects/:id", withTaskAuth(1), h.Update)

	req := httptest.NewRequest(http.MethodPut, "/projects/1", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProjectHandler_Delete_Success(t *testing.T) {
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewProjectUseCase(projectRepo, nil)
	h := NewProjectHandler(uc)

	project := sampleProject()
	project.OwnerID = 1
	projectRepo.On("GetByID", anyCtx, int64(1)).Return(project, nil)
	projectRepo.On("Delete", anyCtx, int64(1)).Return(nil)

	router := setupTaskRouter()
	router.DELETE("/projects/:id", withTaskAuth(1), h.Delete)

	w := performTaskRequest(router, http.MethodDelete, "/projects/1", nil)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestProjectHandler_Delete_NoAuth(t *testing.T) {
	h := NewProjectHandler(nil)
	router := setupTaskRouter()
	router.DELETE("/projects/:id", h.Delete)

	w := performTaskRequest(router, http.MethodDelete, "/projects/1", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProjectHandler_Delete_InvalidID(t *testing.T) {
	h := NewProjectHandler(nil)
	router := setupTaskRouter()
	router.DELETE("/projects/:id", withTaskAuth(1), h.Delete)

	w := performTaskRequest(router, http.MethodDelete, "/projects/abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProjectHandler_List_Success(t *testing.T) {
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewProjectUseCase(projectRepo, nil)
	h := NewProjectHandler(uc)

	projectRepo.On("List", anyCtx, anyFilter, 20, 0).Return([]*entities.Project{sampleProject()}, nil)
	projectRepo.On("Count", anyCtx, anyFilter).Return(int64(1), nil)

	router := setupTaskRouter()
	router.GET("/projects", h.List)

	w := performTaskRequest(router, http.MethodGet, "/projects", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProjectHandler_List_WithParams(t *testing.T) {
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewProjectUseCase(projectRepo, nil)
	h := NewProjectHandler(uc)

	projectRepo.On("List", anyCtx, anyFilter, 10, 5).Return([]*entities.Project{}, nil)
	projectRepo.On("Count", anyCtx, anyFilter).Return(int64(0), nil)

	router := setupTaskRouter()
	router.GET("/projects", h.List)

	w := performTaskRequest(router, http.MethodGet, "/projects?limit=10&offset=5", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProjectHandler_List_Error(t *testing.T) {
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewProjectUseCase(projectRepo, nil)
	h := NewProjectHandler(uc)

	projectRepo.On("List", anyCtx, anyFilter, 20, 0).Return(nil, assert.AnError)

	router := setupTaskRouter()
	router.GET("/projects", h.List)

	w := performTaskRequest(router, http.MethodGet, "/projects", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestProjectHandler_Activate_Success(t *testing.T) {
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewProjectUseCase(projectRepo, nil)
	h := NewProjectHandler(uc)

	project := sampleProject()
	project.OwnerID = 1
	projectRepo.On("GetByID", anyCtx, int64(1)).Return(project, nil)
	projectRepo.On("Save", anyCtx, anyProject).Return(nil)

	router := setupTaskRouter()
	router.POST("/projects/:id/activate", withTaskAuth(1), h.Activate)

	w := performTaskRequest(router, http.MethodPost, "/projects/1/activate", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProjectHandler_Activate_NoAuth(t *testing.T) {
	h := NewProjectHandler(nil)
	router := setupTaskRouter()
	router.POST("/projects/:id/activate", h.Activate)

	w := performTaskRequest(router, http.MethodPost, "/projects/1/activate", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProjectHandler_Activate_InvalidID(t *testing.T) {
	h := NewProjectHandler(nil)
	router := setupTaskRouter()
	router.POST("/projects/:id/activate", withTaskAuth(1), h.Activate)

	w := performTaskRequest(router, http.MethodPost, "/projects/abc/activate", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProjectHandler_PutOnHold_Success(t *testing.T) {
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewProjectUseCase(projectRepo, nil)
	h := NewProjectHandler(uc)

	project := sampleProject()
	project.OwnerID = 1
	projectRepo.On("GetByID", anyCtx, int64(1)).Return(project, nil)
	projectRepo.On("Save", anyCtx, anyProject).Return(nil)

	router := setupTaskRouter()
	router.POST("/projects/:id/hold", withTaskAuth(1), h.PutOnHold)

	w := performTaskRequest(router, http.MethodPost, "/projects/1/hold", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProjectHandler_PutOnHold_NoAuth(t *testing.T) {
	h := NewProjectHandler(nil)
	router := setupTaskRouter()
	router.POST("/projects/:id/hold", h.PutOnHold)

	w := performTaskRequest(router, http.MethodPost, "/projects/1/hold", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProjectHandler_Complete_Success(t *testing.T) {
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewProjectUseCase(projectRepo, nil)
	h := NewProjectHandler(uc)

	project := sampleProject()
	project.OwnerID = 1
	projectRepo.On("GetByID", anyCtx, int64(1)).Return(project, nil)
	projectRepo.On("Save", anyCtx, anyProject).Return(nil)

	router := setupTaskRouter()
	router.POST("/projects/:id/complete", withTaskAuth(1), h.Complete)

	w := performTaskRequest(router, http.MethodPost, "/projects/1/complete", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProjectHandler_Complete_NoAuth(t *testing.T) {
	h := NewProjectHandler(nil)
	router := setupTaskRouter()
	router.POST("/projects/:id/complete", h.Complete)

	w := performTaskRequest(router, http.MethodPost, "/projects/1/complete", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProjectHandler_Cancel_Success(t *testing.T) {
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewProjectUseCase(projectRepo, nil)
	h := NewProjectHandler(uc)

	project := sampleProject()
	project.OwnerID = 1
	projectRepo.On("GetByID", anyCtx, int64(1)).Return(project, nil)
	projectRepo.On("Save", anyCtx, anyProject).Return(nil)

	router := setupTaskRouter()
	router.POST("/projects/:id/cancel", withTaskAuth(1), h.Cancel)

	w := performTaskRequest(router, http.MethodPost, "/projects/1/cancel", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProjectHandler_Cancel_NoAuth(t *testing.T) {
	h := NewProjectHandler(nil)
	router := setupTaskRouter()
	router.POST("/projects/:id/cancel", h.Cancel)

	w := performTaskRequest(router, http.MethodPost, "/projects/1/cancel", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProjectHandler_handleError_AllCases(t *testing.T) {
	h := NewProjectHandler(nil)

	tests := []struct {
		name       string
		err        error
		statusCode int
	}{
		{"ProjectNotFound", usecases.ErrProjectNotFound, http.StatusNotFound},
		{"Unauthorized", usecases.ErrUnauthorized, http.StatusForbidden},
		{"CannotModify", usecases.ErrCannotModifyProject, http.StatusConflict},
		{"InvalidInput", usecases.ErrInvalidInput, http.StatusBadRequest},
		{"DefaultError", assert.AnError, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTaskRouter()
			router.GET("/test", func(c *gin.Context) {
				h.handleError(c, tt.err)
			})
			w := performTaskRequest(router, http.MethodGet, "/test", nil)
			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}
