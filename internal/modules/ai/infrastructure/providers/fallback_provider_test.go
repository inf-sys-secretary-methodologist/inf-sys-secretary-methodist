package providers

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

// --- Mock LLM Provider ---

type mockLLMProvider struct {
	response string
	tokens   int
	err      error
	called   bool
}

func (m *mockLLMProvider) GenerateResponse(_ context.Context, _ string, _ []entities.Message, _ string) (string, int, error) {
	m.called = true
	return m.response, m.tokens, m.err
}

func (m *mockLLMProvider) GenerateResponseStream(_ context.Context, _ string, _ []entities.Message, _ string, onChunk func(string) error) (string, int, error) {
	m.called = true
	if m.err != nil {
		return "", 0, m.err
	}
	if err := onChunk(m.response); err != nil {
		return m.response, m.tokens, err
	}
	return m.response, m.tokens, nil
}

// --- Mock Embedding Provider ---

type mockEmbeddingProvider struct {
	embedding  []float32
	embeddings [][]float32
	err        error
	called     bool
}

func (m *mockEmbeddingProvider) GenerateEmbedding(_ context.Context, _ string) ([]float32, error) {
	m.called = true
	return m.embedding, m.err
}

func (m *mockEmbeddingProvider) GenerateEmbeddings(_ context.Context, _ []string) ([][]float32, error) {
	m.called = true
	return m.embeddings, m.err
}

func (m *mockEmbeddingProvider) GenerateQueryEmbedding(_ context.Context, _ string) ([]float32, error) {
	m.called = true
	return m.embedding, m.err
}

// --- chunkThenFailProvider: sends a chunk then returns an error ---

type chunkThenFailProvider struct {
	content string
	err     error
	called  bool
}

func (p *chunkThenFailProvider) GenerateResponse(_ context.Context, _ string, _ []entities.Message, _ string) (string, int, error) {
	p.called = true
	return "", 0, p.err
}

func (p *chunkThenFailProvider) GenerateResponseStream(_ context.Context, _ string, _ []entities.Message, _ string, onChunk func(string) error) (string, int, error) {
	p.called = true
	_ = onChunk(p.content) // send chunk first
	return p.content, 0, p.err
}

// --- FallbackLLMProvider Tests ---

func TestFallbackLLM_PrimarySucceeds(t *testing.T) {
	primary := &mockLLMProvider{response: "primary answer", tokens: 10}
	fallback := &mockLLMProvider{response: "fallback answer", tokens: 5}
	provider := NewFallbackLLMProvider(primary, fallback, slog.Default())

	content, tokens, err := provider.GenerateResponse(context.Background(), "sys", nil, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "primary answer" {
		t.Errorf("expected primary answer, got %q", content)
	}
	if tokens != 10 {
		t.Errorf("expected 10 tokens, got %d", tokens)
	}
	if !primary.called {
		t.Error("primary should have been called")
	}
	if fallback.called {
		t.Error("fallback should not have been called")
	}
}

func TestFallbackLLM_PrimaryFails_FallbackSucceeds(t *testing.T) {
	primary := &mockLLMProvider{err: errors.New("primary down")}
	fallback := &mockLLMProvider{response: "fallback answer", tokens: 5}
	provider := NewFallbackLLMProvider(primary, fallback, slog.Default())

	content, tokens, err := provider.GenerateResponse(context.Background(), "sys", nil, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "fallback answer" {
		t.Errorf("expected fallback answer, got %q", content)
	}
	if tokens != 5 {
		t.Errorf("expected 5 tokens, got %d", tokens)
	}
	if !primary.called || !fallback.called {
		t.Error("both providers should have been called")
	}
}

func TestFallbackLLM_BothFail(t *testing.T) {
	primary := &mockLLMProvider{err: errors.New("primary down")}
	fallback := &mockLLMProvider{err: errors.New("fallback down")}
	provider := NewFallbackLLMProvider(primary, fallback, slog.Default())

	_, _, err := provider.GenerateResponse(context.Background(), "sys", nil, "")

	if err == nil {
		t.Fatal("expected error when both providers fail")
	}
	if err.Error() != "fallback down" {
		t.Errorf("expected fallback error, got %q", err.Error())
	}
}

func TestFallbackLLM_RateLimitTriggersFallback(t *testing.T) {
	primary := &mockLLMProvider{err: &ErrRateLimited{RetryAfter: 30 * time.Second}}
	fallback := &mockLLMProvider{response: "fallback answer", tokens: 5}
	provider := NewFallbackLLMProvider(primary, fallback, slog.Default())

	content, _, err := provider.GenerateResponse(context.Background(), "sys", nil, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "fallback answer" {
		t.Errorf("expected fallback answer, got %q", content)
	}
}

// --- FallbackEmbeddingProvider Tests ---

func TestFallbackEmbedding_PrimarySucceeds(t *testing.T) {
	primary := &mockEmbeddingProvider{embedding: []float32{1.0, 2.0}}
	fallback := &mockEmbeddingProvider{embedding: []float32{3.0, 4.0}}
	provider := NewFallbackEmbeddingProvider(primary, fallback, slog.Default())

	result, err := provider.GenerateEmbedding(context.Background(), "test")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 || result[0] != 1.0 {
		t.Errorf("expected primary embedding, got %v", result)
	}
	if fallback.called {
		t.Error("fallback should not have been called")
	}
}

func TestFallbackEmbedding_PrimaryFails_FallbackSucceeds(t *testing.T) {
	primary := &mockEmbeddingProvider{err: errors.New("primary down")}
	fallback := &mockEmbeddingProvider{embedding: []float32{3.0, 4.0}}
	provider := NewFallbackEmbeddingProvider(primary, fallback, slog.Default())

	result, err := provider.GenerateEmbedding(context.Background(), "test")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 || result[0] != 3.0 {
		t.Errorf("expected fallback embedding, got %v", result)
	}
}

func TestFallbackEmbedding_BothFail(t *testing.T) {
	primary := &mockEmbeddingProvider{err: errors.New("primary down")}
	fallback := &mockEmbeddingProvider{err: errors.New("fallback down")}
	provider := NewFallbackEmbeddingProvider(primary, fallback, slog.Default())

	_, err := provider.GenerateEmbedding(context.Background(), "test")

	if err == nil {
		t.Fatal("expected error when both providers fail")
	}
	if err.Error() != "fallback down" {
		t.Errorf("expected fallback error, got %q", err.Error())
	}
}

func TestFallbackEmbeddings_PrimaryFails_FallbackSucceeds(t *testing.T) {
	primary := &mockEmbeddingProvider{err: errors.New("primary down")}
	fallback := &mockEmbeddingProvider{embeddings: [][]float32{{1.0}, {2.0}}}
	provider := NewFallbackEmbeddingProvider(primary, fallback, slog.Default())

	result, err := provider.GenerateEmbeddings(context.Background(), []string{"a", "b"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 embeddings, got %d", len(result))
	}
}

func TestFallbackQueryEmbedding_PrimaryFails_FallbackSucceeds(t *testing.T) {
	primary := &mockEmbeddingProvider{err: errors.New("primary down")}
	fallback := &mockEmbeddingProvider{embedding: []float32{5.0, 6.0}}
	provider := NewFallbackEmbeddingProvider(primary, fallback, slog.Default())

	result, err := provider.GenerateQueryEmbedding(context.Background(), "query")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 || result[0] != 5.0 {
		t.Errorf("expected fallback query embedding, got %v", result)
	}
}

// --- FallbackLLMProvider Streaming Tests ---

func TestFallbackLLMStream_PrimarySucceeds(t *testing.T) {
	primary := &mockLLMProvider{response: "streamed", tokens: 10}
	fallback := &mockLLMProvider{response: "fallback", tokens: 5}
	provider := NewFallbackLLMProvider(primary, fallback, slog.Default())

	var chunks []string
	content, tokens, err := provider.GenerateResponseStream(context.Background(), "sys", nil, "", func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "streamed" {
		t.Errorf("expected 'streamed', got %q", content)
	}
	if tokens != 10 {
		t.Errorf("expected 10 tokens, got %d", tokens)
	}
	if fallback.called {
		t.Error("fallback should not have been called")
	}
}

func TestFallbackLLMStream_PrimaryFails_NoChunksSent(t *testing.T) {
	primary := &mockLLMProvider{err: errors.New("primary down")}
	fallback := &mockLLMProvider{response: "fallback", tokens: 5}
	provider := NewFallbackLLMProvider(primary, fallback, slog.Default())

	var chunks []string
	content, tokens, err := provider.GenerateResponseStream(context.Background(), "sys", nil, "", func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "fallback" {
		t.Errorf("expected 'fallback', got %q", content)
	}
	if tokens != 5 {
		t.Errorf("expected 5 tokens, got %d", tokens)
	}
}

func TestFallbackLLMStream_PrimaryFails_ChunksAlreadySent(t *testing.T) {
	// Primary sends a chunk via onChunk and then returns an error
	primaryErr := errors.New("mid-stream failure")
	primary := &mockLLMProvider{}
	// Override the mock to simulate partial data sent
	originalStream := primary.GenerateResponseStream
	_ = originalStream

	// Create a custom primary that sends a chunk then fails
	customPrimary := &chunkThenFailProvider{content: "partial", err: primaryErr}
	fallback := &mockLLMProvider{response: "fallback", tokens: 5}
	provider := NewFallbackLLMProvider(customPrimary, fallback, slog.Default())

	var chunks []string
	_, _, err := provider.GenerateResponseStream(context.Background(), "sys", nil, "", func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	})
	// Should return primary error since chunks were already sent (no fallback)
	if err == nil {
		t.Fatal("expected error when chunks already sent")
	}
	if err != primaryErr {
		t.Errorf("expected primary error, got: %v", err)
	}
	if fallback.called {
		t.Error("fallback should not be called when chunks already sent")
	}
}

func TestFallbackLLMStream_BothFail(t *testing.T) {
	primary := &mockLLMProvider{err: errors.New("primary down")}
	fallback := &mockLLMProvider{err: errors.New("fallback down")}
	provider := NewFallbackLLMProvider(primary, fallback, slog.Default())

	_, _, err := provider.GenerateResponseStream(context.Background(), "sys", nil, "", func(chunk string) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error when both fail")
	}
	if err.Error() != "fallback down" {
		t.Errorf("expected fallback error, got: %v", err)
	}
}

// --- FallbackEmbeddingProvider additional tests ---

func TestFallbackEmbeddings_PrimarySucceeds(t *testing.T) {
	primary := &mockEmbeddingProvider{embeddings: [][]float32{{1.0}, {2.0}}}
	fallback := &mockEmbeddingProvider{embeddings: [][]float32{{3.0}, {4.0}}}
	provider := NewFallbackEmbeddingProvider(primary, fallback, slog.Default())

	result, err := provider.GenerateEmbeddings(context.Background(), []string{"a", "b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 || result[0][0] != 1.0 {
		t.Errorf("expected primary embeddings, got %v", result)
	}
	if fallback.called {
		t.Error("fallback should not have been called")
	}
}

func TestFallbackQueryEmbedding_PrimarySucceeds(t *testing.T) {
	primary := &mockEmbeddingProvider{embedding: []float32{1.0, 2.0}}
	fallback := &mockEmbeddingProvider{embedding: []float32{3.0, 4.0}}
	provider := NewFallbackEmbeddingProvider(primary, fallback, slog.Default())

	result, err := provider.GenerateQueryEmbedding(context.Background(), "query")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 || result[0] != 1.0 {
		t.Errorf("expected primary query embedding, got %v", result)
	}
	if fallback.called {
		t.Error("fallback should not have been called")
	}
}

func TestFallbackEmbedding_RateLimitTriggersFallback(t *testing.T) {
	primary := &mockEmbeddingProvider{err: &ErrRateLimited{RetryAfter: 10 * time.Second}}
	fallback := &mockEmbeddingProvider{embedding: []float32{1.0}}
	provider := NewFallbackEmbeddingProvider(primary, fallback, slog.Default())

	result, err := provider.GenerateEmbedding(context.Background(), "test")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != 1.0 {
		t.Errorf("expected fallback embedding, got %v", result)
	}
}
