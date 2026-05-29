package handlers

import (
	"context"
	"encoding/json"
	"net/http"
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

// ===== Fake ports =====

type fakeRecordOrder struct {
	result   *entities.MinobrnaukiOrder
	err      error
	called   bool
	gotIn    wpUsecases.RecordMinobrnaukiOrderInput
	gotActor int64
	gotRole  string
}

func (f *fakeRecordOrder) Execute(_ context.Context, actorID int64, role string, in wpUsecases.RecordMinobrnaukiOrderInput) (*entities.MinobrnaukiOrder, error) {
	f.called = true
	f.gotIn = in
	f.gotActor = actorID
	f.gotRole = role
	return f.result, f.err
}

type fakeGetOrder struct {
	result   *entities.MinobrnaukiOrder
	affected []int64
	err      error
	called   bool
	gotRole  string
	gotID    int64
}

func (f *fakeGetOrder) Execute(_ context.Context, role string, id int64) (*entities.MinobrnaukiOrder, []int64, error) {
	f.called = true
	f.gotRole = role
	f.gotID = id
	return f.result, f.affected, f.err
}

type fakeListOrders struct {
	result    repositories.MinobrnaukiOrderListResult
	err       error
	called    bool
	gotRole   string
	gotFilter repositories.MinobrnaukiOrderListFilter
}

func (f *fakeListOrders) Execute(_ context.Context, role string, filter repositories.MinobrnaukiOrderListFilter) (repositories.MinobrnaukiOrderListResult, error) {
	f.called = true
	f.gotRole = role
	f.gotFilter = filter
	return f.result, f.err
}

func sampleOrderEntity() *entities.MinobrnaukiOrder {
	return entities.ReconstituteMinobrnaukiOrder(entities.ReconstituteMinobrnaukiOrderInput{
		ID:          100,
		OrderNumber: "№ 1078",
		Title:       "Об изменении ФГОС 09.03.01",
		PublishedAt: time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC),
		ChangeScope: domain.MinobrnaukiOrderChangeScopeMajor,
		Summary:     "сводка",
		UploadedBy:  42,
		CreatedAt:   time.Date(2026, 5, 12, 8, 0, 0, 0, time.UTC),
	})
}

func newOrderRouter(rec *fakeRecordOrder, get *fakeGetOrder, list *fakeListOrders, mw ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if rec == nil {
		rec = &fakeRecordOrder{}
	}
	if get == nil {
		get = &fakeGetOrder{}
	}
	if list == nil {
		list = &fakeListOrders{}
	}
	h := NewMinobrnaukiOrderHandler(rec, get, list)
	api := r.Group("/api/v1")
	for _, m := range mw {
		api.Use(m)
	}
	RegisterMinobrnaukiOrderRoutes(api, h)
	return r
}

func validRecordBody() RecordMinobrnaukiOrderRequest {
	return RecordMinobrnaukiOrderRequest{
		OrderNumber:            "№ 1078 от 12.05.2026",
		Title:                  "Об изменении ФГОС 09.03.01",
		PublishedAt:            "2026-05-12",
		ChangeScope:            "major",
		Summary:                "Обновлён перечень компетенций",
		AffectedWorkProgramIDs: []int64{11, 22},
	}
}

// ===== Record =====

func TestMinobrnaukiOrderHandler_Record_HappyPath(t *testing.T) {
	rec := &fakeRecordOrder{result: sampleOrderEntity()}
	r := newOrderRouter(rec, nil, nil, withAuth(42, "methodist"))

	w := doJSON(t, r, http.MethodPost, "/api/v1/minobrnauki-orders", validRecordBody())

	assert.Equal(t, http.StatusCreated, w.Code)
	require.True(t, rec.called)
	assert.Equal(t, int64(42), rec.gotActor, "uploaded_by derives from JWT, not body")
	assert.Equal(t, "methodist", rec.gotRole)
	assert.Equal(t, "major", rec.gotIn.ChangeScope)
	assert.Equal(t, []int64{11, 22}, rec.gotIn.AffectedWorkProgramIDs)
	assert.True(t, rec.gotIn.PublishedAt.Equal(time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC)))

	var env map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	data := env["data"].(map[string]any)
	assert.Equal(t, "№ 1078", data["order_number"])
	assert.Equal(t, []any{float64(11), float64(22)}, data["affected_work_program_ids"])
}

func TestMinobrnaukiOrderHandler_Record_Unauthorized(t *testing.T) {
	r := newOrderRouter(nil, nil, nil) // no withAuth
	w := doJSON(t, r, http.MethodPost, "/api/v1/minobrnauki-orders", validRecordBody())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMinobrnaukiOrderHandler_Record_BadBody(t *testing.T) {
	rec := &fakeRecordOrder{}
	r := newOrderRouter(rec, nil, nil, withAuth(42, "methodist"))
	body := validRecordBody()
	body.OrderNumber = "" // violates binding:"required"
	w := doJSON(t, r, http.MethodPost, "/api/v1/minobrnauki-orders", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, rec.called)
}

func TestMinobrnaukiOrderHandler_Record_BadDate(t *testing.T) {
	rec := &fakeRecordOrder{}
	r := newOrderRouter(rec, nil, nil, withAuth(42, "methodist"))
	body := validRecordBody()
	body.PublishedAt = "12.05.2026" // wrong format
	w := doJSON(t, r, http.MethodPost, "/api/v1/minobrnauki-orders", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, rec.called, "must not reach usecase on unparseable date")
}

func TestMinobrnaukiOrderHandler_Record_Forbidden(t *testing.T) {
	rec := &fakeRecordOrder{err: domain.ErrMinobrnaukiOrderScopeForbidden}
	r := newOrderRouter(rec, nil, nil, withAuth(7, "teacher"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/minobrnauki-orders", validRecordBody())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMinobrnaukiOrderHandler_Record_Invalid(t *testing.T) {
	rec := &fakeRecordOrder{err: domain.ErrInvalidMinobrnaukiOrder}
	r := newOrderRouter(rec, nil, nil, withAuth(42, "methodist"))
	w := doJSON(t, r, http.MethodPost, "/api/v1/minobrnauki-orders", validRecordBody())
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

// ===== Get =====

func TestMinobrnaukiOrderHandler_Get_HappyPath(t *testing.T) {
	get := &fakeGetOrder{result: sampleOrderEntity(), affected: []int64{11, 22}}
	r := newOrderRouter(nil, get, nil, withAuth(7, "teacher"))

	w := doJSON(t, r, http.MethodGet, "/api/v1/minobrnauki-orders/100", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.True(t, get.called)
	assert.Equal(t, int64(100), get.gotID)
	assert.Equal(t, "teacher", get.gotRole)

	var env map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	data := env["data"].(map[string]any)
	assert.Equal(t, "2026-05-12", data["published_at"])
	assert.Equal(t, []any{float64(11), float64(22)}, data["affected_work_program_ids"])
}

func TestMinobrnaukiOrderHandler_Get_BadID(t *testing.T) {
	get := &fakeGetOrder{}
	r := newOrderRouter(nil, get, nil, withAuth(7, "teacher"))
	w := doJSON(t, r, http.MethodGet, "/api/v1/minobrnauki-orders/0", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, get.called)
}

func TestMinobrnaukiOrderHandler_Get_NotFound(t *testing.T) {
	get := &fakeGetOrder{err: repositories.ErrMinobrnaukiOrderNotFound}
	r := newOrderRouter(nil, get, nil, withAuth(7, "teacher"))
	w := doJSON(t, r, http.MethodGet, "/api/v1/minobrnauki-orders/999", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMinobrnaukiOrderHandler_Get_StudentForbidden(t *testing.T) {
	get := &fakeGetOrder{err: domain.ErrMinobrnaukiOrderScopeForbidden}
	r := newOrderRouter(nil, get, nil, withAuth(5, "student"))
	w := doJSON(t, r, http.MethodGet, "/api/v1/minobrnauki-orders/100", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ===== List =====

func TestMinobrnaukiOrderHandler_List_HappyPath_ParsesFilter(t *testing.T) {
	list := &fakeListOrders{result: repositories.MinobrnaukiOrderListResult{
		Items: []repositories.MinobrnaukiOrderListItem{
			{ID: 1, OrderNumber: "A", PublishedAt: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), ChangeScope: domain.MinobrnaukiOrderChangeScopeMinor},
		},
		Total: 1,
	}}
	r := newOrderRouter(nil, nil, list, withAuth(42, "methodist"))

	w := doJSON(t, r, http.MethodGet, "/api/v1/minobrnauki-orders?change_scope=major&uploaded_by=42&limit=10&offset=5", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	require.True(t, list.called)
	require.NotNil(t, list.gotFilter.ChangeScope)
	assert.Equal(t, domain.MinobrnaukiOrderChangeScopeMajor, *list.gotFilter.ChangeScope)
	require.NotNil(t, list.gotFilter.UploadedBy)
	assert.Equal(t, int64(42), *list.gotFilter.UploadedBy)
	assert.Equal(t, 10, list.gotFilter.Limit)
	assert.Equal(t, 5, list.gotFilter.Offset)

	var env map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	data := env["data"].(map[string]any)
	assert.Equal(t, float64(1), data["total"])
}

func TestMinobrnaukiOrderHandler_List_StudentForbidden(t *testing.T) {
	list := &fakeListOrders{err: domain.ErrMinobrnaukiOrderScopeForbidden}
	r := newOrderRouter(nil, nil, list, withAuth(5, "student"))
	w := doJSON(t, r, http.MethodGet, "/api/v1/minobrnauki-orders", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestNewMinobrnaukiOrderHandler_PanicsOnNilPort(t *testing.T) {
	defer func() {
		if rec := recover(); rec == nil {
			t.Fatal("NewMinobrnaukiOrderHandler(nil...) did not panic")
		}
	}()
	NewMinobrnaukiOrderHandler(nil, &fakeGetOrder{}, &fakeListOrders{})
}
