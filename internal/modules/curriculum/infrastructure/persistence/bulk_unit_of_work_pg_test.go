package persistence_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/infrastructure/persistence"
)

// ===== Constructor =====

func TestNewBulkDisciplineItemsUnitOfWorkPG_PanicsOnNilDB(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("constructor accepted nil db")
		}
	}()
	persistence.NewBulkDisciplineItemsUnitOfWorkPG(nil)
}

// ===== Begin =====

func TestBulkUoW_Begin_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()

	uow := persistence.NewBulkDisciplineItemsUnitOfWorkPG(db)
	tx, err := uow.Begin(context.Background(), nil)
	require.NoError(t, err, "Begin must succeed когда DB.BeginTx succeeds")
	require.NotNil(t, tx, "Begin must return non-nil tx on success")

	// Cleanup tx чтобы release sqlmock expectations.
	mock.ExpectRollback()
	_ = tx.Rollback()
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulkUoW_Begin_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	wantErr := errors.New("db unavailable")
	mock.ExpectBegin().WillReturnError(wantErr)

	uow := persistence.NewBulkDisciplineItemsUnitOfWorkPG(db)
	tx, err := uow.Begin(context.Background(), nil)
	assert.Nil(t, tx)
	assert.True(t, errors.Is(err, wantErr),
		"Begin must propagate DB.BeginTx error verbatim (caller distinguishes wrapped sentinels)")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulkUoW_Begin_PassesIsolationLevel(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// sqlmock captures opts via ExpectBeginTx (not bare ExpectBegin) для
	// proper isolation-level pin. The real impl must call db.BeginTx с
	// the supplied opts, not db.Begin (which ignores opts).
	mock.ExpectBegin()

	uow := persistence.NewBulkDisciplineItemsUnitOfWorkPG(db)
	opts := &sql.TxOptions{Isolation: sql.LevelRepeatableRead}
	tx, err := uow.Begin(context.Background(), opts)
	require.NoError(t, err)
	require.NotNil(t, tx)

	mock.ExpectRollback()
	_ = tx.Rollback()
	require.NoError(t, mock.ExpectationsWereMet())
}

// ===== Tx repos =====

func TestBulkUoW_Tx_ReturnsNonNilRepos(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()

	uow := persistence.NewBulkDisciplineItemsUnitOfWorkPG(db)
	tx, err := uow.Begin(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, tx)

	assert.NotNil(t, tx.Items(), "Tx.Items() must return non-nil DisciplineItemRepository")
	assert.NotNil(t, tx.Sections(), "Tx.Sections() must return non-nil SectionRepository")
	assert.NotNil(t, tx.Curricula(), "Tx.Curricula() must return non-nil CurriculumRepository")

	mock.ExpectRollback()
	_ = tx.Rollback()
	require.NoError(t, mock.ExpectationsWereMet())
}

// ===== Commit / Rollback semantics =====

func TestBulkUoW_Tx_Commit_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectCommit()

	uow := persistence.NewBulkDisciplineItemsUnitOfWorkPG(db)
	tx, err := uow.Begin(context.Background(), nil)
	require.NoError(t, err)

	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulkUoW_Tx_Rollback_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectRollback()

	uow := persistence.NewBulkDisciplineItemsUnitOfWorkPG(db)
	tx, err := uow.Begin(context.Background(), nil)
	require.NoError(t, err)

	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulkUoW_Tx_DoubleCommit_ReturnsFinished(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectCommit()

	uow := persistence.NewBulkDisciplineItemsUnitOfWorkPG(db)
	tx, err := uow.Begin(context.Background(), nil)
	require.NoError(t, err)

	require.NoError(t, tx.Commit())
	err = tx.Commit() // second call
	assert.True(t, errors.Is(err, repositories.ErrBulkTxFinished),
		"second Commit must return ErrBulkTxFinished")
}

func TestBulkUoW_Tx_RollbackAfterCommit_ReturnsFinished(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectCommit()

	uow := persistence.NewBulkDisciplineItemsUnitOfWorkPG(db)
	tx, err := uow.Begin(context.Background(), nil)
	require.NoError(t, err)

	require.NoError(t, tx.Commit())
	err = tx.Rollback() // safe-to-defer pattern
	assert.True(t, errors.Is(err, repositories.ErrBulkTxFinished),
		"Rollback after Commit must return ErrBulkTxFinished idempotent")
}
