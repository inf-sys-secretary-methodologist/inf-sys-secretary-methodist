package entities

import "testing"

func TestNewDocument(t *testing.T) {
	title := "Test Document"
	documentTypeID := int64(1)
	authorID := int64(42)

	doc := NewDocument(title, documentTypeID, authorID)

	if doc.Title != title {
		t.Errorf("expected title %q, got %q", title, doc.Title)
	}
	if doc.DocumentTypeID != documentTypeID {
		t.Errorf("expected document type ID %d, got %d", documentTypeID, doc.DocumentTypeID)
	}
	if doc.AuthorID != authorID {
		t.Errorf("expected author ID %d, got %d", authorID, doc.AuthorID)
	}
	if doc.Status != DocumentStatusDraft {
		t.Errorf("expected status %q, got %q", DocumentStatusDraft, doc.Status)
	}
	if doc.Version != 1 {
		t.Errorf("expected version 1, got %d", doc.Version)
	}
	if doc.IsPublic {
		t.Error("expected IsPublic to be false")
	}
	if doc.Importance != ImportanceNormal {
		t.Errorf("expected importance %q, got %q", ImportanceNormal, doc.Importance)
	}
	if doc.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestDocument_SetFile(t *testing.T) {
	doc := NewDocument("Test", 1, 1)
	fileName := "document.pdf"
	filePath := "/files/document.pdf"
	mimeType := "application/pdf"
	fileSize := int64(1024)

	doc.SetFile(fileName, filePath, mimeType, fileSize)

	if doc.FileName == nil || *doc.FileName != fileName {
		t.Errorf("expected file name %q, got %v", fileName, doc.FileName)
	}
	if doc.FilePath == nil || *doc.FilePath != filePath {
		t.Errorf("expected file path %q, got %v", filePath, doc.FilePath)
	}
	if doc.MimeType == nil || *doc.MimeType != mimeType {
		t.Errorf("expected mime type %q, got %v", mimeType, doc.MimeType)
	}
	if doc.FileSize == nil || *doc.FileSize != fileSize {
		t.Errorf("expected file size %d, got %v", fileSize, doc.FileSize)
	}
}

func TestDocument_ClearFile(t *testing.T) {
	doc := NewDocument("Test", 1, 1)
	doc.SetFile("doc.pdf", "/files/doc.pdf", "application/pdf", 1024)

	doc.ClearFile()

	if doc.FileName != nil {
		t.Error("expected file name to be nil")
	}
	if doc.FilePath != nil {
		t.Error("expected file path to be nil")
	}
	if doc.MimeType != nil {
		t.Error("expected mime type to be nil")
	}
	if doc.FileSize != nil {
		t.Error("expected file size to be nil")
	}
}

func TestDocument_Register(t *testing.T) {
	doc := NewDocument("Test", 1, 1)
	regNumber := "REG-2024-001"

	doc.Register(regNumber)

	if doc.RegistrationNumber == nil || *doc.RegistrationNumber != regNumber {
		t.Errorf("expected registration number %q, got %v", regNumber, doc.RegistrationNumber)
	}
	if doc.RegistrationDate == nil {
		t.Error("expected registration date to be set")
	}
	if doc.Status != DocumentStatusRegistered {
		t.Errorf("expected status %q, got %q", DocumentStatusRegistered, doc.Status)
	}
}

func TestDocument_IsDraft(t *testing.T) {
	doc := NewDocument("Test", 1, 1)

	if !doc.IsDraft() {
		t.Error("expected new document to be draft")
	}

	doc.Register("REG-001")

	if doc.IsDraft() {
		t.Error("expected registered document to not be draft")
	}
}

func TestDocument_IsDeleted(t *testing.T) {
	doc := NewDocument("Test", 1, 1)

	if doc.IsDeleted() {
		t.Error("expected new document to not be deleted")
	}

	doc.SoftDelete()

	if !doc.IsDeleted() {
		t.Error("expected soft-deleted document to be deleted")
	}
}

func TestDocument_SoftDelete(t *testing.T) {
	doc := NewDocument("Test", 1, 1)

	doc.SoftDelete()

	if doc.DeletedAt == nil {
		t.Error("expected DeletedAt to be set")
	}
}

func TestDocument_Restore(t *testing.T) {
	doc := NewDocument("Test", 1, 1)
	doc.SoftDelete()

	doc.Restore()

	if doc.DeletedAt != nil {
		t.Error("expected DeletedAt to be nil after restore")
	}
}

func TestDocument_HasFile(t *testing.T) {
	doc := NewDocument("Test", 1, 1)

	if doc.HasFile() {
		t.Error("expected new document to not have file")
	}

	doc.SetFile("doc.pdf", "/files/doc.pdf", "application/pdf", 1024)

	if !doc.HasFile() {
		t.Error("expected document with file to return true")
	}
}

func TestDocument_HasFile_EmptyPath(t *testing.T) {
	doc := NewDocument("Test", 1, 1)
	emptyPath := ""
	doc.FilePath = &emptyPath

	if doc.HasFile() {
		t.Error("expected document with empty path to not have file")
	}
}

func TestDocumentStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   DocumentStatus
		expected string
	}{
		{"draft", DocumentStatusDraft, "draft"},
		{"registered", DocumentStatusRegistered, "registered"},
		{"routing", DocumentStatusRouting, "routing"},
		{"approval", DocumentStatusApproval, "approval"},
		{"approved", DocumentStatusApproved, "approved"},
		{"rejected", DocumentStatusRejected, "rejected"},
		{"execution", DocumentStatusExecution, "execution"},
		{"executed", DocumentStatusExecuted, "executed"},
		{"archived", DocumentStatusArchived, "archived"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.status)
			}
		})
	}
}

func TestDocumentImportanceConstants(t *testing.T) {
	tests := []struct {
		name       string
		importance DocumentImportance
		expected   string
	}{
		{"low", ImportanceLow, "low"},
		{"normal", ImportanceNormal, "normal"},
		{"high", ImportanceHigh, "high"},
		{"urgent", ImportanceUrgent, "urgent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.importance) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.importance)
			}
		})
	}
}
