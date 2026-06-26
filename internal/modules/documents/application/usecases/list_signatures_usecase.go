package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// ListSignaturesUseCase returns all signatures applied to a document.
type ListSignaturesUseCase struct {
	repo SignatureRepository
}

// NewListSignaturesUseCase wires the use case.
func NewListSignaturesUseCase(repo SignatureRepository) *ListSignaturesUseCase {
	if repo == nil {
		panic("documents: NewListSignaturesUseCase requires non-nil repo")
	}
	return &ListSignaturesUseCase{repo: repo}
}

// Execute lists the document's signatures, oldest first.
func (uc *ListSignaturesUseCase) Execute(ctx context.Context, documentID int64) ([]*entities.DocumentSignature, error) {
	return uc.repo.ListByDocument(ctx, documentID)
}
