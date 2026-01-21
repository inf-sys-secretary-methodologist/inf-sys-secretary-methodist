package entities

import (
	"testing"
	"time"
)

func TestNewFileMetadata(t *testing.T) {
	originalName := "document.pdf"
	storageKey := "files/abc123.pdf"
	mimeType := "application/pdf"
	checksum := "sha256:abc123"
	size := int64(1024)
	uploadedBy := int64(42)

	fm := NewFileMetadata(originalName, storageKey, mimeType, checksum, size, uploadedBy)

	if fm.OriginalName != originalName {
		t.Errorf("expected original name %q, got %q", originalName, fm.OriginalName)
	}
	if fm.StorageKey != storageKey {
		t.Errorf("expected storage key %q, got %q", storageKey, fm.StorageKey)
	}
	if fm.MimeType != mimeType {
		t.Errorf("expected mime type %q, got %q", mimeType, fm.MimeType)
	}
	if fm.Checksum != checksum {
		t.Errorf("expected checksum %q, got %q", checksum, fm.Checksum)
	}
	if fm.Size != size {
		t.Errorf("expected size %d, got %d", size, fm.Size)
	}
	if fm.UploadedBy != uploadedBy {
		t.Errorf("expected uploaded by %d, got %d", uploadedBy, fm.UploadedBy)
	}
	if !fm.IsTemporary {
		t.Error("expected IsTemporary to be true")
	}
	if fm.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestFileMetadata_AttachToDocument(t *testing.T) {
	fm := NewFileMetadata("doc.pdf", "key", "application/pdf", "hash", 1024, 1)
	documentID := int64(42)

	fm.AttachToDocument(documentID)

	if fm.DocumentID == nil || *fm.DocumentID != documentID {
		t.Errorf("expected document ID %d, got %v", documentID, fm.DocumentID)
	}
	if fm.IsTemporary {
		t.Error("expected IsTemporary to be false after attach")
	}
	if fm.ExpiresAt != nil {
		t.Error("expected ExpiresAt to be nil after attach")
	}
}

func TestFileMetadata_AttachToTask(t *testing.T) {
	fm := NewFileMetadata("doc.pdf", "key", "application/pdf", "hash", 1024, 1)
	taskID := int64(55)

	fm.AttachToTask(taskID)

	if fm.TaskID == nil || *fm.TaskID != taskID {
		t.Errorf("expected task ID %d, got %v", taskID, fm.TaskID)
	}
	if fm.IsTemporary {
		t.Error("expected IsTemporary to be false after attach")
	}
}

func TestFileMetadata_AttachToAnnouncement(t *testing.T) {
	fm := NewFileMetadata("doc.pdf", "key", "application/pdf", "hash", 1024, 1)
	announcementID := int64(77)

	fm.AttachToAnnouncement(announcementID)

	if fm.AnnouncementID == nil || *fm.AnnouncementID != announcementID {
		t.Errorf("expected announcement ID %d, got %v", announcementID, fm.AnnouncementID)
	}
	if fm.IsTemporary {
		t.Error("expected IsTemporary to be false after attach")
	}
}

func TestFileMetadata_MarkAsDeleted(t *testing.T) {
	fm := NewFileMetadata("doc.pdf", "key", "application/pdf", "hash", 1024, 1)

	fm.MarkAsDeleted()

	if fm.DeletedAt == nil {
		t.Error("expected DeletedAt to be set")
	}
}

func TestFileMetadata_IsDeleted(t *testing.T) {
	fm := NewFileMetadata("doc.pdf", "key", "application/pdf", "hash", 1024, 1)

	if fm.IsDeleted() {
		t.Error("expected new file to not be deleted")
	}

	fm.MarkAsDeleted()

	if !fm.IsDeleted() {
		t.Error("expected marked file to be deleted")
	}
}

func TestFileMetadata_IsExpired(t *testing.T) {
	fm := NewFileMetadata("doc.pdf", "key", "application/pdf", "hash", 1024, 1)

	// New file is temporary but has no expiry
	if fm.IsExpired() {
		t.Error("expected file without expiry to not be expired")
	}

	// Set expiry in the past
	pastTime := time.Now().Add(-1 * time.Hour)
	fm.ExpiresAt = &pastTime

	if !fm.IsExpired() {
		t.Error("expected file with past expiry to be expired")
	}

	// Set expiry in the future
	futureTime := time.Now().Add(1 * time.Hour)
	fm.ExpiresAt = &futureTime

	if fm.IsExpired() {
		t.Error("expected file with future expiry to not be expired")
	}
}

func TestFileMetadata_IsExpired_NotTemporary(t *testing.T) {
	fm := NewFileMetadata("doc.pdf", "key", "application/pdf", "hash", 1024, 1)
	fm.AttachToDocument(1) // Makes it non-temporary

	pastTime := time.Now().Add(-1 * time.Hour)
	fm.ExpiresAt = &pastTime

	// Non-temporary files are never expired (expiry only applies to temporary files)
	if fm.IsExpired() {
		t.Error("expected non-temporary file to not be expired")
	}
}
