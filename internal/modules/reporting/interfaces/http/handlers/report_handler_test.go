package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/interfaces/http/handlers"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newRouter() *gin.Engine {
	return gin.New()
}

func withAuthUser(userID int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

func doRequest(router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
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

func newReportHandler() *handlers.ReportHandler {
	uc := usecases.NewReportUseCase(nil, nil, nil, nil, nil)
	return handlers.NewReportHandler(uc)
}

func TestReportHandler_Create(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports", h.Create)

		w := doRequest(router, http.MethodPost, "/reports", map[string]interface{}{
			"title":          "Report",
			"report_type_id": 1,
		})
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports", withAuthUser(1, "methodist"), h.Create)

		w := doRequest(router, http.MethodPost, "/reports", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports", withAuthUser(1, "methodist"), h.Create)

		w := doRequest(router, http.MethodPost, "/reports", map[string]interface{}{
			"title": "",
		})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestReportHandler_GetByID(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.GET("/reports/:id", h.GetByID)

		w := doRequest(router, http.MethodGet, "/reports/1", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.GET("/reports/:id", withAuthUser(1, "methodist"), h.GetByID)

		w := doRequest(router, http.MethodGet, "/reports/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing id", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.GET("/reports/:id", withAuthUser(1, "methodist"), h.GetByID)

		w := doRequest(router, http.MethodGet, "/reports/", nil)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestReportHandler_Update(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.PUT("/reports/:id", h.Update)

		w := doRequest(router, http.MethodPut, "/reports/1", map[string]interface{}{"title": "X"})
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.PUT("/reports/:id", withAuthUser(1, "methodist"), h.Update)

		w := doRequest(router, http.MethodPut, "/reports/abc", map[string]interface{}{"title": "X"})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.PUT("/reports/:id", withAuthUser(1, "methodist"), h.Update)

		w := doRequest(router, http.MethodPut, "/reports/1", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestReportHandler_Delete(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.DELETE("/reports/:id", h.Delete)

		w := doRequest(router, http.MethodDelete, "/reports/1", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.DELETE("/reports/:id", withAuthUser(1, "methodist"), h.Delete)

		w := doRequest(router, http.MethodDelete, "/reports/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestReportHandler_List(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.GET("/reports", h.List)

		w := doRequest(router, http.MethodGet, "/reports", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestReportHandler_Generate(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports/:id/generate", h.Generate)

		w := doRequest(router, http.MethodPost, "/reports/1/generate", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports/:id/generate", withAuthUser(1, "methodist"), h.Generate)

		w := doRequest(router, http.MethodPost, "/reports/abc/generate", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestReportHandler_SubmitForReview(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports/:id/submit", h.SubmitForReview)

		w := doRequest(router, http.MethodPost, "/reports/1/submit", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports/:id/submit", withAuthUser(1, "methodist"), h.SubmitForReview)

		w := doRequest(router, http.MethodPost, "/reports/abc/submit", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestReportHandler_Review(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports/:id/review", h.Review)

		w := doRequest(router, http.MethodPost, "/reports/1/review", map[string]interface{}{
			"approved": true,
			"comment":  "OK",
		})
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports/:id/review", withAuthUser(1, "methodist"), h.Review)

		w := doRequest(router, http.MethodPost, "/reports/abc/review", map[string]interface{}{
			"approved": true,
			"comment":  "OK",
		})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports/:id/review", withAuthUser(1, "methodist"), h.Review)

		w := doRequest(router, http.MethodPost, "/reports/1/review", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestReportHandler_Publish(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports/:id/publish", h.Publish)

		w := doRequest(router, http.MethodPost, "/reports/1/publish", map[string]interface{}{})
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports/:id/publish", withAuthUser(1, "methodist"), h.Publish)

		w := doRequest(router, http.MethodPost, "/reports/abc/publish", map[string]interface{}{})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports/:id/publish", withAuthUser(1, "methodist"), h.Publish)

		w := doRequest(router, http.MethodPost, "/reports/1/publish", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestReportHandler_AddAccess(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports/:id/access", h.AddAccess)

		w := doRequest(router, http.MethodPost, "/reports/1/access", map[string]interface{}{
			"user_id":    2,
			"permission": "read",
		})
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports/:id/access", withAuthUser(1, "methodist"), h.AddAccess)

		w := doRequest(router, http.MethodPost, "/reports/abc/access", map[string]interface{}{
			"user_id":    2,
			"permission": "read",
		})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports/:id/access", withAuthUser(1, "methodist"), h.AddAccess)

		w := doRequest(router, http.MethodPost, "/reports/1/access", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestReportHandler_RemoveAccess(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.DELETE("/reports/:id/access/:access_id", h.RemoveAccess)

		w := doRequest(router, http.MethodDelete, "/reports/1/access/1", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid report id", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.DELETE("/reports/:id/access/:access_id", withAuthUser(1, "methodist"), h.RemoveAccess)

		w := doRequest(router, http.MethodDelete, "/reports/abc/access/1", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid access id", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.DELETE("/reports/:id/access/:access_id", withAuthUser(1, "methodist"), h.RemoveAccess)

		w := doRequest(router, http.MethodDelete, "/reports/1/access/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestReportHandler_GetAccess(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.GET("/reports/:id/access", h.GetAccess)

		w := doRequest(router, http.MethodGet, "/reports/1/access", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.GET("/reports/:id/access", withAuthUser(1, "methodist"), h.GetAccess)

		w := doRequest(router, http.MethodGet, "/reports/abc/access", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestReportHandler_AddComment(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports/:id/comments", h.AddComment)

		w := doRequest(router, http.MethodPost, "/reports/1/comments", map[string]interface{}{
			"content": "Nice report",
		})
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports/:id/comments", withAuthUser(1, "methodist"), h.AddComment)

		w := doRequest(router, http.MethodPost, "/reports/abc/comments", map[string]interface{}{
			"content": "Nice report",
		})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.POST("/reports/:id/comments", withAuthUser(1, "methodist"), h.AddComment)

		w := doRequest(router, http.MethodPost, "/reports/1/comments", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestReportHandler_GetComments(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.GET("/reports/:id/comments", h.GetComments)

		w := doRequest(router, http.MethodGet, "/reports/1/comments", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.GET("/reports/:id/comments", withAuthUser(1, "methodist"), h.GetComments)

		w := doRequest(router, http.MethodGet, "/reports/abc/comments", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestReportHandler_GetHistory(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.GET("/reports/:id/history", h.GetHistory)

		w := doRequest(router, http.MethodGet, "/reports/1/history", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.GET("/reports/:id/history", withAuthUser(1, "methodist"), h.GetHistory)

		w := doRequest(router, http.MethodGet, "/reports/abc/history", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestReportHandler_GetReportTypeByID(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.GET("/report-types/:id", h.GetReportTypeByID)

		w := doRequest(router, http.MethodGet, "/report-types/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("panics on nil usecase with valid id", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.GET("/report-types/:id", h.GetReportTypeByID)

		assert.Panics(t, func() {
			doRequest(router, http.MethodGet, "/report-types/1", nil)
		})
	})
}

func TestReportHandler_GetReportTypes(t *testing.T) {
	t.Run("panics on nil usecase", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.GET("/report-types", h.GetReportTypes)

		assert.Panics(t, func() {
			doRequest(router, http.MethodGet, "/report-types", nil)
		})
	})
}

func TestReportHandler_Create_WithDescription(t *testing.T) {
	h := newReportHandler()
	router := newRouter()
	router.POST("/reports", withAuthUser(1, "methodist"), h.Create)

	// Valid input with description - panics because usecase has nil repos
	assert.Panics(t, func() {
		doRequest(router, http.MethodPost, "/reports", map[string]interface{}{
			"title":          "Test Report",
			"report_type_id": 1,
			"description":    "A description",
		})
	})
}

func TestReportHandler_GetHistory_WithQueryParams(t *testing.T) {
	t.Run("with limit and offset", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.GET("/reports/:id/history", withAuthUser(1, "methodist"), h.GetHistory)

		assert.Panics(t, func() {
			doRequest(router, http.MethodGet, "/reports/1/history?limit=10&offset=5", nil)
		})
	})

	t.Run("with invalid limit and offset", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.GET("/reports/:id/history", withAuthUser(1, "methodist"), h.GetHistory)

		assert.Panics(t, func() {
			doRequest(router, http.MethodGet, "/reports/1/history?limit=abc&offset=xyz", nil)
		})
	})

	t.Run("with negative limit and offset", func(t *testing.T) {
		h := newReportHandler()
		router := newRouter()
		router.GET("/reports/:id/history", withAuthUser(1, "methodist"), h.GetHistory)

		assert.Panics(t, func() {
			doRequest(router, http.MethodGet, "/reports/1/history?limit=-1&offset=-1", nil)
		})
	})
}

func TestReportHandler_Review_ValidationError(t *testing.T) {
	h := newReportHandler()
	router := newRouter()
	router.POST("/reports/:id/review", withAuthUser(1, "methodist"), h.Review)

	// Invalid action value - should fail validation
	w := doRequest(router, http.MethodPost, "/reports/1/review", map[string]interface{}{
		"action":  "invalid_action",
		"comment": "OK",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReportHandler_AddComment_ValidationError(t *testing.T) {
	h := newReportHandler()
	router := newRouter()
	router.POST("/reports/:id/comments", withAuthUser(1, "methodist"), h.AddComment)

	w := doRequest(router, http.MethodPost, "/reports/1/comments", map[string]interface{}{
		"content": "",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReportHandler_AddAccess_ValidationError(t *testing.T) {
	h := newReportHandler()
	router := newRouter()
	router.POST("/reports/:id/access", withAuthUser(1, "methodist"), h.AddAccess)

	w := doRequest(router, http.MethodPost, "/reports/1/access", map[string]interface{}{
		"permission": "invalid_permission",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReportHandler_List_PanicsOnNilUsecase(t *testing.T) {
	h := newReportHandler()
	router := newRouter()
	router.GET("/reports", withAuthUser(1, "methodist"), h.List)

	assert.Panics(t, func() {
		doRequest(router, http.MethodGet, "/reports", nil)
	})
}

func TestReportHandler_Generate_PanicsOnNilUsecase(t *testing.T) {
	h := newReportHandler()
	router := newRouter()
	router.POST("/reports/:id/generate", withAuthUser(1, "methodist"), h.Generate)

	assert.Panics(t, func() {
		doRequest(router, http.MethodPost, "/reports/1/generate", nil)
	})
}

func TestReportHandler_SubmitForReview_PanicsOnNilUsecase(t *testing.T) {
	h := newReportHandler()
	router := newRouter()
	router.POST("/reports/:id/submit", withAuthUser(1, "methodist"), h.SubmitForReview)

	assert.Panics(t, func() {
		doRequest(router, http.MethodPost, "/reports/1/submit", nil)
	})
}

func TestReportHandler_GetByID_PanicsOnNilUsecase(t *testing.T) {
	h := newReportHandler()
	router := newRouter()
	router.GET("/reports/:id", withAuthUser(1, "methodist"), h.GetByID)

	assert.Panics(t, func() {
		doRequest(router, http.MethodGet, "/reports/1", nil)
	})
}

func TestReportHandler_Update_PanicsOnNilUsecase(t *testing.T) {
	h := newReportHandler()
	router := newRouter()
	router.PUT("/reports/:id", withAuthUser(1, "methodist"), h.Update)

	assert.Panics(t, func() {
		doRequest(router, http.MethodPut, "/reports/1", map[string]interface{}{"title": "X"})
	})
}

func TestReportHandler_Delete_PanicsOnNilUsecase(t *testing.T) {
	h := newReportHandler()
	router := newRouter()
	router.DELETE("/reports/:id", withAuthUser(1, "methodist"), h.Delete)

	assert.Panics(t, func() {
		doRequest(router, http.MethodDelete, "/reports/1", nil)
	})
}

func TestReportHandler_RemoveAccess_PanicsOnNilUsecase(t *testing.T) {
	h := newReportHandler()
	router := newRouter()
	router.DELETE("/reports/:id/access/:access_id", withAuthUser(1, "methodist"), h.RemoveAccess)

	assert.Panics(t, func() {
		doRequest(router, http.MethodDelete, "/reports/1/access/1", nil)
	})
}

func TestReportHandler_GetAccess_PanicsOnNilUsecase(t *testing.T) {
	h := newReportHandler()
	router := newRouter()
	router.GET("/reports/:id/access", withAuthUser(1, "methodist"), h.GetAccess)

	assert.Panics(t, func() {
		doRequest(router, http.MethodGet, "/reports/1/access", nil)
	})
}

func TestReportHandler_GetComments_PanicsOnNilUsecase(t *testing.T) {
	h := newReportHandler()
	router := newRouter()
	router.GET("/reports/:id/comments", withAuthUser(1, "methodist"), h.GetComments)

	assert.Panics(t, func() {
		doRequest(router, http.MethodGet, "/reports/1/comments", nil)
	})
}

func TestReportHandler_AddComment_PanicsOnNilUsecase(t *testing.T) {
	h := newReportHandler()
	router := newRouter()
	router.POST("/reports/:id/comments", withAuthUser(1, "methodist"), h.AddComment)

	assert.Panics(t, func() {
		doRequest(router, http.MethodPost, "/reports/1/comments", map[string]interface{}{
			"content": "Nice report",
		})
	})
}

func TestReportHandler_Review_PanicsOnNilUsecase(t *testing.T) {
	h := newReportHandler()
	router := newRouter()
	router.POST("/reports/:id/review", withAuthUser(1, "methodist"), h.Review)

	assert.Panics(t, func() {
		doRequest(router, http.MethodPost, "/reports/1/review", map[string]interface{}{
			"action":  "approve",
			"comment": "Good work",
		})
	})
}

func TestReportHandler_Publish_PanicsOnNilUsecase(t *testing.T) {
	h := newReportHandler()
	router := newRouter()
	router.POST("/reports/:id/publish", withAuthUser(1, "methodist"), h.Publish)

	assert.Panics(t, func() {
		doRequest(router, http.MethodPost, "/reports/1/publish", map[string]interface{}{
			"is_public": true,
		})
	})
}

func TestReportHandler_AddAccess_PanicsOnNilUsecase(t *testing.T) {
	h := newReportHandler()
	router := newRouter()
	router.POST("/reports/:id/access", withAuthUser(1, "methodist"), h.AddAccess)

	userID := int64(2)
	assert.Panics(t, func() {
		doRequest(router, http.MethodPost, "/reports/1/access", map[string]interface{}{
			"user_id":    userID,
			"permission": "read",
		})
	})
}

func TestNewReportHandler(t *testing.T) {
	h := newReportHandler()
	assert.NotNil(t, h)
}

// Suppress unused import warnings
var _ = usecases.ErrReportNotFound
