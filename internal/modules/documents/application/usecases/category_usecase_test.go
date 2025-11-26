package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// MockCategoryRepository is a mock implementation of DocumentCategoryRepository
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) Create(ctx context.Context, category *entities.DocumentCategory) error {
	args := m.Called(ctx, category)
	if args.Error(0) == nil {
		category.ID = 1
		category.CreatedAt = time.Now()
		category.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockCategoryRepository) Update(ctx context.Context, category *entities.DocumentCategory) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCategoryRepository) GetByID(ctx context.Context, id int64) (*entities.DocumentCategory, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentCategory), args.Error(1)
}

func (m *MockCategoryRepository) GetAll(ctx context.Context) ([]*entities.DocumentCategory, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentCategory), args.Error(1)
}

func (m *MockCategoryRepository) GetByParentID(ctx context.Context, parentID *int64) ([]*entities.DocumentCategory, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentCategory), args.Error(1)
}

func (m *MockCategoryRepository) GetTree(ctx context.Context) ([]*entities.CategoryTreeNode, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.CategoryTreeNode), args.Error(1)
}

func (m *MockCategoryRepository) GetChildren(ctx context.Context, parentID int64) ([]*entities.DocumentCategory, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentCategory), args.Error(1)
}

func (m *MockCategoryRepository) GetAncestors(ctx context.Context, id int64) ([]*entities.DocumentCategory, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentCategory), args.Error(1)
}

func (m *MockCategoryRepository) HasChildren(ctx context.Context, id int64) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockCategoryRepository) GetDocumentCount(ctx context.Context, id int64, includeSubcategories bool) (int64, error) {
	args := m.Called(ctx, id, includeSubcategories)
	return args.Get(0).(int64), args.Error(1)
}

func TestCategoryUseCase_Create(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	usecase := NewCategoryUseCase(mockRepo)
	ctx := context.Background()

	t.Run("create category without parent", func(t *testing.T) {
		input := dto.CreateCategoryInput{
			Name: "Test Category",
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.DocumentCategory")).Return(nil).Once()
		mockRepo.On("HasChildren", ctx, int64(1)).Return(false, nil).Once()
		mockRepo.On("GetDocumentCount", ctx, int64(1), false).Return(int64(0), nil).Once()

		result, err := usecase.Create(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Category", result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("create category with parent", func(t *testing.T) {
		parentID := int64(1)
		input := dto.CreateCategoryInput{
			Name:     "Child Category",
			ParentID: &parentID,
		}

		parent := &entities.DocumentCategory{ID: 1, Name: "Parent"}
		mockRepo.On("GetByID", ctx, int64(1)).Return(parent, nil).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.DocumentCategory")).Return(nil).Once()
		mockRepo.On("HasChildren", ctx, int64(1)).Return(false, nil).Once()
		mockRepo.On("GetDocumentCount", ctx, int64(1), false).Return(int64(0), nil).Once()

		result, err := usecase.Create(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Child Category", result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("create category with non-existent parent", func(t *testing.T) {
		parentID := int64(999)
		input := dto.CreateCategoryInput{
			Name:     "Child Category",
			ParentID: &parentID,
		}

		mockRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		result, err := usecase.Create(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "родительская категория не найдена")
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryUseCase_Update(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	usecase := NewCategoryUseCase(mockRepo)
	ctx := context.Background()

	t.Run("update category name", func(t *testing.T) {
		existing := &entities.DocumentCategory{
			ID:   1,
			Name: "Old Name",
		}
		newName := "New Name"
		input := dto.UpdateCategoryInput{
			Name: &newName,
		}

		mockRepo.On("GetByID", ctx, int64(1)).Return(existing, nil).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entities.DocumentCategory")).Return(nil).Once()
		mockRepo.On("HasChildren", ctx, int64(1)).Return(false, nil).Once()
		mockRepo.On("GetDocumentCount", ctx, int64(1), false).Return(int64(5), nil).Once()

		result, err := usecase.Update(ctx, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "New Name", result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("update non-existent category", func(t *testing.T) {
		newName := "New Name"
		input := dto.UpdateCategoryInput{
			Name: &newName,
		}

		mockRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		result, err := usecase.Update(ctx, 999, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("prevent circular reference", func(t *testing.T) {
		existing := &entities.DocumentCategory{ID: 1, Name: "Category"}
		selfID := int64(1)
		input := dto.UpdateCategoryInput{
			ParentID: &selfID,
		}

		mockRepo.On("GetByID", ctx, int64(1)).Return(existing, nil).Once()

		result, err := usecase.Update(ctx, 1, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "не может быть родителем самой себя")
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryUseCase_Delete(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	usecase := NewCategoryUseCase(mockRepo)
	ctx := context.Background()

	t.Run("delete existing category", func(t *testing.T) {
		existing := &entities.DocumentCategory{ID: 1, Name: "Test"}
		mockRepo.On("GetByID", ctx, int64(1)).Return(existing, nil).Once()
		mockRepo.On("Delete", ctx, int64(1)).Return(nil).Once()

		err := usecase.Delete(ctx, 1)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("delete non-existent category", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		err := usecase.Delete(ctx, 999)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryUseCase_GetTree(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	usecase := NewCategoryUseCase(mockRepo)
	ctx := context.Background()

	t.Run("get category tree", func(t *testing.T) {
		tree := []*entities.CategoryTreeNode{
			{
				ID:            1,
				Name:          "Root",
				DocumentCount: 5,
				Children: []*entities.CategoryTreeNode{
					{ID: 2, Name: "Child", DocumentCount: 3, Children: []*entities.CategoryTreeNode{}},
				},
			},
		}
		mockRepo.On("GetTree", ctx).Return(tree, nil).Once()

		result, err := usecase.GetTree(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "Root", result[0].Name)
		assert.Len(t, result[0].Children, 1)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryUseCase_GetWithBreadcrumb(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	usecase := NewCategoryUseCase(mockRepo)
	ctx := context.Background()

	t.Run("get category with breadcrumb", func(t *testing.T) {
		category := &entities.DocumentCategory{ID: 3, Name: "Grandchild"}
		ancestors := []*entities.DocumentCategory{
			{ID: 1, Name: "Root"},
			{ID: 2, Name: "Child"},
		}

		mockRepo.On("GetByID", ctx, int64(3)).Return(category, nil).Once()
		mockRepo.On("HasChildren", ctx, int64(3)).Return(false, nil).Once()
		mockRepo.On("GetDocumentCount", ctx, int64(3), false).Return(int64(2), nil).Once()
		mockRepo.On("GetAncestors", ctx, int64(3)).Return(ancestors, nil).Once()

		result, err := usecase.GetWithBreadcrumb(ctx, 3)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Grandchild", result.Category.Name)
		assert.Len(t, result.Breadcrumbs, 2)
		assert.Equal(t, "Root", result.Breadcrumbs[0].Name)
		mockRepo.AssertExpectations(t)
	})
}
