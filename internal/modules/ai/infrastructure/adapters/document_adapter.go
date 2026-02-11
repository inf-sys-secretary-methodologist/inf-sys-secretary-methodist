// Package adapters contains adapters for external module dependencies.
package adapters

import (
	"context"
	"fmt"

	docRepositories "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
)

// DocumentAdapter adapts the documents module for AI use
type DocumentAdapter struct {
	documentRepo docRepositories.DocumentRepository
}

// NewDocumentAdapter creates a new document adapter
func NewDocumentAdapter(documentRepo docRepositories.DocumentRepository) *DocumentAdapter {
	return &DocumentAdapter{
		documentRepo: documentRepo,
	}
}

// GetDocumentContent retrieves the text content of a document
func (a *DocumentAdapter) GetDocumentContent(ctx context.Context, documentID int64) (string, string, error) {
	doc, err := a.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get document: %w", err)
	}

	// Get the latest version content
	version, err := a.documentRepo.GetLatestVersion(ctx, documentID)
	if err != nil {
		// If no version exists, return empty content with title
		return "", doc.Title, nil
	}

	// The content is stored in the version
	content := ""
	if version.Content != nil {
		content = *version.Content
	}
	return content, doc.Title, nil
}
