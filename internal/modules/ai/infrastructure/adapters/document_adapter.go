// Package adapters contains adapters for external module dependencies.
package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/services"
	docRepositories "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
)

// DocumentAdapter adapts the documents module for AI use.
// It bridges the AI module with the documents module and S3 storage,
// providing text extraction from binary document formats.
type DocumentAdapter struct {
	documentRepo          docRepositories.DocumentRepository
	s3Client              *storage.S3Client
	textExtractionService *services.TextExtractionService
	db                    *sql.DB
	logger                *slog.Logger
}

// NewDocumentAdapter creates a new document adapter.
func NewDocumentAdapter(
	documentRepo docRepositories.DocumentRepository,
	s3Client *storage.S3Client,
	textExtractionService *services.TextExtractionService,
	db *sql.DB,
	logger *slog.Logger,
) *DocumentAdapter {
	return &DocumentAdapter{
		documentRepo:          documentRepo,
		s3Client:              s3Client,
		textExtractionService: textExtractionService,
		db:                    db,
		logger:                logger,
	}
}

// GetDocumentContent retrieves the text content of a document.
// It first checks the cached content in document_versions. If not available,
// it downloads the file from S3, extracts text, and caches it.
func (a *DocumentAdapter) GetDocumentContent(ctx context.Context, documentID int64) (string, string, error) {
	doc, err := a.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get document: %w", err)
	}

	// Get the latest version content (may already have cached text)
	version, err := a.documentRepo.GetLatestVersion(ctx, documentID)
	if err != nil {
		// If no version exists, try to extract from file directly
		return a.extractFromFile(ctx, documentID, doc.FilePath, doc.MimeType, doc.Title)
	}

	// If content is already cached in the version, return it
	if version.Content != nil && *version.Content != "" {
		return *version.Content, doc.Title, nil
	}

	// No cached content — try to extract from the stored file
	content, title, err := a.extractFromFile(ctx, documentID, doc.FilePath, doc.MimeType, doc.Title)
	if err != nil {
		return "", title, err
	}

	// Cache extracted content in document_versions (best-effort)
	if content != "" && version.ID > 0 {
		a.cacheContent(ctx, version.ID, content)
	}

	return content, title, nil
}

// extractFromFile downloads a file from S3 and extracts text content.
func (a *DocumentAdapter) extractFromFile(ctx context.Context, documentID int64, filePath, mimeType *string, title string) (string, string, error) {
	if filePath == nil || *filePath == "" {
		return "", title, nil
	}

	if mimeType == nil || *mimeType == "" {
		a.logger.Warn("document has file but no MIME type, skipping extraction",
			"document_id", documentID,
			"file_path", *filePath,
		)
		return "", title, nil
	}

	if !a.textExtractionService.CanExtract(*mimeType) {
		a.logger.Debug("unsupported MIME type for text extraction",
			"document_id", documentID,
			"mime_type", *mimeType,
		)
		return "", title, nil
	}

	if a.s3Client == nil {
		a.logger.Warn("S3 client not available, cannot extract text from file",
			"document_id", documentID,
		)
		return "", title, nil
	}

	// Download from S3
	reader, _, err := a.s3Client.Download(ctx, *filePath)
	if err != nil {
		return "", title, fmt.Errorf("failed to download file from S3: %w", err)
	}
	defer func() { _ = reader.Close() }()

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", title, fmt.Errorf("failed to read file data: %w", err)
	}

	// Extract text
	content, err := a.textExtractionService.Extract(data, *mimeType)
	if err != nil {
		return "", title, fmt.Errorf("failed to extract text from %s: %w", *mimeType, err)
	}

	return content, title, nil
}

// cacheContent writes extracted text to document_versions.content for future use.
// Uses direct SQL to avoid changing the DocumentRepository interface.
func (a *DocumentAdapter) cacheContent(ctx context.Context, versionID int64, content string) {
	if a.db == nil {
		return
	}

	_, err := a.db.ExecContext(ctx,
		"UPDATE document_versions SET content = $1 WHERE id = $2",
		content, versionID,
	)
	if err != nil {
		a.logger.Warn("failed to cache extracted content in document_versions",
			"version_id", versionID,
			"error", err,
		)
	}
}
