// Package usecases contains application use cases for the AI module.
package usecases

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/ports"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/cache"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// Deprecated: Use ports.EmbeddingProvider directly. This alias exists for backward compatibility.
type EmbeddingProvider = ports.EmbeddingProvider

// DocumentProvider defines the interface for accessing document content
type DocumentProvider interface {
	// GetDocumentContent retrieves the text content of a document
	GetDocumentContent(ctx context.Context, documentID int64) (string, string, error) // returns content, title, error
}

const (
	embeddingCachePrefix = "ai:embedding:query:"
	embeddingCacheTTL    = 24 * time.Hour
)

// EmbeddingUseCase handles document embedding and search operations
type EmbeddingUseCase struct {
	embeddingRepo     repositories.EmbeddingRepository
	embeddingProvider EmbeddingProvider
	documentProvider  DocumentProvider
	chunkingService   *services.ChunkingService
	auditLogger       *logging.AuditLogger
	cache             *cache.RedisCache
	embeddingModel    string
}

// NewEmbeddingUseCase creates a new EmbeddingUseCase
func NewEmbeddingUseCase(
	embeddingRepo repositories.EmbeddingRepository,
	embeddingProvider EmbeddingProvider,
	documentProvider DocumentProvider,
	auditLogger *logging.AuditLogger,
	embeddingModel string,
	chunkingConfig ...services.ChunkingConfig,
) *EmbeddingUseCase {
	if embeddingModel == "" {
		embeddingModel = "text-embedding-3-small"
	}
	cfg := services.DefaultChunkingConfig()
	if len(chunkingConfig) > 0 {
		cfg = chunkingConfig[0]
	}
	return &EmbeddingUseCase{
		embeddingRepo:     embeddingRepo,
		embeddingProvider: embeddingProvider,
		documentProvider:  documentProvider,
		chunkingService:   services.NewChunkingService(cfg),
		auditLogger:       auditLogger,
		embeddingModel:    embeddingModel,
	}
}

// SetCache sets the Redis cache for query embedding caching
func (uc *EmbeddingUseCase) SetCache(c *cache.RedisCache) {
	uc.cache = c
}

// queryEmbeddingCacheKey returns a deterministic cache key for a query string and model.
func queryEmbeddingCacheKey(query, model string) string {
	h := sha256.Sum256([]byte(query))
	return fmt.Sprintf("%s%s:%x", embeddingCachePrefix, model, h[:16])
}

// cachedQueryEmbedding tries to get query embedding from cache, falling back to the provider.
func (uc *EmbeddingUseCase) cachedQueryEmbedding(ctx context.Context, query string) ([]float32, error) {
	cacheKey := queryEmbeddingCacheKey(query, uc.embeddingModel)

	// Try cache
	if uc.cache != nil {
		var cached []float32
		found, err := uc.cache.Get(ctx, cacheKey, &cached)
		if err == nil && found && len(cached) > 0 {
			return cached, nil
		}
	}

	// Generate via provider
	embedding, err := uc.embeddingProvider.GenerateQueryEmbedding(ctx, query)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if uc.cache != nil {
		if err := uc.cache.Set(ctx, cacheKey, embedding, embeddingCacheTTL); err != nil {
			slog.Warn("failed to cache query embedding", "error", err)
		}
	}

	return embedding, nil
}

// IndexDocument indexes a document by chunking and generating embeddings
func (uc *EmbeddingUseCase) IndexDocument(ctx context.Context, documentID int64, forceReindex bool) (*dto.IndexDocumentResponse, error) {
	// Check existing status
	status, err := uc.embeddingRepo.GetIndexStatus(ctx, documentID)
	if err == nil && status != nil {
		if !forceReindex && status.Status == entities.IndexStatusIndexed {
			return &dto.IndexDocumentResponse{
				DocumentID:    documentID,
				ChunksCreated: status.ChunksCount,
				Status:        "already_indexed",
				Message:       "Document is already indexed",
			}, nil
		}
	}

	// Set status to indexing
	indexStatus := &entities.DocumentIndexStatus{
		DocumentID: documentID,
		Status:     entities.IndexStatusIndexing,
	}
	if err := uc.embeddingRepo.SetIndexStatus(ctx, indexStatus); err != nil {
		return nil, fmt.Errorf("failed to set index status: %w", err)
	}

	// Get document content
	content, _, err := uc.documentProvider.GetDocumentContent(ctx, documentID)
	if err != nil {
		errMsg := err.Error()
		indexStatus.Status = entities.IndexStatusFailed
		indexStatus.ErrorMessage = &errMsg
		_ = uc.embeddingRepo.SetIndexStatus(ctx, indexStatus) // best-effort status update
		return nil, fmt.Errorf("failed to get document content: %w", err)
	}

	// Delete existing chunks if reindexing
	if forceReindex {
		if err := uc.embeddingRepo.DeleteChunksByDocumentID(ctx, documentID); err != nil {
			return nil, fmt.Errorf("failed to delete existing chunks: %w", err)
		}
	}

	// Chunk the document
	chunks := uc.chunkingService.ChunkDocument(documentID, content)
	if len(chunks) == 0 {
		indexStatus.Status = entities.IndexStatusIndexed
		indexStatus.ChunksCount = 0
		_ = uc.embeddingRepo.SetIndexStatus(ctx, indexStatus)
		return &dto.IndexDocumentResponse{
			DocumentID:    documentID,
			ChunksCreated: 0,
			Status:        "indexed",
			Message:       "Document has no content to index",
		}, nil
	}

	// Create chunks in database
	if err := uc.embeddingRepo.CreateChunks(ctx, chunks); err != nil {
		errMsg := err.Error()
		indexStatus.Status = entities.IndexStatusFailed
		indexStatus.ErrorMessage = &errMsg
		_ = uc.embeddingRepo.SetIndexStatus(ctx, indexStatus) // best-effort status update
		return nil, fmt.Errorf("failed to create chunks: %w", err)
	}

	// Reload chunks to get IDs
	chunks, err = uc.embeddingRepo.GetChunksByDocumentID(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload chunks: %w", err)
	}

	// Extract texts for embedding
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.ChunkText
	}

	// Generate embeddings
	vectors, err := uc.embeddingProvider.GenerateEmbeddings(ctx, texts)
	if err != nil {
		errMsg := err.Error()
		indexStatus.Status = entities.IndexStatusFailed
		indexStatus.ErrorMessage = &errMsg
		_ = uc.embeddingRepo.SetIndexStatus(ctx, indexStatus) // best-effort status update
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Create embeddings in database
	embeddings := make([]entities.Embedding, len(chunks))
	for i, chunk := range chunks {
		embeddings[i] = *entities.NewEmbedding(chunk.ID, vectors[i], uc.embeddingModel)
	}

	if err := uc.embeddingRepo.CreateEmbeddings(ctx, embeddings); err != nil {
		errMsg := err.Error()
		indexStatus.Status = entities.IndexStatusFailed
		indexStatus.ErrorMessage = &errMsg
		_ = uc.embeddingRepo.SetIndexStatus(ctx, indexStatus) // best-effort status update
		return nil, fmt.Errorf("failed to create embeddings: %w", err)
	}

	// Update status to indexed
	indexStatus.Status = entities.IndexStatusIndexed
	indexStatus.ChunksCount = len(chunks)
	if err := uc.embeddingRepo.SetIndexStatus(ctx, indexStatus); err != nil {
		return nil, fmt.Errorf("failed to update index status: %w", err)
	}

	// Log audit event
	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "create", "ai_index", map[string]any{
			"document_id":    documentID,
			"chunks_created": len(chunks),
		})
	}

	return &dto.IndexDocumentResponse{
		DocumentID:    documentID,
		ChunksCreated: len(chunks),
		Status:        "indexed",
		Message:       fmt.Sprintf("Document indexed successfully with %d chunks", len(chunks)),
	}, nil
}

// SearchSimilar performs hybrid semantic + full-text search for similar document chunks.
// It combines vector and FTS results, expands context with adjacent chunks, and applies keyword reranking.
func (uc *EmbeddingUseCase) SearchSimilar(ctx context.Context, query string, limit int, threshold float64) ([]entities.ChunkWithScore, error) {
	if limit <= 0 {
		limit = 10
	}
	if threshold <= 0 {
		threshold = 0.7
	}

	// Generate embedding for query (cached to avoid redundant API calls)
	embedding, err := uc.cachedQueryEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Use hybrid search (vector + FTS with RRF)
	results, err := uc.embeddingRepo.SearchHybrid(ctx, embedding, query, limit, threshold)
	if err != nil {
		// Fallback to vector-only search if hybrid fails (e.g. search_vector column not yet migrated)
		results, err = uc.embeddingRepo.SearchSimilar(ctx, embedding, limit, threshold)
		if err != nil {
			return nil, fmt.Errorf("failed to search similar chunks: %w", err)
		}
	}

	if len(results) == 0 {
		return results, nil
	}

	// Expand context with adjacent chunks (±1 neighbor)
	results = uc.expandWithAdjacentChunks(ctx, results)

	// Apply keyword reranking
	results = rerankByKeywords(results, query)

	// Limit to requested count
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// expandWithAdjacentChunks fetches neighboring chunks and merges their text into results.
func (uc *EmbeddingUseCase) expandWithAdjacentChunks(ctx context.Context, results []entities.ChunkWithScore) []entities.ChunkWithScore {
	chunkIDs := make([]int64, len(results))
	for i, r := range results {
		chunkIDs[i] = r.Chunk.ID
	}

	adjacentChunks, err := uc.embeddingRepo.GetAdjacentChunks(ctx, chunkIDs, 1)
	if err != nil || len(adjacentChunks) == 0 {
		return results
	}

	// Index adjacent chunks by (document_id, chunk_index)
	type chunkKey struct {
		docID      int64
		chunkIndex int
	}
	adjacentMap := make(map[chunkKey]*entities.DocumentChunk, len(adjacentChunks))
	for i := range adjacentChunks {
		c := &adjacentChunks[i]
		adjacentMap[chunkKey{c.DocumentID, c.ChunkIndex}] = c
	}

	// Merge adjacent text into each result
	for i := range results {
		chunk := results[i].Chunk
		var parts []string

		// Previous chunk
		if prev, ok := adjacentMap[chunkKey{chunk.DocumentID, chunk.ChunkIndex - 1}]; ok {
			parts = append(parts, prev.ChunkText)
		}
		parts = append(parts, chunk.ChunkText)
		// Next chunk
		if next, ok := adjacentMap[chunkKey{chunk.DocumentID, chunk.ChunkIndex + 1}]; ok {
			parts = append(parts, next.ChunkText)
		}

		if len(parts) > 1 {
			results[i].Chunk.ChunkText = strings.Join(parts, "\n\n")
		}
	}

	return results
}

// rerankByKeywords applies a simple keyword overlap boost to search results.
func rerankByKeywords(results []entities.ChunkWithScore, query string) []entities.ChunkWithScore {
	if len(results) == 0 || query == "" {
		return results
	}

	// Extract unique query words (lowercased)
	queryWords := strings.Fields(strings.ToLower(query))
	uniqueWords := make(map[string]struct{}, len(queryWords))
	for _, w := range queryWords {
		if len([]rune(w)) > 2 { // skip very short words
			uniqueWords[w] = struct{}{}
		}
	}

	if len(uniqueWords) == 0 {
		return results
	}

	totalQueryWords := float64(len(uniqueWords))

	for i := range results {
		chunkLower := strings.ToLower(results[i].Chunk.ChunkText)
		matched := 0
		for word := range uniqueWords {
			if strings.Contains(chunkLower, word) {
				matched++
			}
		}
		keywordScore := float64(matched) / totalQueryWords
		// Blend: 70% original score + 30% keyword score
		results[i].SimilarityScore = 0.7*results[i].SimilarityScore + 0.3*keywordScore
	}

	// Re-sort by updated score
	sort.Slice(results, func(i, j int) bool {
		return results[i].SimilarityScore > results[j].SimilarityScore
	})

	return results
}

// Search performs semantic search with optional document type filter
func (uc *EmbeddingUseCase) Search(ctx context.Context, req *dto.SearchRequest) (*dto.SearchResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	threshold := req.Threshold
	if threshold <= 0 {
		threshold = 0.7
	}

	// Generate embedding for query (cached to avoid redundant API calls)
	embedding, err := uc.cachedQueryEmbedding(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search for similar chunks
	var results []entities.ChunkWithScore
	if len(req.DocumentTypes) > 0 {
		results, err = uc.embeddingRepo.SearchSimilarByDocumentTypes(ctx, embedding, req.DocumentTypes, limit, threshold)
	} else {
		results, err = uc.embeddingRepo.SearchSimilar(ctx, embedding, limit, threshold)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	response := &dto.SearchResponse{
		Results: make([]dto.SearchResultResponse, 0, len(results)),
		Query:   req.Query,
		Total:   len(results),
	}

	for _, r := range results {
		response.Results = append(response.Results, *dto.ToSearchResultResponse(&r))
	}

	return response, nil
}

// GetIndexingStatus retrieves the indexing status
func (uc *EmbeddingUseCase) GetIndexingStatus(ctx context.Context) (*dto.IndexStatusResponse, error) {
	total, indexed, pending, lastIndexed, err := uc.embeddingRepo.GetIndexingStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get indexing stats: %w", err)
	}

	return &dto.IndexStatusResponse{
		TotalDocuments:   total,
		IndexedDocuments: indexed,
		PendingDocuments: pending,
		LastIndexedAt:    lastIndexed,
	}, nil
}

// IndexPendingDocuments indexes all pending documents (for background job)
func (uc *EmbeddingUseCase) IndexPendingDocuments(ctx context.Context, batchSize int) (int, error) {
	if batchSize <= 0 {
		batchSize = 10
	}

	documentIDs, err := uc.embeddingRepo.GetPendingDocuments(ctx, batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get pending documents: %w", err)
	}

	indexed := 0
	for _, docID := range documentIDs {
		_, err := uc.IndexDocument(ctx, docID, false)
		if err != nil {
			// Log error but continue with other documents
			if uc.auditLogger != nil {
				uc.auditLogger.LogAuditEvent(ctx, "error", "ai_index", map[string]any{
					"document_id": docID,
					"error":       err.Error(),
				})
			}
			continue
		}
		indexed++
	}

	return indexed, nil
}
