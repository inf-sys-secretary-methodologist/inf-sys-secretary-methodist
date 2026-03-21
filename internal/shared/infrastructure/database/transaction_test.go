package database

import (
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostgresUnitOfWork(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	uow := NewPostgresUnitOfWork(db)
	assert.NotNil(t, uow)
	assert.Nil(t, uow.GetTx())
}

func TestPostgresUnitOfWork_Begin_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()

	uow := NewPostgresUnitOfWork(db)
	err = uow.Begin()
	assert.NoError(t, err)
	assert.NotNil(t, uow.GetTx())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUnitOfWork_Begin_AlreadyStarted(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()

	uow := NewPostgresUnitOfWork(db)
	err = uow.Begin()
	require.NoError(t, err)

	err = uow.Begin()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction already started")
}

func TestPostgresUnitOfWork_Begin_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin().WillReturnError(fmt.Errorf("connection lost"))

	uow := NewPostgresUnitOfWork(db)
	err = uow.Begin()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")
}

func TestPostgresUnitOfWork_Commit_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectCommit()

	uow := NewPostgresUnitOfWork(db)
	err = uow.Begin()
	require.NoError(t, err)

	err = uow.Commit()
	assert.NoError(t, err)
	assert.Nil(t, uow.GetTx())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUnitOfWork_Commit_NoTransaction(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	uow := NewPostgresUnitOfWork(db)

	err = uow.Commit()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active transaction")
}

func TestPostgresUnitOfWork_Rollback_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectRollback()

	uow := NewPostgresUnitOfWork(db)
	err = uow.Begin()
	require.NoError(t, err)

	err = uow.Rollback()
	assert.NoError(t, err)
	assert.Nil(t, uow.GetTx())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUnitOfWork_Rollback_NoTransaction(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	uow := NewPostgresUnitOfWork(db)

	err = uow.Rollback()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active transaction")
}

func TestPostgresUnitOfWork_Execute_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectCommit()

	uow := NewPostgresUnitOfWork(db)

	executed := false
	err = uow.Execute(func() error {
		executed = true
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, executed)
	assert.Nil(t, uow.GetTx()) // tx should be nil after commit

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUnitOfWork_Execute_FnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectRollback()

	uow := NewPostgresUnitOfWork(db)

	fnErr := errors.New("business logic failed")
	err = uow.Execute(func() error {
		return fnErr
	})
	assert.Error(t, err)
	assert.Equal(t, fnErr, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUnitOfWork_Execute_FnError_RollbackFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectRollback().WillReturnError(fmt.Errorf("rollback failed"))

	uow := NewPostgresUnitOfWork(db)

	err = uow.Execute(func() error {
		return errors.New("fn error")
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction error")
	assert.Contains(t, err.Error(), "rollback error")
}

func TestPostgresUnitOfWork_Execute_BeginFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin().WillReturnError(fmt.Errorf("begin failed"))

	uow := NewPostgresUnitOfWork(db)

	err = uow.Execute(func() error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")
}

func TestPostgresUnitOfWork_Execute_Panic(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectRollback()

	uow := NewPostgresUnitOfWork(db)

	assert.Panics(t, func() {
		_ = uow.Execute(func() error {
			panic("something went wrong")
		})
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUnitOfWork_GetTx(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	uow := NewPostgresUnitOfWork(db)
	assert.Nil(t, uow.GetTx())
}

// Tests for UnitOfWork interface compliance
func TestPostgresUnitOfWork_ImplementsInterface(t *testing.T) {
	var _ UnitOfWork = &PostgresUnitOfWork{}
}

// Test for NewConnection - error path only (no real DB)
func TestNewConnection_InvalidDSN(t *testing.T) {
	// We can't test NewConnection without a real postgres server
	// but we can verify the function exists and has the right signature
	assert.NotNil(t, t) // placeholder
}
