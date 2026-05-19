package usecases

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// Each of CategoryUseCase / TagUseCase / DocumentUseCase /
// DocumentVersionUseCase carries its own (uc *X) logAudit method that
// guards a nil *logging.AuditLogger via `if uc.auditLog != nil { ... }`.
// Existing tests all construct с nil auditLog, leaving the non-nil
// branch at 50% per-func. The four tests below pass a real
// *logging.AuditLogger constructed via logging.NewAuditLogger so the
// emit branch fires — closes 50% → 100% per-func.

func TestCategoryUseCase_logAudit_NonNilLoggerEmits(t *testing.T) {
	mockRepo := new(MockCategoryRepository)
	auditLog := logging.NewAuditLogger(logging.NewLogger("debug"))
	uc := NewCategoryUseCase(mockRepo, auditLog)
	ctx := context.Background()

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.DocumentCategory")).Return(nil).Once()
	mockRepo.On("HasChildren", ctx, mock.AnythingOfType("int64")).Return(false, nil).Once()
	mockRepo.On("GetDocumentCount", ctx, mock.AnythingOfType("int64"), false).Return(int64(0), nil).Once()

	out, err := uc.Create(ctx, dto.CreateCategoryInput{Name: "logaudit-coverage"})
	assert.NoError(t, err)
	assert.NotNil(t, out, "logAudit non-nil branch must not interfere с happy path")
	mockRepo.AssertExpectations(t)
}

func TestTagUseCase_logAudit_NonNilLoggerEmits(t *testing.T) {
	mockTagRepo := new(MockTagRepository)
	mockDocRepo := new(MockDocumentRepository)
	auditLog := logging.NewAuditLogger(logging.NewLogger("debug"))
	uc := NewTagUseCase(mockTagRepo, mockDocRepo, auditLog)
	ctx := context.Background()

	mockTagRepo.On("GetByName", ctx, "tag-logaudit").Return(nil, assert.AnError).Once()
	mockTagRepo.On("Create", ctx, mock.AnythingOfType("*entities.DocumentTag")).Return(nil).Once()

	out, err := uc.Create(ctx, dto.CreateTagInput{Name: "tag-logaudit"})
	assert.NoError(t, err)
	assert.NotNil(t, out)
	mockTagRepo.AssertExpectations(t)
}

func TestDocumentUseCase_logAudit_NonNilLoggerEmits(t *testing.T) {
	mockDocRepo := new(MockDocumentRepository)
	mockTypeRepo := new(MockDocumentTypeRepository)
	mockCategoryRepo := new(MockDocumentCategoryRepository)
	auditLog := logging.NewAuditLogger(logging.NewLogger("debug"))
	uc := NewDocumentUseCase(mockDocRepo, mockTypeRepo, mockCategoryRepo, nil /* s3Client unused for no-file Create */, auditLog)
	ctx := context.Background()

	docType := &entities.DocumentType{ID: 1, Name: "Приказ"}
	// Use mock.Anything for ctx: DocumentUseCase.Create wraps ctx via otel tracer Start,
	// so identity comparison против context.Background() would miss.
	mockTypeRepo.On("GetByID", mock.Anything, int64(1)).Return(docType, nil).Once()
	mockDocRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Document")).Return(nil).Once()
	mockDocRepo.On("AddHistory", mock.Anything, mock.AnythingOfType("*entities.DocumentHistory")).Return(nil).Once()

	out, err := uc.Create(ctx, dto.CreateDocumentInput{
		Title:          "logaudit-coverage",
		DocumentTypeID: 1,
	}, int64(42))
	assert.NoError(t, err)
	assert.NotNil(t, out)
	mockDocRepo.AssertExpectations(t)
	mockTypeRepo.AssertExpectations(t)
}

func TestDocumentVersionUseCase_logAudit_NonNilLoggerEmits(t *testing.T) {
	mockDocRepo := new(MockDocumentRepository)
	auditLog := logging.NewAuditLogger(logging.NewLogger("debug"))
	uc := NewDocumentVersionUseCase(mockDocRepo, nil /* s3 unused for GetVersions */, auditLog)
	ctx := context.Background()

	doc := &entities.Document{ID: 1, Version: 3}
	mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil).Once()
	mockDocRepo.On("GetVersions", ctx, int64(1)).Return([]*entities.DocumentVersion{}, nil).Once()

	out, err := uc.GetVersions(ctx, 1, int64(42))
	assert.NoError(t, err)
	assert.NotNil(t, out)
	mockDocRepo.AssertExpectations(t)
}
