package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

// ============================================================
// Helper fixture for EmbeddingUseCase tests
// ============================================================

type embeddingTestFixture struct {
	embRepo     *MockEmbeddingRepo
	embProvider *MockEmbeddingProvider
	docProvider *MockDocumentProvider
	useCase     *EmbeddingUseCase
}

func newEmbeddingTestFixture(chunkCfg ...services.ChunkingConfig) *embeddingTestFixture {
	f := &embeddingTestFixture{
		embRepo:     new(MockEmbeddingRepo),
		embProvider: new(MockEmbeddingProvider),
		docProvider: new(MockDocumentProvider),
	}
	if len(chunkCfg) > 0 {
		f.useCase = NewEmbeddingUseCase(f.embRepo, f.embProvider, f.docProvider, nil, "test-model", chunkCfg[0])
	} else {
		f.useCase = NewEmbeddingUseCase(f.embRepo, f.embProvider, f.docProvider, nil, "test-model")
	}
	return f
}

// ============================================================
// Tests: NewEmbeddingUseCase
// ============================================================

func TestNewEmbeddingUseCase_DefaultModel(t *testing.T) {
	embRepo := new(MockEmbeddingRepo)
	embProvider := new(MockEmbeddingProvider)
	docProvider := new(MockDocumentProvider)
	uc := NewEmbeddingUseCase(embRepo, embProvider, docProvider, nil, "")
	assert.Equal(t, "text-embedding-3-small", uc.embeddingModel)
}

func TestNewEmbeddingUseCase_CustomModel(t *testing.T) {
	embRepo := new(MockEmbeddingRepo)
	embProvider := new(MockEmbeddingProvider)
	docProvider := new(MockDocumentProvider)
	uc := NewEmbeddingUseCase(embRepo, embProvider, docProvider, nil, "custom-model")
	assert.Equal(t, "custom-model", uc.embeddingModel)
}

func TestNewEmbeddingUseCase_CustomChunkingConfig(t *testing.T) {
	cfg := services.ChunkingConfig{MaxTokens: 256, OverlapRatio: 0.1}
	f := newEmbeddingTestFixture(cfg)
	assert.NotNil(t, f.useCase.chunkingService)
}

// ============================================================
// Tests: SetCache
// ============================================================

func TestSetCache(t *testing.T) {
	f := newEmbeddingTestFixture()
	assert.Nil(t, f.useCase.cache)
	// SetCache with nil is a valid operation
	f.useCase.SetCache(nil)
	assert.Nil(t, f.useCase.cache)
}

// ============================================================
// Tests: queryEmbeddingCacheKey
// ============================================================

func TestQueryEmbeddingCacheKey_Deterministic(t *testing.T) {
	key1 := queryEmbeddingCacheKey("test query", "model-1")
	key2 := queryEmbeddingCacheKey("test query", "model-1")
	assert.Equal(t, key1, key2)
}

func TestQueryEmbeddingCacheKey_DifferentQueries(t *testing.T) {
	key1 := queryEmbeddingCacheKey("query 1", "model-1")
	key2 := queryEmbeddingCacheKey("query 2", "model-1")
	assert.NotEqual(t, key1, key2)
}

func TestQueryEmbeddingCacheKey_DifferentModels(t *testing.T) {
	key1 := queryEmbeddingCacheKey("query", "model-1")
	key2 := queryEmbeddingCacheKey("query", "model-2")
	assert.NotEqual(t, key1, key2)
}

func TestQueryEmbeddingCacheKey_HasPrefix(t *testing.T) {
	key := queryEmbeddingCacheKey("query", "model")
	assert.Contains(t, key, embeddingCachePrefix)
}

// ============================================================
// Tests: cachedQueryEmbedding (without cache)
// ============================================================

func TestCachedQueryEmbedding_NoCache(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embProvider.On("GenerateQueryEmbedding", "test").Return([]float32{0.1, 0.2, 0.3}, nil)

	emb, err := f.useCase.cachedQueryEmbedding(context.Background(), "test")
	require.NoError(t, err)
	assert.Equal(t, []float32{0.1, 0.2, 0.3}, emb)
}

func TestCachedQueryEmbedding_ProviderError(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embProvider.On("GenerateQueryEmbedding", "test").Return(nil, errors.New("provider error"))

	emb, err := f.useCase.cachedQueryEmbedding(context.Background(), "test")
	assert.Nil(t, emb)
	assert.Error(t, err)
}

// ============================================================
// Tests: IndexDocument
// ============================================================

func TestIndexDocument_AlreadyIndexed_NoForce(t *testing.T) {
	f := newEmbeddingTestFixture()
	status := &entities.DocumentIndexStatus{
		DocumentID:  1,
		Status:      entities.IndexStatusIndexed,
		ChunksCount: 5,
	}
	f.embRepo.On("GetIndexStatus", int64(1)).Return(status, nil)

	resp, err := f.useCase.IndexDocument(context.Background(), 1, false)
	require.NoError(t, err)
	assert.Equal(t, "already_indexed", resp.Status)
	assert.Equal(t, 5, resp.ChunksCreated)
}

func TestIndexDocument_AlreadyIndexed_ForceReindex(t *testing.T) {
	f := newEmbeddingTestFixture()
	status := &entities.DocumentIndexStatus{
		DocumentID: 1,
		Status:     entities.IndexStatusIndexed,
	}
	f.embRepo.On("GetIndexStatus", int64(1)).Return(status, nil)
	f.embRepo.On("SetIndexStatus", mock.AnythingOfType("*entities.DocumentIndexStatus")).Return(nil)
	f.docProvider.On("GetDocumentContent", int64(1)).Return("Some document text content here.", "Title", nil)
	f.embRepo.On("DeleteChunksByDocumentID", int64(1)).Return(nil)
	f.embRepo.On("CreateChunks", mock.Anything).Return(nil)

	chunks := []entities.DocumentChunk{
		{ID: 10, DocumentID: 1, ChunkIndex: 0, ChunkText: "Some document text content here."},
	}
	f.embRepo.On("GetChunksByDocumentID", int64(1)).Return(chunks, nil)
	f.embProvider.On("GenerateEmbeddings", mock.Anything).Return([][]float32{{0.1, 0.2}}, nil)
	f.embRepo.On("CreateEmbeddings", mock.Anything).Return(nil)

	resp, err := f.useCase.IndexDocument(context.Background(), 1, true)
	require.NoError(t, err)
	assert.Equal(t, "indexed", resp.Status)
	assert.Equal(t, 1, resp.ChunksCreated)
}

func TestIndexDocument_NewDocument_Success(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embRepo.On("GetIndexStatus", int64(1)).Return(nil, errors.New("not found"))
	f.embRepo.On("SetIndexStatus", mock.AnythingOfType("*entities.DocumentIndexStatus")).Return(nil)
	f.docProvider.On("GetDocumentContent", int64(1)).Return("Document content.", "Title", nil)
	f.embRepo.On("CreateChunks", mock.Anything).Return(nil)
	chunks := []entities.DocumentChunk{
		{ID: 10, DocumentID: 1, ChunkIndex: 0, ChunkText: "Document content."},
	}
	f.embRepo.On("GetChunksByDocumentID", int64(1)).Return(chunks, nil)
	f.embProvider.On("GenerateEmbeddings", mock.Anything).Return([][]float32{{0.1}}, nil)
	f.embRepo.On("CreateEmbeddings", mock.Anything).Return(nil)

	resp, err := f.useCase.IndexDocument(context.Background(), 1, false)
	require.NoError(t, err)
	assert.Equal(t, "indexed", resp.Status)
	assert.Equal(t, 1, resp.ChunksCreated)
}

func TestIndexDocument_SetIndexStatusError(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embRepo.On("GetIndexStatus", int64(1)).Return(nil, errors.New("not found"))
	f.embRepo.On("SetIndexStatus", mock.Anything).Return(errors.New("db error"))

	resp, err := f.useCase.IndexDocument(context.Background(), 1, false)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to set index status")
}

func TestIndexDocument_GetDocumentContentError(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embRepo.On("GetIndexStatus", int64(1)).Return(nil, errors.New("not found"))
	f.embRepo.On("SetIndexStatus", mock.Anything).Return(nil)
	f.docProvider.On("GetDocumentContent", int64(1)).Return("", "", errors.New("doc not found"))

	resp, err := f.useCase.IndexDocument(context.Background(), 1, false)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to get document content")
}

func TestIndexDocument_EmptyContent(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embRepo.On("GetIndexStatus", int64(1)).Return(nil, errors.New("not found"))
	f.embRepo.On("SetIndexStatus", mock.Anything).Return(nil)
	f.docProvider.On("GetDocumentContent", int64(1)).Return("", "Title", nil)

	resp, err := f.useCase.IndexDocument(context.Background(), 1, false)
	require.NoError(t, err)
	assert.Equal(t, "indexed", resp.Status)
	assert.Equal(t, 0, resp.ChunksCreated)
}

func TestIndexDocument_CreateChunksError(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embRepo.On("GetIndexStatus", int64(1)).Return(nil, errors.New("not found"))
	f.embRepo.On("SetIndexStatus", mock.Anything).Return(nil)
	f.docProvider.On("GetDocumentContent", int64(1)).Return("Some content here.", "Title", nil)
	f.embRepo.On("CreateChunks", mock.Anything).Return(errors.New("db error"))

	resp, err := f.useCase.IndexDocument(context.Background(), 1, false)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to create chunks")
}

func TestIndexDocument_ReloadChunksError(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embRepo.On("GetIndexStatus", int64(1)).Return(nil, errors.New("not found"))
	f.embRepo.On("SetIndexStatus", mock.Anything).Return(nil)
	f.docProvider.On("GetDocumentContent", int64(1)).Return("Some content here.", "Title", nil)
	f.embRepo.On("CreateChunks", mock.Anything).Return(nil)
	f.embRepo.On("GetChunksByDocumentID", int64(1)).Return(nil, errors.New("db error"))

	resp, err := f.useCase.IndexDocument(context.Background(), 1, false)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to reload chunks")
}

func TestIndexDocument_GenerateEmbeddingsError(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embRepo.On("GetIndexStatus", int64(1)).Return(nil, errors.New("not found"))
	f.embRepo.On("SetIndexStatus", mock.Anything).Return(nil)
	f.docProvider.On("GetDocumentContent", int64(1)).Return("Some content here.", "Title", nil)
	f.embRepo.On("CreateChunks", mock.Anything).Return(nil)
	chunks := []entities.DocumentChunk{{ID: 10, DocumentID: 1, ChunkText: "Some content here."}}
	f.embRepo.On("GetChunksByDocumentID", int64(1)).Return(chunks, nil)
	f.embProvider.On("GenerateEmbeddings", mock.Anything).Return(nil, errors.New("provider error"))

	resp, err := f.useCase.IndexDocument(context.Background(), 1, false)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to generate embeddings")
}

func TestIndexDocument_CreateEmbeddingsError(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embRepo.On("GetIndexStatus", int64(1)).Return(nil, errors.New("not found"))
	f.embRepo.On("SetIndexStatus", mock.Anything).Return(nil)
	f.docProvider.On("GetDocumentContent", int64(1)).Return("Some content here.", "Title", nil)
	f.embRepo.On("CreateChunks", mock.Anything).Return(nil)
	chunks := []entities.DocumentChunk{{ID: 10, DocumentID: 1, ChunkText: "Some content here."}}
	f.embRepo.On("GetChunksByDocumentID", int64(1)).Return(chunks, nil)
	f.embProvider.On("GenerateEmbeddings", mock.Anything).Return([][]float32{{0.1}}, nil)
	f.embRepo.On("CreateEmbeddings", mock.Anything).Return(errors.New("db error"))

	resp, err := f.useCase.IndexDocument(context.Background(), 1, false)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to create embeddings")
}

func TestIndexDocument_FinalSetIndexStatusError(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embRepo.On("GetIndexStatus", int64(1)).Return(nil, errors.New("not found"))
	// First SetIndexStatus (indexing) succeeds, second (indexed) fails
	setStatusCallCount := 0
	f.embRepo.On("SetIndexStatus", mock.AnythingOfType("*entities.DocumentIndexStatus")).Return(nil).Run(func(args mock.Arguments) {
		setStatusCallCount++
		if setStatusCallCount >= 2 {
			// We can't change Return after Run, so we use a different approach below
		}
	}).Maybe()
	f.docProvider.On("GetDocumentContent", int64(1)).Return("Content.", "Title", nil)
	f.embRepo.On("CreateChunks", mock.Anything).Return(nil)
	chunks := []entities.DocumentChunk{{ID: 10, DocumentID: 1, ChunkText: "Content."}}
	f.embRepo.On("GetChunksByDocumentID", int64(1)).Return(chunks, nil)
	f.embProvider.On("GenerateEmbeddings", mock.Anything).Return([][]float32{{0.1}}, nil)
	f.embRepo.On("CreateEmbeddings", mock.Anything).Return(nil)

	// Re-setup: clear and re-register with proper sequencing
	f.embRepo.ExpectedCalls = nil
	f.embRepo.On("GetIndexStatus", int64(1)).Return(nil, errors.New("not found"))
	f.embRepo.On("SetIndexStatus", mock.MatchedBy(func(s *entities.DocumentIndexStatus) bool {
		return s.Status == entities.IndexStatusIndexing
	})).Return(nil).Once()
	f.embRepo.On("SetIndexStatus", mock.MatchedBy(func(s *entities.DocumentIndexStatus) bool {
		return s.Status == entities.IndexStatusIndexed
	})).Return(errors.New("final status error")).Once()
	f.docProvider.ExpectedCalls = nil
	f.docProvider.On("GetDocumentContent", int64(1)).Return("Content.", "Title", nil)
	f.embRepo.On("CreateChunks", mock.Anything).Return(nil)
	f.embRepo.On("GetChunksByDocumentID", int64(1)).Return(chunks, nil)
	f.embProvider.ExpectedCalls = nil
	f.embProvider.On("GenerateEmbeddings", mock.Anything).Return([][]float32{{0.1}}, nil)
	f.embRepo.On("CreateEmbeddings", mock.Anything).Return(nil)

	resp, err := f.useCase.IndexDocument(context.Background(), 1, false)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to update index status")
}

func TestIndexDocument_DeleteChunksError_ForceReindex(t *testing.T) {
	f := newEmbeddingTestFixture()
	status := &entities.DocumentIndexStatus{DocumentID: 1, Status: entities.IndexStatusIndexed}
	f.embRepo.On("GetIndexStatus", int64(1)).Return(status, nil)
	f.embRepo.On("SetIndexStatus", mock.Anything).Return(nil)
	f.docProvider.On("GetDocumentContent", int64(1)).Return("Content.", "Title", nil)
	f.embRepo.On("DeleteChunksByDocumentID", int64(1)).Return(errors.New("delete error"))

	resp, err := f.useCase.IndexDocument(context.Background(), 1, true)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to delete existing chunks")
}

// ============================================================
// Tests: SearchSimilar
// ============================================================

func TestSearchSimilar_Success(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	results := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ID: 1, DocumentID: 10, ChunkText: "result"}, SimilarityScore: 0.9},
	}
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, 10, 0.7).Return(results, nil)
	f.embRepo.On("GetAdjacentChunks", mock.Anything, mock.Anything).Return(nil, nil)

	resp, err := f.useCase.SearchSimilar(context.Background(), "query", 10, 0.7)
	require.NoError(t, err)
	assert.Len(t, resp, 1)
}

func TestSearchSimilar_DefaultValues(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, 10, 0.7).Return([]entities.ChunkWithScore{}, nil)

	resp, err := f.useCase.SearchSimilar(context.Background(), "query", 0, 0)
	require.NoError(t, err)
	assert.Empty(t, resp)
}

func TestSearchSimilar_EmbeddingError(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return(nil, errors.New("embed error"))

	resp, err := f.useCase.SearchSimilar(context.Background(), "query", 10, 0.7)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to generate query embedding")
}

func TestSearchSimilar_HybridFallsBackToVector(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("hybrid not supported"))
	results := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ID: 1, ChunkText: "result"}, SimilarityScore: 0.8},
	}
	f.embRepo.On("SearchSimilar", mock.Anything, 10, 0.7).Return(results, nil)
	f.embRepo.On("GetAdjacentChunks", mock.Anything, mock.Anything).Return(nil, nil)

	resp, err := f.useCase.SearchSimilar(context.Background(), "query", 10, 0.7)
	require.NoError(t, err)
	assert.Len(t, resp, 1)
}

func TestSearchSimilar_BothSearchesFail(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("hybrid error"))
	f.embRepo.On("SearchSimilar", mock.Anything, 10, 0.7).Return(nil, errors.New("vector error"))

	resp, err := f.useCase.SearchSimilar(context.Background(), "query", 10, 0.7)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to search similar chunks")
}

func TestSearchSimilar_EmptyResults(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entities.ChunkWithScore{}, nil)

	resp, err := f.useCase.SearchSimilar(context.Background(), "query", 10, 0.7)
	require.NoError(t, err)
	assert.Empty(t, resp)
}

func TestSearchSimilar_LimitsResults(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	// Return more results than requested limit
	results := make([]entities.ChunkWithScore, 5)
	for i := range results {
		results[i] = entities.ChunkWithScore{
			Chunk:           &entities.DocumentChunk{ID: int64(i), DocumentID: 1, ChunkText: "text"},
			SimilarityScore: float64(5-i) / 10.0,
		}
	}
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, 2, mock.Anything).Return(results, nil)
	f.embRepo.On("GetAdjacentChunks", mock.Anything, mock.Anything).Return(nil, nil)

	resp, err := f.useCase.SearchSimilar(context.Background(), "query", 2, 0.7)
	require.NoError(t, err)
	assert.Len(t, resp, 2)
}

func TestSearchSimilar_WithAdjacentChunks(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embProvider.On("GenerateQueryEmbedding", mock.Anything).Return([]float32{0.1}, nil)
	results := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ID: 10, DocumentID: 1, ChunkIndex: 1, ChunkText: "main chunk"}, SimilarityScore: 0.9},
	}
	f.embRepo.On("SearchHybrid", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(results, nil)
	adjacent := []entities.DocumentChunk{
		{ID: 9, DocumentID: 1, ChunkIndex: 0, ChunkText: "prev chunk"},
		{ID: 11, DocumentID: 1, ChunkIndex: 2, ChunkText: "next chunk"},
	}
	f.embRepo.On("GetAdjacentChunks", mock.Anything, mock.Anything).Return(adjacent, nil)

	resp, err := f.useCase.SearchSimilar(context.Background(), "query", 10, 0.7)
	require.NoError(t, err)
	assert.Len(t, resp, 1)
	// Chunk text should include adjacent chunks merged
	assert.Contains(t, resp[0].Chunk.ChunkText, "prev chunk")
	assert.Contains(t, resp[0].Chunk.ChunkText, "main chunk")
	assert.Contains(t, resp[0].Chunk.ChunkText, "next chunk")
}

// ============================================================
// Tests: rerankByKeywords
// ============================================================

func TestRerankByKeywords_EmptyResults(t *testing.T) {
	result := rerankByKeywords(nil, "query")
	assert.Nil(t, result)
}

func TestRerankByKeywords_EmptyQuery(t *testing.T) {
	results := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ChunkText: "text"}, SimilarityScore: 0.9},
	}
	result := rerankByKeywords(results, "")
	assert.Equal(t, results, result)
}

func TestRerankByKeywords_ShortWordsSkipped(t *testing.T) {
	results := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ChunkText: "some text here"}, SimilarityScore: 0.9},
	}
	// All query words are <= 2 runes
	result := rerankByKeywords(results, "ab cd")
	assert.Equal(t, 0.9, result[0].SimilarityScore)
}

func TestRerankByKeywords_BoostsMatchingChunks(t *testing.T) {
	results := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ChunkText: "unrelated content"}, SimilarityScore: 0.8},
		{Chunk: &entities.DocumentChunk{ChunkText: "document about education system"}, SimilarityScore: 0.7},
	}
	result := rerankByKeywords(results, "education system")
	// The second result should be boosted and possibly reordered
	assert.Equal(t, 2, len(result))
	// The chunk with matching keywords should have higher score after reranking
	if result[0].Chunk.ChunkText == "document about education system" {
		assert.Greater(t, result[0].SimilarityScore, result[1].SimilarityScore)
	}
}

func TestRerankByKeywords_SortsDescending(t *testing.T) {
	results := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ChunkText: "no match here"}, SimilarityScore: 0.5},
		{Chunk: &entities.DocumentChunk{ChunkText: "contains important keyword"}, SimilarityScore: 0.5},
	}
	result := rerankByKeywords(results, "important keyword")
	assert.GreaterOrEqual(t, result[0].SimilarityScore, result[1].SimilarityScore)
}

// ============================================================
// Tests: expandWithAdjacentChunks
// ============================================================

func TestExpandWithAdjacentChunks_NoAdjacentFound(t *testing.T) {
	f := newEmbeddingTestFixture()
	results := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ID: 1, DocumentID: 1, ChunkIndex: 0, ChunkText: "main"}, SimilarityScore: 0.9},
	}
	f.embRepo.On("GetAdjacentChunks", mock.Anything, mock.Anything).Return(nil, nil)

	expanded := f.useCase.expandWithAdjacentChunks(context.Background(), results)
	assert.Equal(t, "main", expanded[0].Chunk.ChunkText)
}

func TestExpandWithAdjacentChunks_ErrorFallsBack(t *testing.T) {
	f := newEmbeddingTestFixture()
	results := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ID: 1, DocumentID: 1, ChunkIndex: 0, ChunkText: "main"}, SimilarityScore: 0.9},
	}
	f.embRepo.On("GetAdjacentChunks", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	expanded := f.useCase.expandWithAdjacentChunks(context.Background(), results)
	assert.Equal(t, "main", expanded[0].Chunk.ChunkText)
}

func TestExpandWithAdjacentChunks_MergesPrevAndNext(t *testing.T) {
	f := newEmbeddingTestFixture()
	results := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ID: 5, DocumentID: 1, ChunkIndex: 1, ChunkText: "center"}, SimilarityScore: 0.9},
	}
	adjacent := []entities.DocumentChunk{
		{ID: 4, DocumentID: 1, ChunkIndex: 0, ChunkText: "before"},
		{ID: 6, DocumentID: 1, ChunkIndex: 2, ChunkText: "after"},
	}
	f.embRepo.On("GetAdjacentChunks", mock.Anything, mock.Anything).Return(adjacent, nil)

	expanded := f.useCase.expandWithAdjacentChunks(context.Background(), results)
	assert.Contains(t, expanded[0].Chunk.ChunkText, "before")
	assert.Contains(t, expanded[0].Chunk.ChunkText, "center")
	assert.Contains(t, expanded[0].Chunk.ChunkText, "after")
}

func TestExpandWithAdjacentChunks_OnlyPrev(t *testing.T) {
	f := newEmbeddingTestFixture()
	results := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ID: 5, DocumentID: 1, ChunkIndex: 1, ChunkText: "center"}, SimilarityScore: 0.9},
	}
	adjacent := []entities.DocumentChunk{
		{ID: 4, DocumentID: 1, ChunkIndex: 0, ChunkText: "before"},
	}
	f.embRepo.On("GetAdjacentChunks", mock.Anything, mock.Anything).Return(adjacent, nil)

	expanded := f.useCase.expandWithAdjacentChunks(context.Background(), results)
	assert.Contains(t, expanded[0].Chunk.ChunkText, "before")
	assert.Contains(t, expanded[0].Chunk.ChunkText, "center")
	assert.NotContains(t, expanded[0].Chunk.ChunkText, "after")
}

// ============================================================
// Tests: Search
// ============================================================

func TestSearch_Success_NoDocumentTypes(t *testing.T) {
	f := newEmbeddingTestFixture()
	req := &dto.SearchRequest{Query: "test query", Limit: 5, Threshold: 0.8}
	f.embProvider.On("GenerateQueryEmbedding", "test query").Return([]float32{0.1}, nil)
	results := []entities.ChunkWithScore{
		{Chunk: &entities.DocumentChunk{ID: 1, DocumentID: 10, ChunkText: "match"}, SimilarityScore: 0.85, DocumentTitle: "Doc"},
	}
	f.embRepo.On("SearchSimilar", mock.Anything, 5, 0.8).Return(results, nil)

	resp, err := f.useCase.Search(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Equal(t, "test query", resp.Query)
}

func TestSearch_WithDocumentTypes(t *testing.T) {
	f := newEmbeddingTestFixture()
	req := &dto.SearchRequest{Query: "test", Limit: 5, Threshold: 0.8, DocumentTypes: []string{"report"}}
	f.embProvider.On("GenerateQueryEmbedding", "test").Return([]float32{0.1}, nil)
	f.embRepo.On("SearchSimilarByDocumentTypes", mock.Anything, []string{"report"}, 5, 0.8).Return([]entities.ChunkWithScore{}, nil)

	resp, err := f.useCase.Search(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
}

func TestSearch_DefaultLimitAndThreshold(t *testing.T) {
	f := newEmbeddingTestFixture()
	req := &dto.SearchRequest{Query: "test"}
	f.embProvider.On("GenerateQueryEmbedding", "test").Return([]float32{0.1}, nil)
	f.embRepo.On("SearchSimilar", mock.Anything, 10, 0.7).Return([]entities.ChunkWithScore{}, nil)

	resp, err := f.useCase.Search(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestSearch_EmbeddingError(t *testing.T) {
	f := newEmbeddingTestFixture()
	req := &dto.SearchRequest{Query: "test"}
	f.embProvider.On("GenerateQueryEmbedding", "test").Return(nil, errors.New("embed error"))

	resp, err := f.useCase.Search(context.Background(), req)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to generate query embedding")
}

func TestSearch_SearchError(t *testing.T) {
	f := newEmbeddingTestFixture()
	req := &dto.SearchRequest{Query: "test"}
	f.embProvider.On("GenerateQueryEmbedding", "test").Return([]float32{0.1}, nil)
	f.embRepo.On("SearchSimilar", mock.Anything, 10, 0.7).Return(nil, errors.New("search error"))

	resp, err := f.useCase.Search(context.Background(), req)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to search")
}

// ============================================================
// Tests: GetIndexingStatus
// ============================================================

func TestGetIndexingStatus_Success(t *testing.T) {
	f := newEmbeddingTestFixture()
	lastIdx := "2024-01-01"
	f.embRepo.On("GetIndexingStats").Return(100, 80, 20, &lastIdx, nil)

	resp, err := f.useCase.GetIndexingStatus(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 100, resp.TotalDocuments)
	assert.Equal(t, 80, resp.IndexedDocuments)
	assert.Equal(t, 20, resp.PendingDocuments)
	assert.Equal(t, &lastIdx, resp.LastIndexedAt)
}

func TestGetIndexingStatus_Error(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embRepo.On("GetIndexingStats").Return(0, 0, 0, (*string)(nil), errors.New("db error"))

	resp, err := f.useCase.GetIndexingStatus(context.Background())
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "failed to get indexing stats")
}

// ============================================================
// Tests: IndexPendingDocuments
// ============================================================

func TestIndexPendingDocuments_Success(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embRepo.On("GetPendingDocuments", 10).Return([]int64{1, 2}, nil)
	// Set up IndexDocument for doc 1 - already indexed
	f.embRepo.On("GetIndexStatus", int64(1)).Return(&entities.DocumentIndexStatus{
		DocumentID: 1, Status: entities.IndexStatusIndexed, ChunksCount: 3,
	}, nil)
	// Set up IndexDocument for doc 2 - already indexed
	f.embRepo.On("GetIndexStatus", int64(2)).Return(&entities.DocumentIndexStatus{
		DocumentID: 2, Status: entities.IndexStatusIndexed, ChunksCount: 5,
	}, nil)

	indexed, err := f.useCase.IndexPendingDocuments(context.Background(), 10)
	require.NoError(t, err)
	assert.Equal(t, 2, indexed)
}

func TestIndexPendingDocuments_DefaultBatchSize(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embRepo.On("GetPendingDocuments", 10).Return([]int64{}, nil)

	indexed, err := f.useCase.IndexPendingDocuments(context.Background(), 0)
	require.NoError(t, err)
	assert.Equal(t, 0, indexed)
}

func TestIndexPendingDocuments_GetPendingError(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embRepo.On("GetPendingDocuments", 10).Return(nil, errors.New("db error"))

	indexed, err := f.useCase.IndexPendingDocuments(context.Background(), 10)
	assert.Equal(t, 0, indexed)
	assert.ErrorContains(t, err, "failed to get pending documents")
}

func TestIndexPendingDocuments_PartialFailure(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embRepo.On("GetPendingDocuments", 10).Return([]int64{1, 2}, nil)
	// Doc 1 succeeds (already indexed)
	f.embRepo.On("GetIndexStatus", int64(1)).Return(&entities.DocumentIndexStatus{
		DocumentID: 1, Status: entities.IndexStatusIndexed, ChunksCount: 3,
	}, nil)
	// Doc 2 fails - SetIndexStatus error
	f.embRepo.On("GetIndexStatus", int64(2)).Return(nil, errors.New("not found"))
	f.embRepo.On("SetIndexStatus", mock.MatchedBy(func(s *entities.DocumentIndexStatus) bool {
		return s.DocumentID == 2
	})).Return(errors.New("db error"))

	indexed, err := f.useCase.IndexPendingDocuments(context.Background(), 10)
	require.NoError(t, err)
	assert.Equal(t, 1, indexed)
}

func TestIndexPendingDocuments_NoPending(t *testing.T) {
	f := newEmbeddingTestFixture()
	f.embRepo.On("GetPendingDocuments", 5).Return([]int64{}, nil)

	indexed, err := f.useCase.IndexPendingDocuments(context.Background(), 5)
	require.NoError(t, err)
	assert.Equal(t, 0, indexed)
}
