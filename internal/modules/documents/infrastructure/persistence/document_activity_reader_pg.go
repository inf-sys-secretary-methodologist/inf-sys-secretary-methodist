package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
)

// DocumentActivityReaderPG is the narrow read-only adapter for the
// annual report's documents-activity aggregate. Constructed independently
// from DocumentRepositoryPG so the cross-module orchestrator does not
// depend on the full repository's construction invariants — if
// DocumentRepositoryPG evolves to require new collaborators (logger,
// cache, metrics), the activity reader is unaffected.
//
// Delegates to the package-private aggregateActivityByType helper —
// the single source of truth for the documents-activity SQL query.
type DocumentActivityReaderPG struct {
	db *sql.DB
}

// NewDocumentActivityReaderPG creates the narrow reader.
func NewDocumentActivityReaderPG(db *sql.DB) *DocumentActivityReaderPG {
	return &DocumentActivityReaderPG{db: db}
}

// AggregateActivityByType counts documents grouped by
// (document_type.name, document.status) over [from, to). Empty result
// is not an error.
func (r *DocumentActivityReaderPG) AggregateActivityByType(ctx context.Context, from, to time.Time) ([]usecases.DocumentActivityByTypeAgg, error) {
	return aggregateActivityByType(ctx, r.db, from, to)
}
