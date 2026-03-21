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

func newEmbRepoMock(t *testing.T) (*EmbeddingRepositoryPg, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewEmbeddingRepositoryPg(db)
	return repo.(*EmbeddingRepositoryPg), mock
}

// ---- CreateChunk ----

func TestEmbeddingCreateChunk_Success(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	chunk := &entities.DocumentChunk{DocumentID: 1, ChunkIndex: 0, ChunkText: "text", Metadata: map[string]interface{}{"key": "val"}}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO ai_document_chunks")).
		WithArgs(chunk.DocumentID, chunk.ChunkIndex, chunk.ChunkText, chunk.ChunkTokens, chunk.PageNumber, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.CreateChunk(context.Background(), chunk)
	require.NoError(t, err)
	assert.Equal(t, int64(1), chunk.ID)
}

func TestEmbeddingCreateChunk_Error(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	chunk := &entities.DocumentChunk{DocumentID: 1, ChunkIndex: 0, ChunkText: "text"}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO ai_document_chunks")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert error"))

	err := repo.CreateChunk(context.Background(), chunk)
	assert.Error(t, err)
}

// ---- CreateChunks ----

func TestEmbeddingCreateChunks_Empty(t *testing.T) {
	repo, _ := newEmbRepoMock(t)
	err := repo.CreateChunks(context.Background(), []entities.DocumentChunk{})
	require.NoError(t, err)
}

func TestEmbeddingCreateChunks_Success(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	chunks := []entities.DocumentChunk{
		{DocumentID: 1, ChunkIndex: 0, ChunkText: "text1"},
		{DocumentID: 1, ChunkIndex: 1, ChunkText: "text2"},
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ai_document_chunks")).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.CreateChunks(context.Background(), chunks)
	require.NoError(t, err)
}

func TestEmbeddingCreateChunks_Error(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	chunks := []entities.DocumentChunk{{DocumentID: 1, ChunkIndex: 0, ChunkText: "text1"}}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ai_document_chunks")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("exec error"))

	err := repo.CreateChunks(context.Background(), chunks)
	assert.Error(t, err)
}

// ---- GetChunksByDocumentID ----

func TestEmbeddingGetChunksByDocumentID_Success(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	now := time.Now()

	cols := []string{"id", "document_id", "chunk_index", "chunk_text", "chunk_tokens", "page_number", "metadata", "created_at"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, document_id, chunk_index")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(int64(1), int64(1), 0, "text1", nil, nil, []byte(`{"key":"val"}`), now).
			AddRow(int64(2), int64(1), 1, "text2", nil, nil, nil, now))

	chunks, err := repo.GetChunksByDocumentID(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, chunks, 2)
	assert.Equal(t, "val", chunks[0].Metadata["key"])
}

func TestEmbeddingGetChunksByDocumentID_QueryError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, document_id")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetChunksByDocumentID(context.Background(), 1)
	assert.Error(t, err)
}

func TestEmbeddingGetChunksByDocumentID_ScanError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, document_id")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.GetChunksByDocumentID(context.Background(), 1)
	assert.Error(t, err)
}

// ---- DeleteChunksByDocumentID ----

func TestEmbeddingDeleteChunksByDocumentID_Success(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ai_embeddings")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 5))

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ai_document_chunks")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 5))

	err := repo.DeleteChunksByDocumentID(context.Background(), 1)
	require.NoError(t, err)
}

func TestEmbeddingDeleteChunksByDocumentID_EmbeddingDeleteError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ai_embeddings")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("delete error"))

	err := repo.DeleteChunksByDocumentID(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete embeddings")
}

func TestEmbeddingDeleteChunksByDocumentID_ChunkDeleteError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ai_embeddings")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ai_document_chunks")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("delete error"))

	err := repo.DeleteChunksByDocumentID(context.Background(), 1)
	assert.Error(t, err)
}

// ---- CreateEmbedding ----

func TestEmbeddingCreateEmbedding_Success(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	emb := &entities.Embedding{ChunkID: 1, Embedding: []float32{0.1, 0.2}, Model: "text-embedding-3-small"}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO ai_embeddings")).
		WithArgs(emb.ChunkID, sqlmock.AnyArg(), emb.Model, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.CreateEmbedding(context.Background(), emb)
	require.NoError(t, err)
	assert.Equal(t, int64(1), emb.ID)
}

func TestEmbeddingCreateEmbedding_Error(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	emb := &entities.Embedding{ChunkID: 1, Embedding: []float32{0.1}, Model: "m"}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO ai_embeddings")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert error"))

	err := repo.CreateEmbedding(context.Background(), emb)
	assert.Error(t, err)
}

// ---- CreateEmbeddings ----

func TestEmbeddingCreateEmbeddings_Empty(t *testing.T) {
	repo, _ := newEmbRepoMock(t)
	err := repo.CreateEmbeddings(context.Background(), []entities.Embedding{})
	require.NoError(t, err)
}

func TestEmbeddingCreateEmbeddings_Success(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	embs := []entities.Embedding{
		{ChunkID: 1, Embedding: []float32{0.1}, Model: "m"},
		{ChunkID: 2, Embedding: []float32{0.2}, Model: "m"},
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ai_embeddings")).
		WithArgs(int64(1), sqlmock.AnyArg(), "m", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ai_embeddings")).
		WithArgs(int64(2), sqlmock.AnyArg(), "m", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.CreateEmbeddings(context.Background(), embs)
	require.NoError(t, err)
}

func TestEmbeddingCreateEmbeddings_Error(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	embs := []entities.Embedding{{ChunkID: 1, Embedding: []float32{0.1}, Model: "m"}}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ai_embeddings")).
		WithArgs(int64(1), sqlmock.AnyArg(), "m", sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("exec error"))

	err := repo.CreateEmbeddings(context.Background(), embs)
	assert.Error(t, err)
}

// ---- SearchSimilar ----

func TestEmbeddingSearchSimilar_Success(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	now := time.Now()

	cols := []string{"id", "document_id", "chunk_index", "chunk_text", "chunk_tokens", "page_number", "metadata", "created_at", "document_title", "similarity"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(sqlmock.AnyArg(), 0.5, 10).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(int64(1), int64(1), 0, "text", nil, nil, []byte(`{"k":"v"}`), now, "Doc", 0.9))

	results, err := repo.SearchSimilar(context.Background(), []float32{0.1, 0.2}, 10, 0.5)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, 0.9, results[0].SimilarityScore)
}

func TestEmbeddingSearchSimilar_QueryError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(sqlmock.AnyArg(), 0.5, 10).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.SearchSimilar(context.Background(), []float32{0.1}, 10, 0.5)
	assert.Error(t, err)
}

func TestEmbeddingSearchSimilar_ScanError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(sqlmock.AnyArg(), 0.5, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.SearchSimilar(context.Background(), []float32{0.1}, 10, 0.5)
	assert.Error(t, err)
}

// ---- SearchSimilarByDocumentTypes ----

func TestEmbeddingSearchSimilarByDocumentTypes_Success(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	now := time.Now()

	cols := []string{"id", "document_id", "chunk_index", "chunk_text", "chunk_tokens", "page_number", "metadata", "created_at", "document_title", "similarity"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(sqlmock.AnyArg(), 0.5, sqlmock.AnyArg(), 10).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(int64(1), int64(1), 0, "text", nil, nil, nil, now, "Doc", 0.85))

	results, err := repo.SearchSimilarByDocumentTypes(context.Background(), []float32{0.1}, []string{"type1"}, 10, 0.5)
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestEmbeddingSearchSimilarByDocumentTypes_QueryError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(sqlmock.AnyArg(), 0.5, sqlmock.AnyArg(), 10).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.SearchSimilarByDocumentTypes(context.Background(), []float32{0.1}, []string{"type1"}, 10, 0.5)
	assert.Error(t, err)
}

func TestEmbeddingSearchSimilarByDocumentTypes_ScanError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(sqlmock.AnyArg(), 0.5, sqlmock.AnyArg(), 10).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.SearchSimilarByDocumentTypes(context.Background(), []float32{0.1}, []string{"type1"}, 10, 0.5)
	assert.Error(t, err)
}

// ---- GetIndexStatus ----

func TestEmbeddingGetIndexStatus_Success(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	now := time.Now()

	cols := []string{"document_id", "status", "chunks_count", "error_message", "indexed_at", "created_at", "updated_at"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT document_id, status, chunks_count")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(int64(1), "indexed", 10, nil, &now, now, now))

	status, err := repo.GetIndexStatus(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, status)
	assert.Equal(t, entities.IndexStatus("indexed"), status.Status)
}

func TestEmbeddingGetIndexStatus_NotFound(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT document_id, status")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	status, err := repo.GetIndexStatus(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, status)
}

func TestEmbeddingGetIndexStatus_DBError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT document_id, status")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetIndexStatus(context.Background(), 1)
	assert.Error(t, err)
}

// ---- SetIndexStatus ----

func TestEmbeddingSetIndexStatus_Success(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	status := &entities.DocumentIndexStatus{DocumentID: 1, Status: entities.IndexStatusIndexed, ChunksCount: 5}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ai_document_index_status")).
		WithArgs(status.DocumentID, status.Status, status.ChunksCount, status.ErrorMessage, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SetIndexStatus(context.Background(), status)
	require.NoError(t, err)
}

func TestEmbeddingSetIndexStatus_Pending(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	status := &entities.DocumentIndexStatus{DocumentID: 1, Status: entities.IndexStatusPending, ChunksCount: 0}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ai_document_index_status")).
		WithArgs(status.DocumentID, status.Status, status.ChunksCount, status.ErrorMessage, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SetIndexStatus(context.Background(), status)
	require.NoError(t, err)
}

func TestEmbeddingSetIndexStatus_Error(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	status := &entities.DocumentIndexStatus{DocumentID: 1, Status: entities.IndexStatusIndexed}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ai_document_index_status")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("exec error"))

	err := repo.SetIndexStatus(context.Background(), status)
	assert.Error(t, err)
}

// ---- GetPendingDocuments ----

func TestEmbeddingGetPendingDocuments_Success(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT d.id")).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)).AddRow(int64(2)))

	ids, err := repo.GetPendingDocuments(context.Background(), 10)
	require.NoError(t, err)
	assert.Len(t, ids, 2)
}

func TestEmbeddingGetPendingDocuments_QueryError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT d.id")).
		WithArgs(10).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetPendingDocuments(context.Background(), 10)
	assert.Error(t, err)
}

func TestEmbeddingGetPendingDocuments_ScanError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT d.id")).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.GetPendingDocuments(context.Background(), 10)
	assert.Error(t, err)
}

// ---- GetIndexingStats ----

func TestEmbeddingGetIndexingStats_Success(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM documents")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM ai_document_index_status")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(80))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(20))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT MAX(indexed_at)")).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(now))

	total, indexed, pending, lastIndexed, err := repo.GetIndexingStats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 100, total)
	assert.Equal(t, 80, indexed)
	assert.Equal(t, 20, pending)
	assert.NotNil(t, lastIndexed)
}

func TestEmbeddingGetIndexingStats_NullLastIndexed(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM documents")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM ai_document_index_status")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT MAX(indexed_at)")).
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(nil))

	_, _, _, lastIndexed, err := repo.GetIndexingStats(context.Background())
	require.NoError(t, err)
	assert.Nil(t, lastIndexed)
}

func TestEmbeddingGetIndexingStats_TotalError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM documents")).
		WillReturnError(fmt.Errorf("error"))

	_, _, _, _, err := repo.GetIndexingStats(context.Background())
	assert.Error(t, err)
}

func TestEmbeddingGetIndexingStats_IndexedError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM documents")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM ai_document_index_status")).
		WillReturnError(fmt.Errorf("error"))

	_, _, _, _, err := repo.GetIndexingStats(context.Background())
	assert.Error(t, err)
}

func TestEmbeddingGetIndexingStats_PendingError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM documents")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM ai_document_index_status")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnError(fmt.Errorf("error"))

	_, _, _, _, err := repo.GetIndexingStats(context.Background())
	assert.Error(t, err)
}

func TestEmbeddingGetIndexingStats_LastIndexedError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM documents")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM ai_document_index_status")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT MAX(indexed_at)")).
		WillReturnError(fmt.Errorf("error"))

	_, _, _, _, err := repo.GetIndexingStats(context.Background())
	assert.Error(t, err)
}

// ---- SearchHybrid ----

func TestEmbeddingSearchHybrid_Success(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	now := time.Now()

	cols := []string{"id", "document_id", "chunk_index", "chunk_text", "chunk_tokens", "page_number", "metadata", "created_at", "document_title", "similarity"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(sqlmock.AnyArg(), 0.5, 10, "test query").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(int64(1), int64(1), 0, "text", nil, nil, nil, now, "Doc", 0.03))

	results, err := repo.SearchHybrid(context.Background(), []float32{0.1}, "test query", 10, 0.5)
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestEmbeddingSearchHybrid_QueryError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(sqlmock.AnyArg(), 0.5, 10, "query").
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.SearchHybrid(context.Background(), []float32{0.1}, "query", 10, 0.5)
	assert.Error(t, err)
}

func TestEmbeddingSearchHybrid_ScanError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(sqlmock.AnyArg(), 0.5, 10, "query").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.SearchHybrid(context.Background(), []float32{0.1}, "query", 10, 0.5)
	assert.Error(t, err)
}

// ---- GetAdjacentChunks ----

func TestEmbeddingGetAdjacentChunks_Empty(t *testing.T) {
	repo, _ := newEmbRepoMock(t)

	chunks, err := repo.GetAdjacentChunks(context.Background(), []int64{}, 1)
	require.NoError(t, err)
	assert.Nil(t, chunks)
}

func TestEmbeddingGetAdjacentChunks_ZeroWindow(t *testing.T) {
	repo, _ := newEmbRepoMock(t)

	chunks, err := repo.GetAdjacentChunks(context.Background(), []int64{1}, 0)
	require.NoError(t, err)
	assert.Nil(t, chunks)
}

func TestEmbeddingGetAdjacentChunks_Success(t *testing.T) {
	repo, mock := newEmbRepoMock(t)
	now := time.Now()

	cols := []string{"id", "document_id", "chunk_index", "chunk_text", "chunk_tokens", "page_number", "metadata", "created_at"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT")).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(int64(2), int64(1), 1, "adjacent", nil, nil, []byte(`{"k":"v"}`), now))

	chunks, err := repo.GetAdjacentChunks(context.Background(), []int64{1}, 1)
	require.NoError(t, err)
	assert.Len(t, chunks, 1)
}

func TestEmbeddingGetAdjacentChunks_QueryError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT")).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetAdjacentChunks(context.Background(), []int64{1}, 1)
	assert.Error(t, err)
}

func TestEmbeddingGetAdjacentChunks_ScanError(t *testing.T) {
	repo, mock := newEmbRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT")).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.GetAdjacentChunks(context.Background(), []int64{1}, 1)
	assert.Error(t, err)
}
