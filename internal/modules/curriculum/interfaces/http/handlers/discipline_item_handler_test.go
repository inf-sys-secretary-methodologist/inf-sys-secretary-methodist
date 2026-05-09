package handlers_test

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

	curUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/interfaces/http/handlers"
)

// postJSON / putJSON — local helpers (not shared с section_handler_test.go's
// doJSON to keep test files independent).
func postJSON(t *testing.T, r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, json.NewEncoder(&buf).Encode(body))
	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func putJSON(t *testing.T, r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, json.NewEncoder(&buf).Encode(body))
	req := httptest.NewRequest(http.MethodPut, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// ===== Fakes =====

type fakeCreateItemPort struct {
	out *entities.DisciplineItem
	err error
}

func (f *fakeCreateItemPort) Execute(_ context.Context, _ int64, _ bool, _ curUsecases.CreateDisciplineItemInput) (*entities.DisciplineItem, error) {
	return f.out, f.err
}

type fakeGetItemPort struct {
	out *entities.DisciplineItem
	err error
}

func (f *fakeGetItemPort) Execute(_ context.Context, _ int64) (*entities.DisciplineItem, error) {
	return f.out, f.err
}

type fakeListItemsPort struct {
	out []*entities.DisciplineItem
	err error
}

func (f *fakeListItemsPort) Execute(_ context.Context, _ int64) ([]*entities.DisciplineItem, error) {
	return f.out, f.err
}

type fakeUpdateItemPort struct {
	out *entities.DisciplineItem
	err error
}

func (f *fakeUpdateItemPort) Execute(_ context.Context, _ int64, _ bool, _ curUsecases.UpdateDisciplineItemInput) (*entities.DisciplineItem, error) {
	return f.out, f.err
}

type fakeDeleteItemPort struct {
	err    error
	called bool
}

func (f *fakeDeleteItemPort) Execute(_ context.Context, _ int64, _ bool, _ int64) error {
	f.called = true
	return f.err
}

func builtItem(t *testing.T, id, sectionID int64) *entities.DisciplineItem {
	t.Helper()
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	return entities.ReconstituteDisciplineItem(id, sectionID, "Математический анализ",
		36, 36, 0, 72, entities.ControlFormExam, 4, 1, 0, 0, now, now)
}

func setupItemRouter(
	t *testing.T,
	create handlers.CreateDisciplineItemPort,
	get handlers.GetDisciplineItemPort,
	list handlers.ListDisciplineItemsPort,
	update handlers.UpdateDisciplineItemPort,
	del handlers.DeleteDisciplineItemPort,
	uid int64,
	role string,
) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if uid != 0 || role != "" {
		r.Use(func(c *gin.Context) {
			if uid != 0 {
				c.Set("user_id", uid)
			}
			if role != "" {
				c.Set("role", role)
			}
			c.Next()
		})
	}
	h := handlers.NewDisciplineItemHandler(create, get, list, update, del)
	r.POST("/api/sections/:sectionID/items", h.Create)
	r.GET("/api/sections/:sectionID/items", h.List)
	r.GET("/api/items/:id", h.Get)
	r.PUT("/api/items/:id", h.Update)
	r.DELETE("/api/items/:id", h.Delete)
	return r
}

func itemStubs() (*fakeCreateItemPort, *fakeGetItemPort, *fakeListItemsPort, *fakeUpdateItemPort, *fakeDeleteItemPort) {
	return &fakeCreateItemPort{}, &fakeGetItemPort{}, &fakeListItemsPort{}, &fakeUpdateItemPort{}, &fakeDeleteItemPort{}
}

// ===== Failure-closed wiring =====

func TestNewDisciplineItemHandler_PanicsOnNilPort(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("did not panic on nil port")
		}
	}()
	handlers.NewDisciplineItemHandler(nil, &fakeGetItemPort{}, &fakeListItemsPort{}, &fakeUpdateItemPort{}, &fakeDeleteItemPort{})
}

// ===== Auth contract — production middleware key =====

func TestDisciplineItemHandler_RoleKeyContract(t *testing.T) {
	create, get, list, update, del := itemStubs()
	get.out = builtItem(t, 202, 11)
	r := setupItemRouter(t, create, get, list, update, del, 42, "methodist")

	req := httptest.NewRequest(http.MethodGet, "/api/items/202", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code,
		"handler must read 'role' key (production middleware contract); body=%s", rec.Body.String())
}

func TestDisciplineItemHandler_MissingAuth_Returns401(t *testing.T) {
	create, get, list, update, del := itemStubs()
	r := setupItemRouter(t, create, get, list, update, del, 0, "")
	req := httptest.NewRequest(http.MethodGet, "/api/items/202", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ===== Per-endpoint error mapping =====

func TestDisciplineItemHandler_Create_HappyPath(t *testing.T) {
	create, get, list, update, del := itemStubs()
	create.out = builtItem(t, 202, 11)
	r := setupItemRouter(t, create, get, list, update, del, 42, "methodist")
	body := map[string]any{
		"title":          "Математический анализ",
		"hours_lectures": 36,
		"hours_practice": 36,
		"hours_self":     72,
		"control_form":   "exam",
		"credits":        4,
		"semester":       1,
	}
	rec := postJSON(t, r, "/api/sections/11/items", body)
	require.Equal(t, http.StatusCreated, rec.Code, "body=%s", rec.Body.String())
}

func TestDisciplineItemHandler_Create_Student403(t *testing.T) {
	create, get, list, update, del := itemStubs()
	r := setupItemRouter(t, create, get, list, update, del, 42, "student")
	rec := postJSON(t, r, "/api/sections/11/items", map[string]any{"title": "T"})
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestDisciplineItemHandler_Get_NotFound(t *testing.T) {
	create, get, list, update, del := itemStubs()
	get.err = repositories.ErrDisciplineItemNotFound
	r := setupItemRouter(t, create, get, list, update, del, 42, "methodist")
	req := httptest.NewRequest(http.MethodGet, "/api/items/999", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDisciplineItemHandler_Update_VersionConflict409(t *testing.T) {
	create, get, list, update, del := itemStubs()
	update.err = repositories.ErrDisciplineItemVersionConflict
	r := setupItemRouter(t, create, get, list, update, del, 42, "methodist")
	rec := putJSON(t, r, "/api/items/202", map[string]any{
		"title": "T", "hours_lectures": 1, "control_form": "zachet", "credits": 1, "semester": 1,
	})
	assert.Equal(t, http.StatusConflict, rec.Code,
		"version conflict must surface as 409 (optimistic-lock contract per ADR-3)")
}

func TestDisciplineItemHandler_Update_Forbidden403(t *testing.T) {
	create, get, list, update, del := itemStubs()
	update.err = entities.ErrDisciplineItemScopeForbidden
	r := setupItemRouter(t, create, get, list, update, del, 99, "methodist")
	rec := putJSON(t, r, "/api/items/202", map[string]any{"title": "T"})
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestDisciplineItemHandler_Update_Frozen422(t *testing.T) {
	create, get, list, update, del := itemStubs()
	update.err = entities.ErrCannotEditDisciplineItem
	r := setupItemRouter(t, create, get, list, update, del, 42, "methodist")
	rec := putJSON(t, r, "/api/items/202", map[string]any{"title": "T"})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestDisciplineItemHandler_Delete_HappyPath(t *testing.T) {
	create, get, list, update, del := itemStubs()
	r := setupItemRouter(t, create, get, list, update, del, 42, "methodist")
	req := httptest.NewRequest(http.MethodDelete, "/api/items/202", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.True(t, del.called)
}

func TestDisciplineItemHandler_List_EmptyResult(t *testing.T) {
	create, get, list, update, del := itemStubs()
	list.out = nil
	r := setupItemRouter(t, create, get, list, update, del, 42, "methodist")
	req := httptest.NewRequest(http.MethodGet, "/api/sections/11/items", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Data handlers.DisciplineItemsListResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotNil(t, resp.Data.Items)
	assert.Len(t, resp.Data.Items, 0)
}
