// Package dto contains Data Transfer Objects for the documents module.
package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// DocumentVersionOutput represents output for a document version
type DocumentVersionOutput struct {
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
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	ChangedBy         int64                  `json:"changed_by"`
	ChangedByName     *string                `json:"changed_by_name,omitempty"`
	ChangeDescription *string                `json:"change_description,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

// DocumentVersionListOutput represents output for listing document versions
type DocumentVersionListOutput struct {
	Versions     []*DocumentVersionOutput `json:"versions"`
	Total        int64                    `json:"total"`
	DocumentID   int64                    `json:"document_id"`
	LatestVersion int                     `json:"latest_version"`
}

// CreateVersionInput represents input for creating a new version manually
type CreateVersionInput struct {
	DocumentID        int64   `json:"-"`
	ChangeDescription *string `json:"change_description,omitempty" validate:"omitempty,max=500"`
}

// RestoreVersionInput represents input for restoring a document to a previous version
type RestoreVersionInput struct {
	DocumentID int64 `json:"-"`
	Version    int   `json:"version" validate:"required,min=1"`
}

// CompareVersionsInput represents input for comparing two versions
type CompareVersionsInput struct {
	DocumentID  int64 `json:"-"`
	FromVersion int   `json:"from_version" validate:"required,min=1"`
	ToVersion   int   `json:"to_version" validate:"required,min=1"`
}

// VersionDiffOutput represents output for version comparison
type VersionDiffOutput struct {
	DocumentID    int64                  `json:"document_id"`
	FromVersion   int                    `json:"from_version"`
	ToVersion     int                    `json:"to_version"`
	ChangedFields []string               `json:"changed_fields"`
	DiffData      map[string]interface{} `json:"diff_data,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

// VersionFileDownloadOutput represents output for downloading a version's file
type VersionFileDownloadOutput struct {
	FileName    string `json:"file_name"`
	FilePath    string `json:"file_path"`
	FileSize    int64  `json:"file_size"`
	MimeType    string `json:"mime_type"`
	DownloadURL string `json:"download_url,omitempty"`
}

// ToDocumentVersionOutput converts entity to DTO
func ToDocumentVersionOutput(v *entities.DocumentVersion) *DocumentVersionOutput {
	if v == nil {
		return nil
	}
	return &DocumentVersionOutput{
		ID:                v.ID,
		DocumentID:        v.DocumentID,
		Version:           v.Version,
		Title:             v.Title,
		Subject:           v.Subject,
		Content:           v.Content,
		Status:            v.Status,
		FileName:          v.FileName,
		FilePath:          v.FilePath,
		FileSize:          v.FileSize,
		MimeType:          v.MimeType,
		Metadata:          v.Metadata,
		ChangedBy:         v.ChangedBy,
		ChangedByName:     v.ChangedByName,
		ChangeDescription: v.ChangeDescription,
		CreatedAt:         v.CreatedAt,
	}
}

// ToDocumentVersionOutputList converts a list of entities to DTOs
func ToDocumentVersionOutputList(versions []*entities.DocumentVersion) []*DocumentVersionOutput {
	result := make([]*DocumentVersionOutput, len(versions))
	for i, v := range versions {
		result[i] = ToDocumentVersionOutput(v)
	}
	return result
}

// ToVersionDiffOutput converts entity to DTO
func ToVersionDiffOutput(d *entities.DocumentVersionDiff) *VersionDiffOutput {
	if d == nil {
		return nil
	}
	return &VersionDiffOutput{
		DocumentID:    d.DocumentID,
		FromVersion:   d.FromVersion,
		ToVersion:     d.ToVersion,
		ChangedFields: d.ChangedFields,
		DiffData:      d.DiffData,
		CreatedAt:     d.CreatedAt,
	}
}
