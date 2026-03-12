// Package persistence provides database implementations for document repositories.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// DocumentTypeRepositoryPG implements DocumentTypeRepository using PostgreSQL
type DocumentTypeRepositoryPG struct {
	db *sql.DB
}

// NewDocumentTypeRepositoryPG creates a new PostgreSQL document type repository
func NewDocumentTypeRepositoryPG(db *sql.DB) *DocumentTypeRepositoryPG {
	return &DocumentTypeRepositoryPG{db: db}
}

// GetAll retrieves all document types
func (r *DocumentTypeRepositoryPG) GetAll(ctx context.Context) ([]*entities.DocumentType, error) {
	query := `
		SELECT id, name, code, description, template_path, template_content, template_variables,
			requires_approval, requires_registration, numbering_pattern, retention_period,
			created_at, updated_at
		FROM document_types ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get document types: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var types []*entities.DocumentType
	for rows.Next() {
		t := &entities.DocumentType{}
		var templateVariablesJSON []byte
		err := rows.Scan(
			&t.ID, &t.Name, &t.Code, &t.Description, &t.TemplatePath,
			&t.TemplateContent, &templateVariablesJSON,
			&t.RequiresApproval, &t.RequiresRegistration, &t.NumberingPattern,
			&t.RetentionPeriod, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document type: %w", err)
		}
		if len(templateVariablesJSON) > 0 {
			if err := json.Unmarshal(templateVariablesJSON, &t.TemplateVariables); err != nil {
				return nil, fmt.Errorf("failed to unmarshal template variables: %w", err)
			}
		}
		types = append(types, t)
	}
	return types, nil
}

// GetByID retrieves a document type by ID
func (r *DocumentTypeRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.DocumentType, error) {
	query := `
		SELECT id, name, code, description, template_path, template_content, template_variables,
			requires_approval, requires_registration, numbering_pattern, retention_period,
			created_at, updated_at
		FROM document_types WHERE id = $1`

	t := &entities.DocumentType{}
	var templateVariablesJSON []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.Name, &t.Code, &t.Description, &t.TemplatePath,
		&t.TemplateContent, &templateVariablesJSON,
		&t.RequiresApproval, &t.RequiresRegistration, &t.NumberingPattern,
		&t.RetentionPeriod, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("document type not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document type: %w", err)
	}
	if len(templateVariablesJSON) > 0 {
		if err := json.Unmarshal(templateVariablesJSON, &t.TemplateVariables); err != nil {
			return nil, fmt.Errorf("failed to unmarshal template variables: %w", err)
		}
	}
	return t, nil
}

// GetByCode retrieves a document type by code
func (r *DocumentTypeRepositoryPG) GetByCode(ctx context.Context, code string) (*entities.DocumentType, error) {
	query := `
		SELECT id, name, code, description, template_path, template_content, template_variables,
			requires_approval, requires_registration, numbering_pattern, retention_period,
			created_at, updated_at
		FROM document_types WHERE code = $1`

	t := &entities.DocumentType{}
	var templateVariablesJSON []byte
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&t.ID, &t.Name, &t.Code, &t.Description, &t.TemplatePath,
		&t.TemplateContent, &templateVariablesJSON,
		&t.RequiresApproval, &t.RequiresRegistration, &t.NumberingPattern,
		&t.RetentionPeriod, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("document type not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document type: %w", err)
	}
	if len(templateVariablesJSON) > 0 {
		if err := json.Unmarshal(templateVariablesJSON, &t.TemplateVariables); err != nil {
			return nil, fmt.Errorf("failed to unmarshal template variables: %w", err)
		}
	}
	return t, nil
}

// UpdateTemplate updates a document type's template content and variables
func (r *DocumentTypeRepositoryPG) UpdateTemplate(ctx context.Context, id int64, content *string, variables []entities.TemplateVariable) error {
	var variablesJSON []byte
	var err error
	if variables != nil {
		variablesJSON, err = json.Marshal(variables)
		if err != nil {
			return fmt.Errorf("failed to marshal template variables: %w", err)
		}
	}

	query := `
		UPDATE document_types
		SET template_content = $1, template_variables = $2, updated_at = NOW()
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, content, variablesJSON, id)
	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("document type not found")
	}
	return nil
}

// TemplateRepositoryAdapter adapts DocumentTypeRepositoryPG for TemplateRepository interface
type TemplateRepositoryAdapter struct {
	repo *DocumentTypeRepositoryPG
}

// NewTemplateRepositoryAdapter creates a new TemplateRepositoryAdapter
func NewTemplateRepositoryAdapter(repo *DocumentTypeRepositoryPG) *TemplateRepositoryAdapter {
	return &TemplateRepositoryAdapter{repo: repo}
}

// GetAll returns all document types (adapts for TemplateRepository interface)
func (a *TemplateRepositoryAdapter) GetAll(ctx context.Context) ([]entities.DocumentType, error) {
	types, err := a.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]entities.DocumentType, len(types))
	for i, t := range types {
		result[i] = *t
	}
	return result, nil
}

// GetByID returns a document type by ID (adapts for TemplateRepository interface)
func (a *TemplateRepositoryAdapter) GetByID(ctx context.Context, id int64) (*entities.DocumentType, error) {
	return a.repo.GetByID(ctx, id)
}

// UpdateTemplate updates a document type's template
func (a *TemplateRepositoryAdapter) UpdateTemplate(ctx context.Context, id int64, content *string, variables []entities.TemplateVariable) error {
	return a.repo.UpdateTemplate(ctx, id, content, variables)
}

// GetAllWithTemplates retrieves all document types that have templates
func (r *DocumentTypeRepositoryPG) GetAllWithTemplates(ctx context.Context) ([]entities.DocumentType, error) {
	query := `
		SELECT id, name, code, description, template_path, template_content, template_variables,
			requires_approval, requires_registration, numbering_pattern, retention_period,
			created_at, updated_at
		FROM document_types
		WHERE template_content IS NOT NULL AND template_content != ''
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get document types with templates: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var types []entities.DocumentType
	for rows.Next() {
		t := entities.DocumentType{}
		var templateVariablesJSON []byte
		err := rows.Scan(
			&t.ID, &t.Name, &t.Code, &t.Description, &t.TemplatePath,
			&t.TemplateContent, &templateVariablesJSON,
			&t.RequiresApproval, &t.RequiresRegistration, &t.NumberingPattern,
			&t.RetentionPeriod, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document type: %w", err)
		}
		if len(templateVariablesJSON) > 0 {
			if err := json.Unmarshal(templateVariablesJSON, &t.TemplateVariables); err != nil {
				return nil, fmt.Errorf("failed to unmarshal template variables: %w", err)
			}
		}
		types = append(types, t)
	}
	return types, nil
}

// DocumentCategoryRepositoryPG implements DocumentCategoryRepository using PostgreSQL
type DocumentCategoryRepositoryPG struct {
	db *sql.DB
}

// NewDocumentCategoryRepositoryPG creates a new PostgreSQL document category repository
func NewDocumentCategoryRepositoryPG(db *sql.DB) *DocumentCategoryRepositoryPG {
	return &DocumentCategoryRepositoryPG{db: db}
}

// GetAll retrieves all document categories
func (r *DocumentCategoryRepositoryPG) GetAll(ctx context.Context) ([]*entities.DocumentCategory, error) {
	query := `
		SELECT id, name, description, parent_id, created_at, updated_at
		FROM document_categories ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get document categories: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var categories []*entities.DocumentCategory
	for rows.Next() {
		c := &entities.DocumentCategory{}
		err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.ParentID, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document category: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// GetByID retrieves a document category by ID
func (r *DocumentCategoryRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.DocumentCategory, error) {
	query := `
		SELECT id, name, description, parent_id, created_at, updated_at
		FROM document_categories WHERE id = $1`

	c := &entities.DocumentCategory{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.Description, &c.ParentID, &c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("document category not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document category: %w", err)
	}
	return c, nil
}

// GetByParentID retrieves document categories by parent ID
func (r *DocumentCategoryRepositoryPG) GetByParentID(ctx context.Context, parentID *int64) ([]*entities.DocumentCategory, error) {
	var query string
	var args []interface{}

	if parentID == nil {
		query = `
			SELECT id, name, description, parent_id, created_at, updated_at
			FROM document_categories WHERE parent_id IS NULL ORDER BY name`
	} else {
		query = `
			SELECT id, name, description, parent_id, created_at, updated_at
			FROM document_categories WHERE parent_id = $1 ORDER BY name`
		args = append(args, *parentID)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get document categories: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var categories []*entities.DocumentCategory
	for rows.Next() {
		c := &entities.DocumentCategory{}
		err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.ParentID, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document category: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// Create creates a new document category
func (r *DocumentCategoryRepositoryPG) Create(ctx context.Context, category *entities.DocumentCategory) error {
	query := `
		INSERT INTO document_categories (name, description, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, category.Name, category.Description, category.ParentID).
		Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create document category: %w", err)
	}
	return nil
}

// Update updates an existing document category
func (r *DocumentCategoryRepositoryPG) Update(ctx context.Context, category *entities.DocumentCategory) error {
	query := `
		UPDATE document_categories
		SET name = $1, description = $2, parent_id = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query, category.Name, category.Description, category.ParentID, category.ID).
		Scan(&category.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("document category not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update document category: %w", err)
	}
	return nil
}

// Delete deletes a document category
func (r *DocumentCategoryRepositoryPG) Delete(ctx context.Context, id int64) error {
	// First, update children to have no parent (or could cascade delete)
	_, err := r.db.ExecContext(ctx, `UPDATE document_categories SET parent_id = NULL WHERE parent_id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to update child categories: %w", err)
	}

	// Update documents in this category to have no category
	_, err = r.db.ExecContext(ctx, `UPDATE documents SET category_id = NULL WHERE category_id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to update documents: %w", err)
	}

	result, err := r.db.ExecContext(ctx, `DELETE FROM document_categories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete document category: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("document category not found")
	}
	return nil
}

// GetTree retrieves the full category tree
func (r *DocumentCategoryRepositoryPG) GetTree(ctx context.Context) ([]*entities.CategoryTreeNode, error) {
	// Get all categories with document counts
	query := `
		SELECT c.id, c.name, c.description, c.parent_id, c.created_at, c.updated_at,
			COALESCE((SELECT COUNT(*) FROM documents d WHERE d.category_id = c.id AND d.deleted_at IS NULL), 0) as doc_count
		FROM document_categories c
		ORDER BY c.name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer func() { _ = rows.Close() }()

	// Build map of all categories
	categoryMap := make(map[int64]*entities.CategoryTreeNode)
	var allNodes []*entities.CategoryTreeNode

	for rows.Next() {
		node := &entities.CategoryTreeNode{}
		err := rows.Scan(&node.ID, &node.Name, &node.Description, &node.ParentID,
			&node.CreatedAt, &node.UpdatedAt, &node.DocumentCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		node.Children = []*entities.CategoryTreeNode{}
		categoryMap[node.ID] = node
		allNodes = append(allNodes, node)
	}

	// Build tree structure
	var roots []*entities.CategoryTreeNode
	for _, node := range allNodes {
		if node.ParentID == nil {
			roots = append(roots, node)
		} else {
			parent, exists := categoryMap[*node.ParentID]
			if exists {
				parent.Children = append(parent.Children, node)
			} else {
				// Orphan node, add to roots
				roots = append(roots, node)
			}
		}
	}

	return roots, nil
}

// GetChildren retrieves direct children of a category
func (r *DocumentCategoryRepositoryPG) GetChildren(ctx context.Context, parentID int64) ([]*entities.DocumentCategory, error) {
	query := `
		SELECT id, name, description, parent_id, created_at, updated_at
		FROM document_categories WHERE parent_id = $1 ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var categories []*entities.DocumentCategory
	for rows.Next() {
		c := &entities.DocumentCategory{}
		err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.ParentID, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// GetAncestors retrieves all ancestors (path to root) of a category
func (r *DocumentCategoryRepositoryPG) GetAncestors(ctx context.Context, id int64) ([]*entities.DocumentCategory, error) {
	// Use recursive CTE to get ancestors
	query := `
		WITH RECURSIVE ancestors AS (
			SELECT id, name, description, parent_id, created_at, updated_at, 0 as depth
			FROM document_categories WHERE id = $1
			UNION ALL
			SELECT c.id, c.name, c.description, c.parent_id, c.created_at, c.updated_at, a.depth + 1
			FROM document_categories c
			INNER JOIN ancestors a ON c.id = a.parent_id
		)
		SELECT id, name, description, parent_id, created_at, updated_at
		FROM ancestors
		WHERE id != $1
		ORDER BY depth DESC`

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get ancestors: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var ancestors []*entities.DocumentCategory
	for rows.Next() {
		c := &entities.DocumentCategory{}
		err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.ParentID, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ancestor: %w", err)
		}
		ancestors = append(ancestors, c)
	}
	return ancestors, nil
}

// HasChildren checks if a category has child categories
func (r *DocumentCategoryRepositoryPG) HasChildren(ctx context.Context, id int64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM document_categories WHERE parent_id = $1`, id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check children: %w", err)
	}
	return count > 0, nil
}

// GetDocumentCount returns the number of documents in a category
func (r *DocumentCategoryRepositoryPG) GetDocumentCount(ctx context.Context, id int64, includeSubcategories bool) (int64, error) {
	var count int64

	if includeSubcategories {
		// Use recursive CTE to get all subcategory IDs
		query := `
			WITH RECURSIVE subcategories AS (
				SELECT id FROM document_categories WHERE id = $1
				UNION ALL
				SELECT c.id FROM document_categories c
				INNER JOIN subcategories s ON c.parent_id = s.id
			)
			SELECT COUNT(*) FROM documents
			WHERE category_id IN (SELECT id FROM subcategories) AND deleted_at IS NULL`
		err := r.db.QueryRowContext(ctx, query, id).Scan(&count)
		if err != nil {
			return 0, fmt.Errorf("failed to count documents: %w", err)
		}
	} else {
		err := r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM documents WHERE category_id = $1 AND deleted_at IS NULL`, id).Scan(&count)
		if err != nil {
			return 0, fmt.Errorf("failed to count documents: %w", err)
		}
	}

	return count, nil
}
