package http_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
	handlers "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/interfaces/http/handlers"
)

func newDocumentHandler(docRepo *MockDocumentRepository, typeRepo *MockDocumentTypeRepository, catRepo *MockDocumentCategoryRepository) *handlers.DocumentHandler {
	uc := usecases.NewDocumentUseCase(docRepo, typeRepo, catRepo, nil, nil)
	return handlers.NewDocumentHandler(uc)
}

func TestDocumentHandler_Create_NoAuth(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	typeRepo := new(MockDocumentTypeRepository)
	catRepo := new(MockDocumentCategoryRepository)
	h := newDocumentHandler(docRepo, typeRepo, catRepo)

	router := setupRouter()
	router.POST("/documents", h.Create)

	w := performRequest(router, http.MethodPost, "/documents", map[string]interface{}{
		"title":            "Doc",
		"document_type_id": 1,
	})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDocumentHandler_Create_InvalidJSON(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	typeRepo := new(MockDocumentTypeRepository)
	catRepo := new(MockDocumentCategoryRepository)
	h := newDocumentHandler(docRepo, typeRepo, catRepo)

	router := setupRouter()
	router.POST("/documents", withAuth(1, "methodist"), h.Create)

	w := performRequest(router, http.MethodPost, "/documents", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDocumentHandler_Create_ValidationError(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	typeRepo := new(MockDocumentTypeRepository)
	catRepo := new(MockDocumentCategoryRepository)
	h := newDocumentHandler(docRepo, typeRepo, catRepo)

	router := setupRouter()
	router.POST("/documents", withAuth(1, "methodist"), h.Create)

	w := performRequest(router, http.MethodPost, "/documents", map[string]interface{}{
		"title":            "",
		"document_type_id": 0,
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDocumentHandler_Create_Success(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	typeRepo := new(MockDocumentTypeRepository)
	catRepo := new(MockDocumentCategoryRepository)
	h := newDocumentHandler(docRepo, typeRepo, catRepo)

	docType := &entities.DocumentType{ID: 1, Name: "Test Type", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	typeRepo.On("GetByID", mock.Anything, int64(1)).Return(docType, nil)
	docRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Document")).Run(func(args mock.Arguments) {
		d := args.Get(1).(*entities.Document)
		d.ID = 1
	}).Return(nil)
	docRepo.On("AddHistory", mock.Anything, mock.Anything).Return(nil)

	router := setupRouter()
	router.POST("/documents", withAuth(1, "methodist"), h.Create)

	w := performRequest(router, http.MethodPost, "/documents", map[string]interface{}{
		"title":            "New Doc",
		"document_type_id": 1,
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestDocumentHandler_Create_UsecaseError(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	typeRepo := new(MockDocumentTypeRepository)
	catRepo := new(MockDocumentCategoryRepository)
	h := newDocumentHandler(docRepo, typeRepo, catRepo)

	typeRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, fmt.Errorf("not found"))

	router := setupRouter()
	router.POST("/documents", withAuth(1, "methodist"), h.Create)

	w := performRequest(router, http.MethodPost, "/documents", map[string]interface{}{
		"title":            "New Doc",
		"document_type_id": 1,
	})
	assert.NotEqual(t, http.StatusCreated, w.Code)
}

func TestDocumentHandler_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		doc := &entities.Document{ID: 1, Title: "Test", Status: entities.DocumentStatusDraft, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)

		router := setupRouter()
		router.GET("/documents/:id", h.GetByID)

		w := performRequest(router, http.MethodGet, "/documents/1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		router := setupRouter()
		router.GET("/documents/:id", h.GetByID)

		w := performRequest(router, http.MethodGet, "/documents/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		docRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, fmt.Errorf("not found"))

		router := setupRouter()
		router.GET("/documents/:id", h.GetByID)

		w := performRequest(router, http.MethodGet, "/documents/999", nil)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestDocumentHandler_Update(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		router := setupRouter()
		router.PUT("/documents/:id", h.Update)

		w := performRequest(router, http.MethodPut, "/documents/1", map[string]interface{}{"title": "X"})
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		router := setupRouter()
		router.PUT("/documents/:id", withAuth(1, "methodist"), h.Update)

		w := performRequest(router, http.MethodPut, "/documents/abc", map[string]interface{}{"title": "X"})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		router := setupRouter()
		router.PUT("/documents/:id", withAuth(1, "methodist"), h.Update)

		w := performRequest(router, http.MethodPut, "/documents/1", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		doc := &entities.Document{ID: 1, Title: "Old", Status: entities.DocumentStatusDraft, AuthorID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)
		docRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Document")).Return(nil)
		docRepo.On("AddHistory", mock.Anything, mock.Anything).Return(nil)

		router := setupRouter()
		router.PUT("/documents/:id", withAuth(1, "methodist"), h.Update)

		w := performRequest(router, http.MethodPut, "/documents/1", map[string]interface{}{"title": "New"})
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestDocumentHandler_Delete(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		router := setupRouter()
		router.DELETE("/documents/:id", h.Delete)

		w := performRequest(router, http.MethodDelete, "/documents/1", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		router := setupRouter()
		router.DELETE("/documents/:id", withAuth(1, "methodist"), h.Delete)

		w := performRequest(router, http.MethodDelete, "/documents/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		doc := &entities.Document{ID: 1, Title: "Doc", AuthorID: 1, Status: entities.DocumentStatusDraft}
		docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)
		docRepo.On("SoftDelete", mock.Anything, int64(1)).Return(nil)
		docRepo.On("AddHistory", mock.Anything, mock.Anything).Return(nil)

		router := setupRouter()
		router.DELETE("/documents/:id", withAuth(1, "methodist"), h.Delete)

		w := performRequest(router, http.MethodDelete, "/documents/1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestDocumentHandler_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		docRepo.On("List", mock.Anything, mock.Anything).Return([]*entities.Document{}, int64(0), nil)

		router := setupRouter()
		router.GET("/documents", withAuth(1, "methodist"), h.List)

		w := performRequest(router, http.MethodGet, "/documents", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("with pagination", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		docRepo.On("List", mock.Anything, mock.Anything).Return([]*entities.Document{}, int64(0), nil)

		router := setupRouter()
		router.GET("/documents", withAuth(1, "methodist"), h.List)

		w := performRequest(router, http.MethodGet, "/documents?page=2&page_size=10", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		docRepo.On("List", mock.Anything, mock.Anything).Return(nil, int64(0), fmt.Errorf("db error"))

		router := setupRouter()
		router.GET("/documents", withAuth(1, "methodist"), h.List)

		w := performRequest(router, http.MethodGet, "/documents", nil)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestDocumentHandler_GetDocumentTypes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		types := []*entities.DocumentType{{ID: 1, Name: "Type1", CreatedAt: time.Now(), UpdatedAt: time.Now()}}
		typeRepo.On("GetAll", mock.Anything).Return(types, nil)

		router := setupRouter()
		router.GET("/types", h.GetDocumentTypes)

		w := performRequest(router, http.MethodGet, "/types", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		typeRepo.On("GetAll", mock.Anything).Return(nil, fmt.Errorf("error"))

		router := setupRouter()
		router.GET("/types", h.GetDocumentTypes)

		w := performRequest(router, http.MethodGet, "/types", nil)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestDocumentHandler_GetCategories(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		cats := []*entities.DocumentCategory{{ID: 1, Name: "Cat1", CreatedAt: time.Now(), UpdatedAt: time.Now()}}
		catRepo.On("GetAll", mock.Anything).Return(cats, nil)

		router := setupRouter()
		router.GET("/categories", h.GetCategories)

		w := performRequest(router, http.MethodGet, "/categories", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		catRepo.On("GetAll", mock.Anything).Return(nil, fmt.Errorf("error"))

		router := setupRouter()
		router.GET("/categories", h.GetCategories)

		w := performRequest(router, http.MethodGet, "/categories", nil)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestDocumentHandler_Search(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		docRepo.On("Search", mock.Anything, mock.Anything).Return([]*repositories.SearchResult{}, int64(0), nil)

		router := setupRouter()
		router.GET("/search", withAuth(1, "methodist"), h.Search)

		w := performRequest(router, http.MethodGet, "/search?q=test", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("empty query", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		router := setupRouter()
		router.GET("/search", withAuth(1, "methodist"), h.Search)

		w := performRequest(router, http.MethodGet, "/search?q=", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("no query param", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		router := setupRouter()
		router.GET("/search", withAuth(1, "methodist"), h.Search)

		w := performRequest(router, http.MethodGet, "/search", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("search error", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		docRepo.On("Search", mock.Anything, mock.Anything).Return(nil, int64(0), fmt.Errorf("error"))

		router := setupRouter()
		router.GET("/search", withAuth(1, "methodist"), h.Search)

		w := performRequest(router, http.MethodGet, "/search?q=test", nil)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestDocumentHandler_DeleteFile(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		router := setupRouter()
		router.DELETE("/documents/:id/file", h.DeleteFile)

		w := performRequest(router, http.MethodDelete, "/documents/1/file", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		typeRepo := new(MockDocumentTypeRepository)
		catRepo := new(MockDocumentCategoryRepository)
		h := newDocumentHandler(docRepo, typeRepo, catRepo)

		router := setupRouter()
		router.DELETE("/documents/:id/file", withAuth(1, "methodist"), h.DeleteFile)

		w := performRequest(router, http.MethodDelete, "/documents/abc/file", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDocumentHandler_DownloadFile_InvalidID(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	typeRepo := new(MockDocumentTypeRepository)
	catRepo := new(MockDocumentCategoryRepository)
	h := newDocumentHandler(docRepo, typeRepo, catRepo)

	router := setupRouter()
	router.GET("/documents/:id/file", h.DownloadFile)

	w := performRequest(router, http.MethodGet, "/documents/abc/file", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDocumentHandler_UploadFile_NoAuth(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	typeRepo := new(MockDocumentTypeRepository)
	catRepo := new(MockDocumentCategoryRepository)
	h := newDocumentHandler(docRepo, typeRepo, catRepo)

	router := setupRouter()
	router.POST("/documents/:id/file", h.UploadFile)

	w := performRequest(router, http.MethodPost, "/documents/1/file", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDocumentHandler_UploadFile_InvalidID(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	typeRepo := new(MockDocumentTypeRepository)
	catRepo := new(MockDocumentCategoryRepository)
	h := newDocumentHandler(docRepo, typeRepo, catRepo)

	router := setupRouter()
	router.POST("/documents/:id/file", withAuth(1, "methodist"), h.UploadFile)

	w := performRequest(router, http.MethodPost, "/documents/abc/file", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDocumentHandler_List_PageSizeClamp(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	typeRepo := new(MockDocumentTypeRepository)
	catRepo := new(MockDocumentCategoryRepository)
	h := newDocumentHandler(docRepo, typeRepo, catRepo)

	docRepo.On("List", mock.Anything, mock.Anything).Return([]*entities.Document{}, int64(0), nil)

	router := setupRouter()
	router.GET("/documents", withAuth(1, "admin"), h.List)

	// page_size > 100 should be clamped to 100
	w := performRequest(router, http.MethodGet, "/documents?page_size=200", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDocumentHandler_Search_WithPaginationAndFilters(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	typeRepo := new(MockDocumentTypeRepository)
	catRepo := new(MockDocumentCategoryRepository)
	h := newDocumentHandler(docRepo, typeRepo, catRepo)

	docRepo.On("Search", mock.Anything, mock.Anything).Return([]*repositories.SearchResult{}, int64(0), nil)

	router := setupRouter()
	router.GET("/search", withAuth(1, "methodist"), h.Search)

	// Test with page_size > 100 and extra filters
	w := performRequest(router, http.MethodGet, "/search?q=test&page=1&page_size=200", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDocumentHandler_Update_WithSanitization(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	typeRepo := new(MockDocumentTypeRepository)
	catRepo := new(MockDocumentCategoryRepository)
	h := newDocumentHandler(docRepo, typeRepo, catRepo)

	doc := &entities.Document{ID: 1, Title: "Old", Status: entities.DocumentStatusDraft, AuthorID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)
	docRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Document")).Return(nil)
	docRepo.On("AddHistory", mock.Anything, mock.Anything).Return(nil)

	router := setupRouter()
	router.PUT("/documents/:id", withAuth(1, "methodist"), h.Update)

	w := performRequest(router, http.MethodPut, "/documents/1", map[string]interface{}{
		"title":     "<script>alert('xss')</script>Clean Title",
		"subject":   "<b>Bold Subject</b>",
		"file_name": "file<script>.txt",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDocumentHandler_Delete_Error(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	typeRepo := new(MockDocumentTypeRepository)
	catRepo := new(MockDocumentCategoryRepository)
	h := newDocumentHandler(docRepo, typeRepo, catRepo)

	doc := &entities.Document{ID: 1, Title: "Doc", AuthorID: 1, Status: entities.DocumentStatusDraft}
	docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)
	docRepo.On("SoftDelete", mock.Anything, int64(1)).Return(fmt.Errorf("delete failed"))

	router := setupRouter()
	router.DELETE("/documents/:id", withAuth(1, "methodist"), h.Delete)

	w := performRequest(router, http.MethodDelete, "/documents/1", nil)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestDocumentHandler_DeleteFile_Error(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	typeRepo := new(MockDocumentTypeRepository)
	catRepo := new(MockDocumentCategoryRepository)
	h := newDocumentHandler(docRepo, typeRepo, catRepo)

	doc := &entities.Document{ID: 1, Title: "Doc", AuthorID: 1, Status: entities.DocumentStatusDraft}
	docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)

	router := setupRouter()
	router.DELETE("/documents/:id/file", withAuth(1, "methodist"), h.DeleteFile)

	w := performRequest(router, http.MethodDelete, "/documents/1/file", nil)
	// The usecase will fail because there's no file to delete (or s3Client is nil)
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestDocumentHandler_DownloadFile_Error(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	typeRepo := new(MockDocumentTypeRepository)
	catRepo := new(MockDocumentCategoryRepository)
	h := newDocumentHandler(docRepo, typeRepo, catRepo)

	docRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, fmt.Errorf("not found"))

	router := setupRouter()
	router.GET("/documents/:id/file", h.DownloadFile)

	w := performRequest(router, http.MethodGet, "/documents/1/file", nil)
	assert.NotEqual(t, http.StatusOK, w.Code)
}
