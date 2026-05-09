package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	curUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/interfaces/http/handlers"
)

// ===== Bulk-edit fake port =====

type fakeBulkEditPort struct {
	out    *curUsecases.BulkEditDisciplineItemsResult
	err    error
	gotIn  curUsecases.BulkEditDisciplineItemsInput
	called bool
}

func (f *fakeBulkEditPort) Execute(_ context.Context, _ int64, _ bool, in curUsecases.BulkEditDisciplineItemsInput) (*curUsecases.BulkEditDisciplineItemsResult, error) {
	f.called = true
	f.gotIn = in
	return f.out, f.err
}

// ===== Router setup =====

func setupBulkRouter(t *testing.T, bulk handlers.BulkEditDisciplineItemsPort, uid int64, role string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if uid != 0 || role != "" {
		r.Use(withAuth(uid, role))
	}
	h := handlers.NewBulkDisciplineItemsHandler(bulk)
	r.POST("/api/sections/:sectionID/items/bulk", h.BulkEdit)
	return r
}

// ===== Constructor =====

func TestNewBulkDisciplineItemsHandler_PanicsOnNilPort(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("constructor accepted nil port")
		}
	}()
	handlers.NewBulkDisciplineItemsHandler(nil)
}

// ===== Auth gates =====

func TestBulkDisciplineItemsHandler_MissingAuth_Returns401(t *testing.T) {
	bulk := &fakeBulkEditPort{}
	r := setupBulkRouter(t, bulk, 0, "")
	rec := postJSON(t, r, "/api/sections/11/items/bulk", handlers.BulkEditRequest{
		Creates: []handlers.BulkCreateItemRequest{{Title: "X"}},
	})
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, bulk.called)
}

func TestBulkDisciplineItemsHandler_StudentRole_Returns403(t *testing.T) {
	bulk := &fakeBulkEditPort{}
	r := setupBulkRouter(t, bulk, 42, "student")
	rec := postJSON(t, r, "/api/sections/11/items/bulk", handlers.BulkEditRequest{
		Creates: []handlers.BulkCreateItemRequest{{Title: "X"}},
	})
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.False(t, bulk.called)
}

// ===== Path validation =====

func TestBulkDisciplineItemsHandler_BadSectionID_Returns400(t *testing.T) {
	bulk := &fakeBulkEditPort{}
	r := setupBulkRouter(t, bulk, 42, "methodist")
	rec := postJSON(t, r, "/api/sections/abc/items/bulk", handlers.BulkEditRequest{
		Creates: []handlers.BulkCreateItemRequest{{Title: "X"}},
	})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.False(t, bulk.called)
}

func TestBulkDisciplineItemsHandler_BadJSON_Returns400(t *testing.T) {
	bulk := &fakeBulkEditPort{}
	r := setupBulkRouter(t, bulk, 42, "methodist")
	req := httptest.NewRequest(http.MethodPost, "/api/sections/11/items/bulk",
		bytes.NewReader([]byte("{not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.False(t, bulk.called)
}

// ===== Happy path =====

func TestBulkDisciplineItemsHandler_HappyPath_AllOps(t *testing.T) {
	created := builtItem(t, 250, 11)
	updated := builtItem(t, 202, 11)
	bulk := &fakeBulkEditPort{
		out: &curUsecases.BulkEditDisciplineItemsResult{
			Created: []*entities.DisciplineItem{created},
			Updated: []*entities.DisciplineItem{updated},
			Deleted: []int64{203},
		},
	}
	r := setupBulkRouter(t, bulk, 42, "methodist")
	rec := postJSON(t, r, "/api/sections/11/items/bulk", handlers.BulkEditRequest{
		Creates: []handlers.BulkCreateItemRequest{{
			Title: "Новая", HoursLectures: 36, HoursPractice: 36, HoursSelf: 72,
			ControlForm: "exam", Credits: 4, Semester: 1,
		}},
		Updates: []handlers.BulkUpdateItemRequest{{
			ID: 202, Title: "Обновлённая",
			HoursLectures: 18, HoursPractice: 18, HoursSelf: 36,
			ControlForm: "zachet", Credits: 2, Semester: 1,
		}},
		Deletes: []int64{203},
	})
	require.Equal(t, http.StatusOK, rec.Code, "body=%s", rec.Body.String())

	var resp struct {
		Data handlers.BulkEditSuccessResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data.Created, 1)
	assert.Equal(t, int64(250), resp.Data.Created[0].ID)
	assert.Len(t, resp.Data.Updated, 1)
	assert.Equal(t, int64(202), resp.Data.Updated[0].ID)
	assert.Equal(t, []int64{203}, resp.Data.Deleted)

	assert.True(t, bulk.called)
	assert.Equal(t, int64(11), bulk.gotIn.SectionID)
	assert.Len(t, bulk.gotIn.Creates, 1)
	assert.Len(t, bulk.gotIn.Updates, 1)
	assert.Equal(t, []int64{203}, bulk.gotIn.Deletes)
}

// ===== Version conflict 409 =====

func TestBulkDisciplineItemsHandler_VersionConflict_Returns409WithConflicts(t *testing.T) {
	bulk := &fakeBulkEditPort{
		out: &curUsecases.BulkEditDisciplineItemsResult{
			Conflicts: []curUsecases.BulkEditConflict{
				{ID: 202, ExpectedVersion: 5, CurrentVersion: 7},
				{ID: 204, ExpectedVersion: 3, CurrentVersion: 4},
			},
		},
		err: curUsecases.ErrBulkVersionConflict,
	}
	r := setupBulkRouter(t, bulk, 42, "methodist")
	rec := postJSON(t, r, "/api/sections/11/items/bulk", handlers.BulkEditRequest{
		Updates: []handlers.BulkUpdateItemRequest{
			{ID: 202, Title: "X"},
			{ID: 204, Title: "Y"},
		},
	})
	require.Equal(t, http.StatusConflict, rec.Code, "body=%s", rec.Body.String())

	var conflict handlers.BulkEditConflictResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &conflict))
	assert.Equal(t, "VERSION_CONFLICT", conflict.Error)
	require.Len(t, conflict.Conflicts, 2,
		"collect-all per ADR-12 — both conflicts reported")
	assert.Equal(t, int64(202), conflict.Conflicts[0].ID)
	assert.Equal(t, 5, conflict.Conflicts[0].ExpectedVersion)
	assert.Equal(t, 7, conflict.Conflicts[0].CurrentVersion)
	assert.Equal(t, int64(204), conflict.Conflicts[1].ID)
}

// ===== Sentinel mapping =====

func TestBulkDisciplineItemsHandler_EmptyInput_Returns422(t *testing.T) {
	bulk := &fakeBulkEditPort{err: curUsecases.ErrEmptyBulkInput}
	r := setupBulkRouter(t, bulk, 42, "methodist")
	rec := postJSON(t, r, "/api/sections/11/items/bulk", handlers.BulkEditRequest{})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, rec.Body.String(), "EMPTY_BULK_INPUT")
}

func TestBulkDisciplineItemsHandler_CrossSection_Returns422(t *testing.T) {
	bulk := &fakeBulkEditPort{err: curUsecases.ErrCrossSectionBulkEdit}
	r := setupBulkRouter(t, bulk, 42, "methodist")
	rec := postJSON(t, r, "/api/sections/11/items/bulk", handlers.BulkEditRequest{
		Updates: []handlers.BulkUpdateItemRequest{{ID: 999, Title: "X"}},
	})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, rec.Body.String(), "CROSS_SECTION_BULK_EDIT")
}

func TestBulkDisciplineItemsHandler_SectionNotFound_Returns404(t *testing.T) {
	bulk := &fakeBulkEditPort{err: repositories.ErrSectionNotFound}
	r := setupBulkRouter(t, bulk, 42, "methodist")
	rec := postJSON(t, r, "/api/sections/999/items/bulk", handlers.BulkEditRequest{
		Creates: []handlers.BulkCreateItemRequest{{Title: "X"}},
	})
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestBulkDisciplineItemsHandler_Forbidden_Returns403(t *testing.T) {
	bulk := &fakeBulkEditPort{err: entities.ErrDisciplineItemScopeForbidden}
	r := setupBulkRouter(t, bulk, 99, "methodist")
	rec := postJSON(t, r, "/api/sections/11/items/bulk", handlers.BulkEditRequest{
		Creates: []handlers.BulkCreateItemRequest{{Title: "X"}},
	})
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestBulkDisciplineItemsHandler_NotEditable_Returns422(t *testing.T) {
	bulk := &fakeBulkEditPort{err: entities.ErrCannotEditDisciplineItem}
	r := setupBulkRouter(t, bulk, 42, "methodist")
	rec := postJSON(t, r, "/api/sections/11/items/bulk", handlers.BulkEditRequest{
		Creates: []handlers.BulkCreateItemRequest{{Title: "X"}},
	})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, rec.Body.String(), "NOT_EDITABLE")
}

func TestBulkDisciplineItemsHandler_InvalidInput_Returns422(t *testing.T) {
	bulk := &fakeBulkEditPort{err: entities.ErrInvalidDisciplineItem}
	r := setupBulkRouter(t, bulk, 42, "methodist")
	rec := postJSON(t, r, "/api/sections/11/items/bulk", handlers.BulkEditRequest{
		Creates: []handlers.BulkCreateItemRequest{{Title: ""}},
	})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, rec.Body.String(), "INVALID_INPUT")
}
