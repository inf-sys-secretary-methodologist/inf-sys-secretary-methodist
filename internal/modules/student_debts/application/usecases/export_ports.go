package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
)

// DebtExporter serializes a registry snapshot into a downloadable document
// (xlsx now, potentially other formats later). The concrete adapter lives
// in infrastructure/ and is wired in main.go — this package owns only the
// port (DIP).
//
// It receives the fully-hydrated aggregates (root + resit attempts) rather
// than a flattened row DTO, because a round-trippable export lays the
// attempt timeline out on its own sheet — a single flat row can't carry it.
// The infrastructure adapter legitimately depends on the domain entities
// (infrastructure → domain is an allowed dependency direction).
type DebtExporter interface {
	Export(ctx context.Context, debts []*entities.StudentDebt) ([]byte, error)
}
