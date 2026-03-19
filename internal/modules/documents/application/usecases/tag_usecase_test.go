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

// MockTagRepository is a mock implementation of DocumentTagRepository
type MockTagRepository struct {
	mock.Mock
}

func (m *MockTagRepository) Create(ctx context.Context, tag *entities.DocumentTag) error {
	args := m.Called(ctx, tag)
	if args.Error(0) == nil {
		tag.ID = 1
		tag.CreatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockTagRepository) Update(ctx context.Context, tag *entities.DocumentTag) error {
	args := m.Called(ctx, tag)
	return args.Error(0)
}

func (m *MockTagRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTagRepository) GetByID(ctx context.Context, id int64) (*entities.DocumentTag, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentTag), args.Error(1)
}

func (m *MockTagRepository) GetByName(ctx context.Context, name string) (*entities.DocumentTag, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.DocumentTag), args.Error(1)
}

func (m *MockTagRepository) GetAll(ctx context.Context) ([]*entities.DocumentTag, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentTag), args.Error(1)
}

func (m *MockTagRepository) Search(ctx context.Context, query string, limit int) ([]*entities.DocumentTag, error) {
	args := m.Called(ctx, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentTag), args.Error(1)
}

func (m *MockTagRepository) AddTagToDocument(ctx context.Context, documentID, tagID int64) error {
	args := m.Called(ctx, documentID, tagID)
	return args.Error(0)
}

func (m *MockTagRepository) RemoveTagFromDocument(ctx context.Context, documentID, tagID int64) error {
	args := m.Called(ctx, documentID, tagID)
	return args.Error(0)
}

func (m *MockTagRepository) GetTagsByDocumentID(ctx context.Context, documentID int64) ([]*entities.DocumentTag, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.DocumentTag), args.Error(1)
}

func (m *MockTagRepository) GetDocumentsByTagID(ctx context.Context, tagID int64, limit, offset int) ([]int64, int64, error) {
	args := m.Called(ctx, tagID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockTagRepository) SetDocumentTags(ctx context.Context, documentID int64, tagIDs []int64) error {
	args := m.Called(ctx, documentID, tagIDs)
	return args.Error(0)
}

func (m *MockTagRepository) GetTagUsageCount(ctx context.Context, tagID int64) (int64, error) {
	args := m.Called(ctx, tagID)
	return args.Get(0).(int64), args.Error(1)
}

func TestTagUseCase_Create(t *testing.T) {
	mockTagRepo := new(MockTagRepository)
	mockDocRepo := new(MockDocumentRepository)
	usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)
	ctx := context.Background()

	t.Run("create tag", func(t *testing.T) {
		color := "#FF0000"
		input := dto.CreateTagInput{
			Name:  "Important",
			Color: &color,
		}

		mockTagRepo.On("GetByName", ctx, "Important").Return(nil, assert.AnError).Once()
		mockTagRepo.On("Create", ctx, mock.AnythingOfType("*entities.DocumentTag")).Return(nil).Once()

		result, err := usecase.Create(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Important", result.Name)
		assert.Equal(t, int64(0), result.UsageCount)
		mockTagRepo.AssertExpectations(t)
	})

	t.Run("create duplicate tag", func(t *testing.T) {
		input := dto.CreateTagInput{
			Name: "Existing",
		}

		existing := &entities.DocumentTag{ID: 1, Name: "Existing"}
		mockTagRepo.On("GetByName", ctx, "Existing").Return(existing, nil).Once()

		result, err := usecase.Create(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "уже существует")
		mockTagRepo.AssertExpectations(t)
	})
}

func TestTagUseCase_Update(t *testing.T) {
	mockTagRepo := new(MockTagRepository)
	mockDocRepo := new(MockDocumentRepository)
	usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)
	ctx := context.Background()

	t.Run("update tag name", func(t *testing.T) {
		existing := &entities.DocumentTag{ID: 1, Name: "Old"}
		newName := "New"
		input := dto.UpdateTagInput{
			Name: &newName,
		}

		mockTagRepo.On("GetByID", ctx, int64(1)).Return(existing, nil).Once()
		mockTagRepo.On("GetByName", ctx, "New").Return(nil, assert.AnError).Once()
		mockTagRepo.On("Update", ctx, mock.AnythingOfType("*entities.DocumentTag")).Return(nil).Once()
		mockTagRepo.On("GetTagUsageCount", ctx, int64(1)).Return(int64(5), nil).Once()

		result, err := usecase.Update(ctx, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "New", result.Name)
		assert.Equal(t, int64(5), result.UsageCount)
		mockTagRepo.AssertExpectations(t)
	})

	t.Run("update non-existent tag", func(t *testing.T) {
		newName := "New"
		input := dto.UpdateTagInput{Name: &newName}

		mockTagRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		result, err := usecase.Update(ctx, 999, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockTagRepo.AssertExpectations(t)
	})
}

func TestTagUseCase_Delete(t *testing.T) {
	mockTagRepo := new(MockTagRepository)
	mockDocRepo := new(MockDocumentRepository)
	usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)
	ctx := context.Background()

	t.Run("delete existing tag", func(t *testing.T) {
		existing := &entities.DocumentTag{ID: 1, Name: "Test"}
		mockTagRepo.On("GetByID", ctx, int64(1)).Return(existing, nil).Once()
		mockTagRepo.On("Delete", ctx, int64(1)).Return(nil).Once()

		err := usecase.Delete(ctx, 1)

		assert.NoError(t, err)
		mockTagRepo.AssertExpectations(t)
	})

	t.Run("delete non-existent tag", func(t *testing.T) {
		mockTagRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		err := usecase.Delete(ctx, 999)

		assert.Error(t, err)
		mockTagRepo.AssertExpectations(t)
	})
}

func TestTagUseCase_GetAll(t *testing.T) {
	mockTagRepo := new(MockTagRepository)
	mockDocRepo := new(MockDocumentRepository)
	usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)
	ctx := context.Background()

	t.Run("get all tags", func(t *testing.T) {
		tags := []*entities.DocumentTag{
			{ID: 1, Name: "Tag1"},
			{ID: 2, Name: "Tag2"},
		}

		mockTagRepo.On("GetAll", ctx).Return(tags, nil).Once()
		mockTagRepo.On("GetTagUsageCount", ctx, int64(1)).Return(int64(5), nil).Once()
		mockTagRepo.On("GetTagUsageCount", ctx, int64(2)).Return(int64(3), nil).Once()

		result, err := usecase.GetAll(ctx)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(5), result[0].UsageCount)
		assert.Equal(t, int64(3), result[1].UsageCount)
		mockTagRepo.AssertExpectations(t)
	})
}

func TestTagUseCase_Search(t *testing.T) {
	mockTagRepo := new(MockTagRepository)
	mockDocRepo := new(MockDocumentRepository)
	usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)
	ctx := context.Background()

	t.Run("search tags", func(t *testing.T) {
		tags := []*entities.DocumentTag{
			{ID: 1, Name: "Important"},
			{ID: 2, Name: "Imp"},
		}

		mockTagRepo.On("Search", ctx, "imp", 10).Return(tags, nil).Once()
		mockTagRepo.On("GetTagUsageCount", ctx, int64(1)).Return(int64(5), nil).Once()
		mockTagRepo.On("GetTagUsageCount", ctx, int64(2)).Return(int64(2), nil).Once()

		result, err := usecase.Search(ctx, "imp", 10)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockTagRepo.AssertExpectations(t)
	})
}

func TestTagUseCase_AddTagToDocument(t *testing.T) {
	mockTagRepo := new(MockTagRepository)
	mockDocRepo := new(MockDocumentRepository)
	usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)
	ctx := context.Background()

	t.Run("add tag to document", func(t *testing.T) {
		doc := &entities.Document{ID: 1, Title: "Doc"}
		tag := &entities.DocumentTag{ID: 1, Name: "Tag"}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil).Once()
		mockTagRepo.On("GetByID", ctx, int64(1)).Return(tag, nil).Once()
		mockTagRepo.On("AddTagToDocument", ctx, int64(1), int64(1)).Return(nil).Once()

		err := usecase.AddTagToDocument(ctx, 1, 1)

		assert.NoError(t, err)
		mockTagRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("add tag to non-existent document", func(t *testing.T) {
		mockDocRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		err := usecase.AddTagToDocument(ctx, 999, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "документ не найден")
		mockDocRepo.AssertExpectations(t)
	})
}

func TestTagUseCase_GetDocumentTags(t *testing.T) {
	mockTagRepo := new(MockTagRepository)
	mockDocRepo := new(MockDocumentRepository)
	usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)
	ctx := context.Background()

	t.Run("get document tags", func(t *testing.T) {
		doc := &entities.Document{ID: 1, Title: "Doc"}
		tags := []*entities.DocumentTag{
			{ID: 1, Name: "Tag1"},
			{ID: 2, Name: "Tag2"},
		}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil).Once()
		mockTagRepo.On("GetTagsByDocumentID", ctx, int64(1)).Return(tags, nil).Once()
		mockTagRepo.On("GetTagUsageCount", ctx, int64(1)).Return(int64(5), nil).Once()
		mockTagRepo.On("GetTagUsageCount", ctx, int64(2)).Return(int64(3), nil).Once()

		result, err := usecase.GetDocumentTags(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result.DocumentID)
		assert.Len(t, result.Tags, 2)
		mockTagRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestTagUseCase_SetDocumentTags(t *testing.T) {
	mockTagRepo := new(MockTagRepository)
	mockDocRepo := new(MockDocumentRepository)
	usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)
	ctx := context.Background()

	t.Run("set document tags", func(t *testing.T) {
		doc := &entities.Document{ID: 1, Title: "Doc"}
		tag1 := &entities.DocumentTag{ID: 1, Name: "Tag1"}
		tag2 := &entities.DocumentTag{ID: 2, Name: "Tag2"}
		resultTags := []*entities.DocumentTag{tag1, tag2}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil).Times(2)
		mockTagRepo.On("GetByID", ctx, int64(1)).Return(tag1, nil).Once()
		mockTagRepo.On("GetByID", ctx, int64(2)).Return(tag2, nil).Once()
		mockTagRepo.On("SetDocumentTags", ctx, int64(1), []int64{1, 2}).Return(nil).Once()
		mockTagRepo.On("GetTagsByDocumentID", ctx, int64(1)).Return(resultTags, nil).Once()
		mockTagRepo.On("GetTagUsageCount", ctx, int64(1)).Return(int64(5), nil).Once()
		mockTagRepo.On("GetTagUsageCount", ctx, int64(2)).Return(int64(3), nil).Once()

		result, err := usecase.SetDocumentTags(ctx, 1, []int64{1, 2})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Tags, 2)
		mockTagRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestTagUseCase_GetDocumentsByTag(t *testing.T) {
	mockTagRepo := new(MockTagRepository)
	mockDocRepo := new(MockDocumentRepository)
	usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)
	ctx := context.Background()

	t.Run("get documents by tag", func(t *testing.T) {
		tag := &entities.DocumentTag{ID: 1, Name: "Important"}
		docIDs := []int64{1, 2, 3}

		mockTagRepo.On("GetByID", ctx, int64(1)).Return(tag, nil).Once()
		mockTagRepo.On("GetDocumentsByTagID", ctx, int64(1), 20, 0).Return(docIDs, int64(3), nil).Once()

		result, err := usecase.GetDocumentsByTag(ctx, 1, 1, 20)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Important", result.Tag.Name)
		assert.Len(t, result.DocumentIDs, 3)
		assert.Equal(t, int64(3), result.Total)
		mockTagRepo.AssertExpectations(t)
	})

	t.Run("get documents by tag with default pagination", func(t *testing.T) {
		mockTagRepo2 := new(MockTagRepository)
		mockDocRepo2 := new(MockDocumentRepository)
		uc := NewTagUseCase(mockTagRepo2, mockDocRepo2, nil)

		tag := &entities.DocumentTag{ID: 1, Name: "Tag"}

		mockTagRepo2.On("GetByID", ctx, int64(1)).Return(tag, nil).Once()
		mockTagRepo2.On("GetDocumentsByTagID", ctx, int64(1), 20, 0).Return([]int64{}, int64(0), nil).Once()

		result, err := uc.GetDocumentsByTag(ctx, 1, 0, 0)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.PageSize)
		mockTagRepo2.AssertExpectations(t)
	})

	t.Run("get documents by tag - tag not found", func(t *testing.T) {
		mockTagRepo2 := new(MockTagRepository)
		mockDocRepo2 := new(MockDocumentRepository)
		uc := NewTagUseCase(mockTagRepo2, mockDocRepo2, nil)

		mockTagRepo2.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		result, err := uc.GetDocumentsByTag(ctx, 999, 1, 20)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockTagRepo2.AssertExpectations(t)
	})

	t.Run("get documents by tag - repo error", func(t *testing.T) {
		mockTagRepo2 := new(MockTagRepository)
		mockDocRepo2 := new(MockDocumentRepository)
		uc := NewTagUseCase(mockTagRepo2, mockDocRepo2, nil)

		tag := &entities.DocumentTag{ID: 1, Name: "Tag"}
		mockTagRepo2.On("GetByID", ctx, int64(1)).Return(tag, nil).Once()
		mockTagRepo2.On("GetDocumentsByTagID", ctx, int64(1), 20, 0).Return(nil, int64(0), assert.AnError).Once()

		result, err := uc.GetDocumentsByTag(ctx, 1, 1, 20)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockTagRepo2.AssertExpectations(t)
	})
}

func TestTagUseCase_GetByID(t *testing.T) {
	mockTagRepo := new(MockTagRepository)
	mockDocRepo := new(MockDocumentRepository)
	usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)
	ctx := context.Background()

	t.Run("get tag by id", func(t *testing.T) {
		tag := &entities.DocumentTag{ID: 1, Name: "Important"}
		mockTagRepo.On("GetByID", ctx, int64(1)).Return(tag, nil).Once()
		mockTagRepo.On("GetTagUsageCount", ctx, int64(1)).Return(int64(10), nil).Once()

		result, err := usecase.GetByID(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Important", result.Name)
		assert.Equal(t, int64(10), result.UsageCount)
		mockTagRepo.AssertExpectations(t)
	})

	t.Run("get tag by id not found", func(t *testing.T) {
		mockTagRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		result, err := usecase.GetByID(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "тег не найден")
		mockTagRepo.AssertExpectations(t)
	})
}

func TestTagUseCase_Update_DuplicateName(t *testing.T) {
	ctx := context.Background()

	t.Run("update tag with duplicate name", func(t *testing.T) {
		mockTagRepo := new(MockTagRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)

		existing := &entities.DocumentTag{ID: 1, Name: "Old"}
		newName := "Existing"

		mockTagRepo.On("GetByID", ctx, int64(1)).Return(existing, nil).Once()
		mockTagRepo.On("GetByName", ctx, "Existing").Return(&entities.DocumentTag{ID: 2, Name: "Existing"}, nil).Once()

		input := dto.UpdateTagInput{Name: &newName}

		result, err := usecase.Update(ctx, 1, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "уже существует")
		mockTagRepo.AssertExpectations(t)
	})

	t.Run("update tag - same name same id allowed", func(t *testing.T) {
		mockTagRepo := new(MockTagRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)

		existing := &entities.DocumentTag{ID: 1, Name: "Same"}
		newName := "Same"

		mockTagRepo.On("GetByID", ctx, int64(1)).Return(existing, nil).Once()
		mockTagRepo.On("GetByName", ctx, "Same").Return(&entities.DocumentTag{ID: 1, Name: "Same"}, nil).Once()
		mockTagRepo.On("Update", ctx, mock.AnythingOfType("*entities.DocumentTag")).Return(nil).Once()
		mockTagRepo.On("GetTagUsageCount", ctx, int64(1)).Return(int64(0), nil).Once()

		input := dto.UpdateTagInput{Name: &newName}

		result, err := usecase.Update(ctx, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockTagRepo.AssertExpectations(t)
	})

	t.Run("update tag color only", func(t *testing.T) {
		mockTagRepo := new(MockTagRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)

		existing := &entities.DocumentTag{ID: 1, Name: "Tag"}
		newColor := "#00FF00"

		mockTagRepo.On("GetByID", ctx, int64(1)).Return(existing, nil).Once()
		mockTagRepo.On("Update", ctx, mock.AnythingOfType("*entities.DocumentTag")).Return(nil).Once()
		mockTagRepo.On("GetTagUsageCount", ctx, int64(1)).Return(int64(0), nil).Once()

		input := dto.UpdateTagInput{Color: &newColor}

		result, err := usecase.Update(ctx, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockTagRepo.AssertExpectations(t)
	})

	t.Run("update tag repo error", func(t *testing.T) {
		mockTagRepo := new(MockTagRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)

		existing := &entities.DocumentTag{ID: 1, Name: "Tag"}
		newColor := "#FF0000"

		mockTagRepo.On("GetByID", ctx, int64(1)).Return(existing, nil).Once()
		mockTagRepo.On("Update", ctx, mock.AnythingOfType("*entities.DocumentTag")).Return(assert.AnError).Once()

		input := dto.UpdateTagInput{Color: &newColor}

		result, err := usecase.Update(ctx, 1, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockTagRepo.AssertExpectations(t)
	})
}

func TestTagUseCase_AddTagToDocument_TagNotFound(t *testing.T) {
	ctx := context.Background()

	t.Run("add non-existent tag to document", func(t *testing.T) {
		mockTagRepo := new(MockTagRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)

		doc := &entities.Document{ID: 1, Title: "Doc"}
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil).Once()
		mockTagRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		err := usecase.AddTagToDocument(ctx, 1, 999)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "тег не найден")
		mockDocRepo.AssertExpectations(t)
		mockTagRepo.AssertExpectations(t)
	})

	t.Run("add tag to document repo error", func(t *testing.T) {
		mockTagRepo := new(MockTagRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)

		doc := &entities.Document{ID: 1, Title: "Doc"}
		tag := &entities.DocumentTag{ID: 1, Name: "Tag"}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil).Once()
		mockTagRepo.On("GetByID", ctx, int64(1)).Return(tag, nil).Once()
		mockTagRepo.On("AddTagToDocument", ctx, int64(1), int64(1)).Return(assert.AnError).Once()

		err := usecase.AddTagToDocument(ctx, 1, 1)

		assert.Error(t, err)
		mockDocRepo.AssertExpectations(t)
		mockTagRepo.AssertExpectations(t)
	})
}

func TestTagUseCase_RemoveTagFromDocument(t *testing.T) {
	ctx := context.Background()

	t.Run("remove tag from document success", func(t *testing.T) {
		mockTagRepo := new(MockTagRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)

		mockTagRepo.On("RemoveTagFromDocument", ctx, int64(1), int64(2)).Return(nil).Once()

		err := usecase.RemoveTagFromDocument(ctx, 1, 2)

		assert.NoError(t, err)
		mockTagRepo.AssertExpectations(t)
	})

	t.Run("remove tag from document error", func(t *testing.T) {
		mockTagRepo := new(MockTagRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)

		mockTagRepo.On("RemoveTagFromDocument", ctx, int64(1), int64(2)).Return(assert.AnError).Once()

		err := usecase.RemoveTagFromDocument(ctx, 1, 2)

		assert.Error(t, err)
		mockTagRepo.AssertExpectations(t)
	})
}

func TestTagUseCase_GetDocumentTags_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("document not found", func(t *testing.T) {
		mockTagRepo := new(MockTagRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)

		mockDocRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		result, err := usecase.GetDocumentTags(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "документ не найден")
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("get tags repo error", func(t *testing.T) {
		mockTagRepo := new(MockTagRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)

		doc := &entities.Document{ID: 1, Title: "Doc"}
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil).Once()
		mockTagRepo.On("GetTagsByDocumentID", ctx, int64(1)).Return(nil, assert.AnError).Once()

		result, err := usecase.GetDocumentTags(ctx, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockTagRepo.AssertExpectations(t)
	})
}

func TestTagUseCase_SetDocumentTags_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("document not found", func(t *testing.T) {
		mockTagRepo := new(MockTagRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)

		mockDocRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		result, err := usecase.SetDocumentTags(ctx, 999, []int64{1})

		assert.Error(t, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("invalid tag id", func(t *testing.T) {
		mockTagRepo := new(MockTagRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)

		doc := &entities.Document{ID: 1, Title: "Doc"}
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil).Once()
		mockTagRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		result, err := usecase.SetDocumentTags(ctx, 1, []int64{999})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "тег с ID 999 не найден")
		mockDocRepo.AssertExpectations(t)
		mockTagRepo.AssertExpectations(t)
	})

	t.Run("set tags repo error", func(t *testing.T) {
		mockTagRepo := new(MockTagRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)

		doc := &entities.Document{ID: 1, Title: "Doc"}
		tag := &entities.DocumentTag{ID: 1, Name: "Tag"}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil).Once()
		mockTagRepo.On("GetByID", ctx, int64(1)).Return(tag, nil).Once()
		mockTagRepo.On("SetDocumentTags", ctx, int64(1), []int64{1}).Return(assert.AnError).Once()

		result, err := usecase.SetDocumentTags(ctx, 1, []int64{1})

		assert.Error(t, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockTagRepo.AssertExpectations(t)
	})
}

func TestTagUseCase_GetAll_Error(t *testing.T) {
	ctx := context.Background()

	t.Run("get all tags error", func(t *testing.T) {
		mockTagRepo := new(MockTagRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)

		mockTagRepo.On("GetAll", ctx).Return(nil, assert.AnError).Once()

		result, err := usecase.GetAll(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockTagRepo.AssertExpectations(t)
	})
}

func TestTagUseCase_Search_Error(t *testing.T) {
	ctx := context.Background()

	t.Run("search tags error", func(t *testing.T) {
		mockTagRepo := new(MockTagRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)

		mockTagRepo.On("Search", ctx, "test", 10).Return(nil, assert.AnError).Once()

		result, err := usecase.Search(ctx, "test", 10)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockTagRepo.AssertExpectations(t)
	})
}

func TestTagUseCase_Create_RepoError(t *testing.T) {
	ctx := context.Background()

	t.Run("create tag repo error", func(t *testing.T) {
		mockTagRepo := new(MockTagRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)

		mockTagRepo.On("GetByName", ctx, "New").Return(nil, assert.AnError).Once()
		mockTagRepo.On("Create", ctx, mock.AnythingOfType("*entities.DocumentTag")).Return(assert.AnError).Once()

		input := dto.CreateTagInput{Name: "New"}

		result, err := usecase.Create(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockTagRepo.AssertExpectations(t)
	})
}

func TestTagUseCase_Delete_RepoError(t *testing.T) {
	ctx := context.Background()

	t.Run("delete tag repo error", func(t *testing.T) {
		mockTagRepo := new(MockTagRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTagUseCase(mockTagRepo, mockDocRepo, nil)

		existing := &entities.DocumentTag{ID: 1, Name: "Tag"}
		mockTagRepo.On("GetByID", ctx, int64(1)).Return(existing, nil).Once()
		mockTagRepo.On("Delete", ctx, int64(1)).Return(assert.AnError).Once()

		err := usecase.Delete(ctx, 1)

		assert.Error(t, err)
		mockTagRepo.AssertExpectations(t)
	})
}
