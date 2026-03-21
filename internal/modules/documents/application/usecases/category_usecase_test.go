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

const testNewCategoryName = "New Name"

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
	usecase := NewCategoryUseCase(mockRepo, nil)
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
	usecase := NewCategoryUseCase(mockRepo, nil)
	ctx := context.Background()

	t.Run("update category name", func(t *testing.T) {
		existing := &entities.DocumentCategory{
			ID:   1,
			Name: "Old Name",
		}
		newName := testNewCategoryName
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
		assert.Equal(t, testNewCategoryName, result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("update non-existent category", func(t *testing.T) {
		newName := testNewCategoryName
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
	usecase := NewCategoryUseCase(mockRepo, nil)
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
	usecase := NewCategoryUseCase(mockRepo, nil)
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
	usecase := NewCategoryUseCase(mockRepo, nil)
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

	t.Run("breadcrumb category not found", func(t *testing.T) {
		mockRepo2 := new(MockCategoryRepository)
		uc := NewCategoryUseCase(mockRepo2, nil)

		mockRepo2.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		result, err := uc.GetWithBreadcrumb(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo2.AssertExpectations(t)
	})

	t.Run("breadcrumb ancestors error", func(t *testing.T) {
		mockRepo2 := new(MockCategoryRepository)
		uc := NewCategoryUseCase(mockRepo2, nil)

		category := &entities.DocumentCategory{ID: 3, Name: "Cat"}
		mockRepo2.On("GetByID", ctx, int64(3)).Return(category, nil).Once()
		mockRepo2.On("HasChildren", ctx, int64(3)).Return(false, nil).Once()
		mockRepo2.On("GetDocumentCount", ctx, int64(3), false).Return(int64(0), nil).Once()
		mockRepo2.On("GetAncestors", ctx, int64(3)).Return(nil, assert.AnError).Once()

		result, err := uc.GetWithBreadcrumb(ctx, 3)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo2.AssertExpectations(t)
	})
}

func TestCategoryUseCase_GetAll(t *testing.T) {
	ctx := context.Background()

	t.Run("get all categories", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		categories := []*entities.DocumentCategory{
			{ID: 1, Name: "Cat1"},
			{ID: 2, Name: "Cat2"},
		}
		mockRepo.On("GetAll", ctx).Return(categories, nil).Once()
		mockRepo.On("HasChildren", ctx, int64(1)).Return(true, nil).Once()
		mockRepo.On("GetDocumentCount", ctx, int64(1), false).Return(int64(5), nil).Once()
		mockRepo.On("HasChildren", ctx, int64(2)).Return(false, nil).Once()
		mockRepo.On("GetDocumentCount", ctx, int64(2), false).Return(int64(3), nil).Once()

		result, err := usecase.GetAll(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.True(t, result[0].HasChildren)
		assert.Equal(t, int64(5), result[0].DocumentCount)
		assert.False(t, result[1].HasChildren)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get all categories error", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		mockRepo.On("GetAll", ctx).Return(nil, assert.AnError).Once()

		result, err := usecase.GetAll(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryUseCase_GetChildren(t *testing.T) {
	ctx := context.Background()

	t.Run("get children", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		children := []*entities.DocumentCategory{
			{ID: 2, Name: "Child1"},
			{ID: 3, Name: "Child2"},
		}
		mockRepo.On("GetChildren", ctx, int64(1)).Return(children, nil).Once()
		mockRepo.On("HasChildren", ctx, int64(2)).Return(false, nil).Once()
		mockRepo.On("GetDocumentCount", ctx, int64(2), false).Return(int64(1), nil).Once()
		mockRepo.On("HasChildren", ctx, int64(3)).Return(false, nil).Once()
		mockRepo.On("GetDocumentCount", ctx, int64(3), false).Return(int64(2), nil).Once()

		result, err := usecase.GetChildren(ctx, 1)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get children error", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		mockRepo.On("GetChildren", ctx, int64(1)).Return(nil, assert.AnError).Once()

		result, err := usecase.GetChildren(ctx, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryUseCase_GetRootCategories(t *testing.T) {
	ctx := context.Background()

	t.Run("get root categories", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		roots := []*entities.DocumentCategory{
			{ID: 1, Name: "Root1"},
		}
		mockRepo.On("GetByParentID", ctx, (*int64)(nil)).Return(roots, nil).Once()
		mockRepo.On("HasChildren", ctx, int64(1)).Return(true, nil).Once()
		mockRepo.On("GetDocumentCount", ctx, int64(1), false).Return(int64(10), nil).Once()

		result, err := usecase.GetRootCategories(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.True(t, result[0].HasChildren)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get root categories error", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		mockRepo.On("GetByParentID", ctx, (*int64)(nil)).Return(nil, assert.AnError).Once()

		result, err := usecase.GetRootCategories(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryUseCase_GetDocumentCount(t *testing.T) {
	ctx := context.Background()

	t.Run("get document count", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		mockRepo.On("GetDocumentCount", ctx, int64(1), true).Return(int64(15), nil).Once()

		count, err := usecase.GetDocumentCount(ctx, 1, true)

		assert.NoError(t, err)
		assert.Equal(t, int64(15), count)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get document count without subcategories", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		mockRepo.On("GetDocumentCount", ctx, int64(1), false).Return(int64(5), nil).Once()

		count, err := usecase.GetDocumentCount(ctx, 1, false)

		assert.NoError(t, err)
		assert.Equal(t, int64(5), count)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryUseCase_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("get by id with parent", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		parentID := int64(1)
		category := &entities.DocumentCategory{ID: 2, Name: "Child", ParentID: &parentID}
		parent := &entities.DocumentCategory{ID: 1, Name: "Parent"}

		mockRepo.On("GetByID", ctx, int64(2)).Return(category, nil).Once()
		mockRepo.On("HasChildren", ctx, int64(2)).Return(false, nil).Once()
		mockRepo.On("GetDocumentCount", ctx, int64(2), false).Return(int64(3), nil).Once()
		mockRepo.On("GetByID", ctx, int64(1)).Return(parent, nil).Once()

		result, err := usecase.GetByID(ctx, 2)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Child", result.Name)
		assert.NotNil(t, result.ParentName)
		assert.Equal(t, "Parent", *result.ParentName)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get by id not found", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		mockRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		result, err := usecase.GetByID(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryUseCase_Update_WithParentChange(t *testing.T) {
	ctx := context.Background()

	t.Run("update with valid parent change", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		existing := &entities.DocumentCategory{ID: 2, Name: "Category"}
		newParentID := int64(3)
		newParent := &entities.DocumentCategory{ID: 3, Name: "New Parent"}

		mockRepo.On("GetByID", ctx, int64(2)).Return(existing, nil).Once()
		mockRepo.On("GetAncestors", ctx, int64(3)).Return([]*entities.DocumentCategory{
			{ID: 4, Name: "Grandparent"},
		}, nil).Once()
		mockRepo.On("GetByID", ctx, int64(3)).Return(newParent, nil).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entities.DocumentCategory")).Return(nil).Once()
		mockRepo.On("HasChildren", ctx, int64(2)).Return(false, nil).Once()
		mockRepo.On("GetDocumentCount", ctx, int64(2), false).Return(int64(0), nil).Once()
		// Getting parent name
		mockRepo.On("GetByID", ctx, int64(3)).Return(newParent, nil).Once()

		input := dto.UpdateCategoryInput{ParentID: &newParentID}

		result, err := usecase.Update(ctx, 2, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("prevent moving category to its own descendant", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		existing := &entities.DocumentCategory{ID: 1, Name: "Parent"}
		newParentID := int64(3)

		mockRepo.On("GetByID", ctx, int64(1)).Return(existing, nil).Once()
		mockRepo.On("GetAncestors", ctx, int64(3)).Return([]*entities.DocumentCategory{
			{ID: 1, Name: "Parent"}, // This is the same as the category being moved
		}, nil).Once()

		input := dto.UpdateCategoryInput{ParentID: &newParentID}

		result, err := usecase.Update(ctx, 1, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "нельзя переместить категорию в её подкатегорию")
		mockRepo.AssertExpectations(t)
	})

	t.Run("update with non-existent new parent", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		existing := &entities.DocumentCategory{ID: 2, Name: "Category"}
		newParentID := int64(999)

		mockRepo.On("GetByID", ctx, int64(2)).Return(existing, nil).Once()
		mockRepo.On("GetAncestors", ctx, int64(999)).Return([]*entities.DocumentCategory{}, nil).Once()
		mockRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		input := dto.UpdateCategoryInput{ParentID: &newParentID}

		result, err := usecase.Update(ctx, 2, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "родительская категория не найдена")
		mockRepo.AssertExpectations(t)
	})

	t.Run("update with description", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		existing := &entities.DocumentCategory{ID: 1, Name: "Cat"}
		desc := "New description"

		mockRepo.On("GetByID", ctx, int64(1)).Return(existing, nil).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entities.DocumentCategory")).Return(nil).Once()
		mockRepo.On("HasChildren", ctx, int64(1)).Return(false, nil).Once()
		mockRepo.On("GetDocumentCount", ctx, int64(1), false).Return(int64(0), nil).Once()

		input := dto.UpdateCategoryInput{Description: &desc}

		result, err := usecase.Update(ctx, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("update repo error", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		existing := &entities.DocumentCategory{ID: 1, Name: "Cat"}
		newName := testNewCategoryName

		mockRepo.On("GetByID", ctx, int64(1)).Return(existing, nil).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entities.DocumentCategory")).Return(assert.AnError).Once()

		input := dto.UpdateCategoryInput{Name: &newName}

		result, err := usecase.Update(ctx, 1, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryUseCase_Create_ParentExistsButNil(t *testing.T) {
	ctx := context.Background()

	t.Run("create with parent returning nil entity", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		parentID := int64(1)
		input := dto.CreateCategoryInput{
			Name:     "Child",
			ParentID: &parentID,
		}

		// GetByID returns nil entity with no error
		mockRepo.On("GetByID", ctx, int64(1)).Return((*entities.DocumentCategory)(nil), nil).Once()

		result, err := usecase.Create(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "родительская категория не найдена")
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryUseCase_Delete_Error(t *testing.T) {
	ctx := context.Background()

	t.Run("delete repo error", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		existing := &entities.DocumentCategory{ID: 1, Name: "Test"}
		mockRepo.On("GetByID", ctx, int64(1)).Return(existing, nil).Once()
		mockRepo.On("Delete", ctx, int64(1)).Return(assert.AnError).Once()

		err := usecase.Delete(ctx, 1)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryUseCase_Create_RepoError(t *testing.T) {
	ctx := context.Background()

	t.Run("create repo error", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		input := dto.CreateCategoryInput{Name: "Test"}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.DocumentCategory")).Return(assert.AnError).Once()

		result, err := usecase.Create(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryUseCase_GetTree_Error(t *testing.T) {
	ctx := context.Background()

	t.Run("get tree error", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		mockRepo.On("GetTree", ctx).Return(nil, assert.AnError).Once()

		result, err := usecase.GetTree(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryUseCase_Create_WithDescription(t *testing.T) {
	ctx := context.Background()

	t.Run("create with description", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		usecase := NewCategoryUseCase(mockRepo, nil)

		desc := "A description"
		input := dto.CreateCategoryInput{
			Name:        "Category",
			Description: &desc,
		}

		mockRepo.On("Create", ctx, mock.MatchedBy(func(c *entities.DocumentCategory) bool {
			return c.Description != nil && *c.Description == "A description"
		})).Return(nil).Once()
		mockRepo.On("HasChildren", ctx, int64(1)).Return(false, nil).Once()
		mockRepo.On("GetDocumentCount", ctx, int64(1), false).Return(int64(0), nil).Once()

		result, err := usecase.Create(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})
}
