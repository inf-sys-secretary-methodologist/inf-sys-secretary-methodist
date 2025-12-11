package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
)

// MockDocumentRepository is a mock implementation of DocumentRepository
type MockDocumentRepository struct {
	mock.Mock
}

func (m *MockDocumentRepository) Create(ctx context.Context, doc *entities.Document) error {
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

func (m *MockDocumentRepository) Update(ctx context.Context, doc *entities.Document) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
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

func (m *MockDocumentRepository) GetByRegistrationNumber(ctx context.Context, number string) (*entities.Document, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Document), args.Error(1)
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

func (m *MockDocumentRepository) GetVersion(ctx context.Context, documentID int64, version int) (*entities.DocumentVersion, error) {
	args := m.Called(ctx, documentID, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentVersion), args.Error(1)
}

func (m *MockDocumentRepository) Search(ctx context.Context, filter repositories.SearchFilter) ([]*repositories.SearchResult, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*repositories.SearchResult), args.Get(1).(int64), args.Error(2)
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

func (m *MockDocumentRepository) CompareVersions(ctx context.Context, documentID int64, fromVersion, toVersion int) (*entities.DocumentVersionDiff, error) {
	args := m.Called(ctx, documentID, fromVersion, toVersion)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentVersionDiff), args.Error(1)
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

// MockDocumentTypeRepository is a mock implementation of DocumentTypeRepository
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

// MockDocumentCategoryRepository is a mock implementation of DocumentCategoryRepository
type MockDocumentCategoryRepository struct {
	mock.Mock
}

func (m *MockDocumentCategoryRepository) GetAll(ctx context.Context) ([]*entities.DocumentCategory, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentCategory), args.Error(1)
}

func (m *MockDocumentCategoryRepository) GetByID(ctx context.Context, id int64) (*entities.DocumentCategory, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentCategory), args.Error(1)
}

func (m *MockDocumentCategoryRepository) GetByParentID(ctx context.Context, parentID *int64) ([]*entities.DocumentCategory, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentCategory), args.Error(1)
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

func TestDocumentUseCase_GetByID(t *testing.T) {
	mockDocRepo := new(MockDocumentRepository)
	mockTypeRepo := new(MockDocumentTypeRepository)
	mockCategoryRepo := new(MockDocumentCategoryRepository)

	usecase := NewDocumentUseCase(mockDocRepo, mockTypeRepo, mockCategoryRepo, nil, nil)

	ctx := context.Background()

	t.Run("existing document", func(t *testing.T) {
		expectedDoc := &entities.Document{
			ID:             1,
			Title:          "Test Document",
			DocumentTypeID: 1,
			AuthorID:       1,
			Status:         entities.DocumentStatusDraft,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(expectedDoc, nil).Once()

		result, err := usecase.GetByID(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedDoc.ID, result.ID)
		assert.Equal(t, expectedDoc.Title, result.Title)

		mockDocRepo.AssertExpectations(t)
	})

	t.Run("non-existing document", func(t *testing.T) {
		mockDocRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		result, err := usecase.GetByID(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, result)

		mockDocRepo.AssertExpectations(t)
	})
}

func TestDocumentUseCase_List(t *testing.T) {
	mockDocRepo := new(MockDocumentRepository)
	mockTypeRepo := new(MockDocumentTypeRepository)
	mockCategoryRepo := new(MockDocumentCategoryRepository)

	usecase := NewDocumentUseCase(mockDocRepo, mockTypeRepo, mockCategoryRepo, nil, nil)

	ctx := context.Background()

	t.Run("list documents with pagination", func(t *testing.T) {
		docs := []*entities.Document{
			{ID: 1, Title: "Doc 1", Status: entities.DocumentStatusDraft},
			{ID: 2, Title: "Doc 2", Status: entities.DocumentStatusDraft},
		}

		mockDocRepo.On("List", ctx, mock.AnythingOfType("repositories.DocumentFilter")).
			Return(docs, int64(2), nil).Once()

		filter := dto.DocumentFilterInput{
			Page:     1,
			PageSize: 10,
		}

		result, err := usecase.List(ctx, filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Documents, 2)
		assert.Equal(t, int64(2), result.Total)

		mockDocRepo.AssertExpectations(t)
	})

	t.Run("empty result", func(t *testing.T) {
		mockDocRepo.On("List", ctx, mock.AnythingOfType("repositories.DocumentFilter")).
			Return([]*entities.Document{}, int64(0), nil).Once()

		filter := dto.DocumentFilterInput{
			Page:     1,
			PageSize: 10,
		}

		result, err := usecase.List(ctx, filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Documents, 0)
		assert.Equal(t, int64(0), result.Total)

		mockDocRepo.AssertExpectations(t)
	})
}

func TestDocumentUseCase_Delete(t *testing.T) {
	mockDocRepo := new(MockDocumentRepository)
	mockTypeRepo := new(MockDocumentTypeRepository)
	mockCategoryRepo := new(MockDocumentCategoryRepository)

	usecase := NewDocumentUseCase(mockDocRepo, mockTypeRepo, mockCategoryRepo, nil, nil)

	ctx := context.Background()

	t.Run("delete existing document", func(t *testing.T) {
		doc := &entities.Document{ID: 1, Title: "Test"}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil).Once()
		mockDocRepo.On("SoftDelete", ctx, int64(1)).Return(nil).Once()
		mockDocRepo.On("AddHistory", ctx, mock.AnythingOfType("*entities.DocumentHistory")).Return(nil).Once()

		err := usecase.Delete(ctx, 1, 1)

		assert.NoError(t, err)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("delete non-existing document", func(t *testing.T) {
		mockDocRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		err := usecase.Delete(ctx, 999, 1)

		assert.Error(t, err)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestDocumentUseCase_GetDocumentTypes(t *testing.T) {
	mockDocRepo := new(MockDocumentRepository)
	mockTypeRepo := new(MockDocumentTypeRepository)
	mockCategoryRepo := new(MockDocumentCategoryRepository)

	usecase := NewDocumentUseCase(mockDocRepo, mockTypeRepo, mockCategoryRepo, nil, nil)

	ctx := context.Background()

	t.Run("get all types", func(t *testing.T) {
		types := []*entities.DocumentType{
			{ID: 1, Name: "Приказ", Code: "order"},
			{ID: 2, Name: "Распоряжение", Code: "directive"},
		}

		mockTypeRepo.On("GetAll", ctx).Return(types, nil).Once()

		result, err := usecase.GetDocumentTypes(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Приказ", result[0].Name)

		mockTypeRepo.AssertExpectations(t)
	})
}

func TestDocumentUseCase_GetCategories(t *testing.T) {
	mockDocRepo := new(MockDocumentRepository)
	mockTypeRepo := new(MockDocumentTypeRepository)
	mockCategoryRepo := new(MockDocumentCategoryRepository)

	usecase := NewDocumentUseCase(mockDocRepo, mockTypeRepo, mockCategoryRepo, nil, nil)

	ctx := context.Background()

	t.Run("get all categories", func(t *testing.T) {
		categories := []*entities.DocumentCategory{
			{ID: 1, Name: "Входящие"},
			{ID: 2, Name: "Исходящие"},
		}

		mockCategoryRepo.On("GetAll", ctx).Return(categories, nil).Once()

		result, err := usecase.GetCategories(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Входящие", result[0].Name)

		mockCategoryRepo.AssertExpectations(t)
	})
}

func TestDocumentUseCase_Update(t *testing.T) {
	mockDocRepo := new(MockDocumentRepository)
	mockTypeRepo := new(MockDocumentTypeRepository)
	mockCategoryRepo := new(MockDocumentCategoryRepository)

	usecase := NewDocumentUseCase(mockDocRepo, mockTypeRepo, mockCategoryRepo, nil, nil)

	ctx := context.Background()

	t.Run("update document title", func(t *testing.T) {
		existingDoc := &entities.Document{
			ID:             1,
			Title:          "Old Title",
			DocumentTypeID: 1,
			AuthorID:       1,
			Status:         entities.DocumentStatusDraft,
		}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(existingDoc, nil).Once()
		mockDocRepo.On("Update", ctx, mock.AnythingOfType("*entities.Document")).Return(nil).Once()
		mockDocRepo.On("AddHistory", ctx, mock.AnythingOfType("*entities.DocumentHistory")).Return(nil).Once()

		newTitle := "New Title"
		input := dto.UpdateDocumentInput{
			Title: &newTitle,
		}

		result, err := usecase.Update(ctx, 1, input, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "New Title", result.Title)

		mockDocRepo.AssertExpectations(t)
	})
}

func TestDocumentUseCase_Search(t *testing.T) {
	mockDocRepo := new(MockDocumentRepository)
	mockTypeRepo := new(MockDocumentTypeRepository)
	mockCategoryRepo := new(MockDocumentCategoryRepository)

	usecase := NewDocumentUseCase(mockDocRepo, mockTypeRepo, mockCategoryRepo, nil, nil)

	ctx := context.Background()

	t.Run("successful search with results", func(t *testing.T) {
		searchResults := []*repositories.SearchResult{
			{
				Document: &entities.Document{
					ID:             1,
					Title:          "Приказ о проведении мероприятия",
					DocumentTypeID: 1,
					AuthorID:       1,
					Status:         entities.DocumentStatusDraft,
				},
				Rank:               0.95,
				HighlightedTitle:   "<b>Приказ</b> о проведении мероприятия",
				HighlightedSubject: "",
				HighlightedContent: "",
			},
			{
				Document: &entities.Document{
					ID:             2,
					Title:          "Приказ о назначении",
					DocumentTypeID: 1,
					AuthorID:       2,
					Status:         entities.DocumentStatusApproved,
				},
				Rank:               0.85,
				HighlightedTitle:   "<b>Приказ</b> о назначении",
				HighlightedSubject: "",
				HighlightedContent: "",
			},
		}

		mockDocRepo.On("Search", ctx, mock.AnythingOfType("repositories.SearchFilter")).
			Return(searchResults, int64(2), nil).Once()

		input := dto.SearchInput{
			Query:    "приказ",
			Page:     1,
			PageSize: 20,
		}

		result, err := usecase.Search(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, 2)
		assert.Equal(t, int64(2), result.Total)
		assert.Equal(t, "приказ", result.Query)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.PageSize)
		assert.Equal(t, 0.95, result.Results[0].Rank)
		assert.Contains(t, result.Results[0].HighlightedTitle, "<b>Приказ</b>")

		mockDocRepo.AssertExpectations(t)
	})

	t.Run("empty search query", func(t *testing.T) {
		input := dto.SearchInput{
			Query:    "",
			Page:     1,
			PageSize: 20,
		}

		result, err := usecase.Search(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "search query is required")
	})

	t.Run("search with no results", func(t *testing.T) {
		mockDocRepo.On("Search", ctx, mock.AnythingOfType("repositories.SearchFilter")).
			Return([]*repositories.SearchResult{}, int64(0), nil).Once()

		input := dto.SearchInput{
			Query:    "несуществующий документ",
			Page:     1,
			PageSize: 20,
		}

		result, err := usecase.Search(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, 0)
		assert.Equal(t, int64(0), result.Total)
		assert.Equal(t, 0, result.TotalPages)

		mockDocRepo.AssertExpectations(t)
	})

	t.Run("search with filters", func(t *testing.T) {
		statusApproved := "approved"
		searchResults := []*repositories.SearchResult{
			{
				Document: &entities.Document{
					ID:             1,
					Title:          "Утвержденный приказ",
					DocumentTypeID: 1,
					AuthorID:       1,
					Status:         entities.DocumentStatusApproved,
				},
				Rank:             0.9,
				HighlightedTitle: "Утвержденный <b>приказ</b>",
			},
		}

		mockDocRepo.On("Search", ctx, mock.AnythingOfType("repositories.SearchFilter")).
			Return(searchResults, int64(1), nil).Once()

		input := dto.SearchInput{
			Query:    "приказ",
			Status:   &statusApproved,
			Page:     1,
			PageSize: 10,
		}

		result, err := usecase.Search(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Results, 1)
		assert.Equal(t, int64(1), result.Total)

		mockDocRepo.AssertExpectations(t)
	})

	t.Run("search with pagination", func(t *testing.T) {
		searchResults := []*repositories.SearchResult{
			{
				Document: &entities.Document{
					ID:    11,
					Title: "Документ на странице 2",
				},
				Rank: 0.8,
			},
		}

		mockDocRepo.On("Search", ctx, mock.MatchedBy(func(filter repositories.SearchFilter) bool {
			return filter.Offset == 10 && filter.Limit == 10
		})).Return(searchResults, int64(15), nil).Once()

		input := dto.SearchInput{
			Query:    "документ",
			Page:     2,
			PageSize: 10,
		}

		result, err := usecase.Search(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.Page)
		assert.Equal(t, 10, result.PageSize)
		assert.Equal(t, int64(15), result.Total)
		assert.Equal(t, 2, result.TotalPages)

		mockDocRepo.AssertExpectations(t)
	})

	t.Run("search with repository error", func(t *testing.T) {
		mockDocRepo.On("Search", ctx, mock.AnythingOfType("repositories.SearchFilter")).
			Return(nil, int64(0), assert.AnError).Once()

		input := dto.SearchInput{
			Query:    "приказ",
			Page:     1,
			PageSize: 20,
		}

		result, err := usecase.Search(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to search documents")

		mockDocRepo.AssertExpectations(t)
	})

	t.Run("search with default pagination", func(t *testing.T) {
		searchResults := []*repositories.SearchResult{}

		mockDocRepo.On("Search", ctx, mock.MatchedBy(func(filter repositories.SearchFilter) bool {
			return filter.Offset == 0 && filter.Limit == 20
		})).Return(searchResults, int64(0), nil).Once()

		input := dto.SearchInput{
			Query:    "тест",
			Page:     0,
			PageSize: 0,
		}

		result, err := usecase.Search(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.PageSize)

		mockDocRepo.AssertExpectations(t)
	})
}
