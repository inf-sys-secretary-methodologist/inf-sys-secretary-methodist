package http_test

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
)

// --- MockDocumentRepository ---

type MockDocumentRepository struct {
	mock.Mock
}

func (m *MockDocumentRepository) Create(ctx context.Context, doc *entities.Document) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

func (m *MockDocumentRepository) Update(ctx context.Context, doc *entities.Document) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

func (m *MockDocumentRepository) GetByID(ctx context.Context, id int64) (*entities.Document, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Document), args.Error(1)
}

func (m *MockDocumentRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDocumentRepository) SoftDelete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDocumentRepository) List(ctx context.Context, filter repositories.DocumentFilter) ([]*entities.Document, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entities.Document), args.Get(1).(int64), args.Error(2)
}

func (m *MockDocumentRepository) GetByAuthorID(ctx context.Context, authorID int64, limit, offset int) ([]*entities.Document, error) {
	args := m.Called(ctx, authorID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Document), args.Error(1)
}

func (m *MockDocumentRepository) GetByStatus(ctx context.Context, status entities.DocumentStatus, limit, offset int) ([]*entities.Document, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Document), args.Error(1)
}

func (m *MockDocumentRepository) Search(ctx context.Context, filter repositories.SearchFilter) ([]*repositories.SearchResult, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*repositories.SearchResult), args.Get(1).(int64), args.Error(2)
}

func (m *MockDocumentRepository) CreateVersion(ctx context.Context, version *entities.DocumentVersion) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

func (m *MockDocumentRepository) GetVersions(ctx context.Context, documentID int64) ([]*entities.DocumentVersion, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentVersion), args.Error(1)
}

func (m *MockDocumentRepository) GetVersion(ctx context.Context, documentID int64, version int) (*entities.DocumentVersion, error) {
	args := m.Called(ctx, documentID, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentVersion), args.Error(1)
}

func (m *MockDocumentRepository) GetLatestVersion(ctx context.Context, documentID int64) (*entities.DocumentVersion, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentVersion), args.Error(1)
}

func (m *MockDocumentRepository) RestoreVersion(ctx context.Context, documentID int64, version int, userID int64) error {
	args := m.Called(ctx, documentID, version, userID)
	return args.Error(0)
}

func (m *MockDocumentRepository) DeleteVersion(ctx context.Context, documentID int64, version int) error {
	args := m.Called(ctx, documentID, version)
	return args.Error(0)
}

func (m *MockDocumentRepository) CreateVersionDiff(ctx context.Context, diff *entities.DocumentVersionDiff) error {
	args := m.Called(ctx, diff)
	return args.Error(0)
}

func (m *MockDocumentRepository) GetVersionDiff(ctx context.Context, documentID int64, fromVersion, toVersion int) (*entities.DocumentVersionDiff, error) {
	args := m.Called(ctx, documentID, fromVersion, toVersion)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentVersionDiff), args.Error(1)
}

func (m *MockDocumentRepository) CompareVersions(ctx context.Context, documentID int64, fromVersion, toVersion int) (*entities.DocumentVersionDiff, error) {
	args := m.Called(ctx, documentID, fromVersion, toVersion)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentVersionDiff), args.Error(1)
}

func (m *MockDocumentRepository) AddHistory(ctx context.Context, history *entities.DocumentHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *MockDocumentRepository) GetHistory(ctx context.Context, documentID int64) ([]*entities.DocumentHistory, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentHistory), args.Error(1)
}

// --- MockDocumentTypeRepository ---

type MockDocumentTypeRepository struct {
	mock.Mock
}

func (m *MockDocumentTypeRepository) GetAll(ctx context.Context) ([]*entities.DocumentType, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentType), args.Error(1)
}

func (m *MockDocumentTypeRepository) GetByID(ctx context.Context, id int64) (*entities.DocumentType, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentType), args.Error(1)
}

func (m *MockDocumentTypeRepository) GetByCode(ctx context.Context, code string) (*entities.DocumentType, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentType), args.Error(1)
}

// --- MockDocumentCategoryRepository ---

type MockDocumentCategoryRepository struct {
	mock.Mock
}

func (m *MockDocumentCategoryRepository) Create(ctx context.Context, category *entities.DocumentCategory) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockDocumentCategoryRepository) Update(ctx context.Context, category *entities.DocumentCategory) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockDocumentCategoryRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDocumentCategoryRepository) GetByID(ctx context.Context, id int64) (*entities.DocumentCategory, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentCategory), args.Error(1)
}

func (m *MockDocumentCategoryRepository) GetAll(ctx context.Context) ([]*entities.DocumentCategory, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentCategory), args.Error(1)
}

func (m *MockDocumentCategoryRepository) GetByParentID(ctx context.Context, parentID *int64) ([]*entities.DocumentCategory, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentCategory), args.Error(1)
}

func (m *MockDocumentCategoryRepository) GetTree(ctx context.Context) ([]*entities.CategoryTreeNode, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.CategoryTreeNode), args.Error(1)
}

func (m *MockDocumentCategoryRepository) GetChildren(ctx context.Context, parentID int64) ([]*entities.DocumentCategory, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentCategory), args.Error(1)
}

func (m *MockDocumentCategoryRepository) GetAncestors(ctx context.Context, id int64) ([]*entities.DocumentCategory, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentCategory), args.Error(1)
}

func (m *MockDocumentCategoryRepository) HasChildren(ctx context.Context, id int64) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockDocumentCategoryRepository) GetDocumentCount(ctx context.Context, id int64, includeSubcategories bool) (int64, error) {
	args := m.Called(ctx, id, includeSubcategories)
	return args.Get(0).(int64), args.Error(1)
}

// --- MockDocumentTagRepository ---

type MockDocumentTagRepository struct {
	mock.Mock
}

func (m *MockDocumentTagRepository) Create(ctx context.Context, tag *entities.DocumentTag) error {
	args := m.Called(ctx, tag)
	return args.Error(0)
}

func (m *MockDocumentTagRepository) Update(ctx context.Context, tag *entities.DocumentTag) error {
	args := m.Called(ctx, tag)
	return args.Error(0)
}

func (m *MockDocumentTagRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDocumentTagRepository) GetByID(ctx context.Context, id int64) (*entities.DocumentTag, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentTag), args.Error(1)
}

func (m *MockDocumentTagRepository) GetByName(ctx context.Context, name string) (*entities.DocumentTag, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentTag), args.Error(1)
}

func (m *MockDocumentTagRepository) GetAll(ctx context.Context) ([]*entities.DocumentTag, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentTag), args.Error(1)
}

func (m *MockDocumentTagRepository) Search(ctx context.Context, query string, limit int) ([]*entities.DocumentTag, error) {
	args := m.Called(ctx, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentTag), args.Error(1)
}

func (m *MockDocumentTagRepository) AddTagToDocument(ctx context.Context, documentID, tagID int64) error {
	args := m.Called(ctx, documentID, tagID)
	return args.Error(0)
}

func (m *MockDocumentTagRepository) RemoveTagFromDocument(ctx context.Context, documentID, tagID int64) error {
	args := m.Called(ctx, documentID, tagID)
	return args.Error(0)
}

func (m *MockDocumentTagRepository) GetTagsByDocumentID(ctx context.Context, documentID int64) ([]*entities.DocumentTag, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentTag), args.Error(1)
}

func (m *MockDocumentTagRepository) GetDocumentsByTagID(ctx context.Context, tagID int64, limit, offset int) ([]int64, int64, error) {
	args := m.Called(ctx, tagID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockDocumentTagRepository) SetDocumentTags(ctx context.Context, documentID int64, tagIDs []int64) error {
	args := m.Called(ctx, documentID, tagIDs)
	return args.Error(0)
}

func (m *MockDocumentTagRepository) GetTagUsageCount(ctx context.Context, tagID int64) (int64, error) {
	args := m.Called(ctx, tagID)
	return args.Get(0).(int64), args.Error(1)
}

// --- MockPermissionRepository ---

type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) Create(ctx context.Context, permission *entities.DocumentPermission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func (m *MockPermissionRepository) Update(ctx context.Context, permission *entities.DocumentPermission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func (m *MockPermissionRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPermissionRepository) GetByID(ctx context.Context, id int64) (*entities.DocumentPermission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentPermission), args.Error(1)
}

func (m *MockPermissionRepository) GetByDocumentID(ctx context.Context, documentID int64) ([]*entities.DocumentPermission, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentPermission), args.Error(1)
}

func (m *MockPermissionRepository) GetByUserID(ctx context.Context, userID int64) ([]*entities.DocumentPermission, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentPermission), args.Error(1)
}

func (m *MockPermissionRepository) GetByUserIDOrRole(ctx context.Context, userID int64, role string) ([]*entities.DocumentPermission, error) {
	args := m.Called(ctx, userID, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentPermission), args.Error(1)
}

func (m *MockPermissionRepository) GetByGrantedBy(ctx context.Context, userID int64) ([]*entities.DocumentPermission, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentPermission), args.Error(1)
}

func (m *MockPermissionRepository) GetByDocumentAndUser(ctx context.Context, documentID, userID int64) (*entities.DocumentPermission, error) {
	args := m.Called(ctx, documentID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentPermission), args.Error(1)
}

func (m *MockPermissionRepository) GetByDocumentAndRole(ctx context.Context, documentID int64, role entities.UserRole) (*entities.DocumentPermission, error) {
	args := m.Called(ctx, documentID, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentPermission), args.Error(1)
}

func (m *MockPermissionRepository) HasPermission(ctx context.Context, documentID, userID int64, permission entities.PermissionLevel) (bool, error) {
	args := m.Called(ctx, documentID, userID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *MockPermissionRepository) HasAnyPermission(ctx context.Context, documentID, userID int64) (bool, error) {
	args := m.Called(ctx, documentID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPermissionRepository) GetUserPermissionLevel(ctx context.Context, documentID, userID int64, userRole entities.UserRole) (*entities.PermissionLevel, error) {
	args := m.Called(ctx, documentID, userID, userRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.PermissionLevel), args.Error(1)
}

func (m *MockPermissionRepository) DeleteByDocumentID(ctx context.Context, documentID int64) error {
	args := m.Called(ctx, documentID)
	return args.Error(0)
}

func (m *MockPermissionRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockPermissionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// --- MockPublicLinkRepository ---

type MockPublicLinkRepository struct {
	mock.Mock
}

func (m *MockPublicLinkRepository) Create(ctx context.Context, link *entities.PublicLink) error {
	args := m.Called(ctx, link)
	return args.Error(0)
}

func (m *MockPublicLinkRepository) Update(ctx context.Context, link *entities.PublicLink) error {
	args := m.Called(ctx, link)
	return args.Error(0)
}

func (m *MockPublicLinkRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPublicLinkRepository) GetByID(ctx context.Context, id int64) (*entities.PublicLink, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.PublicLink), args.Error(1)
}

func (m *MockPublicLinkRepository) GetByToken(ctx context.Context, token string) (*entities.PublicLink, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.PublicLink), args.Error(1)
}

func (m *MockPublicLinkRepository) GetByDocumentID(ctx context.Context, documentID int64) ([]*entities.PublicLink, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.PublicLink), args.Error(1)
}

func (m *MockPublicLinkRepository) GetByCreatedBy(ctx context.Context, userID int64) ([]*entities.PublicLink, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.PublicLink), args.Error(1)
}

func (m *MockPublicLinkRepository) GetActiveByDocumentID(ctx context.Context, documentID int64) ([]*entities.PublicLink, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.PublicLink), args.Error(1)
}

func (m *MockPublicLinkRepository) IncrementUseCount(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPublicLinkRepository) Deactivate(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPublicLinkRepository) Activate(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPublicLinkRepository) DeleteByDocumentID(ctx context.Context, documentID int64) error {
	args := m.Called(ctx, documentID)
	return args.Error(0)
}

func (m *MockPublicLinkRepository) DeactivateExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// --- MockTemplateRepository ---

type MockTemplateRepository struct {
	mock.Mock
}

func (m *MockTemplateRepository) GetAll(ctx context.Context) ([]entities.DocumentType, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.DocumentType), args.Error(1)
}

func (m *MockTemplateRepository) GetByID(ctx context.Context, id int64) (*entities.DocumentType, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentType), args.Error(1)
}

func (m *MockTemplateRepository) UpdateTemplate(ctx context.Context, id int64, content *string, variables []entities.TemplateVariable) error {
	args := m.Called(ctx, id, content, variables)
	return args.Error(0)
}
