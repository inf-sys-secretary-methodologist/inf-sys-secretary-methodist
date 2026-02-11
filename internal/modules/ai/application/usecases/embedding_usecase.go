// Package usecases contains application use cases for the AI module.
package usecases

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// EmbeddingProvider defines the interface for embedding generation
type EmbeddingProvider interface {
	// GenerateEmbedding generates an embedding vector for text
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)

	// GenerateEmbeddings generates embedding vectors for multiple texts
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
}

// DocumentProvider defines the interface for accessing document content
type DocumentProvider interface {
	// GetDocumentContent retrieves the text content of a document
	GetDocumentContent(ctx context.Context, documentID int64) (string, string, error) // returns content, title, error
}

// EmbeddingUseCase handles document embedding and search operations
type EmbeddingUseCase struct {
	embeddingRepo     repositories.EmbeddingRepository
	embeddingProvider EmbeddingProvider
	documentProvider  DocumentProvider
	chunkingService   *services.ChunkingService
	auditLogger       *logging.AuditLogger
}

// NewEmbeddingUseCase creates a new EmbeddingUseCase
func NewEmbeddingUseCase(
	embeddingRepo repositories.EmbeddingRepository,
	embeddingProvider EmbeddingProvider,
	documentProvider DocumentProvider,
	auditLogger *logging.AuditLogger,
) *EmbeddingUseCase {
	return &EmbeddingUseCase{
		embeddingRepo:     embeddingRepo,
		embeddingProvider: embeddingProvider,
		documentProvider:  documentProvider,
		chunkingService:   services.NewChunkingService(services.DefaultChunkingConfig()),
		auditLogger:       auditLogger,
	}
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
		uc.embeddingRepo.SetIndexStatus(ctx, indexStatus)
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
		uc.embeddingRepo.SetIndexStatus(ctx, indexStatus)
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
		uc.embeddingRepo.SetIndexStatus(ctx, indexStatus)
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
		uc.embeddingRepo.SetIndexStatus(ctx, indexStatus)
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Create embeddings in database
	embeddings := make([]entities.Embedding, len(chunks))
	for i, chunk := range chunks {
		embeddings[i] = *entities.NewEmbedding(chunk.ID, vectors[i], "text-embedding-3-small")
	}

	if err := uc.embeddingRepo.CreateEmbeddings(ctx, embeddings); err != nil {
		errMsg := err.Error()
		indexStatus.Status = entities.IndexStatusFailed
		indexStatus.ErrorMessage = &errMsg
		uc.embeddingRepo.SetIndexStatus(ctx, indexStatus)
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

// SearchSimilar performs semantic search for similar document chunks
func (uc *EmbeddingUseCase) SearchSimilar(ctx context.Context, query string, limit int, threshold float64) ([]entities.ChunkWithScore, error) {
	if limit <= 0 {
		limit = 5
	}
	if threshold <= 0 {
		threshold = 0.7
	}

	// Generate embedding for query
	embedding, err := uc.embeddingProvider.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search for similar chunks
	results, err := uc.embeddingRepo.SearchSimilar(ctx, embedding, limit, threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar chunks: %w", err)
	}

	return results, nil
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

	// Generate embedding for query
	embedding, err := uc.embeddingProvider.GenerateEmbedding(ctx, req.Query)
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
