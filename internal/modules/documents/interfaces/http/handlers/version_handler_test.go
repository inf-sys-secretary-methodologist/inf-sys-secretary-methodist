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

func newVersionHandler(docRepo *MockDocumentRepository) *handlers.VersionHandler {
	uc := usecases.NewDocumentVersionUseCase(docRepo, nil, nil)
	return handlers.NewVersionHandler(uc)
}

func TestVersionHandler_GetVersions(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.GET("/documents/:id/versions", h.GetVersions)

		w := performRequest(router, http.MethodGet, "/documents/1/versions", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.GET("/documents/:id/versions", withAuth(1, "methodist"), h.GetVersions)

		w := performRequest(router, http.MethodGet, "/documents/abc/versions", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		h := newVersionHandler(docRepo)

		doc := &entities.Document{ID: 1, Title: "Doc", AuthorID: 1, Version: 1, Status: entities.DocumentStatusDraft}
		docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)
		versions := []*entities.DocumentVersion{
			{ID: 1, DocumentID: 1, Version: 1, CreatedAt: time.Now()},
		}
		docRepo.On("GetVersions", mock.Anything, int64(1)).Return(versions, nil)

		router := setupRouter()
		router.GET("/documents/:id/versions", withAuth(1, "methodist"), h.GetVersions)

		w := performRequest(router, http.MethodGet, "/documents/1/versions", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		h := newVersionHandler(docRepo)

		docRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, fmt.Errorf("not found"))

		router := setupRouter()
		router.GET("/documents/:id/versions", withAuth(1, "methodist"), h.GetVersions)

		w := performRequest(router, http.MethodGet, "/documents/999/versions", nil)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestVersionHandler_GetVersion(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.GET("/documents/:id/versions/:version", h.GetVersion)

		w := performRequest(router, http.MethodGet, "/documents/1/versions/1", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.GET("/documents/:id/versions/:version", withAuth(1, "methodist"), h.GetVersion)

		w := performRequest(router, http.MethodGet, "/documents/abc/versions/1", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid version", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.GET("/documents/:id/versions/:version", withAuth(1, "methodist"), h.GetVersion)

		w := performRequest(router, http.MethodGet, "/documents/1/versions/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		h := newVersionHandler(docRepo)

		doc := &entities.Document{ID: 1, AuthorID: 1}
		docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)
		ver := &entities.DocumentVersion{ID: 1, DocumentID: 1, Version: 1, CreatedAt: time.Now()}
		docRepo.On("GetVersion", mock.Anything, int64(1), 1).Return(ver, nil)

		router := setupRouter()
		router.GET("/documents/:id/versions/:version", withAuth(1, "methodist"), h.GetVersion)

		w := performRequest(router, http.MethodGet, "/documents/1/versions/1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestVersionHandler_CreateVersion(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.POST("/documents/:id/versions", h.CreateVersion)

		w := performRequest(router, http.MethodPost, "/documents/1/versions", map[string]interface{}{})
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.POST("/documents/:id/versions", withAuth(1, "methodist"), h.CreateVersion)

		w := performRequest(router, http.MethodPost, "/documents/abc/versions", map[string]interface{}{})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.POST("/documents/:id/versions", withAuth(1, "methodist"), h.CreateVersion)

		w := performRequest(router, http.MethodPost, "/documents/1/versions", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		h := newVersionHandler(docRepo)

		doc := &entities.Document{ID: 1, AuthorID: 1, Version: 1, Status: entities.DocumentStatusDraft}
		docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)
		docRepo.On("GetLatestVersion", mock.Anything, int64(1)).Return(nil, fmt.Errorf("no versions"))
		docRepo.On("CreateVersion", mock.Anything, mock.AnythingOfType("*entities.DocumentVersion")).Run(func(args mock.Arguments) {
			v := args.Get(1).(*entities.DocumentVersion)
			v.ID = 1
		}).Return(nil)
		docRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Document")).Return(nil)
		docRepo.On("AddHistory", mock.Anything, mock.Anything).Return(nil)

		router := setupRouter()
		router.POST("/documents/:id/versions", withAuth(1, "methodist"), h.CreateVersion)

		w := performRequest(router, http.MethodPost, "/documents/1/versions", map[string]interface{}{
			"change_description": "manual snapshot",
		})
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestVersionHandler_RestoreVersion(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.POST("/documents/:id/versions/:version/restore", h.RestoreVersion)

		w := performRequest(router, http.MethodPost, "/documents/1/versions/1/restore", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.POST("/documents/:id/versions/:version/restore", withAuth(1, "methodist"), h.RestoreVersion)

		w := performRequest(router, http.MethodPost, "/documents/abc/versions/1/restore", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid version", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.POST("/documents/:id/versions/:version/restore", withAuth(1, "methodist"), h.RestoreVersion)

		w := performRequest(router, http.MethodPost, "/documents/1/versions/abc/restore", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestVersionHandler_CompareVersions(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.GET("/documents/:id/versions/compare", h.CompareVersions)

		w := performRequest(router, http.MethodGet, "/documents/1/versions/compare?from=1&to=2", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.GET("/documents/:id/versions/compare", withAuth(1, "methodist"), h.CompareVersions)

		w := performRequest(router, http.MethodGet, "/documents/abc/versions/compare?from=1&to=2", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing from", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.GET("/documents/:id/versions/compare", withAuth(1, "methodist"), h.CompareVersions)

		w := performRequest(router, http.MethodGet, "/documents/1/versions/compare?to=2", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing to", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.GET("/documents/:id/versions/compare", withAuth(1, "methodist"), h.CompareVersions)

		w := performRequest(router, http.MethodGet, "/documents/1/versions/compare?from=1", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestVersionHandler_DeleteVersion(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.DELETE("/documents/:id/versions/:version", h.DeleteVersion)

		w := performRequest(router, http.MethodDelete, "/documents/1/versions/1", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.DELETE("/documents/:id/versions/:version", withAuth(1, "methodist"), h.DeleteVersion)

		w := performRequest(router, http.MethodDelete, "/documents/abc/versions/1", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid version", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.DELETE("/documents/:id/versions/:version", withAuth(1, "methodist"), h.DeleteVersion)

		w := performRequest(router, http.MethodDelete, "/documents/1/versions/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestVersionHandler_GetVersionFile(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.GET("/documents/:id/versions/:version/file", h.GetVersionFile)

		w := performRequest(router, http.MethodGet, "/documents/1/versions/1/file", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid doc id", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.GET("/documents/:id/versions/:version/file", withAuth(1, "methodist"), h.GetVersionFile)

		w := performRequest(router, http.MethodGet, "/documents/abc/versions/1/file", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid version", func(t *testing.T) {
		h := newVersionHandler(new(MockDocumentRepository))
		router := setupRouter()
		router.GET("/documents/:id/versions/:version/file", withAuth(1, "methodist"), h.GetVersionFile)

		w := performRequest(router, http.MethodGet, "/documents/1/versions/abc/file", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
