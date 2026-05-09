package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// BulkDisciplineItemsUnitOfWorkPG is the PostgreSQL implementation of
// BulkDisciplineItemsUnitOfWork (v0.128.3 ADR-10). Wraps `*sql.DB`
// and produces tx-bound repos backed by `*sql.Tx`.
//
// Concurrency-safe — Begin returns a fresh BulkDisciplineItemsTx per
// call; UoW itself holds only the connection pool reference.
type BulkDisciplineItemsUnitOfWorkPG struct {
	db *sql.DB
}

// NewBulkDisciplineItemsUnitOfWorkPG constructs the UoW. Panics on
// nil db — failure-closed DI (mirror к other curriculum constructors).
func NewBulkDisciplineItemsUnitOfWorkPG(db *sql.DB) *BulkDisciplineItemsUnitOfWorkPG {
	if db == nil {
		panic("bulk_uow: NewBulkDisciplineItemsUnitOfWorkPG requires non-nil db")
	}
	return &BulkDisciplineItemsUnitOfWorkPG{db: db}
}

// Begin opens a fresh transaction. Pair 1b RED stub — returns an error
// signaling not-yet-implemented. Pair 1c GREEN replaces with proper
// BeginTx + bulkTxPG construction.
func (u *BulkDisciplineItemsUnitOfWorkPG) Begin(ctx context.Context, opts *sql.TxOptions) (repositories.BulkDisciplineItemsTx, error) {
	_ = ctx
	_ = opts
	return nil, errors.New("bulk_uow: not implemented (Pair 1b RED stub)")
}
