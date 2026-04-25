package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/application/usecases"
)

// UploadAttachment handles POST /announcements/:id/attachments.
//
// Accepts a multipart/form-data request with a single "file" field, persists
// it via the usecase, and returns the new attachment metadata.
func (h *AnnouncementHandler) UploadAttachment(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	announcementID, ok := h.getIDParam(c, "id")
	if !ok {
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file field is required"})
		return
	}
	defer func() { _ = file.Close() }()

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	att, err := h.useCase.AddAttachment(
		c.Request.Context(),
		announcementID,
		header.Filename,
		file,
		header.Size,
		mimeType,
		userID,
	)
	if err != nil {
		h.handleAttachmentError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":              att.ID,
		"announcement_id": att.AnnouncementID,
		"file_name":       att.FileName,
		"file_size":       att.FileSize,
		"mime_type":       att.MimeType,
		"created_at":      att.CreatedAt,
	})
}

// DeleteAttachment handles DELETE /announcements/:id/attachments/:attachmentID.
func (h *AnnouncementHandler) DeleteAttachment(c *gin.Context) {
	if _, ok := h.getUserID(c); !ok {
		return
	}

	if _, ok := h.getIDParam(c, "id"); !ok {
		return
	}

	attachmentID, ok := h.getIDParam(c, "attachmentID")
	if !ok {
		return
	}

	if err := h.useCase.RemoveAttachment(c.Request.Context(), attachmentID); err != nil {
		h.handleAttachmentError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// handleAttachmentError maps usecase errors specific to attachments.
// Falls through to the generic handleError for shared errors.
func (h *AnnouncementHandler) handleAttachmentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecases.ErrAttachmentNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "attachment not found"})
	case errors.Is(err, usecases.ErrStorageNotConfigured):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "attachment storage is not configured"})
	default:
		h.handleError(c, err)
	}
}
