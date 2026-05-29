package persistence

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

func newMORepoMock(t *testing.T) (*MinobrnaukiOrderRepositoryPG, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewMinobrnaukiOrderRepositoryPG(db), mock
}

func validOrder(t *testing.T) *entities.MinobrnaukiOrder {
	t.Helper()
	docID := int64(7)
	o, err := entities.NewMinobrnaukiOrder(entities.NewMinobrnaukiOrderInput{
		OrderNumber: "№ 1078 от 12.05.2026",
		Title:       "Об изменении ФГОС 09.03.01",
		PublishedAt: time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC),
		DocumentID:  &docID,
		ChangeScope: domain.MinobrnaukiOrderChangeScopeMajor,
		Summary:     "Обновлён перечень компетенций",
		UploadedBy:  42,
	})
	require.NoError(t, err)
	return o
}

func moColumns() []string {
	return []string{
		"id", "order_number", "title", "published_at", "document_id",
		"change_scope", "summary", "uploaded_by", "created_at",
	}
}

// --- Save ---

func TestMinobrnaukiOrderRepositoryPG_Save_NoAffected_HappyPath(t *testing.T) {
	repo, mock := newMORepoMock(t)
	o := validOrder(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO minobrnauki_orders")).
		WithArgs(
			"№ 1078 от 12.05.2026",       // order_number
			"Об изменении ФГОС 09.03.01", // title
			sqlmock.AnyArg(),             // published_at
			sqlmock.AnyArg(),             // document_id (nullable)
			"major",                      // change_scope
			sqlmock.AnyArg(),             // summary (nullable)
			int64(42),                    // uploaded_by
			sqlmock.AnyArg(),             // created_at
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(100)))
	mock.ExpectCommit()

	err := repo.Save(context.Background(), o, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(100), o.ID(), "Save must write the generated id back onto the entity")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMinobrnaukiOrderRepositoryPG_Save_WithAffected_InsertsJunctionRows(t *testing.T) {
	repo, mock := newMORepoMock(t)
	o := validOrder(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO minobrnauki_orders")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(100)))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO minobrnauki_order_affected")).
		WithArgs(int64(100), int64(11)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO minobrnauki_order_affected")).
		WithArgs(int64(100), int64(22)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Save(context.Background(), o, []int64{11, 22})
	require.NoError(t, err)
	assert.Equal(t, int64(100), o.ID())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMinobrnaukiOrderRepositoryPG_Save_AffectedInsertFailure_RollsBack(t *testing.T) {
	repo, mock := newMORepoMock(t)
	o := validOrder(t)
	boom := errors.New("simulated affected insert failure")

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO minobrnauki_orders")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(100)))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO minobrnauki_order_affected")).
		WillReturnError(boom)
	mock.ExpectRollback()

	err := repo.Save(context.Background(), o, []int64{11})
	require.Error(t, err)
	assert.ErrorIs(t, err, boom, "affected-insert failure must surface and roll back the tx")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMinobrnaukiOrderRepositoryPG_Save_BeginTxFailure_Surfaces(t *testing.T) {
	repo, mock := newMORepoMock(t)
	o := validOrder(t)
	beginErr := errors.New("simulated begin failure")
	mock.ExpectBegin().WillReturnError(beginErr)

	err := repo.Save(context.Background(), o, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, beginErr, "BeginTx failure must surface to caller")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMinobrnaukiOrderRepositoryPG_Save_OrderInsertFailure_RollsBack(t *testing.T) {
	repo, mock := newMORepoMock(t)
	o := validOrder(t)
	boom := errors.New("simulated order insert failure")

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO minobrnauki_orders")).
		WillReturnError(boom)
	mock.ExpectRollback()

	err := repo.Save(context.Background(), o, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, boom, "order-insert failure must surface and roll back the tx")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByID ---

func TestMinobrnaukiOrderRepositoryPG_GetByID_HappyPath(t *testing.T) {
	repo, mock := newMORepoMock(t)
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	pub := time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta("FROM minobrnauki_orders WHERE id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows(moColumns()).AddRow(
			int64(100), "№ 1078", "Об изменении ФГОС", pub,
			sql.NullInt64{Int64: 7, Valid: true}, "major",
			sql.NullString{String: "сводка", Valid: true}, int64(42), now,
		))

	got, err := repo.GetByID(context.Background(), 100)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, int64(100), got.ID())
	assert.Equal(t, "№ 1078", got.OrderNumber())
	assert.Equal(t, domain.MinobrnaukiOrderChangeScopeMajor, got.ChangeScope())
	require.NotNil(t, got.DocumentID())
	assert.Equal(t, int64(7), *got.DocumentID())
	assert.Equal(t, "сводка", got.Summary())
	assert.Equal(t, int64(42), got.UploadedBy())
	assert.True(t, got.PublishedAt().Equal(pub))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMinobrnaukiOrderRepositoryPG_GetByID_NullDocumentAndSummary(t *testing.T) {
	repo, mock := newMORepoMock(t)
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	pub := time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta("FROM minobrnauki_orders WHERE id = $1")).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows(moColumns()).AddRow(
			int64(5), "№ 1", "T", pub,
			sql.NullInt64{}, "minor", sql.NullString{}, int64(3), now,
		))

	got, err := repo.GetByID(context.Background(), 5)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, got.DocumentID(), "NULL document_id must hydrate to nil")
	assert.Equal(t, "", got.Summary(), "NULL summary must hydrate to empty string")
	assert.Equal(t, domain.MinobrnaukiOrderChangeScopeMinor, got.ChangeScope())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMinobrnaukiOrderRepositoryPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newMORepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM minobrnauki_orders WHERE id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	got, err := repo.GetByID(context.Background(), 999)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, repositories.ErrMinobrnaukiOrderNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- List ---

func TestMinobrnaukiOrderRepositoryPG_List_EmptyFilter_ReturnsAllRows(t *testing.T) {
	repo, mock := newMORepoMock(t)
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	pub := time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM minobrnauki_orders")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta("FROM minobrnauki_orders")).
		WillReturnRows(sqlmock.NewRows(moColumns()).
			AddRow(int64(1), "№ 1", "A", pub, sql.NullInt64{}, "minor", sql.NullString{}, int64(42), now).
			AddRow(int64(2), "№ 2", "B", pub, sql.NullInt64{Int64: 9, Valid: true}, "major", sql.NullString{String: "s", Valid: true}, int64(43), now))

	got, err := repo.List(context.Background(), repositories.MinobrnaukiOrderListFilter{Limit: 20})
	require.NoError(t, err)
	assert.Equal(t, 2, got.Total)
	require.Len(t, got.Items, 2)
	assert.Equal(t, int64(1), got.Items[0].ID)
	assert.Nil(t, got.Items[0].DocumentID)
	assert.Equal(t, domain.MinobrnaukiOrderChangeScopeMajor, got.Items[1].ChangeScope)
	require.NotNil(t, got.Items[1].DocumentID)
	assert.Equal(t, int64(9), *got.Items[1].DocumentID)
	assert.Equal(t, "s", got.Items[1].Summary)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMinobrnaukiOrderRepositoryPG_List_FilterByScopeAndUploader_PassesArgs(t *testing.T) {
	repo, mock := newMORepoMock(t)
	scope := domain.MinobrnaukiOrderChangeScopeMajor
	uploader := int64(42)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM minobrnauki_orders")).
		WithArgs("major", sql.NullInt64{Int64: 42, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("FROM minobrnauki_orders")).
		WithArgs("major", sql.NullInt64{Int64: 42, Valid: true}, 10, 5).
		WillReturnRows(sqlmock.NewRows(moColumns()))

	got, err := repo.List(context.Background(), repositories.MinobrnaukiOrderListFilter{
		ChangeScope: &scope,
		UploadedBy:  &uploader,
		Limit:       10,
		Offset:      5,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, got.Total)
	assert.Empty(t, got.Items)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMinobrnaukiOrderRepositoryPG_List_CountQueryError_Surfaces(t *testing.T) {
	repo, mock := newMORepoMock(t)
	boom := errors.New("simulated count failure")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM minobrnauki_orders")).
		WillReturnError(boom)

	_, err := repo.List(context.Background(), repositories.MinobrnaukiOrderListFilter{Limit: 20})
	require.Error(t, err)
	assert.ErrorIs(t, err, boom)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMinobrnaukiOrderRepositoryPG_List_ListQueryError_Surfaces(t *testing.T) {
	repo, mock := newMORepoMock(t)
	boom := errors.New("simulated list failure")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM minobrnauki_orders")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("FROM minobrnauki_orders")).
		WillReturnError(boom)

	_, err := repo.List(context.Background(), repositories.MinobrnaukiOrderListFilter{Limit: 20})
	require.Error(t, err)
	assert.ErrorIs(t, err, boom)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMinobrnaukiOrderRepositoryPG_List_RowIterationError_Surfaces(t *testing.T) {
	repo, mock := newMORepoMock(t)
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	pub := time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC)
	boom := errors.New("simulated row iteration failure")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM minobrnauki_orders")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("FROM minobrnauki_orders")).
		WillReturnRows(sqlmock.NewRows(moColumns()).
			AddRow(int64(1), "№ 1", "A", pub, sql.NullInt64{}, "minor", sql.NullString{}, int64(42), now).
			RowError(0, boom))

	_, err := repo.List(context.Background(), repositories.MinobrnaukiOrderListFilter{Limit: 20})
	require.Error(t, err)
	assert.ErrorIs(t, err, boom)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- FindAffected ---

func TestMinobrnaukiOrderRepositoryPG_FindAffected_ReturnsIDs(t *testing.T) {
	repo, mock := newMORepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM minobrnauki_order_affected WHERE order_id = $1")).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"work_program_id"}).
			AddRow(int64(11)).AddRow(int64(22)))

	ids, err := repo.FindAffected(context.Background(), 100)
	require.NoError(t, err)
	assert.Equal(t, []int64{11, 22}, ids)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMinobrnaukiOrderRepositoryPG_FindAffected_EmptyNotAnError(t *testing.T) {
	repo, mock := newMORepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM minobrnauki_order_affected WHERE order_id = $1")).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"work_program_id"}))

	ids, err := repo.FindAffected(context.Background(), 7)
	require.NoError(t, err)
	assert.Empty(t, ids)
	assert.NoError(t, mock.ExpectationsWereMet())
}
