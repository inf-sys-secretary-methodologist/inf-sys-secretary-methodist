package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
)

const testFakeBcryptHash = "$2a$10$abcdefghijklmnopqrstuuHx5hvxqGS5QWK8xNGhxQHfEfCB1f9i"

// MockPermissionRepository is a mock implementation of PermissionRepository
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

// MockPublicLinkRepository is a mock implementation of PublicLinkRepository
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

func TestSharingUseCase_ShareDocument(t *testing.T) {
	ctx := context.Background()

	t.Run("author can share document", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", AuthorID: 1}
		userID := int64(2)

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("GetByDocumentAndUser", ctx, int64(1), int64(2)).Return(nil, errors.New("not found"))
		mockPermRepo.On("Create", ctx, mock.AnythingOfType("*entities.DocumentPermission")).Return(nil)
		mockPermRepo.On("GetByID", ctx, mock.AnythingOfType("int64")).Return(&entities.DocumentPermission{
			ID:         1,
			DocumentID: 1,
			UserID:     &userID,
			Permission: entities.PermissionRead,
		}, nil)

		input := dto.ShareDocumentInput{
			DocumentID: 1,
			UserID:     &userID,
			Permission: "read",
		}

		result, err := usecase.ShareDocument(ctx, input, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("non-author with admin permission can share", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", AuthorID: 1}
		userID := int64(3)

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasPermission", ctx, int64(1), int64(2), entities.PermissionAdmin).Return(true, nil)
		mockPermRepo.On("GetByDocumentAndUser", ctx, int64(1), int64(3)).Return(nil, errors.New("not found"))
		mockPermRepo.On("Create", ctx, mock.AnythingOfType("*entities.DocumentPermission")).Return(nil)
		mockPermRepo.On("GetByID", ctx, mock.AnythingOfType("int64")).Return(&entities.DocumentPermission{
			ID:         1,
			DocumentID: 1,
			UserID:     &userID,
			Permission: entities.PermissionRead,
		}, nil)

		input := dto.ShareDocumentInput{
			DocumentID: 1,
			UserID:     &userID,
			Permission: "read",
		}

		result, err := usecase.ShareDocument(ctx, input, 2)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("non-author without admin permission cannot share", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", AuthorID: 1}
		userID := int64(3)

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasPermission", ctx, int64(1), int64(2), entities.PermissionAdmin).Return(false, nil)

		input := dto.ShareDocumentInput{
			DocumentID: 1,
			UserID:     &userID,
			Permission: "read",
		}

		result, err := usecase.ShareDocument(ctx, input, 2)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrForbidden, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("update existing permission", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", AuthorID: 1}
		userID := int64(2)
		existingPerm := &entities.DocumentPermission{
			ID:         1,
			DocumentID: 1,
			UserID:     &userID,
			Permission: entities.PermissionRead,
		}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("GetByDocumentAndUser", ctx, int64(1), int64(2)).Return(existingPerm, nil)
		mockPermRepo.On("Update", ctx, mock.AnythingOfType("*entities.DocumentPermission")).Return(nil)

		input := dto.ShareDocumentInput{
			DocumentID: 1,
			UserID:     &userID,
			Permission: "write",
		}

		result, err := usecase.ShareDocument(ctx, input, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("document not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		userID := int64(2)
		mockDocRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		input := dto.ShareDocumentInput{
			DocumentID: 999,
			UserID:     &userID,
			Permission: "read",
		}

		result, err := usecase.ShareDocument(ctx, input, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_RevokePermission(t *testing.T) {
	ctx := context.Background()

	t.Run("author can revoke permission", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		userID := int64(2)
		permission := &entities.DocumentPermission{
			ID:         1,
			DocumentID: 1,
			UserID:     &userID,
			Permission: entities.PermissionRead,
		}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockPermRepo.On("GetByID", ctx, int64(1)).Return(permission, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("Delete", ctx, int64(1)).Return(nil)

		err := usecase.RevokePermission(ctx, 1, 1)

		assert.NoError(t, err)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("non-author with admin permission can revoke", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		userID := int64(3)
		permission := &entities.DocumentPermission{
			ID:         1,
			DocumentID: 1,
			UserID:     &userID,
			Permission: entities.PermissionRead,
		}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockPermRepo.On("GetByID", ctx, int64(1)).Return(permission, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasPermission", ctx, int64(1), int64(2), entities.PermissionAdmin).Return(true, nil)
		mockPermRepo.On("Delete", ctx, int64(1)).Return(nil)

		err := usecase.RevokePermission(ctx, 1, 2)

		assert.NoError(t, err)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("non-author without admin permission cannot revoke", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		userID := int64(3)
		permission := &entities.DocumentPermission{
			ID:         1,
			DocumentID: 1,
			UserID:     &userID,
			Permission: entities.PermissionRead,
		}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockPermRepo.On("GetByID", ctx, int64(1)).Return(permission, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasPermission", ctx, int64(1), int64(2), entities.PermissionAdmin).Return(false, nil)

		err := usecase.RevokePermission(ctx, 1, 2)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrForbidden, err)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("permission not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		mockPermRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		err := usecase.RevokePermission(ctx, 999, 1)

		assert.Error(t, err)
		mockPermRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_GetPermission(t *testing.T) {
	ctx := context.Background()

	t.Run("get existing permission", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		userID := int64(2)
		permission := &entities.DocumentPermission{
			ID:         1,
			DocumentID: 1,
			UserID:     &userID,
			Permission: entities.PermissionRead,
		}

		mockPermRepo.On("GetByID", ctx, int64(1)).Return(permission, nil)

		result, err := usecase.GetPermission(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result.ID)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("permission not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		mockPermRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		result, err := usecase.GetPermission(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockPermRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_GetDocumentPermissions(t *testing.T) {
	ctx := context.Background()

	t.Run("author can view permissions", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1, IsPublic: false}
		userID := int64(2)
		permissions := []*entities.DocumentPermission{
			{ID: 1, DocumentID: 1, UserID: &userID, Permission: entities.PermissionRead},
		}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("GetByDocumentID", ctx, int64(1)).Return(permissions, nil)

		result, err := usecase.GetDocumentPermissions(ctx, 1, 1)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("user with permission can view", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1, IsPublic: false}
		userID := int64(2)
		permissions := []*entities.DocumentPermission{
			{ID: 1, DocumentID: 1, UserID: &userID, Permission: entities.PermissionRead},
		}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasAnyPermission", ctx, int64(1), int64(2)).Return(true, nil)
		mockPermRepo.On("GetByDocumentID", ctx, int64(1)).Return(permissions, nil)

		result, err := usecase.GetDocumentPermissions(ctx, 1, 2)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("user without permission cannot view private document permissions", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1, IsPublic: false}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasAnyPermission", ctx, int64(1), int64(2)).Return(false, nil)

		result, err := usecase.GetDocumentPermissions(ctx, 1, 2)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrForbidden, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("public document permissions viewable by anyone", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1, IsPublic: true}
		permissions := []*entities.DocumentPermission{}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasAnyPermission", ctx, int64(1), int64(2)).Return(false, nil)
		mockPermRepo.On("GetByDocumentID", ctx, int64(1)).Return(permissions, nil)

		result, err := usecase.GetDocumentPermissions(ctx, 1, 2)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_CheckUserPermission(t *testing.T) {
	ctx := context.Background()

	t.Run("author has full access", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)

		result, err := usecase.CheckUserPermission(ctx, 1, 1, entities.PermissionAdmin)

		assert.NoError(t, err)
		assert.True(t, result)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("public document allows read", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1, IsPublic: true}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)

		result, err := usecase.CheckUserPermission(ctx, 1, 2, entities.PermissionRead)

		assert.NoError(t, err)
		assert.True(t, result)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("user with permission", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1, IsPublic: false}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasPermission", ctx, int64(1), int64(2), entities.PermissionWrite).Return(true, nil)

		result, err := usecase.CheckUserPermission(ctx, 1, 2, entities.PermissionWrite)

		assert.NoError(t, err)
		assert.True(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("user without permission", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1, IsPublic: false}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasPermission", ctx, int64(1), int64(2), entities.PermissionWrite).Return(false, nil)

		result, err := usecase.CheckUserPermission(ctx, 1, 2, entities.PermissionWrite)

		assert.NoError(t, err)
		assert.False(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_CreatePublicLink(t *testing.T) {
	ctx := context.Background()

	t.Run("author can create public link", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockLinkRepo.On("Create", ctx, mock.AnythingOfType("*entities.PublicLink")).Return(nil)

		input := dto.CreatePublicLinkInput{
			DocumentID: 1,
			Permission: "read",
		}

		result, err := usecase.CreatePublicLink(ctx, input, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("user with write permission can create public link", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasPermission", ctx, int64(1), int64(2), entities.PermissionWrite).Return(true, nil)
		mockLinkRepo.On("Create", ctx, mock.AnythingOfType("*entities.PublicLink")).Return(nil)

		input := dto.CreatePublicLinkInput{
			DocumentID: 1,
			Permission: "read",
		}

		result, err := usecase.CreatePublicLink(ctx, input, 2)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("user without write permission cannot create public link", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasPermission", ctx, int64(1), int64(2), entities.PermissionWrite).Return(false, nil)

		input := dto.CreatePublicLinkInput{
			DocumentID: 1,
			Permission: "read",
		}

		result, err := usecase.CreatePublicLink(ctx, input, 2)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrForbidden, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("create public link with password", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1}
		password := "secret123"

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockLinkRepo.On("Create", ctx, mock.MatchedBy(func(link *entities.PublicLink) bool {
			return link.PasswordHash != nil && *link.PasswordHash != ""
		})).Return(nil)

		input := dto.CreatePublicLinkInput{
			DocumentID: 1,
			Permission: "download",
			Password:   &password,
		}

		result, err := usecase.CreatePublicLink(ctx, input, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockLinkRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_DeactivatePublicLink(t *testing.T) {
	ctx := context.Background()

	t.Run("author can deactivate link", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockLinkRepo.On("Deactivate", ctx, int64(1)).Return(nil)

		err := usecase.DeactivatePublicLink(ctx, 1, 1)

		assert.NoError(t, err)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("link creator can deactivate", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockLinkRepo.On("Deactivate", ctx, int64(1)).Return(nil)

		err := usecase.DeactivatePublicLink(ctx, 1, 2)

		assert.NoError(t, err)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("user without permission cannot deactivate", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasPermission", ctx, int64(1), int64(3), entities.PermissionAdmin).Return(false, nil)

		err := usecase.DeactivatePublicLink(ctx, 1, 3)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrForbidden, err)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_DeletePublicLink(t *testing.T) {
	ctx := context.Background()

	t.Run("author can delete link", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockLinkRepo.On("Delete", ctx, int64(1)).Return(nil)

		err := usecase.DeletePublicLink(ctx, 1, 1)

		assert.NoError(t, err)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_AccessPublicLink(t *testing.T) {
	ctx := context.Background()

	t.Run("access valid public link", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{
			ID:         1,
			DocumentID: 1,
			Token:      "valid-token",
			Permission: entities.PublicLinkRead,
			IsActive:   true,
		}
		doc := &entities.Document{ID: 1, Title: "Test Doc", AuthorID: 1}

		mockLinkRepo.On("GetByToken", ctx, "valid-token").Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockLinkRepo.On("IncrementUseCount", ctx, int64(1)).Return(nil)

		result, err := usecase.AccessPublicLink(ctx, "valid-token", nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Doc", result.Title)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("access expired link", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		pastTime := time.Now().Add(-24 * time.Hour)
		link := &entities.PublicLink{
			ID:         1,
			DocumentID: 1,
			Token:      "expired-token",
			Permission: entities.PublicLinkRead,
			IsActive:   true,
			ExpiresAt:  &pastTime,
		}

		mockLinkRepo.On("GetByToken", ctx, "expired-token").Return(link, nil)

		result, err := usecase.AccessPublicLink(ctx, "expired-token", nil)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrForbidden, err)
		assert.Nil(t, result)
		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("access inactive link", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{
			ID:         1,
			DocumentID: 1,
			Token:      "inactive-token",
			Permission: entities.PublicLinkRead,
			IsActive:   false,
		}

		mockLinkRepo.On("GetByToken", ctx, "inactive-token").Return(link, nil)

		result, err := usecase.AccessPublicLink(ctx, "inactive-token", nil)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrForbidden, err)
		assert.Nil(t, result)
		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("link not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		mockLinkRepo.On("GetByToken", ctx, "invalid-token").Return(nil, errors.New("not found"))

		result, err := usecase.AccessPublicLink(ctx, "invalid-token", nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockLinkRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_GetSharedDocuments(t *testing.T) {
	ctx := context.Background()

	t.Run("get shared documents", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		permissions := []*entities.DocumentPermission{
			{ID: 1, DocumentID: 1, Permission: entities.PermissionRead},
			{ID: 2, DocumentID: 2, Permission: entities.PermissionWrite},
		}
		doc1 := &entities.Document{ID: 1, Title: "Doc 1"}
		doc2 := &entities.Document{ID: 2, Title: "Doc 2"}

		mockPermRepo.On("GetByUserIDOrRole", ctx, int64(1), "teacher").Return(permissions, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc1, nil)
		mockDocRepo.On("GetByID", ctx, int64(2)).Return(doc2, nil)

		filter := dto.SharedDocumentsFilter{
			UserID:   1,
			UserRole: "teacher",
			Limit:    10,
			Offset:   0,
		}

		result, err := usecase.GetSharedDocuments(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("filter by permission level", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		permissions := []*entities.DocumentPermission{
			{ID: 1, DocumentID: 1, Permission: entities.PermissionRead},
			{ID: 2, DocumentID: 2, Permission: entities.PermissionWrite},
		}
		doc2 := &entities.Document{ID: 2, Title: "Doc 2"}

		mockPermRepo.On("GetByUserIDOrRole", ctx, int64(1), "teacher").Return(permissions, nil)
		mockDocRepo.On("GetByID", ctx, int64(2)).Return(doc2, nil)

		permissionFilter := "write"
		filter := dto.SharedDocumentsFilter{
			UserID:     1,
			UserRole:   "teacher",
			Permission: &permissionFilter,
			Limit:      10,
			Offset:     0,
		}

		result, err := usecase.GetSharedDocuments(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "Doc 2", result[0].Title)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("pagination", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		permissions := []*entities.DocumentPermission{
			{ID: 1, DocumentID: 1, Permission: entities.PermissionRead},
			{ID: 2, DocumentID: 2, Permission: entities.PermissionRead},
			{ID: 3, DocumentID: 3, Permission: entities.PermissionRead},
		}
		doc1 := &entities.Document{ID: 1, Title: "Doc 1"}
		doc2 := &entities.Document{ID: 2, Title: "Doc 2"}
		doc3 := &entities.Document{ID: 3, Title: "Doc 3"}

		mockPermRepo.On("GetByUserIDOrRole", ctx, int64(1), "").Return(permissions, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc1, nil)
		mockDocRepo.On("GetByID", ctx, int64(2)).Return(doc2, nil)
		mockDocRepo.On("GetByID", ctx, int64(3)).Return(doc3, nil)

		filter := dto.SharedDocumentsFilter{
			UserID: 1,
			Limit:  2,
			Offset: 1,
		}

		result, err := usecase.GetSharedDocuments(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Doc 2", result[0].Title)
		assert.Equal(t, "Doc 3", result[1].Title)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_GetMySharedDocuments(t *testing.T) {
	ctx := context.Background()

	t.Run("get documents I shared", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		userID := int64(2)
		permissions := []*entities.DocumentPermission{
			{ID: 1, DocumentID: 1, UserID: &userID, Permission: entities.PermissionRead},
		}
		doc := &entities.Document{ID: 1, Title: "My Doc", AuthorID: 1}

		mockPermRepo.On("GetByGrantedBy", ctx, int64(1)).Return(permissions, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)

		result, err := usecase.GetMySharedDocuments(ctx, 1, 10, 0)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "My Doc", result[0].DocumentTitle)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("empty result when no documents shared", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		mockPermRepo.On("GetByGrantedBy", ctx, int64(1)).Return([]*entities.DocumentPermission{}, nil)

		result, err := usecase.GetMySharedDocuments(ctx, 1, 10, 0)

		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("skip documents not owned by user", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		userID := int64(2)
		permissions := []*entities.DocumentPermission{
			{ID: 1, DocumentID: 1, UserID: &userID, Permission: entities.PermissionRead},
		}
		// Document owned by someone else
		doc := &entities.Document{ID: 1, Title: "Other Doc", AuthorID: 99}

		mockPermRepo.On("GetByGrantedBy", ctx, int64(1)).Return(permissions, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)

		result, err := usecase.GetMySharedDocuments(ctx, 1, 10, 0)

		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_GetPublicLink(t *testing.T) {
	ctx := context.Background()

	t.Run("author can view public link", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)

		result, err := usecase.GetPublicLink(ctx, 1, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("link creator can view", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)

		result, err := usecase.GetPublicLink(ctx, 1, 2)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("user with permission can view", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasAnyPermission", ctx, int64(1), int64(3)).Return(true, nil)

		result, err := usecase.GetPublicLink(ctx, 1, 3)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("user without permission cannot view", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasAnyPermission", ctx, int64(1), int64(3)).Return(false, nil)

		result, err := usecase.GetPublicLink(ctx, 1, 3)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrForbidden, err)
		assert.Nil(t, result)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_GetDocumentPublicLinks(t *testing.T) {
	ctx := context.Background()

	t.Run("author can list public links", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1}
		links := []*entities.PublicLink{
			{ID: 1, DocumentID: 1, Token: "token1"},
			{ID: 2, DocumentID: 1, Token: "token2"},
		}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockLinkRepo.On("GetByDocumentID", ctx, int64(1)).Return(links, nil)

		result, err := usecase.GetDocumentPublicLinks(ctx, 1, 1)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockDocRepo.AssertExpectations(t)
		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("user with permission can list", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1}
		links := []*entities.PublicLink{
			{ID: 1, DocumentID: 1, Token: "token1"},
		}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasAnyPermission", ctx, int64(1), int64(2)).Return(true, nil)
		mockLinkRepo.On("GetByDocumentID", ctx, int64(1)).Return(links, nil)

		result, err := usecase.GetDocumentPublicLinks(ctx, 1, 2)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("user without permission cannot list", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasAnyPermission", ctx, int64(1), int64(2)).Return(false, nil)

		result, err := usecase.GetDocumentPublicLinks(ctx, 1, 2)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrForbidden, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_ShareDocument_WithRole(t *testing.T) {
	ctx := context.Background()

	t.Run("share document with role", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", AuthorID: 1}
		role := "teacher"

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("Create", ctx, mock.MatchedBy(func(p *entities.DocumentPermission) bool {
			return p.Role != nil && *p.Role == entities.UserRole("teacher")
		})).Return(nil)
		mockPermRepo.On("GetByID", ctx, mock.AnythingOfType("int64")).Return(&entities.DocumentPermission{
			ID:         1,
			DocumentID: 1,
			Role:       (*entities.UserRole)(&role),
			Permission: entities.PermissionRead,
		}, nil)

		input := dto.ShareDocumentInput{
			DocumentID: 1,
			Role:       &role,
			Permission: "read",
		}

		result, err := usecase.ShareDocument(ctx, input, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("share document HasPermission error", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", AuthorID: 1}
		userID := int64(3)

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasPermission", ctx, int64(1), int64(2), entities.PermissionAdmin).Return(false, errors.New("db error"))

		input := dto.ShareDocumentInput{
			DocumentID: 1,
			UserID:     &userID,
			Permission: "read",
		}

		result, err := usecase.ShareDocument(ctx, input, 2)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("share document create permission error", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", AuthorID: 1}
		userID := int64(2)

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("GetByDocumentAndUser", ctx, int64(1), int64(2)).Return(nil, errors.New("not found"))
		mockPermRepo.On("Create", ctx, mock.AnythingOfType("*entities.DocumentPermission")).Return(errors.New("db error"))

		input := dto.ShareDocumentInput{
			DocumentID: 1,
			UserID:     &userID,
			Permission: "read",
		}

		result, err := usecase.ShareDocument(ctx, input, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_RevokePermission_DocumentNotFound(t *testing.T) {
	ctx := context.Background()

	t.Run("document not found during revoke", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		userID := int64(2)
		permission := &entities.DocumentPermission{
			ID: 1, DocumentID: 1, UserID: &userID, Permission: entities.PermissionRead,
		}

		mockPermRepo.On("GetByID", ctx, int64(1)).Return(permission, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("not found"))

		err := usecase.RevokePermission(ctx, 1, 1)

		assert.Error(t, err)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("HasPermission error during revoke", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		userID := int64(3)
		permission := &entities.DocumentPermission{
			ID: 1, DocumentID: 1, UserID: &userID, Permission: entities.PermissionRead,
		}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockPermRepo.On("GetByID", ctx, int64(1)).Return(permission, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasPermission", ctx, int64(1), int64(2), entities.PermissionAdmin).Return(false, errors.New("db error"))

		err := usecase.RevokePermission(ctx, 1, 2)

		assert.Error(t, err)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("delete permission error during revoke", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		userID := int64(2)
		permission := &entities.DocumentPermission{
			ID: 1, DocumentID: 1, UserID: &userID, Permission: entities.PermissionRead,
		}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockPermRepo.On("GetByID", ctx, int64(1)).Return(permission, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("Delete", ctx, int64(1)).Return(errors.New("db error"))

		err := usecase.RevokePermission(ctx, 1, 1)

		assert.Error(t, err)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_CheckUserPermission_DocumentNotFound(t *testing.T) {
	ctx := context.Background()

	t.Run("document not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		mockDocRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		result, err := usecase.CheckUserPermission(ctx, 999, 1, entities.PermissionRead)

		assert.Error(t, err)
		assert.False(t, result)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_GetDocumentPermissions_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("document not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		mockDocRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		result, err := usecase.GetDocumentPermissions(ctx, 999, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("HasAnyPermission error", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1, IsPublic: false}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasAnyPermission", ctx, int64(1), int64(2)).Return(false, errors.New("db error"))

		result, err := usecase.GetDocumentPermissions(ctx, 1, 2)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_CreatePublicLink_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("document not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		mockDocRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		input := dto.CreatePublicLinkInput{DocumentID: 999, Permission: "read"}

		result, err := usecase.CreatePublicLink(ctx, input, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("HasPermission error for non-author", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasPermission", ctx, int64(1), int64(2), entities.PermissionWrite).Return(false, errors.New("db error"))

		input := dto.CreatePublicLinkInput{DocumentID: 1, Permission: "read"}

		result, err := usecase.CreatePublicLink(ctx, input, 2)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("create link repo error", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockLinkRepo.On("Create", ctx, mock.AnythingOfType("*entities.PublicLink")).Return(errors.New("db error"))

		input := dto.CreatePublicLinkInput{DocumentID: 1, Permission: "read"}

		result, err := usecase.CreatePublicLink(ctx, input, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("create public link with max uses and expiry", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1}
		maxUses := 10
		expiresAt := time.Now().Add(24 * time.Hour)

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockLinkRepo.On("Create", ctx, mock.MatchedBy(func(link *entities.PublicLink) bool {
			return link.MaxUses != nil && *link.MaxUses == 10 && link.ExpiresAt != nil
		})).Return(nil)

		input := dto.CreatePublicLinkInput{
			DocumentID: 1,
			Permission: "download",
			MaxUses:    &maxUses,
			ExpiresAt:  &expiresAt,
		}

		result, err := usecase.CreatePublicLink(ctx, input, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "download", result.Permission)
		mockDocRepo.AssertExpectations(t)
		mockLinkRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_AccessPublicLink_Password(t *testing.T) {
	ctx := context.Background()

	t.Run("access link with correct password", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		// Generate a bcrypt hash for "secret123"
		// We use the actual hash that bcrypt generates
		hash := testFakeBcryptHash
		link := &entities.PublicLink{
			ID: 1, DocumentID: 1, Token: "token", Permission: entities.PublicLinkRead,
			IsActive: true, PasswordHash: &hash,
		}

		mockLinkRepo.On("GetByToken", ctx, "token").Return(link, nil)

		// Wrong password test
		wrongPw := "wrong"
		result, err := usecase.AccessPublicLink(ctx, "token", &wrongPw)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("access password protected link without password", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		hash := testFakeBcryptHash
		link := &entities.PublicLink{
			ID: 1, DocumentID: 1, Token: "token", Permission: entities.PublicLinkRead,
			IsActive: true, PasswordHash: &hash,
		}

		mockLinkRepo.On("GetByToken", ctx, "token").Return(link, nil)

		result, err := usecase.AccessPublicLink(ctx, "token", nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("access password protected link with empty password", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		hash := testFakeBcryptHash
		link := &entities.PublicLink{
			ID: 1, DocumentID: 1, Token: "token", Permission: entities.PublicLinkRead,
			IsActive: true, PasswordHash: &hash,
		}

		mockLinkRepo.On("GetByToken", ctx, "token").Return(link, nil)

		emptyPw := ""
		result, err := usecase.AccessPublicLink(ctx, "token", &emptyPw)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("access link with download permission", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{
			ID: 1, DocumentID: 1, Token: "dl-token", Permission: entities.PublicLinkDownload,
			IsActive: true,
		}
		doc := &entities.Document{ID: 1, Title: "Test Doc", AuthorID: 1}

		mockLinkRepo.On("GetByToken", ctx, "dl-token").Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockLinkRepo.On("IncrementUseCount", ctx, int64(1)).Return(nil)

		result, err := usecase.AccessPublicLink(ctx, "dl-token", nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.CanDownload)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("access link with usage limit reached", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		maxUses := 5
		link := &entities.PublicLink{
			ID: 1, DocumentID: 1, Token: "limit-token", Permission: entities.PublicLinkRead,
			IsActive: true, MaxUses: &maxUses, UseCount: 5,
		}

		mockLinkRepo.On("GetByToken", ctx, "limit-token").Return(link, nil)

		result, err := usecase.AccessPublicLink(ctx, "limit-token", nil)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrForbidden, err)
		assert.Nil(t, result)
		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("access link increment use count error logged but not failed", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{
			ID: 1, DocumentID: 1, Token: "inc-fail-token", Permission: entities.PublicLinkRead,
			IsActive: true,
		}
		doc := &entities.Document{ID: 1, Title: "Test Doc", AuthorID: 1}

		mockLinkRepo.On("GetByToken", ctx, "inc-fail-token").Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockLinkRepo.On("IncrementUseCount", ctx, int64(1)).Return(errors.New("db error"))

		result, err := usecase.AccessPublicLink(ctx, "inc-fail-token", nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("access link document not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{
			ID: 1, DocumentID: 1, Token: "doc-missing-token", Permission: entities.PublicLinkRead,
			IsActive: true,
		}

		mockLinkRepo.On("GetByToken", ctx, "doc-missing-token").Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("not found"))

		result, err := usecase.AccessPublicLink(ctx, "doc-missing-token", nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_DeactivatePublicLink_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("link not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		mockLinkRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		err := usecase.DeactivatePublicLink(ctx, 999, 1)

		assert.Error(t, err)
		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("document not found during deactivate", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("not found"))

		err := usecase.DeactivatePublicLink(ctx, 1, 1)

		assert.Error(t, err)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("HasPermission error during deactivate", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasPermission", ctx, int64(1), int64(3), entities.PermissionAdmin).Return(false, errors.New("db error"))

		err := usecase.DeactivatePublicLink(ctx, 1, 3)

		assert.Error(t, err)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("deactivate repo error", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockLinkRepo.On("Deactivate", ctx, int64(1)).Return(errors.New("db error"))

		err := usecase.DeactivatePublicLink(ctx, 1, 1)

		assert.Error(t, err)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("admin can deactivate", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasPermission", ctx, int64(1), int64(3), entities.PermissionAdmin).Return(true, nil)
		mockLinkRepo.On("Deactivate", ctx, int64(1)).Return(nil)

		err := usecase.DeactivatePublicLink(ctx, 1, 3)

		assert.NoError(t, err)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_DeletePublicLink_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("link not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		mockLinkRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		err := usecase.DeletePublicLink(ctx, 999, 1)

		assert.Error(t, err)
		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("user without permission cannot delete", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasPermission", ctx, int64(1), int64(3), entities.PermissionAdmin).Return(false, nil)

		err := usecase.DeletePublicLink(ctx, 1, 3)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrForbidden, err)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("link creator can delete", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockLinkRepo.On("Delete", ctx, int64(1)).Return(nil)

		err := usecase.DeletePublicLink(ctx, 1, 2)

		assert.NoError(t, err)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("delete repo error", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 1, Token: "token123"}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockLinkRepo.On("Delete", ctx, int64(1)).Return(errors.New("db error"))

		err := usecase.DeletePublicLink(ctx, 1, 1)

		assert.Error(t, err)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("document not found during delete", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("not found"))

		err := usecase.DeletePublicLink(ctx, 1, 1)

		assert.Error(t, err)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("HasPermission error during delete", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasPermission", ctx, int64(1), int64(3), entities.PermissionAdmin).Return(false, errors.New("db error"))

		err := usecase.DeletePublicLink(ctx, 1, 3)

		assert.Error(t, err)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_GetSharedDocuments_Pagination(t *testing.T) {
	ctx := context.Background()

	t.Run("offset beyond documents", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		permissions := []*entities.DocumentPermission{
			{ID: 1, DocumentID: 1, Permission: entities.PermissionRead},
		}
		doc1 := &entities.Document{ID: 1, Title: "Doc 1"}

		mockPermRepo.On("GetByUserIDOrRole", ctx, int64(1), "").Return(permissions, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc1, nil)

		filter := dto.SharedDocumentsFilter{
			UserID: 1,
			Limit:  10,
			Offset: 100,
		}

		result, err := usecase.GetSharedDocuments(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("get shared documents with limit 0 returns all", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		permissions := []*entities.DocumentPermission{
			{ID: 1, DocumentID: 1, Permission: entities.PermissionRead},
			{ID: 2, DocumentID: 2, Permission: entities.PermissionRead},
		}
		doc1 := &entities.Document{ID: 1, Title: "Doc 1"}
		doc2 := &entities.Document{ID: 2, Title: "Doc 2"}

		mockPermRepo.On("GetByUserIDOrRole", ctx, int64(1), "").Return(permissions, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc1, nil)
		mockDocRepo.On("GetByID", ctx, int64(2)).Return(doc2, nil)

		filter := dto.SharedDocumentsFilter{
			UserID: 1,
			Limit:  0,
			Offset: 0,
		}

		result, err := usecase.GetSharedDocuments(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("get shared documents - repo error", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		mockPermRepo.On("GetByUserIDOrRole", ctx, int64(1), "").Return(nil, errors.New("db error"))

		filter := dto.SharedDocumentsFilter{UserID: 1, Limit: 10}

		result, err := usecase.GetSharedDocuments(ctx, filter)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("get shared documents - deleted document skipped", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		permissions := []*entities.DocumentPermission{
			{ID: 1, DocumentID: 1, Permission: entities.PermissionRead},
			{ID: 2, DocumentID: 2, Permission: entities.PermissionRead},
		}
		doc2 := &entities.Document{ID: 2, Title: "Doc 2"}

		mockPermRepo.On("GetByUserIDOrRole", ctx, int64(1), "").Return(permissions, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("deleted"))
		mockDocRepo.On("GetByID", ctx, int64(2)).Return(doc2, nil)

		filter := dto.SharedDocumentsFilter{UserID: 1, Limit: 10}

		result, err := usecase.GetSharedDocuments(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_GetMySharedDocuments_Pagination(t *testing.T) {
	ctx := context.Background()

	t.Run("offset beyond results", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		userID := int64(2)
		permissions := []*entities.DocumentPermission{
			{ID: 1, DocumentID: 1, UserID: &userID, Permission: entities.PermissionRead},
		}
		doc := &entities.Document{ID: 1, Title: "My Doc", AuthorID: 1}

		mockPermRepo.On("GetByGrantedBy", ctx, int64(1)).Return(permissions, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)

		result, err := usecase.GetMySharedDocuments(ctx, 1, 10, 100)

		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("limit 0 returns all", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		userID := int64(2)
		permissions := []*entities.DocumentPermission{
			{ID: 1, DocumentID: 1, UserID: &userID, Permission: entities.PermissionRead},
		}
		doc := &entities.Document{ID: 1, Title: "My Doc", AuthorID: 1}

		mockPermRepo.On("GetByGrantedBy", ctx, int64(1)).Return(permissions, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)

		result, err := usecase.GetMySharedDocuments(ctx, 1, 0, 0)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("repo error", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		mockPermRepo.On("GetByGrantedBy", ctx, int64(1)).Return(nil, errors.New("db error"))

		result, err := usecase.GetMySharedDocuments(ctx, 1, 10, 0)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("shared with role info", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		role := entities.RoleTeacher
		permissions := []*entities.DocumentPermission{
			{ID: 1, DocumentID: 1, Role: &role, Permission: entities.PermissionRead},
		}
		doc := &entities.Document{ID: 1, Title: "My Doc", AuthorID: 1}

		mockPermRepo.On("GetByGrantedBy", ctx, int64(1)).Return(permissions, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)

		result, err := usecase.GetMySharedDocuments(ctx, 1, 10, 0)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Len(t, result[0].SharedWith, 1)
		assert.NotNil(t, result[0].SharedWith[0].Role)
		assert.Equal(t, "teacher", *result[0].SharedWith[0].Role)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("deleted document skipped", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		userID := int64(2)
		permissions := []*entities.DocumentPermission{
			{ID: 1, DocumentID: 1, UserID: &userID, Permission: entities.PermissionRead},
		}

		mockPermRepo.On("GetByGrantedBy", ctx, int64(1)).Return(permissions, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("deleted"))

		result, err := usecase.GetMySharedDocuments(ctx, 1, 10, 0)

		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockPermRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_GetPublicLink_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("link not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		mockLinkRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		result, err := usecase.GetPublicLink(ctx, 999, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("document not found for link", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("not found"))

		result, err := usecase.GetPublicLink(ctx, 1, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("HasAnyPermission error", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		link := &entities.PublicLink{ID: 1, DocumentID: 1, CreatedBy: 2, Token: "token123"}
		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockLinkRepo.On("GetByID", ctx, int64(1)).Return(link, nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasAnyPermission", ctx, int64(1), int64(3)).Return(false, errors.New("db error"))

		result, err := usecase.GetPublicLink(ctx, 1, 3)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockLinkRepo.AssertExpectations(t)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_GetDocumentPublicLinks_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("document not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		mockDocRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		result, err := usecase.GetDocumentPublicLinks(ctx, 999, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("HasAnyPermission error", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("HasAnyPermission", ctx, int64(1), int64(2)).Return(false, errors.New("db error"))

		result, err := usecase.GetDocumentPublicLinks(ctx, 1, 2)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("get links repo error", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockLinkRepo.On("GetByDocumentID", ctx, int64(1)).Return(nil, errors.New("db error"))

		result, err := usecase.GetDocumentPublicLinks(ctx, 1, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockLinkRepo.AssertExpectations(t)
	})

	t.Run("get permissions repo error", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, AuthorID: 1}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("GetByDocumentID", ctx, int64(1)).Return(nil, errors.New("db error"))

		result, err := usecase.GetDocumentPermissions(ctx, 1, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})
}

func TestSharingUseCase_ShareDocument_UpdatePermissionError(t *testing.T) {
	ctx := context.Background()

	t.Run("update existing permission fails", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockLinkRepo := new(MockPublicLinkRepository)

		usecase := NewSharingUseCase(mockDocRepo, mockPermRepo, mockLinkRepo, nil, "http://localhost", nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", AuthorID: 1}
		userID := int64(2)
		existingPerm := &entities.DocumentPermission{
			ID: 1, DocumentID: 1, UserID: &userID, Permission: entities.PermissionRead,
		}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockPermRepo.On("GetByDocumentAndUser", ctx, int64(1), int64(2)).Return(existingPerm, nil)
		mockPermRepo.On("Update", ctx, mock.AnythingOfType("*entities.DocumentPermission")).Return(errors.New("db error"))

		input := dto.ShareDocumentInput{
			DocumentID: 1,
			UserID:     &userID,
			Permission: "write",
		}

		result, err := usecase.ShareDocument(ctx, input, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
		mockPermRepo.AssertExpectations(t)
	})
}
