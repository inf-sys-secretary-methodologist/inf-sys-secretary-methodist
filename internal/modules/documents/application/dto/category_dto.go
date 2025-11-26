// Package dto contains Data Transfer Objects for the documents module.
package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// CreateCategoryInput represents input for creating a new category
type CreateCategoryInput struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	ParentID    *int64  `json:"parent_id,omitempty"`
}

// UpdateCategoryInput represents input for updating a category
type UpdateCategoryInput struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	ParentID    *int64  `json:"parent_id,omitempty"`
}

// CategoryOutput represents output for a single category
type CategoryOutput struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Description   *string   `json:"description,omitempty"`
	ParentID      *int64    `json:"parent_id,omitempty"`
	ParentName    *string   `json:"parent_name,omitempty"`
	DocumentCount int64     `json:"document_count"`
	HasChildren   bool      `json:"has_children"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CategoryTreeOutput represents a category tree node with children
type CategoryTreeOutput struct {
	ID            int64                 `json:"id"`
	Name          string                `json:"name"`
	Description   *string               `json:"description,omitempty"`
	ParentID      *int64                `json:"parent_id,omitempty"`
	Children      []*CategoryTreeOutput `json:"children,omitempty"`
	DocumentCount int64                 `json:"document_count"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

// CategoryBreadcrumb represents a breadcrumb item for category navigation
type CategoryBreadcrumb struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// CategoryWithBreadcrumbOutput represents a category with its ancestor path
type CategoryWithBreadcrumbOutput struct {
	Category    *CategoryOutput       `json:"category"`
	Breadcrumbs []*CategoryBreadcrumb `json:"breadcrumbs"`
}

// CategoryFromEntity converts entity to output DTO
func CategoryFromEntity(c *entities.DocumentCategory) *CategoryOutput {
	return &CategoryOutput{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		ParentID:    c.ParentID,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// CategoryTreeFromEntity converts tree node entity to output DTO
func CategoryTreeFromEntity(node *entities.CategoryTreeNode) *CategoryTreeOutput {
	output := &CategoryTreeOutput{
		ID:            node.ID,
		Name:          node.Name,
		Description:   node.Description,
		ParentID:      node.ParentID,
		DocumentCount: node.DocumentCount,
		CreatedAt:     node.CreatedAt,
		UpdatedAt:     node.UpdatedAt,
	}

	if len(node.Children) > 0 {
		output.Children = make([]*CategoryTreeOutput, 0, len(node.Children))
		for _, child := range node.Children {
			output.Children = append(output.Children, CategoryTreeFromEntity(child))
		}
	}

	return output
}
