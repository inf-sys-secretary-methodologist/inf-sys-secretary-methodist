package handlers

// v0.153.11 Phase 6 #196 backfill — covers handleAttachmentError +
// the storage-not-configured/attachment-not-found branches of
// UploadAttachment/DeleteAttachment that the existing handler tests
// can't reach with NewAnnouncementHandler(nil).
//
// Approach: real *usecases.AnnouncementUseCase wired with minimal
// in-handler fakes (AnnouncementRepository, optional AttachmentStorage).
// Mirror к v0.153.10 fakes pattern from tasks handler integration tests.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
)

// ----- minimal AnnouncementRepository fake -----
//
// Returns zero-value (nil, nil) by default. Per-branch tests assign
// per-method errors / payloads.

type fakeAnnRepo struct {
	getByIDResult *entities.Announcement
	getByIDErr    error
	getAttResult  *entities.AnnouncementAttachment
	getAttErr     error
}

func (r *fakeAnnRepo) Create(_ context.Context, _ *entities.Announcement) error { return nil }
func (r *fakeAnnRepo) Save(_ context.Context, _ *entities.Announcement) error   { return nil }
func (r *fakeAnnRepo) GetByID(_ context.Context, _ int64) (*entities.Announcement, error) {
	return r.getByIDResult, r.getByIDErr
}
func (r *fakeAnnRepo) Delete(_ context.Context, _ int64) error { return nil }
func (r *fakeAnnRepo) List(_ context.Context, _ usecases.AnnouncementFilter, _, _ int) ([]*entities.Announcement, error) {
	return nil, nil
}
func (r *fakeAnnRepo) Count(_ context.Context, _ usecases.AnnouncementFilter) (int64, error) {
	return 0, nil
}
func (r *fakeAnnRepo) GetByAuthor(_ context.Context, _ int64, _, _ int) ([]*entities.Announcement, error) {
	return nil, nil
}
func (r *fakeAnnRepo) GetPublished(_ context.Context, _ domain.TargetAudience, _, _ int) ([]*entities.Announcement, error) {
	return nil, nil
}
func (r *fakeAnnRepo) GetPinned(_ context.Context, _ int) ([]*entities.Announcement, error) {
	return nil, nil
}
func (r *fakeAnnRepo) GetRecent(_ context.Context, _ int) ([]*entities.Announcement, error) {
	return nil, nil
}
func (r *fakeAnnRepo) IncrementViewCount(_ context.Context, _ int64) error { return nil }
func (r *fakeAnnRepo) AddAttachment(_ context.Context, _ *entities.AnnouncementAttachment) error {
	return nil
}
func (r *fakeAnnRepo) RemoveAttachment(_ context.Context, _ int64) error { return nil }
func (r *fakeAnnRepo) GetAttachments(_ context.Context, _ int64) ([]*entities.AnnouncementAttachment, error) {
	return nil, nil
}
func (r *fakeAnnRepo) GetAttachmentByID(_ context.Context, _ int64) (*entities.AnnouncementAttachment, error) {
	return r.getAttResult, r.getAttErr
}

// ----- minimal AttachmentStorage fake -----

type fakeAttStorage struct{}

func (s *fakeAttStorage) Upload(_ context.Context, _ string, _ io.Reader, _ int64, _ string) (*storage.FileInfo, error) {
	return &storage.FileInfo{}, nil
}
func (s *fakeAttStorage) Delete(_ context.Context, _ string) error { return nil }
func (s *fakeAttStorage) GetPresignedURL(_ context.Context, _ string, _ time.Duration) (string, error) {
	return "", nil
}

// ----- tests -----

func newUploadRouter(h *AnnouncementHandler, userID int64) *gin.Engine {
	r := gin.New()
	r.POST("/announcements/:id/attachments", func(c *gin.Context) {
		c.Set("user_id", userID)
		h.UploadAttachment(c)
	})
	return r
}

func newDeleteRouter(h *AnnouncementHandler, userID int64) *gin.Engine {
	r := gin.New()
	r.DELETE("/announcements/:id/attachments/:attachmentID", func(c *gin.Context) {
		c.Set("user_id", userID)
		h.DeleteAttachment(c)
	})
	return r
}

func newMultipartFile(t *testing.T, fieldName, fileName, content string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	// v0.163.0 ADR-5 (#303): explicit text/plain Content-Type per-part —
	// the legacy CreateFormFile defaults к application/octet-stream which
	// the validator now rejects (octet-stream loophole closed). Tests
	// using .txt fixtures expect to traverse the validator into deeper
	// error paths, so we mirror what browsers actually send.
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name=%q; filename=%q`, fieldName, fileName))
	h.Set("Content-Type", "text/plain")
	part, err := mw.CreatePart(h)
	if err != nil {
		t.Fatalf("create part: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write part: %v", err)
	}
	_ = mw.Close()
	return body, mw.FormDataContentType()
}

func TestUploadAttachment_StorageNotConfigured(t *testing.T) {
	uc := usecases.NewAnnouncementUseCase(&fakeAnnRepo{}, nil, nil, nil)
	h := NewAnnouncementHandler(uc)
	r := newUploadRouter(h, 42)

	body, contentType := newMultipartFile(t, "file", "note.txt", "hello")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/announcements/1/attachments", body)
	req.Header.Set("Content-Type", contentType)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "attachment storage is not configured")
}

func TestDeleteAttachment_StorageNotConfigured(t *testing.T) {
	uc := usecases.NewAnnouncementUseCase(&fakeAnnRepo{}, nil, nil, nil)
	h := NewAnnouncementHandler(uc)
	r := newDeleteRouter(h, 42)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/announcements/1/attachments/2", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "attachment storage is not configured")
}

func TestDeleteAttachment_AttachmentNotFound(t *testing.T) {
	uc := usecases.NewAnnouncementUseCase(&fakeAnnRepo{getAttResult: nil, getAttErr: nil}, nil, nil, nil)
	uc.SetAttachmentStorage(&fakeAttStorage{})
	h := NewAnnouncementHandler(uc)
	r := newDeleteRouter(h, 42)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/announcements/1/attachments/999", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "attachment not found")
}

func TestUploadAttachment_AnnouncementNotFound(t *testing.T) {
	// storage configured; repo.GetByID returns (nil, nil) → ErrAnnouncementNotFound
	// → handleAttachmentError default branch → handleError → 404
	uc := usecases.NewAnnouncementUseCase(&fakeAnnRepo{getByIDResult: nil, getByIDErr: nil}, nil, nil, nil)
	uc.SetAttachmentStorage(&fakeAttStorage{})
	h := NewAnnouncementHandler(uc)
	r := newUploadRouter(h, 42)

	body, contentType := newMultipartFile(t, "file", "note.txt", "hello")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/announcements/1/attachments", body)
	req.Header.Set("Content-Type", contentType)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ----- list-family handler happy paths -----
//
// fakeAnnRepo returns nil + nil for List/GetPublished/GetPinned/GetRecent
// → handlers return 200 + empty announcements list.

func TestListHandler_HappyPath_EmptyResult(t *testing.T) {
	uc := usecases.NewAnnouncementUseCase(&fakeAnnRepo{}, nil, nil, nil)
	h := NewAnnouncementHandler(uc)

	r := gin.New()
	r.GET("/announcements", h.List)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/announcements?author_id=7&status=draft&priority=high&target_audience=all&is_pinned=true&search=test&tags=a&tags=b&limit=10&offset=0", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListHandler_NoFilters(t *testing.T) {
	uc := usecases.NewAnnouncementUseCase(&fakeAnnRepo{}, nil, nil, nil)
	h := NewAnnouncementHandler(uc)

	r := gin.New()
	r.GET("/announcements", h.List)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/announcements", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetPublishedHandler_HappyPath(t *testing.T) {
	uc := usecases.NewAnnouncementUseCase(&fakeAnnRepo{}, nil, nil, nil)
	h := NewAnnouncementHandler(uc)

	r := gin.New()
	r.GET("/announcements/published", h.GetPublished)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/announcements/published?audience=all&limit=10&offset=0", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetPublishedHandler_InvalidAudienceFallsToAll(t *testing.T) {
	uc := usecases.NewAnnouncementUseCase(&fakeAnnRepo{}, nil, nil, nil)
	h := NewAnnouncementHandler(uc)

	r := gin.New()
	r.GET("/announcements/published", h.GetPublished)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/announcements/published?audience=invalid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetPinnedHandler_HappyPath(t *testing.T) {
	uc := usecases.NewAnnouncementUseCase(&fakeAnnRepo{}, nil, nil, nil)
	h := NewAnnouncementHandler(uc)

	r := gin.New()
	r.GET("/announcements/pinned", h.GetPinned)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/announcements/pinned?limit=3", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetRecentHandler_HappyPath(t *testing.T) {
	uc := usecases.NewAnnouncementUseCase(&fakeAnnRepo{}, nil, nil, nil)
	h := NewAnnouncementHandler(uc)

	r := gin.New()
	r.GET("/announcements/recent", h.GetRecent)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/announcements/recent?limit=5", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUploadAttachment_RepoLookupError(t *testing.T) {
	// storage configured; repo.GetByID returns generic error
	// → AddAttachment returns wrapped error → handleAttachmentError default
	// → handleError default → 500
	uc := usecases.NewAnnouncementUseCase(
		&fakeAnnRepo{getByIDErr: errors.New("db error")},
		nil, nil, nil,
	)
	uc.SetAttachmentStorage(&fakeAttStorage{})
	h := NewAnnouncementHandler(uc)
	r := newUploadRouter(h, 42)

	body, contentType := newMultipartFile(t, "file", "note.txt", "hello")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/announcements/1/attachments", body)
	req.Header.Set("Content-Type", contentType)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
