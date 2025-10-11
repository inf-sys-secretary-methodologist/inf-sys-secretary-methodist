package database

import (
	"database/sql"
	"fmt"
)

// UnitOfWork manages database transactions
type UnitOfWork interface {
	Begin() error
	Commit() error
	Rollback() error
	Execute(fn func() error) error
}

// PostgresUnitOfWork implements UnitOfWork for PostgreSQL
type PostgresUnitOfWork struct {
	db *sql.DB
	tx *sql.Tx
}

// NewPostgresUnitOfWork creates a new unit of work
func NewPostgresUnitOfWork(db *sql.DB) *PostgresUnitOfWork {
	return &PostgresUnitOfWork{
		db: db,
	}
}

// Begin starts a new transaction
func (uow *PostgresUnitOfWork) Begin() error {
	if uow.tx != nil {
		return fmt.Errorf("transaction already started")
	}

	tx, err := uow.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	uow.tx = tx
	return nil
}

// Commit commits the current transaction
func (uow *PostgresUnitOfWork) Commit() error {
	if uow.tx == nil {
		return fmt.Errorf("no active transaction")
	}

	err := uow.tx.Commit()
	uow.tx = nil
	return err
}

// Rollback rolls back the current transaction
func (uow *PostgresUnitOfWork) Rollback() error {
	if uow.tx == nil {
		return fmt.Errorf("no active transaction")
	}

	err := uow.tx.Rollback()
	uow.tx = nil
	return err
}

// Execute executes a function within a transaction
func (uow *PostgresUnitOfWork) Execute(fn func() error) error {
	if err := uow.Begin(); err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			uow.Rollback()
			panic(r)
		}
	}()

	if err := fn(); err != nil {
		if rbErr := uow.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	return uow.Commit()
}

// GetTx returns the current transaction
func (uow *PostgresUnitOfWork) GetTx() *sql.Tx {
	return uow.tx
}
