// Package repositories defines interfaces for document persistence.
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// DocumentRepository defines the interface for document persistence
type DocumentRepository interface {
	// CRUD operations
	Create(ctx context.Context, doc *entities.Document) error
	Update(ctx context.Context, doc *entities.Document) error
	GetByID(ctx context.Context, id int64) (*entities.Document, error)
	Delete(ctx context.Context, id int64) error
	SoftDelete(ctx context.Context, id int64) error

	// Query operations
	List(ctx context.Context, filter DocumentFilter) ([]*entities.Document, int64, error)
	GetByAuthorID(ctx context.Context, authorID int64, limit, offset int) ([]*entities.Document, error)
	GetByStatus(ctx context.Context, status entities.DocumentStatus, limit, offset int) ([]*entities.Document, error)

	// Full-text search operations
	Search(ctx context.Context, filter SearchFilter) ([]*SearchResult, int64, error)

	// Version operations
	CreateVersion(ctx context.Context, version *entities.DocumentVersion) error
	GetVersions(ctx context.Context, documentID int64) ([]*entities.DocumentVersion, error)
	GetVersion(ctx context.Context, documentID int64, version int) (*entities.DocumentVersion, error)
	GetLatestVersion(ctx context.Context, documentID int64) (*entities.DocumentVersion, error)
	RestoreVersion(ctx context.Context, documentID int64, version int, userID int64) error
	DeleteVersion(ctx context.Context, documentID int64, version int) error

	// Version diff operations
	CreateVersionDiff(ctx context.Context, diff *entities.DocumentVersionDiff) error
	GetVersionDiff(ctx context.Context, documentID int64, fromVersion, toVersion int) (*entities.DocumentVersionDiff, error)
	CompareVersions(ctx context.Context, documentID int64, fromVersion, toVersion int) (*entities.DocumentVersionDiff, error)

	// History operations
	AddHistory(ctx context.Context, history *entities.DocumentHistory) error
	GetHistory(ctx context.Context, documentID int64) ([]*entities.DocumentHistory, error)
}

// DocumentFilter contains filter options for listing documents
type DocumentFilter struct {
	AuthorID       *int64
	RecipientID    *int64
	DocumentTypeID *int64
	CategoryID     *int64
	Status         *entities.DocumentStatus
	Importance     *entities.DocumentImportance
	IsPublic       *bool
	SearchQuery    *string // search in title and subject
	FromDate       *string // created_at >= from_date
	ToDate         *string // created_at <= to_date
	IncludeDeleted bool
	Limit          int
	Offset         int
	OrderBy        string // e.g., "created_at DESC"
	// Access control fields
	CurrentUserID   int64  // Required for access control
	CurrentUserRole string // User role for role-based permissions
}

// SearchFilter contains options for full-text search
type SearchFilter struct {
	Query          string                       // search query text
	DocumentTypeID *int64                       // filter by document type
	CategoryID     *int64                       // filter by category
	AuthorID       *int64                       // filter by author
	Status         *entities.DocumentStatus     // filter by status
	Importance     *entities.DocumentImportance // filter by importance
	FromDate       *string                      // created_at >= from_date
	ToDate         *string                      // created_at <= to_date
	IncludeDeleted bool                         // include soft-deleted documents
	Limit          int                          // pagination limit
	Offset         int                          // pagination offset
	// Access control fields
	CurrentUserID   int64  // Required for access control
	CurrentUserRole string // User role for role-based permissions
}

// SearchResult represents a document search result with highlighted matches
type SearchResult struct {
	Document           *entities.Document `json:"document"`
	Rank               float64            `json:"rank"`                 // relevance score
	HighlightedTitle   string             `json:"highlighted_title"`    // title with highlighted matches
	HighlightedSubject string             `json:"highlighted_subject"`  // subject with highlighted matches
	HighlightedContent string             `json:"highlighted_content"`  // content snippet with highlighted matches
}

// DocumentTypeRepository defines the interface for document type persistence
type DocumentTypeRepository interface {
	GetAll(ctx context.Context) ([]*entities.DocumentType, error)
	GetByID(ctx context.Context, id int64) (*entities.DocumentType, error)
	GetByCode(ctx context.Context, code string) (*entities.DocumentType, error)
}

// DocumentCategoryRepository defines the interface for document category persistence
type DocumentCategoryRepository interface {
	// CRUD operations
	Create(ctx context.Context, category *entities.DocumentCategory) error
	Update(ctx context.Context, category *entities.DocumentCategory) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*entities.DocumentCategory, error)
	GetAll(ctx context.Context) ([]*entities.DocumentCategory, error)

	// Hierarchy operations
	GetByParentID(ctx context.Context, parentID *int64) ([]*entities.DocumentCategory, error)
	GetTree(ctx context.Context) ([]*entities.CategoryTreeNode, error)
	GetChildren(ctx context.Context, parentID int64) ([]*entities.DocumentCategory, error)
	GetAncestors(ctx context.Context, id int64) ([]*entities.DocumentCategory, error)
	HasChildren(ctx context.Context, id int64) (bool, error)
	GetDocumentCount(ctx context.Context, id int64, includeSubcategories bool) (int64, error)
}

// DocumentTagRepository defines the interface for document tag persistence
type DocumentTagRepository interface {
	// CRUD operations
	Create(ctx context.Context, tag *entities.DocumentTag) error
	Update(ctx context.Context, tag *entities.DocumentTag) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*entities.DocumentTag, error)
	GetByName(ctx context.Context, name string) (*entities.DocumentTag, error)
	GetAll(ctx context.Context) ([]*entities.DocumentTag, error)
	Search(ctx context.Context, query string, limit int) ([]*entities.DocumentTag, error)

	// Document-tag relations
	AddTagToDocument(ctx context.Context, documentID, tagID int64) error
	RemoveTagFromDocument(ctx context.Context, documentID, tagID int64) error
	GetTagsByDocumentID(ctx context.Context, documentID int64) ([]*entities.DocumentTag, error)
	GetDocumentsByTagID(ctx context.Context, tagID int64, limit, offset int) ([]int64, int64, error)
	SetDocumentTags(ctx context.Context, documentID int64, tagIDs []int64) error
	GetTagUsageCount(ctx context.Context, tagID int64) (int64, error)
}
