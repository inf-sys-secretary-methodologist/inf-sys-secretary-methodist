package handlers

// v0.153.11 Phase 6 #196 backfill — extends file_handler_test.go beyond
// the parsing-error branches by wiring a real FileUseCase with an in-memory
// FileMetadataRepository. Covers List + CleanupExpired (both 0% before)
// plus happy-path branches for GetByDocument/GetByTask/GetByAnnouncement
// /Delete/Attach.
//
// FileVersionRepository / storageClient / auditLogger are intentionally
// nil — touched only by paths not exercised here (Upload/Download/Version*).
// CleanupExpired's storage-delete loop is bypassed by keeping
// GetExpiredTemporaryFiles empty.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/repositories"
)

// fakeFileMetaRepo — in-memory FileMetadataRepository satisfying
// the full interface. Behavior is controlled через map lookups +
// per-method error overrides.
type fakeFileMetaRepo struct {
	files          map[int64]*entities.FileMetadata
	byDocument     map[int64][]*entities.FileMetadata
	byTask         map[int64][]*entities.FileMetadata
	byAnnouncement map[int64][]*entities.FileMetadata
	expired        []*entities.FileMetadata

	listErr           error
	countErr          error
	getByIDErr        error
	updateErr         error
	deleteErr         error
	getByDocumentErr  error
	getByTaskErr      error
	getByAnnErr       error
	getExpiredErr     error
	cleanupExpiredErr error

	count         int64
	cleanupCount  int64
	updateCalled  bool
	deleteCalled  bool
	cleanupCalled bool
}

func (r *fakeFileMetaRepo) Create(_ context.Context, _ *entities.FileMetadata) error { return nil }
func (r *fakeFileMetaRepo) GetByID(_ context.Context, id int64) (*entities.FileMetadata, error) {
	if r.getByIDErr != nil {
		return nil, r.getByIDErr
	}
	f, ok := r.files[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return f, nil
}
func (r *fakeFileMetaRepo) GetByStorageKey(_ context.Context, _ string) (*entities.FileMetadata, error) {
	return nil, nil
}
func (r *fakeFileMetaRepo) Update(_ context.Context, _ *entities.FileMetadata) error {
	r.updateCalled = true
	return r.updateErr
}
func (r *fakeFileMetaRepo) Delete(_ context.Context, _ int64) error {
	r.deleteCalled = true
	return r.deleteErr
}
func (r *fakeFileMetaRepo) HardDelete(_ context.Context, _ int64) error { return nil }
func (r *fakeFileMetaRepo) List(_ context.Context, _, _ int) ([]*entities.FileMetadata, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	out := make([]*entities.FileMetadata, 0, len(r.files))
	for _, f := range r.files {
		out = append(out, f)
	}
	return out, nil
}
func (r *fakeFileMetaRepo) Count(_ context.Context) (int64, error) {
	return r.count, r.countErr
}
func (r *fakeFileMetaRepo) GetByDocumentID(_ context.Context, docID int64) ([]*entities.FileMetadata, error) {
	if r.getByDocumentErr != nil {
		return nil, r.getByDocumentErr
	}
	return r.byDocument[docID], nil
}
func (r *fakeFileMetaRepo) GetByTaskID(_ context.Context, taskID int64) ([]*entities.FileMetadata, error) {
	if r.getByTaskErr != nil {
		return nil, r.getByTaskErr
	}
	return r.byTask[taskID], nil
}
func (r *fakeFileMetaRepo) GetByAnnouncementID(_ context.Context, annID int64) ([]*entities.FileMetadata, error) {
	if r.getByAnnErr != nil {
		return nil, r.getByAnnErr
	}
	return r.byAnnouncement[annID], nil
}
func (r *fakeFileMetaRepo) GetByUploadedBy(_ context.Context, _ int64, _, _ int) ([]*entities.FileMetadata, error) {
	return nil, nil
}
func (r *fakeFileMetaRepo) GetExpiredTemporaryFiles(_ context.Context, _ int) ([]*entities.FileMetadata, error) {
	return r.expired, r.getExpiredErr
}
func (r *fakeFileMetaRepo) CleanupExpired(_ context.Context) (int64, error) {
	r.cleanupCalled = true
	return r.cleanupCount, r.cleanupExpiredErr
}

// satisfy compile-time interface check
var _ repositories.FileMetadataRepository = (*fakeFileMetaRepo)(nil)

// authMW устанавливает user_id + role в Gin context, как production middleware.
//
// role mirrors auth/middleware behavior — defaults to student if caller
// omits it (most tests want the uploader case where role doesn't matter
// for the rule, only id-match does).
func authMW(userID int64, roles ...authDomain.RoleType) gin.HandlerFunc {
	role := authDomain.RoleStudent
	if len(roles) > 0 {
		role = roles[0]
	}
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

func newHandlerWithUC(repo *fakeFileMetaRepo) *FileHandler {
	uc := usecases.NewFileUseCase(repo, nil, nil, nil, nil)
	return NewFileHandler(uc, nil)
}

// ----- List -----

func TestFileHandler_List_HappyPath(t *testing.T) {
	repo := &fakeFileMetaRepo{
		files: map[int64]*entities.FileMetadata{
			1: entities.NewFileMetadata("doc.pdf", "s3/key1", "application/pdf", "abc", 1024, 7),
			2: entities.NewFileMetadata("img.png", "s3/key2", "image/png", "def", 2048, 7),
		},
		count: 2,
	}
	h := newHandlerWithUC(repo)
	r := setupFileRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files?page=1&limit=10", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestFileHandler_List_DefaultsAppliedForBadQuery(t *testing.T) {
	repo := &fakeFileMetaRepo{files: map[int64]*entities.FileMetadata{}, count: 0}
	h := newHandlerWithUC(repo)
	r := setupFileRouter(h)

	w := httptest.NewRecorder()
	// Bad page/limit → strconv.Atoi returns 0; ListFiles applies defaults.
	req := httptest.NewRequest(http.MethodGet, "/files?page=abc&limit=xyz", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "default page/limit applied; List succeeds on empty repo")
}

func TestFileHandler_List_RepoError(t *testing.T) {
	repo := &fakeFileMetaRepo{listErr: errors.New("db down")}
	h := newHandlerWithUC(repo)
	r := setupFileRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files", nil)
	r.ServeHTTP(w, req)

	assert.GreaterOrEqual(t, w.Code, 400, "repo error must surface non-2xx")
}

// ----- CleanupExpired -----

func TestFileHandler_CleanupExpired_HappyPath_NoExpiredFiles(t *testing.T) {
	repo := &fakeFileMetaRepo{expired: nil, cleanupCount: 0}
	h := newHandlerWithUC(repo)
	r := setupFileRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/cleanup", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.True(t, repo.cleanupCalled, "CleanupExpired must be invoked")
}

func TestFileHandler_CleanupExpired_GetExpiredError(t *testing.T) {
	repo := &fakeFileMetaRepo{getExpiredErr: errors.New("db error")}
	h := newHandlerWithUC(repo)
	r := setupFileRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/cleanup", nil)
	r.ServeHTTP(w, req)

	assert.GreaterOrEqual(t, w.Code, 400)
}

// ----- GetByDocument / GetByTask / GetByAnnouncement (happy paths) -----

func TestFileHandler_GetByDocument_HappyPath(t *testing.T) {
	repo := &fakeFileMetaRepo{
		byDocument: map[int64][]*entities.FileMetadata{
			42: {entities.NewFileMetadata("a.pdf", "key", "application/pdf", "h", 100, 7)},
		},
	}
	h := newHandlerWithUC(repo)
	r := setupFileRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/documents/42/files", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestFileHandler_GetByDocument_RepoError(t *testing.T) {
	repo := &fakeFileMetaRepo{getByDocumentErr: errors.New("db error")}
	h := newHandlerWithUC(repo)
	r := setupFileRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/documents/42/files", nil)
	r.ServeHTTP(w, req)

	assert.GreaterOrEqual(t, w.Code, 400)
}

func TestFileHandler_GetByTask_HappyPath(t *testing.T) {
	repo := &fakeFileMetaRepo{
		byTask: map[int64][]*entities.FileMetadata{
			55: {entities.NewFileMetadata("rep.docx", "k", "application/msword", "h", 500, 7)},
		},
	}
	h := newHandlerWithUC(repo)
	r := setupFileRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tasks/55/files", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestFileHandler_GetByAnnouncement_HappyPath(t *testing.T) {
	repo := &fakeFileMetaRepo{
		byAnnouncement: map[int64][]*entities.FileMetadata{
			77: {entities.NewFileMetadata("a.png", "k", "image/png", "h", 200, 7)},
		},
	}
	h := newHandlerWithUC(repo)
	r := setupFileRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/announcements/77/files", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

// ----- Delete (happy + forbidden branches) -----

func TestFileHandler_Delete_HappyPath(t *testing.T) {
	f := entities.NewFileMetadata("doc.pdf", "key", "application/pdf", "h", 100, 7) // uploaded by user 7
	f.ID = 11
	repo := &fakeFileMetaRepo{files: map[int64]*entities.FileMetadata{11: f}}
	h := newHandlerWithUC(repo)

	r := gin.New()
	r.Use(authMW(7))
	r.DELETE("/files/:id", h.Delete)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/files/11", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.True(t, repo.deleteCalled)
}

func TestFileHandler_Delete_Forbidden_WrongOwner(t *testing.T) {
	f := entities.NewFileMetadata("doc.pdf", "key", "application/pdf", "h", 100, 7) // uploaded by user 7
	f.ID = 11
	repo := &fakeFileMetaRepo{files: map[int64]*entities.FileMetadata{11: f}}
	h := newHandlerWithUC(repo)

	r := gin.New()
	r.Use(authMW(99)) // different user
	r.DELETE("/files/:id", h.Delete)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/files/11", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, repo.deleteCalled, "Delete must not be invoked when ownership check fails")
}

// ----- Attach (happy path with document_id + validation error) -----

func TestFileHandler_Attach_HappyPath_Document(t *testing.T) {
	f := entities.NewFileMetadata("doc.pdf", "key", "application/pdf", "h", 100, 7)
	f.ID = 11
	f.IsTemporary = true
	repo := &fakeFileMetaRepo{files: map[int64]*entities.FileMetadata{11: f}}
	h := newHandlerWithUC(repo)
	r := setupFileRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/11/attach", strings.NewReader(`{"document_id": 42}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.True(t, repo.updateCalled)
}

func TestFileHandler_Attach_ValidationError_AlreadyAttached(t *testing.T) {
	f := entities.NewFileMetadata("doc.pdf", "key", "application/pdf", "h", 100, 7)
	f.ID = 11
	f.IsTemporary = false // already attached → ValidationError
	repo := &fakeFileMetaRepo{files: map[int64]*entities.FileMetadata{11: f}}
	h := newHandlerWithUC(repo)
	r := setupFileRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/11/attach", strings.NewReader(`{"document_id": 42}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFileHandler_Attach_ValidationError_NoTargetID(t *testing.T) {
	f := entities.NewFileMetadata("doc.pdf", "key", "application/pdf", "h", 100, 7)
	f.ID = 11
	f.IsTemporary = true
	repo := &fakeFileMetaRepo{files: map[int64]*entities.FileMetadata{11: f}}
	h := newHandlerWithUC(repo)
	r := setupFileRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/11/attach", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ----- GetByID happy path -----

func TestFileHandler_GetByID_HappyPath(t *testing.T) {
	f := entities.NewFileMetadata("doc.pdf", "key", "application/pdf", "h", 100, 7)
	f.ID = 11
	repo := &fakeFileMetaRepo{files: map[int64]*entities.FileMetadata{11: f}}
	h := newHandlerWithUC(repo)
	r := setupFileRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/11", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
