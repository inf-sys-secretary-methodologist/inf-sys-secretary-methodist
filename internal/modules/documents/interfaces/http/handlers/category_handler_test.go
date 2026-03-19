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
)

func newCategoryHandler(catRepo *MockDocumentCategoryRepository) *handlers.CategoryHandler {
	uc := usecases.NewCategoryUseCase(catRepo, nil)
	return handlers.NewCategoryHandler(uc)
}

func TestCategoryHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		cat := &entities.DocumentCategory{ID: 1, Name: "Test Cat", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		catRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.DocumentCategory")).
			Run(func(args mock.Arguments) {
				c := args.Get(1).(*entities.DocumentCategory)
				c.ID = 1
				c.CreatedAt = cat.CreatedAt
				c.UpdatedAt = cat.UpdatedAt
			}).Return(nil)
		catRepo.On("HasChildren", mock.Anything, int64(1)).Return(false, nil)
		catRepo.On("GetDocumentCount", mock.Anything, int64(1), false).Return(int64(0), nil)

		router := setupRouter()
		router.POST("/categories", h.Create)

		w := performRequest(router, http.MethodPost, "/categories", map[string]interface{}{
			"name": "Test Cat",
		})

		assert.Equal(t, http.StatusCreated, w.Code)
		resp := parseResponseBody(w)
		assert.True(t, resp["success"].(bool))
	})

	t.Run("invalid json", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		router := setupRouter()
		router.POST("/categories", h.Create)

		w := performRequest(router, http.MethodPost, "/categories", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("validation error - empty name", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		router := setupRouter()
		router.POST("/categories", h.Create)

		w := performRequest(router, http.MethodPost, "/categories", map[string]interface{}{
			"name": "",
		})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		parentID := int64(999)
		catRepo.On("GetByID", mock.Anything, parentID).Return(nil, fmt.Errorf("not found"))

		router := setupRouter()
		router.POST("/categories", h.Create)

		w := performRequest(router, http.MethodPost, "/categories", map[string]interface{}{
			"name":      "Child Cat",
			"parent_id": 999,
		})
		// The usecase returns a generic error, MapDomainError maps it
		require.NotEqual(t, http.StatusCreated, w.Code)
	})

	t.Run("sanitize description", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		catRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.DocumentCategory")).Return(nil)
		catRepo.On("HasChildren", mock.Anything, mock.Anything).Return(false, nil)
		catRepo.On("GetDocumentCount", mock.Anything, mock.Anything, false).Return(int64(0), nil)

		router := setupRouter()
		router.POST("/categories", h.Create)

		desc := "<script>alert('xss')</script>Clean"
		w := performRequest(router, http.MethodPost, "/categories", map[string]interface{}{
			"name":        "Safe Cat",
			"description": desc,
		})

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestCategoryHandler_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		cat := &entities.DocumentCategory{ID: 1, Name: "Cat 1", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		catRepo.On("GetByID", mock.Anything, int64(1)).Return(cat, nil)
		catRepo.On("HasChildren", mock.Anything, int64(1)).Return(false, nil)
		catRepo.On("GetDocumentCount", mock.Anything, int64(1), false).Return(int64(5), nil)

		router := setupRouter()
		router.GET("/categories/:id", h.GetByID)

		w := performRequest(router, http.MethodGet, "/categories/1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponseBody(w)
		assert.True(t, resp["success"].(bool))
	})

	t.Run("invalid id", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		router := setupRouter()
		router.GET("/categories/:id", h.GetByID)

		w := performRequest(router, http.MethodGet, "/categories/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		catRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, fmt.Errorf("not found"))

		router := setupRouter()
		router.GET("/categories/:id", h.GetByID)

		w := performRequest(router, http.MethodGet, "/categories/999", nil)
		require.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestCategoryHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		cat := &entities.DocumentCategory{ID: 1, Name: "Old Name", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		catRepo.On("GetByID", mock.Anything, int64(1)).Return(cat, nil)
		catRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.DocumentCategory")).Return(nil)
		catRepo.On("HasChildren", mock.Anything, int64(1)).Return(false, nil)
		catRepo.On("GetDocumentCount", mock.Anything, int64(1), false).Return(int64(0), nil)

		router := setupRouter()
		router.PUT("/categories/:id", h.Update)

		name := "New Name"
		w := performRequest(router, http.MethodPut, "/categories/1", map[string]interface{}{
			"name": name,
		})
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		router := setupRouter()
		router.PUT("/categories/:id", h.Update)

		w := performRequest(router, http.MethodPut, "/categories/abc", map[string]interface{}{"name": "X"})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		router := setupRouter()
		router.PUT("/categories/:id", h.Update)

		w := performRequest(router, http.MethodPut, "/categories/1", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("sanitize name and description", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		cat := &entities.DocumentCategory{ID: 1, Name: "Old", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		catRepo.On("GetByID", mock.Anything, int64(1)).Return(cat, nil)
		catRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.DocumentCategory")).Return(nil)
		catRepo.On("HasChildren", mock.Anything, int64(1)).Return(false, nil)
		catRepo.On("GetDocumentCount", mock.Anything, int64(1), false).Return(int64(0), nil)

		router := setupRouter()
		router.PUT("/categories/:id", h.Update)

		w := performRequest(router, http.MethodPut, "/categories/1", map[string]interface{}{
			"name":        "<b>Bold</b>",
			"description": "<script>x</script>desc",
		})
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestCategoryHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		cat := &entities.DocumentCategory{ID: 1, Name: "Cat"}
		catRepo.On("GetByID", mock.Anything, int64(1)).Return(cat, nil)
		catRepo.On("HasChildren", mock.Anything, int64(1)).Return(false, nil)
		catRepo.On("GetDocumentCount", mock.Anything, int64(1), false).Return(int64(0), nil)
		catRepo.On("Delete", mock.Anything, int64(1)).Return(nil)

		router := setupRouter()
		router.DELETE("/categories/:id", h.Delete)

		w := performRequest(router, http.MethodDelete, "/categories/1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		router := setupRouter()
		router.DELETE("/categories/:id", h.Delete)

		w := performRequest(router, http.MethodDelete, "/categories/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		cat := &entities.DocumentCategory{ID: 1, Name: "Cat"}
		catRepo.On("GetByID", mock.Anything, int64(1)).Return(cat, nil)
		catRepo.On("HasChildren", mock.Anything, int64(1)).Return(true, nil)

		router := setupRouter()
		router.DELETE("/categories/:id", h.Delete)

		w := performRequest(router, http.MethodDelete, "/categories/1", nil)
		require.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestCategoryHandler_GetAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		cats := []*entities.DocumentCategory{
			{ID: 1, Name: "Cat 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: 2, Name: "Cat 2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		catRepo.On("GetAll", mock.Anything).Return(cats, nil)
		catRepo.On("HasChildren", mock.Anything, mock.Anything).Return(false, nil)
		catRepo.On("GetDocumentCount", mock.Anything, mock.Anything, false).Return(int64(0), nil)

		router := setupRouter()
		router.GET("/categories", h.GetAll)

		w := performRequest(router, http.MethodGet, "/categories", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		catRepo.On("GetAll", mock.Anything).Return(nil, fmt.Errorf("db error"))

		router := setupRouter()
		router.GET("/categories", h.GetAll)

		w := performRequest(router, http.MethodGet, "/categories", nil)
		require.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestCategoryHandler_GetTree(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		tree := []*entities.CategoryTreeNode{
			{ID: 1, Name: "Root", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		catRepo.On("GetTree", mock.Anything).Return(tree, nil)

		router := setupRouter()
		router.GET("/categories/tree", h.GetTree)

		w := performRequest(router, http.MethodGet, "/categories/tree", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		catRepo.On("GetTree", mock.Anything).Return(nil, fmt.Errorf("error"))

		router := setupRouter()
		router.GET("/categories/tree", h.GetTree)

		w := performRequest(router, http.MethodGet, "/categories/tree", nil)
		require.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestCategoryHandler_GetChildren(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		parent := &entities.DocumentCategory{ID: 1, Name: "Parent"}
		catRepo.On("GetByID", mock.Anything, int64(1)).Return(parent, nil)
		catRepo.On("GetChildren", mock.Anything, int64(1)).Return([]*entities.DocumentCategory{
			{ID: 2, Name: "Child", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}, nil)
		catRepo.On("HasChildren", mock.Anything, mock.Anything).Return(false, nil)
		catRepo.On("GetDocumentCount", mock.Anything, mock.Anything, false).Return(int64(0), nil)

		router := setupRouter()
		router.GET("/categories/:id/children", h.GetChildren)

		w := performRequest(router, http.MethodGet, "/categories/1/children", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		router := setupRouter()
		router.GET("/categories/:id/children", h.GetChildren)

		w := performRequest(router, http.MethodGet, "/categories/abc/children", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCategoryHandler_GetRootCategories(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		catRepo.On("GetByParentID", mock.Anything, (*int64)(nil)).Return([]*entities.DocumentCategory{
			{ID: 1, Name: "Root 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}, nil)
		catRepo.On("HasChildren", mock.Anything, mock.Anything).Return(false, nil)
		catRepo.On("GetDocumentCount", mock.Anything, mock.Anything, false).Return(int64(0), nil)

		router := setupRouter()
		router.GET("/categories/root", h.GetRootCategories)

		w := performRequest(router, http.MethodGet, "/categories/root", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		catRepo.On("GetByParentID", mock.Anything, (*int64)(nil)).Return(nil, fmt.Errorf("error"))

		router := setupRouter()
		router.GET("/categories/root", h.GetRootCategories)

		w := performRequest(router, http.MethodGet, "/categories/root", nil)
		require.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestCategoryHandler_GetWithBreadcrumb(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		cat := &entities.DocumentCategory{ID: 2, Name: "Child", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		catRepo.On("GetByID", mock.Anything, int64(2)).Return(cat, nil)
		catRepo.On("GetAncestors", mock.Anything, int64(2)).Return([]*entities.DocumentCategory{
			{ID: 1, Name: "Parent"},
		}, nil)
		catRepo.On("HasChildren", mock.Anything, int64(2)).Return(false, nil)
		catRepo.On("GetDocumentCount", mock.Anything, int64(2), false).Return(int64(0), nil)

		router := setupRouter()
		router.GET("/categories/:id/breadcrumb", h.GetWithBreadcrumb)

		w := performRequest(router, http.MethodGet, "/categories/2/breadcrumb", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		router := setupRouter()
		router.GET("/categories/:id/breadcrumb", h.GetWithBreadcrumb)

		w := performRequest(router, http.MethodGet, "/categories/abc/breadcrumb", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		catRepo := new(MockDocumentCategoryRepository)
		h := newCategoryHandler(catRepo)

		catRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, fmt.Errorf("not found"))

		router := setupRouter()
		router.GET("/categories/:id/breadcrumb", h.GetWithBreadcrumb)

		w := performRequest(router, http.MethodGet, "/categories/999/breadcrumb", nil)
		require.NotEqual(t, http.StatusOK, w.Code)
	})
}
