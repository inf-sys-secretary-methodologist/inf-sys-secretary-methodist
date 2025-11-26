// Package dto contains Data Transfer Objects for the documents module.
package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// CreateTagInput represents input for creating a new tag
type CreateTagInput struct {
	Name  string  `json:"name" validate:"required,min=1,max=100"`
	Color *string `json:"color,omitempty" validate:"omitempty,hexcolor"`
}

// UpdateTagInput represents input for updating a tag
type UpdateTagInput struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Color *string `json:"color,omitempty" validate:"omitempty,hexcolor"`
}

// TagOutput represents output for a single tag
type TagOutput struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Color      *string   `json:"color,omitempty"`
	UsageCount int64     `json:"usage_count"`
	CreatedAt  time.Time `json:"created_at"`
}

// SetDocumentTagsInput represents input for setting document tags
type SetDocumentTagsInput struct {
	TagIDs []int64 `json:"tag_ids" validate:"required"`
}

// AddTagToDocumentInput represents input for adding a tag to a document
type AddTagToDocumentInput struct {
	TagID int64 `json:"tag_id" validate:"required"`
}

// DocumentTagsOutput represents output for document tags
type DocumentTagsOutput struct {
	DocumentID int64        `json:"document_id"`
	Tags       []*TagOutput `json:"tags"`
}

// TagWithDocumentsOutput represents a tag with its associated document IDs
type TagWithDocumentsOutput struct {
	Tag         *TagOutput `json:"tag"`
	DocumentIDs []int64    `json:"document_ids"`
	Total       int64      `json:"total"`
	Page        int        `json:"page"`
	PageSize    int        `json:"page_size"`
}

// TagFromEntity converts entity to output DTO
func TagFromEntity(t *entities.DocumentTag) *TagOutput {
	return &TagOutput{
		ID:        t.ID,
		Name:      t.Name,
		Color:     t.Color,
		CreatedAt: t.CreatedAt,
	}
}

// TagsFromEntities converts multiple entities to output DTOs
func TagsFromEntities(tags []*entities.DocumentTag) []*TagOutput {
	outputs := make([]*TagOutput, 0, len(tags))
	for _, t := range tags {
		outputs = append(outputs, TagFromEntity(t))
	}
	return outputs
}
