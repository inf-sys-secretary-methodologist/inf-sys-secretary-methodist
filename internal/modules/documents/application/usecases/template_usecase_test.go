package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

const testTemplateContent = "Hello {{name}}"

// MockTemplateRepository is a mock implementation of TemplateRepository
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

func TestTemplateUseCase_GetAllTemplates(t *testing.T) {
	ctx := context.Background()

	t.Run("get all templates successfully", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		content := testTemplateContent
		types := []entities.DocumentType{
			{ID: 1, Name: "Order", Code: "order", TemplateContent: &content},
			{ID: 2, Name: "Letter", Code: "letter"},
		}

		mockTemplateRepo.On("GetAll", ctx).Return(types, nil)

		result, err := usecase.GetAllTemplates(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.Total)
		assert.Len(t, result.Templates, 1)
		assert.Equal(t, "Order", result.Templates[0].Name)
		mockTemplateRepo.AssertExpectations(t)
	})

	t.Run("get all templates error", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		mockTemplateRepo.On("GetAll", ctx).Return(nil, errors.New("database error"))

		result, err := usecase.GetAllTemplates(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get templates")
		mockTemplateRepo.AssertExpectations(t)
	})
}

func TestTemplateUseCase_GetTemplate(t *testing.T) {
	ctx := context.Background()

	t.Run("get template successfully", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		content := "Template {{var}}"
		docType := &entities.DocumentType{ID: 1, Name: "Order", Code: "order", TemplateContent: &content}

		mockTemplateRepo.On("GetByID", ctx, int64(1)).Return(docType, nil)

		result, err := usecase.GetTemplate(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Order", result.Name)
		assert.True(t, result.HasTemplate)
		mockTemplateRepo.AssertExpectations(t)
	})

	t.Run("get template not found", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		mockTemplateRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		result, err := usecase.GetTemplate(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get template")
		mockTemplateRepo.AssertExpectations(t)
	})
}

func TestTemplateUseCase_PreviewTemplate(t *testing.T) {
	ctx := context.Background()

	t.Run("preview template successfully", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		content := "Dear {{name}}, your order {{order_id}} is ready."
		docType := &entities.DocumentType{
			ID:              1,
			Name:            "Order",
			TemplateContent: &content,
			TemplateVariables: []entities.TemplateVariable{
				{Name: "name", Label: "Name", Required: true},
				{Name: "order_id", Label: "Order ID", Required: true},
			},
		}

		mockTemplateRepo.On("GetByID", ctx, int64(1)).Return(docType, nil)

		variables := map[string]string{
			"name":     "John",
			"order_id": "12345",
		}

		result, err := usecase.PreviewTemplate(ctx, 1, variables)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Dear John, your order 12345 is ready.", result.Content)
		mockTemplateRepo.AssertExpectations(t)
	})

	t.Run("preview template with missing required variable", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		content := "Dear {{name}}"
		docType := &entities.DocumentType{
			ID:              1,
			Name:            "Order",
			TemplateContent: &content,
			TemplateVariables: []entities.TemplateVariable{
				{Name: "name", Label: "Name", Required: true},
			},
		}

		mockTemplateRepo.On("GetByID", ctx, int64(1)).Return(docType, nil)

		result, err := usecase.PreviewTemplate(ctx, 1, map[string]string{})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "required variable")
		mockTemplateRepo.AssertExpectations(t)
	})

	t.Run("preview template with empty required variable", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		content := "Dear {{name}}"
		docType := &entities.DocumentType{
			ID:              1,
			Name:            "Order",
			TemplateContent: &content,
			TemplateVariables: []entities.TemplateVariable{
				{Name: "name", Label: "Name", Required: true},
			},
		}

		mockTemplateRepo.On("GetByID", ctx, int64(1)).Return(docType, nil)

		result, err := usecase.PreviewTemplate(ctx, 1, map[string]string{"name": "  "})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "required variable")
		mockTemplateRepo.AssertExpectations(t)
	})

	t.Run("preview template not found", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		mockTemplateRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		result, err := usecase.PreviewTemplate(ctx, 999, map[string]string{})

		assert.Error(t, err)
		assert.Nil(t, result)
		mockTemplateRepo.AssertExpectations(t)
	})

	t.Run("preview template with no content", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		docType := &entities.DocumentType{ID: 1, Name: "Order", TemplateContent: nil}

		mockTemplateRepo.On("GetByID", ctx, int64(1)).Return(docType, nil)

		result, err := usecase.PreviewTemplate(ctx, 1, map[string]string{})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "template not found")
		mockTemplateRepo.AssertExpectations(t)
	})

	t.Run("preview template with empty content string", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		emptyContent := ""
		docType := &entities.DocumentType{ID: 1, Name: "Order", TemplateContent: &emptyContent}

		mockTemplateRepo.On("GetByID", ctx, int64(1)).Return(docType, nil)

		result, err := usecase.PreviewTemplate(ctx, 1, map[string]string{})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "template not found")
		mockTemplateRepo.AssertExpectations(t)
	})

	t.Run("preview template with optional variables not provided", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		content := "Dear {{name}}, {{greeting}}"
		docType := &entities.DocumentType{
			ID:              1,
			Name:            "Order",
			TemplateContent: &content,
			TemplateVariables: []entities.TemplateVariable{
				{Name: "name", Label: "Name", Required: true},
				{Name: "greeting", Label: "Greeting", Required: false},
			},
		}

		mockTemplateRepo.On("GetByID", ctx, int64(1)).Return(docType, nil)

		result, err := usecase.PreviewTemplate(ctx, 1, map[string]string{"name": "John"})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Dear John, {{greeting}}", result.Content)
		mockTemplateRepo.AssertExpectations(t)
	})
}

func TestTemplateUseCase_CreateDocumentFromTemplate(t *testing.T) {
	ctx := context.Background()

	t.Run("create document from template successfully", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		content := "Dear {{name}}, subject: {{subject}}"
		docType := &entities.DocumentType{
			ID:              1,
			Name:            "Order",
			TemplateContent: &content,
			TemplateVariables: []entities.TemplateVariable{
				{Name: "name", Label: "Name", Required: true},
				{Name: "subject", Label: "Subject", Required: false},
			},
		}

		mockTemplateRepo.On("GetByID", ctx, int64(1)).Return(docType, nil)
		mockDocRepo.On("Create", ctx, mock.AnythingOfType("*entities.Document")).Return(nil)

		req := &dto.CreateFromTemplateRequest{
			Title: "Test Document",
			Variables: map[string]string{
				"name":    "John",
				"subject": "Meeting",
			},
		}

		result, err := usecase.CreateDocumentFromTemplate(ctx, 1, req, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Document", result.Title)
		assert.Equal(t, int64(1), result.DocumentTypeID)
		assert.NotNil(t, result.Content)
		assert.Equal(t, "Dear John, subject: Meeting", *result.Content)
		assert.NotNil(t, result.Subject)
		assert.Equal(t, "Meeting", *result.Subject)
		mockTemplateRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("create document from template with category", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		content := testTemplateContent
		docType := &entities.DocumentType{
			ID:              1,
			Name:            "Order",
			TemplateContent: &content,
			TemplateVariables: []entities.TemplateVariable{
				{Name: "name", Label: "Name", Required: true},
			},
		}

		mockTemplateRepo.On("GetByID", ctx, int64(1)).Return(docType, nil)
		mockDocRepo.On("Create", ctx, mock.MatchedBy(func(doc *entities.Document) bool {
			return doc.CategoryID != nil && *doc.CategoryID == int64(5)
		})).Return(nil)

		catID := int64(5)
		req := &dto.CreateFromTemplateRequest{
			Title:      "Test Document",
			Variables:  map[string]string{"name": "John"},
			CategoryID: &catID,
		}

		result, err := usecase.CreateDocumentFromTemplate(ctx, 1, req, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockTemplateRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("create document - template not found", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		mockTemplateRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		req := &dto.CreateFromTemplateRequest{Title: "Test", Variables: map[string]string{}}

		result, err := usecase.CreateDocumentFromTemplate(ctx, 999, req, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockTemplateRepo.AssertExpectations(t)
	})

	t.Run("create document - no template content", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		docType := &entities.DocumentType{ID: 1, Name: "Order", TemplateContent: nil}
		mockTemplateRepo.On("GetByID", ctx, int64(1)).Return(docType, nil)

		req := &dto.CreateFromTemplateRequest{Title: "Test", Variables: map[string]string{}}

		result, err := usecase.CreateDocumentFromTemplate(ctx, 1, req, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "template not found")
		mockTemplateRepo.AssertExpectations(t)
	})

	t.Run("create document - missing required variable", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		content := testTemplateContent
		docType := &entities.DocumentType{
			ID: 1, Name: "Order", TemplateContent: &content,
			TemplateVariables: []entities.TemplateVariable{{Name: "name", Label: "Name", Required: true}},
		}
		mockTemplateRepo.On("GetByID", ctx, int64(1)).Return(docType, nil)

		req := &dto.CreateFromTemplateRequest{Title: "Test", Variables: map[string]string{}}

		result, err := usecase.CreateDocumentFromTemplate(ctx, 1, req, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "required variable")
		mockTemplateRepo.AssertExpectations(t)
	})

	t.Run("create document - repo create fails", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		content := testTemplateContent
		docType := &entities.DocumentType{
			ID: 1, Name: "Order", TemplateContent: &content,
			TemplateVariables: []entities.TemplateVariable{{Name: "name", Label: "Name", Required: true}},
		}
		mockTemplateRepo.On("GetByID", ctx, int64(1)).Return(docType, nil)
		mockDocRepo.On("Create", ctx, mock.AnythingOfType("*entities.Document")).Return(errors.New("db error"))

		req := &dto.CreateFromTemplateRequest{Title: "Test", Variables: map[string]string{"name": "John"}}

		result, err := usecase.CreateDocumentFromTemplate(ctx, 1, req, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create document")
		mockTemplateRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("create document without subject in variables", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		content := testTemplateContent
		docType := &entities.DocumentType{
			ID: 1, Name: "Order", TemplateContent: &content,
			TemplateVariables: []entities.TemplateVariable{{Name: "name", Label: "Name", Required: true}},
		}
		mockTemplateRepo.On("GetByID", ctx, int64(1)).Return(docType, nil)
		mockDocRepo.On("Create", ctx, mock.MatchedBy(func(doc *entities.Document) bool {
			return doc.Subject == nil
		})).Return(nil)

		req := &dto.CreateFromTemplateRequest{Title: "Test", Variables: map[string]string{"name": "John"}}

		result, err := usecase.CreateDocumentFromTemplate(ctx, 1, req, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Subject)
		mockTemplateRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("create document with empty template content string", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		emptyContent := ""
		docType := &entities.DocumentType{ID: 1, Name: "Order", TemplateContent: &emptyContent}
		mockTemplateRepo.On("GetByID", ctx, int64(1)).Return(docType, nil)

		req := &dto.CreateFromTemplateRequest{Title: "Test", Variables: map[string]string{}}

		result, err := usecase.CreateDocumentFromTemplate(ctx, 1, req, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "template not found")
		mockTemplateRepo.AssertExpectations(t)
	})
}

func TestTemplateUseCase_UpdateTemplate(t *testing.T) {
	ctx := context.Background()

	t.Run("update template successfully", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		newContent := "Updated {{content}}"
		variables := []entities.TemplateVariable{{Name: "content", Label: "Content", Required: true}}

		mockTemplateRepo.On("UpdateTemplate", ctx, int64(1), &newContent, variables).Return(nil)

		req := &dto.UpdateTemplateRequest{TemplateContent: &newContent, TemplateVariables: variables}

		err := usecase.UpdateTemplate(ctx, 1, req)

		assert.NoError(t, err)
		mockTemplateRepo.AssertExpectations(t)
	})

	t.Run("update template error", func(t *testing.T) {
		mockTemplateRepo := new(MockTemplateRepository)
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewTemplateUseCase(mockTemplateRepo, mockDocRepo, nil)

		mockTemplateRepo.On("UpdateTemplate", ctx, int64(1), mock.Anything, mock.Anything).Return(errors.New("db error"))

		req := &dto.UpdateTemplateRequest{}

		err := usecase.UpdateTemplate(ctx, 1, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update template")
		mockTemplateRepo.AssertExpectations(t)
	})
}
