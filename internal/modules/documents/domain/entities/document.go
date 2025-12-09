// Package entities contains domain entities for the documents module.
package entities

import (
	"time"
)

// DocumentStatus represents the status of a document in workflow
type DocumentStatus string

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

const (
	ImportanceLow    DocumentImportance = "low"
	ImportanceNormal DocumentImportance = "normal"
	ImportanceHigh   DocumentImportance = "high"
	ImportanceUrgent DocumentImportance = "urgent"
)

// Document represents a document entity in the documents domain
// Aligned with migrations/003_create_documents_schema.up.sql
type Document struct {
	ID             int64  `db:"id" json:"id"`
	DocumentTypeID int64  `db:"document_type_id" json:"document_type_id"`
	CategoryID     *int64 `db:"category_id" json:"category_id,omitempty"`

	// Registration data
	RegistrationNumber *string    `db:"registration_number" json:"registration_number,omitempty"`
	RegistrationDate   *time.Time `db:"registration_date" json:"registration_date,omitempty"`

	// Main information
	Title   string  `db:"title" json:"title"`
	Subject *string `db:"subject" json:"subject,omitempty"`
	Content *string `db:"content" json:"content,omitempty"`

	// Author details
	AuthorID         int64   `db:"author_id" json:"author_id"`
	AuthorName       *string `db:"-" json:"author_name,omitempty"` // Populated via JOIN, not stored in documents table
	AuthorDepartment *string `db:"author_department" json:"author_department,omitempty"`
	AuthorPosition   *string `db:"author_position" json:"author_position,omitempty"`

	// Recipient details
	RecipientID         *int64  `db:"recipient_id" json:"recipient_id,omitempty"`
	RecipientName       *string `db:"-" json:"recipient_name,omitempty"` // Populated via JOIN, not stored in documents table
	RecipientDepartment *string `db:"recipient_department" json:"recipient_department,omitempty"`
	RecipientPosition   *string `db:"recipient_position" json:"recipient_position,omitempty"`
	RecipientExternal   *string `db:"recipient_external" json:"recipient_external,omitempty"`

	// Status and workflow
	Status DocumentStatus `db:"status" json:"status"`

	// File information
	FileName *string `db:"file_name" json:"file_name,omitempty"`
	FilePath *string `db:"file_path" json:"file_path,omitempty"`
	FileSize *int64  `db:"file_size" json:"file_size,omitempty"`
	MimeType *string `db:"mime_type" json:"mime_type,omitempty"`

	// Versioning
	Version          int    `db:"version" json:"version"`
	ParentDocumentID *int64 `db:"parent_document_id" json:"parent_document_id,omitempty"`

	// Deadlines
	Deadline      *time.Time `db:"deadline" json:"deadline,omitempty"`
	ExecutionDate *time.Time `db:"execution_date" json:"execution_date,omitempty"`

	// Metadata
	Metadata   map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
	IsPublic   bool                   `db:"is_public" json:"is_public"`
	Importance DocumentImportance     `db:"importance" json:"importance"`

	// Audit
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
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
