// Package entities contains domain entities for the AI module.
package entities

import "time"

// DocumentChunk represents a chunk of document text for RAG
type DocumentChunk struct {
	ID          int64                  `json:"id"`
	DocumentID  int64                  `json:"document_id"`
	ChunkIndex  int                    `json:"chunk_index"`
	ChunkText   string                 `json:"chunk_text"`
	ChunkTokens *int                   `json:"chunk_tokens,omitempty"`
	PageNumber  *int                   `json:"page_number,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// NewDocumentChunk creates a new document chunk
func NewDocumentChunk(documentID int64, index int, text string, tokens int) *DocumentChunk {
	return &DocumentChunk{
		DocumentID:  documentID,
		ChunkIndex:  index,
		ChunkText:   text,
		ChunkTokens: &tokens,
		Metadata:    make(map[string]interface{}),
		CreatedAt:   time.Now(),
	}
}

// ChunkWithScore represents a chunk with its similarity score for search results
type ChunkWithScore struct {
	Chunk           *DocumentChunk `json:"chunk"`
	DocumentTitle   string         `json:"document_title"`
	SimilarityScore float64        `json:"similarity_score"`
}
