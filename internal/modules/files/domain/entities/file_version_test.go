package entities

import "testing"

func TestNewFileVersion(t *testing.T) {
	fileMetadataID := int64(1)
	versionNumber := 2
	storageKey := "files/v2/doc.pdf"
	checksum := "sha256:xyz789"
	comment := "Updated document"
	size := int64(2048)
	createdBy := int64(42)

	fv := NewFileVersion(fileMetadataID, versionNumber, storageKey, checksum, comment, size, createdBy)

	if fv.FileMetadataID != fileMetadataID {
		t.Errorf("expected file metadata ID %d, got %d", fileMetadataID, fv.FileMetadataID)
	}
	if fv.VersionNumber != versionNumber {
		t.Errorf("expected version number %d, got %d", versionNumber, fv.VersionNumber)
	}
	if fv.StorageKey != storageKey {
		t.Errorf("expected storage key %q, got %q", storageKey, fv.StorageKey)
	}
	if fv.Checksum != checksum {
		t.Errorf("expected checksum %q, got %q", checksum, fv.Checksum)
	}
	if fv.Comment != comment {
		t.Errorf("expected comment %q, got %q", comment, fv.Comment)
	}
	if fv.Size != size {
		t.Errorf("expected size %d, got %d", size, fv.Size)
	}
	if fv.CreatedBy != createdBy {
		t.Errorf("expected created by %d, got %d", createdBy, fv.CreatedBy)
	}
	if fv.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestNewFileVersion_EmptyComment(t *testing.T) {
	fv := NewFileVersion(1, 1, "key", "hash", "", 1024, 1)

	if fv.Comment != "" {
		t.Errorf("expected empty comment, got %q", fv.Comment)
	}
}
