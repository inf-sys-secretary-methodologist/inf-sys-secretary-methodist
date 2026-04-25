package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Handler-level unit tests for attachment endpoints. Cover authorization,
// URL parsing, and form parsing — the parts that don't require a usecase
// instance. End-to-end happy-path coverage lives in the usecase tests
// (TestAddAttachment_*) and Playwright e2e per project convention.

func TestUploadAttachment_Unauthorized(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	r := gin.New()
	r.POST("/announcements/:id/attachments", handler.UploadAttachment)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/announcements/1/attachments", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUploadAttachment_InvalidID(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	r := gin.New()
	r.POST("/announcements/:id/attachments", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.UploadAttachment(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/announcements/abc/attachments", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadAttachment_MissingFileInForm(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	r := gin.New()
	r.POST("/announcements/:id/attachments", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.UploadAttachment(c)
	})

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	_ = mw.WriteField("not-the-file", "x")
	_ = mw.Close()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/announcements/1/attachments", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteAttachment_Unauthorized(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	r := gin.New()
	r.DELETE("/announcements/:id/attachments/:attachmentID", handler.DeleteAttachment)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/announcements/1/attachments/2", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteAttachment_InvalidAttachmentID(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	r := gin.New()
	r.DELETE("/announcements/:id/attachments/:attachmentID", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.DeleteAttachment(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/announcements/1/attachments/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
