package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wpUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// ===== Fake usecase ports =====

type fakeCreate struct {
	result   *entities.WorkProgram
	err      error
	called   bool
	gotIn    wpUsecases.CreateWorkProgramInput
	gotActor int64
	gotRole  string
}

func (f *fakeCreate) Execute(_ context.Context, actorID int64, role string, in wpUsecases.CreateWorkProgramInput) (*entities.WorkProgram, error) {
	f.called = true
	f.gotIn = in
	f.gotActor = actorID
	f.gotRole = role
	return f.result, f.err
}

type fakeGet struct {
	result *entities.WorkProgram
	err    error
	called bool
}

func (f *fakeGet) Execute(_ context.Context, _ int64, _ string, _ wpUsecases.GetWorkProgramInput) (*entities.WorkProgram, error) {
	f.called = true
	return f.result, f.err
}

type fakeList struct {
	result wpUsecases.ListWorkProgramsResult
	err    error
	called bool
}

func (f *fakeList) Execute(_ context.Context, _ int64, _ string, _ wpUsecases.ListWorkProgramsInput) (wpUsecases.ListWorkProgramsResult, error) {
	f.called = true
	return f.result, f.err
}

// withAuth pre-sets user_id + role в the gin context — mirrors what the
// RequireAuth middleware does in production. Pinning the exact context
// keys (`user_id`, `role`) catches drift per
// feedback_handler_context_key_must_match_middleware.
func withAuth(userID int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

func sampleWP(t *testing.T) *entities.WorkProgram {
	t.Helper()
	ts := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	return entities.ReconstituteWorkProgram(entities.ReconstituteWorkProgramInput{
		ID:                 99,
		DisciplineID:       7,
		SpecialtyCode:      "09.03.01",
		ApplicableFromYear: 2026,
		Title:              "Базы данных",
		Annotation:         "Аннотация",
		Status:             domain.StatusDraft,
		AuthorID:           42,
		Version:            0,
		CreatedAt:          ts,
		UpdatedAt:          ts,
	})
}

func newRouter(fc *fakeCreate, fg *fakeGet, fl *fakeList, mw ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewWorkProgramHandler(fc, fg, fl)
	api := r.Group("/api/v1")
	for _, m := range mw {
		api.Use(m)
	}
	RegisterWorkProgramRoutes(api, h)
	return r
}

func doJSON(t *testing.T, r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func validCreateBody() CreateWorkProgramRequest {
	return CreateWorkProgramRequest{
		DisciplineID:       7,
		SpecialtyCode:      "09.03.01",
		ApplicableFromYear: 2026,
		Title:              "Базы данных",
		Annotation:         "Аннотация",
	}
}

// ===== Create =====

func TestWorkProgramHandler_Create_HappyPath(t *testing.T) {
	fc := &fakeCreate{result: sampleWP(t)}
	r := newRouter(fc, &fakeGet{}, &fakeList{}, withAuth(42, "teacher"))

	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs", validCreateBody())

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.True(t, fc.called)
	// Author + role derive from JWT context, not request body.
	assert.Equal(t, int64(42), fc.gotActor)
	assert.Equal(t, "teacher", fc.gotRole)
	assert.Equal(t, int64(7), fc.gotIn.DisciplineID)
}

func TestWorkProgramHandler_Create_Unauthenticated(t *testing.T) {
	r := newRouter(&fakeCreate{}, &fakeGet{}, &fakeList{}) // no withAuth
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs", validCreateBody())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWorkProgramHandler_Create_InvalidBodyMaps400(t *testing.T) {
	r := newRouter(&fakeCreate{}, &fakeGet{}, &fakeList{}, withAuth(42, "teacher"))
	// Missing required title / specialty_code.
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs", CreateWorkProgramRequest{
		DisciplineID: 7,
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWorkProgramHandler_Create_ForbiddenMaps403(t *testing.T) {
	// Create is a collection POST — a role-based denial is a true 403
	// (no resource id to enumerate, so no IDOR collapse here).
	fc := &fakeCreate{err: domain.ErrWorkProgramScopeForbidden}
	r := newRouter(fc, &fakeGet{}, &fakeList{}, withAuth(7, "student"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs", validCreateBody())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestWorkProgramHandler_Create_InvalidWorkProgramMaps422(t *testing.T) {
	fc := &fakeCreate{err: domain.ErrInvalidWorkProgram}
	r := newRouter(fc, &fakeGet{}, &fakeList{}, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs", validCreateBody())
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestWorkProgramHandler_Create_IdentityExistsMaps409(t *testing.T) {
	fc := &fakeCreate{err: repositories.ErrWorkProgramIdentityExists}
	r := newRouter(fc, &fakeGet{}, &fakeList{}, withAuth(42, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/work-programs", validCreateBody())
	assert.Equal(t, http.StatusConflict, w.Code)
}

// ===== Get =====

func TestWorkProgramHandler_Get_HappyPath(t *testing.T) {
	fg := &fakeGet{result: sampleWP(t)}
	r := newRouter(&fakeCreate{}, fg, &fakeList{}, withAuth(7, "student"))
	w := doJSON(t, r, http.MethodGet, "/api/v1/work-programs/99", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, fg.called)
}

func TestWorkProgramHandler_Get_Unauthenticated(t *testing.T) {
	r := newRouter(&fakeCreate{}, &fakeGet{}, &fakeList{})
	w := doJSON(t, r, http.MethodGet, "/api/v1/work-programs/99", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWorkProgramHandler_Get_InvalidIDMaps400(t *testing.T) {
	r := newRouter(&fakeCreate{}, &fakeGet{}, &fakeList{}, withAuth(7, "student"))
	w := doJSON(t, r, http.MethodGet, "/api/v1/work-programs/abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWorkProgramHandler_Get_NotFoundMaps404(t *testing.T) {
	fg := &fakeGet{err: repositories.ErrWorkProgramNotFound}
	r := newRouter(&fakeCreate{}, fg, &fakeList{}, withAuth(7, "student"))
	w := doJSON(t, r, http.MethodGet, "/api/v1/work-programs/404", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// IDOR mitigation: a non-admin denied by scope must see 404, not 403,
// so resource ids cannot be enumerated by privilege class (OWASP).
func TestWorkProgramHandler_Get_ForbiddenHiddenAs404ForNonAdmin(t *testing.T) {
	fg := &fakeGet{err: domain.ErrWorkProgramScopeForbidden}
	r := newRouter(&fakeCreate{}, fg, &fakeList{}, withAuth(7, "teacher"))
	w := doJSON(t, r, http.MethodGet, "/api/v1/work-programs/99", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// Admins are entitled to know the resource exists — they keep the 403
// signal (they can see all РПД per ADR-5, so this is defensive).
func TestWorkProgramHandler_Get_ForbiddenStays403ForAdmin(t *testing.T) {
	fg := &fakeGet{err: domain.ErrWorkProgramScopeForbidden}
	r := newRouter(&fakeCreate{}, fg, &fakeList{}, withAuth(1, "system_admin"))
	w := doJSON(t, r, http.MethodGet, "/api/v1/work-programs/99", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
