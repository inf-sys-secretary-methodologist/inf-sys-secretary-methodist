package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
)

const testTitleV2 = "Test Title v2"

func TestDocumentVersionUseCase_GetVersions(t *testing.T) {
	ctx := context.Background()

	t.Run("get all versions successfully", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 3}
		versions := []*entities.DocumentVersion{
			{ID: 1, DocumentID: 1, Version: 1},
			{ID: 2, DocumentID: 1, Version: 2},
			{ID: 3, DocumentID: 1, Version: 3},
		}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockDocRepo.On("GetVersions", ctx, int64(1)).Return(versions, nil)

		result, err := usecase.GetVersions(ctx, 1, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(3), result.Total)
		assert.Equal(t, int64(1), result.DocumentID)
		assert.Equal(t, 3, result.LatestVersion)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("document not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		mockDocRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		result, err := usecase.GetVersions(ctx, 999, 1)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrNotFound, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("failed to get versions", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 3}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockDocRepo.On("GetVersions", ctx, int64(1)).Return(nil, errors.New("database error"))

		result, err := usecase.GetVersions(ctx, 1, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get versions")
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestDocumentVersionUseCase_GetVersion(t *testing.T) {
	ctx := context.Background()

	t.Run("get specific version successfully", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 3}
		title := testTitleV2
		version := &entities.DocumentVersion{
			ID:         2,
			DocumentID: 1,
			Version:    2,
			Title:      &title,
		}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockDocRepo.On("GetVersion", ctx, int64(1), 2).Return(version, nil)

		result, err := usecase.GetVersion(ctx, 1, 2, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.Version)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("document not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		mockDocRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		result, err := usecase.GetVersion(ctx, 999, 1, 1)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrNotFound, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("version not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 3}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockDocRepo.On("GetVersion", ctx, int64(1), 99).Return(nil, errors.New("not found"))

		result, err := usecase.GetVersion(ctx, 1, 99, 1)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrNotFound, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestDocumentVersionUseCase_CreateVersion(t *testing.T) {
	ctx := context.Background()

	t.Run("create version successfully", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 2}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockDocRepo.On("GetLatestVersion", ctx, int64(1)).Return(nil, errors.New("no versions"))
		mockDocRepo.On("CreateVersion", ctx, mock.AnythingOfType("*entities.DocumentVersion")).Return(nil)
		mockDocRepo.On("Update", ctx, mock.AnythingOfType("*entities.Document")).Return(nil)
		mockDocRepo.On("AddHistory", ctx, mock.AnythingOfType("*entities.DocumentHistory")).Return(nil)

		result, err := usecase.CreateVersion(ctx, 1, 1, "New snapshot")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 3, result.Version)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("create version with existing versions in db", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 2}
		latestVersion := &entities.DocumentVersion{ID: 3, DocumentID: 1, Version: 3}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockDocRepo.On("GetLatestVersion", ctx, int64(1)).Return(latestVersion, nil)
		mockDocRepo.On("CreateVersion", ctx, mock.MatchedBy(func(v *entities.DocumentVersion) bool {
			return v.Version == 4
		})).Return(nil)
		mockDocRepo.On("Update", ctx, mock.AnythingOfType("*entities.Document")).Return(nil)
		mockDocRepo.On("AddHistory", ctx, mock.AnythingOfType("*entities.DocumentHistory")).Return(nil)

		result, err := usecase.CreateVersion(ctx, 1, 1, "New snapshot")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 4, result.Version)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("document not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		mockDocRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		result, err := usecase.CreateVersion(ctx, 999, 1, "New snapshot")

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrNotFound, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("failed to create version", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 2}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockDocRepo.On("GetLatestVersion", ctx, int64(1)).Return(nil, errors.New("no versions"))
		mockDocRepo.On("CreateVersion", ctx, mock.AnythingOfType("*entities.DocumentVersion")).Return(errors.New("database error"))

		result, err := usecase.CreateVersion(ctx, 1, 1, "New snapshot")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create version")
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestDocumentVersionUseCase_RestoreVersion(t *testing.T) {
	ctx := context.Background()

	t.Run("restore version successfully", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 3}
		title := testTitleV2
		version := &entities.DocumentVersion{
			ID:         2,
			DocumentID: 1,
			Version:    2,
			Title:      &title,
		}
		restoredDoc := &entities.Document{ID: 1, Title: testTitleV2, Version: 4}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil).Once()
		mockDocRepo.On("GetVersion", ctx, int64(1), 2).Return(version, nil)
		mockDocRepo.On("RestoreVersion", ctx, int64(1), 2, int64(1)).Return(nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(restoredDoc, nil).Once()
		mockDocRepo.On("AddHistory", ctx, mock.AnythingOfType("*entities.DocumentHistory")).Return(nil)

		result, err := usecase.RestoreVersion(ctx, 1, 2, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testTitleV2, result.Title)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("cannot restore to current version", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 3}
		version := &entities.DocumentVersion{ID: 3, DocumentID: 1, Version: 3}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockDocRepo.On("GetVersion", ctx, int64(1), 3).Return(version, nil)

		result, err := usecase.RestoreVersion(ctx, 1, 3, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot restore to current version")
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("document not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		mockDocRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		result, err := usecase.RestoreVersion(ctx, 999, 1, 1)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrNotFound, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("version not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 3}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockDocRepo.On("GetVersion", ctx, int64(1), 99).Return(nil, errors.New("not found"))

		result, err := usecase.RestoreVersion(ctx, 1, 99, 1)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrNotFound, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("restore fails", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 3}
		version := &entities.DocumentVersion{ID: 2, DocumentID: 1, Version: 2}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockDocRepo.On("GetVersion", ctx, int64(1), 2).Return(version, nil)
		mockDocRepo.On("RestoreVersion", ctx, int64(1), 2, int64(1)).Return(errors.New("restore failed"))

		result, err := usecase.RestoreVersion(ctx, 1, 2, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to restore version")
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestDocumentVersionUseCase_CompareVersions(t *testing.T) {
	ctx := context.Background()

	t.Run("compare versions successfully", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 3}
		diff := &entities.DocumentVersionDiff{
			DocumentID:    1,
			FromVersion:   1,
			ToVersion:     2,
			ChangedFields: []string{"title", "content"},
		}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockDocRepo.On("CompareVersions", ctx, int64(1), 1, 2).Return(diff, nil)

		result, err := usecase.CompareVersions(ctx, 1, 1, 2, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.FromVersion)
		assert.Equal(t, 2, result.ToVersion)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("swap versions if from > to", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 3}
		diff := &entities.DocumentVersionDiff{
			DocumentID:    1,
			FromVersion:   1,
			ToVersion:     3,
			ChangedFields: []string{"title"},
		}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		// Should be called with swapped order (1, 3) not (3, 1)
		mockDocRepo.On("CompareVersions", ctx, int64(1), 1, 3).Return(diff, nil)

		result, err := usecase.CompareVersions(ctx, 1, 3, 1, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("document not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		mockDocRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		result, err := usecase.CompareVersions(ctx, 999, 1, 2, 1)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrNotFound, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("compare fails", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 3}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockDocRepo.On("CompareVersions", ctx, int64(1), 1, 2).Return(nil, errors.New("compare failed"))

		result, err := usecase.CompareVersions(ctx, 1, 1, 2, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to compare versions")
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestDocumentVersionUseCase_DeleteVersion(t *testing.T) {
	ctx := context.Background()

	t.Run("delete version successfully", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 3}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockDocRepo.On("DeleteVersion", ctx, int64(1), 2).Return(nil)
		mockDocRepo.On("AddHistory", ctx, mock.AnythingOfType("*entities.DocumentHistory")).Return(nil)

		err := usecase.DeleteVersion(ctx, 1, 2, 1)

		assert.NoError(t, err)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("cannot delete current version", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 3}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)

		err := usecase.DeleteVersion(ctx, 1, 3, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete current version")
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("document not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		mockDocRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found"))

		err := usecase.DeleteVersion(ctx, 999, 1, 1)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrNotFound, err)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("delete fails", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 3}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockDocRepo.On("DeleteVersion", ctx, int64(1), 2).Return(errors.New("delete failed"))

		err := usecase.DeleteVersion(ctx, 1, 2, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete version")
		mockDocRepo.AssertExpectations(t)
	})
}

func TestDocumentVersionUseCase_GetVersionFile(t *testing.T) {
	ctx := context.Background()

	t.Run("get version file successfully", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		filePath := "/uploads/doc1_v2.pdf"
		fileName := "document.pdf"
		fileSize := int64(1024)
		mimeType := "application/pdf"
		version := &entities.DocumentVersion{
			ID:         2,
			DocumentID: 1,
			Version:    2,
			FilePath:   &filePath,
			FileName:   &fileName,
			FileSize:   &fileSize,
			MimeType:   &mimeType,
		}

		mockDocRepo.On("GetVersion", ctx, int64(1), 2).Return(version, nil)

		result, err := usecase.GetVersionFile(ctx, 1, 2, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "document.pdf", result.FileName)
		assert.Equal(t, "/uploads/doc1_v2.pdf", result.FilePath)
		assert.Equal(t, int64(1024), result.FileSize)
		assert.Equal(t, "application/pdf", result.MimeType)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("version not found", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		mockDocRepo.On("GetVersion", ctx, int64(1), 99).Return(nil, errors.New("not found"))

		result, err := usecase.GetVersionFile(ctx, 1, 99, 1)

		assert.Error(t, err)
		assert.Equal(t, domainErrors.ErrNotFound, err)
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("version has no file", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		version := &entities.DocumentVersion{
			ID:         2,
			DocumentID: 1,
			Version:    2,
			FilePath:   nil,
		}

		mockDocRepo.On("GetVersion", ctx, int64(1), 2).Return(version, nil)

		result, err := usecase.GetVersionFile(ctx, 1, 2, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version has no file attached")
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})

	t.Run("version has empty file path", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		emptyPath := ""
		version := &entities.DocumentVersion{
			ID:         2,
			DocumentID: 1,
			Version:    2,
			FilePath:   &emptyPath,
		}

		mockDocRepo.On("GetVersion", ctx, int64(1), 2).Return(version, nil)

		result, err := usecase.GetVersionFile(ctx, 1, 2, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version has no file attached")
		assert.Nil(t, result)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestDocumentVersionUseCase_CreateVersion_UpdateFails(t *testing.T) {
	ctx := context.Background()

	t.Run("create version succeeds but update doc version fails", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 2}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil)
		mockDocRepo.On("GetLatestVersion", ctx, int64(1)).Return(nil, errors.New("no versions"))
		mockDocRepo.On("CreateVersion", ctx, mock.AnythingOfType("*entities.DocumentVersion")).Return(nil)
		mockDocRepo.On("Update", ctx, mock.AnythingOfType("*entities.Document")).Return(errors.New("update failed"))
		mockDocRepo.On("AddHistory", ctx, mock.AnythingOfType("*entities.DocumentHistory")).Return(nil)

		result, err := usecase.CreateVersion(ctx, 1, 1, "Test snapshot")

		// Should succeed even when update fails (logs warning but doesn't fail)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 3, result.Version)
		mockDocRepo.AssertExpectations(t)
	})
}

func TestDocumentVersionUseCase_RestoreVersion_GetUpdatedDocFails(t *testing.T) {
	ctx := context.Background()

	t.Run("restore version - get updated doc fails", func(t *testing.T) {
		mockDocRepo := new(MockDocumentRepository)
		usecase := NewDocumentVersionUseCase(mockDocRepo, nil, nil)

		doc := &entities.Document{ID: 1, Title: "Test Doc", Version: 3}
		title := testTitleV2
		version := &entities.DocumentVersion{
			ID: 2, DocumentID: 1, Version: 2, Title: &title,
		}

		mockDocRepo.On("GetByID", ctx, int64(1)).Return(doc, nil).Once()
		mockDocRepo.On("GetVersion", ctx, int64(1), 2).Return(version, nil)
		mockDocRepo.On("RestoreVersion", ctx, int64(1), 2, int64(1)).Return(nil)
		mockDocRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("not found")).Once()

		result, err := usecase.RestoreVersion(ctx, 1, 2, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get updated document")
		mockDocRepo.AssertExpectations(t)
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("ptrToStr with nil", func(t *testing.T) {
		result := ptrToStr(nil)
		assert.Equal(t, "", result)
	})

	t.Run("ptrToStr with value", func(t *testing.T) {
		s := "test"
		result := ptrToStr(&s)
		assert.Equal(t, "test", result)
	})

	t.Run("ptrToInt64 with nil", func(t *testing.T) {
		result := ptrToInt64(nil)
		assert.Equal(t, int64(0), result)
	})

	t.Run("ptrToInt64 with value", func(t *testing.T) {
		i := int64(42)
		result := ptrToInt64(&i)
		assert.Equal(t, int64(42), result)
	})
}
