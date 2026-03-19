package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

func TestDefaultOpenAIConfig(t *testing.T) {
	cfg := DefaultOpenAIConfig()
	if cfg.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("unexpected BaseURL: %q", cfg.BaseURL)
	}
	if cfg.EmbeddingModel != "text-embedding-3-small" {
		t.Errorf("unexpected EmbeddingModel: %q", cfg.EmbeddingModel)
	}
	if cfg.ChatModel != "gemini-2.5-flash" {
		t.Errorf("unexpected ChatModel: %q", cfg.ChatModel)
	}
	if cfg.MaxTokens != 2048 {
		t.Errorf("unexpected MaxTokens: %d", cfg.MaxTokens)
	}
	if cfg.Temperature != 0.3 {
		t.Errorf("unexpected Temperature: %f", cfg.Temperature)
	}
	if cfg.Timeout != 60*time.Second {
		t.Errorf("unexpected Timeout: %v", cfg.Timeout)
	}
}

func TestNewOpenAIProvider_Defaults(t *testing.T) {
	p := NewOpenAIProvider(OpenAIConfig{APIKey: "test-key"})
	if p.config.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("expected default BaseURL, got %q", p.config.BaseURL)
	}
	if p.config.EmbeddingModel != "text-embedding-3-small" {
		t.Errorf("expected default EmbeddingModel, got %q", p.config.EmbeddingModel)
	}
	if p.config.ChatModel != "gemini-2.5-flash" {
		t.Errorf("expected default ChatModel, got %q", p.config.ChatModel)
	}
	if p.config.Timeout != 60*time.Second {
		t.Errorf("expected default Timeout, got %v", p.config.Timeout)
	}
}

func TestNewOpenAIProvider_CustomConfig(t *testing.T) {
	cfg := OpenAIConfig{
		APIKey:         "key",
		BaseURL:        "https://custom.api.com/v1",
		EmbeddingModel: "custom-embed",
		ChatModel:      "custom-chat",
		Timeout:        30 * time.Second,
	}
	p := NewOpenAIProvider(cfg)
	if p.config.BaseURL != "https://custom.api.com/v1" {
		t.Errorf("expected custom BaseURL, got %q", p.config.BaseURL)
	}
	if p.config.EmbeddingModel != "custom-embed" {
		t.Errorf("expected custom EmbeddingModel, got %q", p.config.EmbeddingModel)
	}
}

// --- Embedding Tests ---

func TestOpenAI_GenerateEmbeddings_EmptyInput(t *testing.T) {
	p := NewOpenAIProvider(OpenAIConfig{APIKey: "test"})
	result, err := p.GenerateEmbeddings(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty input, got %v", result)
	}
}

func TestOpenAI_GenerateEmbeddings_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		resp := embeddingResponse{
			Data: []struct {
				Embedding []float32 `json:"embedding"`
				Index     int       `json:"index"`
			}{
				{Embedding: []float32{0.1, 0.2, 0.3}, Index: 0},
				{Embedding: []float32{0.4, 0.5, 0.6}, Index: 1},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "test-key", BaseURL: server.URL})
	result, err := p.GenerateEmbeddings(context.Background(), []string{"text1", "text2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 embeddings, got %d", len(result))
	}
	if result[0][0] != 0.1 {
		t.Errorf("expected 0.1, got %f", result[0][0])
	}
}

func TestOpenAI_GenerateEmbeddings_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := embeddingResponse{}
		resp.Error = &struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		}{Message: "invalid api key", Type: "auth", Code: "invalid_api_key"}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "bad-key", BaseURL: server.URL})
	_, err := p.GenerateEmbeddings(context.Background(), []string{"text"})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "invalid api key") {
		t.Errorf("expected API error message, got: %v", err)
	}
}

func TestOpenAI_GenerateEmbeddings_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(embeddingResponse{})
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	_, err := p.GenerateEmbeddings(context.Background(), []string{"text"})
	if err == nil {
		t.Fatal("expected error for 500 status")
	}
	if !strings.Contains(err.Error(), "status 500") {
		t.Errorf("expected status error, got: %v", err)
	}
}

func TestOpenAI_GenerateEmbeddings_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not json"))
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	_, err := p.GenerateEmbeddings(context.Background(), []string{"text"})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestOpenAI_GenerateEmbedding_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := embeddingResponse{
			Data: []struct {
				Embedding []float32 `json:"embedding"`
				Index     int       `json:"index"`
			}{
				{Embedding: []float32{1.0, 2.0}, Index: 0},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	result, err := p.GenerateEmbedding(context.Background(), "text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 || result[0] != 1.0 {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestOpenAI_GenerateEmbedding_NoEmbeddings(t *testing.T) {
	// GenerateEmbedding calls GenerateEmbeddings which returns a slice of len(texts) with nil entries
	// when no data is returned. The nil entry at index 0 is returned (no error).
	// But if GenerateEmbeddings itself errors, that propagates.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Return error in response body
		resp := embeddingResponse{}
		resp.Error = &struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		}{Message: "empty input"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	_, err := p.GenerateEmbedding(context.Background(), "text")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestOpenAI_GenerateQueryEmbedding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := embeddingResponse{
			Data: []struct {
				Embedding []float32 `json:"embedding"`
				Index     int       `json:"index"`
			}{
				{Embedding: []float32{0.5}, Index: 0},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	result, err := p.GenerateQueryEmbedding(context.Background(), "query")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != 0.5 {
		t.Errorf("unexpected result: %v", result)
	}
}

// --- Chat/GenerateResponse Tests ---

func TestOpenAI_GenerateResponse_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req chatRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		// Verify system message includes context
		if len(req.Messages) < 1 || !strings.Contains(req.Messages[0].Content, "context info") {
			t.Error("expected system message with context")
		}

		resp := chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "  Hello there!  "}, FinishReason: "stop"},
			},
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{TotalTokens: 42},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	msgs := []entities.Message{
		{Role: entities.MessageRoleSystem, Content: "should be skipped"},
		{Role: entities.MessageRoleUser, Content: "Hi"},
	}

	content, tokens, err := p.GenerateResponse(context.Background(), "system prompt", msgs, "context info")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "Hello there!" {
		t.Errorf("expected trimmed content, got %q", content)
	}
	if tokens != 42 {
		t.Errorf("expected 42 tokens, got %d", tokens)
	}
}

func TestOpenAI_GenerateResponse_NoContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		// System message should just be the prompt, no extra newlines
		if strings.Contains(req.Messages[0].Content, "\n\n") {
			t.Error("system message should not contain double newline when no context")
		}

		resp := chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "OK"}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	_, _, err := p.GenerateResponse(context.Background(), "prompt", nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenAI_GenerateResponse_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatResponse{}
		resp.Error = &struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		}{Message: "quota exceeded"}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	_, _, err := p.GenerateResponse(context.Background(), "p", nil, "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "quota exceeded") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOpenAI_GenerateResponse_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(chatResponse{})
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	_, _, err := p.GenerateResponse(context.Background(), "p", nil, "")
	if err == nil {
		t.Fatal("expected error for non-OK status")
	}
}

func TestOpenAI_GenerateResponse_NoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatResponse{}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	_, _, err := p.GenerateResponse(context.Background(), "p", nil, "")
	if err == nil {
		t.Fatal("expected error for no choices")
	}
	if !strings.Contains(err.Error(), "no response choices") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOpenAI_GenerateResponse_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{invalid"))
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	_, _, err := p.GenerateResponse(context.Background(), "p", nil, "")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// --- Streaming Tests ---

func TestOpenAI_GenerateResponseStream_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if !req.Stream {
			t.Error("expected stream=true")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		// Send chunks
		chunk1 := chatStreamChunk{
			Choices: []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			}{{Delta: struct {
				Content string `json:"content"`
			}{Content: "Hello"}}},
		}
		data1, _ := json.Marshal(chunk1)
		_, _ = w.Write([]byte("data: " + string(data1) + "\n\n"))
		flusher.Flush()

		chunk2 := chatStreamChunk{
			Choices: []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			}{{Delta: struct {
				Content string `json:"content"`
			}{Content: " World"}}},
		}
		data2, _ := json.Marshal(chunk2)
		_, _ = w.Write([]byte("data: " + string(data2) + "\n\n"))
		flusher.Flush()

		// Usage chunk
		usageChunk := `{"choices":[],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`
		_, _ = w.Write([]byte("data: " + usageChunk + "\n\n"))
		flusher.Flush()

		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})

	var chunks []string
	content, tokens, err := p.GenerateResponseStream(context.Background(), "sys", nil, "", func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", content)
	}
	if tokens != 15 {
		t.Errorf("expected 15 tokens, got %d", tokens)
	}
	if len(chunks) != 2 {
		t.Errorf("expected 2 chunks, got %d", len(chunks))
	}
}

func TestOpenAI_GenerateResponseStream_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("rate limited"))
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	_, _, err := p.GenerateResponseStream(context.Background(), "sys", nil, "", func(chunk string) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for non-OK status")
	}
	if !strings.Contains(err.Error(), "status 429") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOpenAI_GenerateResponseStream_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	_, _, err := p.GenerateResponseStream(context.Background(), "sys", nil, "", func(chunk string) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for empty stream")
	}
	if !strings.Contains(err.Error(), "no text content") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOpenAI_GenerateResponseStream_OnChunkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		chunk := chatStreamChunk{
			Choices: []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			}{{Delta: struct {
				Content string `json:"content"`
			}{Content: "data"}}},
		}
		data, _ := json.Marshal(chunk)
		_, _ = w.Write([]byte("data: " + string(data) + "\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	chunkErr := context.DeadlineExceeded
	_, _, err := p.GenerateResponseStream(context.Background(), "sys", nil, "", func(chunk string) error {
		return chunkErr
	})
	if err != chunkErr {
		t.Errorf("expected onChunk error, got: %v", err)
	}
}

func TestOpenAI_GenerateResponseStream_WithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		// Verify system messages filter and context appended
		for _, m := range req.Messages {
			if m.Role == "system" && !strings.Contains(m.Content, "ctx data") {
				t.Error("expected context in system message")
			}
		}

		w.Header().Set("Content-Type", "text/event-stream")
		chunk := `{"choices":[{"delta":{"content":"OK"}}]}`
		_, _ = w.Write([]byte("data: " + chunk + "\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	msgs := []entities.Message{
		{Role: entities.MessageRoleSystem, Content: "skip me"},
		{Role: entities.MessageRoleUser, Content: "hello"},
	}

	_, _, err := p.GenerateResponseStream(context.Background(), "prompt", msgs, "ctx data", func(chunk string) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenAI_GenerateResponseStream_InvalidChunkJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		// Invalid JSON chunk should be skipped
		_, _ = w.Write([]byte("data: {invalid\n\n"))
		// Valid chunk after
		chunk := `{"choices":[{"delta":{"content":"OK"}}]}`
		_, _ = w.Write([]byte("data: " + chunk + "\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	content, _, err := p.GenerateResponseStream(context.Background(), "sys", nil, "", func(chunk string) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "OK" {
		t.Errorf("expected 'OK', got %q", content)
	}
}

func TestOpenAI_GenerateResponseStream_NonDataLines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		// Non-data lines should be skipped
		_, _ = w.Write([]byte(": comment\n\n"))
		_, _ = w.Write([]byte("event: ping\n\n"))
		chunk := `{"choices":[{"delta":{"content":"Hello"}}]}`
		_, _ = w.Write([]byte("data: " + chunk + "\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	content, _, err := p.GenerateResponseStream(context.Background(), "sys", nil, "", func(chunk string) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "Hello" {
		t.Errorf("expected 'Hello', got %q", content)
	}
}
