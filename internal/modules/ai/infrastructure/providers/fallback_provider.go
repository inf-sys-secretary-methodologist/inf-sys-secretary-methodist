// Package providers contains external service providers for the AI module.
package providers

import (
	"context"
	"errors"
	"log/slog"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/ports"
)

// logFallbackError logs a warning with differentiated messages for rate-limit vs generic errors.
func logFallbackError(logger *slog.Logger, component string, err error) {
	var rateLimitErr *ErrRateLimited
	if errors.As(err, &rateLimitErr) {
		logger.Warn("primary "+component+" provider rate-limited, switching to fallback", "error", err)
	} else {
		logger.Warn("primary "+component+" provider failed, switching to fallback", "error", err)
	}
}

// --- FallbackLLMProvider ---

// FallbackLLMProvider wraps a primary LLM provider with a fallback.
// If the primary provider fails, the request is retried with the fallback provider.
type FallbackLLMProvider struct {
	primary  ports.LLMProvider
	fallback ports.LLMProvider
	logger   *slog.Logger
}

// NewFallbackLLMProvider creates a new FallbackLLMProvider.
func NewFallbackLLMProvider(primary, fallback ports.LLMProvider, logger *slog.Logger) *FallbackLLMProvider {
	return &FallbackLLMProvider{
		primary:  primary,
		fallback: fallback,
		logger:   logger,
	}
}

// GenerateResponse tries the primary provider first. On any error it falls back
// to the secondary provider.
func (f *FallbackLLMProvider) GenerateResponse(ctx context.Context, systemPrompt string, messages []entities.Message, contextText string) (string, int, error) {
	content, tokens, err := f.primary.GenerateResponse(ctx, systemPrompt, messages, contextText)
	if err == nil {
		return content, tokens, nil
	}

	logFallbackError(f.logger, "LLM", err)
	return f.fallback.GenerateResponse(ctx, systemPrompt, messages, contextText)
}

// --- FallbackEmbeddingProvider ---

// FallbackEmbeddingProvider wraps a primary embedding provider with a fallback.
// If the primary provider fails, the request is retried with the fallback provider.
type FallbackEmbeddingProvider struct {
	primary  ports.EmbeddingProvider
	fallback ports.EmbeddingProvider
	logger   *slog.Logger
}

// NewFallbackEmbeddingProvider creates a new FallbackEmbeddingProvider.
func NewFallbackEmbeddingProvider(primary, fallback ports.EmbeddingProvider, logger *slog.Logger) *FallbackEmbeddingProvider {
	return &FallbackEmbeddingProvider{
		primary:  primary,
		fallback: fallback,
		logger:   logger,
	}
}

// GenerateEmbedding tries the primary provider first, falls back on error.
func (f *FallbackEmbeddingProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	result, err := f.primary.GenerateEmbedding(ctx, text)
	if err == nil {
		return result, nil
	}

	logFallbackError(f.logger, "embedding", err)
	return f.fallback.GenerateEmbedding(ctx, text)
}

// GenerateEmbeddings tries the primary provider first, falls back on error.
func (f *FallbackEmbeddingProvider) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	result, err := f.primary.GenerateEmbeddings(ctx, texts)
	if err == nil {
		return result, nil
	}

	logFallbackError(f.logger, "embedding", err)
	return f.fallback.GenerateEmbeddings(ctx, texts)
}

// GenerateQueryEmbedding tries the primary provider first, falls back on error.
func (f *FallbackEmbeddingProvider) GenerateQueryEmbedding(ctx context.Context, text string) ([]float32, error) {
	result, err := f.primary.GenerateQueryEmbedding(ctx, text)
	if err == nil {
		return result, nil
	}

	logFallbackError(f.logger, "embedding", err)
	return f.fallback.GenerateQueryEmbedding(ctx, text)
}
