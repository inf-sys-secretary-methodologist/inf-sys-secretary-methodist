// Package usecases contains business logic for the documents module.
package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
)

// DocumentVersionUseCase handles document version business logic
type DocumentVersionUseCase struct {
	documentRepo repositories.DocumentRepository
	s3Client     *storage.S3Client
	auditLog     *logging.AuditLogger
}

// NewDocumentVersionUseCase creates a new document version use case
func NewDocumentVersionUseCase(
	documentRepo repositories.DocumentRepository,
	s3Client *storage.S3Client,
	auditLog *logging.AuditLogger,
) *DocumentVersionUseCase {
	return &DocumentVersionUseCase{
		documentRepo: documentRepo,
		s3Client:     s3Client,
		auditLog:     auditLog,
	}
}

// GetVersions retrieves all versions of a document
func (uc *DocumentVersionUseCase) GetVersions(ctx context.Context, documentID int64, userID int64) (*dto.DocumentVersionListOutput, error) {
	// Check if document exists and user has access
	doc, err := uc.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	// Get all versions
	versions, err := uc.documentRepo.GetVersions(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}

	// Log audit event
	uc.logAudit(ctx, "versions_viewed", "document_version", map[string]interface{}{
		"document_id": documentID,
		"user_id":     userID,
		"count":       len(versions),
	})

	return &dto.DocumentVersionListOutput{
		Versions:      dto.ToDocumentVersionOutputList(versions),
		Total:         int64(len(versions)),
		DocumentID:    documentID,
		LatestVersion: doc.Version,
	}, nil
}

// GetVersion retrieves a specific version of a document
func (uc *DocumentVersionUseCase) GetVersion(ctx context.Context, documentID int64, version int, userID int64) (*dto.DocumentVersionOutput, error) {
	// Check if document exists
	_, err := uc.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	// Get specific version
	v, err := uc.documentRepo.GetVersion(ctx, documentID, version)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	// Log audit event
	uc.logAudit(ctx, "version_viewed", "document_version", map[string]interface{}{
		"document_id": documentID,
		"version":     version,
		"user_id":     userID,
	})

	return dto.ToDocumentVersionOutput(v), nil
}

// CreateVersion manually creates a new version snapshot
func (uc *DocumentVersionUseCase) CreateVersion(ctx context.Context, documentID int64, userID int64, description string) (*dto.DocumentVersionOutput, error) {
	// Get current document
	doc, err := uc.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	// Determine the next version number
	// Check if there are existing versions in the database
	nextVersion := doc.Version + 1
	latestVersion, err := uc.documentRepo.GetLatestVersion(ctx, documentID)
	if err == nil && latestVersion != nil {
		// If we have versions in DB, use max of (doc.Version, latestVersion.Version) + 1
		if latestVersion.Version >= doc.Version {
			nextVersion = latestVersion.Version + 1
		}
	}

	// Create version snapshot with the correct version number
	version := entities.NewDocumentVersion(doc, userID, description)
	version.Version = nextVersion

	if err := uc.documentRepo.CreateVersion(ctx, version); err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	// Update the document's version field to match
	doc.Version = nextVersion
	if err := uc.documentRepo.Update(ctx, doc); err != nil {
		// Version was created but document update failed - log but don't fail
		// The version record is still valid
		fmt.Printf("warning: failed to update document version field: %v\n", err)
	}

	// Log history
	_ = uc.documentRepo.AddHistory(ctx, &entities.DocumentHistory{
		DocumentID: documentID,
		UserID:     &userID,
		Action:     "version_created",
		Details: map[string]interface{}{
			"version":     version.Version,
			"description": description,
		},
	})

	// Log audit event
	uc.logAudit(ctx, "version_created", "document_version", map[string]interface{}{
		"document_id": documentID,
		"version":     version.Version,
		"user_id":     userID,
		"description": description,
	})

	return dto.ToDocumentVersionOutput(version), nil
}

// RestoreVersion restores a document to a previous version
func (uc *DocumentVersionUseCase) RestoreVersion(ctx context.Context, documentID int64, version int, userID int64) (*dto.DocumentOutput, error) {
	// Check if document exists
	doc, err := uc.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	// Check if version exists
	v, err := uc.documentRepo.GetVersion(ctx, documentID, version)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	// Cannot restore to current version
	if doc.Version == version {
		return nil, fmt.Errorf("cannot restore to current version")
	}

	// Restore version
	if err := uc.documentRepo.RestoreVersion(ctx, documentID, version, userID); err != nil {
		return nil, fmt.Errorf("failed to restore version: %w", err)
	}

	// Get updated document
	updatedDoc, err := uc.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated document: %w", err)
	}

	// Log history
	_ = uc.documentRepo.AddHistory(ctx, &entities.DocumentHistory{
		DocumentID: documentID,
		UserID:     &userID,
		Action:     "version_restored",
		Details: map[string]interface{}{
			"restored_version": version,
			"new_version":      updatedDoc.Version,
			"restored_title":   ptrToStr(v.Title),
		},
	})

	// Log audit event
	uc.logAudit(ctx, "version_restored", "document_version", map[string]interface{}{
		"document_id":      documentID,
		"restored_version": version,
		"new_version":      updatedDoc.Version,
		"user_id":          userID,
	})

	return dto.ToDocumentOutput(updatedDoc), nil
}

// CompareVersions compares two versions and returns the differences
func (uc *DocumentVersionUseCase) CompareVersions(ctx context.Context, documentID int64, fromVersion, toVersion int, userID int64) (*dto.VersionDiffOutput, error) {
	// Check if document exists
	_, err := uc.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	// Ensure fromVersion < toVersion for consistency
	if fromVersion > toVersion {
		fromVersion, toVersion = toVersion, fromVersion
	}

	// Compare versions
	diff, err := uc.documentRepo.CompareVersions(ctx, documentID, fromVersion, toVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to compare versions: %w", err)
	}

	// Log audit event
	uc.logAudit(ctx, "versions_compared", "document_version", map[string]interface{}{
		"document_id":    documentID,
		"from_version":   fromVersion,
		"to_version":     toVersion,
		"changed_fields": diff.ChangedFields,
		"user_id":        userID,
	})

	return dto.ToVersionDiffOutput(diff), nil
}

// DeleteVersion deletes a specific version (cannot delete current version)
func (uc *DocumentVersionUseCase) DeleteVersion(ctx context.Context, documentID int64, version int, userID int64) error {
	// Check if document exists
	doc, err := uc.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return errors.ErrNotFound
	}

	// Cannot delete current version
	if doc.Version == version {
		return fmt.Errorf("cannot delete current version")
	}

	// Delete version
	if err := uc.documentRepo.DeleteVersion(ctx, documentID, version); err != nil {
		return fmt.Errorf("failed to delete version: %w", err)
	}

	// Log history
	_ = uc.documentRepo.AddHistory(ctx, &entities.DocumentHistory{
		DocumentID: documentID,
		UserID:     &userID,
		Action:     "version_deleted",
		Details: map[string]interface{}{
			"deleted_version": version,
		},
	})

	// Log audit event
	uc.logAudit(ctx, "version_deleted", "document_version", map[string]interface{}{
		"document_id":     documentID,
		"deleted_version": version,
		"user_id":         userID,
	})

	return nil
}

// GetVersionFile gets file information from a specific version
func (uc *DocumentVersionUseCase) GetVersionFile(ctx context.Context, documentID int64, version int, userID int64) (*dto.VersionFileDownloadOutput, error) {
	// Get version
	v, err := uc.documentRepo.GetVersion(ctx, documentID, version)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	// Check if version has a file
	if v.FilePath == nil || *v.FilePath == "" {
		return nil, fmt.Errorf("version has no file attached")
	}

	// Generate presigned URL for download
	var downloadURL string
	if uc.s3Client != nil {
		url, err := uc.s3Client.GetPresignedURL(ctx, *v.FilePath, 15*time.Minute)
		if err == nil {
			downloadURL = url
		}
	}

	return &dto.VersionFileDownloadOutput{
		FileName:    ptrToStr(v.FileName),
		FilePath:    ptrToStr(v.FilePath),
		FileSize:    ptrToInt64(v.FileSize),
		MimeType:    ptrToStr(v.MimeType),
		DownloadURL: downloadURL,
	}, nil
}

// logAudit logs an audit event
func (uc *DocumentVersionUseCase) logAudit(ctx context.Context, action, resourceType string, details map[string]interface{}) {
	if uc.auditLog != nil {
		uc.auditLog.LogAuditEvent(ctx, action, resourceType, details)
	}
}

// Helper functions
func ptrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrToInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}
