package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTaskRouter() *gin.Engine {
	return gin.New()
}

func withTaskAuth(userID int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

func performTaskRequest(router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func sampleTask() *entities.Task {
	now := time.Now()
	return &entities.Task{
		ID:        1,
		Title:     "Test Task",
		AuthorID:  1,
		Status:    domain.TaskStatusNew,
		Priority:  domain.TaskPriorityNormal,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func sampleTaskComment() *entities.TaskComment {
	now := time.Now()
	return &entities.TaskComment{
		ID:        1,
		TaskID:    1,
		AuthorID:  1,
		Content:   "Test comment",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func sampleTaskChecklist() *entities.TaskChecklist {
	return &entities.TaskChecklist{
		ID:        1,
		TaskID:    1,
		Title:     "Checklist",
		Position:  0,
		CreatedAt: time.Now(),
	}
}

func sampleTaskChecklistItem() *entities.TaskChecklistItem {
	return &entities.TaskChecklistItem{
		ID:          1,
		ChecklistID: 1,
		Title:       "Item 1",
		IsCompleted: false,
		Position:    0,
		CreatedAt:   time.Now(),
	}
}

// --- TaskHandler tests ---

func TestNewTaskHandler(t *testing.T) {
	h := NewTaskHandler(nil)
	assert.NotNil(t, h)
}

func TestTaskHandler_getUserID_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks", h.Create)

	w := performTaskRequest(router, http.MethodPost, "/tasks", map[string]interface{}{"title": "T"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_getIDParam_Invalid(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.GET("/tasks/:id", withTaskAuth(1), h.GetByID)

	w := performTaskRequest(router, http.MethodGet, "/tasks/abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Create_InvalidJSON(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks", withTaskAuth(1), h.Create)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Create_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	task := sampleTask()
	taskRepo.On("Create", anyCtx, anyTask).Run(func(args mockArgs) {
		t := args.Get(1).(*entities.Task)
		t.ID = 1
	}).Return(nil)
	taskRepo.On("AddHistory", anyCtx, anyHistory).Return(nil)

	router := setupTaskRouter()
	router.POST("/tasks", withTaskAuth(1), h.Create)

	w := performTaskRequest(router, http.MethodPost, "/tasks", map[string]interface{}{
		"title": task.Title,
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestTaskHandler_Create_UseCaseError(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("Create", anyCtx, anyTask).Return(assert.AnError)

	router := setupTaskRouter()
	router.POST("/tasks", withTaskAuth(1), h.Create)

	w := performTaskRequest(router, http.MethodPost, "/tasks", map[string]interface{}{
		"title": "Test",
	})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTaskHandler_GetByID_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	task := sampleTask()
	taskRepo.On("GetByID", anyCtx, int64(1)).Return(task, nil)

	router := setupTaskRouter()
	router.GET("/tasks/:id", h.GetByID)

	w := performTaskRequest(router, http.MethodGet, "/tasks/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_GetByID_NotFound(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("GetByID", anyCtx, int64(1)).Return(nil, nil)

	router := setupTaskRouter()
	router.GET("/tasks/:id", h.GetByID)

	w := performTaskRequest(router, http.MethodGet, "/tasks/1", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTaskHandler_Update_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	task := sampleTask()
	taskRepo.On("GetByID", anyCtx, int64(1)).Return(task, nil)
	taskRepo.On("Save", anyCtx, anyTask).Return(nil)
	taskRepo.On("AddHistory", anyCtx, anyHistory).Return(nil)

	title := "Updated"
	router := setupTaskRouter()
	router.PUT("/tasks/:id", withTaskAuth(1), h.Update)

	w := performTaskRequest(router, http.MethodPut, "/tasks/1", map[string]interface{}{
		"title": title,
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_Update_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.PUT("/tasks/:id", h.Update)

	w := performTaskRequest(router, http.MethodPut, "/tasks/1", map[string]interface{}{"title": "T"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_Update_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.PUT("/tasks/:id", withTaskAuth(1), h.Update)

	w := performTaskRequest(router, http.MethodPut, "/tasks/abc", map[string]interface{}{"title": "T"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Update_InvalidJSON(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.PUT("/tasks/:id", withTaskAuth(1), h.Update)

	req := httptest.NewRequest(http.MethodPut, "/tasks/1", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Delete_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	task := sampleTask()
	task.AuthorID = 1
	taskRepo.On("GetByID", anyCtx, int64(1)).Return(task, nil)
	taskRepo.On("Delete", anyCtx, int64(1)).Return(nil)

	router := setupTaskRouter()
	router.DELETE("/tasks/:id", withTaskAuth(1), h.Delete)

	w := performTaskRequest(router, http.MethodDelete, "/tasks/1", nil)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestTaskHandler_Delete_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.DELETE("/tasks/:id", h.Delete)

	w := performTaskRequest(router, http.MethodDelete, "/tasks/1", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_Delete_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.DELETE("/tasks/:id", withTaskAuth(1), h.Delete)

	w := performTaskRequest(router, http.MethodDelete, "/tasks/abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_List_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("List", anyCtx, anyFilter, 20, 0).Return([]*entities.Task{sampleTask()}, nil)
	taskRepo.On("Count", anyCtx, anyFilter).Return(int64(1), nil)

	router := setupTaskRouter()
	router.GET("/tasks", h.List)

	w := performTaskRequest(router, http.MethodGet, "/tasks", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_List_WithLimit(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("List", anyCtx, anyFilter, 10, 5).Return([]*entities.Task{}, nil)
	taskRepo.On("Count", anyCtx, anyFilter).Return(int64(0), nil)

	router := setupTaskRouter()
	router.GET("/tasks", h.List)

	w := performTaskRequest(router, http.MethodGet, "/tasks?limit=10&offset=5", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_List_Error(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("List", anyCtx, anyFilter, 20, 0).Return(nil, assert.AnError)

	router := setupTaskRouter()
	router.GET("/tasks", h.List)

	w := performTaskRequest(router, http.MethodGet, "/tasks", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTaskHandler_Assign_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	task := sampleTask()
	task.AuthorID = 1
	taskRepo.On("GetByID", anyCtx, int64(1)).Return(task, nil)
	taskRepo.On("Save", anyCtx, anyTask).Return(nil)
	taskRepo.On("AddHistory", anyCtx, anyHistory).Return(nil)

	router := setupTaskRouter()
	router.POST("/tasks/:id/assign", withTaskAuth(1), h.Assign)

	w := performTaskRequest(router, http.MethodPost, "/tasks/1/assign", map[string]interface{}{
		"assignee_id": 2,
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_Assign_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/assign", h.Assign)

	w := performTaskRequest(router, http.MethodPost, "/tasks/1/assign", map[string]interface{}{"assignee_id": 2})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_Assign_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/assign", withTaskAuth(1), h.Assign)

	w := performTaskRequest(router, http.MethodPost, "/tasks/abc/assign", map[string]interface{}{"assignee_id": 2})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Assign_InvalidJSON(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/assign", withTaskAuth(1), h.Assign)

	req := httptest.NewRequest(http.MethodPost, "/tasks/1/assign", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Unassign_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	task := sampleTask()
	assigneeID := int64(2)
	task.AssigneeID = &assigneeID
	task.Status = domain.TaskStatusAssigned
	task.AuthorID = 1
	taskRepo.On("GetByID", anyCtx, int64(1)).Return(task, nil)
	taskRepo.On("Save", anyCtx, anyTask).Return(nil)
	taskRepo.On("AddHistory", anyCtx, anyHistory).Return(nil)

	router := setupTaskRouter()
	router.POST("/tasks/:id/unassign", withTaskAuth(1), h.Unassign)

	w := performTaskRequest(router, http.MethodPost, "/tasks/1/unassign", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_Unassign_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/unassign", h.Unassign)

	w := performTaskRequest(router, http.MethodPost, "/tasks/1/unassign", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_StartWork_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	task := sampleTask()
	task.AuthorID = 1
	taskRepo.On("GetByID", anyCtx, int64(1)).Return(task, nil)
	taskRepo.On("Save", anyCtx, anyTask).Return(nil)
	taskRepo.On("AddHistory", anyCtx, anyHistory).Return(nil)

	router := setupTaskRouter()
	router.POST("/tasks/:id/start", withTaskAuth(1), h.StartWork)

	w := performTaskRequest(router, http.MethodPost, "/tasks/1/start", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_StartWork_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/start", h.StartWork)

	w := performTaskRequest(router, http.MethodPost, "/tasks/1/start", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_SubmitForReview_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	task := sampleTask()
	task.Status = domain.TaskStatusInProgress
	task.AuthorID = 1
	taskRepo.On("GetByID", anyCtx, int64(1)).Return(task, nil)
	taskRepo.On("Save", anyCtx, anyTask).Return(nil)
	taskRepo.On("AddHistory", anyCtx, anyHistory).Return(nil)

	router := setupTaskRouter()
	router.POST("/tasks/:id/review", withTaskAuth(1), h.SubmitForReview)

	w := performTaskRequest(router, http.MethodPost, "/tasks/1/review", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_Complete_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	task := sampleTask()
	task.Status = domain.TaskStatusInProgress
	task.AuthorID = 1
	taskRepo.On("GetByID", anyCtx, int64(1)).Return(task, nil)
	taskRepo.On("Save", anyCtx, anyTask).Return(nil)
	taskRepo.On("AddHistory", anyCtx, anyHistory).Return(nil)

	router := setupTaskRouter()
	router.POST("/tasks/:id/complete", withTaskAuth(1), h.Complete)

	w := performTaskRequest(router, http.MethodPost, "/tasks/1/complete", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_Cancel_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	task := sampleTask()
	task.AuthorID = 1
	taskRepo.On("GetByID", anyCtx, int64(1)).Return(task, nil)
	taskRepo.On("Save", anyCtx, anyTask).Return(nil)
	taskRepo.On("AddHistory", anyCtx, anyHistory).Return(nil)

	router := setupTaskRouter()
	router.POST("/tasks/:id/cancel", withTaskAuth(1), h.Cancel)

	w := performTaskRequest(router, http.MethodPost, "/tasks/1/cancel", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_Reopen_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	task := sampleTask()
	task.Status = domain.TaskStatusCompleted
	task.AuthorID = 1
	taskRepo.On("GetByID", anyCtx, int64(1)).Return(task, nil)
	taskRepo.On("Save", anyCtx, anyTask).Return(nil)
	taskRepo.On("AddHistory", anyCtx, anyHistory).Return(nil)

	router := setupTaskRouter()
	router.POST("/tasks/:id/reopen", withTaskAuth(1), h.Reopen)

	w := performTaskRequest(router, http.MethodPost, "/tasks/1/reopen", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_AddWatcher_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	task := sampleTask()
	task.AuthorID = 1
	taskRepo.On("GetByID", anyCtx, int64(1)).Return(task, nil)
	taskRepo.On("IsWatching", anyCtx, int64(1), int64(2)).Return(false, nil)
	taskRepo.On("AddWatcher", anyCtx, anyWatcher).Return(nil)

	router := setupTaskRouter()
	router.POST("/tasks/:id/watchers", withTaskAuth(1), h.AddWatcher)

	w := performTaskRequest(router, http.MethodPost, "/tasks/1/watchers", map[string]interface{}{
		"user_id": 2,
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestTaskHandler_AddWatcher_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/watchers", h.AddWatcher)

	w := performTaskRequest(router, http.MethodPost, "/tasks/1/watchers", map[string]interface{}{"user_id": 2})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_AddWatcher_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/watchers", withTaskAuth(1), h.AddWatcher)

	w := performTaskRequest(router, http.MethodPost, "/tasks/abc/watchers", map[string]interface{}{"user_id": 2})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_AddWatcher_InvalidJSON(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/watchers", withTaskAuth(1), h.AddWatcher)

	req := httptest.NewRequest(http.MethodPost, "/tasks/1/watchers", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_RemoveWatcher_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	task := sampleTask()
	task.AuthorID = 1
	taskRepo.On("GetByID", anyCtx, int64(1)).Return(task, nil)
	taskRepo.On("RemoveWatcher", anyCtx, int64(1), int64(2)).Return(nil)

	router := setupTaskRouter()
	router.DELETE("/tasks/:id/watchers/:watcher_id", withTaskAuth(1), h.RemoveWatcher)

	w := performTaskRequest(router, http.MethodDelete, "/tasks/1/watchers/2", nil)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestTaskHandler_RemoveWatcher_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.DELETE("/tasks/:id/watchers/:watcher_id", h.RemoveWatcher)

	w := performTaskRequest(router, http.MethodDelete, "/tasks/1/watchers/2", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_RemoveWatcher_InvalidTaskID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.DELETE("/tasks/:id/watchers/:watcher_id", withTaskAuth(1), h.RemoveWatcher)

	w := performTaskRequest(router, http.MethodDelete, "/tasks/abc/watchers/2", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_RemoveWatcher_InvalidWatcherID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.DELETE("/tasks/:id/watchers/:watcher_id", withTaskAuth(1), h.RemoveWatcher)

	w := performTaskRequest(router, http.MethodDelete, "/tasks/1/watchers/abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_GetWatchers_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	watchers := []*entities.TaskWatcher{
		{TaskID: 1, UserID: 2, User: &entities.WatcherUser{ID: 2, Name: "User2", Email: "u2@test.com"}},
	}
	taskRepo.On("GetWatchers", anyCtx, int64(1)).Return(watchers, nil)

	router := setupTaskRouter()
	router.GET("/tasks/:id/watchers", h.GetWatchers)

	w := performTaskRequest(router, http.MethodGet, "/tasks/1/watchers", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_GetWatchers_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.GET("/tasks/:id/watchers", h.GetWatchers)

	w := performTaskRequest(router, http.MethodGet, "/tasks/abc/watchers", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_GetWatchers_Error(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("GetWatchers", anyCtx, int64(1)).Return(nil, assert.AnError)

	router := setupTaskRouter()
	router.GET("/tasks/:id/watchers", h.GetWatchers)

	w := performTaskRequest(router, http.MethodGet, "/tasks/1/watchers", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTaskHandler_GetWatchers_NilUser(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	watchers := []*entities.TaskWatcher{
		{TaskID: 1, UserID: 2, User: nil},
	}
	taskRepo.On("GetWatchers", anyCtx, int64(1)).Return(watchers, nil)

	router := setupTaskRouter()
	router.GET("/tasks/:id/watchers", h.GetWatchers)

	w := performTaskRequest(router, http.MethodGet, "/tasks/1/watchers", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_AddComment_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	task := sampleTask()
	task.AuthorID = 1
	comment := sampleTaskComment()
	taskRepo.On("GetByID", anyCtx, int64(1)).Return(task, nil)
	taskRepo.On("AddComment", anyCtx, anyComment).Run(func(args mockArgs) {
		c := args.Get(1).(*entities.TaskComment)
		c.ID = comment.ID
	}).Return(nil)

	router := setupTaskRouter()
	router.POST("/tasks/:id/comments", withTaskAuth(1), h.AddComment)

	w := performTaskRequest(router, http.MethodPost, "/tasks/1/comments", map[string]interface{}{
		"content": "Test comment",
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestTaskHandler_AddComment_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/comments", h.AddComment)

	w := performTaskRequest(router, http.MethodPost, "/tasks/1/comments", map[string]interface{}{"content": "test"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_AddComment_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/comments", withTaskAuth(1), h.AddComment)

	w := performTaskRequest(router, http.MethodPost, "/tasks/abc/comments", map[string]interface{}{"content": "test"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_AddComment_InvalidJSON(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/comments", withTaskAuth(1), h.AddComment)

	req := httptest.NewRequest(http.MethodPost, "/tasks/1/comments", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_UpdateComment_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	comment := sampleTaskComment()
	comment.AuthorID = 1
	taskRepo.On("GetCommentByID", anyCtx, int64(1)).Return(comment, nil)
	taskRepo.On("UpdateComment", anyCtx, anyComment).Return(nil)

	router := setupTaskRouter()
	router.PUT("/comments/:comment_id", withTaskAuth(1), h.UpdateComment)

	w := performTaskRequest(router, http.MethodPut, "/comments/1", map[string]interface{}{
		"content": "Updated",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_UpdateComment_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.PUT("/comments/:comment_id", h.UpdateComment)

	w := performTaskRequest(router, http.MethodPut, "/comments/1", map[string]interface{}{"content": "x"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_UpdateComment_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.PUT("/comments/:comment_id", withTaskAuth(1), h.UpdateComment)

	w := performTaskRequest(router, http.MethodPut, "/comments/abc", map[string]interface{}{"content": "x"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_DeleteComment_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	comment := sampleTaskComment()
	comment.AuthorID = 1
	taskRepo.On("GetCommentByID", anyCtx, int64(1)).Return(comment, nil)
	taskRepo.On("DeleteComment", anyCtx, int64(1)).Return(nil)

	router := setupTaskRouter()
	router.DELETE("/comments/:comment_id", withTaskAuth(1), h.DeleteComment)

	w := performTaskRequest(router, http.MethodDelete, "/comments/1", nil)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestTaskHandler_DeleteComment_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.DELETE("/comments/:comment_id", h.DeleteComment)

	w := performTaskRequest(router, http.MethodDelete, "/comments/1", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_GetComments_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	comments := []*entities.TaskComment{sampleTaskComment()}
	taskRepo.On("GetComments", anyCtx, int64(1)).Return(comments, nil)

	router := setupTaskRouter()
	router.GET("/tasks/:id/comments", h.GetComments)

	w := performTaskRequest(router, http.MethodGet, "/tasks/1/comments", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_GetComments_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.GET("/tasks/:id/comments", h.GetComments)

	w := performTaskRequest(router, http.MethodGet, "/tasks/abc/comments", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_AddChecklist_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	task := sampleTask()
	task.AuthorID = 1
	checklist := sampleTaskChecklist()
	taskRepo.On("GetByID", anyCtx, int64(1)).Return(task, nil)
	taskRepo.On("GetChecklists", anyCtx, int64(1)).Return([]*entities.TaskChecklist{}, nil)
	taskRepo.On("AddChecklist", anyCtx, anyChecklist).Run(func(args mockArgs) {
		c := args.Get(1).(*entities.TaskChecklist)
		c.ID = checklist.ID
	}).Return(nil)

	router := setupTaskRouter()
	router.POST("/tasks/:id/checklists", withTaskAuth(1), h.AddChecklist)

	w := performTaskRequest(router, http.MethodPost, "/tasks/1/checklists", map[string]interface{}{
		"title": "Checklist",
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestTaskHandler_AddChecklist_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/checklists", h.AddChecklist)

	w := performTaskRequest(router, http.MethodPost, "/tasks/1/checklists", map[string]interface{}{"title": "T"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_AddChecklist_InvalidJSON(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/checklists", withTaskAuth(1), h.AddChecklist)

	req := httptest.NewRequest(http.MethodPost, "/tasks/1/checklists", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_DeleteChecklist_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("GetChecklistByID", anyCtx, int64(1)).Return(sampleTaskChecklist(), nil)
	taskRepo.On("DeleteChecklist", anyCtx, int64(1)).Return(nil)

	router := setupTaskRouter()
	router.DELETE("/checklists/:checklist_id", withTaskAuth(1), h.DeleteChecklist)

	w := performTaskRequest(router, http.MethodDelete, "/checklists/1", nil)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestTaskHandler_DeleteChecklist_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.DELETE("/checklists/:checklist_id", h.DeleteChecklist)

	w := performTaskRequest(router, http.MethodDelete, "/checklists/1", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_GetChecklists_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	checklists := []*entities.TaskChecklist{sampleTaskChecklist()}
	taskRepo.On("GetChecklists", anyCtx, int64(1)).Return(checklists, nil)
	taskRepo.On("GetChecklistItems", anyCtx, int64(1)).Return([]*entities.TaskChecklistItem{}, nil)

	router := setupTaskRouter()
	router.GET("/tasks/:id/checklists", h.GetChecklists)

	w := performTaskRequest(router, http.MethodGet, "/tasks/1/checklists", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_GetChecklists_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.GET("/tasks/:id/checklists", h.GetChecklists)

	w := performTaskRequest(router, http.MethodGet, "/tasks/abc/checklists", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_AddChecklistItem_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	item := sampleTaskChecklistItem()
	taskRepo.On("GetChecklistItems", anyCtx, int64(1)).Return([]*entities.TaskChecklistItem{}, nil)
	taskRepo.On("AddChecklistItem", anyCtx, anyChecklistItem).Run(func(args mockArgs) {
		i := args.Get(1).(*entities.TaskChecklistItem)
		i.ID = item.ID
	}).Return(nil)

	router := setupTaskRouter()
	router.POST("/checklists/:checklist_id/items", withTaskAuth(1), h.AddChecklistItem)

	w := performTaskRequest(router, http.MethodPost, "/checklists/1/items", map[string]interface{}{
		"title": "Item 1",
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestTaskHandler_AddChecklistItem_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/checklists/:checklist_id/items", h.AddChecklistItem)

	w := performTaskRequest(router, http.MethodPost, "/checklists/1/items", map[string]interface{}{"title": "T"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_AddChecklistItem_InvalidJSON(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/checklists/:checklist_id/items", withTaskAuth(1), h.AddChecklistItem)

	req := httptest.NewRequest(http.MethodPost, "/checklists/1/items", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_DeleteChecklistItem_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("GetChecklistItemByID", anyCtx, int64(1)).Return(sampleTaskChecklistItem(), nil)
	taskRepo.On("GetChecklistByID", anyCtx, int64(1)).Return(sampleTaskChecklist(), nil)
	taskRepo.On("GetByID", anyCtx, int64(1)).Return(sampleTask(), nil)
	taskRepo.On("DeleteChecklistItem", anyCtx, int64(1)).Return(nil)

	router := setupTaskRouter()
	router.DELETE("/items/:item_id", withTaskAuth(1), h.DeleteChecklistItem)

	w := performTaskRequest(router, http.MethodDelete, "/items/1", nil)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestTaskHandler_DeleteChecklistItem_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.DELETE("/items/:item_id", h.DeleteChecklistItem)

	w := performTaskRequest(router, http.MethodDelete, "/items/1", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_GetHistory_Success(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	history := []*entities.TaskHistory{
		{ID: 1, TaskID: 1, FieldName: "status", CreatedAt: time.Now()},
	}
	taskRepo.On("GetHistory", anyCtx, int64(1), 50, 0).Return(history, nil)

	router := setupTaskRouter()
	router.GET("/tasks/:id/history", h.GetHistory)

	w := performTaskRequest(router, http.MethodGet, "/tasks/1/history", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_GetHistory_WithParams(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("GetHistory", anyCtx, int64(1), 10, 5).Return([]*entities.TaskHistory{}, nil)

	router := setupTaskRouter()
	router.GET("/tasks/:id/history", h.GetHistory)

	w := performTaskRequest(router, http.MethodGet, "/tasks/1/history?limit=10&offset=5", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTaskHandler_GetHistory_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.GET("/tasks/:id/history", h.GetHistory)

	w := performTaskRequest(router, http.MethodGet, "/tasks/abc/history", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_GetHistory_Error(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("GetHistory", anyCtx, int64(1), 50, 0).Return(nil, assert.AnError)

	router := setupTaskRouter()
	router.GET("/tasks/:id/history", h.GetHistory)

	w := performTaskRequest(router, http.MethodGet, "/tasks/1/history", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTaskHandler_handleError_AllCases(t *testing.T) {
	h := NewTaskHandler(nil)

	tests := []struct {
		name       string
		err        error
		statusCode int
	}{
		{"TaskNotFound", usecases.ErrTaskNotFound, http.StatusNotFound},
		{"Unauthorized", usecases.ErrUnauthorized, http.StatusForbidden},
		{"CannotModify", usecases.ErrCannotModifyTask, http.StatusConflict},
		{"InvalidInput", usecases.ErrInvalidInput, http.StatusBadRequest},
		{"CommentNotFound", usecases.ErrCommentNotFound, http.StatusNotFound},
		{"ChecklistNotFound", usecases.ErrChecklistNotFound, http.StatusNotFound},
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

// Verify List with default limit
func TestTaskHandler_List_DefaultLimit(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("List", anyCtx, anyFilter, 20, 0).Return([]*entities.Task{}, nil)
	taskRepo.On("Count", anyCtx, anyFilter).Return(int64(0), nil)

	router := setupTaskRouter()
	router.GET("/tasks", h.List)

	w := performTaskRequest(router, http.MethodGet, "/tasks", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var result dto.TaskListOutput
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
}

// Error path tests for status transitions
func TestTaskHandler_SubmitForReview_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/review", h.SubmitForReview)
	w := performTaskRequest(router, http.MethodPost, "/tasks/1/review", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_SubmitForReview_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/review", withTaskAuth(1), h.SubmitForReview)
	w := performTaskRequest(router, http.MethodPost, "/tasks/abc/review", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Complete_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/complete", h.Complete)
	w := performTaskRequest(router, http.MethodPost, "/tasks/1/complete", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_Complete_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/complete", withTaskAuth(1), h.Complete)
	w := performTaskRequest(router, http.MethodPost, "/tasks/abc/complete", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Cancel_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/cancel", h.Cancel)
	w := performTaskRequest(router, http.MethodPost, "/tasks/1/cancel", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_Cancel_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/cancel", withTaskAuth(1), h.Cancel)
	w := performTaskRequest(router, http.MethodPost, "/tasks/abc/cancel", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Reopen_NoAuth(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/reopen", h.Reopen)
	w := performTaskRequest(router, http.MethodPost, "/tasks/1/reopen", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTaskHandler_Reopen_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/reopen", withTaskAuth(1), h.Reopen)
	w := performTaskRequest(router, http.MethodPost, "/tasks/abc/reopen", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_StartWork_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/start", withTaskAuth(1), h.StartWork)
	w := performTaskRequest(router, http.MethodPost, "/tasks/abc/start", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Unassign_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/unassign", withTaskAuth(1), h.Unassign)
	w := performTaskRequest(router, http.MethodPost, "/tasks/abc/unassign", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_Delete_Error(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	task := sampleTask()
	task.AuthorID = 1
	taskRepo.On("GetByID", anyCtx, int64(1)).Return(task, nil)
	taskRepo.On("Delete", anyCtx, int64(1)).Return(assert.AnError)

	router := setupTaskRouter()
	router.DELETE("/tasks/:id", withTaskAuth(1), h.Delete)
	w := performTaskRequest(router, http.MethodDelete, "/tasks/1", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTaskHandler_GetComments_Error(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("GetComments", anyCtx, int64(1)).Return(nil, assert.AnError)

	router := setupTaskRouter()
	router.GET("/tasks/:id/comments", h.GetComments)
	w := performTaskRequest(router, http.MethodGet, "/tasks/1/comments", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTaskHandler_GetChecklists_Error(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("GetChecklists", anyCtx, int64(1)).Return(nil, assert.AnError)

	router := setupTaskRouter()
	router.GET("/tasks/:id/checklists", h.GetChecklists)
	w := performTaskRequest(router, http.MethodGet, "/tasks/1/checklists", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTaskHandler_UpdateComment_Error(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("GetCommentByID", anyCtx, int64(1)).Return(nil, nil)

	router := setupTaskRouter()
	router.PUT("/comments/:comment_id", withTaskAuth(1), h.UpdateComment)
	w := performTaskRequest(router, http.MethodPut, "/comments/1", map[string]interface{}{"content": "x"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTaskHandler_UpdateComment_InvalidJSON(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.PUT("/comments/:comment_id", withTaskAuth(1), h.UpdateComment)
	req := httptest.NewRequest(http.MethodPut, "/comments/1", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_DeleteComment_Error(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("GetCommentByID", anyCtx, int64(1)).Return(nil, nil)

	router := setupTaskRouter()
	router.DELETE("/comments/:comment_id", withTaskAuth(1), h.DeleteComment)
	w := performTaskRequest(router, http.MethodDelete, "/comments/1", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTaskHandler_DeleteComment_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.DELETE("/comments/:comment_id", withTaskAuth(1), h.DeleteComment)
	w := performTaskRequest(router, http.MethodDelete, "/comments/abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_DeleteChecklist_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.DELETE("/checklists/:checklist_id", withTaskAuth(1), h.DeleteChecklist)
	w := performTaskRequest(router, http.MethodDelete, "/checklists/abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_DeleteChecklist_Error(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("DeleteChecklist", anyCtx, int64(1)).Return(assert.AnError)

	router := setupTaskRouter()
	router.DELETE("/checklists/:checklist_id", withTaskAuth(1), h.DeleteChecklist)
	w := performTaskRequest(router, http.MethodDelete, "/checklists/1", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTaskHandler_AddChecklistItem_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/checklists/:checklist_id/items", withTaskAuth(1), h.AddChecklistItem)
	w := performTaskRequest(router, http.MethodPost, "/checklists/abc/items", map[string]interface{}{"title": "T"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_DeleteChecklistItem_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.DELETE("/items/:item_id", withTaskAuth(1), h.DeleteChecklistItem)
	w := performTaskRequest(router, http.MethodDelete, "/items/abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_DeleteChecklistItem_Error(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("DeleteChecklistItem", anyCtx, int64(1)).Return(assert.AnError)

	router := setupTaskRouter()
	router.DELETE("/items/:item_id", withTaskAuth(1), h.DeleteChecklistItem)
	w := performTaskRequest(router, http.MethodDelete, "/items/1", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTaskHandler_AddChecklist_InvalidID(t *testing.T) {
	h := NewTaskHandler(nil)
	router := setupTaskRouter()
	router.POST("/tasks/:id/checklists", withTaskAuth(1), h.AddChecklist)
	w := performTaskRequest(router, http.MethodPost, "/tasks/abc/checklists", map[string]interface{}{"title": "T"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTaskHandler_AddComment_Error(t *testing.T) {
	taskRepo := new(mockTaskRepo)
	projectRepo := new(mockProjectRepo)
	uc := usecases.NewTaskUseCase(taskRepo, projectRepo, nil, nil)
	h := NewTaskHandler(uc)

	taskRepo.On("GetByID", anyCtx, int64(1)).Return(nil, nil)

	router := setupTaskRouter()
	router.POST("/tasks/:id/comments", withTaskAuth(1), h.AddComment)
	w := performTaskRequest(router, http.MethodPost, "/tasks/1/comments", map[string]interface{}{"content": "x"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}
