// Package ports defines interfaces (ports) for the AI module's external dependencies.
package ports

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

// LLMProvider defines the interface for LLM chat interactions.
type LLMProvider interface {
	// GenerateResponse generates a response from the LLM.
	GenerateResponse(ctx context.Context, systemPrompt string, messages []entities.Message, context string) (string, int, error)

	// GenerateResponseStream generates a streaming response from the LLM.
	// onChunk is called for each text fragment as it arrives from the provider.
	// Returns the full accumulated response text and total tokens used.
	GenerateResponseStream(ctx context.Context, systemPrompt string, messages []entities.Message, context string, onChunk func(chunk string) error) (string, int, error)
}

// EmbeddingProvider defines the interface for embedding generation.
type EmbeddingProvider interface {
	// GenerateEmbedding generates an embedding vector for a single text (for indexing).
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)

	// GenerateEmbeddings generates embedding vectors for multiple texts (for batch indexing).
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)

	// GenerateQueryEmbedding generates an embedding vector optimized for search queries.
	// Providers that support task types (e.g. Gemini) use RETRIEVAL_QUERY instead of RETRIEVAL_DOCUMENT.
	GenerateQueryEmbedding(ctx context.Context, text string) ([]float32, error)
}
