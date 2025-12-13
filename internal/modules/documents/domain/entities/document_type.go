// Package entities contains domain entities for the documents module.
package entities

import (
	"time"
)

// DocumentType represents a type of document (memo, order, letter, etc.)
// Aligned with migrations/003_create_documents_schema.up.sql
type DocumentType struct {
	ID                   int64     `db:"id" json:"id"`
	Name                 string    `db:"name" json:"name"`
	Code                 string    `db:"code" json:"code"`
	Description          *string   `db:"description" json:"description,omitempty"`
	TemplatePath         *string   `db:"template_path" json:"template_path,omitempty"`
	RequiresApproval     bool      `db:"requires_approval" json:"requires_approval"`
	RequiresRegistration bool      `db:"requires_registration" json:"requires_registration"`
	NumberingPattern     *string   `db:"numbering_pattern" json:"numbering_pattern,omitempty"`
	RetentionPeriod      *int      `db:"retention_period" json:"retention_period,omitempty"` // months
	CreatedAt            time.Time `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time `db:"updated_at" json:"updated_at"`
}

// DocumentCategory represents a category for grouping documents
type DocumentCategory struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description,omitempty"`
	ParentID    *int64    `db:"parent_id" json:"parent_id,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// CategoryTreeNode represents a category with its children for tree structure
type CategoryTreeNode struct {
	ID            int64               `json:"id"`
	Name          string              `json:"name"`
	Description   *string             `json:"description,omitempty"`
	ParentID      *int64              `json:"parent_id,omitempty"`
	Children      []*CategoryTreeNode `json:"children,omitempty"`
	DocumentCount int64               `json:"document_count"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// DocumentVersion represents a version of a document with full snapshot
type DocumentVersion struct {
	ID                int64                  `db:"id" json:"id"`
	DocumentID        int64                  `db:"document_id" json:"document_id"`
	Version           int                    `db:"version" json:"version"`
	Title             *string                `db:"title" json:"title,omitempty"`
	Subject           *string                `db:"subject" json:"subject,omitempty"`
	Content           *string                `db:"content" json:"content,omitempty"`
	Status            *string                `db:"status" json:"status,omitempty"`
	FileName          *string                `db:"file_name" json:"file_name,omitempty"`
	FilePath          *string                `db:"file_path" json:"file_path,omitempty"`
	FileSize          *int64                 `db:"file_size" json:"file_size,omitempty"`
	MimeType          *string                `db:"mime_type" json:"mime_type,omitempty"`
	StorageKey        *string                `db:"storage_key" json:"storage_key,omitempty"`
	Metadata          map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
	ChangedBy         int64                  `db:"changed_by" json:"changed_by"`
	ChangedByName     *string                `db:"-" json:"changed_by_name,omitempty"` // Populated via JOIN
	ChangeDescription *string                `db:"change_description" json:"change_description,omitempty"`
	CreatedAt         time.Time              `db:"created_at" json:"created_at"`
}

// DocumentVersionDiff represents a comparison between two document versions
type DocumentVersionDiff struct {
	ID            int64                  `db:"id" json:"id"`
	DocumentID    int64                  `db:"document_id" json:"document_id"`
	FromVersion   int                    `db:"from_version" json:"from_version"`
	ToVersion     int                    `db:"to_version" json:"to_version"`
	ChangedFields []string               `db:"changed_fields" json:"changed_fields"`
	DiffData      map[string]interface{} `db:"diff_data" json:"diff_data,omitempty"`
	CreatedAt     time.Time              `db:"created_at" json:"created_at"`
}

// FieldDiff represents a single field difference between versions
type FieldDiff struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
}

// NewDocumentVersion creates a new document version from a document
func NewDocumentVersion(doc *Document, changedBy int64, description string) *DocumentVersion {
	status := string(doc.Status)
	return &DocumentVersion{
		DocumentID:        doc.ID,
		Version:           doc.Version,
		Title:             &doc.Title,
		Subject:           doc.Subject,
		Content:           doc.Content,
		Status:            &status,
		FileName:          doc.FileName,
		FilePath:          doc.FilePath,
		FileSize:          doc.FileSize,
		MimeType:          doc.MimeType,
		StorageKey:        doc.FilePath,
		Metadata:          doc.Metadata,
		ChangedBy:         changedBy,
		ChangeDescription: &description,
		CreatedAt:         time.Now(),
	}
}

// DocumentTag represents a tag for documents
type DocumentTag struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Color     *string   `db:"color" json:"color,omitempty"` // hex color
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// DocumentHistory represents an audit log entry for a document
type DocumentHistory struct {
	ID         int64                  `db:"id" json:"id"`
	DocumentID int64                  `db:"document_id" json:"document_id"`
	UserID     *int64                 `db:"user_id" json:"user_id,omitempty"`
	Action     string                 `db:"action" json:"action"`
	Details    map[string]interface{} `db:"details" json:"details,omitempty"`
	IPAddress  *string                `db:"ip_address" json:"ip_address,omitempty"`
	UserAgent  *string                `db:"user_agent" json:"user_agent,omitempty"`
	CreatedAt  time.Time              `db:"created_at" json:"created_at"`
}
