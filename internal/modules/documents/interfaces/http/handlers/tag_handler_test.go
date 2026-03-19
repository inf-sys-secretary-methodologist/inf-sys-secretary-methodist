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
	handlers "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/interfaces/http/handlers"
)

func newTagHandler(tagRepo *MockDocumentTagRepository, docRepo *MockDocumentRepository) *handlers.TagHandler {
	// Always set up GetTagUsageCount as a fallback since it's called in many places
	tagRepo.On("GetTagUsageCount", mock.Anything, mock.Anything).Maybe().Return(int64(0), nil)
	uc := usecases.NewTagUseCase(tagRepo, docRepo, nil)
	return handlers.NewTagHandler(uc)
}

func TestTagHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		tagRepo.On("GetByName", mock.Anything, "TestTag").Return(nil, fmt.Errorf("not found"))
		tagRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.DocumentTag")).Run(func(args mock.Arguments) {
			tag := args.Get(1).(*entities.DocumentTag)
			tag.ID = 1
			tag.CreatedAt = time.Now()
		}).Return(nil)

		router := setupRouter()
		router.POST("/tags", h.Create)

		w := performRequest(router, http.MethodPost, "/tags", map[string]interface{}{
			"name":  "TestTag",
			"color": "#FF0000",
		})
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		router := setupRouter()
		router.POST("/tags", h.Create)

		w := performRequest(router, http.MethodPost, "/tags", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("validation error - empty name", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		router := setupRouter()
		router.POST("/tags", h.Create)

		w := performRequest(router, http.MethodPost, "/tags", map[string]interface{}{
			"name": "",
		})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		tagRepo.On("GetByName", mock.Anything, "TestTag").Return(nil, fmt.Errorf("not found"))
		tagRepo.On("Create", mock.Anything, mock.Anything).Return(fmt.Errorf("duplicate key"))

		router := setupRouter()
		router.POST("/tags", h.Create)

		w := performRequest(router, http.MethodPost, "/tags", map[string]interface{}{
			"name":  "TestTag",
			"color": "#FF0000",
		})
		assert.NotEqual(t, http.StatusCreated, w.Code)
	})
}

func TestTagHandler_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		tag := &entities.DocumentTag{ID: 1, Name: "Tag1", CreatedAt: time.Now()}
		tagRepo.On("GetByID", mock.Anything, int64(1)).Return(tag, nil)
		tagRepo.On("GetTagUsageCount", mock.Anything, int64(1)).Return(int64(5), nil)

		router := setupRouter()
		router.GET("/tags/:id", h.GetByID)

		w := performRequest(router, http.MethodGet, "/tags/1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		router := setupRouter()
		router.GET("/tags/:id", h.GetByID)

		w := performRequest(router, http.MethodGet, "/tags/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		tagRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, fmt.Errorf("not found"))

		router := setupRouter()
		router.GET("/tags/:id", h.GetByID)

		w := performRequest(router, http.MethodGet, "/tags/999", nil)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestTagHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		tag := &entities.DocumentTag{ID: 1, Name: "Old", CreatedAt: time.Now()}
		tagRepo.On("GetByID", mock.Anything, int64(1)).Return(tag, nil)
		tagRepo.On("GetByName", mock.Anything, "New").Return(nil, fmt.Errorf("not found"))
		tagRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.DocumentTag")).Return(nil)
		tagRepo.On("GetTagUsageCount", mock.Anything, int64(1)).Return(int64(0), nil)

		router := setupRouter()
		router.PUT("/tags/:id", h.Update)

		w := performRequest(router, http.MethodPut, "/tags/1", map[string]interface{}{"name": "New"})
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		router := setupRouter()
		router.PUT("/tags/:id", h.Update)

		w := performRequest(router, http.MethodPut, "/tags/abc", map[string]interface{}{"name": "New"})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		router := setupRouter()
		router.PUT("/tags/:id", h.Update)

		w := performRequest(router, http.MethodPut, "/tags/1", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTagHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		tag := &entities.DocumentTag{ID: 1, Name: "Tag1"}
		tagRepo.On("GetByID", mock.Anything, int64(1)).Return(tag, nil)
		tagRepo.On("Delete", mock.Anything, int64(1)).Return(nil)

		router := setupRouter()
		router.DELETE("/tags/:id", h.Delete)

		w := performRequest(router, http.MethodDelete, "/tags/1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		router := setupRouter()
		router.DELETE("/tags/:id", h.Delete)

		w := performRequest(router, http.MethodDelete, "/tags/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("error", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		tag := &entities.DocumentTag{ID: 1, Name: "Tag1"}
		tagRepo.On("GetByID", mock.Anything, int64(1)).Return(tag, nil)
		tagRepo.On("Delete", mock.Anything, int64(1)).Return(fmt.Errorf("error"))

		router := setupRouter()
		router.DELETE("/tags/:id", h.Delete)

		w := performRequest(router, http.MethodDelete, "/tags/1", nil)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestTagHandler_GetAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		tags := []*entities.DocumentTag{{ID: 1, Name: "Tag1", CreatedAt: time.Now()}}
		tagRepo.On("GetAll", mock.Anything).Return(tags, nil)
		tagRepo.On("GetTagUsageCount", mock.Anything, int64(1)).Return(int64(0), nil)

		router := setupRouter()
		router.GET("/tags", h.GetAll)

		w := performRequest(router, http.MethodGet, "/tags", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		tagRepo.On("GetAll", mock.Anything).Return(nil, fmt.Errorf("error"))

		router := setupRouter()
		router.GET("/tags", h.GetAll)

		w := performRequest(router, http.MethodGet, "/tags", nil)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestTagHandler_Search(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		tags := []*entities.DocumentTag{{ID: 1, Name: "Test", CreatedAt: time.Now()}}
		tagRepo.On("Search", mock.Anything, "test", 10).Return(tags, nil)

		router := setupRouter()
		router.GET("/tags/search", h.Search)

		w := performRequest(router, http.MethodGet, "/tags/search?q=test", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("empty query", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		router := setupRouter()
		router.GET("/tags/search", h.Search)

		w := performRequest(router, http.MethodGet, "/tags/search", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("with limit", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		tags := []*entities.DocumentTag{}
		tagRepo.On("Search", mock.Anything, "test", 5).Return(tags, nil)

		router := setupRouter()
		router.GET("/tags/search", h.Search)

		w := performRequest(router, http.MethodGet, "/tags/search?q=test&limit=5", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		tagRepo.On("Search", mock.Anything, "test", 10).Return(nil, fmt.Errorf("error"))

		router := setupRouter()
		router.GET("/tags/search", h.Search)

		w := performRequest(router, http.MethodGet, "/tags/search?q=test", nil)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestTagHandler_GetDocumentTags(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		doc := &entities.Document{ID: 1, Title: "Doc"}
		docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)
		tags := []*entities.DocumentTag{{ID: 1, Name: "Tag1"}}
		tagRepo.On("GetTagsByDocumentID", mock.Anything, int64(1)).Return(tags, nil)

		router := setupRouter()
		router.GET("/documents/:id/tags", h.GetDocumentTags)

		w := performRequest(router, http.MethodGet, "/documents/1/tags", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		router := setupRouter()
		router.GET("/documents/:id/tags", h.GetDocumentTags)

		w := performRequest(router, http.MethodGet, "/documents/abc/tags", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTagHandler_SetDocumentTags(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		doc := &entities.Document{ID: 1, Title: "Doc"}
		docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)
		tagRepo.On("GetByID", mock.Anything, int64(1)).Return(&entities.DocumentTag{ID: 1, Name: "Tag1"}, nil)
		tagRepo.On("GetByID", mock.Anything, int64(2)).Return(&entities.DocumentTag{ID: 2, Name: "Tag2"}, nil)
		tagRepo.On("SetDocumentTags", mock.Anything, int64(1), mock.Anything).Return(nil)
		tags := []*entities.DocumentTag{{ID: 1, Name: "Tag1"}, {ID: 2, Name: "Tag2"}}
		tagRepo.On("GetTagsByDocumentID", mock.Anything, int64(1)).Return(tags, nil)

		router := setupRouter()
		router.PUT("/documents/:id/tags", h.SetDocumentTags)

		w := performRequest(router, http.MethodPut, "/documents/1/tags", map[string]interface{}{
			"tag_ids": []int64{1, 2},
		})
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		router := setupRouter()
		router.PUT("/documents/:id/tags", h.SetDocumentTags)

		w := performRequest(router, http.MethodPut, "/documents/abc/tags", map[string]interface{}{
			"tag_ids": []int64{1},
		})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		router := setupRouter()
		router.PUT("/documents/:id/tags", h.SetDocumentTags)

		w := performRequest(router, http.MethodPut, "/documents/1/tags", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTagHandler_AddTagToDocument(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		docRepo.On("GetByID", mock.Anything, int64(1)).Return(&entities.Document{ID: 1, Title: "Doc"}, nil)
		tagRepo.On("GetByID", mock.Anything, int64(2)).Return(&entities.DocumentTag{ID: 2, Name: "Tag2"}, nil)
		tagRepo.On("AddTagToDocument", mock.Anything, int64(1), int64(2)).Return(nil)

		router := setupRouter()
		router.POST("/documents/:id/tags/:tag_id", h.AddTagToDocument)

		w := performRequest(router, http.MethodPost, "/documents/1/tags/2", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid document id", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		router := setupRouter()
		router.POST("/documents/:id/tags/:tag_id", h.AddTagToDocument)

		w := performRequest(router, http.MethodPost, "/documents/abc/tags/2", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid tag id", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		router := setupRouter()
		router.POST("/documents/:id/tags/:tag_id", h.AddTagToDocument)

		w := performRequest(router, http.MethodPost, "/documents/1/tags/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTagHandler_RemoveTagFromDocument(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		tagRepo.On("RemoveTagFromDocument", mock.Anything, int64(1), int64(2)).Return(nil)

		router := setupRouter()
		router.DELETE("/documents/:id/tags/:tag_id", h.RemoveTagFromDocument)

		w := performRequest(router, http.MethodDelete, "/documents/1/tags/2", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid document id", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		router := setupRouter()
		router.DELETE("/documents/:id/tags/:tag_id", h.RemoveTagFromDocument)

		w := performRequest(router, http.MethodDelete, "/documents/abc/tags/2", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid tag id", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		router := setupRouter()
		router.DELETE("/documents/:id/tags/:tag_id", h.RemoveTagFromDocument)

		w := performRequest(router, http.MethodDelete, "/documents/1/tags/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTagHandler_GetDocumentsByTag(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		tag := &entities.DocumentTag{ID: 1, Name: "Tag1", CreatedAt: time.Now()}
		tagRepo.On("GetByID", mock.Anything, int64(1)).Return(tag, nil)
		tagRepo.On("GetDocumentsByTagID", mock.Anything, int64(1), 20, 0).Return([]int64{1, 2}, int64(2), nil)

		router := setupRouter()
		router.GET("/tags/:id/documents", h.GetDocumentsByTag)

		w := performRequest(router, http.MethodGet, "/tags/1/documents", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		router := setupRouter()
		router.GET("/tags/:id/documents", h.GetDocumentsByTag)

		w := performRequest(router, http.MethodGet, "/tags/abc/documents", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("with pagination", func(t *testing.T) {
		tagRepo := new(MockDocumentTagRepository)
		docRepo := new(MockDocumentRepository)
		h := newTagHandler(tagRepo, docRepo)

		tag := &entities.DocumentTag{ID: 1, Name: "Tag1", CreatedAt: time.Now()}
		tagRepo.On("GetByID", mock.Anything, int64(1)).Return(tag, nil)
		tagRepo.On("GetDocumentsByTagID", mock.Anything, int64(1), 10, 10).Return([]int64{}, int64(0), nil)

		router := setupRouter()
		router.GET("/tags/:id/documents", h.GetDocumentsByTag)

		w := performRequest(router, http.MethodGet, "/tags/1/documents?page=2&page_size=10", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
