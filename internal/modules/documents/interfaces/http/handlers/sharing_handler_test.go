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
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

func newSharingHandler(docRepo *MockDocumentRepository, permRepo *MockPermissionRepository, linkRepo *MockPublicLinkRepository) *handlers.SharingHandler {
	uc := usecases.NewSharingUseCase(docRepo, permRepo, linkRepo, nil, "http://localhost", nil)
	v := validation.NewValidator()
	return handlers.NewSharingHandler(uc, v)
}

func TestSharingHandler_ShareDocument(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.POST("/documents/:id/share", h.ShareDocument)

		w := performRequest(router, http.MethodPost, "/documents/1/share", map[string]interface{}{
			"user_id":    2,
			"permission": "read",
		})
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.POST("/documents/:id/share", withAuth(1, "methodist"), h.ShareDocument)

		w := performRequest(router, http.MethodPost, "/documents/abc/share", map[string]interface{}{
			"user_id":    2,
			"permission": "read",
		})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.POST("/documents/:id/share", withAuth(1, "methodist"), h.ShareDocument)

		w := performRequest(router, http.MethodPost, "/documents/1/share", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		permRepo := new(MockPermissionRepository)
		linkRepo := new(MockPublicLinkRepository)
		h := newSharingHandler(docRepo, permRepo, linkRepo)

		doc := &entities.Document{ID: 1, Title: "Doc", AuthorID: 1, Status: entities.DocumentStatusDraft, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)

		userID := int64(2)
		permRepo.On("GetByDocumentAndUser", mock.Anything, int64(1), userID).Return(nil, fmt.Errorf("not found"))
		permRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.DocumentPermission")).Run(func(args mock.Arguments) {
			p := args.Get(1).(*entities.DocumentPermission)
			p.ID = 10
		}).Return(nil)
		grantedBy := int64(1)
		permRepo.On("GetByID", mock.Anything, int64(10)).Return(&entities.DocumentPermission{
			ID: 10, DocumentID: 1, UserID: &userID, Permission: "read", GrantedBy: &grantedBy, CreatedAt: time.Now(),
		}, nil)

		router := setupRouter()
		router.POST("/documents/:id/share", withAuth(1, "methodist"), h.ShareDocument)

		w := performRequest(router, http.MethodPost, "/documents/1/share", map[string]interface{}{
			"user_id":    2,
			"permission": "read",
		})
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestSharingHandler_RevokePermission(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.DELETE("/documents/:id/permissions/:permissionId", h.RevokePermission)

		w := performRequest(router, http.MethodDelete, "/documents/1/permissions/1", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid permission id", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.DELETE("/documents/:id/permissions/:permissionId", withAuth(1, "methodist"), h.RevokePermission)

		w := performRequest(router, http.MethodDelete, "/documents/1/permissions/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		permRepo := new(MockPermissionRepository)
		linkRepo := new(MockPublicLinkRepository)
		h := newSharingHandler(docRepo, permRepo, linkRepo)

		grantedBy := int64(1)
		perm := &entities.DocumentPermission{ID: 1, DocumentID: 1, GrantedBy: &grantedBy}
		permRepo.On("GetByID", mock.Anything, int64(1)).Return(perm, nil)
		doc := &entities.Document{ID: 1, AuthorID: 1}
		docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)
		permRepo.On("Delete", mock.Anything, int64(1)).Return(nil)

		router := setupRouter()
		router.DELETE("/documents/:id/permissions/:permissionId", withAuth(1, "methodist"), h.RevokePermission)

		w := performRequest(router, http.MethodDelete, "/documents/1/permissions/1", nil)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestSharingHandler_GetDocumentPermissions(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.GET("/documents/:id/permissions", h.GetDocumentPermissions)

		w := performRequest(router, http.MethodGet, "/documents/1/permissions", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.GET("/documents/:id/permissions", withAuth(1, "methodist"), h.GetDocumentPermissions)

		w := performRequest(router, http.MethodGet, "/documents/abc/permissions", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		permRepo := new(MockPermissionRepository)
		linkRepo := new(MockPublicLinkRepository)
		h := newSharingHandler(docRepo, permRepo, linkRepo)

		doc := &entities.Document{ID: 1, AuthorID: 1}
		docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)
		perms := []*entities.DocumentPermission{{ID: 1, DocumentID: 1}}
		permRepo.On("GetByDocumentID", mock.Anything, int64(1)).Return(perms, nil)

		router := setupRouter()
		router.GET("/documents/:id/permissions", withAuth(1, "methodist"), h.GetDocumentPermissions)

		w := performRequest(router, http.MethodGet, "/documents/1/permissions", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSharingHandler_CreatePublicLink(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.POST("/documents/:id/public-links", h.CreatePublicLink)

		w := performRequest(router, http.MethodPost, "/documents/1/public-links", map[string]interface{}{
			"permission": "read",
		})
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.POST("/documents/:id/public-links", withAuth(1, "methodist"), h.CreatePublicLink)

		w := performRequest(router, http.MethodPost, "/documents/abc/public-links", map[string]interface{}{
			"permission": "read",
		})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.POST("/documents/:id/public-links", withAuth(1, "methodist"), h.CreatePublicLink)

		w := performRequest(router, http.MethodPost, "/documents/1/public-links", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestSharingHandler_GetDocumentPublicLinks(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.GET("/documents/:id/public-links", h.GetDocumentPublicLinks)

		w := performRequest(router, http.MethodGet, "/documents/1/public-links", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.GET("/documents/:id/public-links", withAuth(1, "methodist"), h.GetDocumentPublicLinks)

		w := performRequest(router, http.MethodGet, "/documents/abc/public-links", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		permRepo := new(MockPermissionRepository)
		linkRepo := new(MockPublicLinkRepository)
		h := newSharingHandler(docRepo, permRepo, linkRepo)

		doc := &entities.Document{ID: 1, AuthorID: 1}
		docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)
		links := []*entities.PublicLink{{ID: 1, DocumentID: 1}}
		linkRepo.On("GetByDocumentID", mock.Anything, int64(1)).Return(links, nil)

		router := setupRouter()
		router.GET("/documents/:id/public-links", withAuth(1, "methodist"), h.GetDocumentPublicLinks)

		w := performRequest(router, http.MethodGet, "/documents/1/public-links", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSharingHandler_DeactivatePublicLink(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.POST("/documents/:id/public-links/:linkId/deactivate", h.DeactivatePublicLink)

		w := performRequest(router, http.MethodPost, "/documents/1/public-links/1/deactivate", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid link id", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.POST("/documents/:id/public-links/:linkId/deactivate", withAuth(1, "methodist"), h.DeactivatePublicLink)

		w := performRequest(router, http.MethodPost, "/documents/1/public-links/abc/deactivate", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		permRepo := new(MockPermissionRepository)
		linkRepo := new(MockPublicLinkRepository)
		h := newSharingHandler(docRepo, permRepo, linkRepo)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 1}
		linkRepo.On("GetByID", mock.Anything, int64(1)).Return(link, nil)
		doc := &entities.Document{ID: 1, AuthorID: 1}
		docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)
		linkRepo.On("Deactivate", mock.Anything, int64(1)).Return(nil)

		router := setupRouter()
		router.POST("/documents/:id/public-links/:linkId/deactivate", withAuth(1, "methodist"), h.DeactivatePublicLink)

		w := performRequest(router, http.MethodPost, "/documents/1/public-links/1/deactivate", nil)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestSharingHandler_DeletePublicLink(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.DELETE("/documents/:id/public-links/:linkId", h.DeletePublicLink)

		w := performRequest(router, http.MethodDelete, "/documents/1/public-links/1", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid link id", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.DELETE("/documents/:id/public-links/:linkId", withAuth(1, "methodist"), h.DeletePublicLink)

		w := performRequest(router, http.MethodDelete, "/documents/1/public-links/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestSharingHandler_AccessPublicDocument(t *testing.T) {
	t.Run("empty token", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.POST("/public/documents/:token", h.AccessPublicDocument)

		w := performRequest(router, http.MethodPost, "/public/documents/", nil)
		// gin won't match empty param in this route, so we get 404
		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		permRepo := new(MockPermissionRepository)
		linkRepo := new(MockPublicLinkRepository)
		h := newSharingHandler(docRepo, permRepo, linkRepo)

		linkRepo.On("GetByToken", mock.Anything, "invalid-token").Return(nil, fmt.Errorf("not found"))

		router := setupRouter()
		router.POST("/public/documents/:token", h.AccessPublicDocument)

		w := performRequest(router, http.MethodPost, "/public/documents/invalid-token", nil)
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestSharingHandler_GetSharedDocuments(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.GET("/documents/shared", h.GetSharedDocuments)

		w := performRequest(router, http.MethodGet, "/documents/shared", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		permRepo := new(MockPermissionRepository)
		linkRepo := new(MockPublicLinkRepository)
		h := newSharingHandler(docRepo, permRepo, linkRepo)

		perms := []*entities.DocumentPermission{
			{ID: 1, DocumentID: 1, UserID: ptrInt64(1)},
		}
		permRepo.On("GetByUserIDOrRole", mock.Anything, int64(1), "methodist").Return(perms, nil)
		doc := &entities.Document{ID: 1, Title: "Shared", Status: entities.DocumentStatusDraft, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)

		router := setupRouter()
		router.GET("/documents/shared", withAuth(1, "methodist"), h.GetSharedDocuments)

		w := performRequest(router, http.MethodGet, "/documents/shared", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSharingHandler_GetMySharedDocuments(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		h := newSharingHandler(new(MockDocumentRepository), new(MockPermissionRepository), new(MockPublicLinkRepository))
		router := setupRouter()
		router.GET("/documents/my-shared", h.GetMySharedDocuments)

		w := performRequest(router, http.MethodGet, "/documents/my-shared", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		permRepo := new(MockPermissionRepository)
		linkRepo := new(MockPublicLinkRepository)
		h := newSharingHandler(docRepo, permRepo, linkRepo)

		perms := []*entities.DocumentPermission{}
		permRepo.On("GetByGrantedBy", mock.Anything, int64(1)).Return(perms, nil)

		router := setupRouter()
		router.GET("/documents/my-shared", withAuth(1, "methodist"), h.GetMySharedDocuments)

		w := performRequest(router, http.MethodGet, "/documents/my-shared", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("with pagination", func(t *testing.T) {
		docRepo := new(MockDocumentRepository)
		permRepo := new(MockPermissionRepository)
		linkRepo := new(MockPublicLinkRepository)
		h := newSharingHandler(docRepo, permRepo, linkRepo)

		perms := []*entities.DocumentPermission{}
		permRepo.On("GetByGrantedBy", mock.Anything, int64(1)).Return(perms, nil)

		router := setupRouter()
		router.GET("/documents/my-shared", withAuth(1, "methodist"), h.GetMySharedDocuments)

		w := performRequest(router, http.MethodGet, "/documents/my-shared?limit=5&offset=0", nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func ptrInt64(v int64) *int64 {
	return &v
}

func TestSharingHandler_ShareDocument_UsecaseError(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	permRepo := new(MockPermissionRepository)
	linkRepo := new(MockPublicLinkRepository)
	h := newSharingHandler(docRepo, permRepo, linkRepo)

	docRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, fmt.Errorf("not found"))

	router := setupRouter()
	router.POST("/documents/:id/share", withAuth(1, "methodist"), h.ShareDocument)

	w := performRequest(router, http.MethodPost, "/documents/1/share", map[string]interface{}{
		"user_id":    2,
		"permission": "read",
	})
	assert.NotEqual(t, http.StatusCreated, w.Code)
}

func TestSharingHandler_RevokePermission_Error(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	permRepo := new(MockPermissionRepository)
	linkRepo := new(MockPublicLinkRepository)
	h := newSharingHandler(docRepo, permRepo, linkRepo)

	permRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, fmt.Errorf("not found"))

	router := setupRouter()
	router.DELETE("/documents/:id/permissions/:permissionId", withAuth(1, "methodist"), h.RevokePermission)

	w := performRequest(router, http.MethodDelete, "/documents/1/permissions/1", nil)
	assert.NotEqual(t, http.StatusNoContent, w.Code)
}

func TestSharingHandler_GetDocumentPermissions_Error(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	permRepo := new(MockPermissionRepository)
	linkRepo := new(MockPublicLinkRepository)
	h := newSharingHandler(docRepo, permRepo, linkRepo)

	docRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, fmt.Errorf("not found"))

	router := setupRouter()
	router.GET("/documents/:id/permissions", withAuth(1, "methodist"), h.GetDocumentPermissions)

	w := performRequest(router, http.MethodGet, "/documents/1/permissions", nil)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestSharingHandler_CreatePublicLink_Success(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	permRepo := new(MockPermissionRepository)
	linkRepo := new(MockPublicLinkRepository)
	h := newSharingHandler(docRepo, permRepo, linkRepo)

	doc := &entities.Document{ID: 1, Title: "Doc", AuthorID: 1}
	docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)
	linkRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.PublicLink")).Run(func(args mock.Arguments) {
		link := args.Get(1).(*entities.PublicLink)
		link.ID = 1
	}).Return(nil)
	linkRepo.On("GetByID", mock.Anything, int64(1)).Return(&entities.PublicLink{
		ID: 1, DocumentID: 1, Token: "test-token", Permission: "read",
		IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}, nil)

	router := setupRouter()
	router.POST("/documents/:id/public-links", withAuth(1, "methodist"), h.CreatePublicLink)

	w := performRequest(router, http.MethodPost, "/documents/1/public-links", map[string]interface{}{
		"permission": "read",
	})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestSharingHandler_DeletePublicLink_Success(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	permRepo := new(MockPermissionRepository)
	linkRepo := new(MockPublicLinkRepository)
	h := newSharingHandler(docRepo, permRepo, linkRepo)

	link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 1}
	linkRepo.On("GetByID", mock.Anything, int64(1)).Return(link, nil)
	doc := &entities.Document{ID: 1, AuthorID: 1}
	docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)
	linkRepo.On("Delete", mock.Anything, int64(1)).Return(nil)

	router := setupRouter()
	router.DELETE("/documents/:id/public-links/:linkId", withAuth(1, "methodist"), h.DeletePublicLink)

	w := performRequest(router, http.MethodDelete, "/documents/1/public-links/1", nil)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestSharingHandler_AccessPublicDocument_WithToken(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	permRepo := new(MockPermissionRepository)
	linkRepo := new(MockPublicLinkRepository)
	h := newSharingHandler(docRepo, permRepo, linkRepo)

	link := &entities.PublicLink{
		ID: 1, DocumentID: 1, Token: "valid-token", Permission: "read",
		IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	linkRepo.On("GetByToken", mock.Anything, "valid-token").Return(link, nil)
	doc := &entities.Document{ID: 1, Title: "Doc", Status: entities.DocumentStatusDraft, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	docRepo.On("GetByID", mock.Anything, int64(1)).Return(doc, nil)
	linkRepo.On("IncrementUseCount", mock.Anything, int64(1)).Return(nil)

	router := setupRouter()
	router.POST("/public/documents/:token", h.AccessPublicDocument)

	w := performRequest(router, http.MethodPost, "/public/documents/valid-token", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSharingHandler_GetSharedDocuments_Error(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	permRepo := new(MockPermissionRepository)
	linkRepo := new(MockPublicLinkRepository)
	h := newSharingHandler(docRepo, permRepo, linkRepo)

	permRepo.On("GetByUserIDOrRole", mock.Anything, int64(1), "methodist").Return(nil, fmt.Errorf("error"))

	router := setupRouter()
	router.GET("/documents/shared", withAuth(1, "methodist"), h.GetSharedDocuments)

	w := performRequest(router, http.MethodGet, "/documents/shared", nil)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestSharingHandler_GetSharedDocuments_WithFilters(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	permRepo := new(MockPermissionRepository)
	linkRepo := new(MockPublicLinkRepository)
	h := newSharingHandler(docRepo, permRepo, linkRepo)

	permRepo.On("GetByUserIDOrRole", mock.Anything, int64(1), "methodist").Return([]*entities.DocumentPermission{}, nil)

	router := setupRouter()
	router.GET("/documents/shared", withAuth(1, "methodist"), h.GetSharedDocuments)

	w := performRequest(router, http.MethodGet, "/documents/shared?permission=read&limit=5&offset=10", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSharingHandler_GetMySharedDocuments_Error(t *testing.T) {
	docRepo := new(MockDocumentRepository)
	permRepo := new(MockPermissionRepository)
	linkRepo := new(MockPublicLinkRepository)
	h := newSharingHandler(docRepo, permRepo, linkRepo)

	permRepo.On("GetByGrantedBy", mock.Anything, int64(1)).Return(nil, fmt.Errorf("error"))

	router := setupRouter()
	router.GET("/documents/my-shared", withAuth(1, "methodist"), h.GetMySharedDocuments)

	w := performRequest(router, http.MethodGet, "/documents/my-shared", nil)
	assert.NotEqual(t, http.StatusOK, w.Code)
}
