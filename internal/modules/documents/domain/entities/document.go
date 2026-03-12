// Package entities contains domain entities for the documents module.
package entities

import (
	"time"
)

// DocumentStatus represents the status of a document in workflow
type DocumentStatus string

// DocumentStatus values.
const (
	DocumentStatusDraft      DocumentStatus = "draft"
	DocumentStatusRegistered DocumentStatus = "registered"
	DocumentStatusRouting    DocumentStatus = "routing"
	DocumentStatusApproval   DocumentStatus = "approval"
	DocumentStatusApproved   DocumentStatus = "approved"
	DocumentStatusRejected   DocumentStatus = "rejected"
	DocumentStatusExecution  DocumentStatus = "execution"
	DocumentStatusExecuted   DocumentStatus = "executed"
	DocumentStatusArchived   DocumentStatus = "archived"
)

// DocumentImportance represents the importance level of a document
type DocumentImportance string

// DocumentImportance values.
const (
	ImportanceLow    DocumentImportance = "low"
	ImportanceNormal DocumentImportance = "normal"
	ImportanceHigh   DocumentImportance = "high"
	ImportanceUrgent DocumentImportance = "urgent"
)

// Document represents a document entity in the documents domain
type Document struct {
	ID             int64  `json:"id"`
	DocumentTypeID int64  `json:"document_type_id"`
	CategoryID     *int64 `json:"category_id,omitempty"`

	// Registration data
	RegistrationNumber *string    `json:"registration_number,omitempty"`
	RegistrationDate   *time.Time `json:"registration_date,omitempty"`

	// Main information
	Title   string  `json:"title"`
	Subject *string `json:"subject,omitempty"`
	Content *string `json:"content,omitempty"`

	// Author details
	AuthorID         int64   `json:"author_id"`
	AuthorName       *string `json:"author_name,omitempty"` // Populated via JOIN
	AuthorDepartment *string `json:"author_department,omitempty"`
	AuthorPosition   *string `json:"author_position,omitempty"`

	// Recipient details
	RecipientID         *int64  `json:"recipient_id,omitempty"`
	RecipientName       *string `json:"recipient_name,omitempty"` // Populated via JOIN
	RecipientDepartment *string `json:"recipient_department,omitempty"`
	RecipientPosition   *string `json:"recipient_position,omitempty"`
	RecipientExternal   *string `json:"recipient_external,omitempty"`

	// Status and workflow
	Status DocumentStatus `json:"status"`

	// File information
	FileName *string `json:"file_name,omitempty"`
	FilePath *string `json:"file_path,omitempty"`
	FileSize *int64  `json:"file_size,omitempty"`
	MimeType *string `json:"mime_type,omitempty"`

	// Versioning
	Version          int    `json:"version"`
	ParentDocumentID *int64 `json:"parent_document_id,omitempty"`

	// Deadlines
	Deadline      *time.Time `json:"deadline,omitempty"`
	ExecutionDate *time.Time `json:"execution_date,omitempty"`

	// Metadata
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	IsPublic   bool                   `json:"is_public"`
	Importance DocumentImportance     `json:"importance"`

	// Audit
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// NewDocument creates a new document with default values
func NewDocument(title string, documentTypeID, authorID int64) *Document {
	now := time.Now()
	return &Document{
		DocumentTypeID: documentTypeID,
		Title:          title,
		AuthorID:       authorID,
		Status:         DocumentStatusDraft,
		Version:        1,
		IsPublic:       false,
		Importance:     ImportanceNormal,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// SetFile sets file information for the document
func (d *Document) SetFile(fileName, filePath, mimeType string, fileSize int64) {
	d.FileName = &fileName
	d.FilePath = &filePath
	d.MimeType = &mimeType
	d.FileSize = &fileSize
	d.UpdatedAt = time.Now()
}

// ClearFile removes file information from the document
func (d *Document) ClearFile() {
	d.FileName = nil
	d.FilePath = nil
	d.MimeType = nil
	d.FileSize = nil
	d.UpdatedAt = time.Now()
}

// Register registers the document with a number and date
func (d *Document) Register(registrationNumber string) {
	now := time.Now()
	d.RegistrationNumber = &registrationNumber
	d.RegistrationDate = &now
	d.Status = DocumentStatusRegistered
	d.UpdatedAt = now
}

// IsDraft checks if document is in draft status
func (d *Document) IsDraft() bool {
	return d.Status == DocumentStatusDraft
}

// IsDeleted checks if document is soft-deleted
func (d *Document) IsDeleted() bool {
	return d.DeletedAt != nil
}

// SoftDelete marks the document as deleted
func (d *Document) SoftDelete() {
	now := time.Now()
	d.DeletedAt = &now
	d.UpdatedAt = now
}

// Restore restores a soft-deleted document
func (d *Document) Restore() {
	d.DeletedAt = nil
	d.UpdatedAt = time.Now()
}

// HasFile checks if document has an attached file
func (d *Document) HasFile() bool {
	return d.FilePath != nil && *d.FilePath != ""
}
