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
)

type fakeGenerateSvc struct {
	preview *usecases.SchedulePreview
	apply   *usecases.ApplyResult
	err     error
}

func (f *fakeGenerateSvc) Preview(_ context.Context, _ usecases.GenerateParams) (*usecases.SchedulePreview, error) {
	return f.preview, f.err
}

func (f *fakeGenerateSvc) Apply(_ context.Context, _ usecases.GenerateParams) (*usecases.ApplyResult, error) {
	return f.apply, f.err
}

func setupGenerateRouter(h *GenerateScheduleHandler, role string) *gin.Engine {
	r := gin.New()
	r.POST("/schedule/generate", withLessonAuth(h.Preview, 1, role))
	r.POST("/schedule/generate/apply", withLessonAuth(h.Apply, 1, role))
	return r
}

func postJSON(r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestGenerateHandler_Preview_ForbiddenForStudent(t *testing.T) {
	h := NewGenerateScheduleHandler(&fakeGenerateSvc{})
	r := setupGenerateRouter(h, "student")
	w := postJSON(r, "/schedule/generate", `{"semester_id":1}`)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGenerateHandler_Preview_HappyPath(t *testing.T) {
	svc := &fakeGenerateSvc{preview: &usecases.SchedulePreview{
		Lessons:        []usecases.GeneratedLesson{{LoadID: 1, GroupName: "ПИ-101"}},
		TotalRequested: 1,
		PlacedCount:    1,
	}}
	h := NewGenerateScheduleHandler(svc)
	r := setupGenerateRouter(h, "academic_secretary")

	w := postJSON(r, "/schedule/generate", `{"semester_id":1}`)
	require.Equal(t, http.StatusOK, w.Code)
	assertDataIsObject(t, decodeEnvelope(t, w.Body.String()))
}

func TestGenerateHandler_Preview_InvalidBody(t *testing.T) {
	h := NewGenerateScheduleHandler(&fakeGenerateSvc{})
	r := setupGenerateRouter(h, "methodist")
	w := postJSON(r, "/schedule/generate", `{not json`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGenerateHandler_Preview_MapsInvalidInput(t *testing.T) {
	h := NewGenerateScheduleHandler(&fakeGenerateSvc{err: usecases.ErrInvalidInput})
	r := setupGenerateRouter(h, "methodist")
	w := postJSON(r, "/schedule/generate", `{"semester_id":0}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGenerateHandler_Apply_ForbiddenForTeacher(t *testing.T) {
	h := NewGenerateScheduleHandler(&fakeGenerateSvc{})
	r := setupGenerateRouter(h, "teacher")
	w := postJSON(r, "/schedule/generate/apply", `{"semester_id":1}`)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGenerateHandler_Apply_HappyPath(t *testing.T) {
	svc := &fakeGenerateSvc{apply: &usecases.ApplyResult{Created: 2, Unplaced: 0}}
	h := NewGenerateScheduleHandler(svc)
	r := setupGenerateRouter(h, "academic_secretary")

	w := postJSON(r, "/schedule/generate/apply", `{"semester_id":1}`)
	require.Equal(t, http.StatusOK, w.Code)
	assertDataIsObject(t, decodeEnvelope(t, w.Body.String()))
}

func TestGenerateHandler_Apply_ConflictWhenExists(t *testing.T) {
	svc := &fakeGenerateSvc{err: usecases.ErrScheduleAlreadyExists}
	h := NewGenerateScheduleHandler(svc)
	r := setupGenerateRouter(h, "academic_secretary")

	w := postJSON(r, "/schedule/generate/apply", `{"semester_id":1}`)
	assert.Equal(t, http.StatusConflict, w.Code)
}
