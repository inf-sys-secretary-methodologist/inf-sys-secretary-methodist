package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

type fakeLoadSvc struct {
	listRet    []*entities.TeachingLoad
	createRet  *entities.TeachingLoad
	updateRet  *entities.TeachingLoad
	createErr  error
	updateErr  error
	deleteErr  error
	listFilter usecases.TeachingLoadFilter
	createCnt  int
	deleteCnt  int
}

func (f *fakeLoadSvc) List(_ context.Context, filter usecases.TeachingLoadFilter) ([]*entities.TeachingLoad, error) {
	f.listFilter = filter
	return f.listRet, nil
}
func (f *fakeLoadSvc) Create(_ context.Context, _ usecases.TeachingLoadParams) (*entities.TeachingLoad, error) {
	f.createCnt++
	if f.createErr != nil {
		return nil, f.createErr
	}
	return f.createRet, nil
}
func (f *fakeLoadSvc) Update(_ context.Context, _ int64, _ usecases.TeachingLoadParams) (*entities.TeachingLoad, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return f.updateRet, nil
}
func (f *fakeLoadSvc) Delete(context.Context, int64) error {
	f.deleteCnt++
	return f.deleteErr
}

func setupLoadRouter(svc TeachingLoadService, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewTeachingLoadHandler(svc)
	if role != "" {
		r.Use(func(c *gin.Context) {
			c.Set("user_id", int64(1))
			c.Set("role", role)
			c.Next()
		})
	}
	r.GET("/load", h.List)
	r.POST("/load", h.Create)
	r.PUT("/load/:id", h.Update)
	r.DELETE("/load/:id", h.Delete)
	return r
}

func doLoad(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

const validLoadBody = `{"semester_id":1,"group_id":2,"discipline_id":3,"teacher_id":4,"lesson_type_id":5,"pairs_per_week":2,"week_type":"all"}`

func TestLoadHandler_List_OK(t *testing.T) {
	svc := &fakeLoadSvc{listRet: []*entities.TeachingLoad{{ID: 1, Group: &entities.StudentGroup{Name: "IS-21"}}}}
	w := doLoad(setupLoadRouter(svc, "teacher"), http.MethodGet, "/load?semester_id=3", "")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "IS-21")
	// Pin the {success,data} envelope: the frontend unwraps .data, so a
	// regression to a bare body would silently empty the screen.
	assert.Contains(t, w.Body.String(), `"success":true`)
	require.NotNil(t, svc.listFilter.SemesterID)
	assert.Equal(t, int64(3), *svc.listFilter.SemesterID)
}

func TestLoadHandler_Create_Forbidden(t *testing.T) {
	svc := &fakeLoadSvc{}
	w := doLoad(setupLoadRouter(svc, "teacher"), http.MethodPost, "/load", validLoadBody)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, svc.createCnt)
}

func TestLoadHandler_Create_MethodistOK(t *testing.T) {
	svc := &fakeLoadSvc{createRet: &entities.TeachingLoad{ID: 8}}
	w := doLoad(setupLoadRouter(svc, "methodist"), http.MethodPost, "/load", validLoadBody)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"id":8`)
	assert.Contains(t, w.Body.String(), `"success":true`)
}

func TestLoadHandler_Create_ValidationError(t *testing.T) {
	svc := &fakeLoadSvc{createErr: entities.ErrInvalidLoadPairs}
	w := doLoad(setupLoadRouter(svc, "academic_secretary"), http.MethodPost, "/load", validLoadBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoadHandler_Create_Duplicate(t *testing.T) {
	svc := &fakeLoadSvc{createErr: entities.ErrTeachingLoadDuplicate}
	w := doLoad(setupLoadRouter(svc, "system_admin"), http.MethodPost, "/load", validLoadBody)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestLoadHandler_Create_BadBody(t *testing.T) {
	svc := &fakeLoadSvc{}
	w := doLoad(setupLoadRouter(svc, "methodist"), http.MethodPost, "/load", `{"group_id":2}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, 0, svc.createCnt)
}

func TestLoadHandler_Update_NotFound(t *testing.T) {
	svc := &fakeLoadSvc{updateErr: entities.ErrTeachingLoadNotFound}
	w := doLoad(setupLoadRouter(svc, "methodist"), http.MethodPut, "/load/99", validLoadBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestLoadHandler_Update_BadID(t *testing.T) {
	svc := &fakeLoadSvc{}
	w := doLoad(setupLoadRouter(svc, "methodist"), http.MethodPut, "/load/abc", validLoadBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoadHandler_Delete_OK(t *testing.T) {
	svc := &fakeLoadSvc{}
	w := doLoad(setupLoadRouter(svc, "academic_secretary"), http.MethodDelete, "/load/5", "")

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, 1, svc.deleteCnt)
}

func TestLoadHandler_Delete_Forbidden(t *testing.T) {
	svc := &fakeLoadSvc{}
	w := doLoad(setupLoadRouter(svc, "student"), http.MethodDelete, "/load/5", "")

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, svc.deleteCnt)
}

func TestLoadHandler_Delete_NotFound(t *testing.T) {
	svc := &fakeLoadSvc{deleteErr: entities.ErrTeachingLoadNotFound}
	w := doLoad(setupLoadRouter(svc, "system_admin"), http.MethodDelete, "/load/99", "")

	assert.Equal(t, http.StatusNotFound, w.Code)
}
