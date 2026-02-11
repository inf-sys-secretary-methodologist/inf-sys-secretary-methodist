// Package entities contains domain entities for the AI module.
package entities

import "time"

// Embedding represents a vector embedding for a document chunk
type Embedding struct {
	ID        int64     `json:"id"`
	ChunkID   int64     `json:"chunk_id"`
	Embedding []float32 `json:"embedding"` // Vector of floats
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
}

// EmbeddingDimension is the dimension of OpenAI text-embedding-3-small
const EmbeddingDimension = 1536

// NewEmbedding creates a new embedding
func NewEmbedding(chunkID int64, embedding []float32, model string) *Embedding {
	return &Embedding{
		ChunkID:   chunkID,
		Embedding: embedding,
		Model:     model,
		CreatedAt: time.Now(),
	}
}

// IndexStatus represents the indexing status of a document
type IndexStatus string

const (
	IndexStatusPending  IndexStatus = "pending"
	IndexStatusIndexing IndexStatus = "indexing"
	IndexStatusIndexed  IndexStatus = "indexed"
	IndexStatusFailed   IndexStatus = "failed"
)

// DocumentIndexStatus tracks the indexing status of a document
type DocumentIndexStatus struct {
	DocumentID   int64       `json:"document_id"`
	Status       IndexStatus `json:"status"`
	ChunksCount  int         `json:"chunks_count"`
	ErrorMessage *string     `json:"error_message,omitempty"`
	IndexedAt    *time.Time  `json:"indexed_at,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}
