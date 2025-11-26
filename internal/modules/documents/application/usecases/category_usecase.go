// Package usecases contains business logic for the documents module.
package usecases

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
)

// CategoryUseCase handles business logic for document categories
type CategoryUseCase struct {
	categoryRepo repositories.DocumentCategoryRepository
}

// NewCategoryUseCase creates a new category use case
func NewCategoryUseCase(categoryRepo repositories.DocumentCategoryRepository) *CategoryUseCase {
	return &CategoryUseCase{
		categoryRepo: categoryRepo,
	}
}

// Create creates a new category
func (uc *CategoryUseCase) Create(ctx context.Context, input dto.CreateCategoryInput) (*dto.CategoryOutput, error) {
	// Validate parent exists if specified
	if input.ParentID != nil {
		parent, err := uc.categoryRepo.GetByID(ctx, *input.ParentID)
		if err != nil {
			return nil, fmt.Errorf("родительская категория не найдена")
		}
		if parent == nil {
			return nil, fmt.Errorf("родительская категория не найдена")
		}
	}

	category := &entities.DocumentCategory{
		Name:        input.Name,
		Description: input.Description,
		ParentID:    input.ParentID,
	}

	if err := uc.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	output := dto.CategoryFromEntity(category)

	// Get additional info
	hasChildren, _ := uc.categoryRepo.HasChildren(ctx, category.ID)
	docCount, _ := uc.categoryRepo.GetDocumentCount(ctx, category.ID, false)
	output.HasChildren = hasChildren
	output.DocumentCount = docCount

	return output, nil
}

// Update updates an existing category
func (uc *CategoryUseCase) Update(ctx context.Context, id int64, input dto.UpdateCategoryInput) (*dto.CategoryOutput, error) {
	category, err := uc.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("категория не найдена")
	}

	// Validate parent if specified
	if input.ParentID != nil {
		// Prevent circular reference
		if *input.ParentID == id {
			return nil, fmt.Errorf("категория не может быть родителем самой себя")
		}

		// Check if new parent is not a descendant
		ancestors, err := uc.categoryRepo.GetAncestors(ctx, *input.ParentID)
		if err == nil {
			for _, ancestor := range ancestors {
				if ancestor.ID == id {
					return nil, fmt.Errorf("нельзя переместить категорию в её подкатегорию")
				}
			}
		}

		// Verify parent exists
		parent, err := uc.categoryRepo.GetByID(ctx, *input.ParentID)
		if err != nil || parent == nil {
			return nil, fmt.Errorf("родительская категория не найдена")
		}

		category.ParentID = input.ParentID
	}

	if input.Name != nil {
		category.Name = *input.Name
	}
	if input.Description != nil {
		category.Description = input.Description
	}

	if err := uc.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	output := dto.CategoryFromEntity(category)

	// Get additional info
	hasChildren, _ := uc.categoryRepo.HasChildren(ctx, category.ID)
	docCount, _ := uc.categoryRepo.GetDocumentCount(ctx, category.ID, false)
	output.HasChildren = hasChildren
	output.DocumentCount = docCount

	// Get parent name if exists
	if category.ParentID != nil {
		parent, err := uc.categoryRepo.GetByID(ctx, *category.ParentID)
		if err == nil && parent != nil {
			output.ParentName = &parent.Name
		}
	}

	return output, nil
}

// Delete deletes a category
func (uc *CategoryUseCase) Delete(ctx context.Context, id int64) error {
	// Check if category exists
	_, err := uc.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("категория не найдена")
	}

	return uc.categoryRepo.Delete(ctx, id)
}

// GetByID retrieves a category by ID
func (uc *CategoryUseCase) GetByID(ctx context.Context, id int64) (*dto.CategoryOutput, error) {
	category, err := uc.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("категория не найдена")
	}

	output := dto.CategoryFromEntity(category)

	// Get additional info
	hasChildren, _ := uc.categoryRepo.HasChildren(ctx, category.ID)
	docCount, _ := uc.categoryRepo.GetDocumentCount(ctx, category.ID, false)
	output.HasChildren = hasChildren
	output.DocumentCount = docCount

	// Get parent name if exists
	if category.ParentID != nil {
		parent, err := uc.categoryRepo.GetByID(ctx, *category.ParentID)
		if err == nil && parent != nil {
			output.ParentName = &parent.Name
		}
	}

	return output, nil
}

// GetAll retrieves all categories
func (uc *CategoryUseCase) GetAll(ctx context.Context) ([]*dto.CategoryOutput, error) {
	categories, err := uc.categoryRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	outputs := make([]*dto.CategoryOutput, 0, len(categories))
	for _, c := range categories {
		output := dto.CategoryFromEntity(c)
		hasChildren, _ := uc.categoryRepo.HasChildren(ctx, c.ID)
		docCount, _ := uc.categoryRepo.GetDocumentCount(ctx, c.ID, false)
		output.HasChildren = hasChildren
		output.DocumentCount = docCount
		outputs = append(outputs, output)
	}

	return outputs, nil
}

// GetTree retrieves the full category tree
func (uc *CategoryUseCase) GetTree(ctx context.Context) ([]*dto.CategoryTreeOutput, error) {
	tree, err := uc.categoryRepo.GetTree(ctx)
	if err != nil {
		return nil, err
	}

	outputs := make([]*dto.CategoryTreeOutput, 0, len(tree))
	for _, node := range tree {
		outputs = append(outputs, dto.CategoryTreeFromEntity(node))
	}

	return outputs, nil
}

// GetChildren retrieves direct children of a category
func (uc *CategoryUseCase) GetChildren(ctx context.Context, parentID int64) ([]*dto.CategoryOutput, error) {
	children, err := uc.categoryRepo.GetChildren(ctx, parentID)
	if err != nil {
		return nil, err
	}

	outputs := make([]*dto.CategoryOutput, 0, len(children))
	for _, c := range children {
		output := dto.CategoryFromEntity(c)
		hasChildren, _ := uc.categoryRepo.HasChildren(ctx, c.ID)
		docCount, _ := uc.categoryRepo.GetDocumentCount(ctx, c.ID, false)
		output.HasChildren = hasChildren
		output.DocumentCount = docCount
		outputs = append(outputs, output)
	}

	return outputs, nil
}

// GetRootCategories retrieves root categories (no parent)
func (uc *CategoryUseCase) GetRootCategories(ctx context.Context) ([]*dto.CategoryOutput, error) {
	categories, err := uc.categoryRepo.GetByParentID(ctx, nil)
	if err != nil {
		return nil, err
	}

	outputs := make([]*dto.CategoryOutput, 0, len(categories))
	for _, c := range categories {
		output := dto.CategoryFromEntity(c)
		hasChildren, _ := uc.categoryRepo.HasChildren(ctx, c.ID)
		docCount, _ := uc.categoryRepo.GetDocumentCount(ctx, c.ID, false)
		output.HasChildren = hasChildren
		output.DocumentCount = docCount
		outputs = append(outputs, output)
	}

	return outputs, nil
}

// GetWithBreadcrumb retrieves a category with its ancestor path
func (uc *CategoryUseCase) GetWithBreadcrumb(ctx context.Context, id int64) (*dto.CategoryWithBreadcrumbOutput, error) {
	category, err := uc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	ancestors, err := uc.categoryRepo.GetAncestors(ctx, id)
	if err != nil {
		return nil, err
	}

	breadcrumbs := make([]*dto.CategoryBreadcrumb, 0, len(ancestors))
	for _, a := range ancestors {
		breadcrumbs = append(breadcrumbs, &dto.CategoryBreadcrumb{
			ID:   a.ID,
			Name: a.Name,
		})
	}

	return &dto.CategoryWithBreadcrumbOutput{
		Category:    category,
		Breadcrumbs: breadcrumbs,
	}, nil
}

// GetDocumentCount returns document count for a category
func (uc *CategoryUseCase) GetDocumentCount(ctx context.Context, id int64, includeSubcategories bool) (int64, error) {
	return uc.categoryRepo.GetDocumentCount(ctx, id, includeSubcategories)
}
