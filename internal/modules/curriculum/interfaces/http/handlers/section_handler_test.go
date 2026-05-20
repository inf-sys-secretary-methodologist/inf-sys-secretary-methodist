package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	curUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/interfaces/http/handlers"
)

// ===== Fakes for the five SectionHandler ports =====

type fakeCreateSectionPort struct {
	called   bool
	gotActor int64
	gotAdmin bool
	gotInput curUsecases.CreateSectionInput
	out      *entities.Section
	err      error
}

func (f *fakeCreateSectionPort) Execute(_ context.Context, actorID int64, isAdmin bool, in curUsecases.CreateSectionInput) (*entities.Section, error) {
	f.called = true
	f.gotActor = actorID
	f.gotAdmin = isAdmin
	f.gotInput = in
	return f.out, f.err
}

type fakeGetSectionPort struct {
	called bool
	gotID  int64
	out    *entities.Section
	err    error
}

func (f *fakeGetSectionPort) Execute(_ context.Context, id int64) (*entities.Section, error) {
	f.called = true
	f.gotID = id
	return f.out, f.err
}

type fakeListSectionsPort struct {
	called bool
	gotID  int64
	out    []*entities.Section
	err    error
}

func (f *fakeListSectionsPort) Execute(_ context.Context, curriculumID int64) ([]*entities.Section, error) {
	f.called = true
	f.gotID = curriculumID
	return f.out, f.err
}

type fakeUpdateSectionPort struct {
	called   bool
	gotActor int64
	gotAdmin bool
	gotInput curUsecases.UpdateSectionInput
	out      *entities.Section
	err      error
}

func (f *fakeUpdateSectionPort) Execute(_ context.Context, actorID int64, isAdmin bool, in curUsecases.UpdateSectionInput) (*entities.Section, error) {
	f.called = true
	f.gotActor = actorID
	f.gotAdmin = isAdmin
	f.gotInput = in
	return f.out, f.err
}

type fakeDeleteSectionPort struct {
	called   bool
	gotActor int64
	gotAdmin bool
	gotID    int64
	err      error
}

func (f *fakeDeleteSectionPort) Execute(_ context.Context, actorID int64, isAdmin bool, sectionID int64) error {
	f.called = true
	f.gotActor = actorID
	f.gotAdmin = isAdmin
	f.gotID = sectionID
	return f.err
}

// builtSection returns a freshly-reconstituted section for happy-path
// fakes to return.
func builtSection(t *testing.T, id, curriculumID int64) *entities.Section {
	t.Helper()
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	return entities.ReconstituteSection(id, curriculumID, "Базовая часть",
		"desc", 0, 0, now, now)
}

// setupSectionRouter wires a handler with the supplied ports + auth
// middleware and returns the engine ready for ServeHTTP. Routes mounted
// at the canonical paths — same shape main.go DI wiring will use.
func setupSectionRouter(
	t *testing.T,
	create handlers.CreateSectionPort,
	get handlers.GetSectionPort,
	list handlers.ListSectionsPort,
	update handlers.UpdateSectionPort,
	del handlers.DeleteSectionPort,
	uid int64,
	role string,
) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if uid != 0 || role != "" {
		r.Use(withAuth(uid, role))
	}
	h := handlers.NewSectionHandler(create, get, list, update, del)
	r.POST("/api/curricula/:curriculumID/sections", h.Create)
	r.GET("/api/curricula/:curriculumID/sections", h.List)
	r.GET("/api/sections/:sectionID", h.Get)
	r.PUT("/api/sections/:sectionID", h.Update)
	r.DELETE("/api/sections/:sectionID", h.Delete)
	return r
}

// stubAll returns a quintet of fakes wired to never-be-called state so
// route-under-test can swap in only the relevant fake.
func stubAll() (*fakeCreateSectionPort, *fakeGetSectionPort, *fakeListSectionsPort, *fakeUpdateSectionPort, *fakeDeleteSectionPort) {
	return &fakeCreateSectionPort{}, &fakeGetSectionPort{}, &fakeListSectionsPort{}, &fakeUpdateSectionPort{}, &fakeDeleteSectionPort{}
}

func doJSON(t *testing.T, r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if s, ok := body.(string); ok {
			req := httptest.NewRequest(method, path, strings.NewReader(s))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			return rec
		}
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// ===== Failure-closed wiring =====

func TestNewSectionHandler_PanicsOnNilPort(t *testing.T) {
	create, get, list, update, del := stubAll()
	cases := []struct {
		name string
		args [5]any
	}{
		{"nil create", [5]any{nil, get, list, update, del}},
		{"nil get", [5]any{create, nil, list, update, del}},
		{"nil list", [5]any{create, get, nil, update, del}},
		{"nil update", [5]any{create, get, list, nil, del}},
		{"nil delete", [5]any{create, get, list, update, nil}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("NewSectionHandler accepted nil port (%s)", tc.name)
				}
			}()
			handlers.NewSectionHandler(
				toCreatePort(tc.args[0]), toGetPort(tc.args[1]), toListPort(tc.args[2]),
				toUpdatePort(tc.args[3]), toDeletePort(tc.args[4]),
			)
		})
	}
}

// Trivial casters so the table can hold any-typed nil values.
func toCreatePort(v any) handlers.CreateSectionPort {
	if v == nil {
		return nil
	}
	return v.(handlers.CreateSectionPort)
}
func toGetPort(v any) handlers.GetSectionPort {
	if v == nil {
		return nil
	}
	return v.(handlers.GetSectionPort)
}
func toListPort(v any) handlers.ListSectionsPort {
	if v == nil {
		return nil
	}
	return v.(handlers.ListSectionsPort)
}
func toUpdatePort(v any) handlers.UpdateSectionPort {
	if v == nil {
		return nil
	}
	return v.(handlers.UpdateSectionPort)
}
func toDeletePort(v any) handlers.DeleteSectionPort {
	if v == nil {
		return nil
	}
	return v.(handlers.DeleteSectionPort)
}

// ===== Auth contract =====
//
// TestSectionHandler_RoleKeyContract pins that the handler reads
// c.Get("role") (production middleware contract — see v0.126.0 /
// v0.126.1 wrong-key bug class). Test injects production-shaped middleware;
// if the handler were to read c.Get("user_role") instead the panic would
// not happen but Auth would 401 here.
func TestSectionHandler_RoleKeyContract(t *testing.T) {
	create, get, list, update, del := stubAll()
	get.out = builtSection(t, 101, 7)
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")

	rec := doJSON(t, r, http.MethodGet, "/api/sections/101", nil)
	assert.Equal(t, http.StatusOK, rec.Code,
		"handler must read 'role' key (production middleware contract); got %d body=%s",
		rec.Code, rec.Body.String())
}

func TestSectionHandler_MissingAuth_Returns401(t *testing.T) {
	create, get, list, update, del := stubAll()
	r := setupSectionRouter(t, create, get, list, update, del, 0, "")
	rec := doJSON(t, r, http.MethodGet, "/api/sections/101", nil)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ===== Create =====

func TestSectionHandler_Create_HappyPath(t *testing.T) {
	create, get, list, update, del := stubAll()
	create.out = builtSection(t, 101, 7)

	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodPost, "/api/curricula/7/sections", handlers.CreateSectionRequest{
		Title:       "Базовая часть",
		Description: "desc",
		OrderIndex:  0,
	})
	require.Equal(t, http.StatusCreated, rec.Code, "body=%s", rec.Body.String())
	require.True(t, create.called)
	assert.Equal(t, int64(42), create.gotActor)
	assert.False(t, create.gotAdmin, "methodist must not be flagged as admin")
	assert.Equal(t, int64(7), create.gotInput.CurriculumID)
}

func TestSectionHandler_Create_RejectsNonWriteRoles(t *testing.T) {
	// v0.158.0: academic_secretary owns section authoring; methodist
	// is approver and must NOT write sections; teacher reads only;
	// student blocked at outer middleware но handler defense-in-depth
	// also denies.
	cases := []string{"methodist", "teacher", "student", "unknown"}
	for _, role := range cases {
		t.Run(role, func(t *testing.T) {
			create, get, list, update, del := stubAll()
			r := setupSectionRouter(t, create, get, list, update, del, 42, role)
			rec := doJSON(t, r, http.MethodPost, "/api/curricula/7/sections", handlers.CreateSectionRequest{Title: "T"})
			assert.Equal(t, http.StatusForbidden, rec.Code)
			assert.False(t, create.called)
		})
	}
}

func TestSectionHandler_Create_InvalidCurriculumID400(t *testing.T) {
	create, get, list, update, del := stubAll()
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodPost, "/api/curricula/abc/sections", handlers.CreateSectionRequest{Title: "T"})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSectionHandler_Create_InvalidJSON400(t *testing.T) {
	create, get, list, update, del := stubAll()
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodPost, "/api/curricula/7/sections", `{not-json`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.False(t, create.called)
}

func TestSectionHandler_Create_AdminOverridePropagatesIsAdmin(t *testing.T) {
	create, get, list, update, del := stubAll()
	create.out = builtSection(t, 101, 7)
	r := setupSectionRouter(t, create, get, list, update, del, 99, "system_admin")
	rec := doJSON(t, r, http.MethodPost, "/api/curricula/7/sections", handlers.CreateSectionRequest{Title: "T"})
	require.Equal(t, http.StatusCreated, rec.Code)
	assert.True(t, create.gotAdmin, "system_admin must propagate isAdmin=true to use case")
}

func TestSectionHandler_Create_CurriculumNotFound404(t *testing.T) {
	create, get, list, update, del := stubAll()
	create.err = repositories.ErrCurriculumNotFound
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodPost, "/api/curricula/7/sections", handlers.CreateSectionRequest{Title: "T"})
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestSectionHandler_Create_Forbidden_FromUseCase403(t *testing.T) {
	create, get, list, update, del := stubAll()
	create.err = entities.ErrSectionScopeForbidden
	r := setupSectionRouter(t, create, get, list, update, del, 99, "academic_secretary")
	rec := doJSON(t, r, http.MethodPost, "/api/curricula/7/sections", handlers.CreateSectionRequest{Title: "T"})
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestSectionHandler_Create_Frozen422(t *testing.T) {
	create, get, list, update, del := stubAll()
	create.err = entities.ErrCannotEditSection
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodPost, "/api/curricula/7/sections", handlers.CreateSectionRequest{Title: "T"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestSectionHandler_Create_Invalid422(t *testing.T) {
	create, get, list, update, del := stubAll()
	create.err = entities.ErrInvalidSection
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodPost, "/api/curricula/7/sections", handlers.CreateSectionRequest{Title: ""})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// ===== Get =====

func TestSectionHandler_Get_HappyPath(t *testing.T) {
	create, get, list, update, del := stubAll()
	get.out = builtSection(t, 101, 7)
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodGet, "/api/sections/101", nil)
	require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	assert.Equal(t, int64(101), get.gotID)
}

func TestSectionHandler_Get_NotFound(t *testing.T) {
	create, get, list, update, del := stubAll()
	get.err = repositories.ErrSectionNotFound
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodGet, "/api/sections/999", nil)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestSectionHandler_Get_Student403(t *testing.T) {
	create, get, list, update, del := stubAll()
	r := setupSectionRouter(t, create, get, list, update, del, 42, "student")
	rec := doJSON(t, r, http.MethodGet, "/api/sections/101", nil)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.False(t, get.called)
}

// ===== List =====

func TestSectionHandler_List_HappyPath(t *testing.T) {
	create, get, list, update, del := stubAll()
	list.out = []*entities.Section{
		builtSection(t, 101, 7),
		builtSection(t, 102, 7),
	}
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodGet, "/api/curricula/7/sections", nil)
	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Data handlers.SectionsListResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data.Items, 2)
}

func TestSectionHandler_List_EmptyResult(t *testing.T) {
	create, get, list, update, del := stubAll()
	list.out = nil
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodGet, "/api/curricula/7/sections", nil)
	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Data handlers.SectionsListResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotNil(t, resp.Data.Items, "items must be empty array, not null")
	assert.Len(t, resp.Data.Items, 0)
}

// ===== Update =====

func TestSectionHandler_Update_HappyPath(t *testing.T) {
	create, get, list, update, del := stubAll()
	update.out = builtSection(t, 101, 7)
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodPut, "/api/sections/101", handlers.UpdateSectionRequest{
		Title:      "Новый",
		OrderIndex: 1,
	})
	require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())
	assert.True(t, update.called)
	assert.Equal(t, int64(101), update.gotInput.ID)
	assert.Equal(t, "Новый", update.gotInput.Title)
	assert.Equal(t, 1, update.gotInput.OrderIndex)
}

func TestSectionHandler_Update_NotFound(t *testing.T) {
	create, get, list, update, del := stubAll()
	update.err = repositories.ErrSectionNotFound
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodPut, "/api/sections/999", handlers.UpdateSectionRequest{Title: "T"})
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestSectionHandler_Update_VersionConflict409(t *testing.T) {
	create, get, list, update, del := stubAll()
	update.err = repositories.ErrSectionVersionConflict
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodPut, "/api/sections/101", handlers.UpdateSectionRequest{Title: "T"})
	assert.Equal(t, http.StatusConflict, rec.Code,
		"version conflict must surface as 409 (optimistic-lock contract per ADR-3)")
}

func TestSectionHandler_Update_Forbidden403(t *testing.T) {
	create, get, list, update, del := stubAll()
	update.err = entities.ErrSectionScopeForbidden
	r := setupSectionRouter(t, create, get, list, update, del, 99, "academic_secretary")
	rec := doJSON(t, r, http.MethodPut, "/api/sections/101", handlers.UpdateSectionRequest{Title: "T"})
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestSectionHandler_Update_Frozen422(t *testing.T) {
	create, get, list, update, del := stubAll()
	update.err = entities.ErrCannotEditSection
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodPut, "/api/sections/101", handlers.UpdateSectionRequest{Title: "T"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestSectionHandler_Update_Invalid422(t *testing.T) {
	create, get, list, update, del := stubAll()
	update.err = entities.ErrInvalidSection
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodPut, "/api/sections/101", handlers.UpdateSectionRequest{Title: ""})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestSectionHandler_Update_Student403(t *testing.T) {
	create, get, list, update, del := stubAll()
	r := setupSectionRouter(t, create, get, list, update, del, 42, "student")
	rec := doJSON(t, r, http.MethodPut, "/api/sections/101", handlers.UpdateSectionRequest{Title: "T"})
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.False(t, update.called)
}

// ===== Delete =====

func TestSectionHandler_Delete_HappyPath(t *testing.T) {
	create, get, list, update, del := stubAll()
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodDelete, "/api/sections/101", nil)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.True(t, del.called)
	assert.Equal(t, int64(101), del.gotID)
}

func TestSectionHandler_Delete_NotFound(t *testing.T) {
	create, get, list, update, del := stubAll()
	del.err = repositories.ErrSectionNotFound
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodDelete, "/api/sections/999", nil)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestSectionHandler_Delete_Forbidden403(t *testing.T) {
	create, get, list, update, del := stubAll()
	del.err = entities.ErrSectionScopeForbidden
	r := setupSectionRouter(t, create, get, list, update, del, 99, "academic_secretary")
	rec := doJSON(t, r, http.MethodDelete, "/api/sections/101", nil)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestSectionHandler_Delete_Frozen422(t *testing.T) {
	create, get, list, update, del := stubAll()
	del.err = entities.ErrCannotEditSection
	r := setupSectionRouter(t, create, get, list, update, del, 42, "academic_secretary")
	rec := doJSON(t, r, http.MethodDelete, "/api/sections/101", nil)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestSectionHandler_Delete_Student403(t *testing.T) {
	create, get, list, update, del := stubAll()
	r := setupSectionRouter(t, create, get, list, update, del, 42, "student")
	rec := doJSON(t, r, http.MethodDelete, "/api/sections/101", nil)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.False(t, del.called)
}
