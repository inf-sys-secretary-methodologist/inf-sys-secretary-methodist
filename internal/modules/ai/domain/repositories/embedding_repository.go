// Package repositories contains repository interfaces for the AI module.
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

// EmbeddingRepository defines the interface for embedding persistence
type EmbeddingRepository interface {
	// CreateChunk creates a document chunk
	CreateChunk(ctx context.Context, chunk *entities.DocumentChunk) error

	// CreateChunks creates multiple document chunks
	CreateChunks(ctx context.Context, chunks []entities.DocumentChunk) error

	// GetChunksByDocumentID retrieves chunks for a document
	GetChunksByDocumentID(ctx context.Context, documentID int64) ([]entities.DocumentChunk, error)

	// DeleteChunksByDocumentID deletes all chunks for a document
	DeleteChunksByDocumentID(ctx context.Context, documentID int64) error

	// CreateEmbedding creates an embedding for a chunk
	CreateEmbedding(ctx context.Context, embedding *entities.Embedding) error

	// CreateEmbeddings creates multiple embeddings
	CreateEmbeddings(ctx context.Context, embeddings []entities.Embedding) error

	// SearchSimilar finds similar chunks using vector similarity
	SearchSimilar(ctx context.Context, embedding []float32, limit int, threshold float64) ([]entities.ChunkWithScore, error)

	// SearchSimilarByDocumentTypes finds similar chunks filtered by document types
	SearchSimilarByDocumentTypes(ctx context.Context, embedding []float32, docTypes []string, limit int, threshold float64) ([]entities.ChunkWithScore, error)

	// GetIndexStatus retrieves the indexing status for a document
	GetIndexStatus(ctx context.Context, documentID int64) (*entities.DocumentIndexStatus, error)

	// SetIndexStatus sets the indexing status for a document
	SetIndexStatus(ctx context.Context, status *entities.DocumentIndexStatus) error

	// GetPendingDocuments retrieves documents that need indexing
	GetPendingDocuments(ctx context.Context, limit int) ([]int64, error)

	// GetIndexingStats retrieves indexing statistics
	GetIndexingStats(ctx context.Context) (totalDocuments, indexedDocuments, pendingDocuments int, lastIndexedAt *string, err error)
}
