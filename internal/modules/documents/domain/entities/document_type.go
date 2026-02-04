// Package entities contains domain entities for the documents module.
package entities

import (
	"time"
)

// DocumentType represents a type of document (memo, order, letter, etc.)
type DocumentType struct {
	ID                   int64               `json:"id"`
	Name                 string              `json:"name"`
	Code                 string              `json:"code"`
	Description          *string             `json:"description,omitempty"`
	TemplatePath         *string             `json:"template_path,omitempty"`
	TemplateContent      *string             `json:"template_content,omitempty"`      // Template with {{variable}} placeholders
	TemplateVariables    []TemplateVariable  `json:"template_variables,omitempty"`    // Variable definitions
	RequiresApproval     bool                `json:"requires_approval"`
	RequiresRegistration bool                `json:"requires_registration"`
	NumberingPattern     *string             `json:"numbering_pattern,omitempty"`
	RetentionPeriod      *int                `json:"retention_period,omitempty"` // months
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
}

// TemplateVariable defines a variable used in document templates
type TemplateVariable struct {
	Name         string   `json:"name"`                    // Variable name (used in template as {{name}})
	Label        string   `json:"label"`                   // Display label
	Type         string   `json:"type"`                    // string, text, date, number, select
	Required     bool     `json:"required"`                // Is variable required
	DefaultValue *string  `json:"default_value,omitempty"` // Default value
	Options      []string `json:"options,omitempty"`       // Options for select type
}

// DocumentCategory represents a category for grouping documents
type DocumentCategory struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	ParentID    *int64    `json:"parent_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
	ID                int64                  `json:"id"`
	DocumentID        int64                  `json:"document_id"`
	Version           int                    `json:"version"`
	Title             *string                `json:"title,omitempty"`
	Subject           *string                `json:"subject,omitempty"`
	Content           *string                `json:"content,omitempty"`
	Status            *string                `json:"status,omitempty"`
	FileName          *string                `json:"file_name,omitempty"`
	FilePath          *string                `json:"file_path,omitempty"`
	FileSize          *int64                 `json:"file_size,omitempty"`
	MimeType          *string                `json:"mime_type,omitempty"`
	StorageKey        *string                `json:"storage_key,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	ChangedBy         int64                  `json:"changed_by"`
	ChangedByName     *string                `json:"changed_by_name,omitempty"` // Populated via JOIN
	ChangeDescription *string                `json:"change_description,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

// DocumentVersionDiff represents a comparison between two document versions
type DocumentVersionDiff struct {
	ID            int64                  `json:"id"`
	DocumentID    int64                  `json:"document_id"`
	FromVersion   int                    `json:"from_version"`
	ToVersion     int                    `json:"to_version"`
	ChangedFields []string               `json:"changed_fields"`
	DiffData      map[string]interface{} `json:"diff_data,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
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
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Color     *string   `json:"color,omitempty"` // hex color
	CreatedAt time.Time `json:"created_at"`
}

// DocumentHistory represents an audit log entry for a document
type DocumentHistory struct {
	ID         int64                  `json:"id"`
	DocumentID int64                  `json:"document_id"`
	UserID     *int64                 `json:"user_id,omitempty"`
	Action     string                 `json:"action"`
	Details    map[string]interface{} `json:"details,omitempty"`
	IPAddress  *string                `json:"ip_address,omitempty"`
	UserAgent  *string                `json:"user_agent,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}
