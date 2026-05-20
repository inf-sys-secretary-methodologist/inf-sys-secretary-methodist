package repositories

import "errors"

// ErrBulkTxFinished signals that a Commit/Rollback was attempted on
// a BulkDisciplineItemsTx that has already been committed or rolled
// back. Idempotent close-once semantics guard against double-finish
// in defer chains.
var ErrBulkTxFinished = errors.New("bulk_uow: tx already committed or rolled back")

// The BulkDisciplineItemsUnitOfWork + BulkDisciplineItemsTx ports live в
// internal/modules/curriculum/application/usecases/repository_interfaces.go
// (DIP — interface lives with consumer). v0.157.1.
