package http_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	handlers "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/interfaces/http/handlers"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

func newTemplateHandler(tmplRepo *MockTemplateRepository, docRepo *MockDocumentRepository) *handlers.TemplateHandler {
	uc := usecases.NewTemplateUseCase(tmplRepo, docRepo, nil)
	v := validation.NewValidator()
	return handlers.NewTemplateHandler(uc, v)
}

func TestTemplateHandler_GetTemplates(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		content := "<p>Template</p>"
		types := []entities.DocumentType{
			{ID: 1, Name: "Type1", TemplateContent: &content, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		tmplRepo.On("GetAll", mock.Anything).Return(types, nil)

		router := setupRouter()
		router.GET("/templates", h.GetTemplates)

		w := performRequest(router, http.MethodGet, "/templates", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		tmplRepo.On("GetAll", mock.Anything).Return(nil, fmt.Errorf("db error"))

		router := setupRouter()
		router.GET("/templates", h.GetTemplates)

		w := performRequest(router, http.MethodGet, "/templates", nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	// v0.126.0 reviewer round-1 fix: pin that GetTemplates reads the
	// JWT role from the same context key that auth middleware writes.
	// The middleware uses 'role'; the handler used to read 'user_role'
	// — that mismatch silently dropped methodist-only filtering in
	// production for every role.
	t.Run("methodist-only filter — system_admin sees both", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		open := "<p>Open</p>"
		methodist := "<p>Methodist</p>"
		types := []entities.DocumentType{
			{ID: 1, Name: "Open", Code: "open", TemplateContent: &open, MethodistOnly: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: 2, Name: "Methodist", Code: "metonly", TemplateContent: &methodist, MethodistOnly: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		tmplRepo.On("GetAll", mock.Anything).Return(types, nil)

		router := setupRouter()
		router.GET("/templates", withAuth(1, "system_admin"), h.GetTemplates)

		w := performRequest(router, http.MethodGet, "/templates", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		body := parseResponseBody(w)
		templates, _ := body["templates"].([]interface{})
		assert.Len(t, templates, 2, "system_admin must see both open and methodist-only templates")
	})

	t.Run("methodist-only filter — teacher sees open only", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		open := "<p>Open</p>"
		methodist := "<p>Methodist</p>"
		types := []entities.DocumentType{
			{ID: 1, Name: "Open", Code: "open", TemplateContent: &open, MethodistOnly: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: 2, Name: "Methodist", Code: "metonly", TemplateContent: &methodist, MethodistOnly: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		tmplRepo.On("GetAll", mock.Anything).Return(types, nil)

		router := setupRouter()
		router.GET("/templates", withAuth(1, "teacher"), h.GetTemplates)

		w := performRequest(router, http.MethodGet, "/templates", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		body := parseResponseBody(w)
		templates, _ := body["templates"].([]interface{})
		require.Len(t, templates, 1, "teacher must see only open templates")
		first, _ := templates[0].(map[string]interface{})
		assert.Equal(t, "Open", first["name"], "teacher must not see methodist-only template")
	})

	t.Run("methodist-only filter — no auth context falls back to failure-closed", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		open := "<p>Open</p>"
		methodist := "<p>Methodist</p>"
		types := []entities.DocumentType{
			{ID: 1, Name: "Open", Code: "open", TemplateContent: &open, MethodistOnly: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: 2, Name: "Methodist", Code: "metonly", TemplateContent: &methodist, MethodistOnly: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		tmplRepo.On("GetAll", mock.Anything).Return(types, nil)

		router := setupRouter()
		// No middleware — context has no 'role' key; handler must not panic
		// and must default to the failure-closed path (open-only result).
		router.GET("/templates", h.GetTemplates)

		w := performRequest(router, http.MethodGet, "/templates", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		body := parseResponseBody(w)
		templates, _ := body["templates"].([]interface{})
		assert.Len(t, templates, 1, "missing role context must fall through to open-only (failure-closed)")
	})
}

func TestTemplateHandler_GetTemplate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		content := "<p>Template</p>"
		dt := &entities.DocumentType{ID: 1, Name: "Type1", TemplateContent: &content, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		tmplRepo.On("GetByID", mock.Anything, int64(1)).Return(dt, nil)

		router := setupRouter()
		router.GET("/templates/:id", h.GetTemplate)

		w := performRequest(router, http.MethodGet, "/templates/1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		router := setupRouter()
		router.GET("/templates/:id", h.GetTemplate)

		w := performRequest(router, http.MethodGet, "/templates/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		tmplRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, fmt.Errorf("not found"))

		router := setupRouter()
		router.GET("/templates/:id", h.GetTemplate)

		w := performRequest(router, http.MethodGet, "/templates/999", nil)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestTemplateHandler_PreviewTemplate(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		router := setupRouter()
		router.POST("/templates/:id/preview", h.PreviewTemplate)

		w := performRequest(router, http.MethodPost, "/templates/abc/preview", map[string]interface{}{
			"variables": map[string]string{"name": "Test"},
		})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		router := setupRouter()
		router.POST("/templates/:id/preview", h.PreviewTemplate)

		w := performRequest(router, http.MethodPost, "/templates/1/preview", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		content := "Hello {{.name}}"
		dt := &entities.DocumentType{
			ID:              1,
			Name:            "Type1",
			TemplateContent: &content,
			TemplateVariables: []entities.TemplateVariable{
				{Name: "name", Label: "Name", Type: "text"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		tmplRepo.On("GetByID", mock.Anything, int64(1)).Return(dt, nil)

		router := setupRouter()
		router.POST("/templates/:id/preview", h.PreviewTemplate)

		w := performRequest(router, http.MethodPost, "/templates/1/preview", map[string]interface{}{
			"variables": map[string]string{"name": "World"},
		})
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestTemplateHandler_CreateDocumentFromTemplate(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		router := setupRouter()
		router.POST("/templates/:id/create", withAuth(1, "methodist"), h.CreateDocumentFromTemplate)

		w := performRequest(router, http.MethodPost, "/templates/abc/create", map[string]interface{}{
			"title":     "Doc",
			"variables": map[string]string{},
		})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("no auth", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		router := setupRouter()
		router.POST("/templates/:id/create", h.CreateDocumentFromTemplate)

		w := performRequest(router, http.MethodPost, "/templates/1/create", map[string]interface{}{
			"title":     "Doc",
			"variables": map[string]string{},
		})
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		router := setupRouter()
		router.POST("/templates/:id/create", withAuth(1, "methodist"), h.CreateDocumentFromTemplate)

		w := performRequest(router, http.MethodPost, "/templates/1/create", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTemplateHandler_UpdateTemplate(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		router := setupRouter()
		router.PUT("/templates/:id", withAuth(1, "system_admin"), h.UpdateTemplate)

		w := performRequest(router, http.MethodPut, "/templates/abc", map[string]interface{}{
			"content": "<p>new</p>",
		})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("no role", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		router := setupRouter()
		router.PUT("/templates/:id", h.UpdateTemplate)

		w := performRequest(router, http.MethodPut, "/templates/1", map[string]interface{}{
			"content": "<p>new</p>",
		})
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("wrong role", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		router := setupRouter()
		router.PUT("/templates/:id", withAuth(1, "student"), h.UpdateTemplate)

		w := performRequest(router, http.MethodPut, "/templates/1", map[string]interface{}{
			"content": "<p>new</p>",
		})
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		router := setupRouter()
		router.PUT("/templates/:id", withAuth(1, "system_admin"), h.UpdateTemplate)

		w := performRequest(router, http.MethodPut, "/templates/1", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		tmplRepo := new(MockTemplateRepository)
		docRepo := new(MockDocumentRepository)
		h := newTemplateHandler(tmplRepo, docRepo)

		tmplRepo.On("GetByID", mock.Anything, int64(1)).Return(&entities.DocumentType{ID: 1, Name: "T"}, nil)
		tmplRepo.On("UpdateTemplate", mock.Anything, int64(1), mock.Anything, mock.Anything, mock.Anything).Return(nil)

		router := setupRouter()
		router.PUT("/templates/:id", withAuth(1, "system_admin"), h.UpdateTemplate)

		w := performRequest(router, http.MethodPut, "/templates/1", map[string]interface{}{
			"content": "<p>updated</p>",
		})
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
