package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
)

// DocumentActivityReaderPG is the narrow read-only adapter for the
// annual report's documents-activity aggregate. Constructed independently
// from DocumentRepositoryPG so the annual orchestrator does not depend
// on the full repository's construction invariants.
//
// Stub: behavior wired in Pair 1 GREEN.
type DocumentActivityReaderPG struct {
	db *sql.DB
}

// NewDocumentActivityReaderPG creates the narrow reader.
func NewDocumentActivityReaderPG(db *sql.DB) *DocumentActivityReaderPG {
	return &DocumentActivityReaderPG{db: db}
}

// AggregateActivityByType — stub (Pair 1 RED). Real SQL execution wired
// in GREEN via shared package-private helper.
func (r *DocumentActivityReaderPG) AggregateActivityByType(ctx context.Context, from, to time.Time) ([]repositories.DocumentActivityByTypeAgg, error) {
	return nil, errors.New("documents: DocumentActivityReaderPG.AggregateActivityByType not implemented (stub)")
}
