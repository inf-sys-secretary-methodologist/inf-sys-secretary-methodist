package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// SignatureRepository persists and retrieves document signatures. Declared in
// the consumer (usecase) package per the dependency-inversion principle; the
// PostgreSQL implementation lives in infrastructure/persistence.
//
// Issue: #140
type SignatureRepository interface {
	// Save persists a new signature and returns its generated id.
	Save(ctx context.Context, sig *entities.DocumentSignature) (int64, error)
	// ListByDocument returns all signatures for a document, oldest first.
	ListByDocument(ctx context.Context, documentID int64) ([]*entities.DocumentSignature, error)
	// GetByID returns a single signature or ErrSignatureNotFound.
	GetByID(ctx context.Context, id int64) (*entities.DocumentSignature, error)
}
