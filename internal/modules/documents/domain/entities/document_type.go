// Package entities contains domain entities for the documents module.
package entities

import (
	"time"
)

// DocumentType represents a type of document (memo, order, letter, etc.)
// Aligned with migrations/003_create_documents_schema.up.sql
type DocumentType struct {
	ID                   int64   `db:"id" json:"id"`
	Name                 string  `db:"name" json:"name"`
	Code                 string  `db:"code" json:"code"`
	Description          *string `db:"description" json:"description,omitempty"`
	TemplatePath         *string `db:"template_path" json:"template_path,omitempty"`
	RequiresApproval     bool    `db:"requires_approval" json:"requires_approval"`
	RequiresRegistration bool    `db:"requires_registration" json:"requires_registration"`
	NumberingPattern     *string `db:"numbering_pattern" json:"numbering_pattern,omitempty"`
	RetentionPeriod      *int    `db:"retention_period" json:"retention_period,omitempty"` // months
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

// DocumentVersion represents a version of a document
type DocumentVersion struct {
	ID                int64     `db:"id" json:"id"`
	DocumentID        int64     `db:"document_id" json:"document_id"`
	Version           int       `db:"version" json:"version"`
	Content           *string   `db:"content" json:"content,omitempty"`
	FileName          *string   `db:"file_name" json:"file_name,omitempty"`
	FilePath          *string   `db:"file_path" json:"file_path,omitempty"`
	FileSize          *int64    `db:"file_size" json:"file_size,omitempty"`
	ChangedBy         int64     `db:"changed_by" json:"changed_by"`
	ChangeDescription *string   `db:"change_description" json:"change_description,omitempty"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
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
