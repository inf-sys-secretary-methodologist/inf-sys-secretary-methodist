package persistence

import (
	"context"
	"database/sql"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// Compile-time assertion that the PG impls satisfy the wide ports
// declared in the consuming application/usecases layer (DIP). The
// bulkTxPG assertion guards the Items/Sections/Curricula accessor
// signatures together (one Tx surface implementing three repository
// ports plus Commit/Rollback). v0.157.1.
var (
	_ usecases.BulkDisciplineItemsUnitOfWork = (*BulkDisciplineItemsUnitOfWorkPG)(nil)
	_ usecases.BulkDisciplineItemsTx         = (*bulkTxPG)(nil)
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

// Begin opens a fresh transaction with the supplied isolation options
// (pass nil for PostgreSQL default Read Committed; use
// `&sql.TxOptions{Isolation: sql.LevelRepeatableRead}` for bulk-edit
// per ADR-12 phantom-prevention).
//
// Returned BulkDisciplineItemsTx exposes tx-bound repository ports
// constructed from the same `*sql.Tx` (DBTX-based reuse per Pair 1a
// refactor — no SQL duplication между tx и non-tx paths).
func (u *BulkDisciplineItemsUnitOfWorkPG) Begin(ctx context.Context, opts *sql.TxOptions) (usecases.BulkDisciplineItemsTx, error) {
	tx, err := u.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &bulkTxPG{
		tx:        tx,
		items:     NewDisciplineItemRepositoryPG(tx),
		sections:  NewSectionRepositoryPG(tx),
		curricula: NewCurriculumRepositoryPG(tx),
	}, nil
}

// bulkTxPG is the tx-bound view returned by Begin. Holds the underlying
// `*sql.Tx` для Commit/Rollback semantics + tx-bound repo instances.
//
// `finished` flag enables idempotent close-once — defer-Rollback after
// Commit returns ErrBulkTxFinished without firing a real Rollback (which
// would error на the underlying *sql.Tx anyway).
type bulkTxPG struct {
	tx        *sql.Tx
	items     *DisciplineItemRepositoryPG
	sections  *SectionRepositoryPG
	curricula *CurriculumRepositoryPG
	finished  bool
}

// Items returns the tx-bound DisciplineItem repository.
func (t *bulkTxPG) Items() usecases.DisciplineItemRepository {
	return t.items
}

// Sections returns the tx-bound Section repository.
func (t *bulkTxPG) Sections() usecases.SectionRepository {
	return t.sections
}

// Curricula returns the tx-bound Curriculum repository.
func (t *bulkTxPG) Curricula() usecases.CurriculumRepository {
	return t.curricula
}

// Commit finalizes the transaction. Subsequent Commit/Rollback calls
// return ErrBulkTxFinished без firing on the underlying *sql.Tx.
func (t *bulkTxPG) Commit() error {
	if t.finished {
		return repositories.ErrBulkTxFinished
	}
	t.finished = true
	return t.tx.Commit()
}

// Rollback aborts the transaction. Safe to defer immediately after
// Begin — second call (after Commit или another Rollback) returns
// ErrBulkTxFinished idempotent.
func (t *bulkTxPG) Rollback() error {
	if t.finished {
		return repositories.ErrBulkTxFinished
	}
	t.finished = true
	return t.tx.Rollback()
}
