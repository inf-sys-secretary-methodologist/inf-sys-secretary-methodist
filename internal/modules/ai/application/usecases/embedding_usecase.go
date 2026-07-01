// Package usecases contains application use cases for the AI module.
package usecases

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/ports"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/cache"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// EmbeddingProvider is a type alias for ports.EmbeddingProvider.
//
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
	embeddingRepo     EmbeddingRepository
	embeddingProvider EmbeddingProvider
	documentProvider  DocumentProvider
	chunkingService   *services.ChunkingService
	auditLogger       *logging.AuditLogger
	cache             *cache.RedisCache
	embeddingModel    string
	reranker          ports.LLMProvider // optional: LLM reranking stage for modified RAG
}

// WithReranker attaches an LLM provider used as the final reranking stage of the
// modified RAG pipeline (search_mode="hybrid_rerank"). Optional; nil = disabled.
func (uc *EmbeddingUseCase) WithReranker(p ports.LLMProvider) *EmbeddingUseCase {
	uc.reranker = p
	return uc
}

// NewEmbeddingUseCase creates a new EmbeddingUseCase
func NewEmbeddingUseCase(
	embeddingRepo EmbeddingRepository,
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

// llmRerank reorders results using the LLM as a cross-encoder-style reranker
// (final stage of the modified RAG pipeline). Best-effort: on any failure it
// returns the original ordering unchanged.
func (uc *EmbeddingUseCase) llmRerank(ctx context.Context, query string, results []entities.ChunkWithScore) []entities.ChunkWithScore {
	if uc.reranker == nil || len(results) < 2 {
		return results
	}

	var sb strings.Builder
	for i, r := range results {
		snippet := r.Chunk.ChunkText
		if len([]rune(snippet)) > 300 {
			snippet = string([]rune(snippet)[:300])
		}
		fmt.Fprintf(&sb, "[%d] %s: %s\n", i+1, r.DocumentTitle, snippet)
	}

	systemPrompt := "Ты ранжируешь фрагменты документов по релевантности запросу пользователя. " +
		"Верни ТОЛЬКО номера кандидатов через запятую, от самого релевантного к наименее релевантному. " +
		"Без пояснений и текста, только номера."
	userMsg := fmt.Sprintf("Запрос: %s\n\nКандидаты:\n%s", query, sb.String())

	resp, _, err := uc.reranker.GenerateResponse(ctx, systemPrompt,
		[]entities.Message{{Role: entities.MessageRoleUser, Content: userMsg}}, "")
	if err != nil || resp == "" {
		return results
	}

	order := parseRerankOrder(resp, len(results))
	if len(order) == 0 {
		return results
	}

	reranked := make([]entities.ChunkWithScore, 0, len(results))
	used := make([]bool, len(results))
	for _, idx := range order {
		if idx >= 0 && idx < len(results) && !used[idx] {
			reranked = append(reranked, results[idx])
			used[idx] = true
		}
	}
	for i := range results {
		if !used[i] {
			reranked = append(reranked, results[i])
		}
	}
	return reranked
}

// parseRerankOrder extracts 1-based indices from the LLM response, converting to 0-based.
func parseRerankOrder(resp string, n int) []int {
	var out []int
	for _, tok := range strings.FieldsFunc(resp, func(r rune) bool { return r < '0' || r > '9' }) {
		if v, err := strconv.Atoi(tok); err == nil && v >= 1 && v <= n {
			out = append(out, v-1)
		}
	}
	return out
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

	var results []entities.ChunkWithScore
	var err error

	switch req.SearchMode {
	case "hybrid_rerank":
		// Modified RAG + final LLM reranking stage.
		results, err = uc.SearchSimilar(ctx, req.Query, limit, threshold)
		if err != nil {
			return nil, fmt.Errorf("failed to search: %w", err)
		}
		results = uc.llmRerank(ctx, req.Query, results)
	case "hybrid":
		// Modified RAG pipeline: hybrid (vector + FTS) with RRF, adjacent chunk
		// expansion (±1) and keyword reranking. Reuses SearchSimilar.
		results, err = uc.SearchSimilar(ctx, req.Query, limit, threshold)
		if err != nil {
			return nil, fmt.Errorf("failed to search: %w", err)
		}
	case "fts_only":
		// Baseline: single-method full-text (keyword-only) search.
		results, err = uc.embeddingRepo.SearchFullText(ctx, req.Query, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to search: %w", err)
		}
	default:
		// Baseline: single vector-only search (no RRF / expansion / reranking).
		embedding, embErr := uc.cachedQueryEmbedding(ctx, req.Query)
		if embErr != nil {
			return nil, fmt.Errorf("failed to generate query embedding: %w", embErr)
		}
		if len(req.DocumentTypes) > 0 {
			results, err = uc.embeddingRepo.SearchSimilarByDocumentTypes(ctx, embedding, req.DocumentTypes, limit, threshold)
		} else {
			results, err = uc.embeddingRepo.SearchSimilar(ctx, embedding, limit, threshold)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to search: %w", err)
		}
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
