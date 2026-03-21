package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

func newFunFactRepoMock(t *testing.T) (*FunFactRepositoryPg, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewFunFactRepositoryPg(db), mock
}

func TestFunFactCreate_Success(t *testing.T) {
	repo, mock := newFunFactRepoMock(t)
	now := time.Now()
	fact := &entities.FunFact{Content: "Fun fact", Category: "science", Source: "wiki", SourceURL: "http://example.com", Language: "en", IsApproved: true}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO ai_fun_facts")).
		WithArgs(fact.Content, fact.Category, fact.Source, fact.SourceURL, fact.Language, fact.IsApproved).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(1), now, now))

	err := repo.Create(context.Background(), fact)
	require.NoError(t, err)
	assert.Equal(t, int64(1), fact.ID)
}

func TestFunFactCreate_Error(t *testing.T) {
	repo, mock := newFunFactRepoMock(t)
	fact := &entities.FunFact{Content: "Fun fact", Category: "science", Language: "en"}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO ai_fun_facts")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert error"))

	err := repo.Create(context.Background(), fact)
	assert.Error(t, err)
}

func TestFunFactBulkCreate_Success(t *testing.T) {
	repo, mock := newFunFactRepoMock(t)
	facts := []entities.FunFact{
		{Content: "Fact 1", Category: "math", Language: "en", IsApproved: true},
		{Content: "Fact 2", Category: "science", Language: "en", IsApproved: false},
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO ai_fun_facts"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ai_fun_facts")).
		WithArgs(facts[0].Content, facts[0].Category, facts[0].Source, facts[0].SourceURL, facts[0].Language, facts[0].IsApproved).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ai_fun_facts")).
		WithArgs(facts[1].Content, facts[1].Category, facts[1].Source, facts[1].SourceURL, facts[1].Language, facts[1].IsApproved).
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	err := repo.BulkCreate(context.Background(), facts)
	require.NoError(t, err)
}

func TestFunFactBulkCreate_BeginError(t *testing.T) {
	repo, mock := newFunFactRepoMock(t)

	mock.ExpectBegin().WillReturnError(fmt.Errorf("begin error"))

	err := repo.BulkCreate(context.Background(), []entities.FunFact{{Content: "Fact"}})
	assert.Error(t, err)
}

func TestFunFactBulkCreate_PrepareError(t *testing.T) {
	repo, mock := newFunFactRepoMock(t)

	mock.ExpectBegin()
	mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO ai_fun_facts")).WillReturnError(fmt.Errorf("prepare error"))
	mock.ExpectRollback()

	err := repo.BulkCreate(context.Background(), []entities.FunFact{{Content: "Fact"}})
	assert.Error(t, err)
}

func TestFunFactBulkCreate_ExecError(t *testing.T) {
	repo, mock := newFunFactRepoMock(t)

	mock.ExpectBegin()
	mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO ai_fun_facts"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ai_fun_facts")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("exec error"))
	mock.ExpectRollback()

	err := repo.BulkCreate(context.Background(), []entities.FunFact{{Content: "Fact"}})
	assert.Error(t, err)
}

func TestFunFactGetRandom_Success(t *testing.T) {
	repo, mock := newFunFactRepoMock(t)
	now := time.Now()

	cols := []string{"id", "content", "category", "source", "source_url", "language", "is_approved", "used_count", "last_used_at", "created_at", "updated_at"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, content, category")).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(int64(1), "Fun fact", "science", "wiki", "http://url", "en", true, 5, &now, now, now))

	fact, err := repo.GetRandom(context.Background())
	require.NoError(t, err)
	require.NotNil(t, fact)
	assert.Equal(t, "Fun fact", fact.Content)
}

func TestFunFactGetRandom_NoRows(t *testing.T) {
	repo, mock := newFunFactRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, content, category")).
		WillReturnError(sql.ErrNoRows)

	fact, err := repo.GetRandom(context.Background())
	require.NoError(t, err)
	assert.Nil(t, fact)
}

func TestFunFactGetRandom_DBError(t *testing.T) {
	repo, mock := newFunFactRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, content, category")).
		WillReturnError(fmt.Errorf("db error"))

	fact, err := repo.GetRandom(context.Background())
	assert.Error(t, err)
	assert.Nil(t, fact)
}

func TestFunFactGetLeastUsed(t *testing.T) {
	repo, mock := newFunFactRepoMock(t)
	now := time.Now()

	cols := []string{"id", "content", "category", "source", "source_url", "language", "is_approved", "used_count", "last_used_at", "created_at", "updated_at"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, content, category")).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(int64(1), "Fact", "math", "", "", "en", true, 0, nil, now, now))

	fact, err := repo.GetLeastUsed(context.Background())
	require.NoError(t, err)
	require.NotNil(t, fact)
}

func TestFunFactIncrementUsedCount_Success(t *testing.T) {
	repo, mock := newFunFactRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE ai_fun_facts SET used_count")).
		WithArgs(sqlmock.AnyArg(), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.IncrementUsedCount(context.Background(), 1)
	require.NoError(t, err)
}

func TestFunFactIncrementUsedCount_Error(t *testing.T) {
	repo, mock := newFunFactRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE ai_fun_facts SET used_count")).
		WithArgs(sqlmock.AnyArg(), int64(1)).
		WillReturnError(fmt.Errorf("update error"))

	err := repo.IncrementUsedCount(context.Background(), 1)
	assert.Error(t, err)
}

func TestFunFactCount_Success(t *testing.T) {
	repo, mock := newFunFactRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM ai_fun_facts")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(42)))

	count, err := repo.Count(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(42), count)
}

func TestFunFactCount_Error(t *testing.T) {
	repo, mock := newFunFactRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM ai_fun_facts")).
		WillReturnError(fmt.Errorf("count error"))

	_, err := repo.Count(context.Background())
	assert.Error(t, err)
}
