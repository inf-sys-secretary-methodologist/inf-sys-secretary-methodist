package repositories

import (
	"context"
	"database/sql"
	"errors"
)

// ErrBulkTxFinished signals that a Commit/Rollback was attempted on
// a BulkDisciplineItemsTx that has already been committed or rolled
// back. Idempotent close-once semantics guard against double-finish
// in defer chains.
var ErrBulkTxFinished = errors.New("bulk_uow: tx already committed or rolled back")

// BulkDisciplineItemsUnitOfWork opens a transactional boundary spanning
// DisciplineItem + Section + Curriculum repositories. Used by the
// bulk-edit endpoint (v0.128.3 ADR-10) to apply creates+updates+deletes
// atomically — все succeed (Commit) или ничего не applies (Rollback).
//
// The pattern intentionally excludes leaking `*sql.Tx` или any infra
// primitive — caller получает only the narrow BulkDisciplineItemsTx
// interface, which composes tx-bound repository ports for use by the
// usecase. Mirror к Vaughn Vernon's "Aggregate transactional consistency
// boundary" pattern, but spans 3 ARs because bulk-edit-РПД is a single
// editorial action в the methodist's mental model.
type BulkDisciplineItemsUnitOfWork interface {
	// Begin opens a fresh transaction. opts may pass an isolation level
	// (Repeatable Read recommended per ADR-12 — phantom-prevention для
	// long bulk operations); pass nil for PostgreSQL default (Read
	// Committed). Each call returns an independent BulkDisciplineItemsTx
	// — UoW is concurrency-safe (state lives on the returned Tx, not
	// the UoW itself).
	Begin(ctx context.Context, opts *sql.TxOptions) (BulkDisciplineItemsTx, error)
}

// BulkDisciplineItemsTx is the tx-bound view exposing repository ports
// backed by the same `*sql.Tx`. Caller MUST eventually call Commit() or
// Rollback() — never both, never neither (recommend `defer tx.Rollback()`
// pattern; second call returns ErrBulkTxFinished idempotent).
type BulkDisciplineItemsTx interface {
	// Items returns DisciplineItem repository operations bound to the tx.
	Items() DisciplineItemRepository

	// Sections returns Section repository operations bound to the tx.
	Sections() SectionRepository

	// Curricula returns Curriculum repository operations bound to the tx.
	Curricula() CurriculumRepository

	// Commit finalizes the transaction. Subsequent calls return
	// ErrBulkTxFinished.
	Commit() error

	// Rollback aborts the transaction. Subsequent calls return
	// ErrBulkTxFinished. Safe to defer immediately after Begin —
	// Rollback after Commit is a no-op (returns ErrBulkTxFinished
	// which caller may ignore).
	Rollback() error
}
