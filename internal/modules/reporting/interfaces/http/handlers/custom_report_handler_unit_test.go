package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/interfaces/http/handlers"
)

func newCustomReportHandlerUnit() *handlers.CustomReportHandler {
	uc := usecases.NewCustomReportUseCase(nil, nil)
	return handlers.NewCustomReportHandler(uc)
}

func TestCustomReportHandler_Create_Unit(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.POST("/reports/custom", h.Create)

		w := doRequest(router, http.MethodPost, "/reports/custom", map[string]interface{}{
			"name":        "Report",
			"data_source": "documents",
			"fields":      []map[string]interface{}{{"field_key": "id", "display_name": "ID", "order": 1}},
		})
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.POST("/reports/custom", withAuthUser(1, "methodist"), h.Create)

		w := doRequest(router, http.MethodPost, "/reports/custom", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCustomReportHandler_GetByID_Unit(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.GET("/reports/custom/:id", h.GetByID)

		w := doRequest(router, http.MethodGet, "/reports/custom/550e8400-e29b-41d4-a716-446655440000", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.GET("/reports/custom/:id", withAuthUser(1, "methodist"), h.GetByID)

		w := doRequest(router, http.MethodGet, "/reports/custom/not-a-uuid", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCustomReportHandler_Update_Unit(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.PUT("/reports/custom/:id", h.Update)

		w := doRequest(router, http.MethodPut, "/reports/custom/550e8400-e29b-41d4-a716-446655440000", map[string]interface{}{"name": "Updated"})
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.PUT("/reports/custom/:id", withAuthUser(1, "methodist"), h.Update)

		w := doRequest(router, http.MethodPut, "/reports/custom/bad-uuid", map[string]interface{}{"name": "Updated"})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.PUT("/reports/custom/:id", withAuthUser(1, "methodist"), h.Update)

		w := doRequest(router, http.MethodPut, "/reports/custom/550e8400-e29b-41d4-a716-446655440000", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCustomReportHandler_Delete_Unit(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.DELETE("/reports/custom/:id", h.Delete)

		w := doRequest(router, http.MethodDelete, "/reports/custom/550e8400-e29b-41d4-a716-446655440000", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.DELETE("/reports/custom/:id", withAuthUser(1, "methodist"), h.Delete)

		w := doRequest(router, http.MethodDelete, "/reports/custom/bad-uuid", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCustomReportHandler_List_Unit(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.GET("/reports/custom", h.List)

		w := doRequest(router, http.MethodGet, "/reports/custom", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestCustomReportHandler_Execute_Unit(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.POST("/reports/custom/:id/execute", h.Execute)

		w := doRequest(router, http.MethodPost, "/reports/custom/550e8400-e29b-41d4-a716-446655440000/execute", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.POST("/reports/custom/:id/execute", withAuthUser(1, "methodist"), h.Execute)

		w := doRequest(router, http.MethodPost, "/reports/custom/bad-uuid/execute", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCustomReportHandler_Export_Unit(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.POST("/reports/custom/:id/export", h.Export)

		w := doRequest(router, http.MethodPost, "/reports/custom/550e8400-e29b-41d4-a716-446655440000/export", map[string]interface{}{"format": "xlsx"})
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.POST("/reports/custom/:id/export", withAuthUser(1, "methodist"), h.Export)

		w := doRequest(router, http.MethodPost, "/reports/custom/bad-uuid/export", map[string]interface{}{"format": "xlsx"})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.POST("/reports/custom/:id/export", withAuthUser(1, "methodist"), h.Export)

		w := doRequest(router, http.MethodPost, "/reports/custom/550e8400-e29b-41d4-a716-446655440000/export", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCustomReportHandler_GetMyReports_Unit(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.GET("/reports/custom/my", h.GetMyReports)

		w := doRequest(router, http.MethodGet, "/reports/custom/my", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestCustomReportHandler_GetAvailableFields_Unit(t *testing.T) {
	h := newCustomReportHandlerUnit()
	router := newRouter()
	router.GET("/reports/custom/fields", h.GetAvailableFields)

	w := doRequest(router, http.MethodGet, "/reports/custom/fields", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCustomReportHandler_getUserID_Types(t *testing.T) {
	h := newCustomReportHandlerUnit()

	t.Run("int type", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Set("user_id", int(42))

		// getUserID succeeds with int type, but usecase panics on nil repos
		// Use recover to verify getUserID passed (no 401)
		assert.Panics(t, func() { h.GetMyReports(c) })
		assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("uint64 type", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Set("user_id", uint64(42))

		assert.Panics(t, func() { h.GetMyReports(c) })
		assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("missing user_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

		h.GetMyReports(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("unsupported type string", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Set("user_id", "not-a-number")

		h.GetMyReports(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestCustomReportHandler_GetPublicReports_Unit(t *testing.T) {
	t.Run("panics on nil usecase", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.GET("/reports/custom/public", h.GetPublicReports)

		assert.Panics(t, func() {
			doRequest(router, http.MethodGet, "/reports/custom/public", nil)
		})
	})

	t.Run("with custom page params", func(t *testing.T) {
		h := newCustomReportHandlerUnit()
		router := newRouter()
		router.GET("/reports/custom/public", h.GetPublicReports)

		assert.Panics(t, func() {
			doRequest(router, http.MethodGet, "/reports/custom/public?page=2&pageSize=5", nil)
		})
	})
}

func TestCustomReportHandler_Execute_WithQueryParams(t *testing.T) {
	h := newCustomReportHandlerUnit()
	router := newRouter()
	router.POST("/reports/custom/:id/execute", withAuthUser(1, "methodist"), h.Execute)

	// Valid UUID but nil usecase -> panics
	assert.Panics(t, func() {
		doRequest(router, http.MethodPost, "/reports/custom/550e8400-e29b-41d4-a716-446655440000/execute?page=2&pageSize=10", nil)
	})
}

func TestCustomReportHandler_List_PanicsOnNilUsecase(t *testing.T) {
	h := newCustomReportHandlerUnit()
	router := newRouter()
	router.GET("/reports/custom", withAuthUser(1, "methodist"), h.List)

	assert.Panics(t, func() {
		doRequest(router, http.MethodGet, "/reports/custom?page=1&pageSize=10", nil)
	})
}

func TestCustomReportHandler_Create_InvalidDataSource(t *testing.T) {
	h := newCustomReportHandlerUnit()
	router := newRouter()
	router.POST("/reports/custom", withAuthUser(1, "methodist"), h.Create)

	// Validation catches invalid data source
	w := doRequest(router, http.MethodPost, "/reports/custom", map[string]interface{}{
		"name":       "Report",
		"dataSource": "invalid_source",
		"fields":     []map[string]interface{}{{"field_key": "id", "display_name": "ID", "order": 1}},
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCustomReportHandler_Create_EmptyFields(t *testing.T) {
	h := newCustomReportHandlerUnit()
	router := newRouter()
	router.POST("/reports/custom", withAuthUser(1, "methodist"), h.Create)

	// Validation catches empty fields
	w := doRequest(router, http.MethodPost, "/reports/custom", map[string]interface{}{
		"name":       "Report",
		"dataSource": "documents",
		"fields":     []map[string]interface{}{},
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCustomReportHandler_Create_ReachesUsecase(t *testing.T) {
	h := newCustomReportHandlerUnit()
	router := newRouter()
	router.POST("/reports/custom", withAuthUser(1, "methodist"), h.Create)

	// Valid input that passes validation - usecase returns ErrInvalidDataSource because
	// the DTO data_source "documents" maps to an entity DataSourceType which is then validated
	assert.Panics(t, func() {
		doRequest(router, http.MethodPost, "/reports/custom", map[string]interface{}{
			"name":       "Report",
			"dataSource": "documents",
			"fields":     []map[string]interface{}{{"fieldKey": "id", "displayName": "ID", "order": 1}},
		})
	})
}

func TestCustomReportHandler_GetByID_PanicsOnNilUsecase(t *testing.T) {
	h := newCustomReportHandlerUnit()
	router := newRouter()
	router.GET("/reports/custom/:id", withAuthUser(1, "methodist"), h.GetByID)

	assert.Panics(t, func() {
		doRequest(router, http.MethodGet, "/reports/custom/550e8400-e29b-41d4-a716-446655440000", nil)
	})
}

func TestCustomReportHandler_Update_PanicsOnNilUsecase(t *testing.T) {
	h := newCustomReportHandlerUnit()
	router := newRouter()
	router.PUT("/reports/custom/:id", withAuthUser(1, "methodist"), h.Update)

	assert.Panics(t, func() {
		doRequest(router, http.MethodPut, "/reports/custom/550e8400-e29b-41d4-a716-446655440000", map[string]interface{}{
			"name": "Updated",
		})
	})
}

func TestCustomReportHandler_Delete_PanicsOnNilUsecase(t *testing.T) {
	h := newCustomReportHandlerUnit()
	router := newRouter()
	router.DELETE("/reports/custom/:id", withAuthUser(1, "methodist"), h.Delete)

	assert.Panics(t, func() {
		doRequest(router, http.MethodDelete, "/reports/custom/550e8400-e29b-41d4-a716-446655440000", nil)
	})
}

func TestCustomReportHandler_Execute_PanicsOnNilUsecase(t *testing.T) {
	h := newCustomReportHandlerUnit()
	router := newRouter()
	router.POST("/reports/custom/:id/execute", withAuthUser(1, "methodist"), h.Execute)

	assert.Panics(t, func() {
		doRequest(router, http.MethodPost, "/reports/custom/550e8400-e29b-41d4-a716-446655440000/execute", map[string]interface{}{
			"page":      1,
			"page_size": 50,
		})
	})
}

func TestCustomReportHandler_Export_PanicsOnNilUsecase(t *testing.T) {
	h := newCustomReportHandlerUnit()
	router := newRouter()
	router.POST("/reports/custom/:id/export", withAuthUser(1, "methodist"), h.Export)

	assert.Panics(t, func() {
		doRequest(router, http.MethodPost, "/reports/custom/550e8400-e29b-41d4-a716-446655440000/export", map[string]interface{}{
			"format": "csv",
		})
	})
}

func TestCustomReportHandler_GetMyReports_PanicsWithPaginationParams(t *testing.T) {
	h := newCustomReportHandlerUnit()
	router := newRouter()
	router.GET("/reports/custom/my", withAuthUser(1, "methodist"), h.GetMyReports)

	assert.Panics(t, func() {
		doRequest(router, http.MethodGet, "/reports/custom/my?page=2&pageSize=5", nil)
	})
}

func TestNewCustomReportHandler(t *testing.T) {
	h := newCustomReportHandlerUnit()
	assert.NotNil(t, h)
}
