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

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

type fakeSlotSvc struct {
	listRet   []*entities.LessonSlot
	createRet *entities.LessonSlot
	updateRet *entities.LessonSlot
	createErr error
	updateErr error
	deleteErr error
	listErr   error
	createCnt int
	deleteCnt int
}

func (f *fakeSlotSvc) List(context.Context) ([]*entities.LessonSlot, error) {
	return f.listRet, f.listErr
}

func (f *fakeSlotSvc) Create(_ context.Context, number int, timeStart, timeEnd string) (*entities.LessonSlot, error) {
	f.createCnt++
	if f.createErr != nil {
		return nil, f.createErr
	}
	return f.createRet, nil
}

func (f *fakeSlotSvc) Update(_ context.Context, id int64, number int, timeStart, timeEnd string) (*entities.LessonSlot, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return f.updateRet, nil
}

func (f *fakeSlotSvc) Delete(context.Context, int64) error {
	f.deleteCnt++
	return f.deleteErr
}

func setupSlotRouter(svc LessonSlotService, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewLessonSlotHandler(svc)
	if role != "" {
		r.Use(func(c *gin.Context) {
			c.Set("user_id", int64(1))
			c.Set("role", role)
			c.Next()
		})
	}
	r.GET("/slots", h.List)
	r.POST("/slots", h.Create)
	r.PUT("/slots/:id", h.Update)
	r.DELETE("/slots/:id", h.Delete)
	return r
}

func doSlot(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
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

func TestSlotHandler_List_OK(t *testing.T) {
	svc := &fakeSlotSvc{listRet: []*entities.LessonSlot{{ID: 1, Number: 1, TimeStart: "08:30", TimeEnd: "10:00"}}}
	w := doSlot(setupSlotRouter(svc, "student"), http.MethodGet, "/slots", "")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "08:30")
}

func TestSlotHandler_Create_Forbidden(t *testing.T) {
	svc := &fakeSlotSvc{}
	w := doSlot(setupSlotRouter(svc, "teacher"), http.MethodPost, "/slots", `{"number":1,"time_start":"08:30","time_end":"10:00"}`)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, svc.createCnt, "denied request must not reach the service")
}

func TestSlotHandler_Create_OK(t *testing.T) {
	svc := &fakeSlotSvc{createRet: &entities.LessonSlot{ID: 5, Number: 1, TimeStart: "08:30", TimeEnd: "10:00"}}
	w := doSlot(setupSlotRouter(svc, "academic_secretary"), http.MethodPost, "/slots", `{"number":1,"time_start":"08:30","time_end":"10:00"}`)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"id":5`)
}

func TestSlotHandler_Create_ValidationError(t *testing.T) {
	svc := &fakeSlotSvc{createErr: entities.ErrInvalidSlotTimeRange}
	w := doSlot(setupSlotRouter(svc, "system_admin"), http.MethodPost, "/slots", `{"number":1,"time_start":"10:00","time_end":"08:30"}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSlotHandler_Create_DuplicateConflict(t *testing.T) {
	svc := &fakeSlotSvc{createErr: entities.ErrLessonSlotNumberTaken}
	w := doSlot(setupSlotRouter(svc, "system_admin"), http.MethodPost, "/slots", `{"number":1,"time_start":"08:30","time_end":"10:00"}`)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestSlotHandler_Create_BadBody(t *testing.T) {
	svc := &fakeSlotSvc{}
	w := doSlot(setupSlotRouter(svc, "system_admin"), http.MethodPost, "/slots", `{"number":0}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, 0, svc.createCnt)
}

func TestSlotHandler_Update_NotFound(t *testing.T) {
	svc := &fakeSlotSvc{updateErr: entities.ErrLessonSlotNotFound}
	w := doSlot(setupSlotRouter(svc, "system_admin"), http.MethodPut, "/slots/99", `{"number":1,"time_start":"08:30","time_end":"10:00"}`)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSlotHandler_Update_BadID(t *testing.T) {
	svc := &fakeSlotSvc{}
	w := doSlot(setupSlotRouter(svc, "system_admin"), http.MethodPut, "/slots/abc", `{"number":1,"time_start":"08:30","time_end":"10:00"}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSlotHandler_Delete_OK(t *testing.T) {
	svc := &fakeSlotSvc{}
	w := doSlot(setupSlotRouter(svc, "academic_secretary"), http.MethodDelete, "/slots/5", "")

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, 1, svc.deleteCnt)
}

func TestSlotHandler_Delete_NotFound(t *testing.T) {
	svc := &fakeSlotSvc{deleteErr: entities.ErrLessonSlotNotFound}
	w := doSlot(setupSlotRouter(svc, "system_admin"), http.MethodDelete, "/slots/99", "")

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSlotHandler_Delete_Forbidden(t *testing.T) {
	svc := &fakeSlotSvc{}
	w := doSlot(setupSlotRouter(svc, "student"), http.MethodDelete, "/slots/5", "")

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, 0, svc.deleteCnt)
	require.NotNil(t, svc)
}
